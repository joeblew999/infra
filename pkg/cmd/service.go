package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/joeblew999/infra/pkg/bento"
	"github.com/joeblew999/infra/pkg/caddy"
	"github.com/joeblew999/infra/pkg/config"
	"github.com/joeblew999/infra/pkg/deck"
	"github.com/joeblew999/infra/pkg/goreman"
	"github.com/joeblew999/infra/pkg/gops"
	"github.com/joeblew999/infra/pkg/log"
	"github.com/joeblew999/infra/pkg/mox"
	"github.com/joeblew999/infra/pkg/nats"
	"github.com/joeblew999/infra/pkg/pocketbase"
	"github.com/joeblew999/infra/pkg/xtemplate"
	"github.com/joeblew999/infra/web"
	"github.com/spf13/cobra"
)

// portToInt converts a port string to int, returns 0 on error
func portToInt(portStr string) int {
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return 0
	}
	return port
}

var serviceCmd = &cobra.Command{
	Use:   "service",
	Short: "Run in service mode (same as root command)",
	Long:  "Start all infrastructure services with goreman supervision. This is identical to running the root command without arguments.",
	Run: func(cmd *cobra.Command, args []string) {
		env, _ := cmd.Flags().GetString("env")
		noMox, _ := cmd.Flags().GetBool("no-mox")
		RunService(true, false, false, noMox, env) // Use embedded docs in production
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
	serviceCmd.Flags().Bool("no-mox", false, "Disable mox mail server")
}

func RunService(noDevDocs bool, noNATS bool, noPocketbase bool, noMox bool, mode string) {
	log.Info("Running in Service mode with goreman supervision...")

	var natsCleanup func()
	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup graceful shutdown using goreman
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		log.Info("🛑 Received shutdown signal, stopping all supervised processes...")
		if natsCleanup != nil {
			natsCleanup()
		}
		goreman.StopAll()
		cancel()
		os.Exit(0)
	}()

	// Start all services using goreman supervision
	log.Info("🚀 Starting all infrastructure services...")
	
	// Start web server FIRST for fast startup and health checks
	log.Info("🚀 Step 1: Starting web server (priority for health checks)...")
	// Check web server port availability
	webPort := config.GetWebServerPort()
	if !gops.IsPortAvailable(portToInt(webPort)) {
		log.Error("❌ Web server port "+webPort+" is already in use. Please free the port and try again.")
		os.Exit(1)
	}

	log.Info("🌐 Starting web server", "address", "http://0.0.0.0:"+webPort, "embedded_docs", noDevDocs)
	// Start web server in background so other services can start
	go func() {
		natsURL := "nats://localhost:" + config.GetNATSPort()
		if err := web.StartServer(natsURL, noDevDocs); err != nil {
			log.Error("❌ Failed to start web server", "error", err)
			os.Exit(1)
		}
	}()
	
	// Give web server a moment to start listening
	time.Sleep(500 * time.Millisecond)
	log.Info("✅ Web server started on port "+webPort)

	// Start embedded NATS server  
	var natsAddr string
	if !noNATS {
		log.Info("🚀 Step 2: Starting embedded NATS server...")
		var err error
		natsAddr, natsCleanup, err = nats.StartEmbeddedNATS(context.Background())
		if err != nil {
			log.Warn("⚠️  Failed to start embedded NATS server, continuing without NATS", "error", err)
			natsAddr = "" // Mark as disabled
		} else {
			log.Info("✅ Embedded NATS server started", "address", natsAddr)
			nats.StartS3GatewaySupervised(natsAddr)
		}
	}

	// Start embedded PocketBase server
	if !noPocketbase {
		log.Info("🚀 Step 3: Starting embedded PocketBase server...")
		pbEnv := "production"
		if mode == "development" {
			pbEnv = "development"
		}
		
		pbServer := pocketbase.NewServer(pbEnv)
		// Start PocketBase in a goroutine since it blocks
		go func() {
			if err := pbServer.Start(context.Background()); err != nil {
				log.Warn("PocketBase failed to start", "error", err)
			}
		}()
		log.Info("✅ Embedded PocketBase server started", "port", config.GetPocketBasePort())
	}

	// Start Caddy reverse proxy
	log.Info("🚀 Step 4: Starting Caddy reverse proxy...")
	if err := caddy.StartSupervised(nil); err != nil {
		log.Warn("Caddy failed to start", "error", err)  
	} else {
		log.Info("✅ Caddy reverse proxy started supervised")
	}

	// Start Bento service
	log.Info("🚀 Step 5: Starting Bento stream processing service...")
	bentoPort := portToInt(config.GetBentoPort())
	if err := bento.StartSupervised(bentoPort); err != nil {
		log.Warn("Bento failed to start", "error", err)
	} else {
		log.Info("✅ Bento service started supervised", "port", config.GetBentoPort())
	}

	// Start deck services
	log.Info("🚀 Step 6: Starting deck services...")
	
	// Start deck API service
	deckAPIPort := portToInt(config.GetDeckAPIPort())
	if err := deck.StartAPISupervised(deckAPIPort); err != nil {
		log.Warn("Deck API failed to start", "error", err)
	} else {
		log.Info("✅ Deck API service started supervised", "port", config.GetDeckAPIPort())
	}
	
	// Start deck file watcher service  
	if err := deck.StartWatcherSupervised([]string{"test/deck"}, []string{"svg", "png", "pdf"}); err != nil {
		log.Warn("Deck watcher failed to start", "error", err)
	} else {
		log.Info("✅ Deck watcher service started supervised")
	}

	// Start XTemplate development server
	log.Info("🚀 Step 7: Starting XTemplate development server...")
	if err := xtemplate.StartSupervised(); err != nil {
		log.Warn("XTemplate failed to start", "error", err)
	} else {
		log.Info("✅ XTemplate development server started supervised", "port", config.GetXTemplatePort())
	}

	// Start mox mail server
	if !noMox {
		log.Info("🚀 Step 8: Starting mox mail server...")
		if err := mox.StartSupervised("localhost", "admin@localhost"); err != nil {
			log.Warn("Mox failed to start", "error", err)
		} else {
			log.Info("✅ Mox mail server started supervised")
		}
	}

	// Show process status from goreman (external services only)
	log.Info("📊 External services started with goreman supervision")
	status := goreman.GetAllStatus()
	for name, stat := range status {
		log.Info("External process status", "name", name, "status", stat)
	}
	
	// All services started - keep main process running
	log.Info("🎉 All infrastructure services started successfully!")
	log.Info("💡 Web server accessible at http://0.0.0.0:"+webPort)
	
	// Block forever to keep all background services running
	select {}
}


// runContainerizedService builds and runs containerized service using ko and Docker
func runContainerizedService(environment string) error {
	log.Info("🐳 Building and running containerized service...")
	
	// First, bootstrap NATS cluster (idempotent)
	log.Info("🚀 Ensuring NATS cluster is running...")
	ctx := context.Background()
	
	if err := nats.StartLocalCluster(ctx); err != nil {
		log.Warn("Failed to start NATS cluster, continuing anyway", "error", err)
		// Don't fail completely - the containerized service might still work
	}
	
	// Use config for image naming
	imageName := config.GetDockerImageFullName()
	
	// Build image directly into Docker using ko
	log.Info("📦 Building container image with ko...")
	
	// Use ko to build directly into Docker daemon
	koPath := config.GetKoBinPath()
	if _, err := os.Stat(koPath); err != nil {
		return fmt.Errorf("ko binary not found at %s. Run 'go run . dep install ko' first", koPath)
	}
	
	// Set environment variables for ko build
	os.Setenv("KO_DOCKER_REPO", "ko.local")
	if environment == "production" || config.IsProduction() {
		os.Setenv("ENVIRONMENT", "production") 
	} else {
		os.Setenv("ENVIRONMENT", "development")
	}
	
	// Get git commit for build metadata
	commit := getGitCommit()
	if commit == "" {
		commit = "dev" // fallback for dirty git state
	}
	
	// Build directly into Docker daemon (no tarball needed)
	buildCmd := exec.Command(koPath, "build", "--push=false", "--platform=linux/amd64", "github.com/joeblew999/infra")
	buildCmd.Env = append(os.Environ(), "GIT_HASH="+commit)
	
	output, err := buildCmd.CombinedOutput()
	if err != nil {
		log.Error("ko build failed", "error", err, "output", string(output))
		return fmt.Errorf("failed to build container image: %w", err)
	}
	
	log.Info("✅ Built container image with ko")
	
	// Since ko builds into Docker daemon, find the latest image with ko.local prefix
	listCmd := exec.Command("docker", "images", "--format", "{{.Repository}}:{{.Tag}}", "--filter", "reference=ko.local/*", "--no-trunc")
	listOutput, err := listCmd.Output()
	if err != nil {
		log.Warn("Failed to list Docker images", "error", err)
		// Fallback to a reasonable default
		imageName = "ko.local/infra-bc4829dfbf7b0b49d219aad7c8cfa3f9:latest"
	} else {
		lines := strings.Split(strings.TrimSpace(string(listOutput)), "\n")
		// Find the most recent image with the right name pattern
		for _, line := range lines {
			if strings.Contains(line, "ko.local/infra") && (strings.HasSuffix(line, ":latest") || strings.Contains(line, ":")) {
				// Tag it with our desired name
				tagCmd := exec.Command("docker", "tag", line, imageName)
				if err := tagCmd.Run(); err != nil {
					log.Warn("Failed to tag image", "from", line, "to", imageName, "error", err)
					// Use the original image name if tagging fails
					imageName = line
				} else {
					log.Info("✅ Tagged image", "from", line, "to", imageName)
				}
				break
			}
		}
	}
	
	log.Info("✅ Container image ready", "image", imageName)
	
	// Step 3: Stop any existing containers using the same ports or name (idempotent behavior)
	log.Info("🧹 Stopping any existing containers on conflicting ports or with same name...")
	
	// First, stop any container with our specific name
	stopNameCmd := exec.Command("docker", "stop", "infra-service")
	if stopNameCmd.Run() == nil {
		log.Info("✅ Stopped existing container by name", "name", "infra-service")
	}
	
	// Also stop by ports for any other containers
	ports := []string{
		config.GetWebServerPort(),
		config.GetNATSPort(), 
		config.GetPocketBasePort(),
		config.GetBentoPort(),
		config.GetDeckAPIPort(),
		config.GetCaddyPort(),
		"443", // HTTPS
	}
	
	stoppedContainers := 0
	for _, port := range ports {
		// Find containers using this port
		psCmd := exec.Command("docker", "ps", "--filter", fmt.Sprintf("publish=%s", port), "--format", "{{.ID}}")
		if output, err := psCmd.Output(); err == nil {
			containerIDs := strings.Fields(strings.TrimSpace(string(output)))
			for _, containerID := range containerIDs {
				if containerID != "" {
					stopCmd := exec.Command("docker", "stop", containerID)
					if stopCmd.Run() == nil {
						stoppedContainers++
						log.Info("✅ Stopped existing container", "container_id", containerID, "port", port)
					}
				}
			}
		}
	}
	
	if stoppedContainers > 0 {
		log.Info("🧹 Cleanup complete", "containers_stopped", stoppedContainers)
		// Give containers time to fully stop
		time.Sleep(1 * time.Second)
	}
	
	// Step 4: Prepare Docker run command
	cwd, _ := os.Getwd()
	dataDir := filepath.Join(cwd, ".data")
	
	// Ensure .data directory exists
	os.MkdirAll(dataDir, 0755)
	
	// Build port mappings using config
	portMappings := []string{
		"-p", fmt.Sprintf("%s:%s", config.GetWebServerPort(), config.GetWebServerPort()),     // Web server
		"-p", fmt.Sprintf("%s:%s", config.GetNATSPort(), config.GetNATSPort()),             // NATS
		"-p", fmt.Sprintf("%s:%s", config.GetPocketBasePort(), config.GetPocketBasePort()), // PocketBase
		"-p", fmt.Sprintf("%s:%s", config.GetBentoPort(), config.GetBentoPort()),           // Bento
		"-p", fmt.Sprintf("%s:%s", config.GetDeckAPIPort(), config.GetDeckAPIPort()),       // Deck API
		"-p", fmt.Sprintf("%s:%s", config.GetXTemplatePort(), config.GetXTemplatePort()),   // XTemplate
		"-p", fmt.Sprintf("%s:%s", config.GetCaddyPort(), config.GetCaddyPort()),           // Caddy HTTP
		"-p", "443:443", // Caddy HTTPS
	}
	
	// Build Docker command (no -it for non-interactive mode)
	args := []string{"run", "--rm", "--name", "infra-service"}
	args = append(args, portMappings...)
	args = append(args, 
		"-v", fmt.Sprintf("%s:/app/.data", dataDir), // Mount data directory
		"-e", fmt.Sprintf("ENVIRONMENT=%s", environment),
		imageName,
		"service", // Run the service command inside container
	)
	
	log.Info("🚀 Starting container...", "image", imageName, "data_dir", dataDir)
	log.Info("📝 Container ports mapped using config functions")
	
	// Step 5: Setup signal handling for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		log.Info("🛑 Received shutdown signal, stopping container...")
		cancel()
	}()
	
	// Step 6: Run Docker container
	dockerCmd := exec.CommandContext(ctx, "docker", args...)
	dockerCmd.Stdout = os.Stdout
	dockerCmd.Stderr = os.Stderr
	dockerCmd.Stdin = os.Stdin
	
	log.Info("💡 Container accessible at the same URLs as direct mode")
	if err := dockerCmd.Run(); err != nil {
		if ctx.Err() != nil {
			log.Info("✅ Container stopped gracefully")
			return nil
		}
		return fmt.Errorf("docker run failed: %w", err)
	}
	
	return nil
}

// getGitCommit retrieves git commit information
func getGitCommit() string {
	// Try to get git commit
	if cmd := exec.Command("git", "rev-parse", "HEAD"); cmd != nil {
		if output, err := cmd.Output(); err == nil {
			return strings.TrimSpace(string(output))
		}
	}
	return ""
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
		
		// Try to kill the specific infra "go run ." process from this directory
		if err := gops.KillInfraGoRunProcess(); err == nil {
			log.Info("✅ Sent shutdown signal to infra go run process")
			mainProcessKilled = true
			// Give it time to shutdown gracefully
			time.Sleep(2 * time.Second)
		}
		
		// Also try to find compiled infra binary and send SIGTERM for graceful shutdown
		if err := gops.KillProcessByName("infra"); err == nil {
			log.Info("✅ Sent graceful shutdown signal to infra binary process")
			mainProcessKilled = true
			// Give it time to shutdown gracefully
			time.Sleep(1 * time.Second)
		}
		
		// Kill by ports using config functions
		log.Info("🔌 Shutting down services by port...")
		ports := []int{
			portToInt(config.GetWebServerPort()), // Web server
			portToInt(config.GetNATSPort()),      // NATS server
			portToInt(config.GetPocketBasePort()), // PocketBase
			portToInt(config.GetBentoPort()),      // Bento
			portToInt(config.GetDeckAPIPort()),    // Deck API
			portToInt(config.GetXTemplatePort()),  // XTemplate
			portToInt(config.GetCaddyPort()),      // Caddy HTTP
			443,  // Caddy HTTPS (fixed)
			portToInt(config.GetNatsS3Port()), // NATS S3 Gateway
			25,    // mox smtp
			143,   // mox imap
			465,   // mox smtps
			587,   // mox submission
			993,   // mox imaps
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
			"infra",       // Compiled binary (exact match only)
			"caddy",       // Caddy reverse proxy
			"bento",       // Bento stream processor
			"deck",        // Deck API server
			"nats-server", // NATS server binary
			"pocketbase",  // PocketBase server
			"xtemplate",   // XTemplate development server
			"nats-s3",     // NATS S3 Gateway
			"mox",         // Mox mail server
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

// containerCmd builds and runs containerized service using ko and Docker
var containerCmd = &cobra.Command{
	Use:   "container",
	Short: "Build and run containerized service with ko and Docker",
	Long: `Build the application with ko and run it in a Docker container.

This command:
- Builds the container image using ko
- Stops any conflicting containers (idempotent)
- Runs the container with proper port mappings
- Mounts data directory for persistence

This provides a containerized alternative to 'go run . service'.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		environment, _ := cmd.Flags().GetString("env")
		return runContainerizedService(environment)
	},
}

func init() {
	rootCmd.AddCommand(shutdownCmd)
	rootCmd.AddCommand(containerCmd)
	
	serviceCmd.Flags().String("env", "production", "Environment (production/development)")
	
	containerCmd.Flags().String("env", "production", "Environment (production/development)")
}

