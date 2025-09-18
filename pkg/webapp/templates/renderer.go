package templates

import (
	"html/template"
	"net/http"
	"strings"

	"github.com/joeblew999/infra/pkg/log"
)

// PageData holds the template data for rendering pages
type PageData struct {
	Navigation template.HTML
	Footer     template.HTML
	DataStar   template.HTML
	Header     template.HTML
}

// Renderer handles template rendering operations
type Renderer struct{}

// NewRenderer creates a new template renderer
func NewRenderer() *Renderer {
	return &Renderer{}
}

// RenderPageContent renders just the inner content and wraps it with the base template
func (r *Renderer) RenderPageContent(w http.ResponseWriter, contentHTML []byte, currentPath, title string) {
	// Parse and execute the content template
	tmpl, err := template.New("content").Parse(string(contentHTML))
	if err != nil {
		log.Error("Error parsing content template", "title", title, "error", err)
		w.Write(contentHTML) // Fallback to static HTML
		return
	}

	// Execute content template to get the inner HTML
	var contentBuf strings.Builder
	err = tmpl.Execute(&contentBuf, nil) // No data needed for simple content
	if err != nil {
		log.Error("Error executing content template", "title", title, "error", err)
		w.Write(contentHTML) // Fallback to static HTML
		return
	}

	// Use the centralized base template
	fullHTML, err := RenderBasePage(title, contentBuf.String(), currentPath)
	if err != nil {
		log.Error("Error rendering base page", "title", title, "error", err)
		w.Write(contentHTML) // Fallback to static HTML
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(fullHTML))
}

// RenderHomePage renders the home page
func (r *Renderer) RenderHomePage(w http.ResponseWriter, _ *http.Request) {
	r.RenderPageContent(w, IndexHTML, "/", "Infrastructure Management System")
}

// Render404Page renders the 404 error page
func (r *Renderer) Render404Page(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	r.RenderPageContent(w, NotFoundHTML, "/404", "Page Not Found")
}
