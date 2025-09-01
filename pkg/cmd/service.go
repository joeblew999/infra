package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joeblew999/infra/pkg/bento"
	"github.com/joeblew999/infra/pkg/caddy"
	"github.com/joeblew999/infra/pkg/deck"
	"github.com/joeblew999/infra/pkg/goreman"
	"github.com/joeblew999/infra/pkg/gops"
	"github.com/joeblew999/infra/pkg/log"
	"github.com/joeblew999/infra/pkg/nats"
	"github.com/joeblew999/infra/pkg/pocketbase"
	"github.com/joeblew999/infra/web"
	"github.com/spf13/cobra"
)

var serviceCmd = &cobra.Command{
	Use:   "service",
	Short: "Run in service mode (same as root command)",
	Long:  "Start all infrastructure services with goreman supervision. This is identical to running the root command without arguments.",
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
	log.Info("Running in Service mode with goreman supervision...")

	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup graceful shutdown using goreman
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		log.Info("🛑 Received shutdown signal, stopping all supervised processes...")
		goreman.StopAll()
		cancel()
		os.Exit(0)
	}()

	// Start all services using goreman supervision
	log.Info("🚀 Starting all infrastructure services...")
	
	// Start NATS server  
	if !noNATS {
		log.Info("🚀 Step 1: Starting NATS server...")
		if err := nats.StartSupervised(4222); err != nil {
			log.Warn("NATS failed to start", "error", err)
		} else {
			log.Info("✅ NATS server started supervised", "port", 4222)
		}
	}

	// Start PocketBase server
	if !noPocketbase {
		log.Info("🚀 Step 2: Starting PocketBase server...")
		pbEnv := "production"
		if mode == "development" {
			pbEnv = "development"
		}
		if err := pocketbase.StartSupervised(pbEnv, "8090"); err != nil {
			log.Warn("PocketBase failed to start", "error", err)
		} else {
			log.Info("✅ PocketBase server started supervised", "port", 8090)
		}
	}

	// Start Caddy reverse proxy
	log.Info("🚀 Step 3: Starting Caddy reverse proxy...")
	if err := caddy.StartSupervised(nil); err != nil {
		log.Warn("Caddy failed to start", "error", err)  
	} else {
		log.Info("✅ Caddy reverse proxy started supervised")
	}

	// Start Bento service
	log.Info("🚀 Step 4: Starting Bento stream processing service...")
	if err := bento.StartSupervised(4195); err != nil {
		log.Warn("Bento failed to start", "error", err)
	} else {
		log.Info("✅ Bento service started supervised", "port", 4195)
	}

	// Start deck services
	log.Info("🚀 Step 5: Starting deck services...")
	
	// Start deck API service
	if err := deck.StartAPISupervised(8888); err != nil {
		log.Warn("Deck API failed to start", "error", err)
	} else {
		log.Info("✅ Deck API service started supervised", "port", 8888)
	}
	
	// Start deck file watcher service  
	if err := deck.StartWatcherSupervised([]string{"test/deck", "docs/deck"}, []string{"svg", "png", "pdf"}); err != nil {
		log.Warn("Deck watcher failed to start", "error", err)
	} else {
		log.Info("✅ Deck watcher service started supervised")
	}

	// Show process status from goreman
	log.Info("📊 All services started with goreman supervision")
	status := goreman.GetAllStatus()
	for name, stat := range status {
		log.Info("Process status", "name", name, "status", stat)
	}
	
	// Start web server (this runs in current process for now)
	log.Info("🚀 Step 6: Starting web server...")
	// Check web server port availability
	if !gops.IsPortAvailable(1337) {
		log.Error("❌ Web server port 1337 is already in use. Please free the port and try again.")
		os.Exit(1)
	}

	log.Info("🌐 Starting web server", "address", "http://localhost:1337", "embedded_docs", noDevDocs)
	// Start the web server (blocking) - for now we use basic NATS connection
	natsAddr := "nats://localhost:4222"
	if err := web.StartServer(natsAddr, noDevDocs); err != nil {
		log.Error("❌ Failed to start web server", "error", err)
		os.Exit(1)
	}
}

// shutdownCmd provides a way to find and kill running services
var shutdownCmd = &cobra.Command{
	Use:   "shutdown",
	Short: "Kill running service processes",
	Long:  "Find and kill all running service processes (goreman-supervised and standalone)",
	Run: func(cmd *cobra.Command, args []string) {
		log.Info("🛑 Shutting down all infrastructure services...")
		
		// First attempt: Try to find and signal the main service process for graceful shutdown
		log.Info("🔍 Looking for main service process...")
		mainProcessKilled := false
		
		// Try to find the main infra process and send SIGTERM for graceful shutdown
		if err := gops.KillProcessByName("infra"); err == nil {
			log.Info("✅ Sent graceful shutdown signal to main service process")
			mainProcessKilled = true
			// Give it time to shutdown gracefully
			time.Sleep(3 * time.Second)
		}
		
		// Kill by ports (including deck API port)
		log.Info("🔌 Shutting down services by port...")
		ports := []int{
			1337, // Web server
			4222, // NATS server  
			8090, // PocketBase
			4195, // Bento
			8888, // Deck API (NEW)
			80,   // Caddy HTTP
			443,  // Caddy HTTPS
		}
		
		portsKilled := 0
		for _, port := range ports {
			if err := gops.KillProcessByPort(port); err == nil {
				log.Info("✅ Stopped service on port", "port", port)
				portsKilled++
			}
		}
		
		// Kill by process name (goreman-supervised processes)
		log.Info("📝 Shutting down goreman-supervised processes...")
		processNames := []string{
			"go run",      // Main service process
			"infra",       // Compiled binary
			"caddy",       // Caddy reverse proxy
			"bento",       // Bento stream processor
			"deck",        // Deck API server
			"nats-server", // NATS server binary
			"pocketbase",  // PocketBase server
		}
		
		processesKilled := 0
		for _, name := range processNames {
			if err := gops.KillProcessByName(name); err == nil {
				log.Info("✅ Stopped process", "name", name)
				processesKilled++
			}
		}
		
		// Summary
		if mainProcessKilled {
			log.Info("✅ Main service process shutdown gracefully")
		}
		if portsKilled > 0 {
			log.Info("✅ Stopped services on ports", "count", portsKilled)
		}
		if processesKilled > 0 {
			log.Info("✅ Stopped processes by name", "count", processesKilled)  
		}
		
		if mainProcessKilled || portsKilled > 0 || processesKilled > 0 {
			log.Info("🎉 All infrastructure services shutdown complete!")
		} else {
			log.Info("ℹ️  No running services found to shutdown")
		}
	},
}

func init() {
	rootCmd.AddCommand(shutdownCmd)
	
	serviceCmd.Flags().String("env", "production", "Environment (production/development)")
}

