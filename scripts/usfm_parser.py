# -*- coding: utf-8 -*-
"""USFM 3.x -> plain verse records.

The parser is deliberately conservative: it drops theological markup that is
not part of the scripture text (footnotes, cross-reference notes, section
headings, psalm titles, speaker labels, acrostic letters, chapter labels and
BSB's `\\r` parallel-passage lines) but it never rewrites a word of the verse
text itself.  Markup that wraps words (\\w, \\add, \\nd, \\wj, \\qs, \\k, \\tl,
\\bk, \\fq, \\mi, and the \\+... nested forms) is unwrapped so only the
surface words survive.

Output:
    verses   : dict[(chapter, verse)] -> {"text": str, "paragraph_start": int}
    headings : list of {"chapter", "verse", "kind", "level", "text", "marker"}
    meta     : dict with \\id, \\h, \\toc* values and parse statistics

`\sp` speaker lines in the Song of Songs are kept as running text: they stand
inside a verse, and the publisher's own plain-text export prints them inline
("...into his rooms. Friends We will be glad..."), so treating them as headings
would both invent a position the source does not have and change the verse text.

Headings are attached to the verse that FOLLOWS them, which is how readers
expect them to be rendered.  A heading that trails the last verse of a chapter
is attached to last_verse + 1.
"""

import re

# ---------------------------------------------------------------- markers ---

# Paragraph-level markers whose inline text is scripture (continuation of the
# open verse).  None of these take a numeric argument in the sources used.
PARAGRAPH_MARKERS = {
    "p", "m", "mi", "nb", "pb", "pc", "pmo", "pm", "pmc", "pmr",
    "pi1", "pi2", "pi3", "pi4", "ph1", "ph2", "ph3",
    "q1", "q2", "q3", "q4", "qc", "qr", "qm1", "qm2", "qm3",
    "li1", "li2", "li3", "li4", "lf1", "lf2",
    "b", "cls", "mr", "sp",
}

# Paragraph-level markers that carry a heading (not scripture text).
HEADING_MARKERS = {
    "s1": ("section", 1), "s2": ("section", 2), "s3": ("section", 3),
    "s4": ("section", 4),
    "d": ("psalm_title", 1),
    "r": ("reference", 1),
    "qa": ("acrostic", 1),
    "cl": ("chapter_label", 1),
    "ms": ("major_section", 1), "ms1": ("major_section", 1),
    "ms2": ("major_section", 2), "ms3": ("major_section", 3),
    "is1": ("intro_section", 1), "is2": ("intro_section", 2),
}

# Book-level / document-level markers: whole paragraph is metadata, discarded.
BOOK_META_MARKERS = {
    "id", "usfm", "ide", "h", "toc1", "toc2", "toc3",
    "mt", "mt1", "mt2", "mt3", "mt4",
    "imt1", "imt2", "imt3", "imt4",
}

# Markers that wrap words and must be unwrapped (text kept).
INLINE_WRAPPERS = {
    "add", "nd", "wj", "qs", "k", "tl", "bk", "fq", "mi", "em", "bd", "it",
    "sc", "no", "pn", "qt", "sig", "sls", "ord", "w", "wg", "wh", "wa", "wk",
    "ref", "xt", "xo", "fr", "ft", "fqa", "fv", "fdc",
    "add*", "nd*", "wj*", "qs*", "k*", "tl*", "bk*", "fq*", "mi*", "em*",
    "bd*", "it*", "sc*", "no*", "pn*", "qt*", "sig*", "sls*", "ord*", "w*",
    "wg*", "wh*", "wa*", "wk*", "ref*", "xt*", "xo*", "fr*", "ft*", "fqa*",
    "fv*", "fdc*",
    "+add", "+nd", "+wj", "+w", "+wh", "+k", "+bk", "+qt", "+sig", "+em",
    "+bd", "+it", "+sc", "+no", "+pn",
    "+add*", "+nd*", "+wj*", "+w*", "+wh*", "+k*", "+bk*", "+qt*", "+sig*",
    "+em*", "+bd*", "+it*", "+sc*", "+no*", "+pn*",
}

# Markers whose text may carry a `|attribute` suffix that must be stripped.
ATTR_MARKERS = {"w", "+w", "k", "+k", "wk", "ref", "xt", "xo", "rb", "rb*"}

NOTE_RE = re.compile(r"\\f(?![a-zA-Z0-9])(?:[^\\]|\\(?!f\*))*\\f\*"
                     r"|\\x(?![a-zA-Z0-9])(?:[^\\]|\\(?!x\*))*\\x\*",
                     re.DOTALL)

MARKER_RE = re.compile(r"(\\\+?[A-Za-z0-9]+\*?)")

PILCROW = "\u00b6"

# Prose paragraph markers that may open a new paragraph.  A verse is flagged as
# a paragraph start when one of these markers stands alone on its own line
# immediately before the `\v` marker (BSB/WEBP write `\p \v 1 In the
# beginning...`), i.e. when nothing but whitespace follows it.
PROSE_PARA_MARKERS = {
    "p", "m", "pc", "pm", "pmc", "pmr", "nb", "pb",
    "pi1", "pi2", "pi3", "pi4", "ph1", "ph2", "ph3",
}


def _note_replacement(match):
    """Replace a note with nothing, unless that would fuse two words.

    A note normally sits between existing spaces (`God\\f ...\\f* created`) or
    against punctuation (`offspring.\\f ...\\f*”`).  Substituting a space
    unconditionally produces `offspring. ”`, so the replacement is empty
    unless dropping the note would join the characters on either side:
    `Samuel,\\f ...\\f*saying` must become `Samuel, saying`.
    """
    text = match.string
    before = text[match.start() - 1] if match.start() > 0 else " "
    after = text[match.end()] if match.end() < len(text) else " "
    if after.isalnum() and (before.isalnum() or before in ",.;:!?)]"):
        return " "
    # `Hebrew,\f ...\f*\wj ‘Saul` — the note separates a comma from an opening
    # quotation mark, and the publisher keeps the space there.
    if before in ",.;:!?" and after in "\u201c\u2018\"([\u2014\u2013":
        return " "
    return ""


def strip_notes(text):
    """Remove footnote (\\f...\\f*) and cross-reference (\\x...\\x*) notes.

    Repeated until stable: a note body can itself contain a `\\x`-shaped run,
    and a second pass cleans up anything the first left behind.
    """
    for _ in range(3):
        new = NOTE_RE.sub(_note_replacement, text)
        if new == text:
            break
        text = new
    return text


# Punctuation spurs produced by the BSB converter (bsb2usfm) and, in smaller
# numbers, by the eBible exporters.  They are typography, never wording: the
# words, verse numbers and canonical order are untouched.  Every rule below is
# validated by comparing the finished text against the publisher's own plain
# text export (see scripts/reference_compare.py), which matches 100% after
# tidying.
TIDY_RULES = [
    (re.compile(r"\[\u2019\u2019\]"), ""),          # bsb2usfm debris: [’’]
    (re.compile(r"(?:(?<=\s)|^)vvv(?=\s|$)"), ""),  # bsb2usfm debris: vvv
    (re.compile(r"(?<=\s)\.\s+\.\s+\.(?=\s)"), ""),  # " . . . " -> ""
    (re.compile(r"\s+([,;:!?\)\]\u201d\u2019])"), r"\1"),
    (re.compile(r"([(\u201c\u2018\[])\s+"), r"\1"),
    (re.compile(r"([:;])([A-Za-z])"), r"\1 \2"),    # ":This" -> ": This"
    # Removing a note between a comma and an opening quote leaves
    # `Hebrew,‘Saul`; the publisher's export has `Hebrew, ‘Saul`.
    (re.compile(r"([,;:!?])(?=[\u201c\u2018\"])"), r"\1 "),
    # Chinese is written without spaces between characters; the eBible CUV
    # converter leaves one wherever a \pn / \add marker was wrapped
    # ("他侄儿 \pn 被掳去\pn*"), and the publisher's plain-text export has
    # none.  A space between two Han characters is never legitimate typography.
    (re.compile(r"(?<=[\u3400-\u4dbf\u4e00-\u9fff])\s+"
                r"(?=[\u3400-\u4dbf\u4e00-\u9fff])"), ""),
]


def tidy(text):
    """Normalise punctuation spacing left behind by the USFM exporters."""
    text = _protect_ideographic_space(text)
    for pattern, repl in TIDY_RULES:
        text = pattern.sub(repl, text)
    return _restore_ideographic_space(re.sub(r"\s+", " ", text)).strip()


def norm_ws(text):
    """Collapse every run of whitespace (incl. newlines) to one space."""
    text = text.replace("\u00a0", " ").replace("\u200b", "")
    text = text.replace("\ufeff", "")
    text = _protect_ideographic_space(text)
    text = re.sub(r"\s+", " ", text)
    return _restore_ideographic_space(text).strip()


# The Chinese Union Version sets a full-width space (U+3000) before 神 and 耶和华
# as a reverential convention ("起初，　神创造天地。").  It is typography the
# publisher ships in both the USFM and the plain-text export, so it survives the
# whitespace normalisation above: collapsing it to an ASCII space would quietly
# alter the edition.  The sentinel never occurs in the source files.
IDEOGRAPHIC_SPACE = "\u3000"
_IDEO_SENTINEL = "\x00"


def _protect_ideographic_space(text):
    return text.replace(IDEOGRAPHIC_SPACE, _IDEO_SENTINEL)


def _restore_ideographic_space(text):
    return text.replace(_IDEO_SENTINEL, IDEOGRAPHIC_SPACE)


class UsfmParseError(Exception):
    pass


class UsfmParser(object):
    def __init__(self):
        self.verses = {}
        self.merged = []            # verse numbers this edition merges
        self.headings = []
        self.meta = {}
        self.unknown_markers = {}
        self.stats = {
            "chapters": 0,
            "verses": 0,
            "headings": 0,
            "notes_removed": 0,
            "paragraph_marks": 0,
            "merged_numbers": 0,
        }

    # -- helpers ---------------------------------------------------------
    def _flush_heading(self):
        if self._heading_buf is None:
            return
        txt = tidy(norm_ws("".join(self._heading_buf)))
        kind = self._heading_kind
        level = self._heading_level
        marker = self._heading_marker
        self._heading_buf = None
        if txt:
            self._pending_headings.append(
                {"kind": kind, "level": level, "text": txt, "marker": marker})
            self.stats["headings"] += 1

    def _assign_pending(self, verse_no):
        for h in self._pending_headings:
            h["chapter"] = self._chapter
            h["verse"] = verse_no
            self.headings.append(h)
        self._pending_headings = []

    def _emit(self, txt):
        if not txt:
            return
        if self._in_attr and "|" in txt:
            txt = txt.split("|", 1)[0]
        if self._mode == "speaker":
            if self._speaker_buf is not None:
                self._speaker_buf.append(txt)
        elif self._mode == "heading":
            if self._heading_buf is not None:
                self._heading_buf.append(txt)
        elif self._mode == "verse" and self._cur_verse is not None:
            self._verse_parts.append(txt)

    def _resolve_speaker(self, bare, has_star, text):
        """Place a pending \sp label once the next marker is known.

        eBible writes `\sp Beloved` on its own line in two different
        positions:

            \sp Beloved          <- heads the next verse
            \q1
            \v 2 Let him kiss me ...

            \sp Friends          <- spoken in the middle of this verse
            \q1 We will be glad ...

        The publisher's plain-text export prints the second kind inline and
        drops the first kind, so the distinction has to be made here: a label
        whose next content is a `\v` / `\c` marker becomes a heading for that
        verse, anything else rejoins the text of the verse it stands in.
        """
        if self._speaker_buf is None:
            return
        label = tidy(norm_ws("".join(self._speaker_buf)))
        if bare in ("v", "c") and not has_star:
            if label:
                self._pending_headings.append(
                    {"kind": "speaker", "level": 1, "text": label,
                     "marker": "sp"})
                self.stats["headings"] += 1
            self._speaker_buf = None
            self._mode = "plain"
            return
        if not has_star and bare in PARAGRAPH_MARKERS and not text.strip():
            return          # a bare \q1 line: keep the label pending
        raw = "".join(self._speaker_buf)
        self._speaker_buf = None
        self._mode = "verse" if self._cur_verse is not None else "plain"
        self._emit(raw)

    def _close_verse(self):
        if self._cur_verse is None:
            return
        raw = "".join(self._verse_parts)
        para = self._verse_para
        if PILCROW in raw:
            para = 1
            self.stats["paragraph_marks"] += 1
            raw = raw.replace(PILCROW, " ")
        txt = tidy(norm_ws(raw))
        key = (self._chapter, self._cur_verse)
        if key in self.verses and self.verses[key]["text"] != txt:
            # Same verse number repeated: concatenate rather than overwrite.
            prev = self.verses[key]
            merged = norm_ws(prev["text"] + " " + txt) if txt else prev["text"]
            prev["text"] = merged
            prev["paragraph_start"] = prev["paragraph_start"] or para
        elif key in self.verses:
            self.verses[key]["paragraph_start"] = (
                self.verses[key]["paragraph_start"] or para)
        else:
            self.verses[key] = {"text": txt, "paragraph_start": para}
            self.stats["verses"] += 1
        self._verse_parts = []

    # -- main ------------------------------------------------------------
    def parse(self, path):
        with open(path, "rb") as fh:
            raw = fh.read().decode("utf-8-sig", errors="strict")
        raw = raw.replace("\r\n", "\n").replace("\r", "\n")

        notes = NOTE_RE.findall(raw)
        self.stats["notes_removed"] += len(notes)
        raw = strip_notes(raw)

        parts = MARKER_RE.split(raw)
        pairs = []
        if parts and parts[0]:
            pairs.append(("", parts[0]))
        for i in range(1, len(parts) - 1, 2):
            pairs.append((parts[i], parts[i + 1]))

        self._chapter = 0
        self._cur_verse = None
        self._last_verse = 0
        self._verse_parts = []
        self._mode = "plain"
        self._heading_buf = None
        self._heading_kind = None
        self._heading_level = 0
        self._heading_marker = None
        self._pending_headings = []
        self._in_attr = False
        self._seen_chapter = False
        self._pending_para = 0
        self._verse_para = 0
        self._speaker_buf = None

        for marker, text in pairs:
            if marker == "":
                self._emit(text)
                continue
            name = marker[1:].lstrip("+")
            has_star = name.endswith("*")
            bare = name[:-1] if has_star else name

            if bare != "sp":
                self._resolve_speaker(bare, has_star, text)

            if bare in ("f", "x"):
                # Should have been removed already.
                continue

            if bare == "c" and not has_star:
                m = re.match(r"\s*(\d+)", text)
                if not m:
                    raise UsfmParseError(
                        "%s: \\c without a number near %r" % (path, text[:40]))
                self._flush_heading()
                self._close_verse()
                self._assign_pending(self._last_verse + 1)
                self._chapter = int(m.group(1))
                self._cur_verse = None
                self._last_verse = 0
                self._mode = "plain"
                self._seen_chapter = True
                self.stats["chapters"] += 1
                rest = text[m.end():]
                if rest.strip():
                    self._emit(rest)
                continue

            if bare == "v" and not has_star:
                # eBible marks a verse that the edition merges with the next
                # one as either `\\v 29-30` (the number of the pair it keeps)
                # or `\\v 6 -` ; only the first number is a verse.
                m = re.match(r"\s*(\d+)(?:\s*-\s*(\d+))?", text)
                if not m:
                    raise UsfmParseError(
                        "%s: \\v without a number near %r" % (path, text[:40]))
                self._flush_heading()
                self._close_verse()
                self._cur_verse = int(m.group(1))
                self._last_verse = self._cur_verse
                if m.group(2):
                    # `\v 20-21` prints one verse whose number covers both; the
                    # second number still exists in the edition's numbering, so
                    # it is recorded and stored as an empty verse row.
                    for extra in range(self._cur_verse + 1,
                                       int(m.group(2)) + 1):
                        self.merged.append((self._chapter, extra))
                        self.stats["merged_numbers"] += 1
                self._verse_para = self._pending_para
                self._pending_para = 0
                self._assign_pending(self._cur_verse)
                self._mode = "verse"
                self._verse_parts = []
                rest = re.sub(r"^\s*-\s*", "", text[m.end():])
                if rest:
                    self._emit(rest)
                continue

            if bare in HEADING_MARKERS and not has_star:
                self._close_verse()
                self._flush_heading()
                kind, level = HEADING_MARKERS[bare]
                self._heading_kind = kind
                self._heading_level = level
                self._heading_marker = bare
                self._heading_buf = [text]
                self._mode = "heading"
                continue

            if bare in BOOK_META_MARKERS and not has_star:
                self._flush_heading()
                if bare in ("id", "h", "toc1", "toc2", "toc3"):
                    self.meta.setdefault(bare, norm_ws(text))
                self._mode = "plain"
                continue

            if bare == "sp" and not has_star:
                self._flush_heading()
                self._speaker_buf = [text]
                self._mode = "speaker"
                continue

            if bare in PARAGRAPH_MARKERS and not has_star:
                if bare in PROSE_PARA_MARKERS and not text.strip():
                    self._pending_para = 1
                self._flush_heading()
                if self._cur_verse is not None:
                    self._mode = "verse"
                self._emit(text)
                continue

            if bare in ATTR_MARKERS:
                # A word/keyword marker carries one USFM delimiter space that
                # is not part of the word: "\w the|strong=...\w*".  Dropping
                # exactly one space keeps `“\w There` from becoming `“ There`.
                self._in_attr = not has_star
                if not has_star and text.startswith(" "):
                    text = text[1:]
                self._emit(text)
                continue

            if name in INLINE_WRAPPERS or bare in INLINE_WRAPPERS:
                # As with \\w, the one space that follows a character marker is
                # the marker/text delimiter, not content: `又\\add 造\\add*众星`
                # must not become `又 造众星` in Chinese text, which has no
                # spaces to absorb the extra one.
                if not has_star and text.startswith(" "):
                    text = text[1:]
                self._emit(text)
                continue

            # Unknown marker: keep the words, log it.
            self.unknown_markers[marker] = self.unknown_markers.get(marker, 0) + 1
            self._emit(text)

        self._resolve_speaker("v", False, "")
        self._close_verse()
        self._flush_heading()
        self._assign_pending(self._last_verse + 1)

        if not self._seen_chapter:
            raise UsfmParseError("%s: no \\c marker found" % path)

        for h in self.headings:
            if "chapter" not in h:
                h["chapter"] = self._chapter
                h["verse"] = self._last_verse + 1
        return self.verses, self.headings, self.meta


def parse_file(path):
    p = UsfmParser()
    verses, headings, meta = p.parse(path)
    return {
        "verses": verses,
        "merged": sorted(p.merged),
        "headings": headings,
        "meta": meta,
        "unknown_markers": p.unknown_markers,
        "stats": p.stats,
    }
