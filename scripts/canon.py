# -*- coding: utf-8 -*-
"""Canonical metadata for the 66-book Protestant canon.

Each entry: (order, osis_id, usfm_id, display_name, testament, section, chapters)

`chapters` is the traditional chapter count for the book and is used by the
verification step to prove that a translation carries all 66 books with the
expected number of chapters.  It is metadata, not data taken from any one
translation.
"""

BOOKS = [
    (1,  "Gen",   "GEN", "Genesis",              "OT", "Law",                50),
    (2,  "Exod",  "EXO", "Exodus",               "OT", "Law",                40),
    (3,  "Lev",   "LEV", "Leviticus",            "OT", "Law",                27),
    (4,  "Num",   "NUM", "Numbers",              "OT", "Law",                36),
    (5,  "Deut",  "DEU", "Deuteronomy",          "OT", "Law",                34),
    (6,  "Josh",  "JOS", "Joshua",               "OT", "History",            24),
    (7,  "Judg",  "JDG", "Judges",               "OT", "History",            21),
    (8,  "Ruth",  "RUT", "Ruth",                 "OT", "History",             4),
    (9,  "1Sam",  "1SA", "1 Samuel",             "OT", "History",            31),
    (10, "2Sam",  "2SA", "2 Samuel",             "OT", "History",            24),
    (11, "1Kgs",  "1KI", "1 Kings",              "OT", "History",            22),
    (12, "2Kgs",  "2KI", "2 Kings",              "OT", "History",            25),
    (13, "1Chr",  "1CH", "1 Chronicles",         "OT", "History",            29),
    (14, "2Chr",  "2CH", "2 Chronicles",         "OT", "History",            36),
    (15, "Ezra",  "EZR", "Ezra",                 "OT", "History",            10),
    (16, "Neh",   "NEH", "Nehemiah",             "OT", "History",            13),
    (17, "Esth",  "EST", "Esther",               "OT", "History",            10),
    (18, "Job",   "JOB", "Job",                  "OT", "Wisdom",             42),
    (19, "Ps",    "PSA", "Psalms",               "OT", "Wisdom",            150),
    (20, "Prov",  "PRO", "Proverbs",             "OT", "Wisdom",             31),
    (21, "Eccl",  "ECC", "Ecclesiastes",         "OT", "Wisdom",             12),
    (22, "Song",  "SNG", "Song of Solomon",      "OT", "Wisdom",              8),
    (23, "Isa",   "ISA", "Isaiah",               "OT", "Major Prophets",     66),
    (24, "Jer",   "JER", "Jeremiah",             "OT", "Major Prophets",     52),
    (25, "Lam",   "LAM", "Lamentations",         "OT", "Major Prophets",      5),
    (26, "Ezek",  "EZK", "Ezekiel",              "OT", "Major Prophets",     48),
    (27, "Dan",   "DAN", "Daniel",               "OT", "Major Prophets",     12),
    (28, "Hos",   "HOS", "Hosea",                "OT", "Minor Prophets",     14),
    (29, "Joel",  "JOL", "Joel",                 "OT", "Minor Prophets",      3),
    (30, "Amos",  "AMO", "Amos",                 "OT", "Minor Prophets",      9),
    (31, "Obad",  "OBA", "Obadiah",              "OT", "Minor Prophets",      1),
    (32, "Jonah", "JON", "Jonah",                "OT", "Minor Prophets",      4),
    (33, "Mic",   "MIC", "Micah",                "OT", "Minor Prophets",      7),
    (34, "Nah",   "NAM", "Nahum",                "OT", "Minor Prophets",      3),
    (35, "Hab",   "HAB", "Habakkuk",             "OT", "Minor Prophets",      3),
    (36, "Zeph",  "ZEP", "Zephaniah",            "OT", "Minor Prophets",      3),
    (37, "Hag",   "HAG", "Haggai",               "OT", "Minor Prophets",      2),
    (38, "Zech",  "ZEC", "Zechariah",            "OT", "Minor Prophets",     14),
    (39, "Mal",   "MAL", "Malachi",              "OT", "Minor Prophets",      4),
    (40, "Matt",  "MAT", "Matthew",              "NT", "Gospels",            28),
    (41, "Mark",  "MRK", "Mark",                 "NT", "Gospels",            16),
    (42, "Luke",  "LUK", "Luke",                 "NT", "Gospels",            24),
    (43, "John",  "JHN", "John",                 "NT", "Gospels",            21),
    (44, "Acts",  "ACT", "Acts",                 "NT", "Acts",               28),
    (45, "Rom",   "ROM", "Romans",               "NT", "Pauline Epistles",   16),
    (46, "1Cor",  "1CO", "1 Corinthians",        "NT", "Pauline Epistles",   16),
    (47, "2Cor",  "2CO", "2 Corinthians",        "NT", "Pauline Epistles",   13),
    (48, "Gal",   "GAL", "Galatians",            "NT", "Pauline Epistles",    6),
    (49, "Eph",   "EPH", "Ephesians",            "NT", "Pauline Epistles",    6),
    (50, "Phil",  "PHP", "Philippians",          "NT", "Pauline Epistles",    4),
    (51, "Col",   "COL", "Colossians",           "NT", "Pauline Epistles",    4),
    (52, "1Thess","1TH", "1 Thessalonians",      "NT", "Pauline Epistles",    5),
    (53, "2Thess","2TH", "2 Thessalonians",      "NT", "Pauline Epistles",    3),
    (54, "1Tim",  "1TI", "1 Timothy",            "NT", "Pauline Epistles",    6),
    (55, "2Tim",  "2TI", "2 Timothy",            "NT", "Pauline Epistles",    4),
    (56, "Titus", "TIT", "Titus",                "NT", "Pauline Epistles",    3),
    (57, "Phlm",  "PHM", "Philemon",             "NT", "Pauline Epistles",    1),
    (58, "Heb",   "HEB", "Hebrews",              "NT", "General Epistles",   13),
    (59, "Jas",   "JAS", "James",                "NT", "General Epistles",    5),
    (60, "1Pet",  "1PE", "1 Peter",              "NT", "General Epistles",    5),
    (61, "2Pet",  "2PE", "2 Peter",              "NT", "General Epistles",    3),
    (62, "1John", "1JN", "1 John",               "NT", "General Epistles",    5),
    (63, "2John", "2JN", "2 John",               "NT", "General Epistles",    1),
    (64, "3John", "3JN", "3 John",               "NT", "General Epistles",    1),
    (65, "Jude",  "JUD", "Jude",                 "NT", "General Epistles",    1),
    (66, "Rev",   "REV", "Revelation",           "NT", "Apocalyptic",        22),
]

FIELDS = ("book_id", "osis_id", "usfm_id", "name", "testament", "section", "chapters")

BY_USFM = {b[2]: dict(zip(FIELDS, b)) for b in BOOKS}
BY_OSIS = {b[1]: dict(zip(FIELDS, b)) for b in BOOKS}

OT_BOOKS = [b for b in BOOKS if b[4] == "OT"]
NT_BOOKS = [b for b in BOOKS if b[4] == "NT"]

TOTAL_CHAPTERS = sum(b[6] for b in BOOKS)
TOTAL_BOOKS = len(BOOKS)


def validate_canon():
    """Sanity checks on the table itself; returns a list of problems."""
    problems = []
    if len(BOOKS) != 66:
        problems.append("expected 66 books, found %d" % len(BOOKS))
    for i, b in enumerate(BOOKS):
        if b[0] != i + 1:
            problems.append("book order gap at %s" % (b[1],))
    if [b[0] for b in BOOKS] != list(range(1, 67)):
        problems.append("book_id sequence is not 1..66")
    if len(BY_USFM) != 66:
        problems.append("duplicate usfm_id")
    if len(BY_OSIS) != 66:
        problems.append("duplicate osis_id")
    if sum(b[6] for b in OT_BOOKS) != 929:
        problems.append("OT chapter total != 929")
    if sum(b[6] for b in NT_BOOKS) != 260:
        problems.append("NT chapter total != 260")
    return problems


if __name__ == "__main__":
    errs = validate_canon()
    print("books=%d chapters=%d OT=%d NT=%d" % (
        TOTAL_BOOKS, TOTAL_CHAPTERS, len(OT_BOOKS), len(NT_BOOKS)))
    print("canon problems:", errs or "none")
