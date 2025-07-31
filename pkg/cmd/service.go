package cmd

import (
	"context"
	"os"
	"time"

	gonats "github.com/nats-io/nats.go"

	"github.com/joeblew999/infra/pkg/config"
	"github.com/joeblew999/infra/pkg/gops"
	"github.com/joeblew999/infra/pkg/log"
	"github.com/joeblew999/infra/pkg/nats"
	"github.com/joeblew999/infra/pkg/pocketbase"
	"github.com/joeblew999/infra/web"
	"github.com/spf13/cobra"
)

var serviceCmd = &cobra.Command{
	Use:   "service",
	Short: "Run in service mode",
	Run: func(cmd *cobra.Command, args []string) {
		env, _ := cmd.Flags().GetString("env")
		RunService(false, false, false, env) // Always start all services
	},
}

// apiCheckCmd provides API compatibility checking
var apiCheckCmd = &cobra.Command{
	Use:   "api-check",
	Short: "Check API compatibility between commits",
	Long: `Check API compatibility between two Git commits using apidiff.
This command helps ensure that public APIs remain backward compatible.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		oldCommit, _ := cmd.Flags().GetString("old")
		newCommit, _ := cmd.Flags().GetString("new")
		
		if oldCommit == "" {
			oldCommit = "HEAD~1"
		}
		if newCommit == "" {
			newCommit = "HEAD"
		}
		
		return runAPICompatibilityCheck(oldCommit, newCommit)
	},
}

func init() {
	rootCmd.AddCommand(serviceCmd)
	rootCmd.AddCommand(apiCheckCmd)
	
		serviceCmd.Flags().String("env", "production", "Environment (production/development)")
	
	apiCheckCmd.Flags().String("old", "HEAD~1", "Old commit to compare against")
	apiCheckCmd.Flags().String("new", "HEAD", "New commit to compare")
}

func RunService(noDevDocs bool, noNATS bool, noPocketbase bool, mode string) {
	log.Info("Running in Service mode...")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Always start NATS server
	log.Info("🚀 Step 1: Starting embedded NATS server...")
	var err error
	natsAddr, err := nats.StartEmbeddedNATS(ctx)
	if err != nil {
		log.Error("❌ Failed to start embedded NATS server", "error", err)
		os.Exit(1)
	}
	log.Info("✅ NATS server started", "address", natsAddr)

	// Connect to NATS for client operations
	log.Info("🔗 Connecting to NATS client...")
	nc, err := gonats.Connect(natsAddr)
	if err != nil {
		log.Error("❌ Failed to connect to NATS client", "error", err)
		os.Exit(1)
	}
	defer nc.Close()
	log.Info("✅ NATS client connected")

	// Always start PocketBase server
	pbEnv := "production"
	if mode == "development" {
		pbEnv = "development"
	}
	
	log.Info("🚀 Step 2: Starting PocketBase server...")
	pbPort := config.GetPocketBasePort()
	log.Info("📱 PocketBase configuration", "port", pbPort, "env", pbEnv, "data_dir", config.GetPocketBaseDataPath())
	
	pbServer := pocketbase.NewServer(pbEnv)
	if err := pbServer.Start(ctx); err != nil {
		log.Error("❌ Failed to start PocketBase server", "error", err)
		os.Exit(1)
	}
	defer func() {
		log.Info("⏹️  Stopping PocketBase server...")
		pbServer.Stop()
	}()
	
	log.Info("✅ PocketBase server started", "url", pocketbase.GetAppURL(pbPort))

	// Initialize multi-destination logging with NATS support
	loggingConfig := log.LoadConfig()
	if len(loggingConfig.Destinations) > 0 {
		// Use NATS-aware initialization since we always start NATS
		if err := log.InitMultiLoggerWithNATS(loggingConfig, nc); err != nil {
			log.Warn("Failed to initialize NATS-aware multi-destination logging", "error", err)
		}
	} else {
		// Use basic logging if no config
		log.Info("Using basic logging (no multi-destination config found)")
	}

	// Determine the service role based on the mode flag
	if mode == "" {
		mode = "service" // Default to service mode if not specified
	}

	switch mode {
	case "metric-agent":
		log.Info("Starting as Metric Agent")
		// Start metric collection in a goroutine
		go gops.StartMetricCollection(ctx, nc, 5*time.Second) // Collect every 5 seconds
	case "autoscaling-orchestrator":
		log.Info("Starting as Autoscaling Orchestrator")
		// TODO: Implement autoscaling orchestration logic here
	default:
		log.Info("🚀 Step 3: Starting web server...")
		// Check web server port availability
		if !gops.IsPortAvailable(1337) {
			log.Error("❌ Web server port 1337 is already in use. Please free the port and try again.")
			os.Exit(1)
		}

		// Check MCP server port availability
		if !gops.IsPortAvailable(8080) {
			log.Error("❌ MCP server port 8080 is already in use. Please free the port and try again.")
			os.Exit(1)
		}

		log.Info("🌐 Starting web server", "address", "http://localhost:1337", "dev_docs", !noDevDocs)
		// Start the web server (blocking)
		if err := web.StartServer(natsAddr, !noDevDocs); err != nil {
			log.Error("❌ Failed to start web server", "error", err)
			os.Exit(1)
		}
	}
}
