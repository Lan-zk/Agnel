package logo

import (
	"bytes"
	_ "embed"
	"image/png"
	"sync"

	"github.com/vivaui/viva/ui"
)

// mark.png is the green mark alone, rasterised from assets/logo-mark.svg on a
// transparent background (the app icon keeps the paper-coloured square). It is a bitmap
// rather than a path because Canvas.DrawPath (and the compiled ui.Icon built
// from the same shapes) produced no window at all on macOS 27 with Viva v0.3.0:
// the painter died and the process stayed alive with nothing on screen.
//
//go:embed mark.png
var markPNG []byte

var (
	once   sync.Once
	bitmap ui.Bitmap
)

// Bitmap returns the mark as premultiplied BGRA, decoded once.
func Bitmap() ui.Bitmap {
	once.Do(func() {
		img, err := png.Decode(bytes.NewReader(markPNG))
		if err != nil {
			return
		}
		b := img.Bounds()
		w, h := b.Dx(), b.Dy()
		pix := make([]byte, w*h*4)
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				r, g, bl, a := img.At(b.Min.X+x, b.Min.Y+y).RGBA()
				// RGBA() returns alpha-premultiplied 16-bit values; the bitmap
				// wants premultiplied BGRA, so a shift is all that is needed.
				i := (y*w + x) * 4
				pix[i+0] = byte(bl >> 8)
				pix[i+1] = byte(g >> 8)
				pix[i+2] = byte(r >> 8)
				pix[i+3] = byte(a >> 8)
			}
		}
		bitmap = ui.Bitmap{Pix: pix, Width: w, Height: h, Stride: w * 4}
	})
	return bitmap
}

// Draw paints the mark into bounds.
func Draw(canvas *ui.Canvas, bounds ui.Rect) {
	if bounds.Width <= 0 || bounds.Height <= 0 {
		return
	}
	bm := Bitmap()
	if bm.Width == 0 {
		return
	}
	canvas.DrawBitmap(bm, bounds)
}
