package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/joeblew999/infra/pkg/config"
	"github.com/joeblew999/infra/pkg/dep"
	"github.com/joeblew999/infra/pkg/log"
	"github.com/joeblew999/infra/pkg/workflows"
)

// RunDevelopmentPreflight performs preflight checks and development-time tasks
// when the application is run in a development environment.
func RunDevelopmentPreflight(ctx context.Context) {
	// Check if we are in a development environment and if it's a service startup command
	// This heuristic assumes 'go run .' or 'go run . service' implies development startup.
	// We avoid running generation for CLI commands like 'go run . cli gozero ...'
	isDevelopment := config.IsDevelopment()

	// Only run preflight for actual service startup, not CLI commands
	isServiceStartup := false
	if len(os.Args) == 1 {
		// 'go run .' - default service startup
		isServiceStartup = true
	} else if len(os.Args) > 1 && os.Args[1] == "service" {
		// 'go run . service' - explicit service startup
		isServiceStartup = true
	}
	// For any other commands (cli, deploy, config, etc.) - do NOT run preflight

	if isDevelopment && isServiceStartup {
		log.Info("🚀 Ensuring Go-Zero API code is up-to-date for development startup...")

		// Discover and generate all API services
		if err := generateAllApiServices(ctx); err != nil {
			log.Error("❌ Failed to generate Go-Zero API code", "error", err)
			os.Exit(1)
		}
		log.Info("✅ Go-Zero API code generation complete.")
	}
}

// RunDevelopmentPreflightIfNeeded performs preflight checks for service startup.
// This should only be called from service command paths (RunService).
func RunDevelopmentPreflightIfNeeded(ctx context.Context) {
	// Only run preflight when EXPLICITLY in development mode
	// This prevents production binaries from accidentally running code generation
	if !isExplicitDevelopmentMode() {
		return
	}

	// Run preflight for service startup (including default 'go run .')
	log.Info("🚀 Ensuring code generation is up-to-date for development startup...")

	// Generate binary constants from dep.json first
	if err := generateBinaryConstants(); err != nil {
		log.Error("❌ Failed to generate binary constants", "error", err)
		os.Exit(1)
	}
	log.Info("✅ Binary constants generation complete.")

	// Ensure required dependencies are installed
	if err := ensurePreflightDependencies(ctx); err != nil {
		log.Error("❌ Failed to ensure preflight dependencies", "error", err)
		os.Exit(1)
	}

	// Generate all API services
	if err := generateAllApiServices(ctx); err != nil {
		log.Error("❌ Failed to generate Go-Zero API code", "error", err)
		os.Exit(1)
	}
	log.Info("✅ Go-Zero API code generation complete.")

	log.Info("🎉 All code generation complete!")
}

// ensurePreflightDependencies ensures all required dependencies are installed for preflight
func ensurePreflightDependencies(_ context.Context) error {
	// Only goctl is needed for go-zero code generation
	// It builds from source so no external dependencies required
	log.Info("Ensuring goctl is available for go-zero code generation")
	if err := dep.InstallBinary("goctl", false); err != nil {
		return fmt.Errorf("failed to install goctl: %w", err)
	}
	log.Info("goctl ready for code generation")

	return nil
}

// generateBinaryConstants generates pkg/config/binaries_gen.go from pkg/dep/dep.json
func generateBinaryConstants() error {
	log.Info("Generating binary constants from dep.json")

	// Run go generate in the config package
	cmd := exec.Command("go", "generate", "./pkg/config")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run go generate ./pkg/config: %w", err)
	}

	return nil
}

// isExplicitDevelopmentMode returns true only when we're EXPLICITLY in development mode.
// This is more conservative than config.IsDevelopment() which defaults to true.
func isExplicitDevelopmentMode() bool {
	env := os.Getenv("ENVIRONMENT")

	// Only run preflight when explicitly set to development
	if env == "development" {
		return true
	}

	// Also run when using 'go run .' (development workflow)
	// Check if we're running via 'go run' by looking at the executable path
	if len(os.Args) > 0 {
		executable := os.Args[0]
		// 'go run .' creates temp executables with specific patterns
		if strings.Contains(executable, "/go-build/") ||
		   strings.Contains(executable, "\\go-build\\") ||
		   strings.HasSuffix(executable, ".test") {
			return true
		}
	}

	return false
}

// generateAllApiServices generates go-zero code for all configured API services.
// Services are defined in pkg/config and can be overridden via API_SERVICES env var.
func generateAllApiServices(ctx context.Context) error {
	// Get configured API services from config
	apiServices := config.GetAPIServices()

	if len(apiServices) == 0 {
		log.Info("No API services configured, skipping go-zero generation")
		return nil
	}

	// Generate each configured service
	for _, serviceDir := range apiServices {
		// Construct the .api file path
		apiFileName := filepath.Base(serviceDir) + ".api"
		apiFilePath := filepath.Join(serviceDir, apiFileName)

		// Check if the .api file exists
		if _, err := os.Stat(apiFilePath); os.IsNotExist(err) {
			log.Info("API file not found, skipping", "service", serviceDir, "file", apiFilePath)
			continue
		}

		log.Info("Generating API service", "service", serviceDir, "apiFile", apiFilePath)

		// Generate go-zero code for this service
		if err := workflows.GenerateGoZeroCode(ctx, apiFilePath, serviceDir); err != nil {
			log.Error("Failed to generate service", "service", serviceDir, "error", err)
			return err
		}
	}

	return nil
}
