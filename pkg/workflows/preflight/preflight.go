package preflight

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

// RunDevelopment executes preflight checks and development-time tasks for standalone runs.
func RunDevelopment(ctx context.Context) {
	isDev := config.IsDevelopment()
	isServiceStartup := len(os.Args) == 1 || (len(os.Args) > 1 && os.Args[1] == "service")
	if !(isDev && isServiceStartup) {
		return
	}

	log.Info("🚀 Ensuring Go-Zero API code is up-to-date for development startup...")
	if err := generateAllAPIServices(ctx); err != nil {
		log.Error("❌ Failed to generate Go-Zero API code", "error", err)
		os.Exit(1)
	}
	log.Info("✅ Go-Zero API code generation complete.")
}

// RunIfNeeded performs preflight checks during service startup when explicitly in development.
func RunIfNeeded(ctx context.Context) {
	if !isExplicitDevelopmentMode() {
		return
	}

	log.Info("🚀 Ensuring code generation is up-to-date for development startup...")

	if err := generateBinaryConstants(); err != nil {
		log.Error("❌ Failed to generate binary constants", "error", err)
		os.Exit(1)
	}
	log.Info("✅ Binary constants generation complete.")

	if err := ensureDependencies(ctx); err != nil {
		log.Error("❌ Failed to ensure preflight dependencies", "error", err)
		os.Exit(1)
	}

	if err := generateAllAPIServices(ctx); err != nil {
		log.Error("❌ Failed to generate Go-Zero API code", "error", err)
		os.Exit(1)
	}
	log.Info("✅ Go-Zero API code generation complete.")

	log.Info("🎉 All code generation complete!")
}

func ensureDependencies(_ context.Context) error {
	log.Info("Ensuring goctl is available for go-zero code generation")
	if err := dep.InstallBinary("goctl", false); err != nil {
		return fmt.Errorf("failed to install goctl: %w", err)
	}
	log.Info("goctl ready for code generation")
	return nil
}

func generateBinaryConstants() error {
	log.Info("Generating binary constants from dep.json")
	cmd := exec.Command("go", "generate", "./pkg/config")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run go generate ./pkg/config: %w", err)
	}
	return nil
}

func isExplicitDevelopmentMode() bool {
	env := os.Getenv("ENVIRONMENT")
	if env == "development" {
		return true
	}
	if len(os.Args) > 0 {
		executable := os.Args[0]
		if strings.Contains(executable, "/go-build/") ||
			strings.Contains(executable, "\\go-build\\") ||
			strings.HasSuffix(executable, ".test") {
			return true
		}
	}
	return false
}

func generateAllAPIServices(ctx context.Context) error {
	apiServices := config.GetAPIServices()
	if len(apiServices) == 0 {
		log.Info("No API services configured, skipping go-zero generation")
		return nil
	}

	for _, serviceDir := range apiServices {
		apiFileName := filepath.Base(serviceDir) + ".api"
		apiFilePath := filepath.Join(serviceDir, apiFileName)

		if _, err := os.Stat(apiFilePath); os.IsNotExist(err) {
			log.Info("API file not found, skipping", "service", serviceDir, "file", apiFilePath)
			continue
		}

		log.Info("Generating API service", "service", serviceDir, "apiFile", apiFilePath)
		if err := workflows.GenerateGoZeroCode(ctx, apiFilePath, serviceDir); err != nil {
			log.Error("Failed to generate service", "service", serviceDir, "error", err)
			return err
		}
	}

	return nil
}
