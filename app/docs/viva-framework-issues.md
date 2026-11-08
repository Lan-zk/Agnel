# Viva 框架共性问题与可复用改进建议（基于 v0.3.0）

本文档记录在打磨 `bible-en-study`（Agnel）桌面阅读器排版、字体与微交互过程中，定位到的 **Viva `v0.3.0` 框架级共性限制与根因**，并给出向下兼容、可抽象复用的框架演进方案。

---

## 1. 单帧全局唯一比例字体与延迟光栅化状态耦合

### 现象与限制
- 应用无法在同一帧中为不同区域指定不同的比例字体族（例如：UI 顶栏使用无衬线字体 `Inter` / `PingFang SC`，正文阅读区使用衬线字体 `Literata` / `思源宋体`）。
- 若在 `Paint()` 过程中多次调用 `engine.SetFontFamily(family)` 试图切换字体：
  1. `SetFontFamily` 每次调用都会执行 `engine.cache.Clear()` 并递增 `engine.generation`（`internal/platform/darwin/text_engine.go:68-84`），导致整帧文本纹理缓存全部失效；
  2. `canvas.TextWidthStyle` 在 `Paint()` 期间同步测量，但 `canvas.Label` 的实际字形光栅化发生在 `Paint()` 结束后的 Metal 纹理上传阶段（`internal/platform/darwin/native.m:2747` `drawLineInContext` → `viva_shared_text_line_style`），此时读取的是帧末最后一次设置的全局静态变量 `viva_interface_family`，造成**测量用 A 字体、绘制用 B 字体**的错位。

### 框架根因（Viva v0.3.0）
- `ui/canvas.go:152-168` 中，`ui.TextStyle` 结构体为等宽字体提供了 `Mono bool` 与 `MonoFace MonoFont` 字段，但**比例字体没有对应的 `Font` / `Family` 字段**。
- `internal/platform/darwin/native.m:948` 使用单一全局 C 缓冲区 `static char viva_interface_family[256]`；`viva_text_line_key`（`native.m:2663`）的纹理缓存键仅包含 `(value, fontSize, fontWeight, italic, mono, monoFamily)`，未包含比例字体族名称。

### 可复用改进方案
1. **在 `ui.TextStyle` 中引入 `Font FontFace`（或复用统一的 `FontFace` 类型）**：
   - 支持 `ui.FontDefault`（走全局界面字体）、`ui.FontFamily("Literata")`、`ui.FontStack("Literata", "Source Han Serif SC")`。
2. **将比例字体标识透传至 `TextCommand`、`StyledTextEngine` 与 `viva_text_line_key`**：
   - 让比例字体像 `MonoFace` 一样按绘制指令携带字体描述符，彻底解除测量阶段与 Metal 延迟光栅化阶段对可变全局状态 `viva_interface_family` 的时序耦合。

---

## 2. 比例字体缺少 Fallback 级联栈（`NSFontCascadeListAttribute` 仅对等宽字体生效）

### 现象与限制
在中英双语或包含国际音标（IPA）、古希腊语/希伯来语、数学符号的桌面应用中：
- 若将全局字体设为中文字体（如 macOS 系统自带 `Songti SC`），由于中文字体内置了 ASCII 拉丁字形（`U+0020..U+007E`），CoreText 不会触发回退，英文正文会被迫使用中文字体附带的窄体 Times 风格西文字形；而遇到中文字体缺失的 IPA 音标字符（如 `ɹ, ˈ, ɪ, ə`）时，CoreText 会直接回退到系统无衬线字体 `SF Pro`，导致同一个音标词内出现衬线与无衬线混排。
- 若将全局字体设为西文字体（如 `Literata` 或 `Charter`），中文字符会跳过衬线宋体、直接回退到系统默认黑体 `PingFang SC`。

### 框架根因（Viva v0.3.0）
- 在 `internal/platform/darwin/native.m:1484-1500` 中，等宽字体函数 `viva_mono_font` 已经实现了完整的逗号分隔字体栈解析，并通过 `NSFontCascadeListAttribute` 构造级联回退链（配合 `ui.MonoStack(primary, fallbacks...)`）。
- 但比例字体函数 `viva_system_font`（`native.m:948-999`）只把 `viva_interface_family` 当作单一家族名传给 `NSFontFamilyAttribute`，未解析级联列表，也未设置 `NSFontCascadeListAttribute`。

### 可复用改进方案
- 将 `viva_mono_font` 中的级联列表构造逻辑提取为公共辅助函数 `viva_font_descriptor_with_cascade(const char *spec, CGFloat weight)`，在 `viva_system_font` 中同样支持逗号分隔的字体栈（例如 `"Literata, Source Han Serif SC, STIX Two Text"`），并提供 `ui.FontStack(primary string, fallbacks ...string)`。
- 在当前项目中的临时绕过方案：我们通过构建期脚本 `scripts/build_fonts.py` 将 `Literata`（Latin）、`STIX Two Text`（IPA 补齐）与 `Source Han Serif SC`（精简汉字子集）在 `unitsPerEm = 1000` 下合成为同一家族 `Bible Literata` 的 Regular / Bold / Italic 三个 `.ttf` 并通过 `app.RegisterFont` 注册。若框架原生支持比例字体 `FontStack` 与 `NSFontCascadeListAttribute`，应用层将不再需要离线合成字体表。

---

## 3. `app.RegisterFont` 进程级注册字体的字重与斜体特征匹配鲁棒性

### 现象与限制
通过 `app.RegisterFont(data)`（底层调用 `CTFontManagerRegisterFontsForURL(..., kCTFontManagerScopeProcess, ...)`）注册自定义多字重字体时：
- 当 UI 请求 `ui.WeightSemibold`（600）或 `ui.WeightMedium`（500），而注册的自定义字体家族只提供 Regular（400）与 Bold（700），或 `OS/2.usWeightClass` 与 CoreText 的 `NSFontWeightTrait` 浮点映射存在细微偏差时，`[NSFont fontWithDescriptor:descriptor size:size]` 在进程级注册字体上可能静默回退到 Regular（400），而不是选择最接近的粗体面。
- 当同时请求 `WeightSemibold` + `Italic: true` 时，`viva_system_font_style`（`native.m:1001-1014`）先按 `NSFontWeightTrait` 查找，再调用 `CTFontCreateCopyWithSymbolicTraits(..., kCTFontItalicTrait, kCTFontItalicTrait)`，若家族中没有独立的 BoldItalic 字体文件，符号特征拷贝可能丢失粗体或斜体。

### 可复用改进方案
- 在 `viva_system_font_style` 中：
  1. 当 `weight >= 600` 通过 `NSFontWeightTrait` 解析出的 `CTFont` 与 `weight = 400` 的 PostScript 名称相同时，尝试使用 `CTFontCreateCopyWithSymbolicTraits(font, 0.0, NULL, kCTFontBoldTrait, kCTFontBoldTrait)` 匹配同家族的 Bold 字面；
  2. 将 `kCTFontBoldTrait` 与 `kCTFontItalicTrait` 合并在一次 `CTFontCreateCopyWithSymbolicTraits` 调用中完成。

---

## 4. 缺少纵向字体度量（Ascent / Descent / Baseline）与基线对齐原语

### 现象与限制
- `ui.TextEngine` 接口（`ui/text_engine.go:19-31`）目前只暴露水平宽度测量 `Width(text string, size float32, weight Weight) float32`，未暴露纵向度量（`Ascent`、`Descent`、`CapHeight`、`XHeight`）。
- `canvas.Label(text, rect, style)` 始终将单行文本在 `rect.Height` 内垂直居中。当需要将不同字号的文本按**排版基线（Baseline）**对齐时（例如：10.5pt 的节号与 15pt 的正文首行对齐、词典卡片中 20pt 词头与 12pt 词性标签/音标对齐），调用方只能通过调整 `rect.Y` 与 `rect.Height` 手工补偿像素偏移。
- `flex.Row().Items(...)` 支持 `Start`、`Center`、`End`、`Stretch`，尚不支持 `flex.Baseline`。

### 可复用改进方案
1. 在 `ui.TextEngine`（或扩展接口 `MetricsTextEngine`）中增加：
   ```go
   type FontMetrics struct {
       Ascent    float32
       Descent   float32
       Leading   float32
       CapHeight float32
       XHeight   float32
   }
   ```
2. 在 `ui.Canvas` 上提供 `LabelBaseline(text string, x, baselineY float32, style TextStyle)`，并在 `flex.Align` 中支持 `flex.Baseline`（按子项首行 `Ascent` 对齐）。

---

## 5. `flex.Box.Hit(id, label)` 的显式无障碍标签被子文本节点覆盖

### 现象与限制
当使用 `flex` 构建组合可点击控件时：
```go
pick := f.Row().Items(flex.Center).Gap(6)
pick.Children(func() {
    f.Text("创世记 1")
    f.Icon(iconChevronDown)
})
pick.Hit(openPicker, "选择书卷与章")
```
调用方在 `pick.Hit(openPicker, "选择书卷与章")` 中显式传入了语义化标签 `"选择书卷与章"`，但子级 `f.Text("创世记 1")` 在布局收敛时会将该 `Hit` 条目的 `Label` 覆盖为 `"创世记 1"`。这会导致：
- 无头截图工具 `cmd/shot` 的 `"label:选择书卷与章"` 或自动化测试 `app.Driver` 无法按显式声明的 `Hit` 标签定位控件；
- 当动态文本变化（如从 `"创世记 1"` 变为 `"出埃及记 2"`）时，外层容器失去了稳定的语义名称。

### 可复用改进方案
- 当 `box.Hit(id, label)` 显式传入非空 `label` 时，优先保留显式 `label`（或同时在 `ui.Hit` 上保留 `Label` 与自动聚合的子文本 `Text`），避免子级 `f.Text` 静默覆盖外层容器显式声明的无障碍标签。

---

## 6. `ui.Icon` 仅支持描边路径（Stroke）与 `LineHeight` 单位语义易混淆

### 现象与限制
1. **`ui.Icon` 缺少填充模式（Fill）**：
   - `ui.Icon`（`ui/icon.go`）只包含 `Stroke float32`、`Cap`、`Join`，`DrawIcon` 仅对路径进行描边。在绘制实心小图标（如已激活的实心书签、实心徽章或品牌字标）时，无法直接使用 `ui.Icon` 填充闭合路径，只能用多个 `c.Fill(ui.Rect{...})` 拼接或退回位图 `f.Image`。
2. **`flex.Text.LineHeight(v)` 的像素单位语义**：
   - 在 CSS 与主流 UI 框架中，`LineHeight(1.5)` 通常表示 `1.5 × FontSize`；而在 Viva 中 `LineHeight(1.5)` 表示 `1.5px` 绝对像素行高，导致多行文本重叠在一起。

### 可复用改进方案
1. 为 `ui.Icon` 增加 `Fill bool`（默认 `false` 保持向后兼容；当 `Fill: true` 且 `Stroke == 0` 时执行路径填充，两者兼有时先填充再描边）。
2. 为 `flex.Text` 增加语义明确的 `LineHeightPx(px float32)` 与 `LineHeightScale(mult float32)`，或在 `LineHeight(v)` 中当 `v > 0 && v <= 4` 时按字号倍率解释。

---

## 7. 路径图元（`ui.Icon` / `canvas.DrawPath` 的三角形）在窗口里不落像素，离屏渲染却正常

### 现象与限制
- 用 `ui.Icon{ViewBox: 24, Stroke: 2, Cap: ui.CapRound, Join: ui.JoinRound, Commands: ...}` 画的翻页箭头、书签、下拉箭头，在 `go run -tags vivahead ./cmd/shot`（离屏软件渲染）里正常出现；在真实窗口（`viva dev`，Metal）里一个都不出现。
- 同一坐标、同一尺寸的圆角矩形 `c.Rounded(ui.Rect{X: cx-18, Y: cy-18, Width: 36, Height: 36}, 18, theme.Surface)` 在窗口里正常出现，而叠在它正中的图标不可见——排除了坐标、裁剪、图层与颜色 alpha 的可能，问题只在“三角形光栅化”这一段。
- 把同一个图标改成填充轮廓（`Stroke: 0`，闭合三角形）后窗口里仍然不出现，说明**整条三角形管线**在窗口里都不出像素，与描边/填充无关；矩形、圆角、文字、`flex.Image` 都正常。
- `ui.Path` 的字段未导出（`Path{width, height, commands}` 全小写），应用无法从外部构造一条路径去单独验证某一种画法；区分“描边还是填充失效”只能靠把 `ui.Icon` 的 `Stroke` 改成 0 重试。

### 框架根因（Viva v0.3.0）
- `ui/icon.go:502 paintIconShapes` 把图标编译成 `PathStyle{Stroke: tint, StrokeWidth: icon.Stroke * units.y}` 并调用 `canvas.DrawPath`（`ui/path.go:142`）；`DrawPath` 把轮廓展平成三角形追加进 `canvas.triangles`。
- `app/painter.go:1387 applyClipping` 用 `placeTrianglesBefore` 把 `canvas.Triangles()` 按绘制顺序插回图元序列，因此在 `Snapshot().Primitives` 里三角形是存在的（`Kind == TrianglePrimitive`），只是没有变成窗口像素。声明侧、排序侧都成立，且填充与描边一样不出现，因此差异在三角形这一批次的上屏路径（着色器绑定、`EdgeMask` 覆盖率通道，或窗口渲染器与离屏渲染器走了不同分支），不在某一种画法。

### 可复用改进方案
1. 提供公开的路径或图形入口（`ui.NewPath`、`canvas.DrawShape(shape, bounds, style)`），让应用和框架测试都能独立复现“一条描边路径”。
2. 增加“窗口像素 vs `Snapshot()` 图元”的一致性测试：现在快照里有点亮三角形的记录、窗口里却没有像素，正是这个 bug 一直藏着的原因。
3. 检查 Darwin 三角形管线的 `EdgeMask` 覆盖率分支，与 `scene.KindTriangle` 的离屏展开逻辑对齐。
4. 项目内的临时绕过：图标全部改用字形（`‹` `›` 作翻页箭头、`▾` 作下拉提示），界面里不再出现路径图元。

---

## 8. `Driver.Pump()` 在应用未请求帧时静默不重绘

### 现象与限制
- 离屏测试里直接改 `State`（例如 `m.openChapter(2)`）后调用 `d.Pump()`，`Snapshot()` 返回的仍是上一帧，测试读到的是过期画面；断言要么假通过，要么报“控件不存在”。
- `app/driver.go:185-191` 的 `Pump()` 只有在 `driver.wake` 为真时才 `redraw()`；而 `wake` 只由绘制期间的 `canvas.RequestFrame*` 或真实输入事件置位。测试无法“强制一帧”。

### 可复用改进方案
1. `Pump()` 无条件重绘，或另提供 `driver.Redraw()`。
2. `Snapshot()` 带上帧号，测试断言前可以确认拿到的是新帧，而不是靠“再 Pump 一次”。
