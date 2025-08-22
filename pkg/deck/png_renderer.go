package deck

import (
	"bytes"
	"fmt"
	"image/png"
	"io"
	"strings"

	"github.com/fogleman/gg"
)

// PNGRenderer handles conversion from decksh DSL to PNG format
type PNGRenderer struct {
	renderer *Renderer
	width    int
	height   int
}

// NewPNGRenderer creates a new PNG renderer
func NewPNGRenderer(width, height float64) *PNGRenderer {
	return &PNGRenderer{
		renderer: NewRenderer(width, height),
		width:    int(width),
		height:   int(height),
	}
}

// DeckshToPNG converts decksh DSL to PNG bytes
func (p *PNGRenderer) DeckshToPNG(dshInput string, opts RenderOptions) ([]byte, error) {
	// First convert decksh to XML
	xmlContent, err := p.renderer.DeckshToXML(dshInput)
	if err != nil {
		return nil, fmt.Errorf("failed to convert decksh to XML: %w", err)
	}

	// Then render XML to PNG
	pngBytes, err := p.XMLToPNG(xmlContent, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to convert XML to PNG: %w", err)
	}

	return pngBytes, nil
}

// XMLToPNG converts deck XML to PNG bytes using gg graphics library
func (p *PNGRenderer) XMLToPNG(xmlInput string, opts RenderOptions) ([]byte, error) {
	// Parse the deck XML using the main renderer
	processedXML := p.renderer.wrapInSlideIfNeeded(xmlInput)
	xmlReader := strings.NewReader(processedXML)
	
	d, err := readDeck(io.NopCloser(xmlReader), p.width, p.height)
	if err != nil {
		return nil, fmt.Errorf("failed to read deck XML: %w", err)
	}

	// Create graphics context
	dc := gg.NewContext(p.width, p.height)

	// Render the first slide (for now)
	if len(d.Slide) > 0 {
		err := p.renderSlideToPNG(dc, d, 0, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to render slide to PNG: %w", err)
		}
	}

	// Get the image and encode to PNG
	img := dc.Image()
	var buf bytes.Buffer
	err = png.Encode(&buf, img)
	if err != nil {
		return nil, fmt.Errorf("failed to encode PNG: %w", err)
	}

	return buf.Bytes(), nil
}

// renderSlideToPNG renders a single slide to PNG using gg
func (p *PNGRenderer) renderSlideToPNG(dc *gg.Context, d Deck, slideIndex int, opts RenderOptions) error {
	if slideIndex < 0 || slideIndex >= len(d.Slide) {
		return fmt.Errorf("slide index %d out of range", slideIndex)
	}

	slide := d.Slide[slideIndex]
	cw, ch := float64(p.width), float64(p.height)

	// Set background
	if slide.Bg != "" {
		dc.SetHexColor(slide.Bg)
	} else {
		dc.SetRGB(1, 1, 1) // white background
	}
	dc.Clear()

	// Set default foreground
	if slide.Fg == "" {
		slide.Fg = "#000000"
	}

	// Process layers in order
	layerList := strings.Split(opts.Layers, ":")
	for _, layer := range layerList {
		switch layer {
		case "rect":
			p.renderRectsToPNG(dc, slide.Rect, cw, ch)
		case "ellipse":
			p.renderEllipsesToPNG(dc, slide.Ellipse, cw, ch)
		case "line":
			p.renderLinesToPNG(dc, slide.Line, cw, ch)
		case "text":
			p.renderTextToPNG(dc, slide.Text, slide.Fg, cw, ch, opts)
		case "list":
			p.renderListsToPNG(dc, slide.List, slide.Fg, cw, ch, opts)
		}
	}

	return nil
}

// Helper methods for rendering different elements to PNG
func (p *PNGRenderer) renderRectsToPNG(dc *gg.Context, rects []Rect, cw, ch float64) {
	for _, rect := range rects {
		x, y := p.pct(rect.Xp, cw), p.pct(100-rect.Yp, ch)
		w, h := p.pct(rect.Wp, cw), p.pct(rect.Hp, ch)

		if rect.Color != "" {
			dc.SetHexColor(rect.Color)
		} else {
			dc.SetRGB(0.5, 0.5, 0.5)
		}

		opacity := rect.Opacity / 100
		if opacity == 0 {
			opacity = 1.0
		}

		// TODO: Apply opacity (gg doesn't have direct opacity support, needs alpha blending)
		dc.DrawRectangle(x-w/2, y-h/2, w, h)
		dc.Fill()
	}
}

func (p *PNGRenderer) renderEllipsesToPNG(dc *gg.Context, ellipses []Ellipse, cw, ch float64) {
	for _, ellipse := range ellipses {
		x, y := p.pct(ellipse.Xp, cw), p.pct(100-ellipse.Yp, ch)
		rx, ry := p.pct(ellipse.Wp, cw)/2, p.pct(ellipse.Hp, ch)/2

		if ellipse.Color != "" {
			dc.SetHexColor(ellipse.Color)
		} else {
			dc.SetRGB(0.5, 0.5, 0.5)
		}

		dc.DrawEllipse(x, y, rx, ry)
		dc.Fill()
	}
}

func (p *PNGRenderer) renderLinesToPNG(dc *gg.Context, lines []Line, cw, ch float64) {
	for _, line := range lines {
		x1, y1 := p.pct(line.Xp1, cw), p.pct(100-line.Yp1, ch)
		x2, y2 := p.pct(line.Xp2, cw), p.pct(100-line.Yp2, ch)

		if line.Color != "" {
			dc.SetHexColor(line.Color)
		} else {
			dc.SetRGB(0.5, 0.5, 0.5)
		}

		sw := line.Sp
		if sw == 0 {
			sw = 2.0
		}
		dc.SetLineWidth(sw)

		dc.DrawLine(x1, y1, x2, y2)
		dc.Stroke()
	}
}

func (p *PNGRenderer) renderTextToPNG(dc *gg.Context, texts []Text, defaultColor string, cw, ch float64, opts RenderOptions) {
	for _, text := range texts {
		x, y := p.pct(text.Xp, cw), p.pct(100-text.Yp, ch)
		fontSize := p.pct(text.Sp, cw)

		color := text.Color
		if color == "" {
			color = defaultColor
		}
		if color != "" {
			dc.SetHexColor(color)
		} else {
			dc.SetRGB(0, 0, 0)
		}

		// Load font - use default if font loading fails
		fontPath := "/System/Library/Fonts/Arial.ttf"
		if text.Font != "" && p.renderer.fontManager != nil {
			if cachedPath, err := p.renderer.GetFontPath(text.Font, 400); err == nil {
				fontPath = cachedPath
			}
		}

		if err := dc.LoadFontFace(fontPath, fontSize); err != nil {
			// Fallback to built-in font if loading fails
			dc.SetFontFace(nil)
		}

		// Handle text alignment
		switch text.Align {
		case "center", "middle", "mid", "c":
			dc.DrawStringAnchored(text.Tdata, x, y, 0.5, 0.5)
		case "right", "end", "e":
			dc.DrawStringAnchored(text.Tdata, x, y, 1.0, 0.5)
		default:
			dc.DrawStringAnchored(text.Tdata, x, y, 0.0, 0.5)
		}
	}
}

func (p *PNGRenderer) renderListsToPNG(dc *gg.Context, lists []List, defaultColor string, cw, ch float64, opts RenderOptions) {
	for _, list := range lists {
		x, y := p.pct(list.Xp, cw), p.pct(100-list.Yp, ch)
		fontSize := p.pct(list.Sp, cw)

		color := list.Color
		if color == "" {
			color = defaultColor
		}
		if color != "" {
			dc.SetHexColor(color)
		} else {
			dc.SetRGB(0, 0, 0)
		}

		// Load font
		fontPath := "/System/Library/Fonts/Arial.ttf"
		if list.Font != "" && p.renderer.fontManager != nil {
			if cachedPath, err := p.renderer.GetFontPath(list.Font, 400); err == nil {
				fontPath = cachedPath
			}
		}

		if err := dc.LoadFontFace(fontPath, fontSize); err != nil {
			dc.SetFontFace(nil)
		}

		lp := list.Lp
		if lp == 0 {
			lp = 2.0
		}
		lineSpacing := lp * fontSize

		for i, item := range list.Li {
			itemY := y + float64(i)*lineSpacing
			text := item.ListText

			if list.Type == "number" {
				text = fmt.Sprintf("%d. %s", i+1, text)
			} else if list.Type == "bullet" {
				text = "• " + text
			}

			dc.DrawStringAnchored(text, x, itemY, 0.0, 0.5)
		}
	}
}

// pct converts percentages to canvas measures
func (p *PNGRenderer) pct(percent, measure float64) float64 {
	return (percent / 100.0) * measure
}

// Simple struct definitions for the decoder (avoiding import cycles)
type Deck struct {
	Slide []Slide `xml:"slide"`
}

type Slide struct {
	Bg         string     `xml:"bg,attr"`
	Fg         string     `xml:"fg,attr"`
	Gradcolor1 string     `xml:"gradcolor1,attr"`
	Gradcolor2 string     `xml:"gradcolor2,attr"`
	Rect       []Rect     `xml:"rect"`
	Ellipse    []Ellipse  `xml:"ellipse"`
	Line       []Line     `xml:"line"`
	Text       []Text     `xml:"text"`
	List       []List     `xml:"list"`
	Image      []Image    `xml:"image"`
}

type Rect struct {
	Xp      float64 `xml:"xp,attr"`
	Yp      float64 `xml:"yp,attr"`
	Wp      float64 `xml:"wp,attr"`
	Hp      float64 `xml:"hp,attr"`
	Color   string  `xml:"color,attr"`
	Opacity float64 `xml:"opacity,attr"`
}

type Ellipse struct {
	Xp      float64 `xml:"xp,attr"`
	Yp      float64 `xml:"yp,attr"`
	Wp      float64 `xml:"wp,attr"`
	Hp      float64 `xml:"hp,attr"`
	Color   string  `xml:"color,attr"`
	Opacity float64 `xml:"opacity,attr"`
}

type Line struct {
	Xp1     float64 `xml:"xp1,attr"`
	Yp1     float64 `xml:"yp1,attr"`
	Xp2     float64 `xml:"xp2,attr"`
	Yp2     float64 `xml:"yp2,attr"`
	Color   string  `xml:"color,attr"`
	Sp      float64 `xml:"sp,attr"`
	Opacity float64 `xml:"opacity,attr"`
}

type Text struct {
	Xp    float64 `xml:"xp,attr"`
	Yp    float64 `xml:"yp,attr"`
	Sp    float64 `xml:"sp,attr"`
	Color string  `xml:"color,attr"`
	Font  string  `xml:"font,attr"`
	Align string  `xml:"align,attr"`
	Tdata string  `xml:",chardata"`
}

type List struct {
	Xp    float64 `xml:"xp,attr"`
	Yp    float64 `xml:"yp,attr"`
	Sp    float64 `xml:"sp,attr"`
	Lp    float64 `xml:"lp,attr"`
	Color string  `xml:"color,attr"`
	Font  string  `xml:"font,attr"`
	Type  string  `xml:"type,attr"`
	Li    []ListItem `xml:"li"`
}

type ListItem struct {
	ListText string `xml:",chardata"`
}

type Image struct {
	Xp     float64 `xml:"xp,attr"`
	Yp     float64 `xml:"yp,attr"`
	Width  int     `xml:"width,attr"`
	Height int     `xml:"height,attr"`
	Scale  float64 `xml:"scale,attr"`
	Name   string  `xml:"name,attr"`
}

// Simple XML decoder to avoid import cycles
func readDeck(r io.ReadCloser, width, height int) (Deck, error) {
	defer r.Close()
	
	// For now, return empty deck - would need full XML parsing implementation
	// This is a simplified version for the PNG renderer
	return Deck{}, fmt.Errorf("XML parsing for PNG not fully implemented - use SVG renderer instead")
}