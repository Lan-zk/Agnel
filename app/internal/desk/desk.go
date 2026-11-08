package desk

import (
	"fmt"
	"log"
	"math"
	"runtime"
	"sort"
	"strings"
	"time"

	"app/internal/bible"
	"app/internal/dict"
	"app/internal/font"
	"app/internal/logo"
	"app/internal/render"
	"app/internal/store"
	"app/internal/theme"

	"github.com/vivaui/viva/app"
	"github.com/vivaui/viva/flex"
	"github.com/vivaui/viva/ui"
)

// Mode is how many translations the reader opens at once.
type Mode int

const (
	Spread Mode = iota // two translations, verse against verse
	Single             // one translation on a wider page
)

// The desk: how wide a spread may be, the room the desk keeps around the book,
// the margin inside a page, the widest a line of text may run, the strip below
// the book that carries the position, and the margin the page-turn controls
// keep for themselves.
const (
	spreadMax  = 1240
	singleMax  = 700
	deskPad    = 44
	pageMargin = 54
	measureMax = 472
	deskFoot   = 38

	// turnDuration is how long a page takes to settle into place.
	turnDuration = 320 * time.Millisecond
)

// Identities are declared once: a control's identity must be stable from frame
// to frame, so nothing below builds one inside the painter.
var (
	tree     ui.Tree
	read     = tree.Feature("read")
	words    = tree.Sub(read, tree.Name("word"))
	marks    = tree.Sub(read, tree.Name("bookmark"))
	prevPage = tree.ID(read, "prev")
	nextPage = tree.ID(read, "next")
	offWord  = tree.ID(read, "close-card")
	pageArea = tree.ID(read, "page")

	pickerID  = tree.Feature("picker")
	doClose   = tree.ID(pickerID, "close")
	lang      = tree.Feature("lang")
	recentRow = tree.Feature("recent")
	books     = tree.Feature("books")
	chaps     = tree.Feature("chaps")

	bookGrid = tree.ID(pickerID, "books")
)

// Font sizes and leading for the bilingual spread. English and Chinese share
// the same baseline grid (Lead: 26.5) so verse openings and first lines align
// across facing pages.
var (
	enStyle    = render.LineStyle{Size: 15, Weight: ui.WeightRegular, Lead: 26.5}
	zhStyle    = render.LineStyle{Size: 15, Weight: ui.WeightRegular, Lead: 26.5}
	hdStyle    = render.LineStyle{Size: 13.5, Weight: ui.WeightSemibold, Lead: 20}
	hdSub      = render.LineStyle{Size: 12.5, Weight: ui.WeightSemibold, Lead: 18}
	hdSubEn    = render.LineStyle{Size: 12.5, Weight: ui.WeightRegular, Italic: true, Lead: 18}
	refStyle   = render.LineStyle{Size: 11, Weight: ui.WeightRegular, Lead: 16}
	refStyleEn = render.LineStyle{Size: 11, Weight: ui.WeightRegular, Italic: true, Lead: 16}
)

type lookup struct {
	word string
	name string
	osis string
	ch   int
	v    int
	en   string
	zh   string
	rect ui.Rect
}

type heading struct {
	text  string
	style render.LineStyle
}

type block struct {
	v              int
	para           bool
	left, right    render.Paragraph
	headsL, headsR []render.Paragraph
	styleL, styleR []render.LineStyle
	headH          float32
	height         float32
}

type page struct {
	blocks []block
}

// State is everything the window needs to paint itself.
type State struct {
	view   *flex.View
	engine ui.TextEngine
	data   *bible.Store

	osis    string
	chapter int
	spread  int
	// landLast sends the next layout to the last spread of the chapter, which
	// is where a reader turning backwards into a new chapter wants to arrive.
	landLast bool
	mode     Mode
	lang     int // 0 中文 book names, 1 English
	left     string
	right    string

	picker     bool
	pickerBook string

	bookmarks []store.Bookmark
	recent    []string
	saved     []string

	leftIdx, rightIdx int
	leftSel, rightSel flex.SelectState
	leftScope         ui.Scope
	rightScope        ui.Scope

	lookup   *lookup
	cardBox  flex.Box
	cardRect ui.Rect

	layoutKey string
	pages     []page
	lastArea  ui.Rect
	win       ui.Size
	colL      float32
	colR      float32
	dirty     bool

	// The turn: the place has already moved, the picture catches up.
	now       time.Duration
	turning   bool
	turnDir   int
	turnBegan time.Duration

	// frameMode is the mode this frame is drawn in: a control that changes the
	// mode takes effect on the next frame, never halfway through one.
	frameMode Mode
}

func New() *State {
	return &State{view: new(flex.View), engine: app.PlatformTextEngine(), mode: Spread}
}

// Load reads the text, the dictionary and the saved place.
func Load() error {
	s, err := bible.Open()
	if err != nil {
		return err
	}
	if err := dict.Load(); err != nil {
		return err
	}
	loaded = s
	return nil
}

var loaded *bible.Store

// Run opens the window.
func Run(icon app.WindowIcon) {
	// Viva reads the reduce-motion setting with a dispatch_sync onto the main
	// queue before the app loop starts. If the goroutine has drifted off the
	// main thread by then, that wait never resolves and no window appears.
	runtime.LockOSThread()
	if err := Load(); err != nil {
		log.Fatal(err)
	}
	font.Register()
	m := New()
	m.data = loaded
	m.restore()
	var canvas ui.Canvas
	cfg := Config(m)
	// A packaged .app keeps its own icon; this one is for a program run without
	// a bundle, such as viva dev, whose Dock tile would otherwise be generic.
	cfg.Icon = icon
	if err := app.RunCanvas(cfg, &canvas); err != nil {
		log.Fatal(err)
	}
}

func Config(m *State) app.Config {
	m.data = loaded
	if m.osis == "" {
		m.osis = "Gen"
	}
	if m.chapter == 0 {
		m.chapter = 1
	}
	m.pickDefaults()
	return app.Config{
		Title:      "Agnel",
		Width:      1440,
		Height:     900,
		MinWidth:   980,
		MinHeight:  640,
		ClearColor: app.Color{Red: 0.87, Green: 0.89, Blue: 0.90, Alpha: 1},
		Paint:      m.paint,
		OnKey:      m.onKey,
	}
}

func (m *State) restore() {
	if m.data == nil {
		m.data = loaded
	}
	if m.data == nil {
		return
	}
	st, err := store.Load()
	if err != nil {
		return
	}
	if b, ok := m.data.Book(st.Book); ok {
		m.osis = b.OSIS
	}
	if m.osis == "" {
		m.osis = "Gen"
	}
	if st.Chapter > 0 {
		m.chapter = st.Chapter
	}
	if m.chapter < 1 {
		m.chapter = 1
	}
	if st.Mode >= 0 && st.Mode <= int(Single) {
		m.mode = Mode(st.Mode)
	}
	m.spread = max(0, st.Spread)
	m.lang = st.Lang
	if _, ok := m.data.Translation(st.Left); ok {
		m.left = st.Left
	}
	if _, ok := m.data.Translation(st.Right); ok {
		m.right = st.Right
	}
	m.bookmarks = st.Bookmarks
	m.recent = st.Recent
	m.saved = st.Saved
	m.pickDefaults()
}

// pickDefaults fixes the pair the page reads: an English text on the left, the
// Chinese one on the right. The left is the reader's choice, the right is not.
func (m *State) pickDefaults() {
	english := m.data.InLanguage("en")
	chinese := m.data.InLanguage("cmn")
	if len(english) > 0 {
		if !m.isEnglish(m.left) {
			m.left = english[0].ID
		}
	} else if m.left == "" && len(m.data.Translations()) > 0 {
		m.left = m.data.Translations()[0].ID
	}
	if len(chinese) > 0 {
		m.right = chinese[0].ID
	} else if m.right == "" {
		for _, t := range m.data.Translations() {
			if t.ID != m.left {
				m.right = t.ID
				break
			}
		}
	}
	m.syncSelects()
}

func (m *State) isEnglish(id string) bool {
	t, ok := m.data.Translation(id)
	return ok && !t.IsChinese()
}

// rightName is what the fixed right page is called in the bar.
func (m *State) rightName() string {
	if t, ok := m.data.Translation(m.right); ok {
		if t.IsChinese() {
			return "和合本"
		}
		return t.Abbr
	}
	return ""
}

func (m *State) leftChoices() []bible.Translation {
	if list := m.data.InLanguage("en"); len(list) > 0 {
		return list
	}
	return m.data.Translations()
}

func (m *State) syncSelects() {
	m.leftIdx = 0
	list := m.leftChoices()
	if m.mode != Spread {
		// one page can also be read in Chinese
		list = m.data.Translations()
	}
	for i, t := range list {
		if t.ID == m.left {
			m.leftIdx = i
		}
	}
	m.rightIdx = m.indexOf(m.right)
	if m.leftScope.IsZero() {
		tree2 := new(ui.Tree)
		m.leftScope = tree2.Feature("select-left")
		m.rightScope = tree2.Feature("select-right")
	}
}

func (m *State) translationLabels() []string {
	return labelsOf(m.data.Translations())
}

func (m *State) leftLabels() []string {
	return labelsOf(m.leftChoices())
}

func labelsOf(list []bible.Translation) []string {
	out := make([]string, 0, len(list))
	for _, t := range list {
		out = append(out, t.Abbr)
	}
	return out
}

// colWidthFor is the text column of one page.
func (m *State) colWidthFor(tr string) float32 {
	return m.colL
}

// styleFor picks the line style of a translation's text.
func (m *State) styleFor(id string) render.LineStyle {
	if t, ok := m.data.Translation(id); ok && t.IsChinese() {
		return zhStyle
	}
	return enStyle
}

func (m *State) indexOf(id string) int {
	for i, t := range m.data.Translations() {
		if t.ID == id {
			return i
		}
	}
	return 0
}

func (m *State) trans(id string) bible.Translation {
	t, _ := m.data.Translation(id)
	return t
}

// rowKey identifies everything a page layout depends on.
func (m *State) rowKey(page ui.Rect) string {
	return fmt.Sprintf("%s|%d|%d|%s|%s|%d|%d", m.osis, m.chapter, int(m.frameMode),
		m.left, m.right, int(page.Width), int(page.Height))
}

// deskWidth is how wide the book may be: two pages side by side, or one.
func (m *State) deskWidth() float32 {
	if m.frameMode == Spread {
		return spreadMax
	}
	return singleMax
}

// pageRects are the pages of the book on the desk: one page, or two facing
// ones, each as tall as a page of a book and never taller than the desk.
func (m *State) pageRects(desk ui.Rect) (ui.Rect, ui.Rect) {
	pageW := float32(singleMax)
	if m.frameMode == Spread {
		pageW = spreadMax / 2
	}
	if m.frameMode == Spread {
		pageW = min(pageW, desk.Width/2)
	} else {
		pageW = min(pageW, desk.Width)
	}
	height := min(desk.Height, pageW*1.44)
	y := desk.Y + (desk.Height-height)/2
	if m.frameMode == Spread {
		x := desk.X + (desk.Width-pageW*2)/2
		return ui.Rect{X: x, Y: y, Width: pageW, Height: height},
			ui.Rect{X: x + pageW, Y: y, Width: pageW, Height: height}
	}
	x := desk.X + (desk.Width-pageW)/2
	return ui.Rect{X: x, Y: y, Width: pageW, Height: height}, ui.Rect{}
}

func (m *State) paint(canvas *ui.Canvas, size ui.Size) {
	m.data = loaded
	m.now = canvas.Now()
	m.frameMode = m.mode
	m.win = size
	m.setFamily()
	f := m.view.Begin(canvas, m.theme(), ui.Rect{Width: size.Width, Height: size.Height})
	var deskSlot flex.Box
	pickerAtStart := m.picker

	if m.picker {
		m.pickerPage(f)
	} else {
		f.Column().Fill().Children(func() {
			m.bar(f)
			f.Divider()
			f.Row().Grow(1).PadSides(10, deskPad, deskFoot, deskPad).Children(func() {
				deskSlot = f.Slot().Grow(1)
			})
		})
	}

	if m.picker {
		canvas.SetEscapeTarget(doClose)
	}
	if m.lookup != nil {
		// A scrim declared before the card: any click outside it closes the card,
		// while the card's own controls (declared after) still win.
		canvas.Hit(offWord, ui.Rect{Width: size.Width, Height: size.Height})
		if canvas.Clicked(offWord) {
			m.lookup = nil
			m.dirty = true
		}
		if m.lookup != nil {
			canvas.SetEscapeTarget(offWord)
			f.PopupLayer(1)
			m.lookupCard(f)
		}
	}
	f.End()

	if pickerAtStart || m.picker {
		if m.dirty {
			m.dirty = false
			canvas.RequestFrameIn(time.Millisecond)
		}
		return
	}
	if m.lookup != nil {
		m.cardRect = m.cardBox.Rect()
	}
	_ = m.recent

	book, _ := m.data.Book(m.osis)
	chL, okL := m.data.Chapter(m.left, m.osis, m.chapter)
	chR := chL
	if m.frameMode == Spread {
		chR, _ = m.data.Chapter(m.right, m.osis, m.chapter)
	}
	desk := deskSlot.Rect()
	pageL, pageR := m.pageRects(desk)
	m.layoutFor(canvas, pageL, book, chL, chR)
	if len(m.pages) == 0 {
		return
	}
	if m.landLast {
		m.landLast = false
		m.spread = len(m.pages) - 1
		m.remember()
	}
	m.spread = min(m.spread, len(m.pages)-1)

	area := pageL
	if m.frameMode == Spread {
		area = union(pageL, pageR)
	}
	// The outer edge of the book. A single page has no second half, so its own
	// edge stands in for it.
	edgeR := pageR
	if m.frameMode == Single {
		edgeR = pageL
	}

	// The page area takes the keyboard and the empty-space clicks; the words are
	// declared after it so they sit in front of it.
	canvas.HitWithOptions(pageArea, area, "正文", ui.HitOptions{Role: ui.RoleGroup, Autofocus: true})

	// The turn: the place has already moved, the page slides the rest of the way.
	turn := float32(1)
	if m.turning {
		p := float32(m.now-m.turnBegan) / float32(turnDuration)
		if p >= 1 {
			m.turning = false
			p = 1
		} else {
			canvas.RequestFrame()
		}
		turn = easeOut(p)
	}

	pg := m.pages[m.spread]
	from := m.spread - m.turnDir
	if m.turning && m.turnDir != 0 && from >= 0 && from < len(m.pages) {
		// The page underneath stays where it is. The arriving spread is laid
		// over it from the outer edge, on a layer of its own: the paper of the
		// arriving page then covers the words of the page below, which the
		// text layer alone could never do.
		canvas.Clip(desk)
		m.paintSpread(canvas, pageL, pageR, book, chL, chR, m.pages[from], 0, false)
		dx := turnSlide(desk, area) * (1 - turn)
		if m.turnDir < 0 {
			dx = -dx
		}
		m.paintTurnShadow(canvas, pageL, edgeR, dx, m.turnDir, turn)
		canvas.PushLayer(1)
		m.paintSpread(canvas, pageL, pageR, book, chL, chR, pg, dx, true)
		canvas.PopLayer()
		canvas.Unclip()
	} else {
		m.paintSpread(canvas, pageL, pageR, book, chL, chR, pg, 0, true)
	}

	m.paintTurn(canvas, desk, pageL, edgeR)
	m.paintPosition(canvas, desk, pageL, edgeR)

	if m.dirty {
		m.dirty = false
		canvas.RequestFrameIn(time.Millisecond)
	}
	if !okL {
		_ = okL
	}
}

// easeOut is the curve a page follows as it settles into place.
func easeOut(p float32) float32 {
	q := 1 - p
	return 1 - q*q*q
}

// turnSlide is how far the arriving spread travels: it comes in from the far
// edge of the desk, so the reader never sees it appear out of nowhere.
func turnSlide(desk, area ui.Rect) float32 {
	return max(desk.X+desk.Width-area.X, area.X+area.Width-desk.X)
}

func (m *State) setFamily() {
	family := font.Register()
	if f, ok := m.engine.(ui.FontFamilyAware); ok {
		f.SetFontFamily(family)
	}
}

func union(a, b ui.Rect) ui.Rect {
	x0, y0 := math.Min(float64(a.X), float64(b.X)), math.Min(float64(a.Y), float64(b.Y))
	x1 := math.Max(float64(a.X+a.Width), float64(b.X+b.Width))
	y1 := math.Max(float64(a.Y+a.Height), float64(b.Y+b.Height))
	return ui.Rect{X: float32(x0), Y: float32(y0), Width: float32(x1 - x0), Height: float32(y1 - y0)}
}

func (m *State) theme() ui.Theme {
	t := ui.LightTheme()
	t.Background = theme.Bg
	t.Foreground = theme.Ink
	t.Card = theme.Surface
	t.CardElevated = theme.Surface
	t.CardForeground = theme.Ink
	t.Popover = theme.Surface
	t.PopoverForeground = theme.Ink
	t.Muted = theme.Surface2
	t.MutedForeground = theme.InkSoft
	t.Border = theme.Rule
	t.Input = theme.Rule
	t.Hover = theme.Surface2
	t.Primary = theme.Accent
	t.PrimaryForeground = theme.OnAccent
	t.Secondary = theme.Surface2
	t.SecondaryForeground = theme.Ink
	t.Accent = theme.AccentBg
	t.AccentForeground = theme.Accent
	t.Selection = theme.AccentBg
	t.SelectionForeground = theme.Ink
	t.Ring = theme.Accent
	t.Danger = theme.Danger
	t.DangerForeground = theme.OnAccent
	return t
}

// bar is the masthead: what is being read, what it is being read in, and the
// few controls that change the page.
func (m *State) bar(f *flex.Frame) {
	book, _ := m.data.Book(m.osis)
	f.Row().Height(54).PadX(24).Items(flex.Center).Gap(10).Children(func() {
		f.Image(logo.Bitmap()).Size(26, 26)
		title := m.bookName(book) + " " + fmt.Sprint(m.chapter)
		head := f.Row().Items(flex.Center).Gap(6).PadSides(5, 10, 5, 8).Radius(8).
			HoverBackground(theme.Surface)
		head.Children(func() {
			f.Text(title).FontSize(18.5).Weight(ui.WeightSemibold).Color(theme.Ink)
			f.Text("▾").FontSize(10).Color(theme.InkFaint)
		})
		if head.Hit(tree.ID(read, "open-picker"), title).Clicked() {
			m.openPicker(m.osis)
		}
		if capText := m.caption(book); capText != "" {
			f.Text(capText).FontSize(theme.SizeMicro).Color(theme.InkFaint).
				PadSides(3, 8, 3, 8).Radius(5).Background(theme.Surface2)
		}
		f.Spacer()
		f.Row().Items(flex.Center).Gap(10).Children(func() {
			if m.frameMode == Spread {
				f.Row().Items(flex.Center).Gap(6).Children(func() {
					f.Text("左页").FontSize(theme.SizeMicro).Color(theme.InkFaint)
					if i, ok := f.Select(m.leftScope, &m.leftSel, m.leftLabels(), m.leftIdx).Small().Picked(); ok {
						if choices := m.leftChoices(); i < len(choices) {
							m.leftIdx = i
							m.left = choices[i].ID
							m.layoutKey = ""
							m.remember()
						}
					}
					f.Column().Width(3).Height(3).Radius(1.5).Background(theme.InkFaint)
					f.Text("右页").FontSize(theme.SizeMicro).Color(theme.InkFaint)
					f.Row().Height(28).Items(flex.Center).PadSides(0, 10, 0, 10).Radius(6).
						Background(theme.Surface).Border(1, theme.Rule).Children(func() {
						f.Text(m.rightName()).FontSize(theme.SizeSmall).Color(theme.InkSoft)
					})
				})
			} else {
				if i, ok := f.Select(m.leftScope, &m.leftSel, m.translationLabels(), m.leftIdx).Small().Picked(); ok {
					if list := m.data.Translations(); i < len(list) {
						m.leftIdx = i
						m.left = list[i].ID
						m.layoutKey = ""
						m.remember()
					}
				}
			}
			m.barRule(f)
			f.Row().Items(flex.Center).Gap(2).Pad(3).Radius(8).Background(theme.DeskTrack).Children(func() {
				for i, label := range []string{"左右对照", "单栏"} {
					mode := Mode(i)
					chip := f.Text(label).FontSize(theme.SizeSmall).PadSides(5, 11, 5, 11).Radius(6)
					if m.mode == mode {
						chip.Background(theme.Card).Color(theme.Accent).Weight(ui.WeightSemibold)
					} else {
						chip.Color(theme.InkSoft).HoverBackground(theme.Surface2)
					}
					if chip.Hit(tree.ID(read, "mode-"+label), label).Clicked() && m.mode != mode {
						m.mode = mode
						if mode == Spread {
							// the spread is English on the left, Chinese on the right
							m.pickDefaults()
						}
						m.syncSelects()
						m.layoutKey = ""
						m.spread = 0
						m.remember()
					}
				}
			})
			m.barRule(f)
			bookmarked := m.bookmarked()
			label := "加书签"
			if bookmarked {
				label = "已加书签"
			}
			markBox := f.Row().Items(flex.Center).Gap(5).PadSides(11, 11, 6, 9).Radius(7)
			if bookmarked {
				markBox.Background(theme.AccentBg).HoverBackground(theme.Surface)
			} else {
				markBox.HoverBackground(theme.Surface2)
			}
			markBox.Children(func() {
				textColor := theme.InkSoft
				weight := ui.WeightRegular
				if bookmarked {
					textColor = theme.Accent
					weight = ui.WeightSemibold
				}
				f.Text(label).FontSize(theme.SizeSmall).Weight(weight).Color(textColor)
			})
			if markBox.Hit(tree.ID(read, "mark"), label).Clicked() {
				m.toggleBookmark()
			}
		})
	})
}

func (m *State) barRule(f *flex.Frame) {
	f.Column().Width(1).Height(16).Background(theme.DeskEdge)
}

// caption is the quiet line beside the title: how much of the chapter is here.
func (m *State) caption(book bible.Book) string {
	n := m.verseCount()
	if n == 0 {
		return ""
	}
	return fmt.Sprintf("%d 节", n)
}

func (m *State) verseCount() int {
	ch, ok := m.data.Chapter(m.left, m.osis, m.chapter)
	if !ok {
		return 0
	}
	return len(ch.Verses)
}

func (m *State) bookName(b bible.Book) string {
	if m.lang == 0 {
		return bible.ZH(b.OSIS)
	}
	return b.Name
}

func (m *State) mark(b bible.Book) string {
	if m.lang == 0 {
		return bible.Mark(b.OSIS)
	}
	return m.data.Title(m.left, b.OSIS)[:min(3, len(m.data.Title(m.left, b.OSIS)))]
}

func (m *State) markLabel(b store.Bookmark) string {
	book, _ := m.data.Book(b.Book)
	return fmt.Sprintf("%s %d", m.bookName(book), b.Chapter)
}

func (m *State) bookmarked() bool {
	for _, b := range m.bookmarks {
		if b.Book == m.osis && b.Chapter == m.chapter && b.Spread == m.spread {
			return true
		}
	}
	return false
}

func (m *State) toggleBookmark() {
	if m.bookmarked() {
		kept := m.bookmarks[:0]
		for _, b := range m.bookmarks {
			if !(b.Book == m.osis && b.Chapter == m.chapter && b.Spread == m.spread) {
				kept = append(kept, b)
			}
		}
		m.bookmarks = kept
	} else {
		m.bookmarks = append(m.bookmarks, store.Bookmark{Book: m.osis, Chapter: m.chapter, Spread: m.spread})
	}
	m.remember()
	m.dirty = true
}

func (m *State) jumpTo(b store.Bookmark) {
	m.osis = b.Book
	m.chapter = b.Chapter
	m.spread = b.Spread
	m.layoutKey = ""
	m.picker = false
	m.dirty = true
	m.remember()
}

// goSpread moves the place in the chapter and lets the page slide into place.
// Turning past either end of the chapter carries on into the chapter beside it,
// so reading never stops at a chapter or a book boundary.
func (m *State) goSpread(delta int) {
	next := m.spread + delta
	if next < 0 || (len(m.pages) > 0 && next > len(m.pages)-1) {
		m.goChapter(delta)
		return
	}
	if next == m.spread {
		return
	}
	m.spread = next
	m.turning = true
	m.turnDir = delta
	m.turnBegan = m.now
	// The card belongs to the page being left behind.
	m.lookup = nil
	m.dirty = true
	m.remember()
}

// canTurn says whether the turn goes anywhere: another spread here, or the
// chapter beside this one when this chapter has no more pages that way.
func (m *State) canTurn(delta int) bool {
	if delta < 0 {
		return m.spread > 0 || m.hasChapter(-1)
	}
	return m.spread < len(m.pages)-1 || m.hasChapter(1)
}

// turnLabel names what the turn does, so the control says 下一章 or 下一卷 where
// the chapter ends instead of promising another page of this one.
func (m *State) turnLabel(delta int) string {
	if delta < 0 {
		switch {
		case m.spread > 0:
			return "上一页"
		case m.sameBook(-1):
			return "上一章"
		}
		return "上一卷"
	}
	switch {
	case m.spread < len(m.pages)-1:
		return "下一页"
	case m.sameBook(1):
		return "下一章"
	}
	return "下一卷"
}

func (m *State) sameBook(delta int) bool {
	book, ok := m.data.Book(m.osis)
	if !ok {
		return false
	}
	c := m.chapter + delta
	return c >= 1 && c <= book.Chapters
}

func (m *State) hasChapter(delta int) bool {
	_, _, ok := m.chapterStep(delta)
	return ok
}

// chapterStep is the chapter one step away in canon order: the next chapter of
// this book, else the nearest chapter of the neighbouring book that has any.
func (m *State) chapterStep(delta int) (string, int, bool) {
	books := m.data.Books()
	i := -1
	for k, b := range books {
		if b.OSIS == m.osis {
			i = k
			break
		}
	}
	if i < 0 {
		return "", 0, false
	}
	if c := m.chapter + delta; c >= 1 && c <= books[i].Chapters {
		return m.osis, c, true
	}
	for k := i + delta; k >= 0 && k < len(books); k += delta {
		if books[k].Chapters <= 0 {
			continue
		}
		if delta > 0 {
			return books[k].OSIS, 1, true
		}
		return books[k].OSIS, books[k].Chapters, true
	}
	return "", 0, false
}

// goChapter opens the neighbouring chapter at the edge the reader was turning
// towards: its first spread going forward, its last going back.
func (m *State) goChapter(delta int) {
	book, chapter, ok := m.chapterStep(delta)
	if !ok {
		return
	}
	m.osis = book
	m.openChapter(chapter)
	m.landLast = delta < 0
}

func (m *State) openChapter(c int) {
	m.chapter = c
	m.spread = 0
	m.lookup = nil
	m.layoutKey = ""
	m.picker = false
	m.noteVisit()
	m.dirty = true
	m.remember()
}

// openPicker shows the passage picker. An empty book opens it at the list of
// books; a book opens it at its chapters.
func (m *State) openPicker(book string) {
	m.picker = true
	m.pickerBook = book
	m.dirty = true
}

func (m *State) noteVisit() {
	key := fmt.Sprintf("%s:%d", m.osis, m.chapter)
	out := []string{key}
	for _, r := range m.recent {
		if r != key && len(out) < 6 {
			out = append(out, r)
		}
	}
	m.recent = out
}

func (m *State) onKey(e ui.KeyEvent) ui.EventResult {
	if e.Release {
		return ui.EventResult{}
	}
	if m.picker {
		switch e.Key {
		case ui.KeyEscape:
			m.picker = false
			m.dirty = true
			return ui.EventResult{Repaint: true}
		}
		return ui.EventResult{}
	}
	if m.lookup != nil && e.Key == ui.KeyEscape {
		m.lookup = nil
		return ui.EventResult{Repaint: true}
	}
	switch e.Key {
	case ui.KeyLeft, ui.KeyPageUp:
		m.goSpread(-1)
	case ui.KeyRight, ui.KeyDown, ui.KeyPageDown, ui.KeySpace:
		m.goSpread(1)
	case ui.KeyHome:
		m.goSpread(-m.spread)
	case ui.KeyEnd:
		if len(m.pages) > 0 {
			m.goSpread(len(m.pages) - 1 - m.spread)
		}
	default:
		return ui.EventResult{}
	}
	return ui.EventResult{Repaint: true}
}

func (m *State) onPointer(e ui.PointerEvent) ui.EventResult {
	if m.lookup == nil {
		return ui.EventResult{}
	}
	if e.Kind != ui.PointerDown {
		return ui.EventResult{}
	}
	if e.X < m.cardRect.X || e.X > m.cardRect.X+m.cardRect.Width ||
		e.Y < m.cardRect.Y || e.Y > m.cardRect.Y+m.cardRect.Height {
		m.lookup = nil
		return ui.EventResult{Repaint: true}
	}
	return ui.EventResult{}
}

// layoutFor wraps the chapter once per column width and breaks it into pages.
func (m *State) layoutFor(c *ui.Canvas, sheet ui.Rect, book bible.Book, left, right bible.Chapter) {
	key := m.rowKey(sheet)
	if key == m.layoutKey {
		return
	}
	m.layoutKey = key
	m.lastArea = sheet

	// The measure: a line of text as wide as a book line, and no wider.
	measure := min(sheet.Width-2*pageMargin, measureMax)
	if measure < 200 {
		measure = max(sheet.Width-2*pageMargin, 120)
	}
	// Both pages are set to the same measure: the verse numbers line up, and
	// the Chinese gets the room it needs without breaking the alignment.
	m.colL, m.colR = measure, measure
	textWidth := measure - 28
	avail := sheet.Height - 62 - 46

	headsL := headingsByVerse(left.Headings, !m.english(m.left))
	headsR := headingsByVerse(right.Headings, !m.english(m.right))
	bodyL := m.styleFor(m.left)
	bodyR := m.styleFor(m.right)

	blocks := make([]block, 0, len(left.Verses))
	for i, v := range left.Verses {
		b := block{v: v.N, para: v.Para}
		b.left = render.Layout(c, v.Text, bodyL, textWidth)
		if m.frameMode == Spread && i < len(right.Verses) && right.Verses[i].N == v.N {
			b.right = render.Layout(c, right.Verses[i].Text, bodyR, textWidth)
		}
		var headL, headR float32
		for _, h := range headsL[v.N] {
			p := render.Layout(c, h.text, h.style, textWidth)
			b.headsL = append(b.headsL, p)
			b.styleL = append(b.styleL, h.style)
			headL += p.Height
		}
		if m.frameMode == Spread {
			for _, h := range headsR[v.N] {
				p := render.Layout(c, h.text, h.style, textWidth)
				b.headsR = append(b.headsR, p)
				b.styleR = append(b.styleR, h.style)
				headR += p.Height
			}
		}
		head := max(headL, headR)
		if head > 0 {
			head += 6
		}
		b.headH = head
		h := max(b.left.Height, b.right.Height) + 6
		if v.Para && i > 0 {
			h += 7
		}
		b.height = h + head
		blocks = append(blocks, b)
	}

	var pages []page
	cur := page{}
	used := float32(0)
	for _, b := range blocks {
		if used+b.height > avail && len(cur.blocks) > 0 {
			pages = append(pages, cur)
			cur = page{}
			used = 0
		}
		cur.blocks = append(cur.blocks, b)
		used += b.height
	}
	if len(cur.blocks) > 0 {
		pages = append(pages, cur)
	}
	m.pages = pages
}

func headingsByVerse(list []bible.Heading, chinese bool) map[int][]heading {
	out := make(map[int][]heading, len(list))
	sub := hdSubEn
	ref := refStyleEn
	if chinese {
		sub = hdSub
		ref = refStyle
	}
	for _, h := range list {
		switch h.Kind {
		case "section":
			if h.Level <= 1 {
				out[h.Verse] = append(out[h.Verse], heading{text: h.Text, style: hdStyle})
			} else {
				out[h.Verse] = append(out[h.Verse], heading{text: h.Text, style: sub})
			}
		case "psalm_title", "acrostic", "speaker", "major_section", "chapter_label":
			out[h.Verse] = append(out[h.Verse], heading{text: h.Text, style: hdStyle})
		case "reference":
			out[h.Verse] = append(out[h.Verse], heading{text: h.Text, style: ref})
		}
	}
	return out
}

// remember writes the place to disk; a failure is not worth interrupting reading.
func (m *State) remember() {
	_ = store.Save(store.State{
		Book: m.osis, Chapter: m.chapter, Spread: m.spread, Mode: int(m.mode),
		Lang: m.lang, Left: m.left, Right: m.right,
		Bookmarks: m.bookmarks, Recent: m.recent, Saved: sorted(m.saved),
	})
}

func sorted(in []string) []string {
	out := append([]string(nil), in...)
	sort.Strings(out)
	return out
}

func firstRune(s string) string {
	for _, r := range s {
		return string(r)
	}
	return s
}

var _ = strings.TrimSpace
