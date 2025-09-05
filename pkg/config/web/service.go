package web

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/joeblew999/infra/pkg/config"
	pkgweb "github.com/joeblew999/infra/pkg/web"
)

//go:embed templates/config-home.html
var configHomeHTML []byte

//go:embed templates/env-status.html
var envStatusHTML []byte

//go:embed templates/secrets.html
var secretsHTML []byte

// ConfigWebService provides web interface for configuration management
type ConfigWebService struct {
	templates *template.Template
}

// NewConfigWebService creates a new config web service
func NewConfigWebService() *ConfigWebService {
	// Load templates (for now, simple JSON rendering)
	return &ConfigWebService{
		templates: template.New("config"),
	}
}

// RegisterRoutes mounts all config routes on the provided router
func (s *ConfigWebService) RegisterRoutes(r chi.Router) {
	// Configuration display routes
	r.Get("/", s.handleConfigPage)
	r.Get("/json", s.handleConfigJSON)
	r.Get("/status", s.handleEnvStatus)
	r.Get("/secrets", s.handleSecretsManagement)
	
	// API routes
	r.Get("/api/config", s.handleConfigAPI)
	r.Get("/api/env-status", s.handleEnvStatusAPI)
	r.Post("/api/secrets/set", s.handleSetSecret)
}

// NewConfigRouter creates a subrouter with all config routes configured
func (s *ConfigWebService) NewConfigRouter() chi.Router {
	r := chi.NewRouter()
	s.RegisterRoutes(r)
	return r
}

// Mount mounts the config subrouter at the specified path
func (s *ConfigWebService) Mount(mainRouter chi.Router, path string) {
	mainRouter.Mount(path, s.NewConfigRouter())
}

// PageData holds the template data for rendering config pages
type PageData struct {
	Navigation    template.HTML
	Footer        template.HTML
	DataStar      template.HTML
	Header        template.HTML
	ConfigSubNav  template.HTML
	Config        interface{}
	EnvStatus     map[string]string
	MissingVars   []string
}

// renderTemplate is a helper for rendering config templates with nav/footer
func (s *ConfigWebService) renderTemplate(w http.ResponseWriter, templateHTML []byte, currentPath, templateName string, data interface{}) {
	// Get centralized navigation and footer
	navHTML, err := pkgweb.RenderNav(currentPath)
	if err != nil {
		navHTML = ""
	}
	
	footerHTML, err := pkgweb.RenderFooter()
	if err != nil {
		footerHTML = ""
	}
	
	// Parse and execute template
	tmpl, err := template.New(templateName).Parse(string(templateHTML))
	if err != nil {
		w.Write(templateHTML) // Fallback to static HTML
		return
	}
	
	// Render config sub-navigation
	configSubNavHTML, err := s.renderConfigSubNav(currentPath)
	if err != nil {
		configSubNavHTML = ""
	}
	
	pageData := PageData{
		Navigation:   template.HTML(navHTML),
		Footer:       template.HTML(footerHTML),
		DataStar:     pkgweb.GetDataStarScript(),
		Header:       pkgweb.GetHeaderHTML(),
		ConfigSubNav: template.HTML(configSubNavHTML),
	}
	
	// Set specific data based on template
	switch v := data.(type) {
	case *config.Config:
		pageData.Config = v
	case map[string]interface{}:
		if envStatus, ok := v["EnvStatus"].(map[string]string); ok {
			pageData.EnvStatus = envStatus
		}
		if missingVars, ok := v["MissingVars"].([]string); ok {
			pageData.MissingVars = missingVars
		}
	}
	
	if err := tmpl.Execute(w, pageData); err != nil {
		w.Write(templateHTML) // Fallback to static HTML
	}
}

// handleConfigPage serves the main configuration page
func (s *ConfigWebService) handleConfigPage(w http.ResponseWriter, r *http.Request) {
	cfg := config.GetConfig()
	s.renderTemplate(w, configHomeHTML, "/config", "config-home", cfg)
}

// handleConfigJSON serves configuration as JSON
func (s *ConfigWebService) handleConfigJSON(w http.ResponseWriter, r *http.Request) {
	cfg := config.GetConfig()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cfg)
}

// handleEnvStatus serves environment status page
func (s *ConfigWebService) handleEnvStatus(w http.ResponseWriter, r *http.Request) {
	envStatus := config.GetEnvStatus()
	missing := config.GetMissingEnvVars()
	
	data := map[string]interface{}{
		"EnvStatus":   envStatus,
		"MissingVars": missing,
	}
	
	s.renderTemplate(w, envStatusHTML, "/config/status", "env-status", data)
}

// handleSecretsManagement serves secrets management interface
func (s *ConfigWebService) handleSecretsManagement(w http.ResponseWriter, r *http.Request) {
	s.renderTemplate(w, secretsHTML, "/config/secrets", "secrets", nil)
}

// API handlers

// handleConfigAPI serves configuration as JSON API
func (s *ConfigWebService) handleConfigAPI(w http.ResponseWriter, r *http.Request) {
	s.handleConfigJSON(w, r)
}

// handleEnvStatusAPI serves environment status as JSON API
func (s *ConfigWebService) handleEnvStatusAPI(w http.ResponseWriter, r *http.Request) {
	envStatus := config.GetEnvStatus()
	missing := config.GetMissingEnvVars()
	
	response := map[string]interface{}{
		"env_status": envStatus,
		"missing":    missing,
		"all_set":    len(missing) == 0,
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleSetSecret handles secret storage requests
func (s *ConfigWebService) handleSetSecret(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	
	if req.Key == "" || req.Value == "" {
		http.Error(w, "Key and value are required", http.StatusBadRequest)
		return
	}
	
	// Use the secrets functionality we created
	if err := config.SetEnvSecret(req.Key, req.Value); err != nil {
		http.Error(w, fmt.Sprintf("Failed to store secret: %v", err), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success", "message": "Secret stored successfully"})
}

// ConfigSubNavItem represents a config sub-navigation menu item
type ConfigSubNavItem struct {
	Href   string
	Text   string
	Icon   string
	Active bool
}

// renderConfigSubNav renders the config sub-navigation HTML for a given path
func (s *ConfigWebService) renderConfigSubNav(currentPath string) (string, error) {
	items := []ConfigSubNavItem{
		{Href: "/config/", Text: "Config Home", Icon: "🏠"},
		{Href: "/config/json", Text: "JSON View", Icon: "📄"},
		{Href: "/config/status", Text: "ENV Status", Icon: "🔍"},
		{Href: "/config/secrets", Text: "Secrets", Icon: "🔐"},
	}
	
	// Mark current item as active
	for i := range items {
		if items[i].Href == currentPath {
			items[i].Active = true
		}
	}
	
	// Sub-navigation template
	const subNavHTML = `
<div class="mb-6 p-4 bg-gray-50 dark:bg-gray-800 rounded-lg">
    <div class="flex flex-wrap items-center gap-4">
        {{range .}}
        <a href="{{.Href}}" class="px-3 py-2 {{if .Active}}bg-blue-100 text-blue-800 dark:bg-blue-900/20 dark:text-blue-400 font-medium{{else}}text-gray-600 dark:text-gray-400 hover:bg-gray-200 dark:hover:bg-gray-700 transition-colors{{end}} rounded-lg text-sm">
            {{.Icon}} {{.Text}}
        </a>
        {{end}}
    </div>
</div>`
	
	tmpl, err := template.New("configSubNav").Parse(subNavHTML)
	if err != nil {
		return "", err
	}
	
	var result strings.Builder
	err = tmpl.Execute(&result, items)
	if err != nil {
		return "", err
	}
	
	return result.String(), nil
}