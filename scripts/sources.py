# -*- coding: utf-8 -*-
"""Provenance, licensing and file layout of the four source editions.

Everything here was read off the publisher's own pages before a single file was
downloaded (see reports/validation_report.md for the checks and quotes):

* BSB  - https://berean.bible/downloads.htm  -> USFM zip on bereanbible.com.
         https://berean.bible/licensing.htm places the text in the public
         domain as of 2023-04-30 and states that licensing is not required.
* WEBP - https://ebible.org/find/details.php?id=engwebp
         Public Domain, "2020 stable text edition", protocanon only (the
         Deuterocanon/Apocrypha edition is `engweb`).
* KJV  - https://ebible.org/find/details.php?id=engkjv  (eng-kjv2006)
         Public Domain, 1769 standardised text with Strong's numbers.  The
         Crown letters patent restrict printing inside the United Kingdom
         only; the text is public domain everywhere else.
* CUV  - https://ebible.org/find/details.php?id=cmn-cu89s  (cmn-cu89s, CHNCU1)
         新标点和合本 / Chinese Union Version (simplified), Public Domain,
         last updated 2021-10-15.  The 1919 CUV text is public domain.

The `files` glob is resolved inside data/raw/<source_dir>.
"""

import glob
import os

RAW_DIR = os.path.join(os.path.dirname(os.path.dirname(
    os.path.abspath(__file__))), "data", "raw")

TRANSLATIONS = [
    {
        "translation_id": "BSB",
        "format": "usfm",
        "name": "Berean Standard Bible",
        "abbreviation": "BSB",
        "language_code": "en",
        "language_name": "English",
        "edition": "Updated to the 3rd printing (USFM export)",
        "copyright": (
            "The Berean Standard Bible is in the public domain. "
            "The Berean Bible and Majority Bible texts were officially "
            "placed into the public domain as of April 30, 2023."
        ),
        "license": "Public Domain",
        "license_url": "https://berean.bible/licensing.htm",
        "source_url": "https://berean.bible/downloads.htm",
        "download_url": "https://bereanbible.com/bsb_usfm.zip",
        "archive": "bsb_usfm.zip",
        "source_dir": "bsb/bsb_usfm",
        "files": "*.usfm",
        "verify_file": "bsb.txt",
        "verify_kind": "tsv",
        "filename_style": "osis",
        "notes": (
            "bsb2usfm export. Psalm superscriptions are embedded in verse 1 "
            "as \\d immediately before \\v 1; the five Psalter divisions use "
            "\\d as well. 16 verses of the traditional numbering carry no \\v "
            "marker in the USFM but are listed as blank rows in the "
            "publisher's own bsb.txt, so they are restored here as empty rows "
            "to keep the numbering intact."
        ),
    },
    {
        "translation_id": "WEBP",
        "format": "usfm",
        "name": "World English Bible, Protestant Edition",
        "abbreviation": "WEBP",
        "language_code": "en",
        "language_name": "English",
        "edition": "2020 stable text edition",
        "copyright": (
            "The World English Bible is in the public domain. "
            "\"World English Bible\" and \"WEB\" are trademarks of eBible.org; "
            "a modified text may not be redistributed under that name."
        ),
        "license": "Public Domain",
        "license_url": "https://ebible.org/find/details.php?id=engwebp",
        "source_url": "https://ebible.org/find/details.php?id=engwebp",
        "download_url": "https://eBible.org/Scriptures/engwebp_usfm.zip",
        "archive": "engwebp_usfm.zip",
        "source_dir": "webp",
        "files": "*-*.usfm",
        "verify_file": "verify/engwebp_vpl.txt",
        "verify_kind": "vpl",
        "filename_style": "numbered",
        "notes": (
            "Protestant canon only: this is `engwebp`, not the `engweb` "
            "edition that carries the Deuterocanon/Apocrypha. Five verses "
            "(Luke 17:36, Acts 8:37, Acts 15:34, Acts 24:7, Romans 16:25) are "
            "footnote-only in the WEB and therefore stored with empty text. "
            "Romans 14:24-26 are numbered here as in the WEB, i.e. the "
            "doxology sits at the end of chapter 14 rather than at 16:25-27."
        ),
    },
    {
        "translation_id": "KJV",
        "format": "usfm",
        "name": "King James Version",
        "abbreviation": "KJV",
        "language_code": "en",
        "language_name": "English",
        "edition": "1769 standardised text (eng-kjv2006)",
        "copyright": (
            "The King James Version is in the public domain. In the United "
            "Kingdom, printing is restricted by Crown letters patent; this "
            "dataset is distributed outside that restriction."
        ),
        "license": "Public Domain",
        "license_url": "https://ebible.org/find/details.php?id=engkjv",
        "source_url": "https://ebible.org/find/details.php?id=engkjv",
        "download_url": "https://eBible.org/Scriptures/eng-kjv2006_usfm.zip",
        "archive": "eng-kjv2006_usfm.zip",
        "source_dir": "kjv",
        "files": "*-*.usfm",
        "verify_file": "verify/eng-kjv2006_vpl.txt",
        "verify_kind": "vpl",
        "filename_style": "numbered",
        "notes": (
            "Prepared by Crosswire Bible Society for eBible.org; supplied "
            "words are wrapped in \\add and appear in square brackets in the "
            "publisher's VPL export. Paragraph starts are marked with a "
            "literal pilcrow (\\u00b6) inside the verse text and are recorded "
            "in verses.paragraph_start, not in the verse text."
        ),
    },
    {
        "translation_id": "CUV",
        "format": "usfm",
        "name": "Chinese Union Version (simplified)",
        "abbreviation": "CUVs",
        "language_code": "cmn",
        "language_name": "简体中文",
        "edition": "新标点和合本 (Chinese Union New Punctuation), eBible.org "
                   "text of 2021-10-15",
        "copyright": (
            "The Chinese Union Version (1919) is in the public domain. "
            "eBible.org distributes this simplified-script edition as Public "
            "Domain (CHNCU1 / cmn-cu89s)."
        ),
        "license": "Public Domain",
        "license_url": "https://ebible.org/find/details.php?id=cmn-cu89s",
        "source_url": "https://ebible.org/find/details.php?id=cmn-cu89s",
        "download_url": "https://eBible.org/Scriptures/cmn-cu89s_usfm.zip",
        "archive": "cmn-cu89s_usfm.zip",
        "source_dir": "cuv",
        "files": "*-*.usfm",
        "verify_file": "verify/cmn-cu89s_vpl.txt",
        "verify_kind": "vpl",
        "filename_style": "numbered",
        "notes": (
            "新标点和合本, simplified script, 神版 wording. 80 verse numbers are "
            "merged into the preceding verse by this edition (70 written as "
            "\\v 29-30, 12 as \\v 6 -); only the first number of each pair is "
            "a verse, so those 82 references do not exist here. 116 psalm "
            "superscriptions are \\d before \\v 1 and are stored as "
            "psalm_title sections attached to verse 1, while the publisher's "
            "plain-text export prints them inside verse 1. \\pn proper-name "
            "markers are unwrapped, \\add supplied words are unwrapped (the "
            "export brackets them), \\qs Selah stays in the verse text, "
            "\\r parallel-passage lines become reference sections, and the "
            "U+3000 reverential space before 神/耶和华 is preserved verbatim."
        ),
    },
]

BY_ID = {t["translation_id"]: t for t in TRANSLATIONS}
TRANSLATION_IDS = [t["translation_id"] for t in TRANSLATIONS]


def source_files(translation):
    """Return {usfm_book_id: path} for one translation."""
    from canon import BY_USFM
    directory = os.path.join(RAW_DIR, translation["source_dir"])
    found = {}
    for path in glob.glob(os.path.join(directory, translation["files"])):
        stem = os.path.basename(path).split(".")[0]
        if translation["filename_style"] == "osis":
            # bsb_usfm/GEN.usfm
            code = stem
        else:
            # engwebp: 01-GENengwebp.usfm ; eng-kjv2006: 01-GENeng-kjv2006.usfm
            head, _, tail = stem.partition("-")
            code = tail[:3] if head.isdigit() else ""
        if code in BY_USFM:
            found[code] = path
    return found


def archive_path(translation):
    return os.path.join(RAW_DIR, translation["archive"])


def verify_path(translation):
    path = os.path.join(RAW_DIR, translation["verify_file"])
    return path if os.path.exists(path) else None


# ---------------------------------------------------------------------------
# Repairs to exporter defects, each verified against the publisher's own
# plain-text export before it is applied.  Nothing else in the verse text is
# rewritten: every other difference from the USFM is typographic whitespace
# that usfm_parser.tidy() normalises, and it is validated verse-by-verse
# against bsb.txt / the eBible VPL exports (see scripts/reference_compare.py).
# ---------------------------------------------------------------------------

TEXT_REPAIRS = [
    {
        "translation_id": "BSB",
        "format": "usfm", "usfm_id": "JER", "chapter": 22, "verse": 24,
        "find": "\u201cevenif you, Coniah",
        "replace": "\u201ceven if you, Coniah",
        "reason": "bsb2usfm dropped the space between 'even' and 'if'; "
                  "bsb.txt reads 'even if'.",
    },
    {
        "translation_id": "BSB",
        "format": "usfm", "usfm_id": "JER", "chapter": 29, "verse": 19,
        "find": "declares the LORD,which I sent",
        "replace": "declares the LORD, which I sent",
        "reason": "bsb2usfm dropped the space after the comma; "
                  "bsb.txt reads 'LORD, which'.",
    },
]
