# 圣经英语 · Agnel

面向中文母语者的英文《圣经》阅读与词汇学习桌面工具。打开即阅读；遇到生词点一下就地查，
后面用读过的文本做跟打、默写和填空。Go + [Viva](https://vivaui.dev) 原生窗口，
GPU 直接绘制，无 WebView；经文和词典离线内嵌，运行时不联网。

## 现在能做什么

现在能用的是阅读器与生词本：

- 左右对照：左页 BSB / WEBP / KJV 任选，右页和合本，逐节对齐；也可单栏。
- 66 卷分卷选章，中文 / English 卷名切换；书签、键盘翻页、记住上次位置。
- 点词查卡：词性、音标、中文义、英文释义、本书搭配、词形变化、出处与原句。
- 词典 41,873 条 / 10,113 forms，覆盖三译本 99.6% 的词次；没收录的词会明说「还没有收录」。
- 生词本：查词卡收藏、三栏词库（筛选 / 列表 / 详情）、三个阶段、错词计数、按词 / 中文 / 原句搜索、
  手动添加、从原句跳回经文；词库还能切到「全部词典」，41,873 条按字母排序，可按首字母跳转、按前缀
  搜索、直接收藏。练习与英语基础还没有页面。

## 跑起来

需要 macOS 12+ 和 Go 1.26。

```sh
cd app
go build -o build/agnel . && ./build/agnel   # 直接开窗
./run.sh dev                                 # viva dev .：热重载
go test ./...                                # 行为测试，不开窗
```

首次构建要拉依赖；`proxy.golang.org` 不通时用国内代理：

```sh
GOPROXY=https://goproxy.cn,direct go run .
```

不开窗、直接出真实 Metal / CoreText 像素：

```sh
cd app
go run -tags vivahead ./cmd/shot /tmp/a.png
go run -tags vivahead ./cmd/shot /tmp/a.png "label:全部书卷" "shot:/tmp/b.png"
go run -tags vivahead ./cmd/shot /tmp/a.png "locate:beginning" "shot:/tmp/c.png"
```

## 仓库结构

| 路径 | 内容 |
|---|---|
| `app/` | Go + Viva 应用；`internal/desk` 阅读器与词库，`internal/vocab` 生词本，`internal/dict` 词典，`internal/bible` 经文 |
| `data/` | SQLite 经文库、词典构建源与报告；数据说明见 `data/README.md` |
| `scripts/` | 下载、USFM 解析、建库、校验、字体子集化（Python 标准库） |
| `design/` | 产品概要说明书与设计资料（HTML 原型已删除） |
| `refer/` | 参考项目，只借鉴交互，不复制代码 |

运行时只读 `//go:embed` 进二进制的 `app/internal/bible/store.bin`、`app/internal/dict/dict.bin`，
以及用户目录下的 `state.json` 与 `vocab.json`（`~/Library/Application Support/bible-en-study/`）。
不联网、不读 SQLite、不写数据库。

## 数据与重建

```sh
python3 scripts/download_sources.py          # 唯一需要网络的步骤
python3 scripts/build_database.py
python3 scripts/verify_database.py           # 83 项检查
python3 data/lex/build_dict.py               # 默认 -vocab all
python3 scripts/build_fonts.py               # 词典新增中文后重做字体子集
```

词典格式、来源与压缩说明见 [data/lex/README.md](data/lex/README.md)。

## 许可

- **代码**（`app/`、`scripts/`、`data/lex/build_dict.py` 等）：[MIT](LICENSE)。
- **词典数据** `app/internal/dict/dict.bin`：**CC BY-SA 4.0**。它是 Open Dictionary、ACAI、
  CC-CEDICT 的衍生数据，分发时必须保留署名、说明修改，并以相同许可共享数据部分。
  完整署名与修改说明见 [data/lex/NOTICE.md](data/lex/NOTICE.md)。
- 经文（BSB / WEBP / KJV / 和合本）为公有领域；来源与授权逐项列在 `data/README.md`。
- WordNet、CMU Pronouncing Dictionary、FrequencyWords 的许可见 `data/lex/README.md`。

代码许可不覆盖词典数据；不能把 `dict.bin` 当作 MIT 数据分发。
