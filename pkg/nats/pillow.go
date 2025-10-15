package nats

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Nintron27/pillow"
	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/jwt/v2"
	"github.com/joeblew999/infra/pkg/config"
	"github.com/joeblew999/infra/pkg/nats/auth"
	"github.com/joeblew999/infra/pkg/log"
)

// PillowClusterConfig represents configuration for Pillow-based NATS clustering
type PillowClusterConfig struct {
	AppName         string
	Region          string
	AuthArtifacts   *auth.Artifacts
	EnableJetStream bool
	UseHubAndSpoke  bool // false = FlyioClustering, true = FlyioHubAndSpoke
}

// StartFlyClusterWithPillow starts a NATS cluster using Pillow's Fly.io adapters
func StartFlyClusterWithPillow(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}

	log.Info("Starting NATS cluster with Pillow integration")

	// Ensure auth artifacts are available
	authArtifacts, err := auth.Ensure(ctx)
	if err != nil {
		return fmt.Errorf("ensure auth artifacts: %w", err)
	}

	regions := config.GetFlyRegions()
	useHubAndSpoke := config.GetUsePillowHubAndSpoke()

	// Deploy to each region
	for _, region := range regions {
		pillowConfig := PillowClusterConfig{
			AppName:         fmt.Sprintf("nats-%s", region),
			Region:          region,
			AuthArtifacts:   authArtifacts,
			EnableJetStream: true,
			UseHubAndSpoke:  useHubAndSpoke,
		}

		if err := startPillowNode(ctx, pillowConfig); err != nil {
			return fmt.Errorf("failed to start Pillow node in region %s: %w", region, err)
		}
	}

	log.Info("NATS cluster started with Pillow", "regions", len(regions), "mode", pillowModeString(useHubAndSpoke))
	return nil
}

// EnsureFlyNodeWithPillow ensures a single NATS node using Pillow for the specified Fly app
func EnsureFlyNodeWithPillow(ctx context.Context, appName string) error {
	if ctx == nil {
		ctx = context.Background()
	}

	log.Info("Ensuring single NATS node with Pillow", "app", appName)

	// Extract region from app name (assumes format "nats-{region}")
	region := appName
	if len(appName) > 5 && appName[:5] == "nats-" {
		region = appName[5:]
	}

	// Ensure auth artifacts are available
	authArtifacts, err := auth.Ensure(ctx)
	if err != nil {
		return fmt.Errorf("ensure auth artifacts: %w", err)
	}

	pillowConfig := PillowClusterConfig{
		AppName:         appName,
		Region:          region,
		AuthArtifacts:   authArtifacts,
		EnableJetStream: true,
		UseHubAndSpoke:  config.GetUsePillowHubAndSpoke(),
	}

	if err := startPillowNode(ctx, pillowConfig); err != nil {
		return fmt.Errorf("failed to ensure Pillow node %s: %w", appName, err)
	}

	log.Info("Single NATS node ensured with Pillow", "app", appName, "region", region)
	return nil
}

// startPillowNode starts a single NATS node using Pillow with our auth integration
func startPillowNode(ctx context.Context, cfg PillowClusterConfig) error {
	// Create JetStream directory
	jetstreamDir, err := createJetStreamDir(cfg.AppName)
	if err != nil {
		return fmt.Errorf("create JetStream directory: %w", err)
	}

	// Configure NATS server options with our auth and JetStream setup
	natsOpts, err := createPillowNATSOptions(cfg, jetstreamDir)
	if err != nil {
		return fmt.Errorf("create NATS server options: %w", err)
	}

	// Configure Pillow options
	var opts []pillow.Option

	// Add our custom NATS server options first
	opts = append(opts, pillow.WithNATSServerOptions(natsOpts))

	// Enable logging for debugging
	opts = append(opts, pillow.WithLogging(true))

	// Add platform adapter based on configuration (must be last according to docs)
	isProduction := config.IsProduction()
	clusterName := config.GetNATSClusterName()

	if cfg.UseHubAndSpoke {
		opts = append(opts, pillow.WithPlatformAdapter(ctx, isProduction, &pillow.FlyioHubAndSpoke{
			ClusterName: clusterName,
		}))
	} else {
		opts = append(opts, pillow.WithPlatformAdapter(ctx, isProduction, &pillow.FlyioClustering{
			ClusterName: clusterName,
		}))
	}

	// Start the NATS server with Pillow
	log.Info("Starting NATS node with Pillow", "app", cfg.AppName, "region", cfg.Region, "jetstream", jetstreamDir)

	_, err = pillow.Run(opts...)
	if err != nil {
		return fmt.Errorf("pillow run failed: %w", err)
	}

	log.Info("NATS node started successfully with Pillow", "app", cfg.AppName)
	return nil
}

// createJetStreamDir creates a directory for JetStream storage
func createJetStreamDir(appName string) (string, error) {
	clusterDataPath := config.GetNATSClusterDataPath()
	jetstreamDir := filepath.Join(clusterDataPath, appName, "jetstream")

	if err := os.MkdirAll(jetstreamDir, 0755); err != nil {
		return "", fmt.Errorf("create jetstream dir: %w", err)
	}

	return jetstreamDir, nil
}

// createPillowNATSOptions creates NATS server options compatible with Pillow that includes our auth setup
func createPillowNATSOptions(cfg PillowClusterConfig, jetstreamDir string) (*server.Options, error) {
	// Parse the operator JWT to get the operator claims
	operatorClaims, err := jwt.DecodeOperatorClaims(cfg.AuthArtifacts.OperatorJWT)
	if err != nil {
		return nil, fmt.Errorf("decode operator JWT: %w", err)
	}

	// Create memory account resolver
	memResolver := &server.MemAccResolver{}

	// Add system account
	if err := memResolver.Store(cfg.AuthArtifacts.SystemAccountID, cfg.AuthArtifacts.SystemAccountJWT); err != nil {
		return nil, fmt.Errorf("store system account JWT: %w", err)
	}

	// Add application account
	if err := memResolver.Store(cfg.AuthArtifacts.ApplicationAccountID, cfg.AuthArtifacts.ApplicationAccountJWT); err != nil {
		return nil, fmt.Errorf("store application account JWT: %w", err)
	}

	opts := &server.Options{
		// Authentication using NSC-generated artifacts
		TrustedOperators: []*jwt.OperatorClaims{operatorClaims},
		AccountResolver:  memResolver,
		SystemAccount:    cfg.AuthArtifacts.SystemAccountID,

		// JetStream configuration
		JetStream: cfg.EnableJetStream,
		StoreDir:  jetstreamDir,

		// Logging
		Debug:   false,
		Trace:   false,
		Logtime: true,

		// Pillow will override these, but we set reasonable defaults:
		// - Server name and host binding
		// - Port configuration
		// - Cluster routing and discovery
		// - Leaf node setup for cross-region connectivity
	}

	return opts, nil
}

// GetPillowLeafRemotes returns leaf node URLs for connecting to a Pillow-managed cluster
func GetPillowLeafRemotes() []string {
	regions := config.GetFlyRegions()
	remotes := make([]string, 0, len(regions))

	for _, region := range regions {
		// Pillow uses standard NATS ports and Fly.io's internal networking
		appName := fmt.Sprintf("nats-%s", region)
		remotes = append(remotes, fmt.Sprintf("nats://%s.fly.dev:7422", appName))
	}

	return remotes
}

// pillowModeString returns a human-readable string for the Pillow clustering mode
func pillowModeString(useHubAndSpoke bool) string {
	if useHubAndSpoke {
		return "FlyioHubAndSpoke"
	}
	return "FlyioClustering"
}