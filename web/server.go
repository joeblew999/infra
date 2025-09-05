package web

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/nats-io/nats.go"
	"github.com/starfederation/datastar-go/datastar"

	"github.com/joeblew999/infra/pkg/docs"
	"github.com/joeblew999/infra/pkg/log"
	"github.com/joeblew999/infra/pkg/config"
	configweb "github.com/joeblew999/infra/pkg/config/web"
	"github.com/joeblew999/infra/pkg/auth"
	bentoweb "github.com/joeblew999/infra/pkg/bento/web"
	goremanweb "github.com/joeblew999/infra/pkg/goreman/web"
	gopsweb "github.com/joeblew999/infra/pkg/gops/web"
	logsweb "github.com/joeblew999/infra/pkg/logs/web"
	"github.com/joeblew999/infra/pkg/metrics"
	metricsweb "github.com/joeblew999/infra/pkg/metrics/web"
	pkgweb "github.com/joeblew999/infra/pkg/web"
)

//go:embed index.html
var helloWorldHTML []byte

//go:embed templates/logs.html
var logsHTML []byte

//go:embed templates/status.html
var statusHTML []byte

//go:embed templates/bento-playground.html
var bentoPlaygroundHTML []byte

//go:embed templates/404.html
var notFoundHTML []byte

const port = 1337

type App struct {
	natsConn     *nats.Conn
	router       *chi.Mux
	docsService  *docs.Service
	docsRenderer *docs.Renderer
}

type Store struct {
	Delay time.Duration `json:"delay"`
}

type Message struct {
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

func StartServer(natsAddr string, devDocs bool) error {
	ctx := context.Background()

	// Start metrics collection
	collector := metrics.GetCollector()
	go collector.Start(ctx, 2*time.Second) // Collect metrics every 2 seconds

	// Connect to NATS server (optional for debugging)
	nc, err := nats.Connect(natsAddr)
	if err != nil {
		log.Warn("Failed to connect to NATS, continuing without NATS features", "error", err)
		nc = nil // Allow web server to start without NATS
	} else {
		defer nc.Close()
	}

	app := &App{
		natsConn:     nc,
		router:       chi.NewRouter(),
		docsService:  docs.New(devDocs, config.DocsDir),
		docsRenderer: docs.NewRenderer(),
	}

	app.setupRoutes()
	if nc != nil {
		if err := app.setupNATSStreams(ctx); err != nil {
			log.Warn("Failed to setup NATS streams", "error", err)
		}
	}

	log.Info("Starting web server", "address", fmt.Sprintf("http://localhost:%d", port))
	if nc != nil {
		log.Info("Connected to NATS", "address", natsAddr)
	} else {
		log.Info("Running without NATS (debug mode)")
	}

	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), app.router); err != nil {
		return fmt.Errorf("Failed to start web server: %w", err)
	}
	return nil
}

// PageData holds the template data for rendering pages
type PageData struct {
	Navigation template.HTML
	Footer     template.HTML
	DataStar   template.HTML
	Header     template.HTML
}

// renderPageContent renders just the inner content and wraps it with the base template
func (app *App) renderPageContent(w http.ResponseWriter, contentHTML []byte, currentPath, title string) {
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
	fullHTML, err := pkgweb.RenderBasePage(title, contentBuf.String(), currentPath)
	if err != nil {
		log.Error("Error rendering base page", "title", title, "error", err)
		w.Write(contentHTML) // Fallback to static HTML
		return
	}
	
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(fullHTML))
}

func (app *App) renderHomePage(w http.ResponseWriter, _ *http.Request) {
	app.renderPageContent(w, helloWorldHTML, "/", "Infrastructure Management System")
}

func (app *App) setupRoutes() {
	app.router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		app.renderHomePage(w, r)
	})

	// Original hello-world endpoint
	app.router.Get("/hello-world", func(w http.ResponseWriter, r *http.Request) {
		store := &Store{}
		if err := datastar.ReadSignals(r, store); err != nil {
			log.Error("Error reading signals", "error", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		sse := datastar.NewSSE(w, r)
		const message = "Hello, world!"

		for i := range len(message) {
			if err := sse.PatchElements(`<div id="message">` + message[:i+1] + `</div>`); err != nil {
				log.Error("Error patching elements", "error", err)
				return
			}
			time.Sleep(100 * time.Millisecond) // 100ms delay for typing effect
		}
	})

	// New NATS-powered endpoint
	app.router.Get("/nats-stream", func(w http.ResponseWriter, r *http.Request) {
		app.handleNATSStream(w, r)
	})

	// Endpoint to publish messages to NATS
	app.router.Post("/publish", func(w http.ResponseWriter, r *http.Request) {
		app.handlePublish(w, r)
	})

	// API routes
	app.router.Get("/api/build", app.handleBuildInfo)
	app.router.Get("/api/system-status", app.handleSystemStatus)
	app.router.Get("/api/logs/stream", logsweb.StreamLogs)
	app.router.Get("/api/logs/config", logsweb.HandleLogConfig)
	app.router.Post("/api/logs/config", logsweb.HandleLogConfig)
	
	// Metrics API routes
	app.router.Get("/api/metrics", metricsweb.HandleMetricsAPI)
	app.router.Get("/api/metrics/history", metricsweb.HandleMetricsHistory)
	app.router.Get("/api/metrics/stream", metricsweb.HandleMetricsStream)
	
	// Bento pipeline builder API routes
	app.router.Get("/api/bento/builder", bentoweb.HandlePipelineBuilder)
	app.router.Get("/api/bento/templates", bentoweb.HandleGetTemplates)
	app.router.Post("/api/bento/validate", bentoweb.HandlePipelineValidate)
	app.router.Post("/api/bento/export", bentoweb.HandlePipelineExport)
	
	// Navigation routes
	app.router.Get("/bento-playground", app.handleBentoPlayground)
	app.router.Get(config.MetricsHTTPPath, metricsweb.HandleMetricsPage)
	app.router.Get(config.LogsHTTPPath, app.handleLogs)
	app.router.Get(config.StatusHTTPPath, app.handleStatus)

	// Docs handler
	app.router.Get(config.DocsHTTPPath+"*", app.handleDocs)

	// Process monitoring routes (goreman web GUI) - using sub-router pattern
	webHandler := goremanweb.NewWebHandler("pkg/goreman/web")
	app.router.Route("/processes", webHandler.SetupRoutes)
	
	// Configuration management routes - using sub-router pattern
	configWebService := configweb.NewConfigWebService()
	app.router.Route("/config", func(r chi.Router) {
		configWebService.RegisterRoutes(r)
	})
	
	// Authentication routes - using sub-router pattern
	// Simple auth setup with in-memory stores for demo
	authConfig := auth.WebAuthnConfig{
		RPDisplayName: "Infrastructure Management",
		RPID:          "localhost",
		RPOrigins:     []string{"http://localhost:1337"},
	}
	userStore := auth.NewInMemoryUserStore()
	sessionStore := auth.NewInMemorySessionStore()
	authService, _ := auth.NewAuthService(authConfig, userStore, sessionStore, "pkg/auth/web")
	app.router.Route("/auth", func(r chi.Router) {
		authService.RegisterRoutes(r)
	})
	
	// 404 handler (must be last)
	app.router.NotFound(app.handle404)
}

func (app *App) handleDocs(w http.ResponseWriter, r *http.Request) {
	filePath := r.URL.Path[len(config.DocsHTTPPath):]
	log.Info("Requested filePath", "path", filePath)

	// Handle folder access - if path ends with /, append README.md
	if strings.HasSuffix(filePath, "/") {
		filePath = filePath + "README.md"
	} else if filePath != "" && !strings.HasSuffix(filePath, ".md") {
		// If no extension, assume it's a folder and redirect to folder/README.md
		filePath = filePath + "/README.md"
	}

	// Read document content
	content, err := app.docsService.ReadFile(filePath)
	if err != nil {
		log.Error("Error reading document", "path", filePath, "error", err)
		http.Error(w, "Document not found", http.StatusNotFound)
		return
	}

	// Convert markdown to HTML
	htmlContent, err := app.docsRenderer.RenderToHTML(content)
	if err != nil {
		log.Error("Error rendering document", "path", filePath, "error", err)
		http.Error(w, "Failed to render document", http.StatusInternalServerError)
		return
	}

	// Wrap in HTML page structure with navigation
	nav := app.docsService.GetNavigation()
	fullHTML := app.docsRenderer.RenderToHTMLPage("Docs", htmlContent, nav, filePath)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(fullHTML))
}

func (app *App) setupNATSStreams(_ context.Context) error {
	// Create JetStream context
	js, err := app.natsConn.JetStream()
	if err != nil {
		log.Error("Failed to create JetStream context", "error", err)
		return fmt.Errorf("Failed to create JetStream context: %w", err)
	}

	// Create a stream for messages
	streamName := "MESSAGES"
	_, err = js.AddStream(&nats.StreamConfig{
		Name:      streamName,
		Subjects:  []string{"messages.*"},
		Retention: nats.LimitsPolicy,
		MaxAge:    time.Hour * 24, // Keep messages for 24 hours
	})
	if err != nil {
		log.Warn("NATS Stream may already exist", "error", err)
	}

	log.Info("NATS JetStream setup complete", "stream", streamName)
	return nil
}

func (app *App) handleNATSStream(w http.ResponseWriter, r *http.Request) {
	sse := datastar.NewSSE(w, r)

	// Subscribe to NATS messages
	sub, err := app.natsConn.Subscribe("messages.demo", func(msg *nats.Msg) {
		var message Message
		if err := json.Unmarshal(msg.Data, &message); err != nil {
			log.Error("Error unmarshaling message", "error", err)
			return
		}

		// Send the message to the browser via DataStar - append to existing messages
		messageHTML := fmt.Sprintf(`<div class="p-2 bg-blue-50 dark:bg-blue-900/20 rounded border-l-4 border-blue-400">` +
			`<div class="text-sm text-gray-600 dark:text-gray-400">%s</div>` +
			`<div class="text-gray-900 dark:text-white">%s</div>` +
			`</div>`,
			message.Timestamp.Format("15:04:05"), message.Content)

		// Append to the messages div using DataStar append mechanism
		if err := sse.PatchElements(fmt.Sprintf(`<div id="messages" _="append '%s'"></div>`, messageHTML)); err != nil {
			log.Error("Error sending SSE", "error", err)
		}
	})

	if err != nil {
		http.Error(w, "Failed to subscribe to NATS", http.StatusInternalServerError)
		return
	}
	defer sub.Unsubscribe()

	// Keep the connection alive
	<-r.Context().Done()
}

func (app *App) handlePublish(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Content string `json:"content"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	message := Message{
		Content:   req.Content,
		Timestamp: time.Now(),
	}

	data, err := json.Marshal(message)
	if err != nil {
		http.Error(w, "Failed to marshal message", http.StatusInternalServerError)
		return
	}

	// Publish to NATS
	if err := app.natsConn.Publish("messages.demo", data); err != nil {
		http.Error(w, "Failed to publish message", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "published"})
}

func (app *App) handleLogs(w http.ResponseWriter, _ *http.Request) {
	app.renderPageContent(w, logsHTML, "/logs", "Logs")
}

func (app *App) handleBuildInfo(w http.ResponseWriter, r *http.Request) {
	buildInfo := struct {
		Version     string `json:"version"`
		GitHash     string `json:"git_hash"`
		ShortHash   string `json:"short_hash"`
		BuildTime   string `json:"build_time"`
		Timestamp   string `json:"timestamp"`
		Environment string `json:"environment"`
	}{
		Version:     config.GetVersion(),
		GitHash:     config.GitHash,
		ShortHash:   config.GetShortHash(),
		BuildTime:   config.BuildTime,
		Timestamp:   time.Now().UTC().Format(time.RFC3339),
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

func (app *App) handleStatus(w http.ResponseWriter, _ *http.Request) {
	app.renderPageContent(w, statusHTML, "/status", "Status")
}

func (app *App) handleBentoPlayground(w http.ResponseWriter, _ *http.Request) {
	app.renderPageContent(w, bentoPlaygroundHTML, "/bento-playground", "Bento Pipeline Builder")
}

func (app *App) handle404(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	app.renderPageContent(w, notFoundHTML, "/404", "Page Not Found")
}

func (app *App) handleSystemStatus(w http.ResponseWriter, r *http.Request) {
	gopsweb.HandleSystemStatus(w, r)
}
