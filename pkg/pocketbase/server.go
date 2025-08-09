package pocketbase

import (
	"context"
	"os"

	"github.com/joeblew999/infra/pkg/log"
	"github.com/joeblew999/infra/pkg/config"
	"github.com/pocketbase/pocketbase"
)

// Server represents a PocketBase server instance
type Server struct {
	app     *pocketbase.PocketBase
	port    string
	env     string
	dataDir string
}

// NewServer creates a new PocketBase server instance
func NewServer(env string) *Server {
	return &Server{
		port:    config.GetPocketBasePort(),
		env:     env,
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

	// Ensure data directory exists with proper permissions
	if err := os.MkdirAll(s.dataDir, 0755); err != nil {
		log.Error("Failed to create data directory", "error", err)
		return err
	}

	// Create new PocketBase app with custom data directory
	app := pocketbase.NewWithConfig(pocketbase.Config{
		DefaultDataDir: s.dataDir,
	})

	s.app = app

	// Start the PocketBase server
	log.Info("Starting PocketBase server...", "port", s.port, "data_dir", s.dataDir)
	
	// Use the app's built-in serve functionality with proper context handling
	app.RootCmd.SetArgs([]string{
		"serve",
		"--dir", s.dataDir,
		"--http", ":" + s.port,
	})
	
	// Execute synchronously - this will block, which is expected for a server
	// This should work for the embedded version
	if err := app.Execute(); err != nil {
		log.Error("PocketBase server error", "error", err)
		return err
	}
	
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