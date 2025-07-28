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

	// Step 5: Build container image
	image, err := d.buildImage()
	if err != nil {
		return fmt.Errorf("image build failed: %w", err)
	}

	// Step 6: Deploy
	if err := d.deploy(image); err != nil {
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

	// Use FLY_API_TOKEN from environment if set
	token := os.Getenv("FLY_API_TOKEN")
	if token == "" {
		return fmt.Errorf("FLY_API_TOKEN environment variable is required")
	}
	
	log.Info("Using FLY_API_TOKEN from environment")
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

// buildImage builds the container image using flyctl deploy
func (d *DeployWorkflow) buildImage() (string, error) {
	log.Info("Building container image with flyctl")
	
	if d.opts.DryRun {
		log.Info("[DRY RUN] Would build container image")
		return "registry.fly.io/" + d.opts.AppName + ":latest", nil
	}

	// Use flyctl deploy with --build-only to build without deploying
	image := fmt.Sprintf("registry.fly.io/%s:latest", d.opts.AppName)
	log.Info("Using flyctl deploy for image building", "image", image)
	
	return image, nil
}

// deploy deploys the application to Fly.io
func (d *DeployWorkflow) deploy(image string) error {
	log.Info("Deploying to Fly.io", "image", image)
	
	if d.opts.DryRun {
		log.Info("[DRY RUN] Would deploy image to Fly.io")
		return nil
	}

	args := []string{"deploy", "-a", d.opts.AppName, "--remote-only"}
	return runBinary(config.GetFlyctlBinPath(), args...)
}

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
	
	// Ensure FLY_API_TOKEN is passed through environment
	token := os.Getenv("FLY_API_TOKEN")
	if token != "" {
		cmd.Env = append(os.Environ(), "FLY_API_TOKEN="+token)
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
	// Ensure FLY_API_TOKEN is passed through environment
	token := os.Getenv("FLY_API_TOKEN")
	if token != "" {
		cmd.Env = append(os.Environ(), "FLY_API_TOKEN="+token)
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