# 圣经英语 · 原生阅读器

`design/05-sidebar-rail-grot-text` 的原生版，Go + [Viva](https://vivaui.dev)：窗口由 GPU 直接绘制，
没有 WebView，也没有打包浏览器。经文与译本全部离线内嵌，不联网。

数据来自 `data/db/bible-en-study.db`：66 卷、1189 章，三个公有领域英译本
（BSB / WEBP / KJV）加和合本（CUV，简体）。右页固定和合本，左页可换英文译本。

## 跑

```sh
cd app
go build -o build/agnel . && ./build/agnel   # 直接开窗
./run.sh dev        # 热重载；内部走 viva dev，Dock 显示应用图标
viva dev .          # 直接跑 viva dev；Dock 显示宿主的通用图标
go test ./...       # 行为测试（不开窗）
```

首次构建要拉依赖，`proxy.golang.org` 不通时用国内代理：

```sh
GOPROXY=https://goproxy.cn,direct go run .
```

## 打包

```sh
viva package -target darwin -sign-identity -          # 本机自签，产出 build/Agnel.app
viva package -target darwin -format app,zip          # 顺带出分发包
```

产物是标准 bundle：`Contents/MacOS/agnel`、`Contents/Resources/AppIcon.icns`（由
`assets/icon.png` 生成）、`Contents/Info.plist`（`dev.agnel.reader`、`LSMinimumSystemVersion 12.0`）。
双击就能起，Dock 里是这枚标志。没有开发者证书时只能 ad-hoc 签名，拷到别的机器上会被 Gatekeeper 拦。

**一个必须记住的坑**：`Run()` 一开始就要 `runtime.LockOSThread()`。Viva 在开窗前会
`dispatch_sync` 到主队列去读「减弱动态效果」，如果 Go 的主 goroutine 这时漂到了别的线程上，
那次同步永远等不到，表现就是**进程活着、窗口不出来**（有时能出来但只有 136×145）。
钉住主线程之后，`./build/agnel` 与打包后的 .app 都能稳定开出 1440×932 的窗口。

## 标志

墨绿的字母 A，字怀镂出一只低头的羊与一本摊开的书；两色，`#668B60` 与 `#FCF7F5`。

- 素材是用户给的矢量图 `app/assets/logo.svg`（`viewBox 0 0 1095 1095`，两个 `path`：纸色方底 + 绿色字标），
  `data/logo/logo.svg` 与 HTML 原型 `design/05-sidebar-rail-grot-text/logo.svg` 是同一份。
- 应用里：顶栏用 `f.Image(logo.Bitmap()).Size(28, 28)` 画 `internal/logo/mark.png`（256²，透明背景，
  只有绿色字标；矢量源是 `app/assets/logo-mark.svg`；`mark.png` 用 `//go:embed` 编进二进制）。
  为什么用位图：这台机器上 `canvas.DrawPath`、`canvas.DrawBitmap` 以及 `viva icons` 编出的 `ui.Icon`
  都画不出东西（`DrawPath` 甚至会让窗口根本不出现），只有 flex 的 `Image` 稳定。
  界面里的翻页箭头、下拉提示因此也不用图标，而是字体里的 `‹` `›` `▾` 三个字形；
  根因与改进建议见 `docs/viva-framework-issues.md` §7。
- 应用图标：`app/assets/icon.png`（1024²，纸色方底 + 绿色字标）→ `sips` + `iconutil` 出
  `Contents/Resources/AppIcon.icns`，打进 `build/Agnel.app`；`viva.toml` 的 `[app] icon` 也指向它，
  `viva package` 会用同一张。同一张图还被 `app/assets/icon.go` 用 `//go:embed` 编进程序，
  `main.go` 解码后交给 `app.Config.Icon`，所以不经 bundle 运行 `./build/agnel` 时 Dock 也是它。
  直接运行 `viva dev .` 时窗口由裸宿主进程创建，Dock 只能显示通用图标；`./run.sh dev` 会把
  viva 宿主包进 `build/Agnel Dev.app` 再启动，Dock 因此也用这枚图标。
- 启动：`./run.sh`（打开 `build/Agnel.app`）；`./run.sh dev` 走 `viva dev .`（热重载）。
  某些时候（这台机器上偶发）原生窗口会以 136×145 这种尺寸出现，`viva dev` 不受影响。

## 数据

```sh
go run ./cmd/bibledata -db ../data/db/bible-en-study.db -out internal/bible/store.bin
```

生成器会调用 `sqlite3 -readonly -json` 读原库（4 个译本 / 124,399 节），按 (译本, 卷, 章) 分组，写成
`internal/bible/store.bin`（7.1 MB）：文件头 + JSON 索引 + 每章一段 zlib 压缩的 JSON
（节号、经文、段落标记、小标题）。运行时只 `//go:embed` 这一个文件，随机取章、16 章缓存，
不需要 SQLite 驱动、不需要 CGO、不往磁盘写。

为什么不直接内嵌那个 25 MB 的库：发布物要单文件、离线、无外部依赖；把库交给运行时读取，
还得带驱动并处理定位，收益不抵成本。

## 看

不开窗、直接出图（真实 Metal / CoreText 像素，便于对比改动）：

```sh
go run -tags vivahead ./cmd/shot /tmp/a.png
go run -tags vivahead ./cmd/shot /tmp/a.png "label:全部书卷" "shot:/tmp/b.png"
go run -tags vivahead ./cmd/shot /tmp/a.png "click:547,189" "shot:/tmp/c.png"
```

步骤：`click:x,y`、`label:文字`、`locate:文字`、`keys:spec`、`shot:路径`。`locate` 只认有角色的控件，
经文里的单词要用 `click:x,y`。

`docs/spread.png` 左英右中对照、`docs/lookup.png` 点词后的词条卡、`docs/picker.png` 分卷页、
`docs/picker-en.png` 分卷页英文态、`docs/chapters.png` 诗篇 150 章。

## 已经做到的

- **左右对照**：左页英文（BSB / WEBP / KJV 任选），右页和合本，逐节对齐。小标题按各译本自己的
  `sections` 打印（BSB "The First Day"、和合本「神的创造」），`paragraph_start` 出段距；页眉是各译本
  自己的卷名（"Psalm" / 「诗篇」），外上角页码，外缘「创/BSB」「诗/和合」书口标，节号挂节末。
  中文行距比英文松（27.5 vs 23.5），栏宽也更宽（436 vs 372），因为同样的高度中文容得下更多字。
- **单栏**：一个译本铺满一页（可选和合本）。
- **选择经文**：66 卷按 律法书 / 历史书 / 诗歌智慧书 / 大先知书 / 小先知书 / 福音书 /
  使徒行传 / 保罗书信 / 普通书信 / 启示录 分组；每卷标章数，多章进一层选章；顶栏 中文 / English
  切换卷名（与阅读页标题、左栏同步），选择记在 state.json。
- **翻页**：书页两侧边缘各一个箭头，不占正文宽度。翻到章末继续往前就进下一章的第一对开，
  在章首往回翻落到上一章的最后一对开，同理跨卷；到正典首尾才停。箭头的标签随位置变化
  （下一页 / 下一章 / 下一卷）。
- **键盘**：← → 翻页，空格同 →，Home / End 到头尾，Esc 关词条卡。
- **书签**：标记当前对开，页上缘挂书签带，左栏列出「卷 章:节」，点击跳回（按节重新定位）。
- **记住位置**：卷、章、页、模式、左右译本、语言、书签写在
  `~/Library/Application Support/bible-en-study/state.json`，下次打开回到原处。
- **查词**：点正文里的词弹卡片（词性、音标、中文义、英文释义、本书搭配、词形变化、出处与原句——
  对照模式下中英两句都给）；词典 41,873 条，没收录的词会明说「还没有收录」。

## 词典

```sh
python3 ../data/lex/build_dict.py                 # 默认 -vocab all：Open Dictionary 单词头 ∪ 三译本词形 ∪ ACAI 专名 ∪ 功能词表
python3 ../data/lex/build_dict.py -vocab bible    # 只做经文里出现的词
```

产物 `internal/dict/dict.bin`（7.1 MB / 41,873 条 / 10,113 forms）：`"LEXI2\n" | uint32 压缩索引长度 |
zlib(索引 JSON) | 负载`。索引含词表、屈折表（form → lemma）与每条记录的偏移；索引和每条记录都单独 zlib。
查询是一次比较加一次 inflate，窗口不等磁盘。

数据来源与许可（全部随数据分发，不联网）：

| 来源 | 用途 | 许可 |
|---|---|---|
| [Open Dictionary](https://github.com/ahpxex/open-dictionary) | 中文学习者释义、词性、词形、IPA | 数据 CC BY-SA 4.0；代码 MIT |
| [ACAI](https://github.com/BibleAquifer/ACAI) | 圣经人物/地点/实体的中文首选名 | CC BY-SA 4.0 |
| [WordNet 3.0](https://wordnet.princeton.edu/) | 英文释义、词性、义项数 | Princeton WordNet License |
| [CMU Pronouncing Dictionary](http://www.speech.cs.cmu.edu/cgi-bin/cmudict) | 音标（ARPAbet → IPA） | BSD 风格 |
| [CC-CEDICT](https://cc-cedict.org/) | 中文候选、专名回退；按和合本词频排序 | CC BY-SA 4.0 |
| [FrequencyWords](https://github.com/hermitdave/FrequencyWords) | 词频、基础词表 | CC BY 4.0 |

没用的东西也记一笔：ECDICT 数据量大（63 MB）且数据上游许可不明确，不采用；kaikki/Wiktionary 最全，
但体量太大，只在构建期从 Open Dictionary 取子集，所以**词源一栏暂时没有**。
圣经专名的中文：ACAI 的 preferred label 优先，常用词仍有手工表（Abram 亚伯兰、Satan 撒但）；
没有可靠中文就不写。专名、KJV 古形和同形词（saw/rose/left/found）另有 `INFLECTION_FIRST` 与覆盖表。

`dict.bin` 是 CC BY-SA 4.0 数据（Open Dictionary + ACAI + CC-CEDICT 的衍生），代码许可以外的数据
不能按 MIT 分发；完整署名见 `data/lex/NOTICE.md`。`data/lex/` 里是构建源文件（约 150 MB），
不随仓库提交，按上面命令可重建。

## 还没做

中文译本（数据里没有，要用和合本得另进一批数据）、词库、练习、英语基础、全文搜索、
脚注与串珠（库里没有注释表）。

## 五个动手前必须知道的约束

1. **窗口必须用 `app.RunCanvas(config, &canvas)`**，不能用 `app.Run(config)`。后者在这台机器上
   会开出一个有标题、有背景色、但 `Paint` 一次都不调用的空窗口（排查：最小窗口探针 + 官方模板对照，
   `Paint` 里打点确认 0 次调用）。
2. **`ui.Theme` 必须从 `ui.LightTheme()` 拷**，只改颜色。手写 `ui.Theme{...}` 字面量会让
   `Metrics` 全为零，于是所有按钮的文字宽度为 0、量出来是 20×20 的空方块。
3. **一帧只能有一种比例字体**。`TextStyle` 没有字体族字段，`SetFontFamily` 只影响下一帧的排版；
   为此 `internal/font`（由 `../scripts/build_fonts.py` 生成）把 **Literata**（英文阅读衬线）、
   **STIX Two Text**（补齐 IPA 音标）与 **思源宋体 Source Han Serif SC**（按圣经/词典/UI 实际字表裁剪的汉字子集）
   在同一度量下合成为内嵌字体族 `Bible Literata`（Regular / Bold / Italic），启动时通过 `app.RegisterFont` 注册。
   Viva 框架层面的根因与可复用改进方案见 `docs/viva-framework-issues.md`。
4. **`LineHeight(1.5)` 是 1.5 像素**，不是倍数。要 1.5 倍行距得自己算像素，或干脆不设。
5. **`f.Slot()` 不给尺寸就是 0**；在 Row 里要 `.Grow(1)`，在 `Items(Center)` 的行里还要 `.Height(…)`。
   另外点击是「后声明的在前」：先声明整页的 `pageArea`，再声明每个词，词才点得到。
