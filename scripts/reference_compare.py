# -*- coding: utf-8 -*-
"""Compare the imported verse text with the publishers' own plain-text exports.

Two independent artefacts are used, both published by the same organisations
that publish the USFM:

    BSB  -> https://bereanbible.com/bsb.txt              (tab separated)
    WEBP -> eBible.org engwebp_vpl.txt                   (verse per line)
    KJV  -> eBible.org eng-kjv2006_vpl.txt               (verse per line)

The USFM package and the plain-text export are produced by different build
steps, so they are a genuine cross-check of the converter: the two texts are
typed by different pipelines and formatted differently (the KJV uses `[added]`
for supplied words, the BSB tab file drops the spaces around em dashes).

Comparison is two tiered:
    exact      after NFC + whitespace collapse
    equivalent after also normalising spacing around dashes / punctuation and
               dropping KJV's square brackets
Anything left is reported as a real mismatch and must be inspected by hand.
"""

import re
import unicodedata

BSB_BOOK_NAMES = {
    "Genesis": "Gen", "Exodus": "Exod", "Leviticus": "Lev", "Numbers": "Num",
    "Deuteronomy": "Deut", "Joshua": "Josh", "Judges": "Judg", "Ruth": "Ruth",
    "1 Samuel": "1Sam", "2 Samuel": "2Sam", "1 Kings": "1Kgs",
    "2 Kings": "2Kgs", "1 Chronicles": "1Chr", "2 Chronicles": "2Chr",
    "Ezra": "Ezra", "Nehemiah": "Neh", "Esther": "Esth", "Job": "Job",
    "Psalm": "Ps", "Proverbs": "Prov", "Ecclesiastes": "Eccl",
    "Song of Solomon": "Song", "Isaiah": "Isa", "Jeremiah": "Jer",
    "Lamentations": "Lam", "Ezekiel": "Ezek", "Daniel": "Dan", "Hosea": "Hos",
    "Joel": "Joel", "Amos": "Amos", "Obadiah": "Obad", "Jonah": "Jonah",
    "Micah": "Mic", "Nahum": "Nah", "Habakkuk": "Hab", "Zephaniah": "Zeph",
    "Haggai": "Hag", "Zechariah": "Zech", "Malachi": "Mal", "Matthew": "Matt",
    "Mark": "Mark", "Luke": "Luke", "John": "John", "Acts": "Acts",
    "Romans": "Rom", "1 Corinthians": "1Cor", "2 Corinthians": "2Cor",
    "Galatians": "Gal", "Ephesians": "Eph", "Philippians": "Phil",
    "Colossians": "Col", "1 Thessalonians": "1Thess",
    "2 Thessalonians": "2Thess", "1 Timothy": "1Tim", "2 Timothy": "2Tim",
    "Titus": "Titus", "Philemon": "Phlm", "Hebrews": "Heb", "James": "Jas",
    "1 Peter": "1Pet", "2 Peter": "2Pet", "1 John": "1John",
    "2 John": "2John", "3 John": "3John", "Jude": "Jude",
    "Revelation": "Rev",
}

VPL_RE = re.compile(r"^([0-9A-Z]{3})\s+(\d+):(\d+)(?:-\d+)?\s?(.*)$")

# The VPL exports use their own 3-letter book codes, several of which differ
# from the USFM codes used by the same publisher for the same books.
VPL_ALIASES = {
    "SOL": "SNG", "1JO": "1JN", "2JO": "2JN", "3JO": "3JN", "EZE": "EZK",
    "JAM": "JAS", "JOE": "JOL", "JOH": "JHN", "MAR": "MRK", "NAH": "NAM",
    "PHI": "PHP",
}
TSV_REF_RE = re.compile(r"^(.*?)\s+(\d+):(\d+)$")


def norm(s):
    s = unicodedata.normalize("NFC", s)
    s = s.replace("\u00a0", " ").replace("\u200b", "")
    s = re.sub(r"\s+", " ", s)
    return s.strip()


def norm_loose(s):
    s = norm(s)
    s = s.replace("\u00b6", "")            # KJV paragraph mark
    s = re.sub(r"\s*[\u2014\u2013]\s*", "\u2014", s)
    s = re.sub(r"\s+([,.;:!?])", r"\1", s)
    s = re.sub(r"\s+([\u201d\u2019\u2019])", r"\1", s)
    s = s.replace("[", "").replace("]", "")
    # eBible marks a verse the edition merges with the next one with a leading
    # hyphen (`\v 6 -连…`, kept as "-连[圣]殿…" in the VPL).  The hyphen is a
    # numbering mark, not scripture, and the database drops it.
    s = re.sub(r"^-(?=[^\s])", "", s)
    # Chinese is set without spaces between characters.  The USFM and the VPL
    # differ on where the converter left one ("以弗 正坐在", "房屋 成为荒凉"),
    # which is typography and not wording, so both sides normalise to no space.
    s = re.sub(r"(?<=[\u3400-\u4dbf\u4e00-\u9fff\uf900-\ufaff])\s+"
               r"(?=[\u3400-\u4dbf\u4e00-\u9fff\uf900-\ufaff])", "", s)
    return re.sub(r"\s+", " ", s).strip()


def load_vpl(path):
    """eBible verse-per-line export -> {(usfm, chapter, verse): text}."""
    out = {}
    with open(path, "r", encoding="utf-8-sig") as fh:
        for line in fh:
            line = line.rstrip("\r\n")
            m = VPL_RE.match(line)
            if not m:
                continue
            book = VPL_ALIASES.get(m.group(1), m.group(1))
            out[(book, int(m.group(2)), int(m.group(3)))] = m.group(4)
    return out


def load_bsb_tsv(path):
    """BereanBible.com tab separated export -> {(osis, chapter, verse): text}."""
    out = {}
    with open(path, "r", encoding="utf-8-sig") as fh:
        for line in fh:
            line = line.rstrip("\r\n")
            if not line.strip():
                continue
            parts = line.split("\t")
            if len(parts) < 2:
                continue
            m = TSV_REF_RE.match(parts[0].strip())
            if not m:
                continue
            book = BSB_BOOK_NAMES.get(m.group(1))
            if book is None:
                continue
            out[(book, int(m.group(2)), int(m.group(3)))] = parts[1].strip()
    return out


def compare(parsed, reference, key_of):
    """Compare {key: text} dicts.

    `key_of` maps a parsed key to the reference key.  Returns a dict with
    counts plus the mismatch list (capped).
    """
    exact = 0
    loose = 0
    missing = []      # in reference, absent from parsed output
    extra = []        # in parsed output, absent from reference
    mismatched = []
    compared = 0
    for key, text in parsed.items():
        rkey = key_of(key)
        if rkey not in reference:
            extra.append(key)
            continue
        compared += 1
        ref = reference[rkey]
        if norm(text) == norm(ref):
            exact += 1
        elif norm_loose(text) == norm_loose(ref):
            loose += 1
        else:
            mismatched.append((key, text, ref))
    ref_keys = set(key_of(k) for k in parsed)
    for rkey in reference:
        if rkey not in ref_keys:
            missing.append(rkey)
    return {
        "compared": compared,
        "exact": exact,
        "equivalent": loose,
        "mismatched": mismatched,
        "missing": missing,
        "extra": extra,
    }
