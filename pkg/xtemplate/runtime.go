package xtemplate

import (
	"context"
	"fmt"

	"github.com/joeblew999/infra/pkg/caddy"
	"github.com/joeblew999/infra/pkg/config"
	"github.com/joeblew999/infra/pkg/log"
	"github.com/joeblew999/infra/pkg/nats/orchestrator"
)

// RuntimeDeps represents external services ensured for xtemplate.
type RuntimeDeps struct {
	LeafURL string
	cleanup func()
}

// EnsureRuntime ensures xtemplate's dependencies (NATS leaf + Caddy) are running.
// It returns a cleanup function for dependencies that support graceful shutdown.
func EnsureRuntime(ctx context.Context) (*RuntimeDeps, error) {
	leafURL, shutdownLeaf, err := ensureNATS(ctx)
	if err != nil {
		return nil, err
	}

	if err := ensureCaddy(); err != nil {
		shutdownLeaf()
		return nil, err
	}

	return &RuntimeDeps{
		LeafURL: leafURL,
		cleanup: shutdownLeaf,
	}, nil
}

// Cleanup stops dependency processes that support graceful shutdown.
func (r *RuntimeDeps) Cleanup() {
	if r == nil {
		return
	}
	if r.cleanup != nil {
		r.cleanup()
	}
}

func ensureNATS(ctx context.Context) (string, func(), error) {
	opts := orchestrator.Options{EnsureCluster: !config.IsProduction()}
	leafURL, cleanup, err := orchestrator.Start(ctx, opts)
	if err != nil {
		return "", nil, fmt.Errorf("ensure nats: %w", err)
	}
	log.Info("XTemplate connected to NATS", "leaf_url", leafURL)
	return leafURL, cleanup, nil
}

func ensureCaddy() error {
	if err := caddy.StartSupervised(nil); err != nil {
		return fmt.Errorf("ensure caddy: %w", err)
	}
	log.Info("Caddy reverse proxy ensured", "port", config.GetCaddyPort())
	return nil
}
