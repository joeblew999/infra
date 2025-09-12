package workflows

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/joeblew999/infra/pkg/config"
	"github.com/joeblew999/infra/pkg/log"
)

// MultiRegistryBuildOptions configures multi-registry build behavior
type MultiRegistryBuildOptions struct {
	GitHash           string
	Environment       string
	PushToGHCR        bool
	PushToFlyRegistry bool
	DryRun            bool
	AppName           string
}

// MultiRegistryBuildWorkflow handles building and pushing to multiple registries
type MultiRegistryBuildWorkflow struct {
	opts MultiRegistryBuildOptions
}

// NewMultiRegistryBuildWorkflow creates a new multi-registry build workflow
func NewMultiRegistryBuildWorkflow(opts MultiRegistryBuildOptions) *MultiRegistryBuildWorkflow {
	// Set defaults
	if opts.GitHash == "" {
		opts.GitHash = config.GetRuntimeGitHash()
		if opts.GitHash == "" {
			opts.GitHash = "dev"
		}
	}
	if opts.Environment == "" {
		if config.IsProduction() {
			opts.Environment = "production"
		} else {
			opts.Environment = "development"
		}
	}
	if opts.AppName == "" {
		opts.AppName = getEnvOrDefault("FLY_APP_NAME", "infra-mgmt")
	}

	return &MultiRegistryBuildWorkflow{opts: opts}
}

// Execute runs the multi-registry build and push workflow
func (m *MultiRegistryBuildWorkflow) Execute() error {
	log.Info("Starting multi-registry build workflow",
		"git_hash", m.opts.GitHash,
		"environment", m.opts.Environment,
		"push_ghcr", m.opts.PushToGHCR,
		"push_fly", m.opts.PushToFlyRegistry,
		"dry_run", m.opts.DryRun)

	shortHash := m.opts.GitHash
	if len(shortHash) > 7 {
		shortHash = shortHash[:7]
	}

	// Registry configurations
	ghcrRepo := "ghcr.io/joeblew999/infra"
	flyRepo := fmt.Sprintf("registry.fly.io/%s", m.opts.AppName)

	var errors []string

	// Build and push to GHCR
	if m.opts.PushToGHCR {
		// Authenticate with GHCR first
		if err := m.authenticateGHCR(); err != nil {
			log.Warn("GHCR authentication failed, Ko will attempt direct push", "error", err)
		}
		
		if err := m.buildAndPushToRegistry(ghcrRepo, []string{"latest", shortHash}); err != nil {
			errorMsg := fmt.Sprintf("GHCR push failed: %v", err)
			log.Error("GHCR push failed", "error", err)
			errors = append(errors, errorMsg)
		} else {
			log.Info("✅ Successfully pushed to GHCR", "repo", ghcrRepo)
		}
	}

	// Build and push to Fly.io registry
	if m.opts.PushToFlyRegistry {
		if err := m.authenticateFlyRegistry(); err != nil {
			errorMsg := fmt.Sprintf("Fly registry auth failed: %v", err)
			log.Error("Fly registry authentication failed", "error", err)
			errors = append(errors, errorMsg)
		} else {
			if err := m.buildAndPushToRegistry(flyRepo, []string{"latest", shortHash}); err != nil {
				errorMsg := fmt.Sprintf("Fly registry push failed: %v", err)
				log.Error("Fly registry push failed", "error", err)
				errors = append(errors, errorMsg)
			} else {
				log.Info("✅ Successfully pushed to Fly registry", "repo", flyRepo)
			}
		}
	}

	// Report results - succeed if at least one registry worked
	if len(errors) > 0 {
		totalRegistries := 0
		if m.opts.PushToGHCR {
			totalRegistries++
		}
		if m.opts.PushToFlyRegistry {
			totalRegistries++
		}
		
		if len(errors) == totalRegistries {
			// All registries failed
			log.Error("All registries failed", "errors", errors)
			return fmt.Errorf("all registry builds failed: %s", strings.Join(errors, "; "))
		} else {
			// Some registries succeeded
			log.Warn("Multi-registry build completed with partial failures", "errors", errors)
		}
	}

	log.Info("🎉 Multi-registry build completed successfully",
		"ghcr_repo", ghcrRepo,
		"fly_repo", flyRepo,
		"tags", []string{"latest", shortHash})
	return nil
}

// buildAndPushToRegistry builds and pushes to a specific registry
func (m *MultiRegistryBuildWorkflow) buildAndPushToRegistry(repo string, tags []string) error {
	for _, tag := range tags {
		if err := m.buildAndPushWithTag(repo, tag); err != nil {
			return fmt.Errorf("failed to push %s:%s: %w", repo, tag, err)
		}
		log.Info("Successfully pushed image", "repo", repo, "tag", tag)
	}
	return nil
}

// buildAndPushWithTag builds and pushes a single image with specific tag
func (m *MultiRegistryBuildWorkflow) buildAndPushWithTag(repo, tag string) error {
	log.Info("Building and pushing image", "repo", repo, "tag", tag)

	if m.opts.DryRun {
		log.Info("[DRY RUN] Would build and push", "repo", repo, "tag", tag)
		return nil
	}

	// Set environment for Ko build
	env := append(os.Environ(),
		fmt.Sprintf("KO_DOCKER_REPO=%s", repo),
		fmt.Sprintf("GIT_HASH=%s", m.opts.GitHash),
		fmt.Sprintf("ENVIRONMENT=%s", m.opts.Environment),
	)
	
	// Add authentication tokens for ko to use
	if githubToken := m.getGitHubToken(); githubToken != "" {
		env = append(env, fmt.Sprintf("GITHUB_TOKEN=%s", githubToken))
	}
	if flyToken := m.getFlyToken(); flyToken != "" {
		env = append(env, fmt.Sprintf("FLY_API_TOKEN=%s", flyToken))
	}

	// Build and push with Ko
	args := []string{
		"build",
		"--push=true",
		"--bare",
		fmt.Sprintf("--tags=%s", tag),
		"github.com/joeblew999/infra",
	}

	cmd := exec.Command(config.GetKoBinPath(), args...)
	cmd.Env = env
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ko build/push failed: %w", err)
	}

	return nil
}

// authenticateGHCR authenticates with GitHub Container Registry
func (m *MultiRegistryBuildWorkflow) authenticateGHCR() error {
	if m.opts.DryRun {
		log.Info("[DRY RUN] Would authenticate with GHCR")
		return nil
	}

	token := m.getGitHubToken()
	if token == "" {
		return fmt.Errorf("GITHUB_TOKEN not set")
	}

	log.Info("Authenticating with GitHub Container Registry (ko will use GITHUB_TOKEN directly)")
	
	// Ko handles authentication directly via GITHUB_TOKEN environment variable
	// No need for docker login - ko has built-in registry authentication
	return nil
}

// authenticateFlyRegistry authenticates with Fly.io registry
func (m *MultiRegistryBuildWorkflow) authenticateFlyRegistry() error {
	if m.opts.DryRun {
		log.Info("[DRY RUN] Would authenticate with Fly registry")
		return nil
	}

	token := m.getFlyToken()
	log.Info("Authenticating with Fly.io registry")

	// Ko can read ~/.docker/config.json for registry authentication
	// We just need to ensure flyctl has configured the registry credentials
	if token != "" {
		// If we have a token, run flyctl auth docker to configure ~/.docker/config.json
		cmd := exec.Command(config.GetFlyctlBinPath(), "auth", "docker")
		env := append(os.Environ(), "FLY_ACCESS_TOKEN="+token)
		cmd.Env = env
		
		// Run silently - we don't need to see the output
		if err := cmd.Run(); err != nil {
			log.Warn("Failed to configure Fly registry credentials", "error", err)
			// Don't fail - ko might still work with existing credentials
		}
	}
	
	// Ko will use the credentials from ~/.docker/config.json
	return nil
}

// CheckCredentials verifies that required credentials are available
func (m *MultiRegistryBuildWorkflow) CheckCredentials() error {
	var missing []string

	if m.opts.PushToGHCR {
		if m.getGitHubToken() == "" {
			missing = append(missing, "GITHUB_TOKEN (for GHCR)")
		}
	}

	if m.opts.PushToFlyRegistry {
		if m.getFlyToken() == "" {
			missing = append(missing, "FLY_API_TOKEN or FLY_ACCESS_TOKEN (for Fly registry)")
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required credentials: %s", strings.Join(missing, ", "))
	}

	return nil
}

// isFlyctlAuthenticated checks if flyctl is authenticated
func (m *MultiRegistryBuildWorkflow) isFlyctlAuthenticated() bool {
	cmd := exec.Command(config.GetFlyctlBinPath(), "auth", "whoami")
	return cmd.Run() == nil
}

// getFlyToken gets FLY token from environment or flyctl
func (m *MultiRegistryBuildWorkflow) getFlyToken() string {
	// Check environment variables first
	token := os.Getenv("FLY_API_TOKEN")
	if token == "" {
		token = os.Getenv("FLY_ACCESS_TOKEN")
	}
	
	// If no env token, try to get from flyctl
	if token == "" && m.isFlyctlAuthenticated() {
		cmd := exec.Command(config.GetFlyctlBinPath(), "auth", "token")
		if output, err := cmd.Output(); err == nil {
			token = strings.TrimSpace(string(output))
			// Cache it in environment for this session
			os.Setenv("FLY_API_TOKEN", token)
		}
	}
	
	return token
}

// getGitHubToken gets GitHub token from environment or gh CLI
func (m *MultiRegistryBuildWorkflow) getGitHubToken() string {
	// Check environment variable first
	token := os.Getenv("GITHUB_TOKEN")
	
	// If no env token, try to get from gh CLI
	if token == "" {
		ghPath := config.GetGhBinPath()
		cmd := exec.Command(ghPath, "auth", "token")
		if output, err := cmd.Output(); err == nil {
			token = strings.TrimSpace(string(output))
			// Cache it in environment for this session
			os.Setenv("GITHUB_TOKEN", token)
		}
	}
	
	return token
}