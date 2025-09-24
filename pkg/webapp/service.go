package webapp

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/nats-io/nats.go"

	"github.com/joeblew999/infra/pkg/auth"
	bentoweb "github.com/joeblew999/infra/pkg/bento/web"
	"github.com/joeblew999/infra/pkg/config"
	configweb "github.com/joeblew999/infra/pkg/config/web"
	demoweb "github.com/joeblew999/infra/pkg/demo"
	docsweb "github.com/joeblew999/infra/pkg/docs/web"
	"github.com/joeblew999/infra/pkg/goreman"
	goremanweb "github.com/joeblew999/infra/pkg/goreman/web"
	"github.com/joeblew999/infra/pkg/log"
	logsweb "github.com/joeblew999/infra/pkg/log/web"
	"github.com/joeblew999/infra/pkg/metrics"
	metricsweb "github.com/joeblew999/infra/pkg/metrics/web"
	natsweb "github.com/joeblew999/infra/pkg/nats/web"
	statusweb "github.com/joeblew999/infra/pkg/status/web"
	"github.com/joeblew999/infra/pkg/webapp/templates"
	xtemplateweb "github.com/joeblew999/infra/pkg/xtemplate/web"
)

// Option configures a Service.
type Option func(*Service)

// WithPort overrides the listening port.
func WithPort(port string) Option {
	return func(s *Service) {
		s.port = port
	}
}

// WithNATSURL overrides the NATS connection target. Empty string disables the connection.
func WithNATSURL(url string) Option {
	return func(s *Service) {
		s.natsURL = url
	}
}

// WithDocsDevMode toggles docs development mode (filesystem-backed navigation).
func WithDocsDevMode(enabled bool) Option {
	return func(s *Service) {
		s.docsDevMode = enabled
	}
}

// Service runs the web UI and supporting routes.
type Service struct {
	port        string
	natsURL     string
	docsDevMode bool

	router *chi.Mux
}

// NewService constructs a Service with repository defaults.
func NewService(opts ...Option) *Service {
	s := &Service{
		port:        config.GetWebServerPort(),
		natsURL:     config.GetNATSURL(),
		docsDevMode: true,
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

// Start boots the web server and blocks until the context is cancelled or the server exits.
func (s *Service) Start(ctx context.Context) error {
	log.Info("Starting web application", "port", s.port, "nats_url", s.natsURL, "docs_dev_mode", s.docsDevMode)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	collector := metrics.GetCollector()
	go collector.Start(ctx, 2*time.Second)

	var nc *nats.Conn
	if s.natsURL != "" {
		conn, err := nats.Connect(s.natsURL)
		if err != nil {
			log.Warn("Failed to connect to NATS, web UI will run without NATS integration", "error", err)
		} else {
			nc = conn
		}
	}
	if nc != nil {
		defer nc.Close()
	}

	if err := s.ensureRouter(ctx, nc); err != nil {
		return err
	}

	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", s.port),
		Handler: s.router,
	}

	errCh := make(chan error, 1)

	go func() {
		err := server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
			return
		}
		errCh <- nil
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelShutdown()
		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Warn("Error during web server shutdown", "error", err)
		}
		<-errCh // drain
		return nil
	case err := <-errCh:
		return err
	}
}

func (s *Service) ensureRouter(ctx context.Context, nc *nats.Conn) error {
	if s.router != nil {
		return nil
	}

	renderer := templates.NewRenderer()
	router := chi.NewRouter()

	app := &appState{
		natsConn: nc,
		router:   router,
		renderer: renderer,
		docsDev:  s.docsDevMode,
	}
	app.setupRoutes(ctx)

	s.router = router
	return nil
}

// appState mirrors the previous web.App struct but scoped to this package.
type appState struct {
	natsConn *nats.Conn
	router   *chi.Mux
	renderer *templates.Renderer
	docsDev  bool
}

func (a *appState) setupRoutes(ctx context.Context) {
	a.router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		a.renderer.RenderHomePage(w, r)
	})

	// Legacy redirect maintained for compatibility with docs links.
	a.router.Get("/bento-playground", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/bento/playground", http.StatusTemporaryRedirect)
	})

	docsWebService := docsweb.NewDocsWebService(a.docsDev)
	a.router.Route(config.DocsHTTPPath, docsWebService.RegisterRoutes)

	demoWebService := demoweb.NewDemoWebService(a.natsConn)
	a.router.Route("/demo", func(r chi.Router) {
		demoWebService.RegisterRoutes(r)
	})
	a.router.Get("/hello-world", demoWebService.HandleHelloWorld)

	metricsWebService := metricsweb.NewMetricsWebService()
	a.router.Route(config.MetricsHTTPPath, func(r chi.Router) {
		metricsWebService.RegisterRoutes(r)
	})

	statusWebService := statusweb.NewStatusWebService()
	a.router.Route(config.StatusHTTPPath, func(r chi.Router) {
		statusWebService.RegisterRoutes(r)
	})
	a.router.Route("/system", func(r chi.Router) {
		statusWebService.RegisterRoutes(r)
	})

	logsWebService := logsweb.NewLogsWebService()
	a.router.Route(config.LogsHTTPPath, func(r chi.Router) {
		logsWebService.RegisterRoutes(r)
	})

	natsWebService := natsweb.NewWebService()
	a.router.Route("/nats", func(r chi.Router) {
		natsWebService.RegisterRoutes(r)
	})

	xtemplateWebService := xtemplateweb.NewWebService()
	a.router.Route("/xtemplate", func(r chi.Router) {
		xtemplateWebService.RegisterRoutes(r)
	})

	bentoWebService := bentoweb.NewBentoWebService()
	a.router.Route("/bento", func(r chi.Router) {
		bentoWebService.RegisterRoutes(r)
	})

	webHandler := goremanweb.NewWebHandler("pkg/goreman/web")
	a.router.Route(config.RuntimeHTTPPath, webHandler.SetupRoutes)

	configWebService := configweb.NewConfigWebService()
	a.router.Route("/config", func(r chi.Router) {
		configWebService.RegisterRoutes(r)
	})

	authConfig := auth.WebAuthnConfig{
		RPDisplayName: "Infrastructure Management",
		RPID:          "localhost",
		RPOrigins:     []string{config.FormatLocalHTTP(config.GetWebServerPort())},
	}
	userStore := auth.NewInMemoryUserStore()
	sessionStore := auth.NewInMemorySessionStore()
	authService, _ := auth.NewAuthService(authConfig, userStore, sessionStore, "pkg/auth/web")
	a.router.Route("/auth", func(r chi.Router) {
		authService.RegisterRoutes(r)
	})

	if a.natsConn != nil {
		if err := goreman.StartCommandListener(ctx, a.natsConn); err != nil {
			log.Error("Failed to start goreman command listener", "error", err)
		}
		if err := demoWebService.SetupNATSStreams(ctx); err != nil {
			log.Warn("Failed to setup NATS streams", "error", err)
		}
	}

	a.router.NotFound(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		a.renderer.Render404Page(w, r)
	})
}
