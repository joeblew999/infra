package pocketbase

import (
	"context"
	"github.com/joeblew999/infra/pkg/log"
	"os"

	"github.com/joeblew999/infra/pkg/config"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/plugins/migratecmd"
)

// Server represents a PocketBase server instance
type Server struct {
	app *pocketbase.PocketBase
	port string
	env  string
	dataDir string
}

// NewServer creates a new PocketBase server instance
func NewServer(env string) *Server {
	return &Server{
		port: config.GetPocketBasePort(),
		env:  env,
		dataDir: config.GetPocketBaseDataPath(),
	}
}

// SetDataDir sets the data directory for the server
func (s *Server) SetDataDir(dataDir string) {
	s.dataDir = dataDir
}

// Start starts the PocketBase server
func (s *Server) Start(ctx context.Context) error {
	log.Info("Starting PocketBase server", "port", s.port, "env", s.env, "data_dir", s.dataDir)
	
	// Ensure data directory exists
	if err := os.MkdirAll(s.dataDir, 0755); err != nil {
		return err
	}
	
	// Create new PocketBase app
	app := pocketbase.New()
	
	
	// Configure the app
	// Add custom routes here if needed
	// app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
	// 	e.Router.GET("/api/health", func(c echo.Context) error {
	// 		return c.JSON(200, map[string]string{"status": "ok"})
	// 	})
	// 	return nil
	// })

	s.app = app
	
	// Set environment variables for PocketBase
	os.Setenv("PB_ENV", s.env)
	os.Setenv("PORT", s.port)
	os.Setenv("PB_DATA_DIR", s.dataDir)
	
	// Configure development mode
	if s.env == "development" {
		// Enable automigration in development
		migratecmd.MustRegister(app, app.RootCmd, migratecmd.Config{
			Automigrate: true,
		})
	}
	
	// Force server mode with correct port format
	app.RootCmd.SetArgs([]string{"serve", "--http", ":" + s.port})
	
	// Start the server in a goroutine
	go func() {
		if err := app.Start(); err != nil {
			log.Error("PocketBase server error", "error", err)
		}
	}()
	
	return nil
}

// Stop gracefully stops the PocketBase server
func (s *Server) Stop() error {
	// PocketBase stops via context cancellation, no explicit Stop method needed
	return nil
}

// GetDataDir returns the PocketBase data directory
func GetDataDir() string {
	return config.GetPocketBaseDataPath()
}

// GetAppURL returns the base URL for the PocketBase app
func GetAppURL(port string) string {
	if port == "" {
		port = "8090"
	}
	return "http://localhost:" + port
}

// GetAPIURL returns the API URL for the PocketBase app
func GetAPIURL(port string) string {
	return GetAppURL(port) + "/api"
}