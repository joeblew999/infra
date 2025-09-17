package web

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/joeblew999/infra/pkg/config"
	"github.com/joeblew999/infra/pkg/docs"
	"github.com/joeblew999/infra/pkg/log"
)

// DocsWebService provides web interface for documentation
type DocsWebService struct {
	service  *docs.Service
	renderer *docs.Renderer
}

// NewDocsWebService creates a new docs web service
func NewDocsWebService(devDocs bool) *DocsWebService {
	return &DocsWebService{
		service:  docs.New(devDocs, config.DocsDir),
		renderer: docs.NewRenderer(),
	}
}

// RegisterRoutes mounts all docs routes on the provided router
func (s *DocsWebService) RegisterRoutes(r chi.Router) {
	// Docs handler - catch-all for docs paths
	r.Get("/*", s.HandleDocs)
}

// HandleDocs handles documentation requests
func (s *DocsWebService) HandleDocs(w http.ResponseWriter, r *http.Request) {
	// Get the file path after /docs
	filePath := strings.TrimPrefix(r.URL.Path, config.DocsHTTPPath)
	log.Info("Requested filePath", "path", filePath)

	// Handle folder access - if path ends with /, append README.md
	if strings.HasSuffix(filePath, "/") {
		filePath = filePath + "README.md"
	} else if filePath != "" && !strings.HasSuffix(filePath, ".md") {
		// If no extension, assume it's a folder and redirect to folder/README.md
		filePath = filePath + "/README.md"
	}

	// Read document content
	content, err := s.service.ReadFile(filePath)
	if err != nil {
		log.Error("Error reading document", "path", filePath, "error", err)
		http.Error(w, "Document not found", http.StatusNotFound)
		return
	}

	// Convert markdown to HTML
	htmlContent, err := s.renderer.RenderToHTML(content)
	if err != nil {
		log.Error("Error rendering document", "path", filePath, "error", err)
		http.Error(w, "Failed to render document", http.StatusInternalServerError)
		return
	}

	// Wrap in HTML page structure with navigation
	nav := s.service.GetNavigation()
	fullHTML := s.renderer.RenderToHTMLPage("Docs", htmlContent, nav, filePath)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(fullHTML))
}