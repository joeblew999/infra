package web

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/joeblew999/infra/pkg/config"
	"github.com/joeblew999/infra/pkg/hugo"
	"github.com/joeblew999/infra/pkg/log"
)

// Handler manages HTTP requests for Hugo-generated documentation
type Handler struct {
	devMode     bool
	hugoService *hugo.Service
	staticDir   string
	proxy       *httputil.ReverseProxy
}

// NewHandler creates a new Hugo web handler
func NewHandler(devMode bool, docsDir string) *Handler {
	hugoService := hugo.NewService(devMode, docsDir)

	handler := &Handler{
		devMode:     devMode,
		hugoService: hugoService,
		staticDir:   hugoService.GetOutputDir(),
	}

	if devMode {
		// Set up reverse proxy to Hugo dev server using config port
		cfg := config.GetConfig()
		targetURL, _ := url.Parse(config.FormatLocalHTTP(cfg.Ports.Hugo))
		handler.proxy = httputil.NewSingleHostReverseProxy(targetURL)
	}

	return handler
}

// SetupRoutes configures the HTTP routes for documentation
func (h *Handler) SetupRoutes(r chi.Router) {
	setupHandler := func(r chi.Router) {
		if h.devMode {
			// In dev mode, proxy everything to Hugo dev server
			r.HandleFunc("/*", h.ProxyToHugo)
		} else {
			// In production, serve static files
			r.HandleFunc("/*", h.ServeStatic)
		}
	}

	// Serve Hugo docs on both /docs-hugo and /docs paths
	r.Route("/docs-hugo", setupHandler)
	r.Route("/docs", setupHandler)
}

// ProxyToHugo proxies requests to the Hugo development server
func (h *Handler) ProxyToHugo(w http.ResponseWriter, r *http.Request) {
	// Check if Hugo dev server is ready
	if !h.hugoService.IsReady() {
		http.Error(w, "Documentation is starting up, please wait...", http.StatusServiceUnavailable)
		return
	}

	// Remove /docs-hugo or /docs prefix for Hugo
	if strings.HasPrefix(r.URL.Path, "/docs-hugo") {
		r.URL.Path = strings.TrimPrefix(r.URL.Path, "/docs-hugo")
	} else if strings.HasPrefix(r.URL.Path, "/docs") {
		r.URL.Path = strings.TrimPrefix(r.URL.Path, "/docs")
	}
	if r.URL.Path == "" {
		r.URL.Path = "/"
	}

	log.Debug("Proxying docs request to Hugo", "path", r.URL.Path)
	h.proxy.ServeHTTP(w, r)
}

// ServeStatic serves pre-built static files
func (h *Handler) ServeStatic(w http.ResponseWriter, r *http.Request) {
	// Remove /docs-hugo or /docs prefix
	path := r.URL.Path
	if strings.HasPrefix(path, "/docs-hugo") {
		path = strings.TrimPrefix(path, "/docs-hugo")
	} else if strings.HasPrefix(path, "/docs") {
		path = strings.TrimPrefix(path, "/docs")
	}
	if path == "" || path == "/" {
		path = "/index.html"
	}

	// Serve from Hugo's output directory
	fileServer := http.FileServer(http.Dir(h.staticDir))

	// Modify request path
	r.URL.Path = path

	log.Debug("Serving static docs file", "path", path, "staticDir", h.staticDir)
	fileServer.ServeHTTP(w, r)
}

// GetService returns the Hugo service for external management
func (h *Handler) GetService() *hugo.Service {
	return h.hugoService
}

// HealthCheck returns the health status of the docs service
func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	if h.hugoService.IsReady() {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy","service":"hugo-docs"}`))
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte(`{"status":"starting","service":"hugo-docs"}`))
	}
}
