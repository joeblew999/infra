package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	gonats "github.com/nats-io/nats.go"

	"github.com/joeblew999/infra/pkg/bento"
	"github.com/joeblew999/infra/pkg/caddy"
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
		RunService(true, false, false, env) // Use embedded docs in production
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
	
	apiCheckCmd.Flags().String("old", "HEAD~1", "Old commit to compare against")
	apiCheckCmd.Flags().String("new", "HEAD", "New commit to compare")
}

func RunService(noDevDocs bool, noNATS bool, noPocketbase bool, mode string) {
	log.Info("Running in Service mode...")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		log.Info("🛑 Received shutdown signal, shutting down gracefully...")
		cancel()
		os.Exit(0)
	}()

	// Start NATS server (but don't fail if it doesn't work)
	log.Info("🚀 Step 1: Starting embedded NATS server...")
	natsAddr, err := nats.StartEmbeddedNATS(ctx)
	if err != nil {
		log.Warn("⚠️  Failed to start embedded NATS server, continuing without NATS", "error", err)
		natsAddr = "" // Mark as disabled
	} else {
		log.Info("✅ NATS server started", "address", natsAddr)
	}

	// Connect to NATS for client operations (only if NATS started)
	var nc *gonats.Conn
	if natsAddr != "" {
		log.Info("🔗 Connecting to NATS client...")
		nc, err = gonats.Connect(natsAddr)
		if err != nil {
			log.Warn("⚠️  Failed to connect to NATS client, continuing without NATS", "error", err)
			natsAddr = "" // Mark as disabled
		} else {
			log.Info("✅ NATS client connected")
			defer nc.Close()
		}
	}

	// Start PocketBase server (optional for now)
	pbEnv := "production"
	if mode == "development" {
		pbEnv = "development"
	}
	
	log.Info("🚀 Step 2: Starting PocketBase server...")
	pbPort := config.GetPocketBasePort()
	dataDir := config.GetPocketBaseDataPath()
	log.Info("📱 PocketBase configuration", "port", pbPort, "env", pbEnv, "data_dir", dataDir)
	
	// Ensure data directory exists
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		log.Error("Failed to create PocketBase data directory", "error", err)
	} else {
		log.Info("✅ PocketBase data directory ready", "path", dataDir)
	}
	
	pbServer := pocketbase.NewServer(pbEnv)
	if err := pbServer.Start(ctx); err != nil {
		log.Error("❌ PocketBase server failed to start", "error", err)
	} else {
		log.Info("✅ PocketBase server started", "url", pocketbase.GetAppURL(pbPort))
		defer func() {
			log.Info("⏹️  Stopping PocketBase server...")
			pbServer.Stop()
		}()
	}

	// Start Caddy reverse proxy (includes bento playground proxy)
	log.Info("🚀 Step 3: Starting Caddy reverse proxy...")
	caddyPort := "80"
	if config.ShouldUseHTTPS() {
		caddyPort = "443"
	}
	
	caddyDir := config.GetCaddyPath()
	if err := os.MkdirAll(caddyDir, 0755); err != nil {
		log.Error("Failed to create Caddy directory", "error", err)
	} else {
		log.Info("✅ Caddy directory ready", "path", caddyDir)
	}
	
	// Generate Caddyfile using new preset API
	caddyConfig := caddy.NewPresetConfig(caddy.PresetDevelopment, 80)
	if err := caddyConfig.GenerateAndSave("Caddyfile"); err != nil {
		log.Error("Failed to generate Caddyfile", "error", err)
	} else {
		log.Info("✅ Caddyfile generated", "path", filepath.Join(caddyDir, "Caddyfile"))
	}
	
	caddyfilePath := filepath.Join(caddyDir, "Caddyfile")
	caddyArgs := []string{"run", "--config", caddyfilePath}
	caddyCmd := exec.CommandContext(ctx, config.GetCaddyBinPath(), caddyArgs...)
	caddyCmd.Stdout = os.Stdout
	caddyCmd.Stderr = os.Stderr
	
	if err := caddyCmd.Start(); err != nil {
		log.Error("❌ Caddy failed to start", "error", err)
	} else {
		log.Info("✅ Caddy reverse proxy started", "url", fmt.Sprintf("http://localhost:%s", caddyPort))
		defer func() {
			log.Info("⏹️  Stopping Caddy reverse proxy...")
			caddyCmd.Process.Signal(syscall.SIGTERM)
			caddyCmd.Wait()
		}()
	}

	// Start Bento service
	log.Info("🚀 Step 4: Starting Bento stream processing service...")
	bentoPort := config.GetBentoPort()
	bentoDataDir := config.GetBentoPath()
	log.Info("🍱 Bento configuration", "port", bentoPort, "data_dir", bentoDataDir)
	
	// Ensure bento data directory exists
	if err := os.MkdirAll(bentoDataDir, 0755); err != nil {
		log.Error("Failed to create Bento data directory", "error", err)
	} else {
		log.Info("✅ Bento data directory ready", "path", bentoDataDir)
	}
	
	// Ensure bento config exists
	if err := bento.CreateDefaultConfig(); err != nil {
		log.Error("Failed to create default bento config", "error", err)
	} else {
		log.Info("✅ Bento configuration ready")
	}
	
	bentoService, err := bento.NewService(4195)
	if err != nil {
		log.Error("❌ Failed to create Bento service", "error", err)
	} else {
		if err := bentoService.Start(); err != nil {
			log.Error("❌ Bento service failed to start", "error", err)
		} else {
			log.Info("✅ Bento service started", "url", "http://localhost:4195", "proxy_url", fmt.Sprintf("http://localhost:%s/bento-playground", caddyPort))
			defer func() {
				log.Info("⏹️  Stopping Bento service...")
				bentoService.Stop()
			}()
		}
	}

	// Initialize multi-destination logging with NATS support (only if NATS started)
	loggingConfig := log.LoadConfig()
	if len(loggingConfig.Destinations) > 0 {
		if natsAddr != "" {
			// Use NATS-aware initialization since NATS is available
			if err := log.InitMultiLoggerWithNATS(loggingConfig, nc); err != nil {
				log.Warn("Failed to initialize NATS-aware multi-destination logging", "error", err)
			}
		} else {
			log.Info("Using basic logging (NATS unavailable)")
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
		if natsAddr != "" {
			// Start metric collection in a goroutine
			go gops.StartMetricCollection(ctx, nc, 5*time.Second) // Collect every 5 seconds
		} else {
			log.Warn("⚠️  Metric Agent requires NATS, but NATS is unavailable")
		}
	case "autoscaling-orchestrator":
		log.Info("Starting as Autoscaling Orchestrator")
		if natsAddr != "" {
			// TODO: Implement autoscaling orchestration logic here
		} else {
			log.Warn("⚠️  Autoscaling Orchestrator requires NATS, but NATS is unavailable")
		}
	default:
		log.Info("🚀 Step 4: Starting web server...")
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

		log.Info("🌐 Starting web server", "address", "http://localhost:1337", "embedded_docs", noDevDocs)
		// Start the web server (blocking)
		if err := web.StartServer(natsAddr, noDevDocs); err != nil {
			log.Error("❌ Failed to start web server", "error", err)
			os.Exit(1)
		}
	}
}

// shutdownCmd provides a way to find and kill running services
var shutdownCmd = &cobra.Command{
	Use:   "shutdown",
	Short: "Kill running service processes",
	Long:  "Find and kill any running service processes (web server, NATS, PocketBase)",
	Run: func(cmd *cobra.Command, args []string) {
		log.Info("🔍 Finding and killing running service processes...")
		
		// Kill by ports
		ports := []int{1337, 4222, 8090, 4195, 80, 443}
		for _, port := range ports {
			if err := gops.KillProcessByPort(port); err != nil {
				log.Warn("Failed to kill process on port", "port", port, "error", err)
			}
		}
		
		// Kill process by name
		gops.KillProcessByName("go run")
		gops.KillProcessByName("infra")
		gops.KillProcessByName("caddy")
		gops.KillProcessByName("bento")
		
		log.Info("✅ All service processes stopped")
	},
}

func init() {
	rootCmd.AddCommand(serviceCmd)
	rootCmd.AddCommand(apiCheckCmd)
	rootCmd.AddCommand(shutdownCmd)
	
	serviceCmd.Flags().String("env", "production", "Environment (production/development)")
	
}
