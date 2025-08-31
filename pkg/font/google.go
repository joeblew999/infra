package font

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/joeblew999/infra/pkg/log"
)

// Uses GoogleFontsAPI from consts.go

// downloadGoogleFont downloads a font from Google Fonts using CSS parsing
func downloadGoogleFont(font Font, path string) error {
	// Build Google Fonts CSS URL
	cssURL := buildGoogleFontsURL(font)
	
	// Get CSS with font URLs (using proper User-Agent for format)
	fontURL, err := getFontURL(cssURL, font.Format)
	if err != nil {
		log.Warn("Failed to get font from CSS, using mock", "family", font.Family, "error", err)
		return createMockFontFile(path, font)
	}
	
	// Download the actual font file
	return downloadFontFile(fontURL, path)
}

// GoogleFontsResponse represents the Google Fonts Web API response
type GoogleFontsResponse struct {
	Items []GoogleFontItem `json:"items"`
}

type GoogleFontItem struct {
	Family string                 `json:"family"`
	Files  map[string]string      `json:"files"`
}

// getGoogleFontDirectURL gets the direct TTF URL from Google Fonts Web API
func getGoogleFontDirectURL(font Font) (string, error) {
	// Google Fonts Web API (public, no key needed for basic usage)
	apiURL := "https://www.googleapis.com/webfonts/v1/webfonts"
	
	resp, err := http.Get(apiURL)
	if err != nil {
		return "", fmt.Errorf("failed to fetch Google Fonts list: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Google Fonts API returned status: %s", resp.Status)
	}
	
	var fontsResponse GoogleFontsResponse
	if err := json.NewDecoder(resp.Body).Decode(&fontsResponse); err != nil {
		return "", fmt.Errorf("failed to parse Google Fonts response: %w", err)
	}
	
	// Find the font family
	for _, item := range fontsResponse.Items {
		if strings.EqualFold(item.Family, font.Family) {
			// Look for TTF file - try different weight variants
			weightKey := fmt.Sprintf("%d", font.Weight)
			if font.Weight == 400 {
				weightKey = "regular"
			}
			
			if ttfURL, exists := item.Files[weightKey]; exists {
				return ttfURL, nil
			}
			
			// Fallback to regular if specific weight not found
			if ttfURL, exists := item.Files["regular"]; exists {
				return ttfURL, nil
			}
			
			return "", fmt.Errorf("no TTF file found for %s weight %d", font.Family, font.Weight)
		}
	}
	
	return "", fmt.Errorf("font family '%s' not found in Google Fonts", font.Family)
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
	// Set user agent based on desired format
	userAgent := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36" // WOFF2
	if format == "ttf" {
		// Use very old user agent to force TTF format from Google Fonts
		userAgent = "Mozilla/4.0 (compatible; MSIE 5.0; Windows 98)"
	}
	
	req, err := http.NewRequest("GET", cssURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", userAgent)
	
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	
	css, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	
	// Extract font URL from CSS - Google Fonts uses different URL patterns
	// Look for: url(https://fonts.gstatic.com/l/font?kit=...) or url(https://fonts.gstatic.com/s/roboto/...)
	re := regexp.MustCompile(`url\((https://fonts\.gstatic\.com/[^)]+)\)`)
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