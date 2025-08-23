package deck

import (
	"os"
	"strings"
	"testing"
	"path/filepath"
)

func TestNewService(t *testing.T) {
	service := NewService()
	
	if service == nil {
		t.Fatal("NewService returned nil")
	}
	
	if service.renderer == nil {
		t.Error("Service should have a renderer")
	}
	
	stats := service.GetStats()
	if stats.Width == 0 || stats.Height == 0 {
		t.Error("Service should have non-zero dimensions")
	}
}

func TestServiceOptions(t *testing.T) {
	service := NewService(
		WithDimensions(800, 600),
		WithFormat("png"),
		WithCacheDir("./test-cache"),
		WithFonts(false),
	)
	
	stats := service.GetStats()
	
	if stats.Width != 800 {
		t.Errorf("Expected width 800, got %f", stats.Width)
	}
	if stats.Height != 600 {
		t.Errorf("Expected height 600, got %f", stats.Height)
	}
	if stats.DefaultFormat != "png" {
		t.Errorf("Expected format png, got %s", stats.DefaultFormat)
	}
	if stats.FontsEnabled {
		t.Error("Fonts should be disabled")
	}
}

func TestServiceStartStop(t *testing.T) {
	// Use temporary directory for testing
	tmpDir := t.TempDir()
	
	service := NewService(
		WithCacheDir(filepath.Join(tmpDir, "cache")),
		WithFonts(true),
	)
	
	err := service.Start()
	if err != nil {
		t.Fatalf("Service start failed: %v", err)
	}
	
	// Check that cache directory was created
	cacheDir := filepath.Join(tmpDir, "cache")
	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		t.Error("Cache directory should have been created")
	}
	
	err = service.Stop()
	if err != nil {
		t.Errorf("Service stop failed: %v", err)
	}
}

func TestRenderDeckshSVG(t *testing.T) {
	service := NewService()
	
	dshInput := `deck
slide "white" "black"
text "Hello World" 50 50 3
rect 25 25 20 15 "blue" 50
eslide
edeck`

	// Test SVG rendering
	result, err := service.RenderDecksh(dshInput, "svg")
	if err != nil {
		t.Fatalf("RenderDecksh SVG failed: %v", err)
	}
	
	if len(result) == 0 {
		t.Error("Expected non-empty SVG output")
	}
	
	svgContent := string(result)
	if !strings.Contains(svgContent, "<svg") {
		t.Error("Expected SVG output to contain <svg element")
	}
	if !strings.Contains(svgContent, "</svg>") {
		t.Error("Expected SVG output to contain closing </svg> tag")
	}
}

func TestRenderDeckshXML(t *testing.T) {
	service := NewService()
	
	dshInput := `deck
slide "white" "black"
text "Hello XML" 50 50 3
eslide
edeck`

	// Test XML rendering
	result, err := service.RenderDecksh(dshInput, "xml")
	if err != nil {
		t.Fatalf("RenderDecksh XML failed: %v", err)
	}
	
	if len(result) == 0 {
		t.Error("Expected non-empty XML output")
	}
	
	xmlContent := string(result)
	if !strings.Contains(xmlContent, "<deck") || !strings.Contains(xmlContent, "<slide") {
		t.Error("Expected XML output to contain deck and slide elements")
	}
}

func TestRenderDeckshUnsupportedFormat(t *testing.T) {
	service := NewService()
	
	dshInput := `deck
slide "white" "black"  
text "Test" 50 50 3
eslide
edeck`

	// Test unsupported format
	_, err := service.RenderDecksh(dshInput, "unsupported")
	if err == nil {
		t.Error("Expected error for unsupported format")
	}
	
	if !strings.Contains(err.Error(), "unsupported format") {
		t.Errorf("Expected 'unsupported format' error, got: %v", err)
	}
}

func TestRenderDeckshToFile(t *testing.T) {
	tmpDir := t.TempDir()
	service := NewService()
	
	dshInput := `deck
slide "white" "black"
text "File Output Test" 50 50 3
eslide
edeck`

	outputPath := filepath.Join(tmpDir, "test-output.svg")
	
	err := service.RenderDeckshToFile(dshInput, outputPath, "svg")
	if err != nil {
		t.Fatalf("RenderDeckshToFile failed: %v", err)
	}
	
	// Check file was created
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Error("Output file should have been created")
	}
	
	// Check file contents
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}
	
	if len(content) == 0 {
		t.Error("Output file should not be empty")
	}
	
	if !strings.Contains(string(content), "<svg") {
		t.Error("Output file should contain SVG content")
	}
}

func TestFontManagement(t *testing.T) {
	service := NewService(WithFonts(true))
	
	err := service.Start()
	if err != nil {
		t.Fatalf("Service start failed: %v", err)
	}
	defer service.Stop()
	
	// Test font loading (may fail with mock fonts, which is OK)
	err = service.LoadFont("Arial", 400)
	if err != nil {
		t.Logf("Font loading returned error (expected with mocks): %v", err)
	}
	
	// Test listing cached fonts
	fonts := service.ListCachedFonts()
	t.Logf("Cached fonts: %d", len(fonts))
	
	// Test stats with font information
	stats := service.GetStats()
	if !stats.FontsEnabled {
		t.Error("Fonts should be enabled in stats")
	}
}

func TestFontManagementDisabled(t *testing.T) {
	service := NewService(WithFonts(false))
	
	err := service.Start()
	if err != nil {
		t.Fatalf("Service start failed: %v", err)
	}
	defer service.Stop()
	
	// Test font loading should fail when disabled
	err = service.LoadFont("Arial", 400)
	if err == nil {
		t.Error("Font loading should fail when fonts are disabled")
	}
	
	if !strings.Contains(err.Error(), "font management is disabled") {
		t.Errorf("Expected 'font management is disabled' error, got: %v", err)
	}
	
	// Test listing cached fonts returns empty when disabled
	fonts := service.ListCachedFonts()
	if fonts != nil {
		t.Error("ListCachedFonts should return nil when fonts are disabled")
	}
}

func TestGetStats(t *testing.T) {
	service := NewService(
		WithDimensions(1000, 750),
		WithFormat("pdf"),
		WithFonts(true),
	)
	
	stats := service.GetStats()
	
	if stats.Width != 1000 {
		t.Errorf("Expected width 1000, got %f", stats.Width)
	}
	if stats.Height != 750 {
		t.Errorf("Expected height 750, got %f", stats.Height)
	}
	if stats.DefaultFormat != "pdf" {
		t.Errorf("Expected format pdf, got %s", stats.DefaultFormat)
	}
	if !stats.FontsEnabled {
		t.Error("Fonts should be enabled in stats")
	}
}