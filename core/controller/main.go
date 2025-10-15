package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/joeblew999/infra/core/controller/pkg/apiserver"
	cloudflareprovider "github.com/joeblew999/infra/core/controller/pkg/providers/cloudflare"
	"github.com/joeblew999/infra/core/controller/pkg/reconcile"
)

func main() {
	var (
		specPath    = flag.String("spec", "controller/spec.yaml", "path to desired state spec")
		addr        = flag.String("addr", "127.0.0.1:4400", "address to bind the controller API")
		cfToken     = flag.String("cloudflare-token", "", "Cloudflare API token (overrides CLOUDFLARE_API_TOKEN)")
		cfTokenFile = flag.String("cloudflare-token-file", "", "Path to Cloudflare API token file (overrides CLOUDFLARE_API_TOKEN_FILE)")
	)
	flag.Parse()

	server, err := apiserver.New(*specPath)
	if err != nil {
		log.Fatalf("load desired state: %v", err)
	}

	srv := &http.Server{
		Addr:    *addr,
		Handler: server.Router(),
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	options := reconcile.Options{Tick: 30 * time.Second}
	if routing, err := loadCloudflareProvider(*cfToken, *cfTokenFile); err != nil {
		log.Fatalf("cloudflare provider: %v", err)
	} else if routing != nil {
		options.Routing = routing
	}
	go reconcile.New(server, options).Run(ctx)

	errCh := make(chan error, 1)
	go func() {
		log.Printf("controller listening on http://%s", *addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-sigCh:
		log.Println("shutting down controller...")
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()

		cancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Printf("server shutdown error: %v", err)
		}

		if err := server.Close(); err != nil {
			log.Printf("persist state error: %v", err)
		}
	case err := <-errCh:
		log.Fatalf("controller error: %v", err)
	}

	fmt.Println("controller stopped")
}

func loadCloudflareProvider(flagToken, flagFile string) (reconcile.RoutingProvider, error) {
	token := strings.TrimSpace(flagToken)
	if token == "" {
		token = strings.TrimSpace(os.Getenv("CLOUDFLARE_API_TOKEN"))
	}
	filePath := strings.TrimSpace(flagFile)
	if filePath == "" {
		filePath = strings.TrimSpace(os.Getenv("CLOUDFLARE_API_TOKEN_FILE"))
	}
	if token == "" && filePath != "" {
		data, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("read cloudflare token file %s: %w", filePath, err)
		}
		token = strings.TrimSpace(string(data))
	}
	if token == "" {
		return nil, nil
	}
	provider, err := cloudflareprovider.New(token)
	if err != nil {
		return nil, err
	}
	return provider, nil
}
