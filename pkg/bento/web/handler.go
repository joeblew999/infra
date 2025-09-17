package web

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// BentoWebService provides web interface for Bento pipeline management
type BentoWebService struct{}

// NewBentoWebService creates a new bento web service
func NewBentoWebService() *BentoWebService {
	return &BentoWebService{}
}

// RegisterRoutes mounts all bento routes on the provided router
func (s *BentoWebService) RegisterRoutes(r chi.Router) {
	// API routes for bento functionality
	r.Get("/api/builder", HandlePipelineBuilder)
	r.Get("/api/templates", HandleGetTemplates)
	r.Post("/api/validate", HandlePipelineValidate)
	r.Post("/api/export", HandlePipelineExport)
}

// HandlePipelineBuilder handles pipeline builder requests
func HandlePipelineBuilder(w http.ResponseWriter, r *http.Request) {
	// Placeholder implementation - integrate with actual bento service
	response := map[string]any{
		"status": "success",
		"builder": "Bento Pipeline Builder",
		"version": "1.0.0",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleGetTemplates handles template retrieval requests
func HandleGetTemplates(w http.ResponseWriter, r *http.Request) {
	// Placeholder implementation - integrate with actual bento templates
	templates := []map[string]any{
		{"name": "basic", "description": "Basic pipeline template"},
		{"name": "advanced", "description": "Advanced pipeline template"},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"templates": templates,
	})
}

// HandlePipelineValidate handles pipeline validation requests
func HandlePipelineValidate(w http.ResponseWriter, r *http.Request) {
	var req map[string]any
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Placeholder implementation - integrate with actual bento validation
	response := map[string]any{
		"valid": true,
		"message": "Pipeline configuration is valid",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandlePipelineExport handles pipeline export requests
func HandlePipelineExport(w http.ResponseWriter, r *http.Request) {
	var req map[string]any
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Placeholder implementation - integrate with actual bento export
	response := map[string]any{
		"exported": true,
		"format": "yaml",
		"content": "# Bento pipeline configuration\ninput: {}\nprocessors: []\noutput: {}",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}