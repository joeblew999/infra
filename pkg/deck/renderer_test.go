package deck

import (
	"strings"
	"testing"
)

func TestNewRenderer(t *testing.T) {
	width, height := 800.0, 600.0
	r := NewRenderer(width, height)

	if r.Width != width {
		t.Errorf("Expected width %f, got %f", width, r.Width)
	}
	if r.Height != height {
		t.Errorf("Expected height %f, got %f", height, r.Height)
	}
	if r.fontManager == nil {
		t.Error("Font manager should be initialized")
	}
}

func TestNewDefaultRenderer(t *testing.T) {
	r := NewDefaultRenderer()

	// Letter size dimensions
	expectedWidth := 792.0
	expectedHeight := 612.0

	if r.Width != expectedWidth {
		t.Errorf("Expected default width %f, got %f", expectedWidth, r.Width)
	}
	if r.Height != expectedHeight {
		t.Errorf("Expected default height %f, got %f", expectedHeight, r.Height)
	}
}

func TestDefaultRenderOptions(t *testing.T) {
	opts := DefaultRenderOptions()

	if opts.GridPercent != 0 {
		t.Errorf("Expected default GridPercent 0, got %f", opts.GridPercent)
	}
	if opts.Title != "" {
		t.Errorf("Expected default Title empty, got %s", opts.Title)
	}
	if opts.Layers == "" {
		t.Error("Expected default Layers to be non-empty")
	}
	if opts.FontFamily != "Arial" {
		t.Errorf("Expected default FontFamily Arial, got %s", opts.FontFamily)
	}
	if opts.FontWeight != 400 {
		t.Errorf("Expected default FontWeight 400, got %d", opts.FontWeight)
	}
}

func TestFontIntegration(t *testing.T) {
	r := NewDefaultRenderer()

	// Test font loading
	err := r.LoadFont("Arial", 400)
	// Error is acceptable - might fallback to mock fonts
	if err != nil {
		t.Logf("Font loading returned error (expected): %v", err)
	}

	// Test font path retrieval
	path, err := r.GetFontPath("Arial", 400)
	if err != nil {
		t.Logf("Font path retrieval error (expected): %v", err)
	} else {
		t.Logf("Font path: %s", path)
	}

	// Test cached fonts listing
	cachedFonts := r.ListCachedFonts()
	t.Logf("Cached fonts: %d", len(cachedFonts))
}

func TestBasicSVGGeneration(t *testing.T) {
	r := NewDefaultRenderer()
	opts := DefaultRenderOptions()

	// Simple deck XML input for testing
	xmlInput := `<?xml version="1.0"?>
<deck>
	<canvas width="792" height="612"/>
	<slide bg="white" fg="black">
		<text xp="50" yp="50" sp="3" color="black">Hello World</text>
	</slide>
</deck>`

	svg, err := r.XMLToSVG(xmlInput, opts)
	if err != nil {
		t.Fatalf("XMLToSVG failed: %v", err)
	}

	// Basic validation
	if svg == "" {
		t.Error("Expected non-empty SVG output")
	}

	// Should contain SVG elements
	if !strings.Contains(svg, "<svg") {
		t.Error("Expected SVG to contain <svg element")
	}
	if !strings.Contains(svg, "</svg>") {
		t.Error("Expected SVG to contain closing </svg> tag")
	}
}

func TestDeckshToXML(t *testing.T) {
	r := NewDefaultRenderer()

	// Content from example decksh
	dshInput := `deck
slide "white" "black"
text "Hello Deck World!" 50 50 3
circle 25 25 10 "red" 0.5
rect 75 75 20 15 "gray" 0.3
text "My First Deck" 50 80 2
eslide
edeck`

	xml, err := r.DeckshToXML(dshInput)
	if err != nil {
		t.Fatalf("DeckshToXML failed: %v", err)
	}

	if xml == "" {
		t.Error("XML output is empty")
	}

	t.Logf("Generated XML:\n%s", xml)
}

func TestDeckshToSVGIntegration(t *testing.T) {
	r := NewDefaultRenderer()
	opts := DefaultRenderOptions()
	opts.Title = "Integration Test Deck"
	opts.FontFamily = "Arial"

	// Proper decksh syntax
	dshInput := `deck
slide "white" "black"
text "Hello Deck World!" 50 50 3
circle 25 25 10 "red" 50
rect 75 75 20 15 "gray" 30
text "My First Deck" 50 80 2
eslide
edeck`

	// Test full pipeline: decksh -> XML -> SVG
	svg, err := r.DeckshToSVG(dshInput, opts)
	if err != nil {
		t.Fatalf("DeckshToSVG integration test failed: %v", err)
	}

	// Validate SVG structure
	if !strings.Contains(svg, "<svg") {
		t.Error("Expected SVG output to contain <svg element")
	}
	if !strings.Contains(svg, "</svg>") {
		t.Error("Expected SVG output to contain closing </svg> tag")
	}

	// Should have substantial content for a complete slide
	if len(svg) < 200 {
		t.Errorf("Expected substantial SVG output, got %d characters", len(svg))
	}

	t.Logf("Generated SVG length: %d characters", len(svg))
}

func TestPctConversion(t *testing.T) {
	r := NewDefaultRenderer()

	// Test percentage conversion
	result := r.pct(50, 100)
	expected := 50.0
	if result != expected {
		t.Errorf("Expected pct(50, 100) = %f, got %f", expected, result)
	}

	result = r.pct(25, 200)
	expected = 50.0
	if result != expected {
		t.Errorf("Expected pct(25, 200) = %f, got %f", expected, result)
	}
}

func TestGetFontFamily(t *testing.T) {
	r := NewDefaultRenderer()

	// Test default font family handling
	fontFamily := r.getFontFamily("Arial")
	if fontFamily == "" {
		t.Error("getFontFamily should return non-empty string")
	}

	fontFamily = r.getFontFamily("NonExistentFont")
	if fontFamily == "" {
		t.Error("getFontFamily should return fallback for non-existent fonts")
	}

	t.Logf("Font family for 'NonExistentFont': %s", fontFamily)
}