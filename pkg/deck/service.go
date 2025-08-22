package deck

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/joeblew999/infra/pkg/log"
)

// Service provides infrastructure integration for deck rendering
type Service struct {
	renderer    *Renderer
	options     ServiceOptions
	mu          sync.RWMutex
}

// ServiceOptions configures the deck service
type ServiceOptions struct {
	Width          float64 // Canvas width in points
	Height         float64 // Canvas height in points  
	DefaultFormat  string  // Default output format (svg, png, pdf)
	CacheDir       string  // Directory for caching rendered outputs
	FontCacheDir   string  // Directory for font caching
	EnableFonts    bool    // Enable font management
}

// DefaultServiceOptions returns sensible default service options
func DefaultServiceOptions() ServiceOptions {
	return ServiceOptions{
		Width:          DefaultWidth,
		Height:         DefaultHeight,
		DefaultFormat:  FormatSVG,
		CacheDir:       GetDeckCachePath(),
		FontCacheDir:   GetDeckFontPath(),
		EnableFonts:    true,
	}
}

// NewService creates a new deck service
func NewService(opts ...ServiceOption) *Service {
	options := DefaultServiceOptions()
	
	for _, opt := range opts {
		opt(&options)
	}
	
	service := &Service{
		renderer: NewRenderer(options.Width, options.Height),
		options:  options,
	}
	
	return service
}

// ServiceOption configures the service
type ServiceOption func(*ServiceOptions)

// WithDimensions sets the canvas dimensions
func WithDimensions(width, height float64) ServiceOption {
	return func(opts *ServiceOptions) {
		opts.Width = width
		opts.Height = height
	}
}

// WithFormat sets the default output format
func WithFormat(format string) ServiceOption {
	return func(opts *ServiceOptions) {
		opts.DefaultFormat = format
	}
}

// WithCacheDir sets the cache directory
func WithCacheDir(dir string) ServiceOption {
	return func(opts *ServiceOptions) {
		opts.CacheDir = dir
	}
}

// WithFonts enables or disables font management
func WithFonts(enabled bool) ServiceOption {
	return func(opts *ServiceOptions) {
		opts.EnableFonts = enabled
	}
}

// Start initializes the deck service
func (s *Service) Start() error {
	log.Info("Starting deck rendering service", 
		"width", s.options.Width,
		"height", s.options.Height,
		"format", s.options.DefaultFormat,
		"fonts_enabled", s.options.EnableFonts)
	
	// Ensure cache directories exist
	if err := s.ensureDirectories(); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}
	
	// Load common fonts if font management is enabled
	if s.options.EnableFonts {
		if err := s.loadCommonFonts(); err != nil {
			log.Warn("Failed to load common fonts", "error", err)
			// Don't fail service start for font loading issues
		}
	}
	
	log.Info("Deck service started successfully")
	return nil
}

// Stop gracefully shuts down the deck service
func (s *Service) Stop() error {
	log.Info("Stopping deck rendering service")
	// Cleanup if needed
	return nil
}

// RenderDecksh renders decksh DSL to the specified format
func (s *Service) RenderDecksh(dshInput, format string) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	opts := DefaultRenderOptions()
	
	switch format {
	case FormatSVG, "":
		svgContent, err := s.renderer.DeckshToSVG(dshInput, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to render SVG: %w", err)
		}
		return []byte(svgContent), nil
		
	case FormatXML:
		xmlContent, err := s.renderer.DeckshToXML(dshInput)
		if err != nil {
			return nil, fmt.Errorf("failed to render XML: %w", err)
		}
		return []byte(xmlContent), nil
		
	case FormatPNG:
		// TODO: Implement PNG rendering
		return nil, fmt.Errorf("PNG rendering not yet implemented")
		
	case FormatPDF:
		// TODO: Implement PDF rendering
		return nil, fmt.Errorf("PDF rendering not yet implemented")
		
	default:
		return nil, fmt.Errorf("%s: %s", ErrUnsupportedFormat, format)
	}
}

// RenderDeckshToFile renders decksh DSL and saves to a file
func (s *Service) RenderDeckshToFile(dshInput, outputPath, format string) error {
	content, err := s.RenderDecksh(dshInput, format)
	if err != nil {
		return err
	}
	
	// Ensure output directory exists
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}
	
	if err := os.WriteFile(outputPath, content, 0644); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}
	
	log.Info("Rendered deck to file",
		"format", format,
		"output", outputPath,
		"size", len(content))
	
	return nil
}

// LoadFont loads a font for use in presentations
func (s *Service) LoadFont(family string, weight int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if !s.options.EnableFonts {
		return fmt.Errorf(ErrFontManagementOff)
	}
	
	return s.renderer.LoadFont(family, weight)
}

// ListCachedFonts returns all cached fonts
func (s *Service) ListCachedFonts() []FontInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	if !s.options.EnableFonts {
		return nil
	}
	
	// Convert from font.FontInfo to our FontInfo
	fontInfos := s.renderer.ListCachedFonts()
	result := make([]FontInfo, len(fontInfos))
	
	for i, info := range fontInfos {
		result[i] = FontInfo{
			Family: info.Family,
			Weight: info.Weight,
			Style:  info.Style,
			Format: info.Format,
			Path:   info.Path,
			Size:   info.Size,
			Source: info.Source,
		}
	}
	
	return result
}

// FontInfo represents information about a cached font
type FontInfo struct {
	Family string `json:"family"`
	Weight int    `json:"weight"`
	Style  string `json:"style"`
	Format string `json:"format"`
	Path   string `json:"path"`
	Size   int64  `json:"size"`
	Source string `json:"source"`
}

// GetStats returns service statistics
func (s *Service) GetStats() ServiceStats {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	return ServiceStats{
		Width:         s.options.Width,
		Height:        s.options.Height,
		DefaultFormat: s.options.DefaultFormat,
		FontsEnabled:  s.options.EnableFonts,
		CachedFonts:   len(s.ListCachedFonts()),
	}
}

// ServiceStats represents service statistics
type ServiceStats struct {
	Width         float64 `json:"width"`
	Height        float64 `json:"height"`
	DefaultFormat string  `json:"default_format"`
	FontsEnabled  bool    `json:"fonts_enabled"`
	CachedFonts   int     `json:"cached_fonts"`
}

// ensureDirectories creates necessary directories
func (s *Service) ensureDirectories() error {
	dirs := []string{
		s.options.CacheDir,
	}
	
	if s.options.EnableFonts {
		dirs = append(dirs, s.options.FontCacheDir)
	}
	
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}
	
	return nil
}

// loadCommonFonts preloads commonly used fonts
func (s *Service) loadCommonFonts() error {
	commonFonts := []struct {
		family string
		weight int
	}{
		{"Arial", 400},
		{"Helvetica", 400},
		{"Times", 400},
		{"Courier", 400},
	}
	
	for _, font := range commonFonts {
		if err := s.renderer.LoadFont(font.family, font.weight); err != nil {
			log.Debug("Failed to load font", 
				"family", font.family,
				"weight", font.weight,
				"error", err)
			// Continue with other fonts
		}
	}
	
	return nil
}