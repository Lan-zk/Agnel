# 圣经离线数据库（BSB / WEBP / KJV / 和合本）

三个英文译本加一本中文和合本的完整 66 卷正文，统一整理为一份 SQLite 数据库，供英语学习应用离线使用。

| 项目 | 数值 |
|---|---|
| 译本 | BSB、WEBP、KJV（英文）、CUV 和合本（简体中文） |
| 书卷 | 每个译本 66 卷（旧约 39 + 新约 27） |
| 章节 | 每个译本 1189 章（旧约 929 + 新约 260） |
| 经节 | BSB 31,102 · WEBP 31,103 · KJV 31,102 · CUV 31,092（合计 124,399 行） |
| 标题/题注 | 8,168 条（`sections` 表） |
| 数据库 | `data/db/bible-en-study.db`，33,443,840 字节（31.9 MiB） |
| 字符编码 | UTF-8（全部为 NFC 规范化形式） |

校验结果：`data/reports/validation_report.md`，83 项检查全部通过。

---

## 1. 译本、来源与授权

所有文件均直接取自版权方/发布方，未使用任何第三方整理版。

| 译本 | 版本 | 来源页面 | 实际下载文件 | 下载日期 (UTC) | 授权 |
|---|---|---|---|---|---|
| **BSB** Berean Standard Bible | 第 3 版印刷文本的 USFM 导出 | <https://berean.bible/downloads.htm> | `https://bereanbible.com/bsb_usfm.zip` | 2026-11-06 | Public Domain |
| **WEBP** World English Bible, Protestant Edition | 2020 stable text edition | <https://ebible.org/find/details.php?id=engwebp> | `https://eBible.org/Scriptures/engwebp_usfm.zip` | 2026-11-06 | Public Domain |
| **KJV** King James Version | 1769 标准文本（`eng-kjv2006`） | <https://ebible.org/find/details.php?id=engkjv> | `https://eBible.org/Scriptures/eng-kjv2006_usfm.zip` | 2026-11-06 | Public Domain |
| **CUV** 和合本（简体，新标点） | 新标点和合本 `cmn-cu89s`（CHNCU1），2021-10-15 | <https://ebible.org/find/details.php?id=cmn-cu89s> | `https://eBible.org/Scriptures/cmn-cu89s_usfm.zip` | 2026-11-06 | Public Domain |

授权依据（下载前逐页核对，原文摘录）：

- **BSB** — <https://berean.bible/licensing.htm>：*"The Berean Bible and Majority Bible texts are officially placed into the public domain as of April 30, 2023 … Licensing is not required for any use."*
- **WEBP** — eBible.org 详情页标注 `public domain`。注意 *"World English Bible"* 与 *"WEB"* 是 eBible.org 的商标，**修改后的文本不得继续以 WEB 名义分发**。本库未改动经文用词（仅规范化标点周围空白，见 §5.4），因此保留原译本名。
- **KJV** — eBible.org 详情页标注 `public domain`，源文件由 Crosswire Bible Society 提供。英国境内的印刷受 Crown letters patent 限制；英国以外属公有领域。
- **CUV** — 和合本 1919 年文本早已进入公有领域；eBible.org 详情页标注 `public domain`，语言 `cmn`（中国语文，简体），版本说明 *"新标点和合本 / Chinese Union Version (simplified)"*，最后更新 2021-10-15。该页另提供同一文本的纯文本导出（`cmn-cu89s_vpl.zip`），本库只把它当作独立比对文件，不当作正文来源。

  > 说明：最初考察过 `github.com/Lan-zk/bible` 提供的 SQLite 版和合本，它同样是 1919 年文本，但**缺全部 116 条诗篇题注**、把 485 条编者注（「或译」「有古卷」）直接混在正文里，且仓库无作者与授权说明，因此改用 eBible.org 的 `cmn-cu89s`（公有领域、带完整 USFM 结构、另有同源纯文本可供交叉核对）。

每个译本的完整元数据（含 URL、下载时间、压缩包 sha256、版本、授权、版权说明）都写在 `translations` 表里，可随时查询，不依赖本文件。

---

## 2. 目录结构

```
data/
  raw/                         下载的原始包（可随时用脚本重新下载）
    download_manifest.json     每个文件的 url / 字节数 / sha256 / 抓取时间
    bsb_usfm.zip  bsb.txt
    engwebp_usfm.zip  engwebp_vpl.zip
    eng-kjv2006_usfm.zip  eng-kjv2006_vpl.zip
    cmn-cu89s_usfm.zip  cmn-cu89s_vpl.zip      CUV（和合本）
    bsb/bsb_usfm/*.usfm        BSB 的 66 卷 USFM
    webp/*.usfm                WEBP 的 66 卷 USFM（另含前言、词汇表）
    kjv/*.usfm                 KJV 的 66 卷 USFM
    cuv/*.usfm                 CUV 的 66 卷 USFM（文件序号是 eBible 自己的编号）
    verify/*_vpl.txt           出版社自己的纯文本导出，用作独立比对
  db/bible-en-study.db         ← 交付物
  reports/
    validation_report.md       ← 数据校验报告
    validation_report.json
    build_summary.json         构建统计
  lex/                        词典构建源与脚本；见 lex/README.md
scripts/
  canon.py                    66 卷正典元数据（OSIS / USFM / 章数）
  usfm_parser.py              USFM → 经节记录
  sources.py                  译本元数据、文件定位、修复规则
  download_sources.py         下载 + 校验 + 解包（唯一需要网络的步骤）
  build_database.py           构建数据库
  reference_compare.py        与出版社纯文本比对
  verify_database.py          校验并生成报告
  schema.sql                  建表语句
```

---

## 3. 数据库结构

```sql
translations        -- 每译本一行：名称、版本、授权、来源 URL、下载日期、sha256、统计
books               -- 66 卷正典：book_id(1..66) / osis_id / usfm_id / 名称 / 约 / 分类 / 章数
translation_books   -- 每译本每卷一行：该译本的卷名（\h、\toc1、\toc3）、章数、节数、源文件
verses              -- 经节正文：(translation_id, book_id, chapter, verse) → text
sections            -- 章节标题、诗篇题注、诗篇 119 字母、雅歌说话人、平行经文行
verses_fts          -- FTS5 全文索引（unicode61，去变音符号）
dataset_meta        -- schema 版本、构建时间、正典说明
```

关键设计：

1. **经节身份是 `(translation_id, book_id, chapter, verse)`**，不是行号、不是自增 id。任何按行对齐的转换都会在译本间经节划分不同时错位（见 §5），所以这里从表结构上排除了这种可能。
2. `ux_verses_ref` 唯一索引保证同译本同节只有一行；`ix_verses_parallel` 支持「某节在其他译本中是什么」的反射查询。
3. 标题一律进 `sections`，**不并入正文**（BSB 的卷首题注除外，见 §5）。
4. `verses.text` 为 `TEXT NOT NULL`；译本故意不收的经节存空字符串而不是删除，保证节号连续。
5. 中文书卷名取自 USFM 的 `\h` / `\toc1`（`创世记` 等），与 `books.name` 的英文名并存，应用可直接用中文显示。

### 视图

```sql
v_verse             -- 经节 + 译本缩写 + OSIS + 卷名，方便直接展示
v_verse_parallel    -- 每个 (卷, 章, 节) 一行：
                       bsb_text / cuv_text / kjv_text / webp_text
                       + translation_count（该节被几个译本收录）
```

`v_verse_parallel` 的列名在构建时按 `translations` 表生成，增加译本不需要改 SQL。

---

## 4. 常用查询

```sql
-- 按译本 + 卷 + 章 + 节取一节
SELECT text FROM verses
WHERE translation_id = 'KJV' AND book_id = 19 AND chapter = 23 AND verse = 1;

-- 取一整章（按节号排序，用 OSIS 定位书卷）
SELECT v.verse, v.text
FROM verses v JOIN books b ON b.book_id = v.book_id
WHERE v.translation_id = 'CUV' AND b.osis_id = 'John' AND v.chapter = 3
ORDER BY v.verse;

-- 同一节四译本对照
SELECT bsb_text, webp_text, kjv_text, cuv_text
FROM v_verse_parallel
WHERE book_id = 19 AND chapter = 23 AND verse = 1;

-- 这一节有几个译本收录（罗马书 14/16 章会看到 3）
SELECT verse, translation_count FROM v_verse_parallel
WHERE book_id = 45 AND chapter = 16 ORDER BY verse;

-- 章节标题 / 诗篇题注
SELECT verse, heading_type, text FROM sections
WHERE translation_id = 'KJV' AND book_id = 19 AND chapter = 23;

-- 英文全文检索
SELECT v.translation_id, b.osis_id, v.chapter, v.verse
FROM verses_fts f
JOIN verses v ON v.id = f.rowid
JOIN books  b ON b.book_id = v.book_id
WHERE verses_fts MATCH '"still waters"'
ORDER BY v.translation_id, v.book_id, v.chapter, v.verse;

-- 中文检索
-- unicode61 分词把中文按标点切成整句，所以整句/整短语能命中：
SELECT v.chapter, v.verse FROM verses_fts f JOIN verses v ON v.id = f.rowid
WHERE verses_fts MATCH '耶和华是我的牧者' AND v.translation_id = 'CUV';
-- 句中任意子串用 LIKE（CUV 3 万行，全表扫描约 5 ms）：
SELECT chapter, verse, text FROM verses
WHERE translation_id = 'CUV' AND text LIKE '%爱世人%';

-- 译本元数据（来源、日期、授权）
SELECT translation_id, name, edition, license, download_date, sha256
FROM translations;
```

Python（标准库即可，无需第三方依赖）：

```python
import sqlite3
con = sqlite3.connect("data/db/bible-en-study.db")
con.row_factory = sqlite3.Row
row = con.execute(
    "SELECT text FROM verses WHERE translation_id=? AND book_id=? "
    "AND chapter=? AND verse=?", ("BSB", 43, 3, 16)).fetchone()
print(row["text"])
```

---

## 5. 数据保真说明

### 5.1 没有做的事

- 没有按行号或顺序把「第 N 行」当作「第 N 节」。
- 没有补写、改写、顺滑任何经文用词；没有改动节号、章号、卷序。
- 没有把标题、题注、说话人并入经文（唯一例外见 5.3 BSB 诗篇题注，那是 BSB 自己的排法）。
- 没有收录次经/旁经（WEBP 用的是 `engwebp`，不是含次经的 `engweb`）。
- 没有擅自「补全」和合本不收的节：源文件不编号的节号，库里就没有这一行。

### 5.2 空经节：出版社本来就不收

| 译本 | 数量 | 来源 |
|---|---:|---|
| BSB | 16 | Matt 17:21, 18:11, 23:14; Mark 7:16, 9:44, 9:46, 11:26, 15:28; Luke 17:36, 23:17; John 5:4; Acts 8:37, 15:34, 24:7, 28:29; Rom 16:24 |
| WEBP | 5 | Luke 17:36; Acts 8:37, 15:34, 24:7; Rom 16:25 |
| KJV | 0 | — |
| CUV | 71 | 和合本把两个英文节号并成一节的 `\v 20-21` 之类（如 Num 1:21、Deut 13:13/18、Ps 116:19），节号保留、正文留空 |

BSB 的空节在其 USFM 里连 `\v` 标记都没有，但出版社自己的 `bsb.txt` 把它们列成空行；WEBP 是只留脚注的 `\v 36 \f … \f*` 行；和合本则用 `\v 20-21` 的节号区间表示「这一节和上一节印在一起」。三种情况下都**保留节号、正文留空**，跨译本查同一节时不会因为缺行而错位。空节不是缺失数据，`translations.empty_verse_count` 与报告 C 组都逐节列出。

另有 11 个节号和合本**完全没有编号**（源文件既无 `\v` 也无区间标记）：Matt 18:11, 23:14; Mark 7:16, 15:28; Luke 17:36, 23:17; John 5:4; Acts 8:37, 15:34, 24:7, 28:29。这些节在库里没有行，报告 C 组会说明「本版不编号，出版社导出同样没有」。

### 5.3 译本之间的经节差异（不强行对齐）

全书 1,189 章中，**有 5 章**的节数在不同译本间不同：

| 章 | BSB | KJV | WEBP | CUV |
|---|---|---|---|---|
| 罗马书 14 | 23 节 | 23 节 | **26 节**（24–26 为颂赞） | 23 节 |
| 罗马书 16 | 27 节 | 27 节 | **25 节** | 27 节 |
| 约翰福音 7 | 53 节 | 53 节 | 53 节 | **52 节**（不收 7:53，并入 8:1） |
| 约翰三书 1 | 14 节 | 14 节 | 14 节 | **15 节**（第 14 节拆成 14、15） |
| 启示录 12 | 17 节 | 17 节 | 17 节 | **18 节**（英文 13:1 首句编为 12:18） |

WEB 把颂赞放在 14:24–26，BSB/KJV/CUV 放在 16:25–27；WEBP 自己的脚注也写明 *"Romans 14:24-26 in the WEB may appear as Romans 16:25-27 in other translations"*。数据库**不做任何重编号**：各自的编号原样保留，`v_verse_parallel.translation_count` 会让某一节「只有 3 个译本」的情况直接显形。

诗篇标题的处理也随出版社而异，因此没有统一「挪到第 1 节前面」：

- **BSB**：题注就在第 1 节里（`\d \v 1`），所以 `Ps 3:1` 的正文以 `A Psalm of David, when he fled…` 开头。
- **WEBP / KJV / CUV**：题注是 `\d`，独立成 `sections` 行（`heading_type = 'psalm_title'`），第 1 节正文从 `LORD, how…` / `耶和华啊，我的敌人何其加增…` 开始。

诗篇 119 的希伯来字母：WEBP 用 `\d BETH`（22 条），BSB 同时给出字母与音译两行（`\qa א` + `\qa ALEPH`，44 条），**两份都保留**，因为源文件就是这么排的。

雅歌的 `\sp`（说话人）标记：出现在经节开始处时记为 `sections`（`heading_type = 'speaker'`）；出现在一节中间时（如 Song 1:4 的 `Friends`）并入正文，因为出版社自己的纯文本导出就是行内排版。

和合本的 `\r`（平行经文提示，如 `\r （诗53）`，618 条）同样进 `sections`（`heading_type = 'reference'`），不混入正文。

### 5.4 排版规范化（不改用词）

USFM 转换器留下的是**空白与标点间距**问题，逐条列在 `scripts/usfm_parser.py` 的 `TIDY_RULES` 里，并已用出版社纯文本导出逐节验证：

| 规则 | 说明 | 影响 / 示例 |
|---|---|---|
| `[’’]` → 删 | bsb2usfm 残留 | 59 节 |
| `vvv` → 删 | bsb2usfm 残留 | Gen 35:18 |
| `. . .` → 删 | 间隔号省略号残留 | Num 26 等 11 节 |
| 标点前空格 → 删 | 收拢 `, ; : ! ? ) ] ” ’` 前的空白 | Lev 7:9（`Likewise , every` → `Likewise, every`）等 21 处 |
| 括号/引号后空格 → 删 | 收拢 `( “ ‘ [` 后的空白 | 11 + 9 + 2 + 4 处 |
| `:文字` → `: 文字` | 冒号后补一个空格 | 逐节比对中未再命中 |
| `,` 后接引号 → 补空格 | 脚注移除后留下的 `Hebrew,‘Saul` 应为 `Hebrew, ‘Saul` | Acts 26:14、Rev 1:11 |
| 汉字之间的空格 → 删 | 中文不设字间距；eBible 转换器在 `\pn`/`\add` 包裹处留下了空格 | 和合本全库 |
| `\v 20-21` → 只取 20 | 合并节号，21 另存空行（见 5.2） | 和合本 71 处 |

另外两处是转换器**丢了空格**，用显式修复表 `sources.TEXT_REPAIRS` 精确修补，构建时断言必须命中恰好一次，否则构建失败：

| 经节 | 源文件 | 修正为 | 依据 |
|---|---|---|---|
| Jer 22:24 | `“evenif you, Coniah` | `“even if you, Coniah` | `bsb.txt` |
| Jer 29:19 | `LORD,which I sent` | `LORD, which I sent` | `bsb.txt` |

**脚注、交叉引用、Strong 编号、`\add` 补充词标记、`\wj` 耶稣语录标记、`\pn` 专名标记**都按「标记去壳、文字保留」处理：包裹词只保留词本身（KJV 的补充词因此不带方括号，和合本的补充词也不带方括号，而出版社的纯文本导出用 `[…]` 标出），脚注与交叉引用整段移除，Strong 编号的 `|strong="H0430"` 属性丢弃。

和合本还有一点与其他三本不同：表示敬畏的**全角空格 U+3000**（`起初，　神创造天地。`）是原书的排法，属于正文而非空白噪声，因此原样保留（只去掉行首行尾的空白）。校验时用 `quote()` 可以看到 `\u3000`。

经过上述处理，四个译本**逐节**与出版社纯文本导出完全一致：

| 译本 | 完全一致 | 仅排版间距差异 | 纯文本导出把诗篇题注并入了经节 | 无法解释的差异 |
|---|---:|---:|---:|---:|
| BSB | 29,846 | 1,256 | 0 | **0** |
| WEBP | 30,961 | 4 | 138 | **0** |
| KJV | 15,200 | 15,786 | 116 | **0** |
| CUV | 27,962 | 2,943 | 116 | **0** |

KJV 的「仅排版间距」比例高，是因为其纯文本导出用 `[was]` 标补充词、用 `¶` 标段落，而 USFM 用 `\add` 与段落标记；剥掉这两者后文字完全一致。CUV 同理：导出把 `\add` 补充词写成 `[造]`、把诗篇题注并进第 1 节。

### 5.5 保留未改的源文件瑕疵

**和合本有 55 节在汉字与标点之间多一个空格**，来自源文件的 `\pn 希伯 \pn*、`（专名标记的正文里带了一个尾随空格），呈现为 `希伯 、幔利`。其中 48 节与出版社自己的纯文本导出逐字相同，7 节的导出没有这个空格；本库按 USFM 原样保留，没有替出版社「订正」，以免与作为校验基准的导出产生新的差异。

需要更「好看」的显示时，建议在展示层处理（例：把 ` 、` 渲染成 `、`），不要在数据库里改字。

另外，和合本**标点后普遍带一个半角空格**（`起初，　神创造天地。`、`那人说： 这是我骨中的骨， 肉中的肉`），这是 eBible 排版该文本时的写法，出版社的纯文本导出完全一致，因此原样保留；这不是转换错误，`LIKE`/全文检索都不受影响。

---

## 6. 重建流程

```bash
# 1) 下载（唯一需要网络的步骤；已下过且 sha256 未变的文件会跳过）
python3 scripts/download_sources.py            # 需要代理时加 --proxy 127.0.0.1:7897

# 2) 构建（可重复执行，先写 .building 临时文件，全部检查通过才替换正式文件）
python3 scripts/build_database.py

# 3) 校验 + 生成报告
python3 scripts/verify_database.py
```

`build_database.py` 另外支持 `--db PATH` 指定输出位置；`verify_database.py` 支持 `--samples N` 调整随机抽样节数。

构建脚本的核心断言：66 卷齐全、每卷章数等于正典章数、章内节号 1..N 连续、无重复 `(译本, 卷, 章, 节)`、修复规则恰好命中一次、`PRAGMA integrity_check` 与 `foreign_key_check` 通过。任一失败则退出非零且不覆盖已有数据库。

---

## 7. 校验报告包含什么

`data/reports/validation_report.md`（同内容另有 JSON）分 9 组，共 83 项：

- **A** 表结构、`integrity_check`、外键、行数、FTS 覆盖
- **B** 66 卷 / 卷序 / 每卷章数 / 每译本 1189 章
- **C** 每章节号 1..N 连续、无重复、空节逐条列出、本版不编号的节号（与出版社导出核对）
- **D** 正文里没有 USFM 标记、脚注残留、pilcrow、双空格、控制字符；全部 NFC
- **E** 授权/URL/下载日期齐全，且 `sha256` 与磁盘上的压缩包一致
- **F** 跨译本覆盖：31,107 个经节中 31,088 个四译本共有；WEBP 独有 Rom 14:24–26，CUV 独有 3JN 1:15、REV 12:18；5 章节数不同，且「本版不编号的经节」逐条与记录比对
- **G** 逐节对照出版社纯文本导出（见 §5.4 表）
- **H** 12 个固定名节 + 8 个固定种子随机节，四译本全文比对
- **I** 文档里写的每条查询实际跑一遍

---

## 8. 再分发注意

- 四个译本均为公有领域，可自由再分发；但 **WEBP 不得改动后仍以 "World English Bible"/"WEB" 名义发布**。
- 英国境内印刷 KJV 受 Crown letters patent 限制。
- 本库保留了每译本的 `copyright` / `license` / `license_url` 字段，随数据一同分发即满足署名与授权说明要求。
