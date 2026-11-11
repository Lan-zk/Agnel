package desk

import (
	"fmt"
	"math"
	"strings"

	"app/internal/bible"
	"app/internal/dict"
	"app/internal/render"
	"app/internal/theme"
	"app/internal/vocab"

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
	// The landing word has had its page. If the page did not contain it, forget
	// it, so it cannot open on an unrelated page later.
	if hits && m.landWord != "" {
		m.landWord = ""
		m.landWordVerse = 0
	}
}

// paintPage draws one page: the stacked fore-edge, the paper and its shadow,
// the folio in the outer corner, the running head, the verses, and the
// colophon at the foot.
func (m *State) paintPage(c *ui.Canvas, r ui.Rect, book bible.Book, ch bible.Chapter,
	pg page, verso bool, tr string, hits bool) {
	if m.warp == nil {
		c.Shadow(r, 4, ui.ShadowStyle{OffsetY: 14, Blur: 42, Spread: -20}, theme.PaperShade)
		m.paintPageStack(c, r, verso)
	}
	m.pRound(c, r, 3, theme.PaperRule)
	m.pRound(c, ui.Rect{X: r.X + 1, Y: r.Y + 1, Width: r.Width - 2, Height: r.Height - 2}, 2.5, theme.Paper)

	colW := m.colL
	if colW == 0 {
		colW = measureMax
	}
	textW := colW - 28
	textX := r.X + (r.Width-textW)/2 + 6

	// The folio, running head, and colophon align flush with the main text block,
	// while verse numbers hang cleanly in the outer left margin.
	folioStr := fmt.Sprint(pg.folio)
	if folioStr == "0" {
		folioStr = fmt.Sprint(m.folio(verso))
	}
	if verso || m.frameMode != Spread {
		m.pLabel(c, folioStr, ui.Rect{X: textX, Y: r.Y + 15, Width: 32, Height: 16},
			ui.TextStyle{Size: 11, Color: theme.PaperSubtle})
	} else {
		m.pLabelAligned(c, folioStr, ui.Rect{X: textX + textW - 32, Y: r.Y + 15, Width: 32, Height: 16},
			ui.TextStyle{Size: 11, Color: theme.PaperSubtle}, ui.AlignRight)
	}

	head := fmt.Sprintf("%s %d", m.data.Title(tr, book.OSIS), m.chapter)
	m.pLabelCenter(c, head, ui.Rect{X: textX, Y: r.Y + 15, Width: textW, Height: 16},
		ui.TextStyle{Size: 11.5, Color: theme.PaperSoft})
	m.pFill(c, ui.Rect{X: textX, Y: r.Y + 37, Width: textW, Height: 1}, theme.PaperRule)

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
			m.pParagraph(c, h, textX, headY, render.LineStyle{Size: style.Size,
				Weight: style.Weight, Italic: style.Italic, Color: style.Color, Lead: style.Lead})
			headY += h.Height
		}
		y += b.headH
		if b.para && i > 0 {
			y += 7
		}
		m.pLabelAligned(c, fmt.Sprint(b.v), ui.Rect{X: textX - 26, Y: y - 1.5, Width: 18, Height: body.Lead},
			ui.TextStyle{Size: 10.5, Weight: ui.WeightSemibold, Color: theme.VerseNum}, ui.AlignRight)
		if verso || m.frameMode != Spread {
			if m.english(tr) {
				m.paintWordMarks(c, b.left, textX, y, b.v, body)
			}
			if hits && m.english(tr) {
				m.hitWords(c, b.left, textX, y, b.v, body)
			}
			m.pParagraph(c, b.left, textX, y, body)
		} else {
			m.pParagraph(c, b.right, textX, y, body)
		}
		y += max(b.left.Height, b.right.Height) + 6
	}

	m.colophon(c, r, textX, textW, tr, verso)

	if m.warp == nil && m.bookmarked() && (m.frameMode != Spread || !verso) {
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

// paintBookTurn draws the rolling sheet in flight. A real page turn is one
// continuous fold: the fold crest travels from the outer fore-edge toward the
// spine, and the sheet is visible on both sides of it at once — the front face
// still carrying the page being left, the rolled-over back face already
// carrying the incoming page. Underneath, the page it lifted from is revealed
// in a widening strip past the crest, and the page it lands on waits until the
// folded flap covers it. In Single mode both faces share one page, so the
// turn folds the sheet over its left binding instead of across a spine.
func (m *State) paintBookTurn(c *ui.Canvas, desk, left, right ui.Rect,
	fromBook bible.Book, fromChL, fromChR bible.Chapter,
	toBook bible.Book, toChL, toChR bible.Chapter,
	fromPg, toPg page, p float32) {
	if fromPg.folio == 0 {
		fromPg.folio = max(m.spread-m.turnDir+1, 1)
	}
	toPg.folio = m.spread + 1

	if m.frameMode != Spread {
		m.paintSingleBookTurn(c, desk, left, fromBook, fromChL, toBook, toChL, fromPg, toPg, p)
		return
	}

	spineX := left.X + left.Width
	c.Clip(desk)

	// The spread underneath: the landing side still shows the old page until
	// the flap covers it, the lifted side already shows the new page through
	// the strip the crest has peeled away.
	forward := m.turnDir > 0
	if forward {
		m.paintPage(c, left, fromBook, fromChL, fromPg, true, m.left, false)
		m.paintPage(c, right, toBook, toChR, toPg, false, m.right, false)
	} else {
		m.paintPage(c, left, toBook, toChL, toPg, true, m.left, false)
		m.paintPage(c, right, fromBook, fromChR, fromPg, false, m.right, false)
	}
	m.paintSpine(c, left, right)

	// The turning sheet: its front face lies on the page being left, its
	// rolled-over back face on the page being carried in.
	front := leafWarp{spineX: spineX, page: right, side: 1, p: p}
	back := leafWarp{spineX: spineX, page: left, side: 1, p: p, back: true}
	frontPage, frontVerso, frontTr := right, false, m.right
	backPage, backVerso, backTr := left, true, m.left
	frontBook, frontCh, frontPg := fromBook, fromChR, fromPg
	backBook, backCh, backPg := toBook, toChL, toPg
	if !forward {
		front.side, back.side = -1, -1
		front.page, back.page = left, right
		frontPage, frontVerso, frontTr = left, true, m.left
		backPage, backVerso, backTr = right, false, m.right
		frontBook, frontCh, frontPg = fromBook, fromChL, fromPg
		backBook, backCh, backPg = toBook, toChR, toPg
	}

	spread := ui.Rect{X: left.X, Y: left.Y,
		Width: right.X + right.Width - left.X, Height: left.Height}

	// The shadows fall on the pages underneath: the curl darkens the newly
	// revealed strip behind the crest, the flap throws a soft shade ahead of
	// its own leading edge.
	front.paintCurlShadows(c, spread)

	// The sheet stands over the book on its own layers: front face first, then
	// the flap's shadow, then the folded flap. The faces need separate layers
	// because text rasterizes above the rectangles of its layer, so only a
	// higher layer lets the flap's paper hide the front page's words.
	c.PushLayer(1)
	m.warp = &front
	m.paintPage(c, frontPage, frontBook, frontCh, frontPg, frontVerso, frontTr, false)
	m.warp = nil
	c.PushLayer(2)
	front.paintFlapShadow(c, spread)
	m.warp = &back
	m.paintPage(c, backPage, backBook, backCh, backPg, backVerso, backTr, false)
	m.warp = nil

	// Crest highlight and cut paper rim on top of the sheet's text.
	c.PushLayer(3)
	_ = front // fore edge off
	// front.paintSheetForeEdge(c)
	c.PopLayer()
	c.PopLayer()
	c.PopLayer()
	c.Unclip()
}

// paintSingleBookTurn rolls one page over its left binding: the front face
// carries the page being left and rolls away toward the spine, the folded
// back face carries the incoming page, and the page underneath is revealed
// past the travelling crest.
func (m *State) paintSingleBookTurn(c *ui.Canvas, desk, pageR ui.Rect,
	fromBook bible.Book, fromCh bible.Chapter,
	toBook bible.Book, toCh bible.Chapter,
	fromPg, toPg page, p float32) {
	forward := m.turnDir > 0
	prog := p
	if !forward {
		// Turning back runs the same roll in reverse: the incoming sheet
		// unrolls from full cover back onto the binding.
		prog = 1 - p
	}

	// Let the arch breathe a little above and below the page, but keep the
	// sheet from spilling over the margins beside it.
	c.Clip(ui.Rect{X: pageR.X, Y: max(pageR.Y-10, 0),
		Width: pageR.Width, Height: pageR.Height + 20})

	m.paintPage(c, pageR, toBook, toCh, toPg, true, m.left, false)

	front := leafWarp{spineX: pageR.X, page: pageR, side: 1, p: prog}
	back := leafWarp{spineX: pageR.X, page: pageR, side: 1, p: prog, back: true}

	front.paintCurlShadows(c, pageR)
	c.PushLayer(1)
	m.warp = &front
	m.paintPage(c, pageR, fromBook, fromCh, fromPg, true, m.left, false)
	m.warp = nil
	c.PushLayer(2)
	front.paintFlapShadow(c, pageR)
	m.warp = &back
	m.paintPage(c, pageR, toBook, toCh, toPg, true, m.left, false)
	m.warp = nil
	c.PushLayer(3)
	_ = front // fore edge off
	// front.paintSheetForeEdge(c)
	c.PopLayer()
	c.PopLayer()
	c.PopLayer()
	c.Unclip()
}

// paintPosition marks where the reader is: the page of this chapter it is on.
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
	// One spread is one page: the two translations face each other on it, so
	// both halves carry the same folio and the count is a count of spreads.
	return fmt.Sprintf("%d / %d 页", m.spread+1, len(m.pages))
}

// folio is the number printed in the corner of a page. Both halves of a spread
// print the same number: the English and Chinese pages are one page.
func (m *State) folio(verso bool) int {
	return m.spread + 1
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
	m.pFill(c, ui.Rect{X: colX, Y: footY - 6, Width: colW, Height: 1}, theme.PaperRule.WithAlpha(0.55))
	w := c.TextWidth(text, 10.5) + 4
	if verso || m.frameMode != Spread {
		m.pLabel(c, text, ui.Rect{X: colX, Y: footY, Width: w, Height: 14},
			ui.TextStyle{Size: 10.5, Color: theme.PaperSubtle})
	} else {
		m.pLabelAligned(c, text, ui.Rect{X: colX + colW - w, Y: footY, Width: w, Height: 14},
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

// wordPad is the wash behind word: drawn inside the line box rather than around
// it, so a mark reads as highlighter on the page.
func wordPad(r ui.Rect) ui.Rect {
	return ui.Rect{
		X:      r.X - 2,
		Y:      r.Y + 3,
		Width:  r.Width + 4,
		Height: r.Height - 6,
	}
}

// activeWord reports whether the lookup card was opened from this word.
func (m *State) activeWord(r ui.Rect, verse int) bool {
	return m.lookup != nil && m.lookup.v == verse &&
		math.Abs(float64(m.lookup.rect.X-r.X)) < 1 &&
		math.Abs(float64(m.lookup.rect.Y-r.Y)) < 1
}

// paintWordMarks lays the washes under a paragraph: the stage of every word
// already in the notebook, and the stronger mark of the word the card is open
// on. It runs whether the page is flat or in flight, so a saved word keeps its
// mark while the sheet is being turned.
func (m *State) paintWordMarks(c *ui.Canvas, p render.Paragraph, x, y float32, verse int, style render.LineStyle) {
	for _, w := range p.Words(c, x, y, style) {
		word := render.TrimWord(w.Text)
		if word == "" {
			continue
		}
		pad := wordPad(w.Rect)
		if m.activeWord(w.Rect, verse) {
			m.fillPad(c, pad, theme.PaperMark)
			m.fillPad(c, ui.Rect{X: pad.X + 1, Y: pad.Y + pad.Height - 1.5, Width: pad.Width - 2, Height: 1.5},
				theme.Accent)
			continue
		}
		stage, ok := m.savedStage(word)
		if !ok {
			continue
		}
		switch stage {
		case vocab.StageNew:
			m.fillPad(c, pad, theme.PaperMarkNew)
		case vocab.StageLearning:
			m.fillPad(c, pad, theme.PaperMarkLearning)
		}
	}
}

// fillPad paints one wash, flat on a still page and carried onto the turning
// sheet's own surface when a sheet is in flight: the wash takes the quad's
// light so it darkens into the curl with the paper under it.
func (m *State) fillPad(c *ui.Canvas, box ui.Rect, color ui.Color) {
	if m.warp == nil {
		c.Rounded(box, 3.5, color)
		return
	}
	quad, ok := m.warp.Quad(box)
	if !ok {
		return
	}
	for i, tint := range quad.Tint {
		quad.Tint[i] = ui.Color{
			Red:   color.Red * tint.Red,
			Green: color.Green * tint.Green,
			Blue:  color.Blue * tint.Blue,
			Alpha: color.Alpha * tint.Alpha,
		}
	}
	c.FillQuad(quad)
}

// hitWords makes every word of a painted paragraph clickable and draws the
// hover and press feedback underneath the word's glyphs. The stage marks are
// already down (paintWordMarks), so nothing here paints over a saved word.
func (m *State) hitWords(c *ui.Canvas, p render.Paragraph, x, y float32, verse int, style render.LineStyle) {
	for _, w := range p.Words(c, x, y, style) {
		word := render.TrimWord(w.Text)
		if word == "" {
			continue
		}
		id := words.At(wordIndex(word, w.Start))
		c.HitWithOptions(id, w.Rect, word, ui.HitOptions{Cursor: ui.PointerCursorHand})
		c.HitLabel(id, w.Rect, word, false)
		// The page just landed on the verse the bank sent it to: open the card
		// for the word it was opened from, in place.
		if m.landWord != "" && m.landWordVerse == verse && strings.EqualFold(word, m.landWord) {
			m.openLookup(bible.Verse{N: verse, Text: ""}, word, w.Rect)
			m.landWord = ""
		}
		active := m.activeWord(w.Rect, verse)
		if !active && c.Hovered() == id {
			m.fillPad(c, wordPad(w.Rect), theme.PaperHover)
		}
		if c.Clicked(id) {
			m.openLookup(bible.Verse{N: verse, Text: ""}, word, w.Rect)
		}
	}
}

// savedStage is the stage of a word kept in the bank, if it is kept at all. A
// mastered word returns known too, and the page leaves it unmarked.
func (m *State) savedStage(word string) (string, bool) {
	if m.vocabIndex == nil {
		return "", false
	}
	w, ok := m.vocabIndex[strings.ToLower(strings.TrimSpace(word))]
	if !ok {
		return "", false
	}
	return w.Stage, true
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
	// In single-column mode the other page is not painted, but its verse still
	// belongs in the card, so a saved word keeps both languages.
	if m.right != m.left {
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
	const cardW = 368
	x := lk.rect.X - 8
	if x+cardW > f.Bounds().Width-16 {
		x = f.Bounds().Width - cardW - 16
	}
	if x < 16 {
		x = 16
	}
	y := lk.rect.Y + lk.rect.Height + 10
	// The card runs tall when a verse is long; if the room below the word does
	// not hold it, it opens upward from the word instead.
	if y+480 > f.Bounds().Height-16 {
		y = lk.rect.Y - 10 - 480
	}
	// Keep it out of the 54px top bar; the popup is placed in window coordinates.
	if y < 66 {
		y = 66
	}
	entry, found := dict.Lookup(lk.word)
	saved := m.vocab != nil && m.vocab.Known(lk.word)

	m.cardBox = f.Popup().At(x, y).Width(cardW).PadSides(18, 20, 18, 20).Gap(12).Radius(11).
		Background(theme.Card).Border(1, theme.CardRule).Shadow(ui.ShadowLG)
	m.cardBox.Children(func() {
		// Headword row: the word, what it is, where it already sits in the
		// collection, and a quiet way out.
		f.Row().Items(flex.Center).Gap(7).Children(func() {
			f.Text(lk.word).FontSize(22).Weight(ui.WeightSemibold).Color(theme.PaperInk)
			if found && entry.Pos != "" {
				f.Text(entry.Pos).FontSize(11).Color(theme.PaperSoft)
			}
			if stage, ok := m.savedStage(lk.word); ok {
				stageDot(f, stageFilter(stage), 6)
				f.Text(stageLabel(stage)).FontSize(10.5).Color(theme.PaperSoft)
			}
			f.Spacer()
			closeBtn := f.Row().Items(flex.Center).PadSides(3, 7, 3, 7).Radius(5).
				HoverBackground(theme.CardInset)
			closeBtn.Children(func() {
				f.Text("关闭").FontSize(11).Color(theme.PaperSoft)
			})
			if closeBtn.Hit(offWord, "关闭").Clicked() {
				m.lookup = nil
			}
		})

		if found {
			// Pronunciation, the form it came from, and how often the book says it.
			ipa := formatIPA(entry.IPA)
			lemma := entry.Lemma != "" && !strings.EqualFold(entry.Lemma, lk.word)
			if ipa != "" || lemma || entry.Verses > 0 {
				f.Row().Items(flex.Center).Gap(8).Children(func() {
					if ipa != "" {
						f.Text(ipa).FontSize(12.5).Color(theme.PaperSoft)
					}
					if lemma {
						lem := f.Row().Items(flex.Center).Gap(3).PadSides(2, 7, 2, 7).Radius(4).
							Background(theme.CardInset).HoverBackground(theme.PaperStack)
						lem.Children(func() {
							f.Text("原形").FontSize(10.5).Color(theme.PaperSoft)
							f.Text(entry.Lemma).FontSize(11.5).Weight(ui.WeightSemibold).Color(theme.Accent)
						})
						if lem.Hit(cardLemma, "这是 "+entry.Lemma+" 的变形").Clicked() {
							lk.word = entry.Lemma
							m.dirty = true
						}
					}
					f.Spacer()
					if entry.Verses > 0 {
						f.Text("本书出现 " + thousands(entry.Verses) + " 次").FontSize(10.5).Color(theme.PaperSoft)
					}
				})
			}

			f.Column().FillWidth().Height(1).Background(theme.PaperRule)

			// The Chinese sense is the line the reader came for.
			if len(entry.Zh) > 0 {
				f.Wrap(strings.Join(entry.Zh, "；")).FontSize(17).LineHeight(26).
					Weight(ui.WeightMedium).Color(theme.PaperInk)
			}
			if len(entry.Senses) > 0 {
				m.cardSection(f, "英文释义", strings.Join(entry.Senses, "; "))
			}
			if len(entry.Collocations) > 0 {
				m.cardSection(f, "本书搭配", strings.Join(entry.Collocations, " · "))
			}
		} else {
			f.Column().FillWidth().Height(1).Background(theme.PaperRule)
			f.Wrap("这部词典还没有收录这个词。").FontSize(theme.SizeSmall).LineHeight(19).Color(theme.PaperSoft)
		}

		// Where the word was met: the verse, in both languages.
		f.Column().FillWidth().Radius(8).Background(theme.CardInset).
			PadSides(11, 13, 11, 12).Gap(5).Children(func() {
			f.Row().Items(flex.Center).Gap(6).Children(func() {
				f.Text(fmt.Sprintf("%s %d:%d", lk.name, lk.ch, lk.v)).
					FontSize(11).Weight(ui.WeightSemibold).Color(theme.PaperSoft)
				f.Text(m.left).FontSize(10).Color(theme.PaperSoft)
				f.Spacer()
			})
			f.Wrap(lk.en).FontSize(13).LineHeight(19).Color(theme.PaperInk)
			if lk.zh != "" {
				f.Wrap(lk.zh).FontSize(12).LineHeight(18).Color(theme.PaperSoft)
			}
		})

		// One primary action; once taken, the same button turns quiet and a way
		// into the collection appears beside it.
		label := "收藏这个词"
		if saved {
			label = "已收藏 · 移出"
		}
		f.Row().FillWidth().Gap(8).Items(flex.Center).Children(func() {
			save := f.Row().Height(32).Grow(1).Justify(flex.Center).Items(flex.Center).Radius(7)
			if saved {
				save = save.Background(theme.CardInset).Border(1, theme.CardRule).
					HoverBackground(theme.PaperStack)
			} else {
				save = save.Background(theme.Accent)
			}
			save.Children(func() {
				t := f.Text(label).FontSize(12).Weight(ui.WeightSemibold)
				if saved {
					t.Color(theme.PaperSoft)
				} else {
					t.Color(theme.OnAccent)
				}
			})
			if save.Hit(cardSave, label).Clicked() {
				m.toggleLookup()
			}
			if saved {
				jump := f.Row().Height(32).PadSides(0, 11, 0, 11).Items(flex.Center).Justify(flex.Center).
					Radius(7).Background(theme.CardInset).HoverBackground(theme.PaperStack)
				jump.Children(func() {
					f.Text("在词库查看 ›").FontSize(11.5).Weight(ui.WeightMedium).Color(theme.PaperSoft)
				})
				if jump.Hit(cardBank, "在词库查看").Clicked() {
					word := lk.word
					m.goPane(paneBank)
					m.dictMode = false
					m.selected = word
					m.syncPageToSelected()
				}
			}
		})
	})
}

func (m *State) cardSection(f *flex.Frame, title, body string) {
	f.Column().Gap(4).Children(func() {
		f.Text(title).FontSize(10.5).Weight(ui.WeightSemibold).Color(theme.PaperSoft)
		f.Wrap(body).FontSize(13).LineHeight(20).Color(theme.PaperSoft)
	})
}
