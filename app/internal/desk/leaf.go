package desk

import (
	"math"

	"app/internal/render"
	"app/internal/theme"

	"github.com/vivaui/viva/ui"
)

// The sheet's numbers live in turnLook (turnstyle.go) so the tuning panel can
// move them while the reader watches.

// leafWarp models one face (either the unpeeled front face or the rolled-over
// back face) of a physical sheet of paper rolling across the book spread.
//
// Unlike a stiff hinged board (which swings around the spine, vanishes into a
// line at 90°, and flips to its back only in the second half), a real book page
// turns as a travelling cylindrical wave:
//   - The fold crest sweeps continuously from the outer fore-edge to the spine.
//   - On the starting side of the fold crest, the front face (recto) of the
//     turning sheet remains visible and curves upward into the rolling cylinder.
//   - On the leading side of the fold crest, the rolled-over back face (verso,
//     carrying the incoming page) is simultaneously visible from the very start
//     of the turn, sweeping across the spine and settling flat onto the
//     opposite page.
type leafWarp struct {
	spineX float32
	page   ui.Rect
	side   float32 // +1 when turning forward (right-to-left), -1 when turning backward (left-to-right)
	p      float32 // eased turn progress in 0..1
	back   bool    // false = unpeeled front face approaching the fold; true = rolled-over back face leading the fold
}

// cylinder geometry for the current progress p:
//   - uFold: material distance from the spine to the rolling fold crest (W -> 0)
//   - radius: radius of the rolling paper cylinder (peaks mid-turn, 0 at both ends)
//   - lift: peak Z height of the cylinder crest toward the viewer
func (w *leafWarp) rollGeometry() (uFold, radius, lift float32) {
	width := max(w.page.Width, 1)
	p := clamp01(w.p)
	sinP := float32(math.Sin(float64(p) * math.Pi))
	// The crest sweeps the full page width; the cylinder it winds on swells
	// through the middle of the turn and vanishes as the sheet settles flat.
	uFold = width * (1 - p)
	radius = width * turnRadius * sinP
	lift = turnLift * sinP
	return uFold, radius, lift
}

// project maps a material coordinate (x, y) on w.page to its 3D screen point
// and shading factor. It returns ok = false when the coordinate lies on the
// other side of the fold crest for this face.
func (w *leafWarp) project(x, y float32) (pt ui.Point, shade float32, ok bool) {
	width := max(w.page.Width, 1)
	u := abs(x - w.spineX)
	uFold, radius, lift := w.rollGeometry()
	sinP := float32(math.Sin(float64(clamp01(w.p)) * math.Pi))

	const eps = 1.5
	var outx, depth float32

	if !w.back {
		// Front face (the page being lifted): only the material between the
		// spine (u = 0) and the fold crest (u = uFold) has not yet rolled over
		// the crest.
		if u > uFold+eps {
			return ui.Point{}, 0, false
		}
		d := max(uFold-u, 0) // distance ahead of the fold crest
		bendBand := max(radius*turnBend, 12)
		if d >= bendBand {
			// Flat on the starting page, with a subtle spine gutter curve and
			// the soft shade the hovering flap throws across its landing line.
			outx = u
			depth = 0
			spineCrease := 1 - clamp01(u/(width*0.18))
			shade = 1 - 0.10*spineCrease*spineCrease*sinP
		} else {
			// Approaching the rolling cylinder crest: curve smoothly upward (Z > 0)
			// and foreshorten slightly in X as the tangent turns vertical at the crest.
			t := 1 - d/bendBand // 0 at bendBand -> 1 at the fold crest
			ang := t * (math.Pi * 0.5)
			sinA := float32(math.Sin(float64(ang)))
			cosA := float32(math.Cos(float64(ang)))
			outx = (uFold - bendBand) + bendBand*(0.32*t+0.68*sinA)
			depth = lift * (1 - cosA) * 0.65
			// Specular highlight just before the crest, rolling into soft shade at the very edge.
			crest := float32(math.Sin(float64(t) * math.Pi))
			shade = 1 + 0.045*crest*sinP - 0.10*t*t*sinP
		}
	} else {
		// Rolled-over back face (the incoming page on the reverse side of the sheet):
		// On the target page, a point has distance u from the spine (0 at the spine,
		// W at the outer margin). When the sheet is folded at uFold from the spine,
		// the outer fore-edge (material distance W) has rolled over the crest to
		// screen position uFold - (W - uFold) = 2*uFold - W on the starting side!
		// Wait: on the back face (target page), the outer fore-edge of the sheet is
		// at u = W (the outer margin of the target page), and the fold crest is at
		// u = uFold! So the material that has already crossed the fold crest is
		// the region u in [uFold, W]!
		if u < uFold-eps {
			return ui.Point{}, 0, false
		}
		d := max(u-uFold, 0) // arc distance from the fold crest toward the free fore-edge
		curlBand := max(radius*turnCurl, 16)

		// In flat unfolded coordinates, the fold crest sits at +w.side * uFold
		// from the spine, and the rolled-over material extends in the -w.side
		// direction (toward the spine and across onto the target page!).
		// Specifically, when p = 1, uFold = 0, and a point at distance u on the
		// target page lands at -w.side * u from the spine, which is its exact
		// flat coordinate on the target page!
		if d < curlBand {
			t := d / curlBand // 0 at the fold crest -> 1 where the sheet flattens out
			ang := t * (math.Pi * 0.5)
			sinA := float32(math.Sin(float64(ang)))
			// Near the crest (t -> 0), the paper emerges vertically out of the fold
			// and curves over the top of the cylinder before sweeping toward -w.side.
			rolledDist := curlBand*(0.28*t+0.72*sinA) - d*0.08*sinP
			outx = uFold - rolledDist
			// Arch rises to peak height at the cylinder crown and descends as it unrolls.
			arch := float32(math.Sin(float64(clamp01(t*0.85+0.15)) * math.Pi))
			depth = lift * (0.65 + 0.35*arch) * (1 - 0.35*t)
			// Cylindrical shading on the rolled back: warm highlight on the crown,
			// deeper roll-off right at the fold crease (t -> 0).
			crease := 1 - t
			crown := float32(math.Sin(float64(t) * math.Pi))
			shade = 1 - 0.18*crease*crease*sinP + 0.05*crown*sinP
		} else {
			// Beyond the cylinder curl, the sheet sweeps in a gentle catenary arch
			// toward the target page, settling flat as p -> 1.
			over := d - curlBand
			rem := max(width-uFold-curlBand, 1)
			q := clamp01(over / rem)
			baseRolled := curlBand * (1 - 0.08*sinP)
			outx = uFold - (baseRolled + over*(1-0.04*sinP))
			// Matches the curl band's depth at its end (lift * 0.65 * 0.65) and
			// decays as the sheet straightens toward its landing point.
			depth = lift * 0.4225 * (1 - q*q)
			// Soft ambient lift near the spine when the sheet bridges across the binding.
			bridge := float32(math.Sin(float64(q) * math.Pi))
			shade = 1 + 0.025*bridge*sinP - 0.05*(1-q)*sinP
		}
	}

	// Perspective projection from viewer eye above the center of the spread.
	eye := max(turnEye, 900)
	k := eye / max(eye-depth, 900)
	midY := w.page.Y + w.page.Height*0.50
	// Gentle upward arch of the lifted cylinder so the top and bottom edges
	// visibly curve like a real flexible page.
	archLift := (depth / 95.0) * turnArch

	screenX := w.spineX + w.side*outx*k
	screenY := midY + (y-midY)*k - archLift
	return ui.Point{X: screenX, Y: screenY}, clamp01(shade), true
}

// crestAndEdgePoints returns the top and bottom screen points of the rolling
// fold crest and of the sheet's leading fore-edge (the outer corner of the
// rolled-over flap).
func (w *leafWarp) crestAndEdgePoints() (crestTop, crestBot, edgeTop, edgeBot ui.Point) {
	uFold, _, _ := w.rollGeometry()

	// The crest is the highest point of the curl: project it on the front face.
	front := *w
	front.back = false
	crestMatX := w.spineX + w.side*uFold
	cTop, _, ok1 := front.project(crestMatX, w.page.Y)
	cBot, _, ok2 := front.project(crestMatX, w.page.Y+w.page.Height)
	if !ok1 || !ok2 {
		cTop = ui.Point{X: crestMatX, Y: w.page.Y}
		cBot = ui.Point{X: crestMatX, Y: w.page.Y + w.page.Height}
	}

	// The leading edge is the outer fore-edge (distance W from the spine) as
	// the back face carries it across the spread.
	back := *w
	back.back = true
	outerMatX := w.spineX - w.side*w.page.Width
	eTop, _, _ := back.project(outerMatX, w.page.Y)
	eBot, _, _ := back.project(outerMatX, w.page.Y+w.page.Height)
	return cTop, cBot, eTop, eBot
}

// Quad places a box of the page on screen, clipping any box that has not yet
// entered (or has already passed) the rolling fold crest.
func (w *leafWarp) Quad(box ui.Rect) (ui.Quad, bool) {
	uFold, _, _ := w.rollGeometry()

	// Distance from spine at both horizontal ends of box.
	u0 := abs(box.X - w.spineX)
	u1 := abs(box.X + box.Width - w.spineX)
	uMin := min(u0, u1)
	uMax := max(u0, u1)

	if !w.back {
		// Front face is visible only for u <= uFold.
		if uMin >= uFold-0.5 {
			return ui.Quad{}, false
		}
		if uMax > uFold {
			// Clamp the outer edge of the box to the fold crest so paper/rules
			// stop cleanly at the rolling crease.
			crestX := w.spineX + w.side*uFold
			if w.side > 0 {
				box.Width = max(crestX-box.X, 0.5)
			} else {
				right := box.X + box.Width
				box.X = min(crestX, right-0.5)
				box.Width = right - box.X
			}
		}
	} else {
		// Rolled-over back face is visible only for u >= uFold.
		if uMax <= uFold+0.5 {
			return ui.Quad{}, false
		}
		if uMin < uFold {
			// For text labels, skip only the extreme sliver just after the fold
			// line crosses them, where glyphs would crush into noise; they grow
			// continuously from there. Wide paper strips and rules always clamp
			// to the fold crest so the curl shows no gaps.
			if box.Height < 40 && box.Width < w.page.Width*0.92 {
				frac := (uMax - uFold) / max(uMax-uMin, 1)
				if frac < 0.18 {
					return ui.Quad{}, false
				}
			}
			// Notice that on the target page, the spine is at w.spineX, and points
			// on the target page lie in direction -w.side from w.spineX.
			crestX := w.spineX - w.side*uFold
			if w.side > 0 {
				// Target page is left of spine (box.X + box.Width <= spineX).
				// u < uFold means X > crestX, so clamp the right edge (box.X + box.Width) to crestX.
				if box.X+box.Width > crestX {
					box.Width = max(crestX-box.X, 0.5)
				}
			} else {
				// Target page is right of spine (box.X >= spineX).
				// u < uFold means X < crestX, so clamp the left edge (box.X) to crestX.
				if box.X < crestX {
					right := box.X + box.Width
					box.X = min(crestX, right-0.5)
					box.Width = right - box.X
				}
			}
		}
	}

	topLeft, sTL, ok1 := w.project(box.X, box.Y)
	topRight, sTR, ok2 := w.project(box.X+box.Width, box.Y)
	bottomLeft, _, ok3 := w.project(box.X, box.Y+box.Height)
	bottomRight, _, ok4 := w.project(box.X+box.Width, box.Y+box.Height)
	if !ok1 || !ok2 || !ok3 || !ok4 {
		return ui.Quad{}, false
	}
	if topRight.X < topLeft.X+0.35 {
		return ui.Quad{}, false
	}
	return ui.Quad{
		Points: [4]ui.Point{topLeft, topRight, bottomLeft, bottomRight},
		Tint: [4]ui.Color{
			w.inkColor(sTL),
			w.inkColor(sTR),
			w.inkColor(sTL),
			w.inkColor(sTR),
		},
	}, true
}

func (w *leafWarp) inkColor(shade float32) ui.Color {
	v := clamp01(shade)
	return ui.RGBA(uint8(255*v), uint8(255*v), uint8(255*v), 1)
}

// paper is the sheet's surface itself: the same quad, with the base colour
// darkened or brightened by the cylindrical shading.
func (w *leafWarp) paper(box ui.Rect, base ui.Color) (ui.Quad, bool) {
	quad, ok := w.Quad(box)
	if !ok {
		return quad, false
	}
	// An edge that has rolled past the crest reports no shade at all; keep the
	// strip's other edge rather than letting the corner go black.
	_, sTL, okTL := w.project(box.X, box.Y)
	_, sTR, okTR := w.project(box.X+box.Width, box.Y)
	if !okTL {
		sTL = 1
	}
	if !okTR {
		sTR = sTL
	}
	quad.Tint = [4]ui.Color{
		dim(base, sTL),
		dim(base, sTR),
		dim(base, sTL),
		dim(base, sTR),
	}
	return quad, true
}

// paintCurvedPaper subdivides a page-wide box into vertical strips so the top
// and bottom edges of the turning page trace a smooth 3D arch and cylindrical
// specular highlight rather than a single flat straight line.
func (w *leafWarp) paintCurvedPaper(c *ui.Canvas, box ui.Rect, base ui.Color) {
	if box.Width < 24 {
		if quad, ok := w.paper(box, base); ok {
			c.FillQuad(quad)
		}
		return
	}
	strips := max(turnStrips, 4)
	step := box.Width / float32(strips)
	for i := 0; i < strips; i++ {
		sx := box.X + float32(i)*step
		sw := step
		if i == strips-1 {
			sw = box.X + box.Width - sx
		}
		strip := ui.Rect{X: sx, Y: box.Y, Width: sw, Height: box.Height}
		if quad, ok := w.paper(strip, base); ok {
			c.FillQuad(quad)
		}
	}
}

// paintCurlShadows darkens the newly revealed strip just behind the rolling
// fold crest, the ambient occlusion the curl throws onto the page it peeled
// away from. The flap's own shadow is a separate pass (paintFlapShadow) that
// runs between the sheet's two faces.
func (w *leafWarp) paintCurlShadows(c *ui.Canvas, spreadBounds ui.Rect) {
	sinP := float32(math.Sin(float64(clamp01(w.p)) * math.Pi))
	if sinP <= 0.002 {
		return
	}
	cTop, cBot, _, _ := w.crestAndEdgePoints()
	_, _, lift := w.rollGeometry()

	c.Clip(ui.Rect{
		X:      spreadBounds.X + 1,
		Y:      spreadBounds.Y + 1,
		Width:  spreadBounds.Width - 2,
		Height: spreadBounds.Height - 2,
	})

	// Under-curl shadow cast onto the newly revealed page just behind the
	// rolling fold crest (in direction +w.side from cTop/cBot). Its width
	// follows the height the paper is lifted, so a sheet barely off the page
	// throws a narrow shade instead of a bar.
	w.shadeBand(c, cTop, cBot, (12+0.35*lift)*w.side, 0.17*sinP, 0)
	c.Unclip()
}

// paintFlapShadow throws the folded flap's shadow across whatever lies just
// beyond its leading edge — the flat front face early in the turn, the page
// being covered once the flap crosses the spine. It must paint after the
// front face and before the back face, so the flap itself stays above it.
func (w *leafWarp) paintFlapShadow(c *ui.Canvas, spreadBounds ui.Rect) {
	sinP := float32(math.Sin(float64(clamp01(w.p)) * math.Pi))
	if sinP <= 0.002 {
		return
	}
	_, _, eTop, eBot := w.crestAndEdgePoints()
	_, _, lift := w.rollGeometry()

	c.Clip(ui.Rect{
		X:      spreadBounds.X + 1,
		Y:      spreadBounds.Y + 1,
		Width:  spreadBounds.Width - 2,
		Height: spreadBounds.Height - 2,
	})

	// The contact shade is darkest a few px inside the flap, where the paper
	// above hides it, and only a third of it shows at the leading edge; a bar
	// of full-strength shadow against bright paper is what reads as a black
	// border rather than as a shadow.
	w.shadeBand(c, eTop, eBot, -(18+0.55*lift)*w.side, 0.11*sinP, 0)
	c.Unclip()
}

// shadeBand paints a horizontal gradient band hanging off two page-height
// points: `near` alpha at the points themselves, `far` at `width` px away. A
// positive width runs toward +X, a negative one toward -X.
func (w *leafWarp) shadeBand(c *ui.Canvas, top, bot ui.Point, width, near, far float32) {
	if width == 0 {
		return
	}
	nearColor := ui.RGBA(16, 18, 14, clamp01(near))
	farColor := ui.RGBA(16, 18, 14, clamp01(far))
	endTop := ui.Point{X: top.X + width, Y: top.Y}
	endBot := ui.Point{X: bot.X + width, Y: bot.Y}
	if width > 0 {
		c.FillQuad(ui.Quad{
			Points: [4]ui.Point{top, endTop, bot, endBot},
			Tint:   [4]ui.Color{nearColor, farColor, nearColor, farColor},
		})
		return
	}
	c.FillQuad(ui.Quad{
		Points: [4]ui.Point{endTop, top, endBot, bot},
		Tint:   [4]ui.Color{farColor, nearColor, farColor, nearColor},
	})
}

// paintSheetForeEdge draws the warm cylindrical highlight along the rolling
// fold crest and the crisp cut-paper edge at the leading fore-edge of the
// rolled-over sheet.
func (w *leafWarp) paintSheetForeEdge(c *ui.Canvas) {
	sinP := float32(math.Sin(float64(clamp01(w.p)) * math.Pi))
	if sinP <= 0.002 {
		return
	}
	cTop, cBot, eTop, eBot := w.crestAndEdgePoints()

	// Soft warm highlight band along the rolling fold crest.
	crestW := (4 + 10*sinP) * (-w.side)
	inTop := ui.Point{X: cTop.X + crestW, Y: cTop.Y}
	inBot := ui.Point{X: cBot.X + crestW, Y: cBot.Y}
	hi := ui.RGBA(255, 253, 247, clamp01(0.42*sinP))
	clearHi := ui.RGBA(255, 253, 247, 0)
	if w.side > 0 {
		c.FillQuad(ui.Quad{
			Points: [4]ui.Point{inTop, cTop, inBot, cBot},
			Tint:   [4]ui.Color{clearHi, hi, clearHi, hi},
		})
	} else {
		c.FillQuad(ui.Quad{
			Points: [4]ui.Point{cTop, inTop, cBot, inBot},
			Tint:   [4]ui.Color{hi, clearHi, hi, clearHi},
		})
	}

	// Cut paper thickness along the leading fore-edge (eTop..eBot).
	rimW := (1.6 + 1.0*sinP) * (-w.side)
	p0 := eTop
	p1 := ui.Point{X: eTop.X + rimW, Y: eTop.Y}
	p2 := eBot
	p3 := ui.Point{X: eBot.X + rimW, Y: eBot.Y}
	if w.side > 0 {
		p0, p1 = p1, p0
		p2, p3 = p3, p2
	}
	c.FillQuad(ui.Quad{
		Points: [4]ui.Point{p0, p1, p2, p3},
		Tint:   [4]ui.Color{theme.PaperBoard, theme.PaperRule, theme.PaperBoard, theme.PaperRule},
	})
}

// dim multiplies a colour's channels, keeping its alpha.
func dim(color ui.Color, k float32) ui.Color {
	dimmed := color
	dimmed.Red = clamp01(dimmed.Red * k)
	dimmed.Green = clamp01(dimmed.Green * k)
	dimmed.Blue = clamp01(dimmed.Blue * k)
	return dimmed
}

func abs(v float32) float32 {
	if v < 0 {
		return -v
	}
	return v
}

func clamp01(v float32) float32 {
	return min(max(v, 0), 1)
}

// Warp is what the reader's paragraphs paint through while a leaf is turning.
var _ render.Warp = (*leafWarp)(nil)

// pFill fills a page box, through the turning sheet's projection while one is
// being painted.
func (m *State) pFill(c *ui.Canvas, box ui.Rect, color ui.Color) {
	if m.warp == nil {
		c.Fill(box, color)
		return
	}
	m.warp.paintCurvedPaper(c, box, color)
}

// pRound fills a rounded page box, curved into strips while it is part of a turning sheet.
func (m *State) pRound(c *ui.Canvas, box ui.Rect, radius float32, color ui.Color) {
	if m.warp == nil {
		c.Rounded(box, radius, color)
		return
	}
	m.warp.paintCurvedPaper(c, box, color)
}

// pLabel draws a line of page text, warped with the sheet it is printed on.
func (m *State) pLabel(c *ui.Canvas, value string, box ui.Rect, style ui.TextStyle) {
	m.pLabelAligned(c, value, box, style, ui.AlignStart)
}

// pLabelCenter draws a centred line of page text.
func (m *State) pLabelCenter(c *ui.Canvas, value string, box ui.Rect, style ui.TextStyle) {
	m.pLabelAligned(c, value, box, style, ui.AlignCenter)
}

// pLabelAligned draws a page line with explicit horizontal placement.
func (m *State) pLabelAligned(c *ui.Canvas, value string, box ui.Rect,
	style ui.TextStyle, align ui.Alignment) {
	if m.warp == nil {
		c.LabelAligned(value, box, style, align)
		return
	}
	if align == ui.AlignCenter || align == ui.AlignRight {
		inset := box.Width - c.TextWidthStyle(value, style)
		if align == ui.AlignCenter {
			inset /= 2
		}
		if inset > 0 {
			box.X += inset
			box.Width = max(box.Width-inset, 1)
		}
	}
	if quad, ok := m.warp.Quad(box); ok {
		c.LabelQuad(value, box, quad, style)
	}
}

// pParagraph draws a laid-out paragraph, warped with the sheet.
func (m *State) pParagraph(c *ui.Canvas, p render.Paragraph, x, y float32, s render.LineStyle) {
	if m.warp == nil {
		p.Paint(c, x, y, s)
		return
	}
	p.PaintWarped(c, x, y, s, m.warp)
}
