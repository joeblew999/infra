package runtime

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/joeblew999/infra/pkg/config"
	"github.com/joeblew999/infra/pkg/log"
	"github.com/joeblew999/infra/pkg/nats"
)

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
		return fmt.Errorf("ko binary not found at %s. Run 'go run . tools dep install ko' first", koPath)
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

	ports := collectServicePorts(Options{})
	ports = append(ports, ServicePort{Service: "HTTPS Proxy", Port: "443"})

	stoppedContainers := 0
	for _, portSpec := range ports {
		psCmd := exec.Command("docker", "ps", "--filter", fmt.Sprintf("publish=%s", portSpec.Port), "--format", "{{.ID}}")
		if output, err := psCmd.Output(); err == nil {
			containerIDs := strings.Fields(strings.TrimSpace(string(output)))
			for _, containerID := range containerIDs {
				if containerID == "" {
					continue
				}
				stopCmd := exec.Command("docker", "stop", containerID)
				if stopCmd.Run() == nil {
					stoppedContainers++
					log.Info("✅ Stopped existing container", "container_id", containerID, "service", portSpec.Service, "port", portSpec.Port)
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
	_ = os.MkdirAll(dataDir, 0o755)

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

func getGitCommit() string {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}
