# app/docs

原生应用（`app/`）的视觉交付截图、后续实现计划与 Viva 框架级问题文档。

## 主要内容

| 文件 | 说明 |
|---|---|
| `roadmap.md` | 以「收藏 → 词库 → 练习」闭环为主线的后续实现计划：M1 生词本、M2 练习、M3 英语基础、M4 阅读器深度，含数据契约与验收标准 |
| `viva-framework-issues.md` | 基于 Viva `v0.3.0` 梳理的 8 项框架级共性问题、源码根因定位与可抽象复用的改进建议 |
| `spread.png` / `spread.jpg` | 左右对照（左页英文 BSB，右页中文和合本）对开页渲染截图 |
| `lookup.png` / `lookup.jpg` | 点击英文单词后弹出的原位词典查词卡与双语经文引文框 |
| `picker.png` / `picker.jpg` | 66 卷正典目录与原位章次展开（中文态） |
| `picker-en.png` / `picker-en.jpg` | 66 卷正典目录与原位章次展开（英文态） |
| `chapters.png` / `chapters.jpg` | 诗篇 150 章全宽章次网格展开态 |
| `logo.png` / `logo.jpg` | 顶栏标志与窗口局部截图 |

## 重新生成截图

在 `app/` 目录下使用无头 Metal/CoreText 渲染器 `cmd/shot` 生成：

```sh
cd app
go run -tags vivahead ./cmd/shot docs/spread.png \
  "label:created" "shot:docs/lookup.png" \
  "label:关闭" "label:创世记 1" "shot:docs/picker.png" \
  "label:English" "shot:docs/picker-en.png" \
  "label:中文" "label:诗篇" "shot:docs/chapters.png"
```
