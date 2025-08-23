package deck

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"codeberg.org/go-pdf/fpdf"
	"github.com/ajstarks/deck"
)

// Use shared constants from consts.go

// PDFRenderer handles conversion from decksh DSL to PDF format using deck library
type PDFRenderer struct {
	renderer *Renderer
	width    float64
	height   float64
	fontMap  map[string]string // Font mapping
}

// NewPDFRenderer creates a new PDF renderer
func NewPDFRenderer(width, height float64) *PDFRenderer {
	return &PDFRenderer{
		renderer: NewRenderer(width, height),
		width:    width,
		height:   height,
		fontMap: map[string]string{
			"sans":   "Arial",
			"serif":  "Times",
			"mono":   "Courier",
			"symbol": "Symbol",
		},
	}
}

// DeckshToPDF converts decksh DSL to PDF bytes using the deck library approach
func (p *PDFRenderer) DeckshToPDF(dshInput string, opts RenderOptions) ([]byte, error) {
	// First convert decksh to XML
	xmlContent, err := p.renderer.DeckshToXML(dshInput)
	if err != nil {
		return nil, fmt.Errorf("failed to convert decksh to XML: %w", err)
	}

	// Parse XML into deck structure
	xmlReader := strings.NewReader(xmlContent)
	d, err := deck.ReadDeck(io.NopCloser(xmlReader), int(p.width), int(p.height))
	if err != nil {
		return nil, fmt.Errorf("failed to parse deck XML: %w", err)
	}

	// Create PDF
	pdf := fpdf.New("L", "pt", "", "")
	pageConfig := fpdf.InitType{
		OrientationStr: "L",
		UnitStr:        "pt",
		SizeStr:        "",
		Size:           fpdf.SizeType{Wd: p.width, Ht: p.height},
		FontDirStr:     "",
	}

	// Render slides
	p.renderSlides(pdf, pageConfig, d, opts)

	// Get PDF bytes
	var buf bytes.Buffer
	err = pdf.Output(&buf)
	if err != nil {
		return nil, fmt.Errorf("failed to generate PDF: %w", err)
	}

	return buf.Bytes(), nil
}

// renderSlides renders all slides in the deck
func (p *PDFRenderer) renderSlides(doc *fpdf.Fpdf, pc fpdf.InitType, d deck.Deck, opts RenderOptions) {
	for i := 0; i < len(d.Slide); i++ {
		p.renderSlide(doc, d, i, opts)
	}
}

// renderSlide renders a single slide (adapted from pdfdeck)
func (p *PDFRenderer) renderSlide(doc *fpdf.Fpdf, d deck.Deck, slideIndex int, opts RenderOptions) {
	if slideIndex >= len(d.Slide) {
		return
	}

	slide := d.Slide[slideIndex]
	doc.AddPageFormat("L", fpdf.SizeType{Wd: p.width, Ht: p.height})

	// Set background
	if slide.Bg != "" {
		p.background(doc, p.width, p.height, slide.Bg)
	}

	// Set gradient background if specified
	if slide.Gradcolor1 != "" && slide.Gradcolor2 != "" {
		p.gradientbg(doc, p.width, p.height, slide.Gradcolor1, slide.Gradcolor2, slide.GradPercent)
	}

	// Process layers in order (similar to SVG renderer)
	layerList := strings.Split(opts.Layers, ":")
	for _, layer := range layerList {
		switch layer {
		case "image":
			p.processImages(doc, slide.Image)
		case "rect":
			p.processRects(doc, slide.Rect)
		case "ellipse":
			p.processEllipses(doc, slide.Ellipse)
		case "line":
			p.processLines(doc, slide.Line)
		case "arc":
			p.processArcs(doc, slide.Arc)
		case "curve":
			p.processCurves(doc, slide.Curve)
		case "polygon":
			p.processPolygons(doc, slide.Polygon)
		case "text":
			p.processTexts(doc, slide.Text)
		case "list":
			p.processLists(doc, slide.List)
		}
	}

	// Add grid if specified
	if opts.GridPercent > 0 {
		p.grid(doc, p.width, p.height, slide.Fg, opts.GridPercent)
	}
}

// Helper functions adapted from pdfdeck source

func (p *PDFRenderer) pct(percent, measure float64) float64 {
	return (percent / 100.0) * measure
}

func (p *PDFRenderer) dimen(w, h, xp, yp, sp float64) (float64, float64, float64) {
	return p.pct(xp, w), p.pct(100-yp, h), p.pct(sp, w) * FontFactor
}

func (p *PDFRenderer) background(doc *fpdf.Fpdf, w, h float64, color string) {
	doc.SetFillColor(colorlookup(color))
	doc.Rect(0, 0, w, h, "F")
}

func (p *PDFRenderer) gradientbg(doc *fpdf.Fpdf, w, h float64, gc1, gc2 string, gp float64) {
	// Basic gradient implementation - fpdf has limited gradient support
	// For now, just use the first color as background
	p.background(doc, w, h, gc1)
}

func (p *PDFRenderer) grid(doc *fpdf.Fpdf, w, h float64, color string, percent float64) {
	gridsize := (w * percent) / 100
	doc.SetDrawColor(colorlookup(color))
	doc.SetLineWidth(0.5)

	// Vertical lines
	for x := 0.0; x <= w; x += gridsize {
		doc.Line(x, 0, x, h)
	}
	// Horizontal lines
	for y := 0.0; y <= h; y += gridsize {
		doc.Line(0, y, w, y)
	}
}

func (p *PDFRenderer) processTexts(doc *fpdf.Fpdf, texts []deck.Text) {
	for _, t := range texts {
		if t.Color == "" {
			t.Color = "black"
		}
		x, y, fs := p.dimen(p.width, p.height, t.Xp, t.Yp, t.Sp)
		p.showtext(doc, x, y, t.Tdata, fs, t.Font, t.Align, t.Link)
	}
}

func (p *PDFRenderer) processLists(doc *fpdf.Fpdf, lists []deck.List) {
	for _, l := range lists {
		if l.Color == "" {
			l.Color = "black"
		}
		x, y, fs := p.dimen(p.width, p.height, l.Xp, l.Yp, l.Sp)
		p.renderList(doc, p.width, x, y, fs, l)
	}
}

func (p *PDFRenderer) processRects(doc *fpdf.Fpdf, rects []deck.Rect) {
	for _, r := range rects {
		p.rectangle(doc, 
			p.pct(r.Xp, p.width), p.pct(100-r.Yp, p.height), 
			p.pct(r.Wp, p.width), p.pct(r.Hp, p.height), 
			r.Color)
	}
}

func (p *PDFRenderer) processEllipses(doc *fpdf.Fpdf, ellipses []deck.Ellipse) {
	for _, e := range ellipses {
		p.ellipse(doc, 
			p.pct(e.Xp, p.width), p.pct(100-e.Yp, p.height), 
			p.pct(e.Wp, p.width), p.pct(e.Hp, p.height), 
			e.Color)
	}
}

func (p *PDFRenderer) processLines(doc *fpdf.Fpdf, lines []deck.Line) {
	for _, l := range lines {
		p.line(doc, 
			p.pct(l.Xp1, p.width), p.pct(100-l.Yp1, p.height),
			p.pct(l.Xp2, p.width), p.pct(100-l.Yp2, p.height),
			l.Sp, l.Color)
	}
}

func (p *PDFRenderer) processArcs(doc *fpdf.Fpdf, arcs []deck.Arc) {
	for _, a := range arcs {
		p.arc(doc, 
			p.pct(a.Xp, p.width), p.pct(100-a.Yp, p.height),
			p.pct(a.Wp, p.width), p.pct(a.Hp, p.height),
			a.A1, a.A2, a.Sp, a.Color)
	}
}

func (p *PDFRenderer) processCurves(doc *fpdf.Fpdf, curves []deck.Curve) {
	for _, c := range curves {
		p.quadcurve(doc, 
			p.pct(c.Xp1, p.width), p.pct(100-c.Yp1, p.height),
			p.pct(c.Xp2, p.width), p.pct(100-c.Yp2, p.height),
			p.pct(c.Xp3, p.width), p.pct(100-c.Yp3, p.height),
			c.Sp, c.Color)
	}
}

func (p *PDFRenderer) processPolygons(doc *fpdf.Fpdf, polygons []deck.Polygon) {
	for _, poly := range polygons {
		p.polygon(doc, poly.XC, poly.YC, poly.Color, p.width, p.height)
	}
}

func (p *PDFRenderer) processImages(doc *fpdf.Fpdf, images []deck.Image) {
	// Image processing would require image loading and embedding
	// For now, render placeholder rectangles
	for _, img := range images {
		x, y := p.pct(img.Xp, p.width), p.pct(100-img.Yp, p.height)
		// Draw placeholder rectangle for image
		p.rectangle(doc, x-50, y-25, 100, 50, "lightgray")
	}
}

// Basic shape functions (simplified versions from pdfdeck)

func (p *PDFRenderer) rectangle(doc *fpdf.Fpdf, x, y, w, h float64, color string) {
	doc.SetFillColor(colorlookup(color))
	doc.Rect(x-w/2, y-h/2, w, h, "F")
}

func (p *PDFRenderer) ellipse(doc *fpdf.Fpdf, x, y, w, h float64, color string) {
	doc.SetFillColor(colorlookup(color))
	doc.Ellipse(x, y, w/2, h/2, 0, "F")
}

func (p *PDFRenderer) line(doc *fpdf.Fpdf, x1, y1, x2, y2, sw float64, color string) {
	doc.SetDrawColor(colorlookup(color))
	doc.SetLineWidth(sw)
	doc.Line(x1, y1, x2, y2)
}

func (p *PDFRenderer) arc(doc *fpdf.Fpdf, x, y, w, h, a1, a2, sw float64, color string) {
	doc.SetDrawColor(colorlookup(color))
	doc.SetLineWidth(sw)
	// fpdf arc support is limited, draw as ellipse for now
	doc.Ellipse(x, y, w/2, h/2, 0, "D")
}

func (p *PDFRenderer) quadcurve(doc *fpdf.Fpdf, x1, y1, x2, y2, x3, y3, sw float64, color string) {
	doc.SetDrawColor(colorlookup(color))
	doc.SetLineWidth(sw)
	// Simplified curve as line for now
	doc.Line(x1, y1, x3, y3)
}

func (p *PDFRenderer) polygon(doc *fpdf.Fpdf, xc, yc, color string, cw, ch float64) {
	// Polygon parsing would require coordinate parsing
	// For now, draw a simple shape
	doc.SetFillColor(colorlookup(color))
	doc.Rect(p.pct(50, cw)-25, p.pct(50, ch)-25, 50, 50, "F")
}

func (p *PDFRenderer) showtext(doc *fpdf.Fpdf, x, y float64, s string, fs float64, fontname, align, link string) {
	if fontname == "" {
		fontname = "sans"
	}
	
	// Use pkg/font for custom fonts if available, fallback to PDF built-in fonts
	font := p.fontlookup(fontname)
	if p.renderer.fontManager != nil {
		// Try to get font from pkg/font system
		if fontPath, err := p.renderer.GetFontPath(fontname, 400); err == nil {
			// For PDF, we would need to use AddFont to add custom fonts
			// This is more complex and requires font conversion
			// For now, stick with built-in fonts but use pkg/font for consistency checking
			_ = fontPath // Use built-in font mapping for now
		}
	}
	
	doc.SetFont(font, "", fs)
	doc.SetTextColor(0, 0, 0) // Default black
	
	// Handle alignment
	switch align {
	case "center", "middle", "c":
		x = x - doc.GetStringWidth(s)/2
	case "right", "end", "e":
		x = x - doc.GetStringWidth(s)
	}
	
	doc.Text(x, y, s)
}

func (p *PDFRenderer) renderList(doc *fpdf.Fpdf, cw, x, y, fs float64, l deck.List) {
	if len(l.Li) == 0 {
		return
	}
	
	lspacing := ListSpacing
	if l.Lp > 0 {
		lspacing = l.Lp
	}
	
	for i, item := range l.Li {
		listitemy := y + (float64(i) * fs * lspacing)
		var t string
		
		switch l.Type {
		case "number":
			t = fmt.Sprintf("%d. ", i+1)
		case "bullet":
		default:
			t = "• "
		}
		
		// Use the same font handling as showtext for consistency
		fontname := l.Font
		if fontname == "" {
			fontname = "sans"
		}
		
		p.showtext(doc, x, listitemy, t+item.ListText, fs, fontname, l.Align, "")
	}
}

func (p *PDFRenderer) fontlookup(s string) string {
	if font, ok := p.fontMap[s]; ok {
		return font
	}
	return "Arial"
}

