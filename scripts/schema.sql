-- Schema for the offline English-study Bible database.
--
-- One database holds all three translations; `translation_id` separates them.
-- Verse identity is (translation_id, book_id, chapter, verse) and never a line
-- number or a row id, so translations whose versification differs (Romans
-- 14:24-26 vs 16:25-27, the 16 verses the BSB leaves blank, the 5 verses the
-- WEB only footnotes) still line up verse-for-verse wherever the numbering
-- really agrees.

PRAGMA foreign_keys = ON;

-- ------------------------------------------------------------------ meta ---

CREATE TABLE dataset_meta (
    key         TEXT PRIMARY KEY,
    value       TEXT NOT NULL
);

CREATE TABLE translations (
    translation_id   TEXT PRIMARY KEY,          -- 'BSB', 'WEBP', 'KJV'
    name             TEXT NOT NULL,
    abbreviation     TEXT NOT NULL,
    language_code    TEXT NOT NULL,             -- ISO 639-1
    language_name    TEXT NOT NULL,
    edition          TEXT NOT NULL,             -- version / edition string
    copyright        TEXT NOT NULL,
    license          TEXT NOT NULL,
    license_url      TEXT,
    source_url       TEXT NOT NULL,             -- publisher's landing page
    download_url     TEXT NOT NULL,             -- exact file that was fetched
    download_date    TEXT NOT NULL,             -- ISO 8601 UTC
    sha256           TEXT,                      -- of the downloaded archive
    book_count       INTEGER NOT NULL,
    chapter_count    INTEGER NOT NULL,
    verse_count      INTEGER NOT NULL,
    empty_verse_count INTEGER NOT NULL,
    notes            TEXT
);

-- -------------------------------------------------------------- canon -----

CREATE TABLE books (
    book_id       INTEGER PRIMARY KEY,          -- 1..66, canonical order
    osis_id       TEXT NOT NULL UNIQUE,         -- 'Gen', '1Sam', 'Rev'
    usfm_id       TEXT NOT NULL UNIQUE,         -- 'GEN', '1SA', 'REV'
    name          TEXT NOT NULL,
    testament     TEXT NOT NULL CHECK (testament IN ('OT', 'NT')),
    section       TEXT NOT NULL,
    chapter_count INTEGER NOT NULL
);

-- Per-translation book titles as printed by that publisher (\h, \toc1, \toc3).
CREATE TABLE translation_books (
    translation_id TEXT    NOT NULL REFERENCES translations(translation_id),
    book_id        INTEGER NOT NULL REFERENCES books(book_id),
    title          TEXT,                        -- running head  (\h)
    long_title     TEXT,                        -- full title    (\toc1)
    abbreviation   TEXT,                        -- short title   (\toc3)
    chapter_count  INTEGER NOT NULL,
    verse_count    INTEGER NOT NULL,
    source_file    TEXT NOT NULL,
    PRIMARY KEY (translation_id, book_id)
);

-- -------------------------------------------------------------- verses ----

CREATE TABLE verses (
    id              INTEGER PRIMARY KEY,
    translation_id  TEXT    NOT NULL REFERENCES translations(translation_id),
    book_id         INTEGER NOT NULL REFERENCES books(book_id),
    chapter         INTEGER NOT NULL CHECK (chapter > 0),
    verse           INTEGER NOT NULL CHECK (verse > 0),
    text            TEXT    NOT NULL,            -- '' only where the edition
                                                -- deliberately omits a verse
    paragraph_start INTEGER NOT NULL DEFAULT 0  -- 1 = opens a prose paragraph
);

-- The natural key: the only lookup path the application needs.
CREATE UNIQUE INDEX ux_verses_ref
    ON verses (translation_id, book_id, chapter, verse);

-- Reverse lookup: "give me every translation of Gen 1:1".
CREATE INDEX ix_verses_parallel
    ON verses (book_id, chapter, verse);

-- ------------------------------------------------------------- headings ---

-- Chapter headings, psalm superscriptions, Psalm 119 stanza letters, speaker
-- labels and BSB's Psalter divisions.  A heading belongs to the verse that
-- follows it; a heading after the last verse of a chapter attaches to
-- (last_verse + 1).  Headings are never merged into verse text.
CREATE TABLE sections (
    section_id     INTEGER PRIMARY KEY,
    translation_id TEXT    NOT NULL REFERENCES translations(translation_id),
    book_id        INTEGER NOT NULL REFERENCES books(book_id),
    chapter        INTEGER NOT NULL,
    verse          INTEGER NOT NULL,            -- verse the heading precedes
    ordinal        INTEGER NOT NULL,            -- order within the same verse
    heading_type   TEXT    NOT NULL,            -- section | psalm_title |
                                                -- acrostic | speaker | ...
    level          INTEGER NOT NULL,
    marker         TEXT NOT NULL,               -- the USFM marker, e.g. 'd'
    text           TEXT    NOT NULL
);

CREATE INDEX ix_sections_ref
    ON sections (translation_id, book_id, chapter, verse);

-- ----------------------------------------------------------- full text ----

CREATE VIRTUAL TABLE verses_fts USING fts5(
    text,
    content    = 'verses',
    content_rowid = 'id',
    tokenize   = 'unicode61 remove_diacritics 2'
);

-- ---------------------------------------------------------------- views ---

-- Readable verse view: translation, book, chapter, verse, text.
CREATE VIEW v_verse AS
SELECT t.translation_id,
       t.abbreviation,
       b.book_id,
       b.osis_id,
       b.name      AS book_name,
       v.chapter,
       v.verse,
       v.text,
       v.paragraph_start
FROM verses v
JOIN translations t ON t.translation_id = v.translation_id
JOIN books        b ON b.book_id        = v.book_id;
