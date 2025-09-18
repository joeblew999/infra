package web

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	datastarlib "github.com/starfederation/datastar-go/datastar"

	"github.com/joeblew999/infra/pkg/config"
	"github.com/joeblew999/infra/pkg/log"
	"github.com/joeblew999/infra/pkg/webapp/templates"
)

func init() {
	templates.RegisterNavItem(templates.NavItem{
		Href:  config.ConfigHTTPPath,
		Text:  "Config",
		Icon:  "🛠️",
		Color: "cyan",
		Order: 70,
	})
}

//go:embed templates/config-page.html
var configPageTemplate string

const configStreamInterval = 4 * time.Second

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
	// Config page routes
	r.Get("/", HandleConfigPage)
	r.Get("/api/stream", HandleConfigStream)
	r.Get("/api/config", HandleConfigAPI)

	// Legacy routes for backward compatibility
	r.Get("/json", s.handleConfigJSON)
	r.Get("/api/build", s.HandleBuildInfo)
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
	Navigation   template.HTML
	Footer       template.HTML
	DataStar     template.HTML
	Header       template.HTML
	ConfigSubNav template.HTML
	Config       any
	EnvStatus    map[string]string
	MissingVars  []string
}

// renderTemplate is a helper for rendering config templates with nav/footer
func (s *ConfigWebService) renderTemplate(w http.ResponseWriter, templateHTML []byte, currentPath, templateName string, data any) {
	// Get centralized navigation and footer
	navHTML, err := templates.RenderNav(currentPath)
	if err != nil {
		navHTML = ""
	}

	footerHTML, err := templates.RenderFooter()
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
		DataStar:     templates.GetDataStarScript(),
		Header:       templates.GetHeaderHTML(),
		ConfigSubNav: template.HTML(configSubNavHTML),
	}

	// Set specific data based on template
	switch v := data.(type) {
	case *config.Config:
		pageData.Config = v
	case map[string]any:
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

// HandleConfigPage renders the config page with the shared base layout.
func HandleConfigPage(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.New("config-page").Parse(configPageTemplate)
	if err != nil {
		log.Error("error parsing config page template", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	var content strings.Builder
	data := struct {
		BasePath string
	}{BasePath: config.ConfigHTTPPath}

	if err := tmpl.Execute(&content, data); err != nil {
		log.Error("error executing config page template", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	fullHTML, err := templates.RenderBasePage("Configuration", content.String(), config.ConfigHTTPPath)
	if err != nil {
		log.Error("error rendering base page", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(fullHTML))
}

// handleConfigJSON serves configuration as JSON
func (s *ConfigWebService) handleConfigJSON(w http.ResponseWriter, r *http.Request) {
	cfg := config.GetConfig()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cfg)
}

// Legacy handlers - these can be removed once migration to DataStar is complete

// HandleConfigStream streams live config updates via DataStar SSE to the browser.
func HandleConfigStream(w http.ResponseWriter, r *http.Request) {
	sse := datastarlib.NewSSE(w, r)

	if err := sendConfigSnapshot(sse); err != nil {
		log.Error("error sending initial config snapshot", "error", err)
		return
	}

	ticker := time.NewTicker(configStreamInterval)
	defer ticker.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case <-ticker.C:
			if err := sendConfigSnapshot(sse); err != nil {
				log.Error("error streaming config snapshot", "error", err)
				return
			}
		}
	}
}

// HandleConfigAPI is an HTTP handler for config endpoint
func HandleConfigAPI(w http.ResponseWriter, r *http.Request) {
	cfg := config.GetConfig()
	envStatus := config.GetEnvStatus()
	missingVars := config.GetMissingEnvVars()

	response := map[string]any{
		"config":     cfg,
		"env_status": envStatus,
		"missing":    missingVars,
		"all_set":    len(missingVars) == 0,
		"timestamp":  time.Now().UTC(),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Error("Error encoding config data", "error", err)
		http.Error(w, "Failed to encode config data", http.StatusInternalServerError)
		return
	}
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

	response := map[string]any{
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

// HandleBuildInfo provides build information as JSON API
func (s *ConfigWebService) HandleBuildInfo(w http.ResponseWriter, r *http.Request) {
	buildInfo := struct {
		Version     string `json:"version"`
		GitHash     string `json:"git_hash"`
		ShortHash   string `json:"short_hash"`
		BuildTime   string `json:"build_time"`
		Timestamp   string `json:"timestamp"`
		Environment string `json:"environment"`
	}{
		Version:   config.GetVersion(),
		GitHash:   config.GitHash,
		ShortHash: config.GetShortHash(),
		BuildTime: config.BuildTime,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Environment: func() string {
			if config.IsProduction() {
				return "production"
			}
			return "development"
		}(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(buildInfo)
}

// Helper functions

func sendConfigSnapshot(sse *datastarlib.ServerSentEventGenerator) error {
	snapshot := mapConfigToTemplate()

	html, err := RenderConfigCards(snapshot)
	if err != nil {
		return err
	}

	return sse.PatchElements(html)
}
