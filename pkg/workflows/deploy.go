package workflows

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/joeblew999/infra/pkg/config"
	"github.com/joeblew999/infra/pkg/log"
)

// DeployOptions configures deployment behavior
type DeployOptions struct {
	AppName     string
	Region      string
	Environment string
	DryRun      bool
	Force       bool
}

// DeployWorkflow handles idempotent deployment to Fly.io
type DeployWorkflow struct {
	opts DeployOptions
}

// NewDeployWorkflow creates a new deployment workflow
func NewDeployWorkflow(opts DeployOptions) *DeployWorkflow {
	// Set defaults
	if opts.AppName == "" {
		opts.AppName = getEnvOrDefault("FLY_APP_NAME", "infra-mgmt")
	}
	if opts.Region == "" {
		opts.Region = getEnvOrDefault("FLY_REGION", "syd")
	}
	if opts.Environment == "" {
		if config.IsProduction() {
			opts.Environment = "production"
		} else {
			opts.Environment = "development"
		}
	}

	return &DeployWorkflow{opts: opts}
}

// Execute runs the complete idempotent deployment workflow
func (d *DeployWorkflow) Execute() error {
	log.Info("Starting idempotent deployment workflow", 
		"app", d.opts.AppName, 
		"region", d.opts.Region,
		"environment", d.opts.Environment,
		"dry_run", d.opts.DryRun)

	// Step 0: Bootstrap required environment variables
	if err := config.BootstrapRequiredEnvs(); err != nil {
		return fmt.Errorf("ENV bootstrap failed: %w", err)
	}

	// Step 1: Ensure flyctl is available
	if err := d.ensureFlyctl(); err != nil {
		return fmt.Errorf("flyctl setup failed: %w", err)
	}

	// Step 2: Check authentication
	if err := d.checkAuth(); err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	// Step 3: Ensure app exists
	if err := d.ensureApp(); err != nil {
		return fmt.Errorf("app setup failed: %w", err)
	}

	// Step 4: Ensure volume exists
	if err := d.ensureVolume(); err != nil {
		return fmt.Errorf("volume setup failed: %w", err)
	}

	// Step 5: Build and push container images to multiple registries
	if err := d.buildMultiRegistryImages(); err != nil {
		return fmt.Errorf("multi-registry build failed: %w", err)
	}

	// Step 6: Deploy using GHCR image
	if err := d.deploy(); err != nil {
		return fmt.Errorf("deployment failed: %w", err)
	}

	// Step 7: Verify deployment
	if err := d.verifyDeployment(); err != nil {
		return fmt.Errorf("deployment verification failed: %w", err)
	}

	log.Info("Deployment workflow completed successfully", 
		"app_url", fmt.Sprintf("https://%s.fly.dev", d.opts.AppName))

	return nil
}

// ensureFlyctl ensures flyctl is available
func (d *DeployWorkflow) ensureFlyctl() error {
	log.Info("Ensuring flyctl is available...")
	
	if d.opts.DryRun {
		log.Info("[DRY RUN] Would check flyctl availability")
		return nil
	}

	// This will automatically install if needed via our dep system
	return runBinary(config.GetFlyctlBinPath(), "version")
}

// checkAuth checks if user is authenticated with Fly.io
func (d *DeployWorkflow) checkAuth() error {
	log.Info("Checking Fly.io authentication...")
	
	if d.opts.DryRun {
		log.Info("[DRY RUN] Would check authentication")
		return nil
	}

	// Use FLY_API_TOKEN or FLY_ACCESS_TOKEN from environment
	token := os.Getenv("FLY_API_TOKEN")
	if token == "" {
		token = os.Getenv("FLY_ACCESS_TOKEN")
	}
	
	if token == "" {
		log.Info("No Fly.io token found, attempting to use flyctl authentication...")
		
		// Check if flyctl is already authenticated
		_, err := runBinaryWithOutput(config.GetFlyctlBinPath(), "auth", "whoami")
		if err == nil {
			log.Info("Using existing flyctl authentication")
			return nil
		}
		
		// Try to authenticate with flyctl
		log.Info("Starting flyctl authentication (this will open a browser)...")
		authErr := runBinary(config.GetFlyctlBinPath(), "auth", "login")
		if authErr != nil {
			return fmt.Errorf("authentication failed: %w. Please set FLY_API_TOKEN or FLY_ACCESS_TOKEN environment variable", authErr)
		}
		
		log.Info("Successfully authenticated with flyctl")
		return nil
	}
	
	log.Info("Using Fly.io token from environment")
	return nil
}

// ensureApp ensures the Fly.io app exists
func (d *DeployWorkflow) ensureApp() error {
	log.Info("Ensuring Fly.io app exists", "app", d.opts.AppName)
	
	if d.opts.DryRun {
		log.Info("[DRY RUN] Would ensure app exists")
		return nil
	}

	// Check if app exists - handle both success and error cases
	output, err := runBinaryWithOutput(config.GetFlyctlBinPath(), "apps", "list")
	if err != nil {
		// If we can't list apps, check if it's an auth issue
		if strings.Contains(err.Error(), "must be authenticated") {
			return fmt.Errorf("authentication failed: %w", err)
		}
		// Otherwise, log the error but continue - app might exist
		log.Warn("Could not list apps, will attempt to deploy to existing app", "error", err)
	} else {
		// Successfully listed apps, check if our app exists
		if !strings.Contains(output, d.opts.AppName) {
			log.Info("App doesn't exist, creating...", "app", d.opts.AppName)
			
			args := []string{"apps", "create", d.opts.AppName}
			args = append(args, "--org", "personal") // Default org
			
			createErr := runBinary(config.GetFlyctlBinPath(), args...)
			if createErr != nil && strings.Contains(createErr.Error(), "already been taken") {
				log.Info("App already exists, continuing...", "app", d.opts.AppName)
			} else if createErr != nil {
				return fmt.Errorf("failed to create app: %w", createErr)
			} else {
				// Only set secrets for new apps
				if err := d.setSecrets(); err != nil {
					return fmt.Errorf("failed to set secrets: %w", err)
				}
			}
		} else {
			log.Info("App already exists, skipping creation", "app", d.opts.AppName)
		}
	}

	return nil
}

// setSecrets sets required environment variables
func (d *DeployWorkflow) setSecrets() error {
	log.Info("Setting secrets for new app")
	
	secrets := map[string]string{
		"ENVIRONMENT": d.opts.Environment,
	}

	for key, value := range secrets {
		args := []string{"secrets", "set", fmt.Sprintf("%s=%s", key, value), "-a", d.opts.AppName}
		if err := runBinary(config.GetFlyctlBinPath(), args...); err != nil {
			return fmt.Errorf("failed to set secret %s: %w", key, err)
		}
	}

	return nil
}

// ensureVolume ensures persistent volume exists
func (d *DeployWorkflow) ensureVolume() error {
	log.Info("Ensuring persistent volume exists")
	
	if d.opts.DryRun {
		log.Info("[DRY RUN] Would ensure volume exists")
		return nil
	}

	// Check if volume exists - handle idempotently
	output, err := runBinaryWithOutput(config.GetFlyctlBinPath(), "volumes", "list", "-a", d.opts.AppName)
	if err != nil {
		// Handle auth or other errors gracefully
		if strings.Contains(err.Error(), "must be authenticated") {
			return fmt.Errorf("authentication failed: %w", err)
		}
		log.Warn("Could not list volumes, will attempt to use existing volume", "error", err)
	} else {
		// Successfully listed volumes
		if !strings.Contains(output, "infra_data") {
			log.Info("Creating persistent volume")
			args := []string{"volumes", "create", "infra_data", "--size", "1", "--region", d.opts.Region, "-a", d.opts.AppName, "--yes"}
			createErr := runBinary(config.GetFlyctlBinPath(), args...)
			if createErr != nil && strings.Contains(createErr.Error(), "already exists") {
				log.Info("Volume already exists, continuing...", "volume", "infra_data")
			} else if createErr != nil {
				return fmt.Errorf("failed to create volume: %w", createErr)
			}
		} else {
			log.Info("Volume already exists, skipping creation", "volume", "infra_data")
		}
	}

	return nil
}

// buildMultiRegistryImages builds and pushes images to multiple registries
func (d *DeployWorkflow) buildMultiRegistryImages() error {
	log.Info("Building multi-registry container images")
	
	// Create multi-registry workflow with intelligent fallback
	opts := MultiRegistryBuildOptions{
		GitHash:           config.GetRuntimeGitHash(),
		Environment:       d.opts.Environment,
		PushToGHCR:        true,  // Try GHCR first
		PushToFlyRegistry: true,  // Fallback to Fly registry
		DryRun:            d.opts.DryRun,
		AppName:           d.opts.AppName,
	}
	
	// Check GHCR credentials and fallback if needed
	if os.Getenv("GITHUB_TOKEN") == "" {
		log.Info("GITHUB_TOKEN not available, using Fly registry only")
		opts.PushToGHCR = false
	}
	
	multiRegistry := NewMultiRegistryBuildWorkflow(opts)
	
	// Check credentials before attempting build
	if err := multiRegistry.CheckCredentials(); err != nil {
		log.Warn("Credential check failed, will attempt build anyway", "error", err)
	}
	
	// Execute multi-registry build
	return multiRegistry.Execute()
}

// deploy deploys the application to Fly.io using best available image
func (d *DeployWorkflow) deploy() error {
	// Choose image based on what's available - prefer GHCR if available
	var image string
	if os.Getenv("GITHUB_TOKEN") != "" {
		image = "ghcr.io/joeblew999/infra:latest"
		log.Info("Deploying using GHCR image", "image", image)
	} else {
		image = fmt.Sprintf("registry.fly.io/%s:latest", d.opts.AppName)
		log.Info("Deploying using Fly.io registry image", "image", image)
	}
	
	if d.opts.DryRun {
		log.Info("[DRY RUN] Would deploy with image", "image", image)
		return nil
	}

	args := []string{"deploy", "-a", d.opts.AppName, "--image", image}
	return runBinary(config.GetFlyctlBinPath(), args...)
}

// Note: buildAndPushWithKo removed - now handled by multi-registry script

// verifyDeployment verifies the deployment was successful
func (d *DeployWorkflow) verifyDeployment() error {
	log.Info("Verifying deployment")
	
	if d.opts.DryRun {
		log.Info("[DRY RUN] Would verify deployment")
		return nil
	}

	// Check app status
	return runBinary(config.GetFlyctlBinPath(), "status", "-a", d.opts.AppName)
}

// Helper functions

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func runBinary(path string, args ...string) error {
	log.Debug("Running command", "binary", path, "args", args)
	
	cmd := exec.Command(path, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	// Ensure FLY_API_TOKEN or FLY_ACCESS_TOKEN is passed through environment
	token := os.Getenv("FLY_API_TOKEN")
	if token == "" {
		token = os.Getenv("FLY_ACCESS_TOKEN")
	}
	if token != "" {
		cmd.Env = append(os.Environ(), "FLY_ACCESS_TOKEN="+token)
		// Add --access-token flag if not already present
		hasTokenFlag := false
		for _, arg := range args {
			if arg == "--access-token" {
				hasTokenFlag = true
				break
			}
		}
		if !hasTokenFlag {
			args = append(args, "--access-token", token)
		}
	}
	
	cmd.Args = append([]string{path}, args...)
	return cmd.Run()
}

func runBinaryWithOutput(path string, args ...string) (string, error) {
	log.Debug("Running command with output", "binary", path, "args", args)
	
	cmd := exec.Command(path, args...)
	// Ensure FLY_API_TOKEN or FLY_ACCESS_TOKEN is passed through environment
	token := os.Getenv("FLY_API_TOKEN")
	if token == "" {
		token = os.Getenv("FLY_ACCESS_TOKEN")
	}
	if token != "" {
		cmd.Env = append(os.Environ(), "FLY_ACCESS_TOKEN="+token)
		// Add --access-token flag if not already present
		hasTokenFlag := false
		for _, arg := range args {
			if arg == "--access-token" {
				hasTokenFlag = true
				break
			}
		}
		if !hasTokenFlag {
			args = append(args, "--access-token", token)
		}
	}
	
	cmd.Args = append([]string{path}, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Debug("Command failed", "error", err, "output", string(output))
		return string(output), fmt.Errorf("command failed: %w", err)
	}
	
	return string(output), nil
}