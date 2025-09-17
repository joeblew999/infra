package runtime

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
	"github.com/joeblew999/infra/pkg/gops"
	"github.com/joeblew999/infra/pkg/goreman"
	"github.com/joeblew999/infra/pkg/log"
	"github.com/joeblew999/infra/pkg/mox"
	"github.com/joeblew999/infra/pkg/nats"
	"github.com/joeblew999/infra/pkg/pocketbase"
	"github.com/joeblew999/infra/pkg/xtemplate"
	"github.com/joeblew999/infra/web"
)

// PreflightFunc allows callers to hook development-time preparation before startup.
type PreflightFunc func(context.Context)

// Options control service startup behaviour.
type Options struct {
	Mode         string
	NoDevDocs    bool
	NoNATS       bool
	NoPocketbase bool
	NoMox        bool
	Preflight    PreflightFunc
}

// Start launches all infrastructure services under goreman supervision.
// It blocks until a shutdown signal is received or a startup error occurs.
func Start(opts Options) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if opts.Preflight != nil {
		opts.Preflight(ctx)
	}

	log.Info("Running in Service mode with goreman supervision...")

	var natsCleanup func()
	errCh := make(chan error, 1)

	defer func() {
		if natsCleanup != nil {
			natsCleanup()
		}
		goreman.StopAll()
	}()

	// Setup graceful shutdown handling.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	go func() {
		select {
		case <-ctx.Done():
			return
		case <-sigCh:
			log.Info("🛑 Received shutdown signal, stopping all supervised processes...")
			cancel()
		}
	}()

	// Helper to record fatal startup errors.
	recordErr := func(err error) {
		if err == nil {
			return
		}
		select {
		case errCh <- err:
		default:
		}
		cancel()
	}

	// Start web server first for health checks.
	log.Info("🚀 Starting all infrastructure services...")
	log.Info("🚀 Step 1: Starting web server (priority for health checks)...")

	webPort := config.GetWebServerPort()
	if !gops.IsPortAvailable(portToInt(webPort)) {
		err := fmt.Errorf("web server port %s is already in use", webPort)
		log.Error("❌ Web server port in use. Please free the port and try again.", "port", webPort)
		return err
	}

	log.Info("🌐 Starting web server", "address", "http://0.0.0.0:"+webPort, "embedded_docs", opts.NoDevDocs)
	go func() {
		natsURL := "nats://localhost:" + config.GetNATSPort()
		if err := web.StartServer(natsURL, opts.NoDevDocs); err != nil {
			log.Error("❌ Failed to start web server", "error", err)
			recordErr(fmt.Errorf("web server failed to start: %w", err))
		}
	}()

	time.Sleep(500 * time.Millisecond)
	log.Info("✅ Web server started on port " + webPort)

	// Start embedded NATS server when required.
	if !opts.NoNATS {
		log.Info("🚀 Step 2: Starting embedded NATS server...")
		addr, conn, cleanup, err := nats.StartEmbeddedNATS(context.Background())
		if err != nil {
			log.Warn("⚠️  Failed to start embedded NATS server, continuing without NATS", "error", err)
		} else {
			natsCleanup = cleanup
			log.Info("✅ Embedded NATS server started", "address", addr)
			nats.StartS3GatewaySupervised(addr)
			if err := goreman.StartCommandListener(ctx, conn); err != nil {
				log.Error("Failed to start goreman control listener", "error", err)
			} else {
				log.Info("✅ Goreman control channel ready", "subject", goreman.CommandSubject)
			}
		}
	}

	// Start embedded PocketBase server.
	if !opts.NoPocketbase {
		log.Info("🚀 Step 3: Starting embedded PocketBase server...")
		pbEnv := "production"
		if opts.Mode == "development" {
			pbEnv = "development"
		}

		pbServer := pocketbase.NewServer(pbEnv)
		go func() {
			if err := pbServer.Start(context.Background()); err != nil {
				log.Warn("PocketBase failed to start", "error", err)
			}
		}()
		log.Info("✅ Embedded PocketBase server started", "port", config.GetPocketBasePort())
	}

	// Start Caddy reverse proxy.
	log.Info("🚀 Step 4: Starting Caddy reverse proxy...")
	if err := caddy.StartSupervised(nil); err != nil {
		log.Warn("Caddy failed to start", "error", err)
	} else {
		log.Info("✅ Caddy reverse proxy started supervised")
	}

	// Start Bento service.
	log.Info("🚀 Step 5: Starting Bento stream processing service...")
	if err := bento.StartSupervised(portToInt(config.GetBentoPort())); err != nil {
		log.Warn("Bento failed to start", "error", err)
	} else {
		log.Info("✅ Bento service started supervised", "port", config.GetBentoPort())
	}

	// Start deck services.
	log.Info("🚀 Step 6: Starting deck services...")
	if err := deck.StartAPISupervised(portToInt(config.GetDeckAPIPort())); err != nil {
		log.Warn("Deck API failed to start", "error", err)
	} else {
		log.Info("✅ Deck API service started supervised", "port", config.GetDeckAPIPort())
	}

	if err := deck.StartWatcherSupervised([]string{"test/deck"}, []string{"svg", "png", "pdf"}); err != nil {
		log.Warn("Deck watcher failed to start", "error", err)
	} else {
		log.Info("✅ Deck watcher service started supervised")
	}

	// Start XTemplate server.
	log.Info("🚀 Step 7: Starting XTemplate development server...")
	if err := xtemplate.StartSupervised(); err != nil {
		log.Warn("XTemplate failed to start", "error", err)
	} else {
		log.Info("✅ XTemplate development server started supervised", "port", config.GetXTemplatePort())
	}

	// Start mox mail server.
	if !opts.NoMox {
		log.Info("🚀 Step 8: Starting mox mail server...")
		if err := mox.StartSupervised("localhost", "admin@localhost"); err != nil {
			log.Warn("Mox failed to start", "error", err)
		} else {
			log.Info("✅ Mox mail server started supervised")
		}
	}

	// Show goreman status for external processes.
	log.Info("📊 External services started with goreman supervision")
	status := goreman.GetAllStatus()
	for name, stat := range status {
		log.Info("External process status", "name", name, "status", stat)
	}

	log.Info("🎉 All infrastructure services started successfully!")
	log.Info("💡 Web server accessible at http://0.0.0.0:" + webPort)

	<-ctx.Done()

	select {
	case err := <-errCh:
		return err
	default:
		return nil
	}
}

// Shutdown stops running infrastructure services by signalling processes, ports, and goreman groups.
func Shutdown() {
	log.Info("🛑 Shutting down all infrastructure services...")

	log.Info("🔍 Looking for main service process...")
	mainProcessKilled := false

	if err := gops.KillInfraGoRunProcess(); err == nil {
		log.Info("✅ Sent shutdown signal to infra go run process")
		mainProcessKilled = true
		time.Sleep(2 * time.Second)
	}

	if err := gops.KillProcessByName("infra"); err == nil {
		log.Info("✅ Sent graceful shutdown signal to infra binary process")
		mainProcessKilled = true
		time.Sleep(1 * time.Second)
	}

	log.Info("🔌 Shutting down services by port...")
	ports := []int{
		portToInt(config.GetWebServerPort()),
		portToInt(config.GetNATSPort()),
		portToInt(config.GetPocketBasePort()),
		portToInt(config.GetBentoPort()),
		portToInt(config.GetDeckAPIPort()),
		portToInt(config.GetXTemplatePort()),
		portToInt(config.GetCaddyPort()),
		443,
		portToInt(config.GetNatsS3Port()),
		25,
		143,
		465,
		587,
		993,
	}

	portsKilled := 0
	for _, port := range ports {
		if err := gops.KillProcessByPort(port); err == nil {
			log.Info("✅ Stopped service on port", "port", port)
			portsKilled++
		}
	}

	log.Info("📝 Shutting down goreman-supervised processes...")
	processNames := []string{
		"infra",
		"caddy",
		"bento",
		"deck",
		"nats-server",
		"pocketbase",
		"xtemplate",
		"nats-s3",
		"mox",
	}

	processesKilled := 0
	for _, name := range processNames {
		if err := gops.KillProcessByName(name); err == nil {
			log.Info("✅ Stopped process", "name", name)
			processesKilled++
		}
	}

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
}

// RunContainer builds and runs the infrastructure container using ko and Docker.
func RunContainer(environment string) error {
	log.Info("🐳 Building and running containerized service...")

	log.Info("🚀 Ensuring NATS cluster is running...")
	ctx := context.Background()
	if err := nats.StartLocalCluster(ctx); err != nil {
		log.Warn("Failed to start NATS cluster, continuing anyway", "error", err)
	}

	imageName := config.GetDockerImageFullName()

	log.Info("📦 Building container image with ko...")
	koPath := config.GetKoBinPath()
	if _, err := os.Stat(koPath); err != nil {
		return fmt.Errorf("ko binary not found at %s. Run 'go run . dep install ko' first", koPath)
	}

	os.Setenv("KO_DOCKER_REPO", "ko.local")
	if environment == "production" || config.IsProduction() {
		os.Setenv("ENVIRONMENT", "production")
	} else {
		os.Setenv("ENVIRONMENT", "development")
	}

	commit := getGitCommit()
	if commit == "" {
		commit = "dev"
	}

	buildCmd := exec.Command(koPath, "build", "--push=false", "--platform=linux/amd64", "github.com/joeblew999/infra")
	buildCmd.Env = append(os.Environ(), "GIT_HASH="+commit)

	output, err := buildCmd.CombinedOutput()
	if err != nil {
		log.Error("ko build failed", "error", err, "output", string(output))
		return fmt.Errorf("failed to build container image: %w", err)
	}

	log.Info("✅ Built container image with ko")

	listCmd := exec.Command("docker", "images", "--format", "{{.Repository}}:{{.Tag}}", "--filter", "reference=ko.local/*", "--no-trunc")
	listOutput, err := listCmd.Output()
	if err != nil {
		log.Warn("Failed to list Docker images", "error", err)
		imageName = "ko.local/infra-bc4829dfbf7b0b49d219aad7c8cfa3f9:latest"
	} else {
		lines := strings.Split(strings.TrimSpace(string(listOutput)), "\n")
		for _, line := range lines {
			if strings.Contains(line, "ko.local/infra") && (strings.HasSuffix(line, ":latest") || strings.Contains(line, ":")) {
				tagCmd := exec.Command("docker", "tag", line, imageName)
				if err := tagCmd.Run(); err != nil {
					log.Warn("Failed to tag image", "from", line, "to", imageName, "error", err)
					imageName = line
				} else {
					log.Info("✅ Tagged image", "from", line, "to", imageName)
				}
				break
			}
		}
	}

	log.Info("✅ Container image ready", "image", imageName)

	log.Info("🧹 Stopping any existing containers on conflicting ports or with same name...")
	stopNameCmd := exec.Command("docker", "stop", "infra-service")
	if stopNameCmd.Run() == nil {
		log.Info("✅ Stopped existing container by name", "name", "infra-service")
	}

	ports := []string{
		config.GetWebServerPort(),
		config.GetNATSPort(),
		config.GetPocketBasePort(),
		config.GetBentoPort(),
		config.GetDeckAPIPort(),
		config.GetCaddyPort(),
		config.GetXTemplatePort(),
		"443",
	}

	stoppedContainers := 0
	for _, port := range ports {
		psCmd := exec.Command("docker", "ps", "--filter", fmt.Sprintf("publish=%s", port), "--format", "{{.ID}}")
		if output, err := psCmd.Output(); err == nil {
			containerIDs := strings.Fields(strings.TrimSpace(string(output)))
			for _, containerID := range containerIDs {
				if containerID == "" {
					continue
				}
				stopCmd := exec.Command("docker", "stop", containerID)
				if stopCmd.Run() == nil {
					stoppedContainers++
					log.Info("✅ Stopped existing container", "container_id", containerID, "port", port)
				}
			}
		}
	}

	if stoppedContainers > 0 {
		log.Info("🧹 Cleanup complete", "containers_stopped", stoppedContainers)
		time.Sleep(1 * time.Second)
	}

	cwd, _ := os.Getwd()
	dataDir := filepath.Join(cwd, ".data")
	_ = os.MkdirAll(dataDir, 0755)

	portMappings := []string{
		"-p", fmt.Sprintf("%s:%s", config.GetWebServerPort(), config.GetWebServerPort()),
		"-p", fmt.Sprintf("%s:%s", config.GetNATSPort(), config.GetNATSPort()),
		"-p", fmt.Sprintf("%s:%s", config.GetPocketBasePort(), config.GetPocketBasePort()),
		"-p", fmt.Sprintf("%s:%s", config.GetBentoPort(), config.GetBentoPort()),
		"-p", fmt.Sprintf("%s:%s", config.GetDeckAPIPort(), config.GetDeckAPIPort()),
		"-p", fmt.Sprintf("%s:%s", config.GetXTemplatePort(), config.GetXTemplatePort()),
		"-p", fmt.Sprintf("%s:%s", config.GetCaddyPort(), config.GetCaddyPort()),
		"-p", "443:443",
	}

	args := []string{"run", "--rm", "--name", "infra-service"}
	args = append(args, portMappings...)
	args = append(args,
		"-v", fmt.Sprintf("%s:/app/.data", dataDir),
		"-e", fmt.Sprintf("ENVIRONMENT=%s", environment),
		imageName,
		"service",
	)

	log.Info("🚀 Starting container...", "image", imageName, "data_dir", dataDir)
	log.Info("📝 Container ports mapped using config functions")

	runCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(c)

	go func() {
		select {
		case <-runCtx.Done():
			return
		case <-c:
			log.Info("🛑 Received shutdown signal, stopping container...")
			cancel()
		}
	}()

	dockerCmd := exec.CommandContext(runCtx, "docker", args...)
	dockerCmd.Stdout = os.Stdout
	dockerCmd.Stderr = os.Stderr
	dockerCmd.Stdin = os.Stdin

	log.Info("💡 Container accessible at the same URLs as direct mode")
	if err := dockerCmd.Run(); err != nil {
		if runCtx.Err() != nil {
			log.Info("✅ Container stopped gracefully")
			return nil
		}
		return fmt.Errorf("docker run failed: %w", err)
	}

	return nil
}

func portToInt(portStr string) int {
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return 0
	}
	return port
}

func getGitCommit() string {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}
