package desk

import (
	"fmt"
	"math"
	"strings"

	"app/internal/bible"
	"app/internal/dict"
	"app/internal/render"
	"app/internal/theme"

	"github.com/vivaui/viva/flex"
	"github.com/vivaui/viva/ui"
)

// paintSpread draws the pages of the spread: the paper stack, the running
// heads, the folios, the verses, and the shadow in the spine between them.
// dx moves the whole spread sideways while a page is turning.
func (m *State) paintSpread(c *ui.Canvas, left, right ui.Rect, book bible.Book,
	chL, chR bible.Chapter, pg page, dx float32, hits bool) {
	m.paintPage(c, shift(left, dx), book, chL, pg, true, m.left, hits)
	if m.frameMode == Spread {
		m.paintPage(c, shift(right, dx), book, chR, pg, false, m.right, hits)
		m.paintSpine(c, shift(left, dx), shift(right, dx))
	}
}

// paintPage draws one page: the stacked fore-edge, the paper and its shadow,
// the folio in the outer corner, the running head, the verses, and the
// colophon at the foot.
func (m *State) paintPage(c *ui.Canvas, r ui.Rect, book bible.Book, ch bible.Chapter,
	pg page, verso bool, tr string, hits bool) {
	c.Shadow(r, 4, ui.ShadowStyle{OffsetY: 14, Blur: 42, Spread: -20}, theme.PaperShade)
	m.paintPageStack(c, r, verso)
	c.Rounded(r, 3, theme.PaperRule)
	c.Rounded(ui.Rect{X: r.X + 1, Y: r.Y + 1, Width: r.Width - 2, Height: r.Height - 2}, 2.5, theme.Paper)

	colW := m.colL
	if colW == 0 {
		colW = measureMax
	}
	textW := colW - 28
	textX := r.X + (r.Width-textW)/2 + 6

	// The folio, running head, and colophon align flush with the main text block,
	// while verse numbers hang cleanly in the outer left margin.
	folioStr := fmt.Sprint(m.folio(verso))
	if verso || m.frameMode != Spread {
		c.Label(folioStr, ui.Rect{X: textX, Y: r.Y + 15, Width: 32, Height: 16},
			ui.TextStyle{Size: 11, Color: theme.PaperSubtle})
	} else {
		c.LabelAligned(folioStr, ui.Rect{X: textX + textW - 32, Y: r.Y + 15, Width: 32, Height: 16},
			ui.TextStyle{Size: 11, Color: theme.PaperSubtle}, ui.AlignRight)
	}

	head := fmt.Sprintf("%s %d", m.data.Title(tr, book.OSIS), m.chapter)
	c.LabelCenter(head, ui.Rect{X: textX, Y: r.Y + 15, Width: textW, Height: 16},
		ui.TextStyle{Size: 11.5, Color: theme.PaperSoft})
	c.Fill(ui.Rect{X: textX, Y: r.Y + 37, Width: textW, Height: 1}, theme.PaperRule)

	y := r.Y + 58
	body := m.styleFor(tr)
	body.Color = theme.PaperInk
	for i, b := range pg.blocks {
		heads, styles := b.headsL, b.styleL
		if !verso && m.frameMode == Spread {
			heads, styles = b.headsR, b.styleR
		}
		headY := y
		for idx, h := range heads {
			style := styles[idx]
			style.Color = theme.PaperInk
			if style.Italic {
				style.Color = theme.PaperSoft
			}
			h.Paint(c, textX, headY, style)
			headY += h.Height
		}
		y += b.headH
		if b.para && i > 0 {
			y += 7
		}
		c.LabelAligned(fmt.Sprint(b.v), ui.Rect{X: textX - 26, Y: y - 1.5, Width: 18, Height: body.Lead},
			ui.TextStyle{Size: 10.5, Weight: ui.WeightSemibold, Color: theme.VerseNum}, ui.AlignRight)
		if verso || m.frameMode != Spread {
			if hits && m.english(tr) {
				m.hitWords(c, b.left, textX, y, b.v, body)
			}
			b.left.Paint(c, textX, y, body)
		} else {
			b.right.Paint(c, textX, y, body)
		}
		y += max(b.left.Height, b.right.Height) + 6
	}

	m.colophon(c, r, textX, textW, tr, verso)

	if m.bookmarked() && (m.frameMode != Spread || !verso) {
		m.paintRibbon(c, r, verso)
	}
}

// paintPageStack draws the subtle stacked paper sheets and outer board edge
// so the spread reads as an open volume resting on the desk.
func (m *State) paintPageStack(c *ui.Canvas, r ui.Rect, verso bool) {
	if m.frameMode == Spread {
		if verso {
			c.Rounded(ui.Rect{X: r.X - 5, Y: r.Y + 3, Width: r.Width + 5, Height: r.Height + 1}, 4, theme.PaperBoard)
			c.Rounded(ui.Rect{X: r.X - 2.5, Y: r.Y + 1.5, Width: r.Width + 2.5, Height: r.Height + 0.5}, 3.5, theme.PaperStack)
		} else {
			c.Rounded(ui.Rect{X: r.X, Y: r.Y + 3, Width: r.Width + 5, Height: r.Height + 1}, 4, theme.PaperBoard)
			c.Rounded(ui.Rect{X: r.X, Y: r.Y + 1.5, Width: r.Width + 2.5, Height: r.Height + 0.5}, 3.5, theme.PaperStack)
		}
		return
	}
	c.Rounded(ui.Rect{X: r.X - 3, Y: r.Y + 3, Width: r.Width + 6, Height: r.Height + 1}, 4, theme.PaperBoard)
	c.Rounded(ui.Rect{X: r.X - 1.5, Y: r.Y + 1.5, Width: r.Width + 3, Height: r.Height + 0.5}, 3.5, theme.PaperStack)
}

// paintRibbon draws a woven bookmark ribbon hanging from the top margin of a
// bookmarked spread.
func (m *State) paintRibbon(c *ui.Canvas, r ui.Rect, verso bool) {
	rx := r.X + r.Width - 38
	if verso && m.frameMode == Spread {
		rx = r.X + 26
	}
	ry := r.Y - 2
	w := float32(12)
	h := float32(36)
	c.Fill(ui.Rect{X: rx, Y: ry, Width: w, Height: 3}, theme.RibbonDark)
	c.Fill(ui.Rect{X: rx, Y: ry + 3, Width: w, Height: h - 7}, theme.Ribbon)
	// Swallowtail notch at the bottom of the ribbon.
	c.Fill(ui.Rect{X: rx, Y: ry + h - 4, Width: 4, Height: 4}, theme.Ribbon)
	c.Fill(ui.Rect{X: rx + w - 4, Y: ry + h - 4, Width: 4, Height: 4}, theme.Ribbon)
	c.Fill(ui.Rect{X: rx + 4, Y: ry + h - 4, Width: 1.5, Height: 2}, theme.Ribbon)
	c.Fill(ui.Rect{X: rx + w - 5.5, Y: ry + h - 4, Width: 1.5, Height: 2}, theme.Ribbon)
}

// paintSpine is the shadow the pages fall into at the binding: dark at the
// seam, fading out across the inner margin of each page.
func (m *State) paintSpine(c *ui.Canvas, left, right ui.Rect) {
	x := left.X + left.Width
	w := float32(32)
	c.FillLinearGradient(ui.Rect{X: x - w, Y: left.Y + 1, Width: w, Height: left.Height - 2},
		ui.LinearGradient{Start: ui.Point{X: x, Y: 0}, End: ui.Point{X: x - w, Y: 0},
			From: ui.RGBA(24, 26, 22, 0.13), To: ui.RGBA(24, 26, 22, 0)})
	c.FillLinearGradient(ui.Rect{X: x, Y: right.Y + 1, Width: w, Height: right.Height - 2},
		ui.LinearGradient{Start: ui.Point{X: x, Y: 0}, End: ui.Point{X: x + w, Y: 0},
			From: ui.RGBA(24, 26, 22, 0.13), To: ui.RGBA(24, 26, 22, 0)})
	c.Fill(ui.Rect{X: x - 0.5, Y: left.Y + 1, Width: 1, Height: left.Height - 2}, theme.PaperBoard)
}

// paintTurn puts the page-turn controls at the two edges of the book. The
// margin beside the page is the target: the chevron only says where it is.
func (m *State) paintTurn(c *ui.Canvas, desk, left, right ui.Rect) {
	if len(m.pages) == 0 {
		return
	}
	lx, lw := desk.X, left.X-desk.X
	if lw < 24 {
		lx, lw = max(desk.X-56, 0), 56
	}
	rx := right.X + right.Width
	rw := desk.X + desk.Width - rx
	if rw < 24 {
		rx, rw = min(rx, max(desk.X+desk.Width-56, 0)), 56
	}
	cy := desk.Y + desk.Height/2
	c.PushLayer(2)
	m.turnButton(c, ui.Rect{X: lx, Y: desk.Y, Width: lw, Height: desk.Height},
		left.X-28, cy, prevPage, m.turnLabel(-1), "‹", m.canTurn(-1))
	m.turnButton(c, ui.Rect{X: rx, Y: desk.Y, Width: rw, Height: desk.Height},
		right.X+right.Width+28, cy, nextPage, m.turnLabel(1), "›", m.canTurn(1))
	c.PopLayer()
	if c.Clicked(prevPage) {
		m.goSpread(-1)
	}
	if c.Clicked(nextPage) {
		m.goSpread(1)
	}
}

// chevronLift is how far the guillemet's ink rides below the middle of its line
// box, as a fraction of the type size: the box is centred, the ink is not, so
// the mark is lifted by this much to sit in the middle of the disc.
const chevronLift = float32(0.13)

// turnButton is a chevron in the margin: quiet until the pointer finds it.
func (m *State) turnButton(c *ui.Canvas, strip ui.Rect, cx, cy float32, id ui.ID,
	label, glyph string, enabled bool) {
	if strip.Width < 20 || strip.Height < 40 {
		return
	}
	if enabled {
		c.HitWithOptions(id, strip, label,
			ui.HitOptions{Role: ui.RoleButton, Cursor: ui.PointerCursorHand})
	}
	size := float32(21)
	color := theme.InkFaint
	switch {
	case !enabled:
		color = theme.InkFaint.WithAlpha(0.28)
	case c.Hovered() == id:
		size = 23
		color = theme.Accent
		c.Rounded(ui.Rect{X: cx - 18, Y: cy - 18, Width: 36, Height: 36}, 18, theme.Surface)
	}
	c.LabelCenter(glyph, ui.Rect{X: cx - 14, Y: cy - size*(0.78+chevronLift), Width: 28, Height: size * 1.56},
		ui.TextStyle{Size: size, Weight: ui.WeightRegular, Color: color})
}

// paintTurnShadow is the shade the arriving page casts on the one beneath it:
// it falls away from the edge that leads the way.
func (m *State) paintTurnShadow(c *ui.Canvas, left, right ui.Rect, dx float32, dir int, p float32) {
	alpha := 0.2 * float32(math.Sin(float64(p)*math.Pi))
	if alpha < 0.01 {
		return
	}
	w := float32(46)
	if dir > 0 {
		edge := left.X + dx
		c.FillLinearGradient(ui.Rect{X: edge - w, Y: left.Y, Width: w, Height: left.Height},
			ui.LinearGradient{Start: ui.Point{X: edge, Y: 0}, End: ui.Point{X: edge - w, Y: 0},
				From: ui.RGBA(18, 22, 20, alpha), To: ui.RGBA(18, 22, 20, 0)})
		return
	}
	edge := right.X + right.Width + dx
	c.FillLinearGradient(ui.Rect{X: edge, Y: left.Y, Width: w, Height: left.Height},
		ui.LinearGradient{Start: ui.Point{X: edge, Y: 0}, End: ui.Point{X: edge + w, Y: 0},
			From: ui.RGBA(18, 22, 20, alpha), To: ui.RGBA(18, 22, 20, 0)})
}

// paintPosition marks where the reader is: the pages of this chapter, and the
// folio range the two pages carry.
func (m *State) paintPosition(c *ui.Canvas, desk, left, right ui.Rect) {
	text := m.positionText()
	if text == "" {
		return
	}
	w := c.TextWidth(text, 11) + 8
	x := desk.X + (desk.Width-w)/2
	y := desk.Y + desk.Height + 11
	ruleW := float32(22)
	c.Fill(ui.Rect{X: x - ruleW - 10, Y: y + 8, Width: ruleW, Height: 1}, theme.DeskEdge)
	c.LabelCenter(text, ui.Rect{X: x, Y: y, Width: w, Height: 16},
		ui.TextStyle{Size: 11, Color: theme.InkFaint})
	c.Fill(ui.Rect{X: x + w + 10, Y: y + 8, Width: ruleW, Height: 1}, theme.DeskEdge)
}

// positionText is where in the chapter the reader is.
func (m *State) positionText() string {
	if len(m.pages) == 0 {
		return ""
	}
	first, last := m.folioRange()
	if m.frameMode != Spread {
		return fmt.Sprintf("%d / %d 页", first, len(m.pages))
	}
	return fmt.Sprintf("%d–%d / %d 页", first, last, len(m.pages)*2)
}

// folio is the number printed in the corner of a page.
func (m *State) folio(verso bool) int {
	if m.frameMode != Spread {
		return m.spread + 1
	}
	n := m.spread*2 + 1
	if !verso {
		n++
	}
	return n
}

// folioRange is the pair of numbers the pages of this spread carry.
func (m *State) folioRange() (int, int) {
	if m.frameMode != Spread {
		return m.spread + 1, m.spread + 1
	}
	first := m.spread*2 + 1
	return first, first + 1
}

// english reports whether a translation is read word by word, which is what
// the dictionary opens on.
func (m *State) english(id string) bool {
	t, ok := m.data.Translation(id)
	return ok && !t.IsChinese()
}

// colophon is the line at the foot of the page: the translation the page is
// shown in, aligned to the outer edge of the page's text grid.
func (m *State) colophon(c *ui.Canvas, page ui.Rect, colX, colW float32, tr string, verso bool) {
	text := m.colophonText(tr)
	if text == "" {
		return
	}
	footY := page.Y + page.Height - 28
	c.Fill(ui.Rect{X: colX, Y: footY - 6, Width: colW, Height: 1}, theme.PaperRule.WithAlpha(0.55))
	w := c.TextWidth(text, 10.5) + 4
	if verso || m.frameMode != Spread {
		c.Label(text, ui.Rect{X: colX, Y: footY, Width: w, Height: 14},
			ui.TextStyle{Size: 10.5, Color: theme.PaperSubtle})
	} else {
		c.LabelAligned(text, ui.Rect{X: colX + colW - w, Y: footY, Width: w, Height: 14},
			ui.TextStyle{Size: 10.5, Color: theme.PaperSubtle}, ui.AlignRight)
	}
}

// colophonText is the name of a translation on the page.
func (m *State) colophonText(tr string) string {
	t, ok := m.data.Translation(tr)
	if !ok {
		return ""
	}
	if t.IsChinese() {
		return "和合本"
	}
	return t.Abbr
}

func shift(r ui.Rect, dx float32) ui.Rect {
	r.X += dx
	return r
}

// hitWords makes every word of a painted paragraph clickable and draws subtle
// hover and active selection highlights underneath the word's glyphs.
func (m *State) hitWords(c *ui.Canvas, p render.Paragraph, x, y float32, verse int, style render.LineStyle) {
	for _, w := range p.Words(c, x, y, style) {
		word := render.TrimWord(w.Text)
		if word == "" {
			continue
		}
		id := words.At(wordIndex(word, w.Start))
		c.HitWithOptions(id, w.Rect, word, ui.HitOptions{Cursor: ui.PointerCursorHand})
		c.HitLabel(id, w.Rect, word, false)
		active := m.lookup != nil && m.lookup.v == verse &&
			math.Abs(float64(m.lookup.rect.X-w.Rect.X)) < 1 &&
			math.Abs(float64(m.lookup.rect.Y-w.Rect.Y)) < 1
		pad := ui.Rect{
			X:      w.Rect.X - 2,
			Y:      w.Rect.Y + 3,
			Width:  w.Rect.Width + 4,
			Height: w.Rect.Height - 6,
		}
		if active {
			c.Rounded(pad, 3.5, theme.PaperMark)
			c.Fill(ui.Rect{X: pad.X + 1, Y: pad.Y + pad.Height - 1.5, Width: pad.Width - 2, Height: 1.5}, theme.Accent)
		} else if c.Hovered() == id {
			c.Rounded(pad, 3.5, theme.PaperHover)
		}
		if c.Clicked(id) {
			m.openLookup(bible.Verse{N: verse, Text: ""}, word, w.Rect)
		}
	}
}

func wordIndex(word string, start int) int {
	h := 0
	for _, r := range word {
		h = h*31 + int(r)
	}
	h ^= start * 7
	if h < 0 {
		h = -h
	}
	return h % (1 << 24)
}

func (m *State) openLookup(v bible.Verse, word string, rect ui.Rect) {
	ch, _ := m.data.Chapter(m.left, m.osis, m.chapter)
	text := ""
	for _, vv := range ch.Verses {
		if vv.N == v.N {
			text = vv.Text
			break
		}
	}
	zh := ""
	if m.frameMode == Spread {
		if chn, ok := m.data.Chapter(m.right, m.osis, m.chapter); ok {
			for _, vv := range chn.Verses {
				if vv.N == v.N {
					zh = vv.Text
					break
				}
			}
		}
	}
	book, _ := m.data.Book(m.osis)
	m.lookup = &lookup{word: word, osis: m.osis, ch: m.chapter, v: v.N,
		name: m.bookName(book), en: text, zh: zh, rect: rect}
	m.dirty = true
}

func (m *State) lookupCard(f *flex.Frame) {
	lk := m.lookup
	if lk == nil {
		return
	}
	const cardW = 352
	x := lk.rect.X
	if x+cardW > f.Bounds().Width-16 {
		x = f.Bounds().Width - cardW - 16
	}
	if x < 16 {
		x = 16
	}
	y := lk.rect.Y + lk.rect.Height + 8
	if y+330 > f.Bounds().Height-16 {
		y = lk.rect.Y - 8 - 310
	}
	if y < 16 {
		y = 16
	}
	entry, found := dict.Lookup(lk.word)

	m.cardBox = f.Popup().At(x, y).Width(cardW).Pad(18).Gap(8).Radius(10).
		Background(theme.Card).Border(1, theme.CardRule).Shadow(ui.ShadowLG)
	m.cardBox.Children(func() {
		f.Row().Items(flex.Center).Gap(8).Children(func() {
			f.Text(lk.word).FontSize(20).Weight(ui.WeightSemibold).Color(theme.Ink)
			if found && entry.Pos != "" {
				f.Text(entry.Pos).FontSize(theme.SizeMicro).Weight(ui.WeightSemibold).Color(theme.Accent).
					PadSides(2, 7, 2, 7).Radius(4).Background(theme.AccentBg)
			}
			f.Spacer()
			if f.Button(offWord, "关闭").Ghost().XSmall().Clicked() {
				m.lookup = nil
			}
		})
		if found {
			if entry.IPA != "" || entry.Verses > 0 {
				f.Row().Items(flex.Center).Gap(8).Children(func() {
					if entry.IPA != "" {
						f.Text(entry.IPA).FontSize(12.5).Color(theme.Accent)
					}
					f.Spacer()
					if entry.Verses > 0 {
						f.Text(fmt.Sprintf("本书出现 %d 次", entry.Verses)).FontSize(theme.SizeMicro).Color(theme.InkFaint)
					}
				})
			}
			if entry.Lemma != "" {
				f.Row().Items(flex.Center).Children(func() {
					f.Text("这是 "+entry.Lemma+" 的变形").FontSize(11.5).Color(theme.InkSoft).
						PadSides(3, 8, 3, 8).Radius(5).Background(theme.CardInset)
				})
			}
			if len(entry.Zh) > 0 {
				f.Wrap(strings.Join(entry.Zh, "；")).FontSize(14).Weight(ui.WeightSemibold).Color(theme.Ink)
			}
			if len(entry.Senses) > 0 {
				m.cardSection(f, "英文释义", strings.Join(entry.Senses, "; "))
			}
			if len(entry.Collocations) > 0 {
				m.cardSection(f, "本书搭配", strings.Join(entry.Collocations, "；"))
			}
		} else {
			f.Wrap("这部词典还没有收录这个词。").FontSize(theme.SizeSmall).Color(theme.InkSoft)
		}
		f.Column().Fill().PadSides(9, 11, 9, 11).Gap(4).Radius(7).
			Background(theme.CardInset).Children(func() {
			f.Text(fmt.Sprintf("%s %d:%d", lk.name, lk.ch, lk.v)).FontSize(11).Weight(ui.WeightSemibold).Color(theme.Accent)
			f.Wrap(lk.en).FontSize(12).Color(theme.Ink)
			if lk.zh != "" {
				f.Wrap(lk.zh).FontSize(12).Color(theme.InkSoft)
			}
		})
	})
}

func (m *State) cardSection(f *flex.Frame, title, body string) {
	f.Column().Gap(2).Children(func() {
		f.Text(title).FontSize(10.5).Weight(ui.WeightSemibold).Color(theme.InkFaint)
		f.Wrap(body).FontSize(12).Color(theme.InkSoft)
	})
}
