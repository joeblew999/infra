package workflows

import (
	"context"

	"github.com/joeblew999/infra/pkg/config"
	"github.com/joeblew999/infra/pkg/log"
	"github.com/joeblew999/infra/pkg/nats/orchestrator"
)

// StartNATSStack boots the NATS stack with the provided options.
func StartNATSStack(ctx context.Context, ensureCluster bool) (string, func(), error) {
	leafURL, cleanup, err := orchestrator.Start(ctx, orchestrator.Options{EnsureCluster: ensureCluster})
	if err != nil {
		return "", nil, err
	}

	log.Info("NATS stack started", "leaf_url", leafURL, "ensure_cluster", ensureCluster)
	return leafURL, cleanup, nil
}

// StartNATSStackWithEnvironment derives options from the current config environment.
// In both development and production, we want to ensure cluster nodes are running.
// The orchestrator handles the topology difference:
// - Development: Start all 6 NATS nodes locally under goreman
// - Production on Fly: Start 1 NATS node for this app/region under goreman
func StartNATSStackWithEnvironment(ctx context.Context) (string, func(), error) {
	return StartNATSStack(ctx, config.ShouldEnsureNATSCluster())
}
