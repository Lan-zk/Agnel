# 圣经英语 · 原生阅读器

Go + [Viva](https://vivaui.dev) 原生桌面应用：窗口由 GPU 直接绘制，
没有 WebView，也没有打包浏览器。经文与译本全部离线内嵌，不联网。

数据来自 `data/db/bible-en-study.db`：66 卷、1189 章，三个公有领域英译本
（BSB / WEBP / KJV）加和合本（CUV，简体）。右页固定和合本，左页可换英文译本。
左侧 68px 模块轨在阅读与词库之间切换；词库是本机的生词本，不联网。

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
  `data/logo/logo.svg` 是同一个矢量源的存档。
- 应用里：窗口标题条中线用 `f.Image(logo.Bitmap()).Size(15, 15)` 画 `internal/logo/mark.png`（256²，透明背景，
  只有绿色字标；矢量源是 `app/assets/logo-mark.svg`；`mark.png` 用 `//go:embed` 编进二进制）。
  为什么用位图：这台机器上 `canvas.DrawPath`、`canvas.DrawBitmap` 以及 `viva icons` 编出的 `ui.Icon`
  都画不出东西（`DrawPath` 甚至会让窗口根本不出现），只有 flex 的 `Image` 稳定。
  界面里的翻页箭头、下拉提示因此也不用图标，而是字体里的 `‹` `›` `▾` 三个字形；
  根因与改进建议见 `docs/viva-framework-issues.md` §7（窗口里三角形不落像素的根因已经定位：
  矩形图层快照只在部分事件路径里刷新，现已改为每帧刷新；`DrawPath` 没有回头再验证，图标仍用位图与字形）。
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

步骤：`click:x,y`、`label:文字`、`locate:文字`、`type:文字`、`size:w,h`、`keys:spec`、`shot:路径`。
`locate` 只认有角色的控件，经文里的单词要用 `click:x,y`。`type` 往当前输入框打字（先用 `label:` 点它），
`size` 改窗口大小，方便看 980×640 下的布局。

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
- **翻页**：书页两侧边缘各一个箭头，不占正文宽度；翻页围绕中央书脊（单栏绕左侧书脊）进行两阶段纸页掀起、背页卷弧高光与落页阴影仿真（`680ms` S 型加减速，支持章内与跨章/跨卷翻页）。翻到章末继续往前就进下一章的第一对开，
  在章首往回翻落到上一章的最后一对开，同理跨卷；到正典首尾才停。箭头的标签随位置变化
  （下一页 / 下一章 / 下一卷）。
- **键盘**：← → 翻页，空格同 →，Home / End 到头尾，Esc 关词条卡。
- **书签**：标记当前对开，页上缘挂书签带，左栏列出「卷 章:节」，点击跳回（按节重新定位）。
- **记住位置**：卷、章、页、模式、左右译本、语言、书签写在
  `~/Library/Application Support/bible-en-study/state.json`，下次打开回到原处；生词本单独写在同目录的
  `vocab.json`，所以翻页的频繁写入不会碰词库文件。
- **查词**：点正文里的词弹卡片（词性、音标、中文义、英文释义、本书搭配、词形变化、出处与原句——
  对照模式下中英两句都给）；词典 41,873 条，没收录的词会明说「还没有收录」。若点击的是屈折变形词，
  可直接点按「原形」跳转查看原形词卡；卡片底部可以收藏，已收藏的词还能一键「在词库查看」。
- **页上标记**：收藏过的词在正文里有淡底纹——生词最浅，学习中最深，已掌握不画，读起来和正文一样。
- **词库**：左侧模块轨切进来。一行身份栏（`词库` + 收录总数、搜索框、`我的词 / 全部词典`、`手动添加`），
  一行阶段筛选（`全部 / 生词 / 学习中 / 已掌握 / 错词`，筛选与卡片用同一个圆点说同一种阶段），
  主区是分页词卡墙（每页 12 张，词头、词性、音标、中文义与首条经文出处），底部页码条与 `PgUp` / `PgDn`；
  摘要行写「N 张词卡 · 按加入时间 · 已掌握 x / y」。点按词卡在右侧展开抽屉：词头、音标、阶段（圆点 + 文字两段式）、
  中英释义、本书搭配、收藏时那几节的中英原句与「回到原文」、底部「移出词库」；「收起」或 `Esc` 关抽屉，
  `↑↓` / `←→` 跨页切词卡。「回到原文」跳回那一节并把卡片开在原处。「手动添加」在同一个抽屉里填单词与可选中文释义，
  词典收录的词自动补齐音标与英译，没收录的在详情里写明。
- **空词库**：不摆一张空墙。居中的引导卡说明词怎么进来，给「去阅读」「手动添加一个词」，下面一行「浏览全部词典 ›」；
  此时筛选条收起。搜索或筛选没有结果时，卡片给出「查看全部词 / 清空搜索」。
- **全部词典**：词库顶栏切「我的词 / 全部词典」。顶部是 `全部` + `A–Z` 字母横栏，
  搜索框按词头前缀找（找不到前缀时退回子串，最多 200 条）；主区每页 24 张词卡展示内嵌的 41,873 条词条，
  点按词卡在右侧抽屉查看详情并直接收进我的词库。

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

中文译本（数据里没有，要用和合本得另进一批数据）、练习、英语基础、全文搜索、
脚注与串珠（库里没有注释表）。

## 七个动手前必须知道的约束

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
6. **窗口顶条由应用自己画**。`Config` 用 `app.WindowCustomTitlebar`，`TitlebarHeight: titlebarH`（32px），
   `paint` 的第一行是 `m.chrome(f)`：整条是拖动区（`MovesWindow`），中线是 `mark.png` + `Agnel` 字标。
   macOS 的红绿灯占住左侧 `app.TitlebarInsets()`（70px），顶上这一条的内容都要让开它；
   `ClearColor` 取自 `theme.Bg`，否则窗口边缘、标题条与页头会是三种颜色。`dev` 宿主窗口也从
   `reloadwire` 里拿到 `WindowMode` 与 `TitlebarHeight`，所以 `viva dev` 与打包后的窗口一致。
7. **形变要用本地 viva**。`app/go.mod` 里有 `replace github.com/vivaui/viva => ../../viva`：翻页的透视纸面
   靠 `ui.Canvas.FillQuad` 与 `ui.Canvas.LabelQuad`（任意四边形 + 每角颜色），这两个入口只在本地 viva
   检出里有，发布版 `v0.3.0` 上没有；`../../viva` 不在时 `app` 编不过。改动与根因见
   `docs/viva-framework-issues.md` §9。
