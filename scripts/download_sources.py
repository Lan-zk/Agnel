#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""Download the four source packages, verify them and unpack them.

    python3 scripts/download_sources.py [--force] [--proxy HOST:PORT]

What it fetches (all from the publisher, no third-party mirrors)
    bereanbible.com/bsb_usfm.zip          USFM  (BSB)
    bereanbible.com/bsb.txt               plain text, used as a cross-check
    eBible.org/Scriptures/engwebp_usfm.zip       USFM  (WEBP)
    eBible.org/Scriptures/engwebp_vpl.zip        plain text, cross-check
    eBible.org/Scriptures/eng-kjv2006_usfm.zip   USFM  (KJV)
    eBible.org/Scriptures/eng-kjv2006_vpl.zip    plain text, cross-check
    eBible.org/Scriptures/cmn-cu89s_usfm.zip     USFM  (CUV, 和合本)
    eBible.org/Scriptures/cmn-cu89s_vpl.zip      plain text, cross-check

Files are written to data/raw/ and recorded in data/raw/download_manifest.json
with url, byte size, sha256 and the time they were fetched.  A second run
re-uses a file whose sha256 still matches the manifest, so the whole pipeline
stays reproducible without hitting the network again.  Pass --force to
re-download.

The download is the only step that needs the network; build_database.py and
verify_database.py run entirely offline.
"""

import argparse
import hashlib
import json
import os
import shutil
import subprocess
import sys
import time
import zipfile
from datetime import datetime, timezone

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))

import sources                                    # noqa: E402

HERE = os.path.dirname(os.path.abspath(__file__))
ROOT = os.path.dirname(HERE)
RAW = os.path.join(ROOT, "data", "raw")
MANIFEST = os.path.join(RAW, "download_manifest.json")

# USFM package -> translation_id used by sources.py / the database
USFM_PACKAGES = {"bsb_usfm": "BSB", "webp_usfm": "WEBP", "kjv_usfm": "KJV",
                 "cuv_usfm": "CUV"}

# (key, url, file name, unpack into, source page the licence was read from)
PACKAGES = [
    ("bsb_usfm", "https://bereanbible.com/bsb_usfm.zip",
     "bsb_usfm.zip", "bsb", "https://berean.bible/downloads.htm"),
    ("bsb_text", "https://bereanbible.com/bsb.txt",
     "bsb.txt", None, "https://berean.bible/downloads.htm"),
    ("webp_usfm", "https://eBible.org/Scriptures/engwebp_usfm.zip",
     "engwebp_usfm.zip", "webp",
     "https://ebible.org/find/details.php?id=engwebp"),
    ("webp_vpl", "https://eBible.org/Scriptures/engwebp_vpl.zip",
     "engwebp_vpl.zip", "verify",
     "https://ebible.org/find/details.php?id=engwebp"),
    ("kjv_usfm", "https://eBible.org/Scriptures/eng-kjv2006_usfm.zip",
     "eng-kjv2006_usfm.zip", "kjv",
     "https://ebible.org/find/details.php?id=engkjv"),
    ("kjv_vpl", "https://eBible.org/Scriptures/eng-kjv2006_vpl.zip",
     "eng-kjv2006_vpl.zip", "verify",
     "https://ebible.org/find/details.php?id=engkjv"),
    ("cuv_usfm", "https://eBible.org/Scriptures/cmn-cu89s_usfm.zip",
     "cmn-cu89s_usfm.zip", "cuv",
     "https://ebible.org/find/details.php?id=cmn-cu89s"),
    ("cuv_vpl", "https://eBible.org/Scriptures/cmn-cu89s_vpl.zip",
     "cmn-cu89s_vpl.zip", "verify",
     "https://ebible.org/find/details.php?id=cmn-cu89s"),
]


def sha256_of(path):
    h = hashlib.sha256()
    with open(path, "rb") as fh:
        for chunk in iter(lambda: fh.read(1 << 20), b""):
            h.update(chunk)
    return h.hexdigest()


def load_manifest():
    if not os.path.exists(MANIFEST):
        return {}
    with open(MANIFEST, "r", encoding="utf-8") as fh:
        return json.load(fh)


def save_manifest(manifest):
    with open(MANIFEST, "w", encoding="utf-8") as fh:
        json.dump(manifest, fh, ensure_ascii=False, indent=2, sort_keys=True)
        fh.write("\n")


def fetch(url, target, proxy=None, min_members=4):
    """curl with retries and resume into `<target>.part`, then verify.

    Writing to a separate .part file matters: `curl -C -` appends, and a
    server that answers a complete file with a short error body would otherwise
    silently glue that body onto a good archive.
    """
    os.makedirs(os.path.dirname(target), exist_ok=True)
    part = target + ".part"
    if os.path.exists(part):
        try:
            check_download(part, target, min_members)
            print("  complete %s.part already present, no download needed"
                  % os.path.basename(target))
            os.replace(part, target)
            return
        except SystemExit:
            pass                      # partial file: resume it below
    cmd = ["curl", "-q", "-L", "--fail", "--retry", "5", "--retry-all-errors",
           "--retry-delay", "2", "-C", "-", "-o", part, url]
    if proxy:
        cmd[4:4] = ["--proxy", proxy]
    print("  GET %s" % url)
    result = subprocess.run(cmd)
    if result.returncode != 0:
        raise SystemExit("download failed (%d): %s" % (result.returncode, url))
    check_download(part, target, min_members)
    os.replace(part, target)


def check_download(part, target, min_members=4):
    """Reject a short/HTML/truncated body before it can replace a good file."""
    size = os.path.getsize(part)
    if size < 100000:
        raise SystemExit("%s: only %d bytes downloaded, looks like an error page"
                         % (target, size))
    if target.endswith(".zip"):
        with zipfile.ZipFile(part) as zf:
            bad = zf.testzip()
            if bad is not None:
                raise SystemExit("%s: corrupt archive member %r" % (target, bad))
            if len(zf.namelist()) < min_members:
                raise SystemExit("%s: only %d members, expected at least %d"
                                 % (target, len(zf.namelist()), min_members))
    else:
        with open(part, "rb") as fh:
            head = fh.read(4096).decode("utf-8-sig", errors="replace")
        if "\ufffd" in head or "\t" not in head:
            raise SystemExit("%s: not the expected tab separated text" % target)


def unpack(archive, subdir):
    directory = os.path.join(RAW, subdir)
    os.makedirs(directory, exist_ok=True)
    with zipfile.ZipFile(archive) as zf:
        bad = zf.testzip()
        if bad is not None:
            raise SystemExit("corrupt archive member %r in %s" % (bad, archive))
        zf.extractall(directory)
    return directory


def count_usfm(directory):
    n = 0
    for root, _dirs, files in os.walk(directory):
        n += sum(1 for f in files if f.endswith(".usfm"))
    return n


def main(argv=None):
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument("--force", action="store_true",
                    help="download again even if the manifest hash matches")
    ap.add_argument("--proxy", help="HTTP proxy, e.g. 127.0.0.1:7897")
    args = ap.parse_args(argv)

    proxy = args.proxy or os.environ.get("HTTPS_PROXY") or None
    os.makedirs(RAW, exist_ok=True)
    manifest = load_manifest()

    for key, url, name, subdir, source_page in PACKAGES:
        target = os.path.join(RAW, name)
        entry = manifest.get(key, {})
        if (not args.force and os.path.exists(target)
                and entry.get("sha256") == sha256_of(target)):
            print("%-10s cached  %s (%s bytes)" % (key, name,
                                                   "{:,}".format(
                                                       os.path.getsize(target))))
            size = os.path.getsize(target)
        else:
            print("%-10s fetch   %s" % (key, name))
            fetch(url, target, proxy,
                  min_members=60 if key in USFM_PACKAGES else 4)
            size = os.path.getsize(target)
            manifest[key] = {
                "url": url,
                "file": name,
                "bytes": size,
                "sha256": sha256_of(target),
                "source_page": source_page,
                "fetched_at": datetime.now(timezone.utc).strftime(
                    "%Y-%m-%dT%H:%M:%SZ"),
            }
            save_manifest(manifest)
            time.sleep(1)

        if subdir and name.endswith(".zip"):
            directory = unpack(target, subdir)
            if name.endswith("usfm.zip"):
                print("           unpacked to data/raw/%s (%d .usfm files)"
                      % (subdir, count_usfm(directory)))
                # the packages also carry front matter and a glossary, so
                # check that all 66 canonical books resolve, not the raw count
                found = sources.source_files(
                    sources.BY_ID[USFM_PACKAGES[key]])
                if len(found) != 66:
                    raise SystemExit("%s: only %d of the 66 canonical books "
                                     "found in data/raw/%s"
                                     % (name, len(found), subdir))
                print("           all 66 canonical books located")
            else:
                print("           plain-text cross-check unpacked")
        elif subdir:
            os.makedirs(os.path.join(RAW, subdir), exist_ok=True)

    # the two plain-text cross-check files must be where sources.py expects them
    expected = [
        ("bsb.txt", os.path.join(RAW, "bsb.txt")),
        ("engwebp_vpl.txt", os.path.join(RAW, "verify", "engwebp_vpl.txt")),
        ("eng-kjv2006_vpl.txt", os.path.join(RAW, "verify",
                                             "eng-kjv2006_vpl.txt")),
        ("cmn-cu89s_vpl.txt", os.path.join(RAW, "verify",
                                           "cmn-cu89s_vpl.txt")),
    ]
    for label, path in expected:
        if not os.path.exists(path):
            raise SystemExit("missing %s after download (%s)" % (label, path))

    print("\nmanifest: data/raw/download_manifest.json")
    print("next:     python3 scripts/build_database.py")
    return 0


if __name__ == "__main__":
    sys.exit(main())
