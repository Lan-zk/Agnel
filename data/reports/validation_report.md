# 圣经数据库校验报告

生成时间 2026-11-08 14:53（UTC），生成脚本 `scripts/verify_database.py`。

**数据库：** `data/db/bible-en-study.db`（33,443,840 字节）

**结果：83 项检查，通过 83 项，失败 0 项。**

## 译本概况

| 译本 | 名称 | 版本 | 经节 | 其中空节 | 书卷 | 章节 | 授权 |
|----|------|---------|-------:|------:|------:|---------:|---------|
| BSB | Berean Standard Bible | Updated to the 3rd printing (USFM export) | 31,102 | 16 | 66 | 1189 | Public Domain |
| CUV | Chinese Union Version (simplified) | 新标点和合本 (Chinese Union New Punctuation), eBible.org text of 2021-10-15 | 31,092 | 71 | 66 | 1189 | Public Domain |
| KJV | King James Version | 1769 standardised text (eng-kjv2006) | 31,102 | 0 | 66 | 1189 | Public Domain |
| WEBP | World English Bible, Protestant Edition | 2020 stable text edition | 31,103 | 5 | 66 | 1189 | Public Domain |

## 校验明细

### A. 表结构、完整性与行数

| 结果 | 校验项 | 详情 |
|:-------|:------|:-------|
| 通过 | PRAGMA integrity_check（库完整性） | ok |
| 通过 | PRAGMA foreign_key_check（外键） | 无违反 |
| 通过 | 核心表齐全 | 7 张表全部存在 |
| 通过 | 视图齐全 | v_verse, v_verse_parallel |
| 通过 | verses 行数与 translations 表声明一致 | 124399 行 / 声明 124399 行 |
| 通过 | 全文索引覆盖每一节 | 已索引 124399 / 共 124399 节 |
| 通过 | dataset_meta 已写入 | {'schema_version': '1.0', 'built_at': '2026-11-06T14:09:02Z', 'generator': 'scripts/build_database.py', 'canon': '66 books, 1189 chapters', 'translation_ids': 'BSB,WEBP,KJV,CUV', 'verse_identity': '(translation_id, book_id, chapter, verse); never a line number'} |

### B. 正典完整性（66 卷 / 1189 章）

| 结果 | 校验项 | 详情 |
|:-------|:------|:-------|
| 通过 | BSB：66 卷齐全 | 共 66 卷 |
| 通过 | BSB：卷序与各卷章数符合正典 | book_id 1..66 按正典顺序，各卷章数与公认正典一致 |
| 通过 | BSB：共 1189 章 | 1189 章（旧约 929 + 新约 260） |
| 通过 | BSB：各卷章数 | 66 卷的章数全部与正典一致 |
| 通过 | CUV：66 卷齐全 | 共 66 卷 |
| 通过 | CUV：卷序与各卷章数符合正典 | book_id 1..66 按正典顺序，各卷章数与公认正典一致 |
| 通过 | CUV：共 1189 章 | 1189 章（旧约 929 + 新约 260） |
| 通过 | CUV：各卷章数 | 66 卷的章数全部与正典一致 |
| 通过 | KJV：66 卷齐全 | 共 66 卷 |
| 通过 | KJV：卷序与各卷章数符合正典 | book_id 1..66 按正典顺序，各卷章数与公认正典一致 |
| 通过 | KJV：共 1189 章 | 1189 章（旧约 929 + 新约 260） |
| 通过 | KJV：各卷章数 | 66 卷的章数全部与正典一致 |
| 通过 | WEBP：66 卷齐全 | 共 66 卷 |
| 通过 | WEBP：卷序与各卷章数符合正典 | book_id 1..66 按正典顺序，各卷章数与公认正典一致 |
| 通过 | WEBP：共 1189 章 | 1189 章（旧约 929 + 新约 260） |
| 通过 | WEBP：各卷章数 | 66 卷的章数全部与正典一致 |

### C. 经节连续性、重复与空节

| 结果 | 校验项 | 详情 |
|:-------|:------|:-------|
| 通过 | BSB：每章节号 1..N 连续无缺口 | 共 1189 章，首节为 1 |
| 通过 | BSB：无重复经节 | 唯一索引生效 |
| 通过 | BSB：空经节已核对 | 空 16 节，声明 16 节：MAT 17:21, MAT 18:11, MAT 23:14, MRK 7:16, MRK 9:44, MRK 9:46, MRK 11:26, MRK 15:28, LUK 17:36, LUK 23:17, JHN 5:4, ACT 8:37, ACT 15:34, ACT 24:7, ACT 28:29, ROM 16:24 |
| 通过 | CUV：每章节号 1..N 连续无缺口 | 共 1189 章，首节为 1；11 个节号本版不编号（出版社导出同样没有）：MAT 18:11、MAT 23:14、MRK 7:16、MRK 15:28、LUK 17:36、LUK 23:17、JHN 5:4、ACT 8:37、ACT 15:34、ACT 24:7、ACT 28:29 |
| 通过 | CUV：无重复经节 | 唯一索引生效 |
| 通过 | CUV：空经节已核对 | 空 71 节，声明 71 节：GEN 24:30, EXO 38:15, LEV 3:15, NUM 1:21, NUM 1:23, NUM 1:25, NUM 1:27, NUM 1:29, NUM 1:31, NUM 1:33, NUM 1:35, NUM 1:37, NUM 1:39, NUM 1:41, NUM 1:43, NUM 1:46, NUM 4:40, NUM 4:44, NUM 4:48, DEU 4:33, DEU 13:13, DEU 13:18, DEU 15:5, DEU 28:14, DEU 28:59, DEU 29:11, DEU 30:10, JOS 3:11, JOS 21:10, JDG 20:28, 1SA 25:31, 2SA 3:10, 1CH 4:18, 1CH 16:13, 1CH 16:40, 2CH 30:19, NEH 12:29, EST 1:14, EST 8:12, JOB 10:21, JOB 38:38, JOB 38:40, PSA 8:7, PSA 8:8, PSA 49:9, PSA 63:6, PSA 76:9, PSA 92:3, PSA 105:6, PSA 116:19, PSA 135:2, PRO 23:32, PRO 26:19, ECC 7:28, ISA 4:4, ISA 65:7, JER 34:9, JER 34:19, JER 43:6, EZK 16:57, EZK 18:11, EZK 39:27, HOS 12:5, ZEP 2:2, LUK 1:2, LUK 2:35, LUK 21:30, ACT 1:25, ACT 4:30, EPH 6:3, COL 2:21 |
| 通过 | KJV：每章节号 1..N 连续无缺口 | 共 1189 章，首节为 1 |
| 通过 | KJV：无重复经节 | 唯一索引生效 |
| 通过 | KJV：空经节已核对 | 空 0 节，声明 0 节 |
| 通过 | WEBP：每章节号 1..N 连续无缺口 | 共 1189 章，首节为 1 |
| 通过 | WEBP：无重复经节 | 唯一索引生效 |
| 通过 | WEBP：空经节已核对 | 空 5 节，声明 5 节：LUK 17:36, ACT 8:37, ACT 15:34, ACT 24:7, ROM 16:25 |

### D. 正文清洁度

| 结果 | 校验项 | 详情 |
|:-------|:------|:-------|
| 通过 | BSB：无 USFM 标记、脚注残留或空白缺陷 | 扫描 31102 节，全部干净 |
| 通过 | BSB：文本为 NFC 规范化 | 全部符合 NFC |
| 通过 | CUV：无 USFM 标记、脚注残留或空白缺陷 | 扫描 31092 节，全部干净 |
| 通过 | CUV：文本为 NFC 规范化 | 全部符合 NFC |
| 通过 | KJV：无 USFM 标记、脚注残留或空白缺陷 | 扫描 31102 节，全部干净 |
| 通过 | KJV：文本为 NFC 规范化 | 全部符合 NFC |
| 通过 | WEBP：无 USFM 标记、脚注残留或空白缺陷 | 扫描 31103 节，全部干净 |
| 通过 | WEBP：文本为 NFC 规范化 | 全部符合 NFC |

### E. 来源与授权信息

| 结果 | 校验项 | 详情 |
|:-------|:------|:-------|
| 通过 | BSB：来源字段完整 | Berean Standard Bible \| Updated to the 3rd printing (USFM export) \| https://bereanbible.com/bsb_usfm.zip \| 2026-11-06T12:34:30Z |
| 通过 | BSB：sha256 与下载的压缩包一致 | f0c56438eb89c98e...（bsb_usfm.zip） |
| 通过 | CUV：来源字段完整 | Chinese Union Version (simplified) \| 新标点和合本 (Chinese Union New Punctuation), eBible.org text of 2021-10-15 \| https://eBible.org/Scriptures/cmn-cu89s_usfm.zip \| 2026-11-06T14:06:32Z |
| 通过 | CUV：sha256 与下载的压缩包一致 | 496f17f4daa62ab1...（cmn-cu89s_usfm.zip） |
| 通过 | KJV：来源字段完整 | King James Version \| 1769 standardised text (eng-kjv2006) \| https://eBible.org/Scriptures/eng-kjv2006_usfm.zip \| 2026-11-06T12:53:37Z |
| 通过 | KJV：sha256 与下载的压缩包一致 | 069646cd8722f6d5...（eng-kjv2006_usfm.zip） |
| 通过 | WEBP：来源字段完整 | World English Bible, Protestant Edition \| 2020 stable text edition \| https://eBible.org/Scriptures/engwebp_usfm.zip \| 2026-11-06T12:34:40Z |
| 通过 | WEBP：sha256 与下载的压缩包一致 | a93df5571009fed5...（engwebp_usfm.zip） |

### F. 跨译本覆盖与经节差异

| 结果 | 校验项 | 详情 |
|:-------|:------|:-------|
| 通过 | 各译本共有的经节 | 31107 处经节中有 31088 处 4 个译本都有 |
| 通过 | BSB：其他译本没有的经节 | 0 处：无 |
| 通过 | CUV：其他译本没有的经节 | 2 处：3JN 1:15, REV 12:18 |
| 通过 | KJV：其他译本没有的经节 | 0 处：无 |
| 通过 | WEBP：其他译本没有的经节 | 3 处：ROM 14:24, ROM 14:25, ROM 14:26 |
| 通过 | BSB：本版不编号的经节与记录一致 | 5 处：3JN 1:15、REV 12:18 等版本差异 |
| 通过 | CUV：本版不编号的经节与记录一致 | 15 处：3JN 1:15、REV 12:18 等版本差异 |
| 通过 | KJV：本版不编号的经节与记录一致 | 5 处：3JN 1:15、REV 12:18 等版本差异 |
| 通过 | WEBP：本版不编号的经节与记录一致 | 4 处：3JN 1:15、REV 12:18 等版本差异 |
| 通过 | 罗马书颂赞的节号与出版社一致 | WEBP 有 14:24-26，KJV 第 14 章到 23 节为止，颂赞在 16:25-27 |
| 通过 | 章节数差异只出现在已记录的版本差异处 | John 第 7 章 → BSB:53 CUV:52 KJV:53 WEBP:53；Rom 第 14 章 → BSB:23 CUV:23 KJV:23 WEBP:26；Rom 第 16 章 → BSB:27 CUV:27 KJV:27 WEBP:25；3John 第 1 章 → BSB:14 CUV:15 KJV:14 WEBP:14；Rev 第 12 章 → BSB:17 CUV:18 KJV:17 WEBP:17 |
| 通过 | 已知版本差异已逐条说明 | 3John 1：和合本把英文版的第 14 节拆为 14、15 两节；John 7：和合本不编号 7:53，其内容并入 8:1「于是各人都回家去了」；Rev 12：和合本把英文版 13:1 的首句「那时龙就站在海边的沙上」编为 12:18；Rom 14：WEBP 把颂赞编为 14:24-26，BSB/KJV/CUV 编为 16:25-27；Rom 16：WEBP 的罗马书 16 章止于第 25 节，颂赞不在本章 |
| 通过 | 没有为对齐其他译本而重新编号 | Rom 16:27 只出现在 BSB、CUV、KJV |

### G. 与出版社纯文本导出逐节比对

| 结果 | 校验项 | 详情 |
|:-------|:------|:-------|
| 通过 | BSB：每一节都与出版社纯文本导出一致 | 29846 节完全相同，1256 节仅标点/空白间距不同，0 节因导出把诗篇题注或 119 篇字母并入经节而不同，0 节无法解释 |
| 通过 | BSB：没有经节在纯文本导出中缺失 | 无缺失 |
| 通过 | CUV：每一节都与出版社纯文本导出一致 | 27962 节完全相同，2943 节仅标点/空白间距不同，116 节因导出把诗篇题注或 119 篇字母并入经节而不同，0 节无法解释 |
| 通过 | CUV：没有经节在纯文本导出中缺失 | 无缺失；71 个合并编号的空节不在导出中 |
| 通过 | KJV：每一节都与出版社纯文本导出一致 | 15200 节完全相同，15786 节仅标点/空白间距不同，116 节因导出把诗篇题注或 119 篇字母并入经节而不同，0 节无法解释 |
| 通过 | KJV：没有经节在纯文本导出中缺失 | 无缺失 |
| 通过 | WEBP：每一节都与出版社纯文本导出一致 | 30961 节完全相同，4 节仅标点/空白间距不同，138 节因导出把诗篇题注或 119 篇字母并入经节而不同，0 节无法解释 |
| 通过 | WEBP：没有经节在纯文本导出中缺失 | 无缺失 |

### H. 固定名节与随机抽样

| 结果 | 校验项 | 详情 |
|:-------|:------|:-------|
| 通过 | 抽样经节与出版社导出一致 | 12 处固定名节 + 8 处随机节 × 4 个译本，逐字比对 |

### I. 查询接口实测

| 结果 | 校验项 | 详情 |
|:-------|:------|:-------|
| 通过 | 查询：按译本+卷+章+节取一节 | John 3:16 BSB → For God so loved the world that He gave His one and only … |
| 通过 | 查询：整章列出（按节号排序） | 诗篇 23 篇 KJV → 6 节，第 1 节为 'The LORD is my shepherd; I shal…' |
| 通过 | 查询：章内节号范围 | 罗马书 14 章 WEBP → 第 1..26 节 |
| 通过 | 查询：某译本的书卷数 | WEBP → 66 卷 |
| 通过 | 查询：某章的标题条数 | 诗篇 119 篇 BSB → 45 条标题（类型示例 acrostic） |
| 通过 | 查询：三译本经文对照视图 | John 3:16 → 3 列中有 3 列非空 |
| 通过 | 查询：对照视图：经节归属不足三译本 | 罗马书 14 章 → 3 处经节不足三个译本收录 |
| 通过 | 查询：全文检索 | 160 节含 shepherd |
| 通过 | 查询：全文检索 + 联接过滤 | 2 个译本含短语 still waters |
| 通过 | 查询：译本授权与下载日期 | BSB Public Domain（2026-11-06）；CUV Public Domain（2026-11-06）；KJV Public Domain（2026-11-06）；WEBP Public Domain（2026-11-06） |

## 各卷章数与节数

| # | 书卷 | OSIS | 章数 | BSB | CUV | KJV | WEBP |
|---:|------|------|---------:|---:|---:|---:|---:|
| 1 | Genesis | Gen | 50 | 50 章 / 1,533 节 | 50 章 / 1,533 节 | 50 章 / 1,533 节 | 50 章 / 1,533 节 |
| 2 | Exodus | Exod | 40 | 40 章 / 1,213 节 | 40 章 / 1,213 节 | 40 章 / 1,213 节 | 40 章 / 1,213 节 |
| 3 | Leviticus | Lev | 27 | 27 章 / 859 节 | 27 章 / 859 节 | 27 章 / 859 节 | 27 章 / 859 节 |
| 4 | Numbers | Num | 36 | 36 章 / 1,288 节 | 36 章 / 1,288 节 | 36 章 / 1,288 节 | 36 章 / 1,288 节 |
| 5 | Deuteronomy | Deut | 34 | 34 章 / 959 节 | 34 章 / 959 节 | 34 章 / 959 节 | 34 章 / 959 节 |
| 6 | Joshua | Josh | 24 | 24 章 / 658 节 | 24 章 / 658 节 | 24 章 / 658 节 | 24 章 / 658 节 |
| 7 | Judges | Judg | 21 | 21 章 / 618 节 | 21 章 / 618 节 | 21 章 / 618 节 | 21 章 / 618 节 |
| 8 | Ruth | Ruth | 4 | 4 章 / 85 节 | 4 章 / 85 节 | 4 章 / 85 节 | 4 章 / 85 节 |
| 9 | 1 Samuel | 1Sam | 31 | 31 章 / 810 节 | 31 章 / 810 节 | 31 章 / 810 节 | 31 章 / 810 节 |
| 10 | 2 Samuel | 2Sam | 24 | 24 章 / 695 节 | 24 章 / 695 节 | 24 章 / 695 节 | 24 章 / 695 节 |
| 11 | 1 Kings | 1Kgs | 22 | 22 章 / 816 节 | 22 章 / 816 节 | 22 章 / 816 节 | 22 章 / 816 节 |
| 12 | 2 Kings | 2Kgs | 25 | 25 章 / 719 节 | 25 章 / 719 节 | 25 章 / 719 节 | 25 章 / 719 节 |
| 13 | 1 Chronicles | 1Chr | 29 | 29 章 / 942 节 | 29 章 / 942 节 | 29 章 / 942 节 | 29 章 / 942 节 |
| 14 | 2 Chronicles | 2Chr | 36 | 36 章 / 822 节 | 36 章 / 822 节 | 36 章 / 822 节 | 36 章 / 822 节 |
| 15 | Ezra | Ezra | 10 | 10 章 / 280 节 | 10 章 / 280 节 | 10 章 / 280 节 | 10 章 / 280 节 |
| 16 | Nehemiah | Neh | 13 | 13 章 / 406 节 | 13 章 / 406 节 | 13 章 / 406 节 | 13 章 / 406 节 |
| 17 | Esther | Esth | 10 | 10 章 / 167 节 | 10 章 / 167 节 | 10 章 / 167 节 | 10 章 / 167 节 |
| 18 | Job | Job | 42 | 42 章 / 1,070 节 | 42 章 / 1,070 节 | 42 章 / 1,070 节 | 42 章 / 1,070 节 |
| 19 | Psalms | Ps | 150 | 150 章 / 2,461 节 | 150 章 / 2,461 节 | 150 章 / 2,461 节 | 150 章 / 2,461 节 |
| 20 | Proverbs | Prov | 31 | 31 章 / 915 节 | 31 章 / 915 节 | 31 章 / 915 节 | 31 章 / 915 节 |
| 21 | Ecclesiastes | Eccl | 12 | 12 章 / 222 节 | 12 章 / 222 节 | 12 章 / 222 节 | 12 章 / 222 节 |
| 22 | Song of Solomon | Song | 8 | 8 章 / 117 节 | 8 章 / 117 节 | 8 章 / 117 节 | 8 章 / 117 节 |
| 23 | Isaiah | Isa | 66 | 66 章 / 1,292 节 | 66 章 / 1,292 节 | 66 章 / 1,292 节 | 66 章 / 1,292 节 |
| 24 | Jeremiah | Jer | 52 | 52 章 / 1,364 节 | 52 章 / 1,364 节 | 52 章 / 1,364 节 | 52 章 / 1,364 节 |
| 25 | Lamentations | Lam | 5 | 5 章 / 154 节 | 5 章 / 154 节 | 5 章 / 154 节 | 5 章 / 154 节 |
| 26 | Ezekiel | Ezek | 48 | 48 章 / 1,273 节 | 48 章 / 1,273 节 | 48 章 / 1,273 节 | 48 章 / 1,273 节 |
| 27 | Daniel | Dan | 12 | 12 章 / 357 节 | 12 章 / 357 节 | 12 章 / 357 节 | 12 章 / 357 节 |
| 28 | Hosea | Hos | 14 | 14 章 / 197 节 | 14 章 / 197 节 | 14 章 / 197 节 | 14 章 / 197 节 |
| 29 | Joel | Joel | 3 | 3 章 / 73 节 | 3 章 / 73 节 | 3 章 / 73 节 | 3 章 / 73 节 |
| 30 | Amos | Amos | 9 | 9 章 / 146 节 | 9 章 / 146 节 | 9 章 / 146 节 | 9 章 / 146 节 |
| 31 | Obadiah | Obad | 1 | 1 章 / 21 节 | 1 章 / 21 节 | 1 章 / 21 节 | 1 章 / 21 节 |
| 32 | Jonah | Jonah | 4 | 4 章 / 48 节 | 4 章 / 48 节 | 4 章 / 48 节 | 4 章 / 48 节 |
| 33 | Micah | Mic | 7 | 7 章 / 105 节 | 7 章 / 105 节 | 7 章 / 105 节 | 7 章 / 105 节 |
| 34 | Nahum | Nah | 3 | 3 章 / 47 节 | 3 章 / 47 节 | 3 章 / 47 节 | 3 章 / 47 节 |
| 35 | Habakkuk | Hab | 3 | 3 章 / 56 节 | 3 章 / 56 节 | 3 章 / 56 节 | 3 章 / 56 节 |
| 36 | Zephaniah | Zeph | 3 | 3 章 / 53 节 | 3 章 / 53 节 | 3 章 / 53 节 | 3 章 / 53 节 |
| 37 | Haggai | Hag | 2 | 2 章 / 38 节 | 2 章 / 38 节 | 2 章 / 38 节 | 2 章 / 38 节 |
| 38 | Zechariah | Zech | 14 | 14 章 / 211 节 | 14 章 / 211 节 | 14 章 / 211 节 | 14 章 / 211 节 |
| 39 | Malachi | Mal | 4 | 4 章 / 55 节 | 4 章 / 55 节 | 4 章 / 55 节 | 4 章 / 55 节 |
| 40 | Matthew | Matt | 28 | 28 章 / 1,071 节 | 28 章 / 1,069 节 | 28 章 / 1,071 节 | 28 章 / 1,071 节 |
| 41 | Mark | Mark | 16 | 16 章 / 678 节 | 16 章 / 676 节 | 16 章 / 678 节 | 16 章 / 678 节 |
| 42 | Luke | Luke | 24 | 24 章 / 1,151 节 | 24 章 / 1,149 节 | 24 章 / 1,151 节 | 24 章 / 1,151 节 |
| 43 | John | John | 21 | 21 章 / 879 节 | 21 章 / 877 节 | 21 章 / 879 节 | 21 章 / 879 节 |
| 44 | Acts | Acts | 28 | 28 章 / 1,007 节 | 28 章 / 1,003 节 | 28 章 / 1,007 节 | 28 章 / 1,007 节 |
| 45 | Romans | Rom | 16 | 16 章 / 433 节 | 16 章 / 433 节 | 16 章 / 433 节 | 16 章 / 434 节 |
| 46 | 1 Corinthians | 1Cor | 16 | 16 章 / 437 节 | 16 章 / 437 节 | 16 章 / 437 节 | 16 章 / 437 节 |
| 47 | 2 Corinthians | 2Cor | 13 | 13 章 / 257 节 | 13 章 / 257 节 | 13 章 / 257 节 | 13 章 / 257 节 |
| 48 | Galatians | Gal | 6 | 6 章 / 149 节 | 6 章 / 149 节 | 6 章 / 149 节 | 6 章 / 149 节 |
| 49 | Ephesians | Eph | 6 | 6 章 / 155 节 | 6 章 / 155 节 | 6 章 / 155 节 | 6 章 / 155 节 |
| 50 | Philippians | Phil | 4 | 4 章 / 104 节 | 4 章 / 104 节 | 4 章 / 104 节 | 4 章 / 104 节 |
| 51 | Colossians | Col | 4 | 4 章 / 95 节 | 4 章 / 95 节 | 4 章 / 95 节 | 4 章 / 95 节 |
| 52 | 1 Thessalonians | 1Thess | 5 | 5 章 / 89 节 | 5 章 / 89 节 | 5 章 / 89 节 | 5 章 / 89 节 |
| 53 | 2 Thessalonians | 2Thess | 3 | 3 章 / 47 节 | 3 章 / 47 节 | 3 章 / 47 节 | 3 章 / 47 节 |
| 54 | 1 Timothy | 1Tim | 6 | 6 章 / 113 节 | 6 章 / 113 节 | 6 章 / 113 节 | 6 章 / 113 节 |
| 55 | 2 Timothy | 2Tim | 4 | 4 章 / 83 节 | 4 章 / 83 节 | 4 章 / 83 节 | 4 章 / 83 节 |
| 56 | Titus | Titus | 3 | 3 章 / 46 节 | 3 章 / 46 节 | 3 章 / 46 节 | 3 章 / 46 节 |
| 57 | Philemon | Phlm | 1 | 1 章 / 25 节 | 1 章 / 25 节 | 1 章 / 25 节 | 1 章 / 25 节 |
| 58 | Hebrews | Heb | 13 | 13 章 / 303 节 | 13 章 / 303 节 | 13 章 / 303 节 | 13 章 / 303 节 |
| 59 | James | Jas | 5 | 5 章 / 108 节 | 5 章 / 108 节 | 5 章 / 108 节 | 5 章 / 108 节 |
| 60 | 1 Peter | 1Pet | 5 | 5 章 / 105 节 | 5 章 / 105 节 | 5 章 / 105 节 | 5 章 / 105 节 |
| 61 | 2 Peter | 2Pet | 3 | 3 章 / 61 节 | 3 章 / 61 节 | 3 章 / 61 节 | 3 章 / 61 节 |
| 62 | 1 John | 1John | 5 | 5 章 / 105 节 | 5 章 / 105 节 | 5 章 / 105 节 | 5 章 / 105 节 |
| 63 | 2 John | 2John | 1 | 1 章 / 13 节 | 1 章 / 13 节 | 1 章 / 13 节 | 1 章 / 13 节 |
| 64 | 3 John | 3John | 1 | 1 章 / 14 节 | 1 章 / 15 节 | 1 章 / 14 节 | 1 章 / 14 节 |
| 65 | Jude | Jude | 1 | 1 章 / 25 节 | 1 章 / 25 节 | 1 章 / 25 节 | 1 章 / 25 节 |
| 66 | Revelation | Rev | 22 | 22 章 / 404 节 | 22 章 / 405 节 | 22 章 / 404 节 | 22 章 / 404 节 |

## 固定名节核验

### GEN 1:1

- **BSB** In the beginning God created the heavens and the earth.
- **CUV** 起初，　神创造天地。
- **KJV** In the beginning God created the heaven and the earth.
- **WEBP** In the beginning, God created the heavens and the earth.

### PSA 3:1

- **BSB** A Psalm of David, when he fled from his son Absalom. O LORD, how my foes have increased! How many rise up against me!
- **CUV** 耶和华啊，我的敌人何其加增； 有许多人起来攻击我。
- **KJV** LORD, how are they increased that trouble me! many are they that rise up against me.
- **WEBP** LORD, how my adversaries have increased! Many are those who rise up against me.

### PSA 23:1

- **BSB** A Psalm of David. The LORD is my shepherd; I shall not want.
- **CUV** 耶和华是我的牧者， 我必不致缺乏。
- **KJV** The LORD is my shepherd; I shall not want.
- **WEBP** The LORD is my shepherd; I shall lack nothing.

### PSA 119:1

- **BSB** Blessed are those whose way is blameless, who walk in the Law of the LORD.
- **CUV** 行为完全、遵行耶和华律法的， 这人便为有福！
- **KJV** Blessed are the undefiled in the way, who walk in the law of the LORD.
- **WEBP** Blessed are those whose ways are blameless, who walk according to the LORD’s law.

### JON 2:1

- **BSB** From inside the fish, Jonah prayed to the LORD his God,
- **CUV** 约拿在鱼腹中祷告耶和华—他的　神，
- **KJV** Then Jonah prayed unto the LORD his God out of the fish’s belly,
- **WEBP** Then Jonah prayed to the LORD, his God, out of the fish’s belly.

### MAT 6:9

- **BSB** So then, this is how you should pray: ‘Our Father in heaven, hallowed be Your name.
- **CUV** 所以，你们祷告要这样说： 我们在天上的父： 愿人都尊你的名为圣。
- **KJV** After this manner therefore pray ye: Our Father which art in heaven, Hallowed be thy name.
- **WEBP** Pray like this: “‘Our Father in heaven, may your name be kept holy.

### JHN 3:16

- **BSB** For God so loved the world that He gave His one and only Son, that everyone who believes in Him shall not perish but have eternal life.
- **CUV** 「　神爱世人，甚至将他的独生子赐给他们，叫一切信他的，不致灭亡，反得永生。
- **KJV** For God so loved the world, that he gave his only begotten Son, that whosoever believeth in him should not perish, but have everlasting life.
- **WEBP** For God so loved the world, that he gave his only born Son, that whoever believes in him should not perish, but have eternal life.

### ROM 8:1

- **BSB** Therefore, there is now no condemnation for those who are in Christ Jesus.
- **CUV** 如今，那些在基督耶稣里的就不定罪了。
- **KJV** There is therefore now no condemnation to them which are in Christ Jesus, who walk not after the flesh, but after the Spirit.
- **WEBP** There is therefore now no condemnation to those who are in Christ Jesus, who don’t walk according to the flesh, but according to the Spirit.

### ROM 16:25

- **BSB** Now to Him who is able to strengthen you by my gospel and by the proclamation of Jesus Christ, according to the revelation of the mystery concealed for ages past
- **CUV** 惟有　神能照我所传的福音和所讲的耶稣基督，并照永古隐藏不言的奥秘，坚固你们的心。
- **KJV** Now to him that is of power to stablish you according to my gospel, and the preaching of Jesus Christ, according to the revelation of the mystery, which was kept secret since the world began,
- **WEBP** _（本译本此节为空）_

### 1JN 5:7

- **BSB** For there are three that testify:
- **CUV** 并且有圣灵作见证，因为圣灵就是真理。
- **KJV** For there are three that bear record in heaven, the Father, the Word, and the Holy Ghost: and these three are one.
- **WEBP** For there are three who testify:

### REV 22:21

- **BSB** The grace of the Lord Jesus be with all the saints. Amen.
- **CUV** 愿主耶稣的恩惠常与众圣徒同在。阿们！
- **KJV** The grace of our Lord Jesus Christ be with you all. Amen.
- **WEBP** The grace of the Lord Jesus Christ be with all the saints. Amen.

### SNG 1:1

- **BSB** This is Solomon’s Song of Songs.
- **CUV** 所罗门的歌，是歌中的雅歌。
- **KJV** The song of songs, which is Solomon’s.
- **WEBP** The Song of songs, which is Solomon’s.

## 随机抽样（8 处，种子 20231106）

- **JDG 5:13**
  - BSB: Then the survivors came down to the nobles; the people of the LORD came down to me against the mighty.
  - CUV: 那时有余剩的贵胄和百姓一同下来； 耶和华降临，为我攻击勇士。
  - KJV: Then he made him that remaineth have dominion over the nobles among the people: the LORD made me have dominion over the mighty.
  - WEBP: “Then a remnant of the nobles and the people came down. The LORD came down for me against the mighty.
- **2CH 24:8**
  - BSB: At the king’s command a chest was made and placed outside, at the gate of the house of the LORD.
  - CUV: 于是王下令，众人做了一柜，放在耶和华殿的门外，
  - KJV: And at the king’s commandment they made a chest, and set it without at the gate of the house of the LORD.
  - WEBP: So the king commanded, and they made a chest, and set it outside at the gate of the LORD’s house.
- **NEH 2:12**
  - BSB: I set out at night with a few men. I did not tell anyone what my God had laid on my heart to do for Jerusalem. The only animal with me was the one on which I was riding.
  - CUV: 我夜间起来，有几个人也一同起来；但　神使我心里要为耶路撒冷做什么事，我并没有告诉人。除了我骑的牲口以外，也没有别的牲口在我那里。
  - KJV: And I arose in the night, I and some few men with me; neither told I any man what my God had put in my heart to do at Jerusalem: neither was there any beast with me, save the beast that I rode upon.
  - WEBP: I arose in the night, I and a few men with me. I didn’t tell anyone what my God put into my heart to do for Jerusalem. There wasn’t any animal with me except the animal that I rode on.
- **ISA 49:25**
  - BSB: Indeed, this is what the LORD says: “Even the captives of the mighty will be taken away, and the plunder of the tyrant will be retrieved; I will contend with those who contend with you, and I will save your children.
  - CUV: 但耶和华如此说： 就是勇士所掳掠的，也可以夺回； 强暴人所抢的，也可以解救。 与你相争的，我必与他相争； 我要拯救你的儿女。
  - KJV: But thus saith the LORD, Even the captives of the mighty shall be taken away, and the prey of the terrible shall be delivered: for I will contend with him that contendeth with thee, and I will save thy children.
  - WEBP: But the LORD says, “Even the captives of the mighty shall be taken away, and the plunder retrieved from the fierce, for I will contend with him who contends with you and I will save your children.
- **JER 19:7**
  - BSB: And in this place I will ruin the plans of Judah and Jerusalem. I will make them fall by the sword before their enemies, by the hands of those who seek their lives, and I will give their carcasses as food to the birds of the air and the beasts of the earth.
  - CUV: 我必在这地方使犹大和耶路撒冷的计谋落空，也必使他们在仇敌面前倒于刀下，并寻索其命的人手下。他们的尸首，我必给空中的飞鸟和地上的野兽作食物。
  - KJV: And I will make void the counsel of Judah and Jerusalem in this place; and I will cause them to fall by the sword before their enemies, and by the hands of them that seek their lives: and their carcases will I give to be meat for the fowls of the heaven, and for the beasts of the earth.
  - WEBP: “‘“I will make the counsel of Judah and Jerusalem void in this place. I will cause them to fall by the sword before their enemies, and by the hand of those who seek their life. I will give their dead bodies to be food for the birds of the sky and for the animals of the earth.
- **EZK 48:19**
  - BSB: The workers of the city who cultivate it will come from all the tribes of Israel.
  - CUV: 所有以色列支派中，在城内做工的，都要耕种这地。
  - KJV: And they that serve the city shall serve it out of all the tribes of Israel.
  - WEBP: Those who labor in the city, out of all the tribes of Israel, shall cultivate it.
- **HAG 2:16**
  - BSB: from that time, when one came expecting a heap of twenty ephahs of grain, there were but ten. When one came to the winepress to draw out fifty baths, there were but twenty.
  - CUV: 在那一切日子，有人来到谷堆，想得二十斗，只得了十斗。有人来到酒池，想得五十桶，只得了二十桶。
  - KJV: Since those days were, when one came to an heap of twenty measures, there were but ten: when one came to the pressfat for to draw out fifty vessels out of the press, there were but twenty.
  - WEBP: Through all that time, when one came to a heap of twenty measures, there were only ten. When one came to the wine vat to draw out fifty, there were only twenty.
- **JAS 2:11**
  - BSB: For He who said, “Do not commit adultery,” also said, “Do not murder.” If you do not commit adultery, but do commit murder, you have become a lawbreaker.
  - CUV: 原来那说「不可奸淫」的，也说「不可杀人」；你就是不奸淫，却杀人，仍是成了犯律法的。
  - KJV: For he that said, Do not commit adultery, said also, Do not kill. Now if thou commit no adultery, yet if thou kill, thou art become a transgressor of the law.
  - WEBP: For he who said, “Do not commit adultery,” also said, “Do not commit murder.” Now if you do not commit adultery but do murder, you have become a transgressor of the law.

## 校验方法

1. 每个译本的每一节都与**同一发布方**另外提供的纯文本导出逐条比对（BSB 用 bereanbible.com 的 `bsb.txt`；WEBP、KJV、CUV 用 eBible.org 的 `engwebp_vpl.txt`、`eng-kjv2006_vpl.txt`、`cmn-cu89s_vpl.txt`）。USFM 与纯文本由不同流水线生成，因此两者一致才算真正的外部核对。
2. 差异必须被归入可解释的类别：完全相同 / 仅破折号、括号与标点周围的排版间距不同 / 纯文本导出把诗篇题注或诗篇 119 篇的字母并入了经节。任何落不进这些类别的差异都会让该项失败。
3. 全流程以 `(translation_id, book_id, chapter, verse)` 作为经节身份，任何环节都不按行号或行序对齐。

## 已知限制

- BSB 有 16 节、WEBP 有 5 节正文为空，因为这两个版本本身就不收这些节；CUV 另有 71 节为空，对应和合本把两个英文节号印成一节（`\v 20-21`）的排法。三种情况下节号一律保留，使跨译本的经节引用保持稳定。
- CUV 另有 11 个节号完全不编号（如 MAT 18:11、ACT 8:37），源文件既无 `\v` 也无区间标记，出版社的纯文本导出同样没有这些节，因此库里没有对应行。
- 全书 1,189 章中有 5 章节数在译本间不同：ROM 14（WEBP 26 节，颂赞在本章末）、ROM 16（WEBP 25 节）、JHN 7（CUV 52 节，不收 7:53）、3JN 1（CUV 15 节）、REV 12（CUV 18 节）。任何一方都没有被重新编号；`v_verse_parallel` 的 `translation_count` 列会让这种差异显形。
- CUV 的 USFM 里有 55 节在汉字与标点之间带一个空格（源文件写作 `\pn 希伯 \pn*、`），其中 48 节与出版社的纯文本导出逐字相同，7 节的导出没有该空格。数据库按 USFM 原样保留，未做统一，以免与出版社导出产生新的差异。
- 段落起始（`verses.paragraph_start`）来自 `\p` 系列标记，KJV 另外来自原文中的段落号 `¶`；它只用于排版提示，不属于经文正文。
- 标题一律保存在 `sections`，从不并入 `verses.text`；唯一例外是 BSB 的诗篇题注，出版社自己把它排在诗篇第 1 节之内。
