package workflows

import (
	"fmt"
	"os"

	"github.com/joeblew999/infra/pkg/config"
	"github.com/joeblew999/infra/pkg/log"
)

// BuildOptions configures build behavior
type BuildOptions struct {
	Push     bool
	Platform string
	Repo     string
	Tag      string
	DryRun   bool
}

// BuildWorkflow handles standardized container image builds
type BuildWorkflow struct {
	opts BuildOptions
}

// NewBuildWorkflow creates a new build workflow
func NewBuildWorkflow(opts BuildOptions) *BuildWorkflow {
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

	return &BuildWorkflow{opts: opts}
}

// Execute runs the build workflow
func (b *BuildWorkflow) Execute() (string, error) {
	log.Info("Starting build workflow", 
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
	os.Setenv("KO_DOCKER_REPO", b.opts.Repo)
	if config.IsProduction() {
		os.Setenv("ENVIRONMENT", "production")
	} else {
		os.Setenv("ENVIRONMENT", "development")
	}

	// Build image
	image, err := runBinaryWithOutput(config.GetKoBinPath(), 
		"build", 
		"--platform="+b.opts.Platform,
		"--image-refs=/dev/null", // Don't write refs file
		"github.com/joeblew999/infra")
	if err != nil {
		return "", fmt.Errorf("ko build failed: %w", err)
	}

	image = fmt.Sprintf("%s:%s", b.opts.Repo, b.opts.Tag)
	log.Info("Built container image", "image", image)

	if b.opts.Push {
		return b.pushImage(image)
	}

	return image, nil
}

// pushImage pushes the built image to registry
func (b *BuildWorkflow) pushImage(image string) (string, error) {
	log.Info("Pushing image to registry", "image", image)

	// Tag and push using ko
	_, err := runBinaryWithOutput(config.GetKoBinPath(), 
		"publish", 
		"--platform="+b.opts.Platform,
		"--tag-only",
		"github.com/joeblew999/infra")
	if err != nil {
		return "", fmt.Errorf("ko publish failed: %w", err)
	}

	log.Info("Successfully pushed image", "image", image)
	return image, nil
}