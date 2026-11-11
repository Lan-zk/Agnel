// Package render turns text into rows of words that can be painted and clicked.
//
// flex lays out blocks of text; the reader needs more than that: every English
// word inside a verse is a target, and verses of different lengths have to fit
// one column before anything is painted. So this package wraps paragraphs into
// rows with the canvas' own engine and keeps the byte offsets of each word, so
// the same numbers place the glyphs and the click region under them.
package render

import (
	"strings"

	"github.com/vivaui/viva/ui"
)

// LineStyle is one paragraph's type: size, weight, colour and the distance from
// one baseline to the next.
type LineStyle struct {
	Size   float32
	Weight float32
	Italic bool
	Color  ui.Color
	Lead   float32
}

// Paint style for this line.
func (s LineStyle) text() ui.TextStyle {
	return ui.TextStyle{Size: s.Size, Weight: s.Weight, Italic: s.Italic, Color: s.Color}
}

// Row is one wrapped line.
type Row struct {
	Text   string  // the bytes of the line, trailing spaces included
	Start  int     // byte offset of the first byte in the paragraph
	Offset float32 // x of the first glyph from the paragraph's left edge
	Width  float32 // advance of the line
}

// Paragraph is a wrapped block of text.
type Paragraph struct {
	Rows   []Row
	Height float32
}

// Layout wraps text at width.
func Layout(c *ui.Canvas, text string, s LineStyle, width float32) Paragraph {
	if text == "" {
		return Paragraph{}
	}
	var buf []ui.TextRow
	buf = c.WrapRows(buf, text, s.Size, s.Weight, width)
	p := Paragraph{Rows: make([]Row, len(buf))}
	for i, r := range buf {
		line := text[r.Start:r.End]
		p.Rows[i] = Row{
			Text:  line,
			Start: r.Start,
			Width: c.TextWidthWeight(line, s.Size, s.Weight),
		}
	}
	p.Height = float32(len(p.Rows)) * s.Lead
	return p
}

// LastRowWidth is the advance of the paragraph's last row. The printed Bible
// sets the verse number just past it.
func (p Paragraph) LastRowWidth() float32 {
	if len(p.Rows) == 0 {
		return 0
	}
	return p.Rows[len(p.Rows)-1].Width
}

// Paint draws the paragraph with its top-left corner at x, y.
func (p Paragraph) Paint(c *ui.Canvas, x, y float32, s LineStyle) {
	p.PaintWarped(c, x, y, s, nil)
}

// Warp places a box laid out on a page into the shape that page takes on
// screen. The reader gives it the lifted leaf of a page turn: its Quad maps the
// row's box to a quadrilateral, and reports false when the box has folded away
// and should not be drawn at all.
type Warp interface {
	Quad(box ui.Rect) (ui.Quad, bool)
}

// PaintWarped paints the paragraph through warp. With a nil warp it is Paint:
// the platform still rasterises each row itself, warp only moves and shades it,
// so a line drawn this way stays as crisp as one drawn flat.
func (p Paragraph) PaintWarped(c *ui.Canvas, x, y float32, s LineStyle, warp Warp) {
	style := s.text()
	for i, r := range p.Rows {
		if r.Text == "" {
			continue
		}
		box := ui.Rect{
			X:      x + r.Offset,
			Y:      y + float32(i)*s.Lead,
			Width:  r.Width + 2,
			Height: s.Lead,
		}
		if warp == nil {
			c.Label(r.Text, box, style)
			continue
		}
		quad, ok := warp.Quad(box)

		if !ok {
			continue
		}
		c.LabelQuad(r.Text, box, quad, style)
	}
}

// Word is one clickable run of text, positioned in window pixels.
type Word struct {
	Text  string
	Start int // byte offset of the word inside the paragraph
	Rect  ui.Rect
}

// Words returns one entry per word of the paragraph, in reading order, laid out
// with the paragraph's top-left corner at x, y.
func (p Paragraph) Words(c *ui.Canvas, x, y float32, s LineStyle) []Word {
	var out []Word
	for i, r := range p.Rows {
		top := y + float32(i)*s.Lead
		for _, span := range wordSpans(r.Text) {
			x0 := c.TextCaret(r.Text, s.Size, s.Weight, span[0])
			x1 := c.TextCaret(r.Text, s.Size, s.Weight, span[1])
			if x1 <= x0 {
				continue
			}
			out = append(out, Word{
				Text:  r.Text[span[0]:span[1]],
				Start: r.Start + span[0],
				Rect: ui.Rect{
					X:      x + r.Offset + x0,
					Y:      top,
					Width:  x1 - x0,
					Height: s.Lead,
				},
			})
		}
	}
	return out
}

// wordSpans splits a line into [start, end) byte pairs, one per word. Spaces
// separate words; punctuation and hyphens stay with the word they touch, which
// is what a reader means by "this word".
func wordSpans(line string) [][2]int {
	var out [][2]int
	start := -1
	for i := 0; i < len(line); i++ {
		ch := line[i]
		if ch == ' ' || ch == '\t' || ch == '\n' {
			if start >= 0 {
				out = append(out, [2]int{start, i})
				start = -1
			}
			continue
		}
		if start < 0 {
			start = i
		}
	}
	if start >= 0 {
		out = append(out, [2]int{start, len(line)})
	}
	return out
}

// TrimWord returns a word without the punctuation a lookup should ignore.
func TrimWord(w string) string {
	return strings.Trim(w, ".,;:!?()[]{}\"'“”‘’—-")
}
