# 词典构建数据（build inputs）

这里放 `app/internal/dict/dict.bin` 的源文件和构建脚本。运行时只读取内嵌的
`dict.bin`，不会访问本目录；这些文件只在重建词典时需要。

## 文件

| 文件 | 用途 | 大小 |
|---|---|---|
| `build_dict.py` | 构建脚本，合并全部来源并写出 `dict.bin` | — |
| `open-dict.jsonl.gz` | Open Dictionary v2.0 全量发布物 | 93 MB |
| `acai-names.json` | 从 ACAI 提取的英文首选名 → 中文首选名 | 106 KB |
| `cc-cedict-1.1.1.tgz` | CC-CEDICT 原始包 | 6.9 MB |
| `cedict.js` | 由 CC-CEDICT 生成的中文索引 | 11 MB |
| `wordnet-db.tgz`、`package/dict/` | WordNet 3.0 数据库 | 10 MB |
| `cmudict.dict` | CMU Pronouncing Dictionary | 3.5 MB |
| `freq-en.txt` | FrequencyWords `en_50k` 词频 | 409 KB |
| `ecdict-0.0.4.tgz` | 考察过但**未使用**的 ECDICT | 22 MB |

`open-dict.jsonl.gz` 与 `acai-names.json` 是构建输入，不随发布物分发；发布物只有
`app/internal/dict/dict.bin`。

## 重建

```sh
# 需要 data/db/bible-en-study.db 已存在
python3 data/lex/build_dict.py            # 默认 -vocab all
python3 data/lex/build_dict.py -vocab bible
python3 data/lex/build_dict.py -vocab top10k
python3 scripts/build_fonts.py            # 词典新增中文后重新做字体子集
cd app && go test ./...
```

`OPEN_DICT=/path/to/open-dict.jsonl.gz` 可以覆盖 Open Dictionary 路径。

输出是 `LEXI2`：

```
"LEXI2\n" | uint32 压缩索引长度 | zlib(JSON index) | payload
index = {"words": [...], "forms": {form: lemma}, "entries": {word: [off, len]}}
entry = zlib(JSON {"w","pos","uk","s":[...],"zh":[...],"c":[...],"n","r","lemma"})
```

当前产物：41,873 条 / 10,113 forms / 7,095,178 字节；压缩索引 544,280 字节，
payload 6,550,888 字节。索引压缩后查询仍是一次比较加一次 inflate。

## 数据来源与许可

| 来源 | 用途 | 许可 |
|---|---|---|
| [Open Dictionary](https://github.com/ahpxex/open-dictionary) | 中文学习者释义、词性、词形、IPA | 数据 CC BY-SA 4.0；代码 MIT |
| [ACAI](https://github.com/BibleAquifer/ACAI) | 圣经人物/地点/实体的中文首选名 | CC BY-SA 4.0 |
| [CC-CEDICT](https://cc-cedict.org/) | 中文候选、专名回退；按和合本词频排序 | CC BY-SA 4.0 |
| [WordNet 3.0](https://wordnet.princeton.edu/) | 英文释义、词性、义项数 | Princeton WordNet License |
| [CMU Pronouncing Dictionary](http://www.speech.cs.cmu.edu/cgi-bin/cmudict) | 音标（ARPAbet → IPA） | BSD 风格 |
| [FrequencyWords](https://github.com/hermitdave/FrequencyWords) | 词频、基础词表 | CC BY 4.0 |

### 许可边界

- **代码**（`build_dict.py`、`app/` 下的 Go 代码）可以继续用项目自己的许可发布。
- **`dict.bin` 是 Open Dictionary、ACAI、CC-CEDICT 的衍生数据**，必须按
  **CC BY-SA 4.0** 分发：保留署名、标明修改、并以相同许可共享数据部分。
- Open Dictionary 的代码许可（MIT）不覆盖数据；`LICENSE-DATA.md` 已明确数据工件为
  CC BY-SA 4.0。不能把 `dict.bin` 当成 MIT 数据。
- 若发布仓库同时包含代码与数据，推荐在根 `LICENSE` 之外单列
  `data/lex/NOTICE.md` 或 `LICENSE-DATA`，说明「代码 X 许可，词典数据 CC BY-SA 4.0」。
- ECDICT 未使用：其仓库代码是 MIT，但数据来自 `cdict-1.0-1.rpm` 等来源，上游许可不明确。

完整署名见 [NOTICE.md](NOTICE.md)。
