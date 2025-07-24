package docs

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/joeblew999/infra/docs"
)

// NavItem represents a navigation item
type NavItem struct {
	Title string
	Path  string
}

// Service handles documentation serving from both filesystem and embedded sources
type Service struct {
	devMode bool
	docsDir string
	nav     []NavItem
}

// New creates a new docs service
func New(devMode bool, docsDir string) *Service {
	s := &Service{
		devMode: devMode,
		docsDir: docsDir,
	}
	s.buildNavigation()
	return s
}

// ReadFile reads a document file from either filesystem or embedded source
func (s *Service) ReadFile(filePath string) ([]byte, error) {
	// Default to main roadmap if no specific file requested
	if filePath == "" {
		filePath = "roadmap/ROADMAP.md"
	}

	if s.devMode {
		fullPath := filepath.Join(s.docsDir, filePath)
		return os.ReadFile(fullPath)
	}

	// Read directly from embedded filesystem
	content, err := fs.ReadFile(docs.EmbeddedFS, filePath)
	if err != nil {
		return nil, fmt.Errorf("embedded file not found: %s", filePath)
	}
	return content, nil
}

// GetNavigation returns the navigation items
func (s *Service) GetNavigation() []NavItem {
	return s.nav
}

// buildNavigation scans the filesystem and builds navigation
func (s *Service) buildNavigation() {
	var navItems []NavItem

	if s.devMode {
		s.scanDirectory(s.docsDir, &navItems)
	} else {
		s.scanEmbeddedFS(&navItems)
	}

	s.nav = navItems
}

// scanDirectory scans a directory on disk for markdown files
func (s *Service) scanDirectory(dir string, navItems *[]NavItem) {
	filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() && strings.HasSuffix(path, ".md") {
			relPath, _ := filepath.Rel(s.docsDir, path)
			relPath = filepath.ToSlash(relPath) // Convert to forward slashes for URLs
			title := s.extractTitleFromFile(path, true)
			*navItems = append(*navItems, NavItem{Title: title, Path: relPath})
		}
		return nil
	})
}

// scanEmbeddedFS scans the embedded filesystem for markdown files
func (s *Service) scanEmbeddedFS(navItems *[]NavItem) {
	fs.WalkDir(docs.EmbeddedFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() && strings.HasSuffix(path, ".md") {
			title := s.extractTitleFromFile(path, false)
			*navItems = append(*navItems, NavItem{Title: title, Path: path})
		}
		return nil
	})
}

// extractTitleFromFile extracts the title from a markdown file
func (s *Service) extractTitleFromFile(filePath string, isDisk bool) string {
	var content []byte
	var err error

	if isDisk {
		content, err = os.ReadFile(filePath)
	} else {
		content, err = fs.ReadFile(docs.EmbeddedFS, filePath)
	}

	if err != nil {
		// Fallback to filename
		return s.fileNameToTitle(filePath)
	}

	// Extract first # heading
	for _, line := range strings.Split(string(content), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "# ") {
			return strings.TrimSpace(line[2:])
		}
	}

	// Fallback to filename
	return s.fileNameToTitle(filePath)
}

// fileNameToTitle converts a file path to a readable title
func (s *Service) fileNameToTitle(filePath string) string {
	base := filepath.Base(filePath)
	name := strings.TrimSuffix(base, filepath.Ext(base))
	// Convert underscores and dashes to spaces, capitalize first letter
	name = strings.ReplaceAll(name, "_", " ")
	name = strings.ReplaceAll(name, "-", " ")
	if len(name) > 0 {
		name = strings.ToUpper(name[:1]) + name[1:]
	}
	return name
}
