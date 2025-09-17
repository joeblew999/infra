package web

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/nats-io/nats.go"

	"github.com/joeblew999/infra/pkg/auth"
	bentoweb "github.com/joeblew999/infra/pkg/bento/web"
	"github.com/joeblew999/infra/pkg/config"
	configweb "github.com/joeblew999/infra/pkg/config/web"
	docsweb "github.com/joeblew999/infra/pkg/docs/web"
	goreman "github.com/joeblew999/infra/pkg/goreman"
	goremanweb "github.com/joeblew999/infra/pkg/goreman/web"
	"github.com/joeblew999/infra/pkg/log"
	logsweb "github.com/joeblew999/infra/pkg/log/web"
	"github.com/joeblew999/infra/pkg/metrics"
	metricsweb "github.com/joeblew999/infra/pkg/metrics/web"
	statusweb "github.com/joeblew999/infra/pkg/status/web"
	"github.com/joeblew999/infra/web/demo"
	"github.com/joeblew999/infra/web/templates"
)

type App struct {
	natsConn *nats.Conn
	router   *chi.Mux
	renderer *templates.Renderer
}

func StartServer(natsAddr string, devDocs bool) error {
	baseCtx := context.Background()
	ctx, cancel := context.WithCancel(baseCtx)
	defer cancel()

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
		natsConn: nc,
		router:   chi.NewRouter(),
		renderer: templates.NewRenderer(),
	}

	app.setupRoutes(devDocs)
	if nc != nil {
		if err := goreman.StartCommandListener(ctx, nc); err != nil {
			log.Error("Failed to start goreman command listener", "error", err)
		}

		demoWebService := demo.NewDemoWebService(nc)
		if err := demoWebService.SetupNATSStreams(ctx); err != nil {
			log.Warn("Failed to setup NATS streams", "error", err)
		}
	}

	log.Info("Starting web server", "address", fmt.Sprintf("http://localhost:%s", config.GetWebServerPort()))
	if nc != nil {
		log.Info("Connected to NATS", "address", natsAddr)
	} else {
		log.Info("Running without NATS (debug mode)")
	}

	if err := http.ListenAndServe(fmt.Sprintf(":%s", config.GetWebServerPort()), app.router); err != nil {
		return fmt.Errorf("Failed to start web server: %w", err)
	}
	return nil
}

func (app *App) setupRoutes(devDocs bool) {
	app.router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		app.renderHomePage(w, r)
	})

	// Navigation routes
	app.router.Get("/bento-playground", app.handleBentoPlayground)
	app.router.Get(config.LogsHTTPPath, app.handleLogs)

	// Docs handler - using sub-router pattern
	docsWebService := docsweb.NewDocsWebService(devDocs)
	app.router.Route(config.DocsHTTPPath, func(r chi.Router) {
		docsWebService.RegisterRoutes(r)
	})

	// Demo routes - using sub-router pattern
	demoWebService := demo.NewDemoWebService(app.natsConn)
	app.router.Route("/demo", func(r chi.Router) {
		demoWebService.RegisterRoutes(r)
	})

	// Direct hello-world route for DataStar compatibility
	app.router.Get("/hello-world", demoWebService.HandleHelloWorld)

	// Sub-router implementations

	// Metrics routes - using sub-router pattern
	metricsWebService := metricsweb.NewMetricsWebService()
	app.router.Route(config.MetricsHTTPPath, func(r chi.Router) {
		metricsWebService.RegisterRoutes(r)
	})

	// Status routes - using sub-router pattern
	statusWebService := statusweb.NewStatusWebService()

	// Logs routes - using sub-router pattern
	logsWebService := logsweb.NewLogsWebService()
	app.router.Route("/logs", func(r chi.Router) {
		logsWebService.RegisterRoutes(r)
	})

	// System monitoring routes - share handlers between /status and /system
	app.router.Route(config.StatusHTTPPath, func(r chi.Router) {
		statusWebService.RegisterRoutes(r)
	})
	app.router.Route("/system", func(r chi.Router) {
		statusWebService.RegisterRoutes(r)
	})

	// Bento routes - using sub-router pattern
	bentoWebService := bentoweb.NewBentoWebService()
	app.router.Route("/bento", func(r chi.Router) {
		bentoWebService.RegisterRoutes(r)
	})

	// Process monitoring routes (goreman web GUI) - using sub-router pattern
	webHandler := goremanweb.NewWebHandler("pkg/goreman/web")
	app.router.Route(config.ProcessesHTTPPath, webHandler.SetupRoutes)

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
