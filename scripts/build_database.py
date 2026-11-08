#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""Build the offline SQLite database from the publisher USFM packages.

    python3 scripts/build_database.py [--db PATH] [--raw DIR]

The build is repeatable: it always writes a fresh database from the source
files, so running it twice produces the same result.  Nothing is downloaded
here; run scripts/download_sources.py first.

Pipeline per translation
    1. parse all 66 USFM books              -> verses + headings + \h metadata
    2. apply the documented exporter repairs (sources.TEXT_REPAIRS)
    3. restore verses the edition numbers but leaves blank, using the
       publisher's own plain-text export as the authority (never inventing text)
    4. write canon, provenance, verses, headings and the FTS index

The database is written to a temporary file and moved into place only after
every check inside this script has passed, so a failed build never leaves a
half-written bible.db behind.
"""

import argparse
import hashlib
import json
import os
import sqlite3
import sys
import time
from datetime import datetime, timezone

HERE = os.path.dirname(os.path.abspath(__file__))
ROOT = os.path.dirname(HERE)
sys.path.insert(0, HERE)

import canon                                     # noqa: E402
import reference_compare                          # noqa: E402
import sources                                    # noqa: E402
from usfm_parser import parse_file                # noqa: E402

SCHEMA_VERSION = "1.0"
DEFAULT_DB = os.path.join(ROOT, "data", "db", "bible-en-study.db")
NOTES_TABLE_MISSING = None


# ------------------------------------------------------------------ utils ---

def sha256_of(path):
    h = hashlib.sha256()
    with open(path, "rb") as fh:
        for chunk in iter(lambda: fh.read(1 << 20), b""):
            h.update(chunk)
    return h.hexdigest()


def utc_from_mtime(path):
    return datetime.fromtimestamp(os.path.getmtime(path), timezone.utc)


def iso(dt):
    return dt.strftime("%Y-%m-%dT%H:%M:%SZ")


# -------------------------------------------------------------- extraction --

def load_translation(translation):
    """Parse one translation into in-memory rows."""
    files = sources.source_files(translation)
    missing = [b[2] for b in canon.BOOKS if b[2] not in files]
    if missing:
        raise SystemExit("missing USFM files for %s: %s"
                         % (translation["translation_id"], ", ".join(missing)))

    verses = {}        # (book_id, chapter, verse) -> [text, paragraph_start]
    filled = []        # verse numbers the edition merges, stored empty
    headings = []      # dicts with book_id
    book_meta = {}     # book_id -> {"title", "long_title", "abbrev", "file"}
    unknown = {}
    totals = {"verses": 0, "chapters": 0, "headings": 0, "notes": 0}

    for order, osis, usfm, name, testament, section, chapters in canon.BOOKS:
        path = files[usfm]
        parsed = parse_file(path)
        for marker, n in parsed["unknown_markers"].items():
            unknown[marker] = unknown.get(marker, 0) + n
        meta = parsed["meta"]
        book_meta[order] = {
            "title": meta.get("h"),
            "long_title": meta.get("toc1"),
            "abbrev": meta.get("toc3"),
            "file": os.path.basename(path),
        }
        for (chapter, verse), row in parsed["verses"].items():
            verses[(order, chapter, verse)] = [row["text"],
                                               row["paragraph_start"]]
        # `\v 20-21` is one printed verse covering both numbers; the second
        # number still belongs to the edition's numbering, so it gets an empty
        # row and every chapter stays 1..N.
        for chapter, verse in parsed["merged"]:
            key = (order, chapter, verse)
            if key not in verses:
                verses[key] = ["", 0]
                filled.append("%s %d:%d" % (usfm, chapter, verse))
        for h in parsed["headings"]:
            h["book_id"] = order
            headings.append(h)
        totals["verses"] += len(parsed["verses"])
        totals["chapters"] += parsed["stats"]["chapters"]
        totals["notes"] += parsed["stats"]["notes_removed"]

    return {
        "verses": verses,
        "filled": sorted(filled),
        "headings": headings,
        "book_meta": book_meta,
        "unknown_markers": unknown,
        "totals": totals,
    }


def apply_repairs(data, translation):
    """Apply the documented exporter repairs; each must match exactly once."""
    applied = []
    for repair in sources.TEXT_REPAIRS:
        if repair["translation_id"] != translation["translation_id"]:
            continue
        book = canon.BY_USFM[repair["usfm_id"]]
        key = (book["book_id"], repair["chapter"], repair["verse"])
        if key not in data["verses"]:
            raise SystemExit("repair target %s %s:%s is not in the parsed text"
                             % (repair["usfm_id"], repair["chapter"],
                                repair["verse"]))
        text = data["verses"][key][0]
        count = text.count(repair["find"])
        if count != 1:
            raise SystemExit(
                "repair %s %s:%s expected exactly one occurrence of %r, found %d"
                % (repair["usfm_id"], repair["chapter"], repair["verse"],
                   repair["find"], count))
        data["verses"][key][0] = text.replace(repair["find"], repair["replace"])
        applied.append({
            "ref": "%s %d:%d" % (repair["usfm_id"], repair["chapter"],
                                 repair["verse"]),
            "find": repair["find"],
            "replace": repair["replace"],
            "reason": repair["reason"],
        })
    return applied


def restore_blank_verses(data, translation):
    """Insert empty rows for verses the edition numbers but does not print.

    The BSB USFM simply omits 16 verse numbers that its own bsb.txt lists as
    blank, and the WEB omits five verses that are footnote-only.  The
    publisher's plain-text export decides which references exist; this function
    never supplies text, it only keeps the numbering continuous.
    """
    path = sources.verify_path(translation)
    if not path:
        return []
    if translation["verify_kind"] == "tsv":
        reference = reference_compare.load_bsb_tsv(path)
    else:
        reference = reference_compare.load_vpl(path)

    restored = []
    for (usfm, chapter, verse), text in reference.items():
        if text.strip():
            continue
        # the tab-separated BSB export keys books by OSIS id, the eBible VPL
        # by USFM id; accept either spelling of the same book.
        book = canon.BY_OSIS.get(usfm) or canon.BY_USFM.get(usfm)
        if book is None:
            raise SystemExit("unknown book code %r in %s" % (usfm, path))
        key = (book["book_id"], chapter, verse)
        if key not in data["verses"]:
            data["verses"][key] = ["", 0]
            restored.append("%s %d:%d" % (usfm, chapter, verse))
    return sorted(restored)


# ----------------------------------------------------------------- schema ---

def build_schema(conn):
    with open(os.path.join(HERE, "schema.sql"), "r", encoding="utf-8") as fh:
        conn.executescript(fh.read())


def insert_canon(conn):
    conn.executemany(
        "INSERT INTO books (book_id, osis_id, usfm_id, name, testament, "
        "section, chapter_count) VALUES (?, ?, ?, ?, ?, ?, ?)",
        [(b[0], b[1], b[2], b[3], b[4], b[5], b[6]) for b in canon.BOOKS])


def insert_translations(conn, translation, data, applied, restored, stats):
    archive = sources.archive_path(translation)
    downloaded = utc_from_mtime(archive)
    empty = sum(1 for t, _ in data["verses"].values() if not t.strip())
    chapters = len(set((b, c) for (b, c, v) in data["verses"]))
    books = len(set(b for (b, c, v) in data["verses"]))
    conn.execute(
        "INSERT INTO translations (translation_id, name, abbreviation, "
        "language_code, language_name, edition, copyright, license, "
        "license_url, source_url, download_url, download_date, sha256, "
        "book_count, chapter_count, verse_count, empty_verse_count, notes) "
        "VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
        (translation["translation_id"], translation["name"],
         translation["abbreviation"], translation["language_code"],
         translation["language_name"], translation["edition"],
         translation["copyright"], translation["license"],
         translation["license_url"], translation["source_url"],
         translation["download_url"], iso(downloaded),
         sha256_of(archive), books, chapters, len(data["verses"]), empty,
         translation["notes"]))

    per_book = {}
    for (book_id, chapter, verse), (text, _) in data["verses"].items():
        rec = per_book.setdefault(book_id, {"chapters": set(), "verses": 0,
                                            "empty": 0})
        rec["chapters"].add(chapter)
        rec["verses"] += 1
        if not text.strip():
            rec["empty"] += 1
    rows = []
    for book_id, meta in sorted(data["book_meta"].items()):
        rec = per_book.get(book_id, {"chapters": set(), "verses": 0})
        rows.append((translation["translation_id"], book_id, meta["title"],
                     meta["long_title"], meta["abbrev"], len(rec["chapters"]),
                     rec["verses"], meta["file"]))
    conn.executemany(
        "INSERT INTO translation_books (translation_id, book_id, title, "
        "long_title, abbreviation, chapter_count, verse_count, source_file) "
        "VALUES (?, ?, ?, ?, ?, ?, ?, ?)", rows)
    stats["empty_verse_count"] = empty
    stats["book_count"] = books
    stats["chapter_count"] = chapters


def insert_verses(conn, translation, data):
    rows = [(translation["translation_id"], book_id, chapter, verse,
             text, int(para))
            for (book_id, chapter, verse), (text, para)
            in sorted(data["verses"].items())]
    conn.executemany(
        "INSERT INTO verses (translation_id, book_id, chapter, verse, text, "
        "paragraph_start) VALUES (?, ?, ?, ?, ?, ?)", rows)
    return len(rows)


def insert_sections(conn, translation, data):
    rows = []
    seen = {}
    for h in data["headings"]:
        key = (h["book_id"], h["chapter"], h["verse"])
        ordinal = seen.get(key, 0)
        seen[key] = ordinal + 1
        rows.append((translation["translation_id"], h["book_id"],
                     h["chapter"], h["verse"], ordinal, h["kind"],
                     h.get("level", 1), h.get("marker", ""), h["text"]))
    rows.sort(key=lambda r: (r[1], r[2], r[3], r[4]))
    conn.executemany(
        "INSERT INTO sections (translation_id, book_id, chapter, verse, "
        "ordinal, heading_type, level, marker, text) "
        "VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)", rows)
    return len(rows)


def create_parallel_view(conn):
    """Pivot view: one row per verse reference, one column per translation."""
    ids = [r[0] for r in conn.execute(
        "SELECT translation_id FROM translations ORDER BY translation_id")]
    cols = ",\n       ".join(
        "MAX(CASE WHEN v.translation_id = '%s' THEN v.text END) AS %s_text"
        % (tid, tid.lower()) for tid in ids)
    sql = (
        "CREATE VIEW v_verse_parallel AS\n"
        "SELECT b.book_id,\n"
        "       b.osis_id,\n"
        "       b.name AS book_name,\n"
        "       v.chapter,\n"
        "       v.verse,\n"
        "       %s,\n"
        "       COUNT(*) AS translation_count\n"
        "FROM verses v\n"
        "JOIN books b ON b.book_id = v.book_id\n"
        "GROUP BY v.book_id, v.chapter, v.verse" % cols)
    conn.execute(sql)


# ------------------------------------------------------------------- main ---

def main(argv=None):
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument("--db", default=DEFAULT_DB, help="output database path")
    ap.add_argument("--quiet", action="store_true")
    args = ap.parse_args(argv)

    db_path = os.path.abspath(args.db)
    db_dir = os.path.dirname(db_path)
    os.makedirs(db_dir, exist_ok=True)
    if not db_path.endswith(".db"):
        raise SystemExit("refusing to overwrite %r: database path must end in "
                         ".db" % db_path)
    tmp_path = db_path + ".building"

    started = time.time()
    canon.validate_canon()

    if os.path.exists(tmp_path):
        os.unlink(tmp_path)
    conn = sqlite3.connect(tmp_path)
    conn.execute("PRAGMA journal_mode = OFF")
    conn.execute("PRAGMA synchronous = OFF")
    conn.execute("PRAGMA page_size = 4096")
    build_schema(conn)

    report = {
        "schema_version": SCHEMA_VERSION,
        "built_at": iso(datetime.now(timezone.utc)),
        "database": os.path.relpath(db_path, ROOT),
        "translations": {},
        "problems": [],
    }

    conn.execute("BEGIN")
    insert_canon(conn)
    for translation in sources.TRANSLATIONS:
        tid = translation["translation_id"]
        data = load_translation(translation)
        applied = apply_repairs(data, translation)
        restored = restore_blank_verses(data, translation)
        stats = {}
        insert_translations(conn, translation, data, applied, restored, stats)
        n_verses = insert_verses(conn, translation, data)
        n_sections = insert_sections(conn, translation, data)

        if data["unknown_markers"]:
            report["problems"].append(
                "%s: unknown USFM markers %r" % (tid, data["unknown_markers"]))
        report["translations"][tid] = {
            "books": stats["book_count"],
            "chapters": stats["chapter_count"],
            "verses": n_verses,
            "empty_verses": stats["empty_verse_count"],
            "sections": n_sections,
            "paragraph_starts": sum(1 for t, p in data["verses"].values()
                                    if p),
            "repairs": applied,
            "restored_blank_verses": restored,
            "merged_blank_verses": data["filled"],
            "unknown_markers": data["unknown_markers"],
        }
        if not args.quiet:
            print("%-5s books=%-3d chapters=%-5d verses=%-6d sections=%-5d "
                  "empty=%-3d" % (tid, stats["book_count"],
                                  stats["chapter_count"], n_verses,
                                  n_sections, stats["empty_verse_count"]))
        del data

    conn.execute("INSERT INTO verses_fts (verses_fts) VALUES ('rebuild')")
    create_parallel_view(conn)
    conn.execute(
        "INSERT INTO dataset_meta (key, value) VALUES "
        "('schema_version', ?)", (SCHEMA_VERSION,))
    conn.execute(
        "INSERT INTO dataset_meta (key, value) VALUES ('built_at', ?)",
        (report["built_at"],))
    conn.execute(
        "INSERT INTO dataset_meta (key, value) VALUES ('generator', ?)",
        ("scripts/build_database.py",))
    conn.execute(
        "INSERT INTO dataset_meta (key, value) VALUES ('canon', ?)",
        ("66 books, %d chapters" % sum(b[6] for b in canon.BOOKS),))
    conn.execute(
        "INSERT INTO dataset_meta (key, value) VALUES ('translation_ids', ?)",
        (",".join(sources.TRANSLATION_IDS),))
    conn.execute(
        "INSERT INTO dataset_meta (key, value) VALUES ('verse_identity', ?)",
        ("(translation_id, book_id, chapter, verse); never a line number",))
    conn.commit()

    if report["problems"]:
        conn.close()
        os.unlink(tmp_path)
        raise SystemExit("build failed:\n  " + "\n  ".join(report["problems"]))

    conn.execute("ANALYZE")
    conn.execute("VACUUM")
    integrity = conn.execute("PRAGMA integrity_check").fetchone()[0]
    fk = conn.execute("PRAGMA foreign_key_check").fetchall()
    conn.close()
    if integrity != "ok" or fk:
        raise SystemExit("integrity check failed: %r %r" % (integrity, fk))

    if os.path.exists(db_path):
        os.unlink(db_path)
    os.replace(tmp_path, db_path)

    report["db_bytes"] = os.path.getsize(db_path)
    report["elapsed_seconds"] = round(time.time() - started, 1)
    report_path = os.path.join(ROOT, "data", "reports", "build_summary.json")
    os.makedirs(os.path.dirname(report_path), exist_ok=True)
    with open(report_path, "w", encoding="utf-8") as fh:
        json.dump(report, fh, ensure_ascii=False, indent=2)
        fh.write("\n")

    if not args.quiet:
        print("\nwrote %s (%.1f MiB) in %.1fs"
              % (os.path.relpath(db_path, ROOT),
                 report["db_bytes"] / 1048576.0, report["elapsed_seconds"]))
        print("summary: %s" % os.path.relpath(report_path, ROOT))
    return 0


if __name__ == "__main__":
    sys.exit(main())
