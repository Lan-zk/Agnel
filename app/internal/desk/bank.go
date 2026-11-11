package desk

import (
	"fmt"
	"sort"
	"strings"

	"github.com/vivaui/viva/flex"
	"github.com/vivaui/viva/ui"

	"app/internal/dict"
	"app/internal/logo"
	"app/internal/store"
	"app/internal/theme"
	"app/internal/vocab"
)

// The word bank: a 68px module rail on the left of the window, a top filter &
// search bar, a paginated Bento card wall of saved words (or the 41,873-word
// lexicon), and a slide-over detail drawer on the right when a card is open.
const (
	bankMax      = 1360
	drawerWidth  = 448
	railWidth    = 68
	bankPageSize = 12
	dictPageSize = 24
)

var (
	bankSearch      = tree.ID(bankID, "search")
	bankClear       = tree.ID(bankID, "search-clear")
	bankQuote       = tree.ID(bankID, "add-word")
	bankZh          = tree.ID(bankID, "add-zh")
	bankSubmit      = tree.ID(bankID, "add-submit")
	bankCancel      = tree.ID(bankID, "add-cancel")
	bankViews       = tree.Sub(bankID, tree.Name("views"))
	bankLetters     = tree.Sub(bankID, tree.Name("letters"))
	bankDict        = tree.Sub(bankID, tree.Name("dict"))
	bankSave        = tree.ID(bankID, "save")
	bankLemmaJump   = tree.ID(bankID, "lemma-jump")
	bankEmptyDict   = tree.ID(bankID, "empty-dict")
	bankEmptyAll    = tree.ID(bankID, "empty-all")
	bankEmptyClear  = tree.ID(bankID, "empty-clear")
	bankEmptyAdd    = tree.ID(bankID, "empty-add")
	bankEmptyBrowse = tree.ID(bankID, "empty-browse")
	bankDrawerClose = tree.ID(bankID, "drawer-close")
	bankPagePrev    = tree.ID(bankID, "page-prev")
	bankPageNext    = tree.ID(bankID, "page-next")
	bankPageNums    = tree.Sub(bankID, tree.Name("page-nums"))
)

// modules is the rail: reading and the word bank are built; the other two are
// listed so the shape of the product is visible, and say so when clicked.
type module struct {
	label string
	pane  pane
	built bool
}

var modules = []module{
	{"阅读", paneReader, true},
	{"词库", paneBank, true},
	{"练习", paneReader, false},
	{"英语基础", paneReader, false},
}

// rail draws the module rail under the bar. The labels are two characters set one
// above the other: an icon font is not part of the binary, and a Chinese label the
// reader already knows beats a glyph that needs explaining.
func (m *State) rail(f *flex.Frame) {
	f.Column().Width(railWidth).PadY(12).Gap(4).Items(flex.Center).Children(func() {
		for i, mod := range modules {
			active := mod.built && mod.pane == m.pane
			b := f.Column().Width(railWidth - 14).Height(64).Justify(flex.Center).Items(flex.Center).Gap(1).Radius(10)
			switch {
			case active:
				b = b.Background(theme.AccentBg)
			case mod.built:
				b = b.HoverBackground(theme.Surface)
			}
			b.Children(func() {
				for _, r := range mod.label {
					t := f.Text(string(r)).FontSize(12).LineHeight(15)
					switch {
					case active:
						t.Color(theme.Accent).Weight(ui.WeightSemibold)
					case mod.built:
						t.Color(theme.InkSoft)
					default:
						t.Color(theme.InkFaint)
					}
				}
			})
			if b.Hit(railID.At(i), mod.label).Clicked() && mod.built {
				m.goPane(mod.pane)
			}
		}
		f.Spacer()
	})
}

// goPane switches module, keeping the reading place in memory.
func (m *State) goPane(p pane) {
	if m.pane == p {
		return
	}
	m.pane = p
	m.adding = false
	m.confirmDelete = ""
	// The card belongs to the reading page it was opened on.
	m.lookup = nil
	if p == paneBank && !m.dictMode {
		m.ensureBankSelection()
	}
	m.dirty = true
}

// ensureBankSelection makes sure the bank opens on a real word when words exist,
// and keeps bankPg aligned with the selected word.
func (m *State) ensureBankSelection() {
	words := m.bankWords()
	if len(words) == 0 {
		m.selected = ""
		m.bankPg = 0
		return
	}
	for i, w := range words {
		if strings.EqualFold(w.Word, m.selected) {
			m.bankPg = i / bankPageSize
			return
		}
	}
	m.selected = words[0].Word
	m.bankPg = 0
}

// syncPageToSelected moves bankPg or dictPg to the page containing m.selected.
func (m *State) syncPageToSelected() {
	if m.selected == "" {
		return
	}
	if m.dictMode {
		words := m.dictWords()
		for i, w := range words {
			if strings.EqualFold(w, m.selected) {
				m.dictPg = i / dictPageSize
				return
			}
		}
		return
	}
	words := m.bankWords()
	for i, w := range words {
		if strings.EqualFold(w.Word, m.selected) {
			m.bankPg = i / bankPageSize
			return
		}
	}
}

// stepBankPage moves the active pagination page by delta (-1 or +1).
func (m *State) stepBankPage(delta int) bool {
	if m.dictMode {
		total := len(m.dictWords())
		pages := pageCount(total, dictPageSize)
		next := clampPage(m.dictPg+delta, pages)
		if next == m.dictPg {
			return false
		}
		m.dictPg = next
		return true
	}
	total := len(m.bankWords())
	pages := pageCount(total, bankPageSize)
	next := clampPage(m.bankPg+delta, pages)
	if next == m.bankPg {
		return false
	}
	m.bankPg = next
	return true
}

func pageCount(total, size int) int {
	if total <= 0 {
		return 1
	}
	return (total + size - 1) / size
}

func clampPage(p, pages int) int {
	if p < 0 {
		return 0
	}
	if p >= pages {
		return pages - 1
	}
	return p
}

// stepBankSelection moves the selected word up or down by delta in the active list,
// automatically turning the page when crossing a page boundary.
func (m *State) stepBankSelection(delta int) bool {
	if m.dictMode {
		words := m.dictWords()
		if len(words) == 0 {
			return false
		}
		cur := -1
		for i, w := range words {
			if strings.EqualFold(w, m.selected) {
				cur = i
				break
			}
		}
		next := cur + delta
		if cur < 0 {
			next = m.dictPg * dictPageSize
		}
		if next < 0 {
			next = 0
		}
		if next >= len(words) {
			next = len(words) - 1
		}
		if strings.EqualFold(words[next], m.selected) {
			return false
		}
		m.selected = words[next]
		m.dictPg = next / dictPageSize
		return true
	}
	words := m.bankWords()
	if len(words) == 0 {
		return false
	}
	cur := -1
	for i, w := range words {
		if strings.EqualFold(w.Word, m.selected) {
			cur = i
			break
		}
	}
	next := cur + delta
	if cur < 0 {
		next = m.bankPg * bankPageSize
	}
	if next < 0 {
		next = 0
	}
	if next >= len(words) {
		next = len(words) - 1
	}
	if strings.EqualFold(words[next].Word, m.selected) {
		return false
	}
	m.selected = words[next].Word
	m.bankPg = next / bankPageSize
	return true
}

// barBank is the top bar over the word bank.
func (m *State) barBank(f *flex.Frame) {
	n := 0
	if m.vocab != nil {
		n = m.vocab.Count()
	}
	f.Row().Height(54).PadX(24).Items(flex.Center).Gap(10).Children(func() {
		f.Text("词库").FontSize(18.5).Weight(ui.WeightSemibold).Color(theme.Ink)
		if m.dictMode {
			f.Text(thousands(dict.Words())+" 个词条").FontSize(theme.SizeMicro).Color(theme.InkFaint).
				PadSides(3, 8, 3, 8).Radius(5).Background(theme.Surface2)
		} else {
			f.Text(fmt.Sprintf("%d 个词", n)).FontSize(theme.SizeMicro).Color(theme.InkFaint).
				PadSides(3, 8, 3, 8).Radius(5).Background(theme.Surface2)
		}
		if m.vocabBroken && !m.dictMode {
			f.Text("词库文件无法读取，这次不会写入").FontSize(theme.SizeMicro).Color(theme.Danger)
		}
		f.Spacer()
		m.bankSearchField(f)
		view := 0
		if m.dictMode {
			view = 1
		}
		if i, ok := f.Choice(bankViews, []string{"我的词", "全部词典"}, view).Width(168).ChosenIndex(); ok {
			m.setDictMode(i == 1)
		}
		if m.dictMode {
			return
		}
		add := f.Row().Items(flex.Center).Gap(5).PadSides(8, 12, 8, 10).Radius(7).Background(theme.Surface)
		if m.adding {
			add = add.Background(theme.AccentBg)
		}
		add.Children(func() {
			color := theme.InkSoft
			label := "手动添加"
			if m.adding {
				color, label = theme.Accent, "取消添加"
			}
			f.Text(label).FontSize(theme.SizeSmall).Color(color)
		})
		if add.Hit(bankAdd, "手动添加").Clicked() {
			m.adding = !m.adding
			if m.adding {
				m.addWord.SetText("")
				m.addZh.SetText("")
			}
		}
	})
}

// bankSearchField is the one search box of the surface: it sits in the masthead
// so it stays in reach of the word list and of the whole lexicon alike.
func (m *State) bankSearchField(f *flex.Frame) {
	placeholder := "词、中文或原句"
	if m.dictMode {
		placeholder = "词或前缀"
	}
	f.Row().Items(flex.Center).Gap(6).Children(func() {
		f.Field(&m.query, bankSearch, "搜索词库", placeholder).Width(228)
		if strings.TrimSpace(m.query.Text()) != "" {
			clr := f.Row().Height(28).PadSides(0, 8, 0, 8).Items(flex.Center).Radius(6).
				Background(theme.Surface2).HoverBackground(theme.DeskTrack)
			clr.Children(func() {
				f.Text("清空").FontSize(11).Color(theme.InkSoft)
			})
			if clr.Hit(bankClear, "清空").Clicked() {
				m.query.SetText("")
				m.bankPg = 0
				m.dictPg = 0
				m.dirty = true
			}
		}
	})
}

// bankPage lays out Scheme 04 with pagination:
//   - Top horizontal filter/alphabet bar + search box
//   - Main paginated Bento card wall + bottom pagination bar
//   - Right-hand slide-over detail drawer when a card is selected
func (m *State) bankPage(f *flex.Frame) {
	f.Column().Grow(1).FillWidth().Children(func() {
		// A filter row over nothing is furniture: the empty state speaks for
		// the surface until there is something to filter.
		if m.dictMode || m.bankCounts()[0] > 0 {
			m.bankTopBar(f)
		}
		f.Row().Grow(1).FillWidth().Children(func() {
			m.bankBentoArea(f)
			if m.selected != "" || m.adding {
				m.bankDrawer(f)
			}
		})
	})
}

var bankFilterLabels = []string{"全部", "生词", "学习中", "已掌握", "错词"}

// bankCounts is how many words each filter would show.
func (m *State) bankCounts() [5]int {
	var out [5]int
	if m.vocab == nil {
		return out
	}
	for _, w := range m.vocab.Words {
		out[0]++
		switch w.Stage {
		case vocab.StageLearning:
			out[2]++
		case vocab.StageKnown:
			out[3]++
		default:
			out[1]++
		}
		if w.Wrong > 0 {
			out[4]++
		}
	}
	return out
}

// bankTopBar draws the one secondary row of the surface: the stage filters (or
// the A–Z index of the lexicon). The share of mastered words rides the summary
// line instead, where it can sit inside the column of cards it describes.
func (m *State) bankTopBar(f *flex.Frame) {
	f.Column().FillWidth().Background(theme.Surface).Children(func() {
		f.Row().FillWidth().Height(46).PadSides(0, 24, 0, 24).Items(flex.Center).Gap(8).Children(func() {
			if m.dictMode {
				m.bankLettersStrip(f)
				return
			}
			m.bankFiltersStrip(f)
			f.Spacer()
		})
		f.Column().FillWidth().Height(1).Background(theme.RuleSoft)
	})
}

// bankFiltersStrip is the stage vocabulary: a dot and a count beside the word,
// the same dot the cards carry, so filter and card name the same thing.
func (m *State) bankFiltersStrip(f *flex.Frame) {
	counts := m.bankCounts()
	f.Row().Items(flex.Center).Gap(4).Children(func() {
		for i, label := range bankFilterLabels {
			active := i == m.filter
			b := f.Row().Height(32).Items(flex.Center).PadSides(0, 11, 0, 11).Gap(6).Radius(7)
			if active {
				b = b.Background(theme.Card).Border(1, theme.CardRule)
			} else {
				b = b.HoverBackground(theme.Surface2)
			}
			b.Children(func() {
				stageDot(f, i, 6)
				t := f.Text(label).FontSize(12.5)
				if active {
					t.Color(theme.Ink).Weight(ui.WeightSemibold)
				} else {
					t.Color(theme.InkSoft)
				}
				cnt := f.Text(fmt.Sprint(counts[i])).FontSize(11)
				if active {
					cnt.Color(theme.Accent).Weight(ui.WeightSemibold)
				} else {
					cnt.Color(theme.InkFaint)
				}
			})
			if b.Hit(bankChips.At(i), label).Clicked() {
				m.filter = i
				m.bankPg = 0
				if m.selected != "" {
					m.ensureBankSelection()
				}
			}
		}
	})
}

// bankMastery states how much of the collection is behind you; the bar is the
// only progress on the surface, and it belongs to the whole list, not a page.
// stageDot is the one mark of a word's stage: warm for a new word, green while
// learning, quiet once known, red for the ones answered wrong.
func stageDot(f *flex.Frame, filter int, size float32) {
	c := theme.InkFaint.WithAlpha(0.45)
	switch filter {
	case 1:
		c = theme.VerseNum.WithAlpha(0.75)
	case 2:
		c = theme.Accent
	case 3:
		c = theme.InkFaint.WithAlpha(0.35)
	case 4:
		c = theme.Danger.WithAlpha(0.75)
	}
	f.Column().Size(size, size).Radius(size / 2).Background(c)
}

// stageFilter maps a saved word's stage onto the filter vocabulary, so the
// dot beside it is the same dot the filter row uses.
func stageFilter(stage string) int {
	switch stage {
	case vocab.StageLearning:
		return 2
	case vocab.StageKnown:
		return 3
	default:
		return 1
	}
}

// stageMark is the card's quiet version of the same: the default state says
// nothing, so a wall of new words stays a wall of words.
func (m *State) stageMark(f *flex.Frame, stage string, wrong int) {
	if wrong > 0 {
		f.Text(fmt.Sprintf("错 %d", wrong)).FontSize(10).Color(theme.Danger)
	}
	switch stage {
	case vocab.StageLearning:
		stageDot(f, 2, 6)
	case vocab.StageKnown:
		stageDot(f, 3, 6)
	default:
		return
	}
	f.Text(stageLabel(stage)).FontSize(10.5).Color(theme.PaperSoft)
}

func (m *State) bankLettersStrip(f *flex.Frame) {
	starts := m.letterStarts()
	curLetter := m.currentDictLetter()
	f.Row().Items(flex.Center).Gap(2).Children(func() {
		allActive := m.dictLetter == 0 && m.dictPg == 0 && strings.TrimSpace(m.query.Text()) == ""
		all := f.Row().Height(26).PadSides(0, 9, 0, 9).Items(flex.Center).Justify(flex.Center).Radius(5)
		if allActive {
			all = all.Background(theme.Card).Border(1, theme.CardRule)
		} else {
			all = all.HoverBackground(theme.Surface2)
		}
		all.Children(func() {
			t := f.Text("全部").FontSize(11.5)
			if allActive {
				t.Color(theme.Ink).Weight(ui.WeightSemibold)
			} else {
				t.Color(theme.InkSoft)
			}
		})
		if all.Hit(bankLetters.At(0), "全部").Clicked() {
			m.jumpToLetter(starts, 0)
		}

		for i := 1; i <= 26; i++ {
			label := string(rune('A' + i - 1))
			active := !allActive && curLetter == i
			cell := f.Row().Width(23).Height(26).Items(flex.Center).Justify(flex.Center).Radius(5)
			if active {
				cell = cell.Background(theme.Card).Border(1, theme.CardRule)
			} else {
				cell = cell.HoverBackground(theme.Surface2)
			}
			cell.Children(func() {
				t := f.Text(label).FontSize(11.5)
				if active {
					t.Color(theme.Ink).Weight(ui.WeightSemibold)
				} else {
					t.Color(theme.InkSoft)
				}
			})
			if cell.Hit(bankLetters.At(i), label).Clicked() {
				m.jumpToLetter(starts, i)
			}
		}
	})
}

// matches applies the filter and the query to one word. The query searches the
// word, its Chinese, and the verses it was met in — which is how a word comes
// back to memory.
func (m *State) matches(w vocab.Word) bool {
	switch m.filter {
	case 1:
		if w.Stage == vocab.StageLearning || w.Stage == vocab.StageKnown {
			return false
		}
	case 2:
		if w.Stage != vocab.StageLearning {
			return false
		}
	case 3:
		if w.Stage != vocab.StageKnown {
			return false
		}
	case 4:
		if w.Wrong == 0 {
			return false
		}
	}
	q := strings.ToLower(strings.TrimSpace(m.query.Text()))
	if q == "" {
		return true
	}
	if strings.Contains(strings.ToLower(w.Word), q) || strings.Contains(strings.ToLower(w.Lemma), q) {
		return true
	}
	for _, zh := range w.Zh {
		if strings.Contains(zh, m.query.Text()) {
			return true
		}
	}
	for _, c := range w.Contexts {
		if strings.Contains(strings.ToLower(c.EN), q) || strings.Contains(c.ZH, m.query.Text()) {
			return true
		}
	}
	return false
}

// bankWords returns the words on show, newest first.
func (m *State) bankWords() []vocab.Word {
	if m.vocab == nil {
		return nil
	}
	out := make([]vocab.Word, 0, len(m.vocab.Words))
	for _, w := range m.vocab.Words {
		if m.matches(w) {
			out = append(out, w)
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Added.After(out[j].Added) })
	return out
}

// bankBentoArea renders the paginated Bento card wall and bottom pagination bar.
func (m *State) bankBentoArea(f *flex.Frame) {
	if m.dictMode {
		m.dictBentoList(f)
		return
	}
	words := m.bankWords()
	pages := pageCount(len(words), bankPageSize)
	m.bankPg = clampPage(m.bankPg, pages)
	start := m.bankPg * bankPageSize
	end := start + bankPageSize
	if end > len(words) {
		end = len(words)
	}
	pageWords := words[start:end]

	// An empty collection is not a list taller than the window: it fills the
	// wall it would have occupied, so the one thing to do sits in the middle of
	// the space rather than hiding under the filter row.
	if len(words) == 0 {
		f.Column().Grow(1).FillWidth().PadSides(20, 28, 24, 28).Children(func() {
			m.bankEmpty(f)
		})
		return
	}

	f.Column().Grow(1).FillHeight().Children(func() {
		f.ScrollArea(bankList).Grow(1).PadSides(20, 28, 24, 28).Children(func() {
			f.Column().MaxWidth(bankMax).FillWidth().Gap(16).Children(func() {
				f.Row().Items(flex.Center).Children(func() {
					summary := fmt.Sprintf("%d 张词卡 · 按加入时间", len(words))
					if strings.TrimSpace(m.query.Text()) != "" {
						summary = fmt.Sprintf("找到 %d 张词卡 · 按加入时间", len(words))
					}
					if counts := m.bankCounts(); counts[0] > 0 {
						summary += fmt.Sprintf(" · 已掌握 %d / %d", counts[3], counts[0])
					}
					f.Text(summary).FontSize(11).Color(theme.InkFaint)
					f.Spacer()
					f.Text("↑↓ 选词 · PgUp/PgDn 翻页 · Esc 收起").FontSize(10.5).Color(theme.InkFaint)
				})
				f.Grid(265).Gap(16).Children(func() {
					for i, w := range pageWords {
						m.bankBentoCard(f, start+i, w)
					}
				})
			})
		})
		if len(words) > 0 {
			m.bankPagerBar(f, m.bankPg, pages, len(words), bankPageSize, false)
		}
	})
}

// bankPagerBar is the quiet end of the surface: where you are in the list, and
// the way to the next stretch of it.
func (m *State) bankPagerBar(f *flex.Frame, cur, pages, total, pageSize int, isDict bool) {
	start := cur*pageSize + 1
	end := (cur + 1) * pageSize
	if end > total {
		end = total
	}

	f.Column().FillWidth().Background(theme.Surface).Children(func() {
		f.Column().FillWidth().Height(1).Background(theme.RuleSoft)
		f.Row().FillWidth().Height(46).PadSides(0, 24, 0, 24).Items(flex.Center).Gap(4).Children(func() {
			position := fmt.Sprintf("第 %s–%s 张 / 共 %s 张", thousands(start), thousands(end), thousands(total))
			if isDict {
				position = fmt.Sprintf("第 %s–%s 条 / 共 %s 条", thousands(start), thousands(end), thousands(total))
			}
			f.Text(position).FontSize(11).Color(theme.InkFaint)
			f.Spacer()

			prev := pagerStep(f, "‹ 上一页", cur > 0)
			if prev.Hit(bankPagePrev, "上一页").Clicked() && cur > 0 {
				m.setBankPage(cur-1, isDict)
			}
			for idx, p := range pagerWindow(cur, pages) {
				if p < 0 {
					f.Text("…").FontSize(11.5).Color(theme.InkFaint).PadSides(0, 3, 0, 3)
					continue
				}
				active := p == cur
				pill := f.Row().MinWidth(26).Height(26).PadSides(0, 6, 0, 6).
					Items(flex.Center).Justify(flex.Center).Radius(6)
				if active {
					pill = pill.Background(theme.Card).Border(1, theme.CardRule)
				} else {
					pill = pill.HoverBackground(theme.Surface2)
				}
				pill.Children(func() {
					t := f.Text(fmt.Sprint(p + 1)).FontSize(11.5)
					if active {
						t.Color(theme.Ink).Weight(ui.WeightSemibold)
					} else {
						t.Color(theme.InkSoft)
					}
				})
				if pill.Hit(bankPageNums.At(idx), fmt.Sprintf("第 %d 页", p+1)).Clicked() && !active {
					m.setBankPage(p, isDict)
				}
			}
			next := pagerStep(f, "下一页 ›", cur+1 < pages)
			if next.Hit(bankPageNext, "下一页").Clicked() && cur+1 < pages {
				m.setBankPage(cur+1, isDict)
			}
		})
	})
}

// pagerStep is one of the two step controls; a step that cannot be taken keeps
// its place and loses its ink.
func pagerStep(f *flex.Frame, label string, on bool) flex.Box {
	b := f.Row().Height(26).PadSides(0, 8, 0, 8).Items(flex.Center).Radius(6)
	if on {
		b = b.HoverBackground(theme.Surface2)
	}
	b.Children(func() {
		c := theme.InkSoft
		if !on {
			c = theme.InkFaint.WithAlpha(0.55)
		}
		f.Text(label).FontSize(11.5).Color(c)
	})
	return b
}

func (m *State) setBankPage(p int, isDict bool) {
	if isDict {
		m.dictPg = p
	} else {
		m.bankPg = p
		words := m.bankWords()
		start := p * bankPageSize
		end := start + bankPageSize
		if end > len(words) {
			end = len(words)
		}
		if start < len(words) && m.selected != "" {
			onPage := false
			for _, w := range words[start:end] {
				if strings.EqualFold(w.Word, m.selected) {
					onPage = true
					break
				}
			}
			if !onPage {
				m.selected = words[start].Word
			}
		}
	}
	if st := m.view.ScrollState(bankList); st != nil {
		st.ScrollTo(0)
	}
	m.dirty = true
}

// pagerWindow returns 0-based page indices to show as buttons, with -1 representing an ellipsis.
func pagerWindow(cur, pages int) []int {
	if pages <= 7 {
		out := make([]int, pages)
		for i := range out {
			out[i] = i
		}
		return out
	}
	out := []int{0}
	low := cur - 1
	high := cur + 1
	if cur <= 2 {
		low, high = 1, 3
	} else if cur >= pages-3 {
		low, high = pages-4, pages-2
	}
	if low > 1 {
		out = append(out, -1)
	}
	for p := low; p <= high; p++ {
		if p > 0 && p < pages-1 {
			out = append(out, p)
		}
	}
	if high < pages-2 {
		out = append(out, -1)
	}
	out = append(out, pages-1)
	return out
}

// bankEmpty is the surface with no cards in it: a quiet welcome that says how a
// word gets here, and the one or two things worth doing next, instead of a
// blank wall with a filter bar above it.
func (m *State) bankEmpty(f *flex.Frame) {
	hasWords := m.vocab != nil && m.vocab.Count() > 0
	searching := strings.TrimSpace(m.query.Text()) != ""
	filtered := hasWords && (m.filter != 0 || searching)

	title := "词库还空着"
	desc := "阅读时点按经文里的生词，卡片上的「收藏这个词」会把词和经节一起收进来。"
	if filtered {
		title = "没有符合条件的词"
		desc = "换个阶段筛选，或者清空搜索词，再看看。"
	}

	f.Column().Grow(1).FillWidth().Justify(flex.Center).Items(flex.Center).Children(func() {
		f.Column().MaxWidth(520).FillWidth().PadSides(34, 38, 34, 32).Gap(9).Radius(14).
			Background(theme.Card).Border(1, theme.CardRule).Shadow(ui.ShadowSM).
			Items(flex.Center).Children(func() {
			f.Image(logo.Bitmap()).Size(24, 24)
			f.Spacer().Height(4)
			f.Text(title).FontSize(16.5).Weight(ui.WeightSemibold).Color(theme.PaperInk).TextCenter()
			f.Wrap(desc).FontSize(12.5).LineHeight(19.5).Color(theme.PaperSoft).TextCenter()
			f.Spacer().Height(6)

			if filtered {
				f.Row().Gap(8).Items(flex.Center).Children(func() {
					if m.filter != 0 {
						if emptyChip(f, bankEmptyAll, "查看全部词", true) {
							m.filter = 0
							m.bankPg = 0
						}
					}
					if searching {
						if emptyChip(f, bankEmptyClear, "清空搜索", m.filter == 0) {
							m.query.SetText("")
							m.bankPg = 0
						}
					}
				})
				return
			}
			f.Row().Gap(8).Items(flex.Center).Children(func() {
				if emptyChip(f, bankEmptyDict, "去阅读", true) {
					m.pane = paneReader
					m.dirty = true
					m.remember()
				}
				if emptyChip(f, bankEmptyAdd, "手动添加一个词", false) {
					m.adding = true
					m.addWord.SetText("")
					m.addZh.SetText("")
				}
			})
			f.Button(bankEmptyBrowse, "浏览全部词典 ›").Link().XSmall()
		})
	})
}

// emptyChip is one action of the empty state: filled for the way forward,
// outlined for the alternative beside it.
func emptyChip(f *flex.Frame, id ui.ID, label string, primary bool) bool {
	b := f.Row().Height(32).Items(flex.Center).PadSides(0, 15, 0, 15).Radius(8)
	if primary {
		b = b.Background(theme.Accent).HoverBackground(theme.Accent.WithAlpha(0.88))
	} else {
		b = b.Background(theme.Card).Border(1, theme.CardRule).HoverBackground(theme.CardInset)
	}
	b.Children(func() {
		t := f.Text(label).FontSize(12.5).Weight(ui.WeightSemibold)
		if primary {
			t.Color(theme.OnAccent)
		} else {
			t.Color(theme.PaperInk)
		}
	})
	return b.Hit(id, label).Clicked()
}

// thousands groups a count so forty thousand words read as 41,873.
func thousands(n int) string {
	s := fmt.Sprint(n)
	for i := len(s) - 3; i > 0; i -= 3 {
		s = s[:i] + "," + s[i:]
	}
	return s
}

// stripIPA drops the slashes or brackets the lexicon sometimes keeps around a
// pronunciation, so a card and a list row read the same.
func stripIPA(s string) string {
	return strings.Trim(strings.TrimSpace(s), "/[]")
}

// formatIPA wraps a phonetic transcription in slashes when present.
func formatIPA(s string) string {
	raw := stripIPA(s)
	if raw == "" {
		return ""
	}
	return "/" + raw + "/"
}

func stageLabel(stage string) string {
	switch stage {
	case vocab.StageLearning:
		return "学习中"
	case vocab.StageKnown:
		return "已掌握"
	}
	return "生词"
}

// bankBentoCard is one page of the index: the word, how it sounds, what it
// means, and the verse it came from. The stage is silent when the word is new,
// which is what most of a fresh collection is, and a dot when it is not.
func (m *State) bankBentoCard(f *flex.Frame, i int, w vocab.Word) {
	selected := strings.EqualFold(w.Word, m.selected)
	card := f.Column().MinHeight(152).PadSides(16, 18, 15, 18).Gap(5).Radius(10).
		Background(theme.Card).HoverBackground(theme.Paper)
	if selected {
		card = card.Background(theme.Paper).Border(1, theme.Accent).Shadow(ui.ShadowSM)
	} else {
		card = card.Border(1, theme.CardRule)
	}
	card.Children(func() {
		f.Row().Items(flex.Center).Gap(7).Children(func() {
			f.Text(w.Word).FontSize(18).Weight(ui.WeightSemibold).Color(theme.PaperInk).Ellipsis()
			if w.Pos != "" {
				f.Text(w.Pos).FontSize(11).Color(theme.PaperSoft).Ellipsis()
			}
			f.Spacer()
			m.stageMark(f, w.Stage, w.Wrong)
		})

		if ipa := formatIPA(w.IPA); ipa != "" {
			f.Text(ipa).FontSize(12).Color(theme.PaperSoft).Ellipsis()
		}

		if len(w.Zh) > 0 {
			f.Text(strings.Join(w.Zh, "；")).FontSize(14).LineHeight(20).
				Weight(ui.WeightMedium).Color(theme.PaperInk).MaxLines(2)
		} else if len(w.Senses) > 0 {
			f.Text(w.Senses[0]).FontSize(12.5).LineHeight(18).Color(theme.PaperSoft).MaxLines(2)
		}

		f.Spacer()
		f.Column().FillWidth().Height(1).Background(theme.PaperRule)

		switch {
		case len(w.Contexts) > 0:
			c0 := w.Contexts[0]
			book, _ := m.data.Book(c0.Book)
			f.Row().Items(flex.Center).Gap(6).Children(func() {
				f.Text(fmt.Sprintf("%s %d:%d", m.bookName(book), c0.Chapter, c0.Verse)).
					FontSize(10.5).Weight(ui.WeightSemibold).Color(theme.PaperSoft)
				if len(w.Contexts) > 1 {
					f.Spacer()
					f.Text(fmt.Sprintf("共 %d 处", len(w.Contexts))).FontSize(10).Color(theme.PaperSoft)
				}
			})
			if c0.EN != "" {
				f.Text(c0.EN).FontSize(11.5).LineHeight(17).Color(theme.PaperSoft).MaxLines(2)
			}
		case m.hasLexiconVerses(w.Word):
			f.Text("圣经出现 " + thousands(m.lexiconVerses(w.Word)) + " 处").FontSize(10.5).Color(theme.PaperSoft)
		default:
			f.Text("手动添加").FontSize(10.5).Color(theme.PaperSoft)
		}
	})
	if card.Hit(bankRows.At(i), w.Word).Clicked() {
		m.selected = w.Word
		m.adding = false
	}
}

// hasLexiconVerses and lexiconVerses answer how often the lexicon saw a word in
// scripture; the reader's own verseCount counts something else.
func (m *State) hasLexiconVerses(word string) bool {
	lex, ok := m.dictEntry(word)
	return ok && lex.Verses > 0
}

func (m *State) lexiconVerses(word string) int {
	lex, _ := m.dictEntry(word)
	return lex.Verses
}

// dictBentoList renders the 41,873 lexicon entries in pages of dictPageSize (24
// cards per page) so it never dumps the whole lexicon at once.
func (m *State) dictBentoList(f *flex.Frame) {
	words := m.dictWords()
	pages := pageCount(len(words), dictPageSize)
	m.dictPg = clampPage(m.dictPg, pages)
	start := m.dictPg * dictPageSize
	end := start + dictPageSize
	if end > len(words) {
		end = len(words)
	}
	pageWords := words[start:end]

	// An empty collection is not a list taller than the window: it fills the
	// wall it would have occupied, so the one thing to do sits in the middle of
	// the space rather than hiding under the filter row.
	if len(words) == 0 {
		f.Column().Grow(1).FillWidth().PadSides(20, 28, 24, 28).Children(func() {
			m.bankEmpty(f)
		})
		return
	}

	f.Column().Grow(1).FillHeight().Children(func() {
		f.ScrollArea(bankList).Grow(1).PadSides(18, 28, 24, 28).Children(func() {
			f.Column().MaxWidth(bankMax).FillWidth().Gap(12).Children(func() {
				if len(words) == 0 {
					f.Text("没有匹配的词典词条。").FontSize(13).Color(theme.InkFaint)
					return
				}
				f.Row().Items(flex.Center).Children(func() {
					summary := fmt.Sprintf("%s 个词条 · 按字母排序", thousands(len(words)))
					if strings.TrimSpace(m.query.Text()) != "" {
						summary = fmt.Sprintf("找到 %s 个词条 · 按字母排序", thousands(len(words)))
					}
					f.Text(summary).FontSize(11).Color(theme.InkFaint)
					f.Spacer()
					f.Text("A–Z 跳字母 · PgUp/PgDn 翻页").FontSize(10.5).Color(theme.InkFaint)
				})
				f.Grid(250).Gap(10).Children(func() {
					for i, w := range pageWords {
						m.dictBentoCard(f, start+i, w)
					}
				})
			})
		})
		if len(words) > 0 {
			m.bankPagerBar(f, m.dictPg, pages, len(words), dictPageSize, true)
		}
	})
}

func (m *State) dictBentoCard(f *flex.Frame, i int, word string) {
	selected := strings.EqualFold(word, m.selected)
	b := f.Column().MinHeight(90).PadSides(11, 14, 10, 14).Gap(3).Radius(9).
		Background(theme.Card).HoverBackground(theme.Paper)
	if selected {
		b = b.Border(1, theme.Accent).Shadow(ui.ShadowSM)
	} else {
		b = b.Border(1, theme.CardRule)
	}
	b.Children(func() {
		e, ok := m.dictEntry(word)
		f.Row().Items(flex.Center).Gap(6).Children(func() {
			f.Text(word).FontSize(15.5).Weight(ui.WeightSemibold).Color(theme.PaperInk).Ellipsis()
			if ok && e.Pos != "" {
				f.Text(e.Pos).FontSize(10.5).Color(theme.PaperSoft)
			}
			f.Spacer()
			if m.vocab != nil && m.vocab.Known(word) {
				f.Text("已收").FontSize(10).Color(theme.PaperSoft)
			} else if ok && e.Verses > 0 {
				f.Text(fmt.Sprintf("%d 处", e.Verses)).FontSize(10).Color(theme.PaperSoft)
			}
		})
		if !ok {
			return
		}
		if ipa := formatIPA(e.IPA); ipa != "" {
			f.Text(ipa).FontSize(11.5).Color(theme.PaperSoft).Ellipsis()
		}
		if len(e.Zh) > 0 {
			f.Text(strings.Join(e.Zh, "；")).FontSize(12.5).LineHeight(18).Color(theme.PaperSoft).MaxLines(2)
		} else if len(e.Senses) > 0 {
			f.Text(e.Senses[0]).FontSize(12).LineHeight(17).Color(theme.PaperSoft).MaxLines(2)
		}
	})
	if b.Hit(bankDict.At(i), word).Clicked() {
		m.selected = word
	}
}

// dictWords is the headwords on show: the whole lexicon (or active letter), or
// what the query matches. A prefix is a binary search into the sorted index;
// anything else is a scan capped so one letter cannot paint the lexicon.
func (m *State) dictWords() []string {
	words := dict.Headwords()
	q := strings.ToLower(strings.TrimSpace(m.query.Text()))
	if q == "" {
		if m.dictLetter >= 1 && m.dictLetter <= 26 {
			starts := m.letterStarts()
			lo := starts[m.dictLetter]
			hi := len(words)
			if m.dictLetter < 26 {
				hi = starts[m.dictLetter+1]
			}
			if lo >= 0 && lo <= hi && hi <= len(words) {
				return words[lo:hi]
			}
		}
		return words
	}
	if i := sort.SearchStrings(words, q); i < len(words) && strings.HasPrefix(words[i], q) {
		j := i
		for j < len(words) && strings.HasPrefix(words[j], q) {
			j++
		}
		return words[i:j]
	}
	out := make([]string, 0, 200)
	for _, w := range words {
		if strings.Contains(strings.ToLower(w), q) {
			out = append(out, w)
			if len(out) == 200 {
				break
			}
		}
	}
	return out
}

// dictEntry reads one record and remembers it: scrolling repaints the same rows
// again and again, and every record is a small inflate.
func (m *State) dictEntry(word string) (dict.Entry, bool) {
	if e, ok := m.dictCache[word]; ok {
		return e, true
	}
	e, ok := dict.Lookup(word)
	if !ok {
		return dict.Entry{}, false
	}
	if m.dictCache == nil || len(m.dictCache) > 4096 {
		m.dictCache = make(map[string]dict.Entry, 256)
	}
	m.dictCache[word] = e
	return e, true
}

// currentDictLetter returns 1..26 for the active letter filter or the letter on
// the current dictionary page (or the selected word's initial), or 0 when on the first page.
func (m *State) currentDictLetter() int {
	if m.dictLetter >= 1 && m.dictLetter <= 26 && strings.TrimSpace(m.query.Text()) == "" {
		return m.dictLetter
	}
	if m.selected != "" {
		r := rune(strings.ToLower(m.selected)[0])
		if r >= 'a' && r <= 'z' {
			return int(r-'a') + 1
		}
	}
	idx := m.dictPg * dictPageSize
	words := m.dictWords()
	if idx >= len(words) {
		idx = len(words) - 1
	}
	if idx >= 0 && len(words[idx]) > 0 {
		r := rune(strings.ToLower(words[idx])[0])
		if r >= 'a' && r <= 'z' {
			return int(r-'a') + 1
		}
	}
	return 0
}

// letterStarts is where each letter begins in the headword list, so a letter is
// a page target rather than a search.
func (m *State) letterStarts() []int {
	if m.letterAt != nil {
		return m.letterAt
	}
	words := dict.Headwords()
	m.letterAt = make([]int, 27)
	for i := 1; i <= 26; i++ {
		m.letterAt[i] = sort.SearchStrings(words, string(rune('a'+i-1)))
	}
	return m.letterAt
}

// jumpToLetter filters to the chosen letter (or 0 for all) starting at page 1.
func (m *State) jumpToLetter(_ []int, i int) {
	m.query.SetText("")
	m.dictLetter = i
	m.dictPg = 0
	if st := m.view.ScrollState(bankList); st != nil {
		st.ScrollTo(0)
	}
	m.dirty = true
}

// setDictMode switches the main wall between the saved words and the whole lexicon.
func (m *State) setDictMode(on bool) {
	if m.dictMode == on {
		return
	}
	m.dictMode = on
	m.selected = ""
	m.adding = false
	m.query.SetText("")
	m.dictLetter = 0
	m.dictPg = 0
	m.bankPg = 0
	if !on {
		m.ensureBankSelection()
	}
	m.dirty = true
}

// saveDictWord puts the open dictionary word into the bank.
func (m *State) saveDictWord() {
	if m.vocab == nil {
		return
	}
	e, ok := m.dictEntry(m.selected)
	if !ok {
		return
	}
	w := vocab.Word{Word: e.Word, Stage: vocab.StageNew}
	w.Lemma, w.Pos, w.IPA, w.Zh, w.Senses = e.Lemma, e.Pos, stripIPA(e.IPA), e.Zh, e.Senses
	m.vocab.Add(w)
	m.saveVocab()
}

// bankDrawer renders the right-hand slide-over detail drawer when a card is open.
func (m *State) bankDrawer(f *flex.Frame) {
	f.Row().FillHeight().Children(func() {
		f.Column().Width(1).FillHeight().Background(theme.CardRule)
		f.Column().Width(drawerWidth).FillHeight().Background(theme.Card).Shadow(ui.ShadowMD).Children(func() {
			// The drawer names itself once; everything under it is the word.
			f.Row().Height(46).PadSides(0, 26, 0, 22).Items(flex.Center).Children(func() {
				title := "词卡详情"
				if m.dictMode {
					title = "词典词条"
				}
				if m.adding {
					title = "手动添加"
				}
				f.Text(title).FontSize(10.5).Weight(ui.WeightSemibold).Color(theme.PaperSoft)
				f.Spacer()
				closeBtn := f.Row().PadSides(3, 8, 3, 8).Radius(5).HoverBackground(theme.CardInset)
				closeBtn.Children(func() {
					f.Text("收起 ✕").FontSize(11).Color(theme.PaperSoft)
				})
				if closeBtn.Hit(bankDrawerClose, "收起详情").Clicked() {
					m.adding = false
					m.selected = ""
				}
			})
			f.Column().FillWidth().Height(1).Background(theme.PaperRule)

			// Drawer scrollable body
			f.ScrollArea(bankPane).Grow(1).PadSides(22, 26, 26, 26).Children(func() {
				if m.adding {
					m.addDrawerBody(f)
					return
				}
				if m.dictMode {
					m.dictDrawerBody(f)
					return
				}
				w := m.selectedWord()
				if w == nil {
					if _, ok := m.dictEntry(m.selected); ok {
						m.dictDrawerBody(f)
						return
					}
					f.Text("点按左侧任意词卡查看详情。").FontSize(13).Color(theme.PaperSubtle)
					return
				}
				m.wordDrawerBody(f, w)
			})
		})
	})
}

func (m *State) dictDrawerBody(f *flex.Frame) {
	e, ok := m.dictEntry(m.selected)
	if !ok {
		f.Text("词典里还没有这个词。").FontSize(13).Color(theme.PaperSoft)
		return
	}
	saved := m.vocab != nil && m.vocab.Known(e.Word)

	f.Column().Gap(16).Children(func() {
		f.Column().Gap(5).Children(func() {
			f.Row().Items(flex.Center).Gap(7).Children(func() {
				f.Text(e.Word).FontSize(30).Weight(ui.WeightSemibold).Color(theme.PaperInk)
				if e.Pos != "" {
					f.Text(e.Pos).FontSize(12).Color(theme.PaperSoft)
				}
			})
			ipa := formatIPA(e.IPA)
			lemma := e.Lemma != "" && !strings.EqualFold(e.Lemma, e.Word)
			if ipa != "" || e.Verses > 0 || lemma {
				f.Row().Items(flex.Center).Gap(10).Children(func() {
					if ipa != "" {
						f.Text(ipa).FontSize(12.5).Color(theme.PaperSoft)
					}
					if e.Verses > 0 {
						f.Text("圣经出现 " + thousands(e.Verses) + " 处").FontSize(11).Color(theme.PaperSoft)
					}
					if lemma {
						lem := f.Row().Items(flex.Center).Gap(4).PadSides(2, 7, 2, 7).Radius(5).
							Background(theme.CardInset).HoverBackground(theme.PaperStack)
						lem.Children(func() {
							f.Text("原形").FontSize(10.5).Color(theme.PaperSoft)
							f.Text(e.Lemma + " ›").FontSize(11.5).Weight(ui.WeightSemibold).Color(theme.Accent)
						})
						if lem.Hit(bankLemmaJump, "原形是 "+e.Lemma).Clicked() {
							if _, exists := m.dictEntry(e.Lemma); exists {
								m.selected = e.Lemma
								m.syncPageToSelected()
							}
						}
					}
					f.Spacer()
				})
			}
		})

		// Saving is the entry's one action, so it is the only thing wearing the
		// accent in the drawer.
		label := "收藏这个词"
		if saved {
			label = "已收藏 · 移出"
		}
		f.Row().Items(flex.Center).Children(func() {
			act := f.Row().Height(30).PadSides(0, 13, 0, 13).Items(flex.Center).Radius(6)
			if saved {
				act = act.Background(theme.CardInset).Border(1, theme.CardRule).HoverBackground(theme.PaperStack)
			} else {
				act = act.Background(theme.Accent)
			}
			act.Children(func() {
				t := f.Text(label).FontSize(11.5).Weight(ui.WeightSemibold)
				if saved {
					t.Color(theme.PaperSoft)
				} else {
					t.Color(theme.OnAccent)
				}
			})
			if act.Hit(bankSave, label).Clicked() {
				if saved {
					m.vocab.Remove(e.Word)
					m.saveVocab()
				} else {
					m.saveDictWord()
				}
			}
			f.Spacer()
		})

		f.Column().FillWidth().Height(1).Background(theme.PaperRule)

		// The Chinese sense leads; the English one supports it.
		f.Column().Gap(14).Children(func() {
			if len(e.Zh) > 0 {
				f.Wrap(strings.Join(e.Zh, "；")).FontSize(17).LineHeight(26).
					Weight(ui.WeightMedium).Color(theme.PaperInk)
			}
			if len(e.Senses) > 0 {
				f.Column().Gap(4).Children(func() {
					drawerLabel(f, "英文释义")
					f.Wrap(strings.Join(e.Senses, "; ")).FontSize(13).LineHeight(20).Color(theme.PaperSoft)
				})
			}
			if len(e.Zh) == 0 && len(e.Senses) == 0 {
				f.Text("这个词条只有读音，还没有释义。").FontSize(13).Color(theme.PaperSoft)
			}
			if len(e.Collocations) > 0 {
				f.Column().Gap(4).Children(func() {
					drawerLabel(f, "本书搭配")
					f.Wrap(strings.Join(e.Collocations, " · ")).FontSize(12.5).LineHeight(19.5).Color(theme.PaperSoft)
				})
			}
		})
	})
}

// selectedWord is the open word, or nil.
func (m *State) selectedWord() *vocab.Word {
	if m.vocab == nil || m.selected == "" {
		return nil
	}
	if i := m.vocab.Find(m.selected); i >= 0 {
		return &m.vocab.Words[i]
	}
	return nil
}

func (m *State) wordDrawerBody(f *flex.Frame, w *vocab.Word) {
	lex, hasLex := m.dictEntry(w.Word)
	f.Column().Gap(16).Children(func() {
		// The word itself, then the lexicon line: how it sounds, how often the
		// text says it, and which form it came from.
		f.Column().Gap(5).Children(func() {
			f.Row().Items(flex.Center).Gap(7).Children(func() {
				f.Text(w.Word).FontSize(30).Weight(ui.WeightSemibold).Color(theme.PaperInk)
				if w.Pos != "" {
					f.Text(w.Pos).FontSize(12).Color(theme.PaperSoft)
				}
			})
			ipa := formatIPA(w.IPA)
			verses := 0
			if hasLex {
				verses = lex.Verses
			}
			lemma := w.Lemma != "" && !strings.EqualFold(w.Lemma, w.Word)
			if ipa != "" || verses > 0 || lemma {
				f.Row().Items(flex.Center).Gap(10).Children(func() {
					if ipa != "" {
						f.Text(ipa).FontSize(12.5).Color(theme.PaperSoft)
					}
					if verses > 0 {
						f.Text("圣经出现 " + thousands(verses) + " 处").FontSize(11).Color(theme.PaperSoft)
					}
					if lemma {
						lem := f.Row().Items(flex.Center).Gap(4).PadSides(2, 7, 2, 7).Radius(5).
							Background(theme.CardInset).HoverBackground(theme.PaperStack)
						lem.Children(func() {
							f.Text("原形").FontSize(10.5).Color(theme.PaperSoft)
							f.Text(w.Lemma + " ›").FontSize(11.5).Weight(ui.WeightSemibold).Color(theme.Accent)
						})
						if lem.Hit(bankLemmaJump, "原形是 "+w.Lemma).Clicked() {
							if _, exists := m.dictEntry(w.Lemma); exists {
								m.selected = w.Lemma
								m.syncPageToSelected()
							}
						}
					}
					f.Spacer()
				})
			}
		})

		// The Chinese sense is the one line worth setting large: it is what the
		// drawer was opened to read.
		if len(w.Zh) > 0 {
			f.Wrap(strings.Join(w.Zh, "；")).FontSize(17).LineHeight(26).
				Weight(ui.WeightMedium).Color(theme.PaperInk)
		}

		if m.vocab != nil {
			m.stagePicker(f, w)
		}

		f.Column().FillWidth().Height(1).Background(theme.PaperRule)

		// The rest is support: the English sense, the book's own phrasing.
		f.Column().Gap(14).Children(func() {
			if len(w.Senses) > 0 {
				f.Column().Gap(4).Children(func() {
					drawerLabel(f, "英文释义")
					f.Wrap(strings.Join(w.Senses, "; ")).FontSize(13).LineHeight(20).Color(theme.PaperSoft)
				})
			}
			if len(w.Zh) == 0 && len(w.Senses) == 0 {
				f.Text("词典未收录，只有你写下的释义。").FontSize(13).Color(theme.PaperSoft)
			}
			if hasLex && len(lex.Collocations) > 0 {
				f.Column().Gap(4).Children(func() {
					drawerLabel(f, "本书搭配")
					f.Wrap(strings.Join(lex.Collocations, " · ")).FontSize(12.5).LineHeight(19.5).Color(theme.PaperSoft)
				})
			}
		})

		if len(w.Contexts) > 0 {
			f.Column().FillWidth().Height(1).Background(theme.PaperRule)
			f.Column().Gap(10).Children(func() {
				f.Row().Items(flex.Center).Children(func() {
					drawerLabel(f, "遇到的经文")
					f.Spacer()
					f.Text(fmt.Sprintf("%d 处", len(w.Contexts))).FontSize(10.5).Color(theme.PaperSoft)
				})
				for i, c := range w.Contexts {
					m.contextCard(f, i, c, w.Word)
				}
			})
		}

		f.Column().FillWidth().Height(1).Background(theme.PaperRule)

		f.Row().Items(flex.Center).Gap(10).Children(func() {
			f.Text(fmt.Sprintf("%s 加入 · 对 %d · 错 %d", w.Added.Local().Format("2006-01-02"), w.Right, w.Wrong)).
				FontSize(10.5).Color(theme.PaperSoft)
			f.Spacer()
			if f.Button(bankClose, "移出词库").Link().XSmall().Destructive().Clicked() {
				m.confirmDelete = w.Word
			}
		})
	})
}

// stagePicker is the stage control: the three names the filters already
// taught, set as text so the top of the drawer stays about the word.
func (m *State) stagePicker(f *flex.Frame, w *vocab.Word) {
	names := []string{"生词", "学习中", "已掌握"}
	stages := []string{vocab.StageNew, vocab.StageLearning, vocab.StageKnown}
	current := 0
	switch w.Stage {
	case vocab.StageLearning:
		current = 1
	case vocab.StageKnown:
		current = 2
	}

	f.Row().Items(flex.Center).Children(func() {
		drawerLabel(f, "阶段")
		for i, name := range names {
			active := i == current
			seg := f.Row().Height(26).PadSides(0, 9, 0, 9).Items(flex.Center).Gap(5).Radius(5)
			if active {
				seg = seg.Background(theme.Card).Border(1, theme.CardRule)
			} else {
				seg = seg.HoverBackground(theme.CardInset)
			}
			seg.Children(func() {
				if active {
					stageDot(f, i+1, 5)
				}
				t := f.Text(name).FontSize(11.5)
				if active {
					t.Weight(ui.WeightSemibold).Color(theme.Ink)
				} else {
					t.Color(theme.InkFaint)
				}
			})
			if seg.Hit(bankStage.At(i), name).Clicked() && !active {
				if m.vocab.SetStage(w.Word, stages[i]) {
					m.saveVocab()
				}
			}
		}
		f.Spacer()
	})
}

// drawerLabel is the small name above a block of the card.
func drawerLabel(f *flex.Frame, title string) {
	f.Text(title).FontSize(10.5).Weight(ui.WeightSemibold).Color(theme.PaperSoft)
}

func (m *State) contextCard(f *flex.Frame, i int, c vocab.Context, word string) {
	book, _ := m.data.Book(c.Book)
	f.Column().FillWidth().Radius(8).Background(theme.CardInset).
		PadSides(10, 13, 11, 13).Gap(5).Children(func() {
		f.Row().Items(flex.Center).Gap(6).Children(func() {
			f.Text(fmt.Sprintf("%s %d:%d", m.bookName(book), c.Chapter, c.Verse)).
				FontSize(11).Weight(ui.WeightSemibold).Color(theme.PaperSoft)
			if c.Translation != "" {
				f.Text(c.Translation).FontSize(10).Color(theme.PaperSoft)
			}
			f.Spacer()
			if f.Button(bankWords.At(i), "回到原文").Link().XSmall().Clicked() {
				m.openPassage(c, word)
			}
		})
		if c.EN != "" {
			f.Wrap(c.EN).FontSize(13).LineHeight(19.5).Color(theme.PaperInk)
		}
		if c.ZH != "" {
			f.Wrap(c.ZH).FontSize(12).LineHeight(18).Color(theme.PaperSoft)
		}
	})
}

// openPassage takes the reader back to the verse a word was met in.
func (m *State) openPassage(c vocab.Context, word string) {
	if b, ok := m.data.Book(c.Book); ok {
		m.osis = b.OSIS
	}
	if c.Chapter > 0 {
		m.chapter = c.Chapter
	}
	m.landVerse = c.Verse
	m.landWord = word
	m.landWordVerse = c.Verse
	m.spread = 0
	m.layoutKey = ""
	m.lookup = nil
	m.pane = paneReader
	m.dirty = true
	m.remember()
}

// addDrawerBody is the manual add form, shaped for the drawer: one column, so
// the fields are as wide as the card they will be saved next to.
func (m *State) addDrawerBody(f *flex.Frame) {
	f.Column().Gap(16).Children(func() {
		f.Column().Gap(6).Children(func() {
			f.Text("添加一个词").FontSize(24).Weight(ui.WeightSemibold).Color(theme.PaperInk)
			f.Text("阅读时点按生词即可连同经节收入；这里补上词典还没有收的词。").
				FontSize(12.5).LineHeight(19).Color(theme.PaperSoft)
		})

		f.Column().FillWidth().Height(1).Background(theme.PaperRule)

		f.Column().Gap(5).Children(func() {
			drawerLabel(f, "单词")
			if f.Field(&m.addWord, bankQuote, "单词", "例如 covenant").Submitted() {
				m.addManual()
			}
		})
		f.Column().Gap(5).Children(func() {
			drawerLabel(f, "中文释义（选填）")
			if f.Field(&m.addZh, bankZh, "中文释义", "留空则使用词典释义").Submitted() {
				m.addManual()
			}
		})

		f.Text("词典已收录的词会自动补齐音标与英译。").FontSize(11).Color(theme.PaperSoft)

		f.Row().Gap(10).Items(flex.Center).Children(func() {
			add := f.Row().Height(32).Grow(1).Items(flex.Center).Justify(flex.Center).
				Radius(7).Background(theme.Accent)
			add.Children(func() {
				f.Text("添加").FontSize(12).Weight(ui.WeightSemibold).Color(theme.OnAccent)
			})
			if add.Hit(bankSubmit, "添加").Clicked() {
				m.addManual()
			}
			if f.Button(bankCancel, "取消").Link().XSmall().Clicked() {
				m.adding = false
			}
		})
	})
}

func (m *State) addManual() {
	word := strings.TrimSpace(m.addWord.Text())
	if word == "" || m.vocab == nil {
		return
	}
	sv := vocab.Word{Word: word, Stage: vocab.StageNew}
	if zh := strings.TrimSpace(m.addZh.Text()); zh != "" {
		sv.Zh = []string{zh}
	}
	if entry, ok := dict.Lookup(word); ok {
		sv.Lemma, sv.Pos, sv.IPA = entry.Lemma, entry.Pos, entry.IPA
		if len(sv.Zh) == 0 {
			sv.Zh = entry.Zh
		}
		sv.Senses = entry.Senses
	}
	i := m.vocab.Add(sv)
	m.selected = m.vocab.Words[i].Word
	m.bankPg = 0
	m.adding = false
	m.addWord.SetText("")
	m.addZh.SetText("")
	m.saveVocab()
}

// saveLookup saves the word on the reading card, or takes it out again.
func (m *State) toggleLookup() {
	lk := m.lookup
	if lk == nil || m.vocab == nil {
		return
	}
	if m.vocab.Known(lk.word) {
		m.vocab.Remove(lk.word)
		if strings.EqualFold(m.selected, lk.word) {
			m.selected = ""
		}
	} else {
		w := vocab.Word{Word: lk.word, Stage: vocab.StageNew}
		if entry, ok := dict.Lookup(lk.word); ok {
			w.Lemma, w.Pos, w.IPA, w.Zh, w.Senses = entry.Lemma, entry.Pos, stripIPA(entry.IPA), entry.Zh, entry.Senses
		}
		w.Contexts = []vocab.Context{{
			Translation: m.left, Book: lk.osis, Chapter: lk.ch, Verse: lk.v,
			EN: lk.en, ZH: lk.zh,
		}}
		m.vocab.Add(w)
	}
	m.saveVocab()
}

// saveVocab writes the list; a broken file is left alone so its words are not
// overwritten by this session's empty one.
func (m *State) saveVocab() {
	if m.vocab == nil || m.vocabBroken {
		return
	}
	_ = vocab.Save(m.vocab)
	m.reindex()
}

// reindex keeps the reading page's word marks off the saved list's walk: a word
// is looked up once per painted word, not once per saved word.
func (m *State) reindex() {
	m.vocabIndex = make(map[string]vocab.Word, len(m.vocab.Words))
	if m.vocab == nil {
		return
	}
	for _, w := range m.vocab.Words {
		m.vocabIndex[strings.ToLower(strings.TrimSpace(w.Word))] = w
	}
}

// loadVocab reads the word list in the directory state.json lives in, and moves
// the bookmark-era word list over once.
func (m *State) loadVocab() {
	vocab.Dir = store.Dir
	f, err := vocab.Load()
	if err != nil {
		m.vocabBroken = true
	}
	if f != nil {
		m.vocab = f
	} else {
		m.vocab = vocab.Empty()
	}
	m.reindex()
	st, err := store.Load()
	if err != nil || len(st.Saved) == 0 {
		return
	}
	for _, word := range st.Saved {
		word = strings.TrimSpace(word)
		if word == "" {
			continue
		}
		w := vocab.Word{Word: word, Stage: vocab.StageNew}
		if entry, ok := dict.Lookup(word); ok {
			w.Lemma, w.Pos, w.IPA, w.Zh, w.Senses = entry.Lemma, entry.Pos, stripIPA(entry.IPA), entry.Zh, entry.Senses
		}
		m.vocab.Add(w)
	}
	if m.vocabBroken {
		return
	}
	if err := vocab.Save(m.vocab); err != nil {
		return
	}
	m.reindex()
	st.Saved = nil
	_ = store.Save(st)
}

// pageOfVerse is the page where a verse is, or -1.
func (m *State) pageOfVerse(v int) int {
	for i, p := range m.pages {
		for _, b := range p.blocks {
			if b.v == v {
				return i
			}
		}
	}
	return -1
}

// deleteDialog asks before a word leaves the bank for good.
func (m *State) deleteDialog(f *flex.Frame) {
	word := m.confirmDelete
	d := f.Dialog(flex.DialogSpec{
		Title:       "把「" + word + "」移出词库？",
		Description: "这个词和它的出处会一起删除。之后在阅读里还能重新收藏。",
		Width:       400,
		Scope:       bankDlg.ID(tree.Name("scope")),
		DismissID:   dlgCancel,
		CloseID:     dlgClose,
	})
	d.Footer(func() {
		f.Button(dlgCancel, "取消").Outline()
		if f.Button(dlgDelete, "删除").Destructive().Clicked() {
			if m.vocab != nil && m.vocab.Remove(word) {
				if strings.EqualFold(m.selected, word) {
					m.selected = ""
				}
				m.saveVocab()
			}
			m.confirmDelete = ""
		}
	})
	if d.Dismissed() {
		m.confirmDelete = ""
	}
}
