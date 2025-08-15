package font

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/joeblew999/infra/pkg/config"
	"github.com/joeblew999/infra/pkg/log"
)

// Font represents a font with family, weight, and style
type Font struct {
	Family string
	Weight int    // 100, 200, 300, 400, 500, 600, 700, 800, 900
	Style  string // normal, italic
	Format string // woff2, woff, ttf
}

// FontInfo contains metadata about a cached font
type FontInfo struct {
	Font
	Path    string
	Size    int64
	Version string
	Source  string // "google", "local"
}

// Manager handles font operations
type Manager struct {
	cacheDir string
	registry *Registry
}

// NewManager creates a new font manager
func NewManager() *Manager {
	return &Manager{
		cacheDir: config.GetFontPath(),
		registry: NewRegistry(),
	}
}

// Get returns the path to a cached font, downloading if necessary
func (m *Manager) Get(family string, weight int) (string, error) {
	font := Font{
		Family: family,
		Weight: weight,
		Style:  DefaultFontStyle,
		Format: DefaultFontFormat,
	}

	// Check if font is already cached
	if path, exists := m.registry.GetPath(font); exists {
		return path, nil
	}

	// Download and cache the font
	return m.cacheFont(font)
}

// List returns all available cached fonts
func (m *Manager) List() []FontInfo {
	return m.registry.List()
}

// Cache downloads and caches a font
func (m *Manager) Cache(family string, weight int) error {
	font := Font{
		Family: family,
		Weight: weight,
		Style:  DefaultFontStyle,
		Format: DefaultFontFormat,
	}

	_, err := m.cacheFont(font)
	return err
}

// Available checks if a font is cached
func (m *Manager) Available(family string, weight int) bool {
	font := Font{
		Family: family,
		Weight: weight,
		Style:  DefaultFontStyle,
		Format: DefaultFontFormat,
	}
	_, exists := m.registry.GetPath(font)
	return exists
}

// cacheFont downloads and caches a font
func (m *Manager) cacheFont(font Font) (string, error) {
	// Ensure cache directory exists
	familyDir := config.GetFontPathForFamily(font.Family)
	if err := ensureDir(familyDir); err != nil {
		return "", fmt.Errorf("failed to create font directory: %w", err)
	}

	// Generate filename
	filename := fmt.Sprintf("%d.%s", font.Weight, DefaultFontFormat)
	path := filepath.Join(familyDir, filename)

	// Download from Google Fonts
	if err := m.downloadGoogleFont(font, path); err != nil {
		return "", fmt.Errorf("failed to download font: %w", err)
	}

	// Register in registry
	info := FontInfo{
		Font:    font,
		Path:    path,
		Source:  "google",
		Version: "latest",
	}

	if err := m.registry.Add(info); err != nil {
		log.Warn("Failed to register font", "error", err)
	}

	return path, nil
}

// downloadGoogleFont downloads a font from Google Fonts
func (m *Manager) downloadGoogleFont(font Font, path string) error {
	// This will be implemented in google.go
	return downloadGoogleFont(font, path)
}

// ensureDir creates a directory if it doesn't exist
func ensureDir(path string) error {
	return os.MkdirAll(path, 0755)
}
