#!/usr/bin/env python3
"""把 design/_build 的模板与 design/_seed/seeds.json 合成原型。

seeds.json 里只留了第 5 个候选（见 design/_seed/seed.py 的 KEEP_INDEX），
所以这里只出一套：design/05-sidebar-rail-grot-text/{index.html, styles.css, app.js}
"""
import json
import pathlib
import re

ROOT = pathlib.Path(__file__).resolve().parent.parent
BUILD = ROOT / "_build"
SEEDS = ROOT / "_seed" / "seeds.json"

MOTION = {
    "quick-crisp": ("130ms", "cubic-bezier(.4,0,.2,1)"),
    "crossfade": ("200ms", "cubic-bezier(.4,0,.2,1)"),
    "settled": ("280ms", "cubic-bezier(.22,.61,.36,1)"),
    "springy": ("240ms", "cubic-bezier(.32,.72,0,1)"),
}


def _hex_rgb(h):
    h = h.strip().lstrip("#")
    return tuple(int(h[i:i + 2], 16) for i in (0, 2, 4))


def _lum(rgb):
    def f(c):
        c = c / 255
        return c / 12.92 if c <= 0.03928 else ((c + 0.055) / 1.055) ** 2.4
    r, g, b = (f(x) for x in rgb)
    return 0.2126 * r + 0.7152 * g + 0.0722 * b


def _contrast(a, b):
    la, lb = _lum(a), _lum(b)
    hi, lo = max(la, lb), min(la, lb)
    return (hi + 0.05) / (lo + 0.05)


def _mix(ink, bg, p):
    return tuple(round(ink[i] * p + bg[i] * (1 - p)) for i in range(3))


def _pick_mix(ink, bgs, target):
    """挑一个 ink 与底色的混合比例，让文字刚好过对比度线（取最弱的那档仍达标）。"""
    for i in range(30, 97):
        p = i / 100
        c = _mix(ink, bgs[0], p)
        if all(_contrast(c, bg) >= target for bg in bgs):
            return p, c
    return 1.0, ink


def _hex(c):
    return "#%02x%02x%02x" % tuple(int(v) for v in c)


def _hsl_rgb(h):
    m = re.match(r"hsl\(\s*([\d.]+)\s+([\d.]+)%\s+([\d.]+)%\s*\)", h.strip())
    if not m:
        return None
    hh = float(m.group(1)) % 360
    ss = float(m.group(2)) / 100
    ll = float(m.group(3)) / 100
    c = (1 - abs(2 * ll - 1)) * ss
    x = c * (1 - abs((hh / 60) % 2 - 1))
    m0 = ll - c / 2
    if hh < 60:
        rgb = (c, x, 0)
    elif hh < 120:
        rgb = (x, c, 0)
    elif hh < 180:
        rgb = (0, c, x)
    elif hh < 240:
        rgb = (0, x, c)
    elif hh < 300:
        rgb = (x, 0, c)
    else:
        rgb = (c, 0, x)
    return tuple(round((v + m0) * 255) for v in rgb)


def _paper_tokens(p):
    """书页用色：纸、正文墨色、次级墨色、界线、纸上强调色。

    纸永远是纸：暖白里掺一点本方案的主色，墨是暖黑；
    强调色要压到纸上仍过 4.6:1。"""
    accent = _hsl_rgb(p["accent"]) or _hex_rgb(p["accent"])
    base = (250, 246, 238)
    paper = _mix(accent, base, 0.055)
    p_ink = _mix(accent, (34, 30, 26), 0.10)
    _, p_soft = _pick_mix(p_ink, [paper], 4.6)
    p_rule = _mix(p_ink, paper, 0.16)
    p_acc = accent
    for i in range(0, 101):
        c = _mix(p_ink, accent, i / 100)
        if _contrast(c, paper) >= 4.6:
            p_acc = c
            break
    p_mark = _mix(p_acc, paper, 0.16)
    return {
        "paper": _hex(paper),
        "paper_ink": _hex(p_ink),
        "paper_soft": _hex(p_soft),
        "paper_rule": _hex(p_rule),
        "paper_accent": _hex(p_acc),
        "paper_mark": _hex(p_mark),
    }


def tokens(scheme):
    p = scheme["palette"]
    t = scheme["typeface"]
    s = scheme["shape"]
    dark = p["mode"] == "dark"
    dur, ease = MOTION[scheme["motion"]["id"]]
    if scheme["slug"].startswith("05"):
        dur, ease = "220ms", "cubic-bezier(.32,.72,0,1)"
    radius = max(8, min(16, int(s["radius"])))
    ink = _hex_rgb(p["ink"])
    bg = _hex_rgb(p["base"])
    surf = _hex_rgb(p["surface"])
    dark = p["mode"] == "dark"
    _, soft_rgb = _pick_mix(ink, [bg, surf], 7.0)
    _, faint_rgb = _pick_mix(ink, [bg, surf], 4.8)
    soft_hex = _hex(soft_rgb)
    faint_hex = _hex(faint_rgb)
    paper = _paper_tokens(p)
    danger = "hsl(6 72% 64%)" if dark else "hsl(4 64% 44%)"
    warn = "hsl(38 84% 62%)" if dark else "hsl(30 74% 36%)"
    shadow = (
        "0 2px 6px rgba(0,0,0,.28), 0 28px 56px -28px rgba(0,0,0,.6)"
        if dark
        else "0 2px 5px rgba(16,20,18,.08), 0 24px 48px -26px rgba(16,20,18,.32)"
    )
    return f""":root {{
  color-scheme: {p["mode"]};

  --bg: {p["base"]};
  --surface: {p["surface"]};
  --surface-2: color-mix(in oklab, var(--ink) 6%, var(--surface));
  --surface-3: color-mix(in oklab, var(--ink) 11%, var(--surface));
  --ink: {p["ink"]};
  --ink-soft: {soft_hex};
  --ink-faint: {faint_hex};
  --rule: {p["rule"]};
  --rule-soft: color-mix(in oklab, var(--ink) 13%, var(--bg));

  --paper: {paper["paper"]};
  --paper-ink: {paper["paper_ink"]};
  --paper-soft: {paper["paper_soft"]};
  --paper-rule: {paper["paper_rule"]};
  --paper-accent: {paper["paper_accent"]};
  --paper-mark: {paper["paper_mark"]};

  --accent: {p["accent"]};
  --accent-soft: {p["accentSoft"]};
  --accent-fill: {p["accentFill"]};
  --accent-ink: {p["accentInk"]};
  --on-accent: {p["onAccent"]};
  --secondary: {p["secondary"]};
  --danger: {danger};
  --warn: {warn};

  --r-card: {radius}px;
  --r-ctl: {min(radius, 10)}px;
  --r-sm: {max(4, radius - 6)}px;
  --lead: {s["leading"]};
  --measure: {s["measureCh"]}ch;
  --scale: {s["scale"]};

  --font-ui: {t["ui"]};
  --font-read: {t["read"]};
  --font-mono: {t["mono"]};
  --font-display: {t["display"]};

  --dur: {dur};
  --ease: {ease};
  --shadow-float: {shadow};
}}
"""


def build():
    schemes = json.loads(SEEDS.read_text(encoding="utf-8"))
    base_css = (BUILD / "css" / "base.css").read_text(encoding="utf-8")
    shell = (BUILD / "shell.html").read_text(encoding="utf-8")
    js_dir = BUILD / "js"
    parts = sorted(p for p in js_dir.glob("*.js"))

    for scheme in schemes:
        slug = scheme["slug"]
        out = ROOT / slug
        out.mkdir(parents=True, exist_ok=True)

        layout_css = (BUILD / "css" / f"layout-{scheme['layout']['id']}.css").read_text(
            encoding="utf-8"
        )
        css = base_css.replace("/*__TOKENS__*/", tokens(scheme)) + "\n" + layout_css
        (out / "styles.css").write_text(css, encoding="utf-8")

        motion_attr = "apple" if slug.startswith("05") else scheme["motion"]["id"]
        html = shell.replace("__LAYOUT__", scheme["layout"]["id"]).replace(
            "__MOTION__", motion_attr
        )
        (out / "index.html").write_text(html, encoding="utf-8")

        chunks = []
        for part in parts:
            if part.name == "07-apple.js" and not slug.startswith("05"):
                continue
            chunks.append(f"/* ---- {part.name} ---- */\n" + part.read_text(encoding="utf-8"))
        (out / "app.js").write_text("\n".join(chunks), encoding="utf-8")

        print(f"[ok] {slug}  layout={scheme['layout']['id']} type={scheme['typeface']['id']}")


if __name__ == "__main__":
    build()
