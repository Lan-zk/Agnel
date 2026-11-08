# `app/internal/font` — 内嵌双语阅读字体

把经过专门配比与字表裁剪的中英双语衬线阅读字体（**Bible Literata**）通过 `//go:embed` 打进二进制，并在启动或无头渲染首帧时通过 `app.RegisterFont` 注册到当前进程。

## 为什么需要合成内嵌字体

Viva v0.3.0 的比例字体引擎（`app.SetInterfaceFontFamily`）全局只接受单一字体家族名，且不带跨语种 `NSFontCascadeListAttribute` 回退栈：

- 若直接设成中文字体（如 `Songti SC`），字体内部自带的拉丁字形会直接接管英文排版，导致英文经文间距局促、缺乏西文书籍衬线气质，且缺失的若干国际音标字符（`ɹ`、`ɪ`、`ə` 等）会在单词中间回退到系统无衬线字体 `SF Pro`。
- 若直接设成西文衬线字体（如 `Literata` 或 `Iowan Old Style`），其不覆盖汉字，CoreText 会将中文经文回退到系统默认的无衬线黑体 `PingFang SC`，失去中文《和合本》应有的书卷宋体质感。

因此本包通过构建期脚本 `scripts/build_fonts.py` 将三套 **SIL Open Font License 1.1** 开源字体在同一度量空间（`unitsPerEm = 1000`）下合成为同一家族 `Bible Literata`：

1. **英文与西文标点（`U+0020..U+1FFF` 及弯引号、破折号）**：采用 **Literata**（TypeTogether 为 Google Play Books 设计的屏显长文阅读衬线体），完整保留原始字形索引 `0..N_base-1`，使原生的 `GPOS` 字距调整（Kerning）、`GSUB` 连字（`fi`、`fl`）与真斜体（Italic）原样生效。
2. **国际音标补齐（`ɔ, ɚ, ɛ, ɝ, ɪ, ɹ, ʃ, ʊ, ʌ, ʒ`）**：从 **STIX Two Text** 补齐 Literata 未覆盖的 10 个词典音标字符，保证查词卡音标在同一衬线风格内完整渲染。
3. **中文与中文标点（`>= U+2000`）**：采用 **Source Han Serif SC（思源宋体）**，按全库 4 个译本经文、41,873 条词典及界面源码实际用到的 4,409 个汉字与标点做子集化，并按 `0.92x` 字面缩放与 `-35` 单位基线微调，使汉字与 Literata 的 x-height 在双栏对照与混排时视觉重心对齐。

## 文件与入口

| 文件 | 作用 |
|---|---|
| `font.go` | `Register() string`：使用 `sync.Once` 向 Viva 注册 Regular / Bold / Italic 三个字重并返回家族名 `Bible Literata` |
| `bible-literata-regular.ttf` | 常规体（Regular 400，含 Literata + STIX IPA + 思源宋体子集，约 2.1 MB） |
| `bible-literata-bold.ttf` | 半粗体（Bold/SemiBold 600，含 Literata SemiBold + STIX IPA + 思源宋体 SemiBold 子集，约 2.1 MB） |
| `bible-literata-italic.ttf` | 真斜体（Italic 400，含 Literata Italic + STIX Italic IPA，约 287 KB） |

## 重新生成

当词典或经文新增了未覆盖的生僻汉字时，在仓库根目录运行（仅依赖 Python 标准库）：

```sh
python3 scripts/build_fonts.py
```
