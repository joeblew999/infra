package docs

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/joeblew999/infra/docs"
)

// Service handles documentation serving from both filesystem and embedded sources
type Service struct {
	devMode bool
	docsDir string
}

// New creates a new docs service
func New(devMode bool, docsDir string) *Service {
	return &Service{
		devMode: devMode,
		docsDir: docsDir,
	}
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
