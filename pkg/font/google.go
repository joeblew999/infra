package font

import (
	"fmt"
	"os"
)

// Uses GoogleFontsAPI from consts.go

// downloadGoogleFont downloads a font from Google Fonts
func downloadGoogleFont(font Font, path string) error {
	// For now, create a mock font file for testing
	// In production, we'd parse the CSS and download the actual font files
	return createMockFontFile(path, font)
}

// createMockFontFile creates a mock font file for testing
func createMockFontFile(path string, font Font) error {
	// Create mock font data (WOFF2 header + minimal content)
	mockData := []byte("mock font data for " + font.Family + " " + fmt.Sprintf("%d", font.Weight))
	return os.WriteFile(path, mockData, 0644)
}


// ListGoogleFonts returns available Google Fonts
func ListGoogleFonts() []string {
	return DefaultFonts
}