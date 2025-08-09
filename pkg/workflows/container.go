package workflows

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/joeblew999/infra/pkg/config"
	"github.com/joeblew999/infra/pkg/log"
)

// ContainerBuildOptions configures container build behavior
type ContainerBuildOptions struct {
	Push     bool
	Platform string
	Repo     string
	Tag      string
	DryRun   bool
}

// ContainerBuildWorkflow handles standardized container image builds
type ContainerBuildWorkflow struct {
	opts ContainerBuildOptions
}

// NewContainerBuildWorkflow creates a new container build workflow
func NewContainerBuildWorkflow(opts ContainerBuildOptions) *ContainerBuildWorkflow {
	// Set defaults
	if opts.Platform == "" {
		opts.Platform = config.PlatformLinuxAmd64
	}
	if opts.Repo == "" {
		opts.Repo = config.GetKoDockerRepo()
	}
	if opts.Tag == "" {
		opts.Tag = "latest"
	}

	return &ContainerBuildWorkflow{opts: opts}
}

// Execute runs the container build workflow
func (b *ContainerBuildWorkflow) Execute() (string, error) {
	log.Info("Starting container build workflow", 
		"push", b.opts.Push,
		"platform", b.opts.Platform,
		"repo", b.opts.Repo,
		"tag", b.opts.Tag,
		"dry_run", b.opts.DryRun)

	if b.opts.DryRun {
		image := b.opts.Repo + ":" + b.opts.Tag
		log.Info("[DRY RUN] Would build image", "image", image)
		return image, nil
	}

	// Authenticate with registry if pushing
	if b.opts.Push {
		if err := runBinary(config.GetFlyctlBinPath(), "auth", "docker"); err != nil {
			return "", fmt.Errorf("failed to authenticate with registry: %w", err)
		}
	}

	// Set ko environment
	if b.opts.Push {
		os.Setenv("KO_DOCKER_REPO", b.opts.Repo)
	} else {
		// Use dummy repo for local builds
		os.Setenv("KO_DOCKER_REPO", "unused")
	}
	
	if config.IsProduction() {
		os.Setenv("ENVIRONMENT", "production")
	} else {
		os.Setenv("ENVIRONMENT", "development")
	}

	// Build image - use direct ko binary path
	koPath := ".dep/ko"
	if _, err := os.Stat(koPath); os.IsNotExist(err) {
		koPath = ".dep/ko_darwin_arm64"
	}
	
	// Build the command
	args := []string{
		"build",
		"--platform=" + b.opts.Platform,
		"--bare",
		"--tags=latest",
	}
	
	if !b.opts.Push {
		// For local builds, use oci-layout and avoid registry
		buildPath := config.GetBuildPath()
		args = append(args, "--push=false", "--oci-layout-path="+buildPath)
		os.Setenv("KO_DOCKER_REPO", "unused")
	} else {
		// For registry push
		args = append(args, "--push=true")
		os.Setenv("KO_DOCKER_REPO", b.opts.Repo)
	}
	
	args = append(args, "github.com/joeblew999/infra")
	
	// Execute the command
	cmd := exec.Command(koPath, args...)
	// Create temporary empty config to avoid .ko.yaml conflicts
	tmpConfig := "/tmp/.ko.yaml"
	if _, err := os.Stat(tmpConfig); os.IsNotExist(err) {
		os.WriteFile(tmpConfig, []byte(""), 0644)
	}
	cmd.Env = append(os.Environ(), "KO_CONFIG_PATH=/tmp")
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("ko build failed: %w\nOutput: %s", err, string(output))
	}

	var image string
	if b.opts.Push {
		image = fmt.Sprintf("%s:%s", b.opts.Repo, b.opts.Tag)
	} else {
		image = filepath.Join(config.GetBuildPath(), "@sha256:latest")
	}
	log.Info("Built container image", "image", image)

	return image, nil
}

// pushImage is now handled by the build command directly