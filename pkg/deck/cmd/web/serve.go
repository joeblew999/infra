package web

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joeblew999/infra/pkg/config"
	"github.com/joeblew999/infra/pkg/deck/web"
	"github.com/joeblew999/infra/pkg/log"
	"github.com/spf13/cobra"
)

var (
	port string
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the deck examples web viewer",
	Long: `Start a web server to view and generate deck examples.
	
This command serves a web interface that allows you to:
- Browse available .dsh example files
- Generate SVG, PNG, and PDF outputs
- View results in your browser

The server uses the testdata structure for inputs and outputs.`,
	RunE: runServe,
}

func runServe(cmd *cobra.Command, args []string) error {
	log.Info("Starting deck web server", "port", port)

	// Create server
	server, err := web.NewServer()
	if err != nil {
		return fmt.Errorf("failed to create server: %w", err)
	}

	// Setup routes
	router := server.SetupRoutes()

	// Create HTTP server
	httpServer := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	// Start server in goroutine
	go func() {
		log.Info("Deck web server started", "url", config.FormatLocalHTTP(port))
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("Server failed", "error", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("server forced to shutdown: %w", err)
	}

	log.Info("Server stopped")
	return nil
}

func init() {
	serveCmd.Flags().StringVarP(&port, "port", "p", "3000", "Port to serve on")
}
