package desk

import (
	"fmt"

	"app/internal/bible"
	"app/internal/logo"
	"app/internal/theme"

	"github.com/vivaui/viva/flex"
	"github.com/vivaui/viva/ui"
)

// The marks that stand in for icons are set in the page's own face — the
// guillemets are already the angles a chevron is made of, and a stroked icon
// path never reaches the screen on this renderer (see docs/viva-framework-issues.md).
// A filled contour (Stroke: 0) does not reach it either — the whole triangle
// pipeline is dark there, so the caret is a glyph too.

// barPicker is the picker's own top strip. Returning sits in the corner the
// reader came in through, so leaving is the same short step as arriving.
func (m *State) barPicker(f *flex.Frame) {
	f.Row().Height(54).PadX(24).Items(flex.Center).Gap(12).Children(func() {
		back := f.Row().Items(flex.Center).Gap(5).PadSides(7, 12, 6, 9).Radius(7).
			Background(theme.Surface).HoverBackground(theme.Surface2)
		back.Children(func() {
			f.Text("‹").FontSize(16).Color(theme.InkSoft)
			f.Text("返回阅读").FontSize(theme.SizeSmall).Weight(ui.WeightMedium).Color(theme.Ink)
		})
		if back.Hit(doClose, "返回阅读").Clicked() {
			m.picker = false
			m.dirty = true
		}
		m.barRule(f)
		f.Image(logo.Bitmap()).Size(24, 24)
		f.Text("选择经文").FontSize(18.5).Weight(ui.WeightSemibold).Color(theme.Ink)
		f.Text(m.pickerScope()).FontSize(theme.SizeMicro).Color(theme.InkFaint).
			PadSides(3, 8, 3, 8).Radius(5).Background(theme.Surface2)
		f.Spacer()
		f.Row().Items(flex.Center).Gap(2).Pad(3).Radius(8).Background(theme.DeskTrack).Children(func() {
			for i, label := range []string{"中文", "English"} {
				chip := f.Text(label).FontSize(theme.SizeSmall).PadSides(5, 12, 5, 12).Radius(6)
				if m.lang == i {
					chip.Background(theme.Card).Color(theme.Accent).Weight(ui.WeightSemibold)
				} else {
					chip.Color(theme.InkSoft).HoverBackground(theme.Surface2)
				}
				if chip.Hit(lang.At(i), label).Clicked() && m.lang != i {
					m.lang = i
					m.remember()
					m.dirty = true
				}
			}
		})
	})
}

// pickerScope is the line beside the picker's title: what is being chosen
// right now — the whole canon, or the chapters of the open book.
func (m *State) pickerScope() string {
	if m.pickerBook != "" {
		book, _ := m.data.Book(m.pickerBook)
		return fmt.Sprintf("%s · %d 章", m.bookName(book), book.Chapters)
	}
	return fmt.Sprintf("%d 卷 · %d 章 · %d 节", len(m.data.Books()),
		m.chapterCount(), m.verseTotal())
}

func (m *State) chapterCount() int {
	n := 0
	for _, b := range m.data.Books() {
		n += b.Chapters
	}
	return n
}

func (m *State) verseTotal() int {
	n := 0
	for _, t := range m.data.Translations() {
		if t.ID == m.left {
			n = t.Verses
		}
	}
	return n
}

// pickerPage is the whole-book browser: the canon in two columns, grouped the
// way a printed Bible groups it. A book opens in place to show its chapters,
// so choosing a passage never takes the reader off this page.
func (m *State) pickerPage(f *flex.Frame) {
	f.Column().Fill().Children(func() {
		m.barPicker(f)
		f.Divider()
		m.bookPage(f)
	})
}

func (m *State) bookPage(f *flex.Frame) {
	sections := m.data.Sections()
	var ot, nt []bible.Section
	for _, s := range sections {
		if len(s.Books) > 0 && s.Books[0].Testament == "NT" {
			nt = append(nt, s)
		} else {
			ot = append(ot, s)
		}
	}
	f.ScrollArea(bookGrid).Grow(1).PadSides(24, 40, 28, 40).Children(func() {
		if len(m.recent) > 0 || len(m.bookmarks) > 0 {
			f.Row().Justify(flex.Center).Children(func() {
				f.Column().Width(952).Gap(8).PadSides(10, 14, 10, 14).Radius(9).
					Background(theme.Surface).Border(1, theme.Rule).Children(func() {
					if len(m.recent) > 0 {
						f.Row().Gap(6).Items(flex.Center).Children(func() {
							f.Text("最近阅读").FontSize(theme.SizeMicro).Weight(ui.WeightSemibold).Color(theme.InkFaint).PadSides(0, 6, 0, 0)
							for i, key := range m.recent {
								var osis string
								var ch int
								fmt.Sscanf(key, "%[^:]:%d", &osis, &ch)
								book, ok := m.data.Book(osis)
								if !ok {
									continue
								}
								label := fmt.Sprintf("%s %d", m.bookName(book), ch)
								chip := f.Text(label).FontSize(theme.SizeSmall).Color(theme.Ink).
									PadSides(4, 10, 4, 10).Radius(6).
									Background(theme.Card).HoverBackground(theme.Surface2)
								if chip.Hit(recentRow.At(i), label).Clicked() {
									m.choose(osis, ch)
								}
							}
						})
					}
					if len(m.bookmarks) > 0 {
						f.Row().Gap(6).Items(flex.Center).Children(func() {
							f.Text("书签夹").FontSize(theme.SizeMicro).Weight(ui.WeightSemibold).Color(theme.Ribbon).PadSides(0, 6, 0, 0)
							for i, b := range m.bookmarks {
								book, ok := m.data.Book(b.Book)
								if !ok {
									continue
								}
								label := fmt.Sprintf("%s %d", m.bookName(book), b.Chapter)
								chip := f.Text(label).FontSize(theme.SizeSmall).Weight(ui.WeightMedium).Color(theme.Accent).
									PadSides(4, 10, 4, 10).Radius(6).
									Background(theme.AccentBg).HoverBackground(theme.Surface2)
								if chip.Hit(marks.At(i), label).Clicked() {
									m.jumpTo(b)
								}
							}
						})
					}
				})
			})
			f.Spacer().Height(16)
		}
		f.Row().Gap(48).Justify(flex.Center).Items(flex.Start).Children(func() {
			otName, otCount := "旧约", "39 卷"
			ntName, ntCount := "新约", "27 卷"
			if m.lang == 1 {
				otName, otCount = "Old Testament", "39 Books"
				ntName, ntCount = "New Testament", "27 Books"
			}
			m.testament(f, otName, otCount, ot)
			m.testament(f, ntName, ntCount, nt)
		})
	})
}

var sectionEN = map[string]string{
	"律法书":   "Pentateuch",
	"历史书":   "Historical Books",
	"诗歌智慧书": "Poetry & Wisdom",
	"大先知书":  "Major Prophets",
	"小先知书":  "Minor Prophets",
	"福音书":   "Gospels",
	"使徒行传":  "Acts",
	"保罗书信":  "Pauline Epistles",
	"普通书信":  "General Epistles",
	"启示录":   "Apocalypse",
}

// testament is one half of the canon: its sections, and the books in them set
// as a structured editorial index.
func (m *State) testament(f *flex.Frame, name, count string, sections []bible.Section) {
	f.Column().Width(452).Gap(16).Children(func() {
		f.Row().Items(flex.Center).Gap(8).PadSides(0, 4, 4, 4).Children(func() {
			f.Text(name).FontSize(16.5).Weight(ui.WeightSemibold).Color(theme.Ink)
			f.Text(count).FontSize(theme.SizeMicro).Color(theme.InkFaint).
				PadSides(2, 7, 2, 7).Radius(4).Background(theme.Surface2)
		})
		f.Divider()
		for _, s := range sections {
			f.Column().Gap(6).Children(func() {
				secTitle := s.ZH
				secCount := fmt.Sprintf("%d 卷", len(s.Books))
				if m.lang == 1 {
					if en, ok := sectionEN[s.ZH]; ok {
						secTitle = en
					}
					secCount = fmt.Sprintf("%d", len(s.Books))
				}
				f.Row().Items(flex.Center).Gap(6).PadSides(2, 4, 0, 4).Children(func() {
					f.Text(secTitle).FontSize(12).Weight(ui.WeightSemibold).Color(theme.InkSoft)
					f.Spacer()
					f.Text(secCount).FontSize(10.5).Color(theme.InkFaint)
				})
				for i := 0; i < len(s.Books); i += 2 {
					b0 := s.Books[i]
					hasSecond := i+1 < len(s.Books)
					var b1 bible.Book
					if hasSecond {
						b1 = s.Books[i+1]
					}
					f.Row().Gap(8).Children(func() {
						m.bookTile(f, b0)
						if hasSecond {
							m.bookTile(f, b1)
						} else {
							f.Column().Width(222)
						}
					})
					if m.pickerBook == b0.OSIS {
						m.chapterStrip(f, b0)
					} else if hasSecond && m.pickerBook == b1.OSIS {
						m.chapterStrip(f, b1)
					}
				}
			})
		}
	})
}

// bookTile is one book in the index: its short mark, its name, and how many
// chapters it holds.
func (m *State) bookTile(f *flex.Frame, b bible.Book) {
	isCurrent := b.OSIS == m.osis
	isOpen := m.pickerBook == b.OSIS
	row := f.Row().Width(222).Height(34).PadSides(5, 10, 5, 8).Radius(7).Items(flex.Center).Gap(7)
	switch {
	case isOpen:
		row.Background(theme.Card).Border(1, theme.Accent)
	case isCurrent:
		row.Background(theme.AccentBg).HoverBackground(theme.Surface)
	default:
		row.Background(theme.Surface).HoverBackground(theme.Surface2)
	}
	row.Children(func() {
		badgeBg := theme.Surface2
		badgeFg := theme.InkSoft
		if isCurrent || isOpen {
			badgeBg = theme.AccentBg
			badgeFg = theme.Accent
		}
		f.Column().Width(30).Height(22).Radius(4).Items(flex.Center).Justify(flex.Center).
			Background(badgeBg).Children(func() {
			f.Text(m.mark(b)).FontSize(10.5).Weight(ui.WeightSemibold).Color(badgeFg)
		})
		name := f.Text(m.bookName(b)).FontSize(theme.SizeBody).Color(theme.Ink).Ellipsis()
		if isCurrent || isOpen {
			name.Color(theme.Accent).Weight(ui.WeightSemibold)
		}
		f.Spacer()
		f.Text(fmt.Sprintf("%d 章", b.Chapters)).FontSize(11).Color(theme.InkFaint)
	})
	row.Hit(books.At(m.bookIndex(b.OSIS)), m.bookName(b))
	if row.Clicked() {
		if b.Chapters <= 1 {
			m.choose(b.OSIS, 1)
		} else {
			m.toggleBook(b.OSIS)
		}
	}
}

// bookIndex is a book's place in the canon, which is the identity of its tile.
func (m *State) bookIndex(osis string) int {
	for i, b := range m.data.Books() {
		if b.OSIS == osis {
			return i
		}
	}
	return 0
}

// toggleBook opens a book's chapters in the index, or closes them again.
func (m *State) toggleBook(osis string) {
	if m.pickerBook == osis {
		m.pickerBook = ""
	} else {
		m.pickerBook = osis
	}
	m.dirty = true
}

const (
	chapCellW = 38
	chapCellH = 28
	chapGap   = 4
)

// chapterStrip is the run of chapters under an open book, spanning the full
// width of the testament column so even 150-chapter Psalms stays legible.
func (m *State) chapterStrip(f *flex.Frame, b bible.Book) {
	f.Column().Width(452).PadSides(10, 12, 12, 12).Gap(8).Radius(8).
		Background(theme.Card).Border(1, theme.CardRule).Children(func() {
		f.Row().Items(flex.Center).Gap(6).Children(func() {
			f.Text(m.bookName(b)).FontSize(12.5).Weight(ui.WeightSemibold).Color(theme.Accent)
			f.Text(fmt.Sprintf("· 选择章次（共 %d 章）", b.Chapters)).FontSize(11).Color(theme.InkFaint)
		})
		f.Grid(chapCellW).Gap(chapGap).Children(func() {
			for c := 1; c <= b.Chapters; c++ {
				m.chapterCell(f, b, c)
			}
		})
	})
}

// chapterCell is one chapter of an open book.
func (m *State) chapterCell(f *flex.Frame, book bible.Book, c int) {
	label := fmt.Sprintf("%s %d", m.bookName(book), c)
	current := book.OSIS == m.osis && c == m.chapter
	box := f.Column().Height(chapCellH).Items(flex.Center).Justify(flex.Center).Radius(5).
		Background(theme.CardInset).HoverBackground(theme.Surface2)
	if current {
		box.Background(theme.Accent)
	}
	box.Hit(chaps.At(c), label)
	box.Children(func() {
		num := f.Text(fmt.Sprint(c)).FontSize(12).Color(theme.Ink)
		if current {
			num.Color(theme.OnAccent).Weight(ui.WeightSemibold)
		}
	})
	if box.Clicked() {
		m.choose(book.OSIS, c)
	}
}

func (m *State) choose(osis string, chapter int) {
	m.osis = osis
	m.picker = false
	m.pickerBook = ""
	m.openChapter(chapter)
}
