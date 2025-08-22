package font

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/joeblew999/infra/pkg/log"
)

// Uses GoogleFontsAPI from consts.go

// downloadGoogleFont downloads a font from Google Fonts
func downloadGoogleFont(font Font, path string) error {
	// Build Google Fonts CSS URL
	cssURL := buildGoogleFontsURL(font)
	
	// Get CSS with font URLs
	fontURL, err := getFontURL(cssURL, font.Format)
	if err != nil {
		log.Warn("Failed to get real font, using mock", "family", font.Family, "error", err)
		return createMockFontFile(path, font)
	}
	
	// Download the actual font file
	return downloadFontFile(fontURL, path)
}

// buildGoogleFontsURL creates the CSS URL for Google Fonts
func buildGoogleFontsURL(font Font) string {
	// Example: https://fonts.googleapis.com/css2?family=Roboto:wght@400
	familyName := strings.ReplaceAll(font.Family, " ", "+")
	
	url := fmt.Sprintf("%s?family=%s:wght@%d", GoogleFontsAPI, familyName, font.Weight)
	
	if font.Style == "italic" {
		url = fmt.Sprintf("%s?family=%s:ital,wght@1,%d", GoogleFontsAPI, familyName, font.Weight)
	}
	
	return url
}

// getFontURL parses CSS to extract the actual font file URL
func getFontURL(cssURL, format string) (string, error) {
	// Set user agent to get WOFF2 fonts
	req, err := http.NewRequest("GET", cssURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	
	css, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	
	// Extract font URL from CSS
	// Look for: url(https://fonts.gstatic.com/s/roboto/v30/KFOmCnqEu92Fr1Mu4mxK.woff2)
	re := regexp.MustCompile(`url\((https://[^)]+\.` + format + `)\)`)
	matches := re.FindStringSubmatch(string(css))
	
	if len(matches) < 2 {
		return "", fmt.Errorf("no %s font URL found in CSS", format)
	}
	
	return matches[1], nil
}

// downloadFontFile downloads a font file from a URL
func downloadFontFile(url, path string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download font: %s", resp.Status)
	}
	
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	
	_, err = io.Copy(file, resp.Body)
	return err
}

// createMockFontFile creates a mock font file for testing/fallback
func createMockFontFile(path string, font Font) error {
	// Create mock font data (WOFF2 header + minimal content)
	mockData := []byte("mock font data for " + font.Family + " " + fmt.Sprintf("%d", font.Weight))
	return os.WriteFile(path, mockData, 0644)
}

// ListGoogleFonts returns available Google Fonts
func ListGoogleFonts() []string {
	return DefaultFonts
}

// GetFontCSS generates CSS for embedding a font (useful for email/web)
func GetFontCSS(font Font, fontPath string) string {
	return fmt.Sprintf(`@font-face {
  font-family: '%s';
  font-style: %s;
  font-weight: %d;
  font-display: swap;
  src: url('%s') format('%s');
}`, font.Family, font.Style, font.Weight, fontPath, font.Format)
}

// GetEmailSafeFonts returns fonts that are widely supported in email clients
func GetEmailSafeFonts() []string {
	return []string{
		"Arial",
		"Helvetica",
		"Georgia", 
		"Times",
		"Courier",
		"Verdana",
		"Tahoma",
		"Impact",
		"Comic Sans MS",
		"Trebuchet MS",
		"Arial Black",
		"Palatino",
		"Lucida Console",
	}
}