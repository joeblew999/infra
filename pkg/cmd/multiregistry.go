package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/joeblew999/infra/pkg/workflows"
	"github.com/joeblew999/infra/pkg/log"
)

var multiRegistryCmd = &cobra.Command{
	Use:   "multiregistry",
	Short: "Build and push container images to multiple registries",
	Long: `Build container images using Ko and push to multiple registries for redundancy.
	
Supports GitHub Container Registry (GHCR) and Fly.io registry with proper credential management.`,
	Example: `  # Build and push to both GHCR and Fly.io registry
  infra multiregistry --ghcr --fly

  # Build and push only to GHCR (recommended)
  infra multiregistry --ghcr

  # Dry run to see what would happen
  infra multiregistry --ghcr --dry-run

  # Build with custom app name
  infra multiregistry --ghcr --app my-custom-app`,
	RunE: runMultiRegistryBuild,
}

var (
	multiRegistryPushGHCR        bool
	multiRegistryPushFly         bool
	multiRegistryDryRun          bool
	multiRegistryAppName         string
	multiRegistryGitHash         string
	multiRegistryEnvironment     string
)

func init() {
	rootCmd.AddCommand(multiRegistryCmd)
	
	multiRegistryCmd.Flags().BoolVar(&multiRegistryPushGHCR, "ghcr", false, "Push to GitHub Container Registry")
	multiRegistryCmd.Flags().BoolVar(&multiRegistryPushFly, "fly", false, "Push to Fly.io registry")
	multiRegistryCmd.Flags().BoolVar(&multiRegistryDryRun, "dry-run", false, "Show what would be done without executing")
	multiRegistryCmd.Flags().StringVar(&multiRegistryAppName, "app", "", "Fly.io app name (default: from FLY_APP_NAME or infra-mgmt)")
	multiRegistryCmd.Flags().StringVar(&multiRegistryGitHash, "git-hash", "", "Git hash to inject (default: auto-detected)")
	multiRegistryCmd.Flags().StringVar(&multiRegistryEnvironment, "environment", "", "Environment: production or development (default: auto-detected)")
}

func runMultiRegistryBuild(cmd *cobra.Command, args []string) error {
	// Validation
	if !multiRegistryPushGHCR && !multiRegistryPushFly {
		return fmt.Errorf("must specify at least one registry: --ghcr or --fly")
	}

	// Create workflow options
	opts := workflows.MultiRegistryBuildOptions{
		GitHash:           multiRegistryGitHash,
		Environment:       multiRegistryEnvironment,
		PushToGHCR:        multiRegistryPushGHCR,
		PushToFlyRegistry: multiRegistryPushFly,
		DryRun:            multiRegistryDryRun,
		AppName:           multiRegistryAppName,
	}

	// Create and execute workflow
	workflow := workflows.NewMultiRegistryBuildWorkflow(opts)

	log.Info("Starting multi-registry build workflow")
	
	// Check credentials first
	if err := workflow.CheckCredentials(); err != nil {
		if multiRegistryDryRun {
			log.Warn("Credential check failed (dry run mode)", "error", err)
		} else {
			return fmt.Errorf("credential check failed: %w", err)
		}
	}

	// Execute the workflow
	if err := workflow.Execute(); err != nil {
		return fmt.Errorf("multi-registry build failed: %w", err)
	}

	log.Info("Multi-registry build completed successfully")
	return nil
}