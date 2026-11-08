#!/usr/bin/env python3
"""
seed.py — 随机种子派生器

用途：为 5 套原型各生成一串很长的随机字母数字，并把这串字符的统计特征
翻译成设计决策（配色 / 布局 / 字体 / 密度 / 圆角 / 动效性格）。

关键约束（来自需求）：
  * 种子字符串本身不出现在任何原型页面上；
  * 同一套派生规则对任意字符串都成立，可复现（固定主种子）；
  * 派生结果避开"AI 默认审美"区间（暖奶油纸 + 陶土红 + 衬线展示体）；
  * 5 套之间不撞车：布局 / 字体 / 底色 / 主色相互错开；
  * 主色对比度在生成时就按 WCAG 校验，不靠肉眼。

用法：
    python3 design/_seed/seed.py            # 生成/刷新 seeds.json
    python3 design/_seed/seed.py --print 3  # 只打印第 3 套的派生结果
"""

from __future__ import annotations

import hashlib
import json
import random
import string
from collections import Counter
from pathlib import Path

MASTER_SEED = 20261104  # 固定主种子：保证每次运行结果一致
SCHEME_COUNT = 5
SEED_LENGTH = 96

# 五个候选是一起派生的：底色、版式、字体要横向错开，主色相要两两拉开 26°，
# 这些都依赖"同时看全部候选"，所以派生仍然跑满 5 个，最后只留第 5 个。
# seeds.json 里只写留下的那一个，构建脚本也只出它。
KEEP_INDEX = 5

ALPHABET = string.ascii_letters + string.digits

# ---------------------------------------------------------------- 派生表 ----

# 版面骨架：决定"屏幕上东西怎么摆"，与配色/字体正交
LAYOUTS = [
    {
        "id": "split-focus",
        "name": "左窄栏 + 右主区",
        "note": "左侧固定窄导航（文字标签，非图标），右侧大面积工作区；阅读与练习都用同一骨架。",
        "grid": "240px 1fr",
        "reading": "主区内单栏居中，行宽 68ch",
    },
    {
        "id": "centered-column",
        "name": "居中单栏 + 顶部导航",
        "note": "全宽单列，正文居中，导航横置在顶部；查找、练习、词库都是同一列的变体。",
        "grid": "1fr / max-width 1180px",
        "reading": "主列居中，行宽 66ch",
    },
    {
        "id": "sidebar-rail",
        "name": "图标轨 + 内容台",
        "note": "极窄竖轨只放模块切换，主区像一张摊开的书页，占满可用宽度。",
        "grid": "64px 1fr",
        "reading": "主区宽排，正文块居中",
    },
    {
        "id": "triptych",
        "name": "三段工作台",
        "note": "左：目录/筛选；中：主内容；右：详情/释义/练习。查词结果常驻右栏，不打断。",
        "grid": "260px minmax(0,1fr) 320px",
        "reading": "中栏正文，右栏是浮层式信息面板",
    },
    {
        "id": "stacked-sheet",
        "name": "顶部操作条 + 下方文档",
        "note": "像一张纸：上方一条极薄的操作带，正文像信笺一样展开，纵向阅读节奏优先。",
        "grid": "1fr",
        "reading": "正文块居中偏左，宽度随视口收缩",
    },
    {
        "id": "mosaic",
        "name": "顶部分区 + 模块切换面板",
        "note": "顶部一条分区导航，下面按模块切换整块布局（阅读是长文，练习是单卡片，词库是表格）。",
        "grid": "1fr",
        "reading": "阅读视图内并排两栏（英文 / 中文）",
    },
    {
        "id": "split-screen-duo",
        "name": "左右对开双页",
        "note": "像摊开的对开本：左英文右中文，中间一条细缝；查词浮层贴在缝的两侧。",
        "grid": "1fr 1fr",
        "reading": "双页对开，逐节左右对齐",
    },
    {
        "id": "band-stack",
        "name": "横向色带 + 下沉内容",
        "note": "顶部一条深色带承载全局操作与进度，下面内容用留白分隔，不画卡片框。",
        "grid": "1fr",
        "reading": "内容在色带下居中排布",
    },
]

# 字体性格：每项给出"界面字体 / 阅读字体 / 数字与标签字体"三档
TYPEFACES = [
    {
        "id": "humanist",
        "name": "人文无衬线 + 过渡衬线",
        "display": "'Optima', 'Palatino', 'Georgia', serif",
        "ui": "'Avenir Next', 'Segoe UI', system-ui, sans-serif",
        "read": "'Iowan Old Style', 'Palatino Linotype', Palatino, Georgia, serif",
        "mono": "'SF Mono', 'JetBrains Mono', Menlo, monospace",
        "note": "圆润人文无衬线做界面，阅读用带暖调的过渡衬线，英文正文可读性优先。",
    },
    {
        "id": "grot-text",
        "name": "中性 grotesque + 书籍衬线",
        "display": "'Didot', 'Bodoni 72', 'Georgia', serif",
        "ui": "'Helvetica Neue', 'Inter', system-ui, sans-serif",
        "read": "'Charter', 'Bitstream Charter', 'Sitka Text', Cambria, serif",
        "mono": "'Menlo', monospace",
        "note": "界面用中性 grotesque 保持克制；阅读用屏幕衬线，弱化装饰、突出词形。",
    },
    {
        "id": "old-style",
        "name": "老式衬线系统",
        "display": "'Athelas', 'Charter', 'Georgia', serif",
        "ui": "'Optima', 'Avenir Next', system-ui, sans-serif",
        "read": "'Hoefler Text', 'Baskerville', 'Georgia', serif",
        "mono": "'Andale Mono', Menlo, monospace",
        "note": "全站衬线气质，界面靠字号与字重分层，不靠颜色分层；数字用等宽。",
    },
    {
        "id": "workhorse-grotesk",
        "name": "工作型 grotesque + 无衬线阅读",
        "display": "'Futura', 'Century Gothic', 'Avenir Next', sans-serif",
        "ui": "'Seravek', 'Gill Sans Nova', 'Segoe UI', system-ui, sans-serif",
        "read": "'Seravek', 'Verdana', 'Segoe UI', system-ui, sans-serif",
        "mono": "'Menlo', monospace",
        "note": "阅读与界面同一套字，靠字号、字重、行距分层；适合长时间高对比阅读。",
    },
    {
        "id": "slab-data",
        "name": "粗衬线标题 + 等宽数据",
        "display": "'Roboto Slab', 'Superclarendon', 'Rockwell', 'Georgia', serif",
        "ui": "'Superclarendon', 'Rockwell', 'Georgia', serif",
        "read": "'Superclarendon', 'Charter', Georgia, serif",
        "mono": "'Andale Mono', 'SF Mono', Menlo, monospace",
        "note": "标题与正文用衬线承载重量，练习区与统计用等宽，数字对齐全宽。",
    },
    {
        "id": "transitional-data",
        "name": "过渡衬线 + 窄体无衬线",
        "display": "'Hoefler Text', 'Baskerville', 'Georgia', serif",
        "ui": "'Avenir Next Condensed', 'Arial Narrow', 'Helvetica Neue', sans-serif",
        "read": "'Cochin', 'Charter', Georgia, serif",
        "mono": "'SF Mono', Menlo, monospace",
        "note": "界面字窄，密度高、占位小；阅读仍交给衬线，长段落不累。",
    },
    {
        "id": "geometric-soft",
        "name": "几何无衬线 + 轻衬线",
        "display": "'Gill Sans', 'Futura', 'Avenir Next', sans-serif",
        "ui": "'Futura', 'Century Gothic', 'Avenir Next', sans-serif",
        "read": "'Cochin', 'Iowan Old Style', Georgia, serif",
        "mono": "'Menlo', monospace",
        "note": "几何体做界面，圆润直白；阅读字温和，两者靠字号拉开距离。",
    },
    {
        "id": "humanist-serif-ui",
        "name": "衬线界面 + 无衬线阅读",
        "display": "'Seravek', 'Gill Sans Nova', 'Helvetica Neue', sans-serif",
        "ui": "'Charter', 'Georgia', serif",
        "read": "'Seravek', 'Verdana', 'Segoe UI', system-ui, sans-serif",
        "mono": "'Menlo', monospace",
        "note": "界面文字带书籍感，英文正文用无衬线以提高生词辨认度，适合初学者。",
    },
]

# 底色材质：影响"光"从哪来
GROUNDS = [
    {"id": "paper-cool", "mode": "light", "base": "#f4f5f3", "surface": "#ffffff", "ink": "#16181a", "desc": "冷调纸白，长时间阅读不刺眼"},
    {"id": "sand", "mode": "light", "base": "#eae3d5", "surface": "#f4efe4", "ink": "#1e1a14", "desc": "砂色纸面，比纯白更沉，适合暖光环境"},
    {"id": "blueprint-ink", "mode": "dark", "base": "#11161c", "surface": "#171d25", "ink": "#e7ecf2", "desc": "深蓝墨底，夜间或暗光环境"},
    {"id": "graphite", "mode": "dark", "base": "#151515", "surface": "#1d1d1d", "ink": "#ececec", "desc": "中性石墨黑，低饱和"},
    {"id": "moss-deep", "mode": "dark", "base": "#12180f", "surface": "#1a2116", "ink": "#e9eee2", "desc": "深苔绿底，带一点户外感"},
    {"id": "ash-light", "mode": "light", "base": "#e8e9e8", "surface": "#f2f3f3", "ink": "#1b1c1d", "desc": "灰阶底，靠黑白对比而非颜色分层"},
    {"id": "ink-plum", "mode": "dark", "base": "#17131a", "surface": "#1f1a23", "ink": "#ece6ef", "desc": "深紫墨底，暖一点的黑"},
    {"id": "sea-glass", "mode": "light", "base": "#e6ecea", "surface": "#f3f7f6", "ink": "#131a19", "desc": "极浅青玻璃底，冷而不冷"},
    {"id": "slate-day", "mode": "light", "base": "#dde2e6", "surface": "#eef1f4", "ink": "#14181c", "desc": "冷灰蓝纸面，偏工作台"},
    {"id": "umber-night", "mode": "dark", "base": "#1b1512", "surface": "#241d18", "ink": "#f0e8e1", "desc": "深棕夜底，最暖的暗色"},
    {"id": "chalk-white", "mode": "light", "base": "#fbfbfa", "surface": "#ffffff", "ink": "#101112", "desc": "近白 chalk 底，最高对比"},
    {"id": "teal-ink", "mode": "dark", "base": "#0d1a1a", "surface": "#132222", "ink": "#e4f0ee", "desc": "深青墨底，冷而稳"},
]

# AI 默认审美雷区：暖奶油 + 陶土红 / 信号红
DEFAULT_ACCENT_BANDS = [(2, 34), (334, 358)]


# ------------------------------------------------------------------ 工具 ----

def _digest(seed: str) -> bytes:
    return hashlib.blake2b(seed.encode("utf-8"), digest_size=48).digest()


def _pick(dg: bytes, offset: int, modulo: int) -> int:
    """从哈希摘要里切一段独立字节做下标选择——各维度之间互不相关。"""
    chunk = dg[offset:offset + 4]
    return int.from_bytes(chunk, "big") % modulo


def _clamp(v: float, lo: float, hi: float) -> float:
    return max(lo, min(hi, v))


def _hsl_rgb(h: float, s: float, l: float) -> tuple[int, int, int]:
    h = h % 360
    s = _clamp(s, 0, 100) / 100
    l = _clamp(l, 0, 100) / 100
    c = (1 - abs(2 * l - 1)) * s
    x = c * (1 - abs((h / 60) % 2 - 1))
    m = l - c / 2
    if h < 60:
        r, g, b = c, x, 0
    elif h < 120:
        r, g, b = x, c, 0
    elif h < 180:
        r, g, b = 0, c, x
    elif h < 240:
        r, g, b = 0, x, c
    elif h < 300:
        r, g, b = x, 0, c
    else:
        r, g, b = c, 0, x
    return (round((r + m) * 255), round((g + m) * 255), round((b + m) * 255))


def _hex_rgb(hx: str) -> tuple[int, int, int]:
    hx = hx.lstrip("#")
    return (int(hx[0:2], 16), int(hx[2:4], 16), int(hx[4:6], 16))


def _lum(rgb: tuple[int, int, int]) -> float:
    def ch(v: int) -> float:
        v /= 255
        return v / 12.92 if v <= 0.04045 else ((v + 0.055) / 1.055) ** 2.4
    r, g, b = (ch(c) for c in rgb)
    return 0.2126 * r + 0.7152 * g + 0.0722 * b


def _contrast(a: tuple[int, int, int], b: tuple[int, int, int]) -> float:
    la, lb = _lum(a), _lum(b)
    hi, lo = max(la, lb), min(la, lb)
    return (hi + 0.05) / (lo + 0.05)


def _hsl(h: float, s: float, l: float) -> str:
    return f"hsl({round(h % 360)} {round(_clamp(s, 0, 100))}% {round(_clamp(l, 0, 100))}%)"


def _shift_out_of_defaults(h: float) -> float:
    """主色若落进 AI 默认色带（陶土红 / 信号红），跨到冷色一侧去。"""
    h = h % 360
    for lo, hi in DEFAULT_ACCENT_BANDS:
        if lo <= h <= hi:
            h = (h + 148) % 360
            break
    return h


def _fit_lightness(h: float, s: float, bg_hex: str, dark: bool) -> float:
    """在偏好明度区间里挑一个既达标（≥4.5:1）又不过刺眼的主色明度。"""
    bg = _hex_rgb(bg_hex)
    lo, hi = (54, 76) if dark else (24, 40)
    target = 7.0 if dark else 5.2
    best = None
    for l in range(lo, hi + 1):
        ratio = _contrast(_hsl_rgb(h, s, l), bg)
        if ratio < 4.5:
            continue
        score = abs(ratio - target)
        if best is None or score < best[1]:
            best = (l, score)
    if best:
        return float(best[0])
    # 兜底：整个区间都不达标时取对比度最高的那个
    cands = [(l, _contrast(_hsl_rgb(h, s, l), bg)) for l in range(6, 95)]
    return float(max(cands, key=lambda t: t[1])[0])


# ------------------------------------------------------------- 派生主体 ----

def derive(seed: str, override: dict | None = None, hue_shift: float = 0.0) -> dict:
    n = len(seed)
    counts = Counter(seed)
    dg = _digest(seed)
    digits = [int(c) for c in seed if c.isdigit()]
    letters = [c for c in seed if c.isalpha()]
    upper = sum(1 for c in letters if c.isupper())
    lower = sum(1 for c in letters if c.islower())

    # --- 字符统计特征
    entropy = len(counts) / n                      # 唯一字符占比
    vowel_ratio = sum(1 for c in letters if c.lower() in "aeiou") / max(1, len(letters))
    digit_ratio = len(digits) / n
    upper_ratio = upper / max(1, upper + lower)

    max_run = 1
    run = 1
    for i in range(1, n):
        if seed[i] == seed[i - 1]:
            run += 1
            max_run = max(max_run, run)
        else:
            run = 1

    digit_sum = sum(digits)
    code_sum = sum(ord(c) for c in seed)

    # ---------------- 配色 ----------------
    gi = _pick(dg, 0, len(GROUNDS))
    li = _pick(dg, 20, len(LAYOUTS))
    fi = _pick(dg, 24, len(TYPEFACES))
    if override:
        gi = override.get("ground", gi)
        li = override.get("layout", li)
        fi = override.get("typeface", fi)

    ground = GROUNDS[gi]
    dark = ground["mode"] == "dark"

    base_hue = (digit_sum * 11 + _pick(dg, 4, 360)) % 360
    accent_hue = _shift_out_of_defaults((base_hue + _pick(dg, 12, 190)) % 360)
    accent_hue = (accent_hue + hue_shift) % 360
    support_hue = (accent_hue + 140 + _pick(dg, 16, 50)) % 360

    sat_accent = 32 + int(entropy * 58)
    sat_main = 16 + int(vowel_ratio * 64)          # 元音多 -> 饱和高

    accent_l = _fit_lightness(accent_hue, sat_accent, ground["base"], dark)
    # 填充色（按钮底）：压到另一端，保证压在上面的文字达标
    fill_l = 30 if dark else accent_l
    fill_rgb = _hsl_rgb(accent_hue, sat_accent, fill_l)
    on_fill = "#ffffff" if _contrast(fill_rgb, (255, 255, 255)) >= 4.5 else "#0d0d0d"
    # 浅底上的按钮用白字时若不够暗，就把填充再压一档
    if on_fill == "#0d0d0d" and not dark:
        fill_l = max(18.0, fill_l - 8)
        fill_rgb = _hsl_rgb(accent_hue, sat_accent, fill_l)
    # 低彩度底纹：浅底用近白的淡彩，深底用近黑的暗彩
    soft_l = 93 if not dark else 21
    soft_rgb = _hsl_rgb(accent_hue, min(sat_accent, 40), soft_l)
    ink_on_soft = _hsl(accent_hue, min(sat_accent, 65), 18 if not dark else 93)

    # ---------------- 布局 ----------------
    layout = LAYOUTS[li]

    # ---------------- 字体 ----------------
    typeface = TYPEFACES[fi]

    # ---------------- 形状与密度 ----------------
    density = ["compact", "balanced", "roomy"][_pick(dg, 28, 3)]
    radius = {"compact": 4, "balanced": 9, "roomy": 14}[density] + _pick(dg, 32, 5) - 2
    if max_run >= 5:
        radius = 0
    radius = int(_clamp(radius, 0, 16))
    leading = {"compact": 1.58, "balanced": 1.7, "roomy": 1.78}[density]
    measure = 62 + _pick(dg, 36, 12)               # 英文正文行宽 ch
    scale = round(1.18 + vowel_ratio * 0.16, 3)    # 字阶比例

    border_style = ["hairline", "solid", "soft"][_pick(dg, 40, 3)]

    motion = [
        {"id": "quick-crisp", "name": "短促线性", "desc": "120–180ms，无回弹"},
        {"id": "settled", "name": "沉稳回落", "desc": "260–360ms，末端减速"},
        {"id": "springy", "name": "轻微回弹", "desc": "弹簧带一次极小回弹"},
        {"id": "crossfade", "name": "交叉淡入", "desc": "只淡入不位移"},
    ][_pick(dg, 44, 4)]

    return {
        "seed": seed,
        "seedLength": n,
        "stats": {
            "uniqueRatio": round(entropy, 3),
            "vowelRatio": round(vowel_ratio, 3),
            "digitRatio": round(digit_ratio, 3),
            "upperRatio": round(upper_ratio, 3),
            "longestRun": max_run,
            "digitSum": digit_sum,
            "codeSum": code_sum,
        },
        "palette": {
            "groundId": ground["id"],
            "mode": ground["mode"],
            "base": ground["base"],
            "surface": ground["surface"],
            "ink": ground["ink"],
            "groundDesc": ground["desc"],
            "accentHue": round(accent_hue, 1),
            "accent": _hsl(accent_hue, sat_accent, accent_l),
            "accentSoft": _hsl(accent_hue, min(sat_accent, 40), soft_l),
            "accentInk": ink_on_soft,
            "accentFill": _hsl(accent_hue, sat_accent, fill_l),
            "onAccent": on_fill,
            "secondary": _hsl(support_hue, sat_main, accent_l),
            "rule": _hsl(accent_hue, 12, 82 if not dark else 30),
        },
        "layout": layout,
        "typeface": typeface,
        "shape": {
            "radius": radius,
            "density": density,
            "leading": leading,
            "measureCh": measure,
            "scale": scale,
            "borderStyle": border_style,
        },
        "motion": motion,
    }


# ----------------------------------------------------------------- 构建 ----

def _spread(raw: list[int], dgs: list[bytes], size: int, offset: int) -> list[int]:
    """把撞车的下标确定性地错开，保证同一维度上 5 套互不重复。"""
    used: set[int] = set()
    out: list[int] = []
    for i, v in enumerate(raw):
        if v in used:
            step = 1 + _pick(dgs[i], offset, 3)
            for k in range(1, size + 1):
                cand = (v + step * k) % size
                if cand not in used:
                    v = cand
                    break
        used.add(v)
        out.append(v)
    return out


def build() -> list[dict]:
    rng = random.Random(MASTER_SEED)
    seeds = [
        "".join(rng.choice(ALPHABET) for _ in range(SEED_LENGTH))
        for _ in range(SCHEME_COUNT)
    ]
    dgs = [_digest(s) for s in seeds]

    grounds = _spread([_pick(dg, 0, len(GROUNDS)) for dg in dgs], dgs, len(GROUNDS), 48)
    layouts = _spread([_pick(dg, 20, len(LAYOUTS)) for dg in dgs], dgs, len(LAYOUTS), 52)
    faces = _spread([_pick(dg, 24, len(TYPEFACES)) for dg in dgs], dgs, len(TYPEFACES), 56)

    schemes = []
    for i, s in enumerate(seeds):
        ov = {"ground": grounds[i], "layout": layouts[i], "typeface": faces[i]}
        d = derive(s, override=ov)

        # 主色相两两拉开至少 26°，避免 5 套并排时主色黏在一起
        shift = 0.0
        for attempt in range(12):
            hue = (d["palette"]["accentHue"] + shift) % 360
            gaps = [
                abs(((hue - schemes[j]["palette"]["accentHue"]) + 180) % 360 - 180)
                for j in range(i)
            ]
            if all(g >= 26 for g in gaps):
                break
            shift += 26
        if shift:
            d = derive(s, override=ov, hue_shift=shift)
        schemes.append(d)

    for i, d in enumerate(schemes):
        d["index"] = i + 1
        d["slug"] = f"{i + 1:02d}-{d['layout']['id']}-{d['typeface']['id']}"
    return schemes


def main() -> None:
    schemes = build()
    kept = [s for s in schemes if s["index"] == KEEP_INDEX]
    dest = Path(__file__).resolve().parent / "seeds.json"
    dest.write_text(json.dumps(kept, ensure_ascii=False, indent=2), encoding="utf-8")
    print(f"wrote {dest}")
    for s in kept:
        st = s["stats"]
        p = s["palette"]
        c = _contrast(_hsl_rgb(p["accentHue"], 60, 40), _hex_rgb(p["base"]))
        print(
            f"\n[{s['index']}] {s['slug']}\n"
            f"  seed      : {s['seed']}\n"
            f"  统计      : uniq={st['uniqueRatio']} vowel={st['vowelRatio']} digit={st['digitRatio']} "
            f"upper={st['upperRatio']} run={st['longestRun']}\n"
            f"  配色      : {p['mode']} / {p['base']} / accent {p['accent']} (hue {p['accentHue']})\n"
            f"  布局      : {s['layout']['name']} ({s['layout']['id']})\n"
            f"  字体      : {s['typeface']['name']} ({s['typeface']['id']})\n"
            f"  形状      : r={s['shape']['radius']} {s['shape']['density']} lead={s['shape']['leading']} "
            f"measure={s['shape']['measureCh']}ch scale={s['shape']['scale']} {s['shape']['borderStyle']}\n"
            f"  动效      : {s['motion']['name']}"
        )
    print(f"\n（另外 {SCHEME_COUNT - 1} 个候选已弃用，seeds.json 只保留第 {KEEP_INDEX} 个。）")


if __name__ == "__main__":
    main()
