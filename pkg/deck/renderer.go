// Package deck provides functionality to render deck presentations from decksh DSL to various formats
package deck

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/ajstarks/deck"
	"github.com/ajstarks/decksh"
	svg "github.com/ajstarks/svgo/float"
	
	"github.com/joeblew999/infra/pkg/font"
)

// Renderer handles the conversion from decksh DSL to various output formats
type Renderer struct {
	Width       float64        // Canvas width in points
	Height      float64        // Canvas height in points
	fontManager *font.Manager  // Font manager for custom fonts
}

// NewRenderer creates a new renderer with the specified canvas dimensions
func NewRenderer(width, height float64) *Renderer {
	return &Renderer{
		Width:       width,
		Height:      height,
		fontManager: font.NewManager(),
	}
}

// NewDefaultRenderer creates a renderer with standard Letter size dimensions
func NewDefaultRenderer() *Renderer {
	return NewRenderer(DefaultWidth, DefaultHeight)
}

// RenderOptions holds configuration for rendering
type RenderOptions struct {
	GridPercent float64 // Grid percentage (0 = no grid)
	Title       string  // Document title
	Layers      string  // Drawing order layers
	FontFamily  string  // Default font family
	FontWeight  int     // Default font weight
}

// DefaultRenderOptions returns sensible default rendering options
func DefaultRenderOptions() RenderOptions {
	return RenderOptions{
		GridPercent: 0,
		Title:       "",
		Layers:      DefaultLayers,
		FontFamily:  DefaultFontFamily,
		FontWeight:  DefaultFontWeight,
	}
}

// DeckshToXML converts decksh DSL input to deck XML format
func (r *Renderer) DeckshToXML(dshInput string) (string, error) {
	// Create input reader from the decksh DSL
	input := strings.NewReader(dshInput)

	// Create output buffer for XML
	var xmlOutput bytes.Buffer

	// Process decksh DSL to generate XML
	err := decksh.Process(&xmlOutput, input)
	if err != nil {
		return "", fmt.Errorf("failed to process decksh: %w", err)
	}

	return xmlOutput.String(), nil
}

// XMLToSVG converts deck XML to SVG format
func (r *Renderer) XMLToSVG(xmlInput string, opts RenderOptions) (string, error) {
	// Check if the XML needs to be wrapped in a slide element
	processedXML := r.wrapInSlideIfNeeded(xmlInput)

	// Parse the deck XML
	xmlReader := strings.NewReader(processedXML)
	d, err := deck.ReadDeck(io.NopCloser(xmlReader), int(r.Width), int(r.Height))
	if err != nil {
		return "", fmt.Errorf("failed to read deck XML: %w", err)
	}

	// Ensure canvas dimensions are set
	d.Canvas.Width = int(r.Width)
	d.Canvas.Height = int(r.Height)

	// Create SVG output buffer
	var svgOutput bytes.Buffer
	svgDoc := svg.New(&svgOutput)

	// Render all slides (for now, just render the first slide)
	if len(d.Slide) > 0 {
		err := r.renderSlideToSVG(svgDoc, d, 0, opts)
		if err != nil {
			return "", fmt.Errorf("failed to render slide to SVG: %w", err)
		}
	}

	return svgOutput.String(), nil
}

// DeckshToSVG converts decksh DSL directly to SVG (convenience method)
func (r *Renderer) DeckshToSVG(dshInput string, opts RenderOptions) (string, error) {
	// First convert decksh to XML
	xmlContent, err := r.DeckshToXML(dshInput)
	if err != nil {
		return "", fmt.Errorf("decksh to XML conversion failed: %w", err)
	}

	// Then convert XML to SVG
	svgContent, err := r.XMLToSVG(xmlContent, opts)
	if err != nil {
		return "", fmt.Errorf("XML to SVG conversion failed: %w", err)
	}

	return svgContent, nil
}

// LoadFont loads a font for use in presentations
func (r *Renderer) LoadFont(family string, weight int) error {
	if r.fontManager == nil {
		return fmt.Errorf("font manager not initialized")
	}
	return r.fontManager.Cache(family, weight)
}

// GetFontPath returns the path to a cached font
func (r *Renderer) GetFontPath(family string, weight int) (string, error) {
	if r.fontManager == nil {
		return "", fmt.Errorf("font manager not initialized")
	}
	return r.fontManager.Get(family, weight)
}

// ListCachedFonts returns all cached fonts
func (r *Renderer) ListCachedFonts() []font.FontInfo {
	if r.fontManager == nil {
		return nil
	}
	return r.fontManager.List()
}

// renderSlideToSVG renders a single slide to SVG
func (r *Renderer) renderSlideToSVG(doc *svg.SVG, d deck.Deck, slideIndex int, opts RenderOptions) error {
	if slideIndex < 0 || slideIndex >= len(d.Slide) {
		return fmt.Errorf("%s: %d", ErrSlideIndexOutOfRange, slideIndex)
	}

	slide := d.Slide[slideIndex]
	cw, ch := r.Width, r.Height

	// Start SVG document
	doc.Start(cw, ch)

	// Add title if specified
	if opts.Title != "" {
		doc.Title(fmt.Sprintf("%s: Slide %d", opts.Title, slideIndex+1))
	}

	// Set background if specified
	if slide.Bg != "" {
		doc.Rect(0, 0, cw, ch, fmt.Sprintf("fill:%s", slide.Bg))
	}

	// Set gradient background if specified
	if slide.Gradcolor1 != "" && slide.Gradcolor2 != "" {
		oc := []svg.Offcolor{
			{Offset: 0, Color: slide.Gradcolor1, Opacity: 1.0},
			{Offset: 100, Color: slide.Gradcolor2, Opacity: 1.0},
		}
		doc.Def()
		doc.LinearGradient("slidegrad", 0, 0, 0, 100, oc)
		doc.DefEnd()
		doc.Rect(0, 0, cw, ch, "fill:url(#slidegrad)")
	}

	// Set default foreground
	if slide.Fg == "" {
		slide.Fg = "black"
	}

	// Process layers in order
	layerList := strings.Split(opts.Layers, ":")
	for _, layer := range layerList {
		switch layer {
		case "rect":
			r.renderRects(doc, slide.Rect, cw, ch)
		case "ellipse":
			r.renderEllipses(doc, slide.Ellipse, cw, ch)
		case "line":
			r.renderLines(doc, slide.Line, cw, ch)
		case "text":
			r.renderTexts(doc, slide.Text, slide.Fg, cw, ch, opts)
		case "list":
			r.renderLists(doc, slide.List, slide.Fg, cw, ch, opts)
		case "image":
			r.renderImages(doc, slide.Image, cw, ch)
		}
	}

	// Add grid if specified
	if opts.GridPercent > 0 {
		r.renderGrid(doc, cw, ch, slide.Fg, opts.GridPercent)
	}

	doc.End()
	return nil
}

// Helper methods for rendering different elements
func (r *Renderer) renderRects(doc *svg.SVG, rects []deck.Rect, cw, ch float64) {
	for _, rect := range rects {
		x, y := r.pct(rect.Xp, cw), r.pct(100-rect.Yp, ch)
		w, h := r.pct(rect.Wp, cw), r.pct(rect.Hp, ch)

		color := rect.Color
		if color == "" {
			color = "rgb(127,127,127)"
		}

		opacity := rect.Opacity
		if opacity == 0 {
			opacity = 1.0
		}

		style := fmt.Sprintf("fill:%s;fill-opacity:%.2f", color, opacity/100)
		doc.Rect(x-(w/2), y-(h/2), w, h, style)
	}
}

func (r *Renderer) renderEllipses(doc *svg.SVG, ellipses []deck.Ellipse, cw, ch float64) {
	for _, ellipse := range ellipses {
		x, y := r.pct(ellipse.Xp, cw), r.pct(100-ellipse.Yp, ch)
		w, h := r.pct(ellipse.Wp, cw), r.pct(ellipse.Hp, ch)

		color := ellipse.Color
		if color == "" {
			color = "rgb(127,127,127)"
		}

		opacity := ellipse.Opacity
		if opacity == 0 {
			opacity = 1.0
		}

		style := fmt.Sprintf("fill:%s;fill-opacity:%.2f", color, opacity/100)
		doc.Ellipse(x, y, w/2, h/2, style)
	}
}

func (r *Renderer) renderLines(doc *svg.SVG, lines []deck.Line, cw, ch float64) {
	for _, line := range lines {
		x1, y1 := r.pct(line.Xp1, cw), r.pct(100-line.Yp1, ch)
		x2, y2 := r.pct(line.Xp2, cw), r.pct(100-line.Yp2, ch)

		color := line.Color
		if color == "" {
			color = "rgb(127,127,127)"
		}

		sw := line.Sp
		if sw == 0 {
			sw = 2.0
		}

		opacity := line.Opacity
		if opacity == 0 {
			opacity = 1.0
		}

		style := fmt.Sprintf("stroke:%s;stroke-width:%.2f;stroke-opacity:%.2f", color, sw, opacity/100)
		doc.Line(x1, y1, x2, y2, style)
	}
}

func (r *Renderer) renderTexts(doc *svg.SVG, texts []deck.Text, defaultColor string, cw, ch float64, opts RenderOptions) {
	for _, text := range texts {
		x, y := r.pct(text.Xp, cw), r.pct(100-text.Yp, ch)
		fs := r.pct(text.Sp, cw)

		color := text.Color
		if color == "" {
			color = defaultColor
		}

		font := text.Font
		if font == "" {
			font = r.getFontFamily(opts.FontFamily)
		}

		align := "start"
		switch text.Align {
		case "center", "middle", "mid", "c":
			align = "middle"
		case "right", "end", "e":
			align = "end"
		}

		style := fmt.Sprintf("fill:%s;font-size:%.2fpx;font-family:%s;text-anchor:%s", color, fs, font, align)
		doc.Text(x, y, text.Tdata, style)
	}
}

func (r *Renderer) renderLists(doc *svg.SVG, lists []deck.List, defaultColor string, cw, ch float64, opts RenderOptions) {
	for _, list := range lists {
		x, y := r.pct(list.Xp, cw), r.pct(100-list.Yp, ch)
		fs := r.pct(list.Sp, cw)

		color := list.Color
		if color == "" {
			color = defaultColor
		}

		font := list.Font
		if font == "" {
			font = r.getFontFamily(opts.FontFamily)
		}

		lp := list.Lp
		if lp == 0 {
			lp = 2.0 // default list spacing
		}
		ls := lp * fs

		for i, item := range list.Li {
			itemY := y + float64(i)*ls
			text := item.ListText

			if list.Type == "number" {
				text = fmt.Sprintf("%d. %s", i+1, text)
			} else if list.Type == "bullet" {
				text = "• " + text
			}

			style := fmt.Sprintf("fill:%s;font-size:%.2fpx;font-family:%s", color, fs, font)
			doc.Text(x, itemY, text, style)
		}
	}
}

func (r *Renderer) renderImages(doc *svg.SVG, images []deck.Image, cw, ch float64) {
	for _, img := range images {
		x, y := r.pct(img.Xp, cw), r.pct(100-img.Yp, ch)

		// For now, just render a placeholder rectangle for images
		w, h := float64(img.Width), float64(img.Height)
		if img.Scale > 0 {
			w *= img.Scale / 100
			h *= img.Scale / 100
		}

		style := "fill:rgb(240,240,240);stroke:rgb(200,200,200);stroke-width:1"
		doc.Rect(x-w/2, y-h/2, w, h, style)

		// Add image name as text
		if img.Name != "" {
			textStyle := "fill:rgb(100,100,100);font-size:12px;font-family:sans-serif;text-anchor:middle"
			doc.Text(x, y, img.Name, textStyle)
		}
	}
}

func (r *Renderer) renderGrid(doc *svg.SVG, cw, ch float64, color string, percent float64) {
	pw := cw * (percent / 100)
	ph := ch * (percent / 100)

	style := fmt.Sprintf("stroke:%s;stroke-width:0.5", color)

	// Vertical lines
	for x := 0.0; x <= cw; x += pw {
		doc.Line(x, 0, x, ch, style)
	}

	// Horizontal lines
	for y := 0.0; y <= ch; y += ph {
		doc.Line(0, y, cw, y, style)
	}
}

// getFontFamily returns the appropriate font family, checking for cached fonts
func (r *Renderer) getFontFamily(defaultFont string) string {
	// Check if we have the font cached
	if r.fontManager != nil && r.fontManager.Available(defaultFont, 400) {
		return defaultFont
	}
	
	// Fall back to email-safe fonts for broader compatibility
	emailSafeFonts := font.GetEmailSafeFonts()
	for _, safeFont := range emailSafeFonts {
		if strings.Contains(strings.ToLower(defaultFont), strings.ToLower(safeFont)) {
			return safeFont
		}
	}
	
	// Default fallback
	return "sans-serif"
}

// wrapInSlideIfNeeded wraps the XML content in a slide element if it's not already present
func (r *Renderer) wrapInSlideIfNeeded(xmlInput string) string {
	// If the XML already has slide elements or is empty, return as-is
	if strings.Contains(xmlInput, "<slide") || strings.TrimSpace(xmlInput) == "" {
		return xmlInput
	}

	// If it has a deck element but no slides, wrap content in a slide
	if strings.Contains(xmlInput, "<deck>") {
		// Find the content between <deck> and </deck>
		start := strings.Index(xmlInput, "<deck>") + 6
		end := strings.Index(xmlInput, "</deck>")
		if start > 6 && end > start {
			content := strings.TrimSpace(xmlInput[start:end])
			if content != "" {
				// Wrap the content in a slide
				return fmt.Sprintf(`<deck>
  <canvas width="%d" height="%d"/>
  <slide>
    %s
  </slide>
</deck>`, int(r.Width), int(r.Height), content)
			}
		}
		return xmlInput
	}

	// If it's just bare elements, wrap them in a complete deck structure
	content := strings.TrimSpace(xmlInput)
	if content != "" {
		return fmt.Sprintf(`<deck>
  <canvas width="%d" height="%d"/>
  <slide>
    %s
  </slide>
</deck>`, int(r.Width), int(r.Height), content)
	}

	return xmlInput
}

// pct converts percentages to canvas measures
func (r *Renderer) pct(p, m float64) float64 {
	return (p / 100.0) * m
}