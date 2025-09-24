package orchestrator

import (
	"context"
	"fmt"
	"os"

	"github.com/joeblew999/infra/pkg/config"
	"github.com/joeblew999/infra/pkg/goreman"
	"github.com/joeblew999/infra/pkg/log"
	"github.com/joeblew999/infra/pkg/nats"
	"github.com/joeblew999/infra/pkg/nats/auth"
	"github.com/joeblew999/infra/pkg/nats/gateway"
)

// Options controls how the NATS stack is brought up.
type Options struct {
	EnsureCluster bool
}

// Start boots the NATS cluster (if requested), embedded leaf, and supporting gateways.
// It returns the leaf URL and a cleanup function to stop the embedded node.
func Start(ctx context.Context, opts Options) (string, func(), error) {
	authArtifacts, err := auth.Ensure(ctx)
	if err != nil {
		return "", nil, fmt.Errorf("prepare nats auth: %w", err)
	}

	if opts.EnsureCluster {
		if config.IsProduction() {
			// On Fly, ensure the entire NATS cluster is deployed first
			flyAppName := os.Getenv("FLY_APP_NAME")
			if flyAppName != "" {
				log.Info("Ensuring NATS cluster deployment to Fly.io")
				if err := nats.DeployFlyCluster(ctx); err != nil {
					log.Warn("Failed to deploy NATS cluster, continuing with existing deployment", "error", err)
				}

				log.Info("Ensuring single NATS node for Fly app", "app", flyAppName)
				if err := nats.EnsureSingleFlyNode(ctx, flyAppName); err != nil {
					return "", nil, fmt.Errorf("failed to ensure Fly NATS node %s: %w", flyAppName, err)
				}
			} else {
				log.Warn("Production mode but no FLY_APP_NAME found, skipping cluster start")
			}
		} else {
			// Local development: start all 6 NATS nodes
			log.Info("Ensuring goreman-managed local NATS cluster")
			if err := nats.EnsureCluster(ctx, nats.GetLocalClusterConfig(), authArtifacts); err != nil {
				return "", nil, fmt.Errorf("failed to ensure local NATS cluster: %w", err)
			}
		}
	}

	var remotes []string
	if opts.EnsureCluster {
		useLocalCluster := !config.IsProduction()
		remotes = nats.GetClusterLeafRemotes(useLocalCluster)
	}
	leafURL, conn, shutdownLeaf, err := nats.StartEmbeddedNATS(ctx, remotes, authArtifacts.ApplicationCredsPath)
	if err != nil {
		return "", nil, err
	}

	log.Info("Embedded NATS leaf ready", "url", leafURL, "remotes", len(remotes))

	gateway.StartS3Gateway(leafURL)
	if err := goreman.StartCommandListener(ctx, conn); err != nil {
		shutdownLeaf()
		return "", nil, fmt.Errorf("failed to start goreman command listener: %w", err)
	}

	cleanup := func() {
		shutdownLeaf()
	}

	return leafURL, cleanup, nil
}

// StartWithEnvironment derives options from the current config environment and starts the stack.
func StartWithEnvironment(ctx context.Context) (string, func(), error) {
	return Start(ctx, Options{EnsureCluster: config.ShouldEnsureNATSCluster()})
}
