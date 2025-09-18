package web

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// WebService provides HTTP handlers for the xtemplate overview.
type WebService struct{}

// NewWebService constructs a WebService.
func NewWebService() *WebService {
	return &WebService{}
}

// RegisterRoutes mounts xtemplate routes on the provided router.
func (s *WebService) RegisterRoutes(r chi.Router) {
	r.Get("/", func(w http.ResponseWriter, _ *http.Request) {
		page, err := RenderOverviewPage()
		if err != nil {
			http.Error(w, "Failed to render xtemplate overview", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(page))
	})
}
