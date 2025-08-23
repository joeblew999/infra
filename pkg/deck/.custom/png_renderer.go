package deck

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/ajstarks/deck"
	"github.com/disintegration/gift"
	"github.com/fogleman/gg"
)

// Use shared constants from consts.go

// PNGRenderer handles conversion from decksh DSL to PNG format (following pngdeck patterns)
type PNGRenderer struct {
	renderer *Renderer
	width    int
	height   int
	fontMap  map[string]string
}

var codemap = strings.NewReplacer("\t", "    ")

// NewPNGRenderer creates a new PNG renderer
func NewPNGRenderer(width, height float64) *PNGRenderer {
	return &PNGRenderer{
		renderer: NewRenderer(width, height),
		width:    int(width),
		height:   int(height),
		fontMap:  make(map[string]string),
	}
}

// DeckshToPNG converts decksh DSL to PNG bytes using pngdeck patterns
func (p *PNGRenderer) DeckshToPNG(dshInput string, opts RenderOptions) ([]byte, error) {
	// First convert decksh to XML
	xmlContent, err := p.renderer.DeckshToXML(dshInput)
	if err != nil {
		return nil, fmt.Errorf("failed to convert decksh to XML: %w", err)
	}

	// Parse XML into deck structure
	xmlReader := strings.NewReader(xmlContent)
	d, err := deck.ReadDeck(io.NopCloser(xmlReader), p.width, p.height)
	if err != nil {
		return nil, fmt.Errorf("failed to parse deck XML: %w", err)
	}

	// Set up font mapping with pkg/font integration
	p.setupFontMap(opts)

	// Create context and render
	doc := gg.NewContext(p.width, p.height)
	
	// Render the first slide (for now)
	if len(d.Slide) > 0 {
		p.pngslide(doc, d, 0, opts.GridPercent, true, opts.Layers)
	}

	// Encode to PNG
	var buf bytes.Buffer
	err = png.Encode(&buf, doc.Image())
	if err != nil {
		return nil, fmt.Errorf("failed to encode PNG: %w", err)
	}

	return buf.Bytes(), nil
}

// setupFontMap configures font mapping with pkg/font integration
func (p *PNGRenderer) setupFontMap(opts RenderOptions) {
	// Initialize with defaults
	p.fontMap["sans"] = "sans"
	p.fontMap["serif"] = "serif"
	p.fontMap["mono"] = "mono"
	p.fontMap["symbol"] = "symbol"

	// Try to get actual font paths from pkg/font
	if p.renderer.fontManager != nil {
		// Try to get font paths for common families
		families := []string{"Arial", "Helvetica", "Times", "Courier"}
		aliases := []string{"sans", "serif", "serif", "mono"}
		
		for i, family := range families {
			if fontPath, err := p.renderer.GetFontPath(family, opts.FontWeight); err == nil {
				p.fontMap[aliases[i]] = fontPath
			}
		}
	}
}

// includefile returns the contents of a file as string
func (p *PNGRenderer) includefile(filename string) string {
	data, err := os.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return ""
	}
	return codemap.Replace(string(data))
}

// pct converts percentages to canvas measures
func (p *PNGRenderer) pct(percent, measure float64) float64 {
	return (percent / 100.0) * measure
}

// dimen returns canvas dimensions from percentages
func (p *PNGRenderer) dimen(w, h, xp, yp, sp float64) (float64, float64, float64) {
	return p.pct(xp, w), p.pct(100-yp, h), p.pct(sp, w) * FontFactor
}

// fontlookup maps font aliases to implementation font names
func (p *PNGRenderer) fontlookup(s string) string {
	font, ok := p.fontMap[s]
	if ok {
		return font
	}
	return "sans"
}

// grid makes a percentage scale
func (p *PNGRenderer) grid(doc *gg.Context, w, h float64, color string, percent float64) {
	pw := w * (percent / 100)
	ph := h * (percent / 100)
	r, g, b := colorlookup(color)
	doc.SetRGB255(r, g, b)
	doc.SetLineWidth(0.25)
	fs := p.pct(1, w)
	for x, pl := 0.0, 0.0; x <= w; x += pw {
		doc.DrawLine(x, 0, x, h)
		doc.Stroke()
		if pl > 0 {
			p.showtext(doc, x, h-fs, fmt.Sprintf("%.0f", pl), fs, "sans", "center")
		}
		pl += percent
	}
	for y, pl := 0.0, 0.0; y <= h; y += ph {
		doc.DrawLine(0, y, w, y)
		doc.Stroke()
		if pl < 100 {
			p.showtext(doc, fs, y+(fs/3), fmt.Sprintf("%.0f", 100-pl), fs, "sans", "center")
		}
		pl += percent
	}
}

// setop sets the opacity as a truncated fraction of 255
func (p *PNGRenderer) setop(v float64) int {
	if v > 0.0 {
		return int(255.0 * (v / 100.0))
	}
	return 255
}

// bullet draws a bullet
func (p *PNGRenderer) bullet(doc *gg.Context, x, y, size float64, color string) {
	rs := size / 2
	r, g, b := colorlookup(color)
	doc.SetRGB255(r, g, b)
	doc.DrawCircle(x-size*2, y-rs, rs)
	doc.Fill()
}

// background places a colored rectangle
func (p *PNGRenderer) background(doc *gg.Context, w, h float64, color string) {
	r, g, b := colorlookup(color)
	doc.SetRGB255(r, g, b)
	doc.Clear()
}

// gradientbg sets the background color gradient
func (p *PNGRenderer) gradientbg(doc *gg.Context, w, h float64, gc1, gc2 string, gp float64) {
	p.gradient(doc, 0, 0, w, h, gc1, gc2, gp)
}

// gradient sets the rect color gradient
func (p *PNGRenderer) gradient(doc *gg.Context, x, y, w, h float64, gc1, gc2 string, gp float64) {
	r1, g1, b1 := colorlookup(gc1)
	r2, g2, b2 := colorlookup(gc2)
	gp /= 100.0
	grad := gg.NewLinearGradient(x, y, x+w, y+h)
	grad.AddColorStop(0, color.RGBA{uint8(r1), uint8(g1), uint8(b1), 1})
	grad.AddColorStop(1, color.RGBA{uint8(r2), uint8(g2), uint8(b2), 1})
	doc.SetFillStyle(grad)
	doc.DrawRectangle(x, y, w, h)
	doc.Fill()
}

// doline draws a line
func (p *PNGRenderer) doline(doc *gg.Context, xp1, yp1, xp2, yp2, sw float64, color string, opacity float64) {
	r, g, b := colorlookup(color)
	doc.SetLineWidth(sw)
	doc.SetRGBA255(r, g, b, p.setop(opacity))
	doc.SetLineCapButt()
	doc.DrawLine(xp1, yp1, xp2, yp2)
	doc.Stroke()
}

// doarc draws an arc
func (p *PNGRenderer) doarc(doc *gg.Context, x, y, w, h, a1, a2, sw float64, color string, opacity float64) {
	r, g, b := colorlookup(color)
	doc.SetLineWidth(sw)
	doc.SetRGBA255(r, g, b, p.setop(opacity))
	doc.SetLineCapButt()
	doc.DrawEllipticalArc(x, y, w, h, gg.Radians(360-a1), gg.Radians(360-a2))
	doc.Stroke()
}

// docurve draws a bezier curve
func (p *PNGRenderer) docurve(doc *gg.Context, xp1, yp1, xp2, yp2, xp3, yp3, sw float64, color string, opacity float64) {
	r, g, b := colorlookup(color)
	doc.SetLineWidth(sw)
	doc.SetLineCapButt()
	doc.SetRGBA255(r, g, b, p.setop(opacity))
	doc.MoveTo(xp1, yp1)
	doc.QuadraticTo(xp2, yp2, xp3, yp3)
	doc.Stroke()
}

// dorect draws a rectangle
func (p *PNGRenderer) dorect(doc *gg.Context, x, y, w, h float64, color string, opacity float64) {
	r, g, b := colorlookup(color)
	doc.SetRGBA255(r, g, b, p.setop(opacity))
	doc.DrawRectangle(x, y, w, h)
	doc.Fill()
}

// doellipse draws an ellipse
func (p *PNGRenderer) doellipse(doc *gg.Context, x, y, w, h float64, color string, opacity float64) {
	r, g, b := colorlookup(color)
	doc.SetRGBA255(r, g, b, p.setop(opacity))
	doc.DrawEllipse(x, y, w, h)
	doc.Fill()
}

// dopoly draws a polygon
func (p *PNGRenderer) dopoly(doc *gg.Context, xc, yc string, cw, ch float64, color string, opacity float64) {
	xs := strings.Split(xc, " ")
	ys := strings.Split(yc, " ")
	if len(xs) != len(ys) {
		return
	}
	if len(xs) < 3 || len(ys) < 3 {
		return
	}
	doc.NewSubPath()
	for i := 0; i < len(xs); i++ {
		x, err := strconv.ParseFloat(xs[i], 64)
		if err != nil {
			x = 0
		} else {
			x = p.pct(x, cw)
		}
		y, err := strconv.ParseFloat(ys[i], 64)
		if err != nil {
			y = 0
		} else {
			y = p.pct(100-y, ch)
		}
		doc.LineTo(x, y)
	}
	doc.ClosePath()
	r, g, b := colorlookup(color)
	doc.SetRGBA255(r, g, b, p.setop(opacity))
	doc.Fill()
}

// dotext places text elements on the canvas according to type
func (p *PNGRenderer) dotext(doc *gg.Context, cw, x, y, fs, wp, rotation, spacing float64, tdata, font, align, ttype, color string, opacity float64) {
	var tw float64

	td := strings.Split(tdata, "\n")
	red, green, blue := colorlookup(color)
	if rotation > 0 {
		doc.Push()
		doc.RotateAbout(gg.Radians(360-rotation), x, y)
	}
	if ttype == "code" {
		font = "mono"
		ch := float64(len(td)) * spacing * fs
		tw = deck.Pwidth(wp, cw, cw-x-20)
		p.dorect(doc, x-fs, y-fs, tw, ch, "rgb(240,240,240)", 100)
	}
	doc.SetRGBA255(red, green, blue, p.setop(opacity))
	if ttype == "block" {
		tw = deck.Pwidth(wp, cw, cw/2)
		p.textwrap(doc, x, y, tw, fs, fs*spacing, tdata, font)
	} else {
		ls := spacing * fs
		for _, t := range td {
			p.showtext(doc, x, y, t, fs, font, align)
			y += ls
		}
	}
	if rotation > 0 {
		doc.Pop()
	}
}

// whitespace determines if a rune is whitespace
func (p *PNGRenderer) whitespace(r rune) bool {
	return r == ' ' || r == '\n' || r == '\t'
}

// loadfont loads a font at the specified size
func (p *PNGRenderer) loadfont(doc *gg.Context, s string, size float64) {
	fontPath := p.fontlookup(s)
	
	// Try to load from pkg/font system first
	var err error
	
	if p.renderer.fontManager != nil {
		// Get font from pkg/font if available
		if actualPath, fontErr := p.renderer.GetFontPath(s, 400); fontErr == nil {
			err = doc.LoadFontFace(actualPath, size)
		} else {
			// Fall back to system font loading
			err = doc.LoadFontFace(fontPath, size)
		}
	} else {
		// Direct font loading
		err = doc.LoadFontFace(fontPath, size)
	}
	
	if err != nil {
		fmt.Fprintf(os.Stderr, "pngdeck %v\n", err)
		return
	}
}

// textwrap draws text at location, wrapping at the specified width
func (p *PNGRenderer) textwrap(doc *gg.Context, x, y, w, fs, leading float64, s, font string) int {
	var factor = 0.3
	if font == "mono" {
		factor = 1.0
	}
	nbreak := 0
	p.loadfont(doc, font, fs)
	wordspacing, _ := doc.MeasureString("M")
	words := strings.FieldsFunc(s, p.whitespace)
	xp := x
	yp := y
	edge := x + w
	for _, s := range words {
		if s == "\\n" { // magic new line
			xp = x
			yp += leading
			nbreak++
			continue
		}
		tw, _ := doc.MeasureString(s)
		doc.DrawString(s, xp, yp)
		xp += tw + (wordspacing * factor)
		if xp > edge {
			xp = x
			yp += leading
			nbreak++
		}
	}
	return nbreak
}

// showtext places fully attributed text at the specified location
func (p *PNGRenderer) showtext(doc *gg.Context, x, y float64, s string, fs float64, font, align string) {
	offset := 0.0
	p.loadfont(doc, font, fs)
	t := s
	tw, _ := doc.MeasureString(t)
	switch align {
	case "center", "middle", "mid", "c":
		offset = (tw / 2)
	case "right", "end", "e":
		offset = tw
	}
	doc.DrawString(t, x-offset, y)
}

// dolist places lists on the canvas
func (p *PNGRenderer) dolist(doc *gg.Context, cw, x, y, fs, lwidth, rotation, spacing float64, list []deck.ListItem, font, ltype, align, color string, opacity float64) {
	if font == "" {
		font = "sans"
	}
	red, green, blue := colorlookup(color)

	if ltype == "bullet" {
		x += fs * 1.2
	}
	ls := spacing * fs
	tw := deck.Pwidth(lwidth, cw, cw/2)

	if rotation > 0 {
		doc.Push()
		doc.RotateAbout(gg.Radians(360-rotation), x, y)
	}
	var t string
	for i, tl := range list {
		p.loadfont(doc, font, fs)
		doc.SetRGBA255(red, green, blue, p.setop(tl.Opacity))
		if ltype == "number" {
			t = fmt.Sprintf("%d. ", i+1) + tl.ListText
		} else {
			t = tl.ListText
		}
		if ltype == "bullet" {
			p.bullet(doc, x, y, fs/2, color)
		}
		if len(tl.Color) > 0 {
			tlred, tlgreen, tlblue := colorlookup(tl.Color)
			doc.SetRGB255(tlred, tlgreen, tlblue)
		}
		if len(tl.Font) > 0 {
			p.loadfont(doc, tl.Font, fs)
		}
		if align == "center" || align == "c" {
			p.showtext(doc, x, y, t, fs, font, align)
			y += ls
		} else {
			yw := p.textwrap(doc, x, y, tw, fs, ls, t, font)
			y += ls
			if yw >= 1 {
				y += ls * float64(yw)
			}
		}
	}
	if rotation > 0 {
		doc.Pop()
	}
}

// pngslide makes a slide, one slide per generated PNG (adapted from pngdeck)
func (p *PNGRenderer) pngslide(doc *gg.Context, d deck.Deck, n int, gp float64, showslide bool, layers string) {
	if n < 0 || n > len(d.Slide)-1 || !showslide {
		return
	}

	var x, y, fs float64

	cw := float64(d.Canvas.Width)
	ch := float64(d.Canvas.Height)
	slide := d.Slide[n]
	
	// set default background
	if slide.Bg == "" {
		slide.Bg = "white"
	}
	p.background(doc, cw, ch, slide.Bg)

	if slide.GradPercent <= 0 || slide.GradPercent > 100 {
		slide.GradPercent = 100
	}
	// set gradient background, if specified. You need both colors
	if len(slide.Gradcolor1) > 0 && len(slide.Gradcolor2) > 0 {
		p.gradientbg(doc, cw, ch, slide.Gradcolor1, slide.Gradcolor2, slide.GradPercent)
	}
	// set the default foreground
	if slide.Fg == "" {
		slide.Fg = "black"
	}
	const defaultColor = "rgb(127,127,127)"
	layerlist := strings.Split(layers, ":")
	
	// draw elements in the order of the layer list
	for il := 0; il < len(layerlist); il++ {
		switch layerlist[il] {
		// for every image on the slide...
		case "image":
			for _, im := range slide.Image {
				x, y, _ = p.dimen(cw, ch, im.Xp, im.Yp, 0)
				iw, ih := im.Width, im.Height
				// scale the image by the specified percentage
				if im.Scale > 0 {
					iw = int(float64(iw) * (im.Scale / 100))
					ih = int(float64(ih) * (im.Scale / 100))
				}
				// scale the image to fit the canvas width
				if im.Autoscale == "on" && iw < d.Canvas.Width {
					ih = int((float64(d.Canvas.Width) / float64(iw)) * float64(ih))
					iw = d.Canvas.Width
				}

				img, err := gg.LoadImage(im.Name)
				if err != nil {
					fmt.Fprintf(os.Stderr, "pngdeck: slide %d (%v)\n", n+1, err)
					return
				}

				bounds := img.Bounds()
				nw, nh := bounds.Max.X-bounds.Min.X, bounds.Max.Y-bounds.Min.Y
				// scale the image to a percentage of the canvas width
				if im.Height == 0 && im.Width > 0 {
					if nh > 0 {
						fw := float64(im.Width)
						imscale := (fw / 100) * cw
						fw = imscale
						fh := imscale / (float64(nw) / float64(nh))
						iw = int(fw)
						ih = int(fh)
					}
				}

				if iw == nw && ih == nh {
					doc.DrawImageAnchored(img, int(x), int(y), 0.5, 0.5)
				} else {
					g := gift.New(gift.Resize(iw, ih, gift.BoxResampling))
					resized := image.NewRGBA(g.Bounds(img.Bounds()))
					g.Draw(resized, img)
					doc.DrawImageAnchored(resized, int(x), int(y), 0.5, 0.5)
				}
				if len(im.Caption) > 0 {
					capsize := deck.Pwidth(im.Sp, cw, p.pct(2, cw))
					if im.Font == "" {
						im.Font = "sans"
					}
					if im.Color == "" {
						im.Color = slide.Fg
					}
					if im.Align == "" {
						im.Align = "center"
					}
					midx := float64(iw) / 2
					midy := float64(ih) / 2
					switch im.Align {
					case "left", "start":
						x -= midx
					case "right", "end":
						x += midx
					}
					capr, capg, capb := colorlookup(im.Color)
					doc.SetRGB255(capr, capg, capb)
					p.showtext(doc, x, y+(midy)+(capsize*1.5), im.Caption, capsize, im.Font, im.Align)
				}
			}
		// every graphic on the slide
		case "rect":
			// rect
			for _, rect := range slide.Rect {
				x, y, _ := p.dimen(cw, ch, rect.Xp, rect.Yp, 0)
				var w, h float64
				w = p.pct(rect.Wp, cw)
				if rect.Hr == 0 {
					h = p.pct(rect.Hp, ch)
				} else {
					h = p.pct(rect.Hr, w)
				}
				if rect.Color == "" {
					rect.Color = defaultColor
				}
				if len(rect.Gradcolor1) > 0 && len(rect.Gradcolor2) > 0 {
					p.gradient(doc, x-(w/2), y-(h/2), w, h, rect.Gradcolor1, rect.Gradcolor2, rect.GradPercent)
				} else {
					p.dorect(doc, x-(w/2), y-(h/2), w, h, rect.Color, rect.Opacity)
				}
			}
		case "ellipse":
			// ellipse
			for _, ellipse := range slide.Ellipse {
				x, y, _ := p.dimen(cw, ch, ellipse.Xp, ellipse.Yp, 0)
				var w, h float64
				w = p.pct(ellipse.Wp, cw)
				if ellipse.Hr == 0 {
					h = p.pct(ellipse.Hp, ch)
				} else {
					h = p.pct(ellipse.Hr, w)
				}
				if ellipse.Color == "" {
					ellipse.Color = defaultColor
				}
				p.doellipse(doc, x, y, w/2, h/2, ellipse.Color, ellipse.Opacity)
			}
		case "curve":
			// curve
			for _, curve := range slide.Curve {
				if curve.Color == "" {
					curve.Color = defaultColor
				}
				x1, y1, sw := p.dimen(cw, ch, curve.Xp1, curve.Yp1, curve.Sp)
				x2, y2, _ := p.dimen(cw, ch, curve.Xp2, curve.Yp2, 0)
				x3, y3, _ := p.dimen(cw, ch, curve.Xp3, curve.Yp3, 0)
				if sw == 0 {
					sw = 2.0
				}
				p.docurve(doc, x1, y1, x2, y2, x3, y3, sw, curve.Color, curve.Opacity)
			}
		case "arc":
			// arc
			for _, arc := range slide.Arc {
				if arc.Color == "" {
					arc.Color = defaultColor
				}
				x, y, sw := p.dimen(cw, ch, arc.Xp, arc.Yp, arc.Sp)
				w := p.pct(arc.Wp, cw)
				h := p.pct(arc.Hp, cw)
				if sw == 0 {
					sw = 2.0
				}
				p.doarc(doc, x, y, w/2, h/2, arc.A1, arc.A2, sw, arc.Color, arc.Opacity)
			}
		case "line":
			// line
			for _, line := range slide.Line {
				if line.Color == "" {
					line.Color = defaultColor
				}
				x1, y1, sw := p.dimen(cw, ch, line.Xp1, line.Yp1, line.Sp)
				x2, y2, _ := p.dimen(cw, ch, line.Xp2, line.Yp2, 0)
				if sw == 0 {
					sw = 2.0
				}
				p.doline(doc, x1, y1, x2, y2, sw, line.Color, line.Opacity)
			}
		case "poly":
			// polygon
			for _, poly := range slide.Polygon {
				if poly.Color == "" {
					poly.Color = defaultColor
				}
				p.dopoly(doc, poly.XC, poly.YC, cw, ch, poly.Color, poly.Opacity)
			}
		case "text":
			// for every text element...
			var tdata string
			for _, t := range slide.Text {
				if t.Color == "" {
					t.Color = slide.Fg
				}
				if t.Font == "" {
					t.Font = "sans"
				}
				x, y, fs = p.dimen(cw, ch, t.Xp, t.Yp, t.Sp)
				if t.File != "" {
					tdata = p.includefile(t.File)
				} else {
					tdata = t.Tdata
				}
				if t.Lp == 0 {
					t.Lp = LineSpacing
				}
				p.dotext(doc, cw, x, y, fs, t.Wp, t.Rotation, t.Lp, tdata, t.Font, t.Align, t.Type, t.Color, t.Opacity)
			}
		case "list":
			// for every list element...
			for _, l := range slide.List {
				if l.Color == "" {
					l.Color = slide.Fg
				}
				if l.Lp == 0 {
					l.Lp = ListSpacing
				}
				if l.Wp == 0 {
					l.Wp = ListWrap
				}
				x, y, fs = p.dimen(cw, ch, l.Xp, l.Yp, l.Sp)
				p.dolist(doc, cw, x, y, fs, l.Wp, l.Rotation, l.Lp, l.Li, l.Font, l.Type, l.Align, l.Color, l.Opacity)
			}
		}
	}
	// add a grid, if specified
	if gp > 0 {
		p.grid(doc, cw, ch, slide.Fg, gp)
	}
}