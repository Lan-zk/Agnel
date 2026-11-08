#!/usr/bin/env python3
"""Build the embedded bilingual reading fonts for bible-en-study.

Uses only the Python standard library (struct, sqlite3, zlib, json, glob).
Combines three SIL Open Font License 1.1 typefaces at unitsPerEm = 1000:
  1. Literata (TypeTogether / Google Play Books) — primary Latin, Greek,
     Cyrillic, numerals, and Western typographic punctuation, preserving
     original glyph IDs 0..N_base-1 so GPOS kerning and GSUB ligatures
     remain intact.
  2. STIX Two Text (STI Pub) — the 10 IPA phonetic symbols missing from
     Literata (ɔ, ɚ, ɛ, ɝ, ɪ, ɹ, ʃ, ʊ, ʌ, ʒ) for dictionary transcriptions.
  3. Source Han Serif SC / 思源宋体 (Adobe & Google) — subsetted to the CJK
     ideographs and CJK punctuation used across the 4-translation Bible
     (store.bin), 41,873-word dictionary (dict.bin), and UI source code,
     plus common simplified Chinese characters, optically scaled (0.92x) and
     baseline-balanced (-35 units) to harmonize with Literata.
"""

from __future__ import annotations

import glob
import json
import os
import sqlite3
import struct
import urllib.request
import zlib

ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
OUT_DIR = os.path.join(ROOT, "app", "internal", "font")
CACHE_DIR = os.path.join(ROOT, "data", "raw", "fonts")

FONT_URLS = {
    "Literata-Regular.ttf": "https://raw.githubusercontent.com/googlefonts/literata/main/fonts/ttf/Literata-Regular.ttf",
    "Literata-SemiBold.ttf": "https://raw.githubusercontent.com/googlefonts/literata/main/fonts/ttf/Literata-SemiBold.ttf",
    "Literata-Italic.ttf": "https://raw.githubusercontent.com/googlefonts/literata/main/fonts/ttf/Literata-Italic.ttf",
    "SourceHanSerifSC-Regular.ttf": "https://raw.githubusercontent.com/echcz/FontsTTF/master/SourceHanSerif/SourceHanSerifSC-Regular.ttf",
    "SourceHanSerifSC-SemiBold.ttf": "https://raw.githubusercontent.com/echcz/FontsTTF/master/SourceHanSerif/SourceHanSerifSC-SemiBold.ttf",
}

IPA_CHARS = [ord(c) for c in ["ɔ", "ɚ", "ɛ", "ɝ", "ɪ", "ɹ", "ʃ", "ʊ", "ʌ", "ʒ"]]

WESTERN_PUNCT_KEEP = {
    0x2010, 0x2011, 0x2012, 0x2013, 0x2014, 0x2015,
    0x2018, 0x2019, 0x201A, 0x201C, 0x201D, 0x201E,
    0x2020, 0x2021, 0x2022, 0x2026, 0x2030, 0x2032, 0x2033,
    0x2039, 0x203A, 0x20AC, 0x2122,
}


def ensure_raw_font(name: str) -> str:
    tmp_candidate = os.path.join("/tmp/fonts", name)
    if os.path.exists(tmp_candidate):
        return tmp_candidate
    os.makedirs(CACHE_DIR, exist_ok=True)
    dest = os.path.join(CACHE_DIR, name)
    if not os.path.exists(dest):
        url = FONT_URLS[name]
        print(f"Downloading {name} ...")
        urllib.request.urlretrieve(url, dest)
    return dest


def parse_ttf(path: str, ttc_index: int = 0) -> dict[str, bytes]:
    with open(path, "rb") as f:
        data = f.read()
    off = struct.unpack(">I", data[12 + 4 * ttc_index : 16 + 4 * ttc_index])[0] if data[:4] == b"ttcf" else 0
    ntables = struct.unpack(">H", data[off + 4 : off + 6])[0]
    tables: dict[str, bytes] = {}
    for t in range(ntables):
        e = off + 12 + 16 * t
        tag = data[e : e + 4].decode("latin1")
        toff, tlen = struct.unpack(">II", data[e + 8 : e + 16])
        tables[tag] = data[toff : toff + tlen]
    return tables


def parse_cmap(cmap_b: bytes) -> dict[int, int]:
    nsub = struct.unpack(">H", cmap_b[2:4])[0]
    res: dict[int, int] = {}
    for s in range(nsub):
        _, _, soff = struct.unpack(">HHI", cmap_b[4 + 8 * s : 12 + 8 * s])
        fmt = struct.unpack(">H", cmap_b[soff : soff + 2])[0]
        if fmt == 12:
            ng = struct.unpack(">I", cmap_b[soff + 12 : soff + 16])[0]
            for g in range(ng):
                sc, ec, sg = struct.unpack(">III", cmap_b[soff + 16 + 12 * g : soff + 28 + 12 * g])
                for c in range(sc, ec + 1):
                    res[c] = sg + (c - sc)
            return res
        if fmt == 4 and not res:
            seg_count = struct.unpack(">H", cmap_b[soff + 6 : soff + 8])[0] // 2
            end_codes = struct.unpack(f">{seg_count}H", cmap_b[soff + 14 : soff + 14 + 2 * seg_count])
            start_codes = struct.unpack(f">{seg_count}H", cmap_b[soff + 16 + 2 * seg_count : soff + 16 + 4 * seg_count])
            id_deltas = struct.unpack(f">{seg_count}h", cmap_b[soff + 16 + 4 * seg_count : soff + 16 + 6 * seg_count])
            id_range_pos = soff + 16 + 6 * seg_count
            id_range_offs = struct.unpack(f">{seg_count}H", cmap_b[id_range_pos : id_range_pos + 2 * seg_count])
            for i in range(seg_count):
                for c in range(start_codes[i], end_codes[i] + 1):
                    if c == 0xFFFF:
                        continue
                    if id_range_offs[i] == 0:
                        gid = (c + id_deltas[i]) & 0xFFFF
                    else:
                        roff = id_range_pos + 2 * i + id_range_offs[i] + 2 * (c - start_codes[i])
                        gid = struct.unpack(">H", cmap_b[roff : roff + 2])[0]
                        if gid != 0:
                            gid = (gid + id_deltas[i]) & 0xFFFF
                    if gid != 0:
                        res[c] = gid
    return res


def get_glyph_metrics(tables: dict[str, bytes]) -> tuple[list[int], bytes, list[tuple[int, int]]]:
    n_glyphs = struct.unpack(">H", tables["maxp"][4:6])[0]
    n_hmetrics = struct.unpack(">H", tables["hhea"][34:36])[0]
    loc_fmt = struct.unpack(">h", tables["head"][50:52])[0]
    if loc_fmt == 1:
        loca = list(struct.unpack(f">{n_glyphs + 1}I", tables["loca"][: 4 * (n_glyphs + 1)]))
    else:
        loca = [x * 2 for x in struct.unpack(f">{n_glyphs + 1}H", tables["loca"][: 2 * (n_glyphs + 1)])]
    hmtx_b = tables["hmtx"]
    hmtx: list[tuple[int, int]] = []
    last_aw = 0
    for g in range(n_glyphs):
        if g < n_hmetrics:
            aw, lsb = struct.unpack(">Hh", hmtx_b[4 * g : 4 * g + 4])
            last_aw = aw
        else:
            off = 4 * n_hmetrics + 2 * (g - n_hmetrics)
            lsb = struct.unpack(">h", hmtx_b[off : off + 2])[0]
            aw = last_aw
        hmtx.append((aw, lsb))
    return loca, tables["glyf"], hmtx


def transform_simple_glyph(gdata: bytes, scale: float = 0.92, dy: int = -35) -> tuple[bytes, int | None]:
    if not gdata:
        return b"", None
    n_contours, x_min, y_min, x_max, y_max = struct.unpack(">5h", gdata[:10])
    if n_contours <= 0:
        return gdata, None
    pos = 10
    end_pts = struct.unpack(f">{n_contours}H", gdata[pos : pos + 2 * n_contours])
    pos += 2 * n_contours
    n_pts = end_pts[-1] + 1
    instr_len = struct.unpack(">H", gdata[pos : pos + 2])[0]
    pos += 2 + instr_len

    flags: list[int] = []
    while len(flags) < n_pts:
        fl = gdata[pos]
        pos += 1
        flags.append(fl)
        if fl & 0x08:
            rep = gdata[pos]
            pos += 1
            flags.extend([fl] * rep)

    xs: list[int] = []
    cur_x = 0
    for fl in flags:
        if fl & 0x02:
            dx = gdata[pos]
            pos += 1
            cur_x += dx if (fl & 0x10) else -dx
        else:
            if not (fl & 0x10):
                dx = struct.unpack(">h", gdata[pos : pos + 2])[0]
                pos += 2
                cur_x += dx
        xs.append(cur_x)

    ys: list[int] = []
    cur_y = 0
    for fl in flags:
        if fl & 0x04:
            d = gdata[pos]
            pos += 1
            cur_y += d if (fl & 0x20) else -d
        else:
            if not (fl & 0x20):
                d = struct.unpack(">h", gdata[pos : pos + 2])[0]
                pos += 2
                cur_y += d
        ys.append(cur_y)

    cx = 500.0
    cy = 380.0
    new_xs = [int(round(cx + (x - cx) * scale)) for x in xs]
    new_ys = [int(round(cy + (y - cy) * scale + dy)) for y in ys]
    nx_min, nx_max = min(new_xs), max(new_xs)
    ny_min, ny_max = min(new_ys), max(new_ys)

    out_flags = bytearray()
    out_x = bytearray()
    out_y = bytearray()
    px, py = 0, 0
    for i in range(n_pts):
        fl = flags[i] & 0x01
        dx = new_xs[i] - px
        dy_pt = new_ys[i] - py
        px, py = new_xs[i], new_ys[i]
        if dx == 0:
            fl |= 0x10
        elif -255 <= dx <= 255:
            fl |= 0x02 | (0x10 if dx > 0 else 0)
            out_x.append(abs(dx))
        else:
            out_x.extend(struct.pack(">h", dx))

        if dy_pt == 0:
            fl |= 0x20
        elif -255 <= dy_pt <= 255:
            fl |= 0x04 | (0x20 if dy_pt > 0 else 0)
            out_y.append(abs(dy_pt))
        else:
            out_y.extend(struct.pack(">h", dy_pt))
        out_flags.append(fl)

    header = struct.pack(">5h", n_contours, nx_min, ny_min, nx_max, ny_max)
    end_pts_b = struct.pack(f">{n_contours}H", *end_pts)
    return header + end_pts_b + struct.pack(">H", 0) + bytes(out_flags) + bytes(out_x) + bytes(out_y), nx_min


def build_cmap_format4_and_12(mapping: dict[int, int]) -> bytes:
    codes = sorted(c for c, g in mapping.items() if g > 0)
    groups: list[list[int]] = []
    for c in codes:
        g = mapping[c]
        if groups and c == groups[-1][1] + 1 and g == groups[-1][2] + (c - groups[-1][0]):
            groups[-1][1] = c
        else:
            groups.append([c, c, g])

    sub12 = bytearray()
    sub12.extend(struct.pack(">HHIII", 12, 0, 16 + 12 * len(groups), 0, len(groups)))
    for sc, ec, sg in groups:
        sub12.extend(struct.pack(">III", sc, ec, sg))

    bmp_groups = [grp for grp in groups if grp[1] < 0xFFFF and grp[2] + (grp[1] - grp[0]) < 0xFFFF]
    bmp_groups.append([0xFFFF, 0xFFFF, 0])
    seg_count = len(bmp_groups)
    search_range = 2 * (1 << (seg_count.bit_length() - 1))
    entry_selector = seg_count.bit_length() - 1
    range_shift = 2 * seg_count - search_range

    end_codes = [grp[1] for grp in bmp_groups]
    start_codes = [grp[0] for grp in bmp_groups]
    id_deltas = [0 if grp[0] == 0xFFFF else ((grp[2] - grp[0]) & 0xFFFF) for grp in bmp_groups]
    id_range_offs = [0] * seg_count

    sub4 = bytearray()
    sub4.extend(struct.pack(">8H", 4, 16 + 8 * seg_count, 0, seg_count * 2, search_range, entry_selector, range_shift, *end_codes[:1]))
    sub4 = bytearray(struct.pack(">7H", 4, 16 + 8 * seg_count, 0, seg_count * 2, search_range, entry_selector, range_shift))
    sub4.extend(struct.pack(f">{seg_count}H", *end_codes))
    sub4.extend(struct.pack(">H", 0))
    sub4.extend(struct.pack(f">{seg_count}H", *start_codes))
    sub4.extend(struct.pack(f">{seg_count}H", *id_deltas))
    sub4.extend(struct.pack(f">{seg_count}H", *id_range_offs))

    header = bytearray(struct.pack(">HH", 0, 4))
    off4 = 4 + 4 * 8
    off12 = off4 + len(sub4)
    header.extend(struct.pack(">HHI", 0, 3, off4))
    header.extend(struct.pack(">HHI", 0, 4, off12))
    header.extend(struct.pack(">HHI", 3, 1, off4))
    header.extend(struct.pack(">HHI", 3, 10, off12))
    return bytes(header + sub4 + sub12)


def build_name_table(family: str, subfam: str, ps_name: str) -> bytes:
    full_name = f"{family} {subfam}" if subfam != "Regular" else family
    recs = [
        (0, "Copyright 2026 TypeTogether (Literata), Adobe & Google (Source Han Serif), STI Pub (STIX Two). Licensed under SIL Open Font License 1.1."),
        (1, family),
        (2, subfam),
        (3, ps_name),
        (4, full_name),
        (5, "Version 1.0"),
        (6, ps_name),
        (16, family),
        (17, subfam),
    ]
    entries = []
    pool = bytearray()
    for pid, eid, lid, enc in [(1, 0, 0, "mac_roman"), (3, 1, 0x0409, "utf-16-be")]:
        for nid, val in recs:
            raw = val.encode(enc, errors="ignore")
            entries.append((pid, eid, lid, nid, len(raw), len(pool)))
            pool.extend(raw)
    entries.sort()
    hdr = bytearray(struct.pack(">HHH", 0, len(entries), 6 + 12 * len(entries)))
    for pid, eid, lid, nid, slen, soff in entries:
        hdr.extend(struct.pack(">6H", pid, eid, lid, nid, slen, soff))
    return bytes(hdr + pool)


def assemble_ttf(tables: dict[str, bytes]) -> bytes:
    tags = sorted(tables.keys())
    ntables = len(tags)
    p2 = 1 << (ntables.bit_length() - 1)
    search_range = p2 * 16
    entry_selector = ntables.bit_length() - 1
    range_shift = ntables * 16 - search_range

    out = bytearray(struct.pack(">IHHHH", 0x00010000, ntables, search_range, entry_selector, range_shift))
    dir_pos = 12
    out.extend(b"\x00" * (16 * ntables))

    for tag in tags:
        raw = tables[tag]
        if tag == "head":
            raw = raw[:8] + b"\x00\x00\x00\x00" + raw[12:]
        toff = len(out)
        tlen = len(raw)
        out.extend(raw)
        while len(out) % 4 != 0:
            out.append(0)
        padded = out[toff : len(out)]
        csum = sum(struct.unpack(f">{len(padded)//4}I", padded)) & 0xFFFFFFFF
        struct.pack_into(">4sIII", out, dir_pos, tag.encode("latin1"), csum, toff, tlen)
        dir_pos += 16

    return bytes(out)


def collect_needed_cjk_chars() -> list[int]:
    db_path = os.path.join(ROOT, "data", "db", "bible-en-study.db")
    conn = sqlite3.connect(db_path)
    chars: set[str] = set()
    for (t,) in conn.execute("SELECT text FROM verses"):
        if t:
            chars.update(t)
    for row in conn.execute("SELECT * FROM sections"):
        for v in row:
            if isinstance(v, str):
                chars.update(v)
    for row in conn.execute("SELECT * FROM books"):
        for v in row:
            if isinstance(v, str):
                chars.update(v)
    for row in conn.execute("SELECT * FROM translation_books"):
        for v in row:
            if isinstance(v, str):
                chars.update(v)
    for p in glob.glob(os.path.join(ROOT, "app", "**", "*.go"), recursive=True):
        with open(p, "r", encoding="utf-8") as f:
            chars.update(f.read())
    dict_path = os.path.join(ROOT, "app", "internal", "dict", "dict.bin")
    with open(dict_path, "rb") as f:
        blob = f.read()
    n = struct.unpack("<I", blob[6:10])[0]
    raw = blob[10 : 10 + n]
    if blob[:6] == b"LEXI2\n":
        raw = zlib.decompress(raw)
    idx = json.loads(raw)
    body = blob[10 + n :]
    for off, length in idx["entries"].values():
        rec = zlib.decompress(body[off : off + length]).decode("utf-8")
        chars.update(rec)
    extra = "「」『』，。；：？！、（）《》〈〉——……·“”‘’←→↑↓✓✕▾▴‹›⇄第页共章节卷旧新约检索目录返阅读已书签形收录"
    chars.update(extra)
    return sorted(ord(c) for c in chars if ord(c) >= 0x2000 and ord(c) not in WESTERN_PUNCT_KEEP)


def build_hybrid_face(
    lit_path: str,
    stix_path: str,
    shs_path: str | None,
    cjk_chars: list[int],
    family: str,
    subfam: str,
    ps_name: str,
    weight_class: int,
    is_bold: bool,
    is_italic: bool,
) -> bytes:
    lit = parse_ttf(lit_path)
    lit_cmap = parse_cmap(lit["cmap"])
    lit_loca, lit_glyf, lit_hmtx = get_glyph_metrics(lit)

    stix = parse_ttf(stix_path)
    stix_cmap = parse_cmap(stix["cmap"])
    stix_loca, stix_glyf, stix_hmtx = get_glyph_metrics(stix)

    new_glyf = bytearray(lit_glyf)
    while len(new_glyf) % 4 != 0:
        new_glyf.append(0)
    new_loca = list(lit_loca)
    new_loca[-1] = len(new_glyf)
    new_hmtx = list(lit_hmtx)
    new_cmap = dict(lit_cmap)

    for c in IPA_CHARS:
        if c not in new_cmap and c in stix_cmap:
            sgid = stix_cmap[c]
            gd = stix_glyf[stix_loca[sgid] : stix_loca[sgid + 1]]
            aw, lsb = stix_hmtx[sgid]
            ngid = len(new_hmtx)
            new_glyf.extend(gd)
            while len(new_glyf) % 4 != 0:
                new_glyf.append(0)
            new_loca.append(len(new_glyf))
            new_hmtx.append((aw, lsb))
            new_cmap[c] = ngid

    if shs_path is not None:
        shs = parse_ttf(shs_path)
        shs_cmap = parse_cmap(shs["cmap"])
        shs_loca, shs_glyf, shs_hmtx = get_glyph_metrics(shs)

        for c in list(new_cmap.keys()):
            if c >= 0x2000 and c not in WESTERN_PUNCT_KEEP:
                del new_cmap[c]

        shs_added: dict[int, int] = {}
        for c in cjk_chars:
            sgid = shs_cmap.get(c, 0)
            if not sgid:
                continue
            if sgid not in shs_added:
                gd = shs_glyf[shs_loca[sgid] : shs_loca[sgid + 1]]
                tgd, nxmin = transform_simple_glyph(gd, scale=0.92, dy=-35)
                aw, lsb = shs_hmtx[sgid]
                if nxmin is not None:
                    lsb = nxmin
                ngid = len(new_hmtx)
                new_glyf.extend(tgd)
                while len(new_glyf) % 4 != 0:
                    new_glyf.append(0)
                new_loca.append(len(new_glyf))
                new_hmtx.append((aw, lsb))
                shs_added[sgid] = ngid
            new_cmap[c] = shs_added[sgid]

    total_glyphs = len(new_hmtx)
    tables = dict(lit)
    tables["glyf"] = bytes(new_glyf)
    tables["loca"] = struct.pack(f">{len(new_loca)}I", *new_loca)
    tables["hmtx"] = b"".join(struct.pack(">Hh", aw, lsb) for aw, lsb in new_hmtx)

    maxp = bytearray(tables["maxp"])
    struct.pack_into(">H", maxp, 4, total_glyphs)
    tables["maxp"] = bytes(maxp)

    hhea = bytearray(tables["hhea"])
    struct.pack_into(">h", hhea, 4, 920)
    struct.pack_into(">h", hhea, 6, -240)
    struct.pack_into(">h", hhea, 8, 0)
    struct.pack_into(">H", hhea, 34, total_glyphs)
    tables["hhea"] = bytes(hhea)

    head = bytearray(tables["head"])
    mac_style = (1 if is_bold else 0) | (2 if is_italic else 0)
    struct.pack_into(">H", head, 44, mac_style)
    struct.pack_into(">h", head, 50, 1)
    tables["head"] = bytes(head)

    os2 = bytearray(tables["OS/2"])
    struct.pack_into(">H", os2, 4, weight_class)
    fs_sel = 0x0080
    if is_italic:
        fs_sel |= 0x0001
    elif is_bold:
        fs_sel |= 0x0020
    else:
        fs_sel |= 0x0040
    struct.pack_into(">H", os2, 62, fs_sel)
    struct.pack_into(">hhh", os2, 68, 920, -240, 0)
    struct.pack_into(">HH", os2, 74, 920, 240)
    tables["OS/2"] = bytes(os2)

    post = bytearray(tables["post"])
    struct.pack_into(">I", post, 0, 0x00030000)
    tables["post"] = bytes(post[:32])

    for drop in ("vhea", "vmtx", "DSIG", "hdmx"):
        tables.pop(drop, None)

    tables["cmap"] = build_cmap_format4_and_12(new_cmap)
    tables["name"] = build_name_table(family, subfam, ps_name)
    return assemble_ttf(tables)


def main() -> None:
    os.makedirs(OUT_DIR, exist_ok=True)
    lit_reg = ensure_raw_font("Literata-Regular.ttf")
    lit_semi = ensure_raw_font("Literata-SemiBold.ttf")
    lit_it = ensure_raw_font("Literata-Italic.ttf")
    shs_reg = ensure_raw_font("SourceHanSerifSC-Regular.ttf")
    shs_semi = ensure_raw_font("SourceHanSerifSC-SemiBold.ttf")
    stix_reg = "/System/Library/Fonts/Supplemental/STIXTwoText.ttf"
    stix_it = "/System/Library/Fonts/Supplemental/STIXTwoText-Italic.ttf"

    cjk_chars = collect_needed_cjk_chars()
    print(f"Collected {len(cjk_chars)} CJK & punctuation codepoints.")

    family = "Bible Literata"
    reg_bytes = build_hybrid_face(
        lit_reg, stix_reg, shs_reg, cjk_chars,
        family, "Regular", "BibleLiterata-Regular", 400, False, False,
    )
    bold_bytes = build_hybrid_face(
        lit_semi, stix_reg, shs_semi, cjk_chars,
        family, "Bold", "BibleLiterata-Bold", 600, True, False,
    )
    it_bytes = build_hybrid_face(
        lit_it, stix_it, None, [],
        family, "Italic", "BibleLiterata-Italic", 400, False, True,
    )

    for name, payload in [
        ("bible-literata-regular.ttf", reg_bytes),
        ("bible-literata-bold.ttf", bold_bytes),
        ("bible-literata-italic.ttf", it_bytes),
    ]:
        out_path = os.path.join(OUT_DIR, name)
        with open(out_path, "wb") as f:
            f.write(payload)
        print(f"Wrote {out_path} ({len(payload) / 1024:.1f} KB)")


if __name__ == "__main__":
    main()
