package builders

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/joeblew999/infra/pkg/config"
	"github.com/joeblew999/infra/pkg/log"
)

// NPMInstaller handles NPM package installations
type NPMInstaller struct{}

func (i *NPMInstaller) Install(name, repo, pkg, version string, debug bool) error {
	log.Info("Installing NPM package", "package", pkg, "version", version)

	// Ensure bun is available through the dep system
	bunPath, err := i.ensureBun(debug)
	if err != nil {
		return fmt.Errorf("failed to ensure bun is available: %w", err)
	}

	// Ensure .dep directory exists
	depPath := config.GetDepPath()
	if err := os.MkdirAll(depPath, 0755); err != nil {
		return fmt.Errorf("failed to create .dep directory: %w", err)
	}

	// Use bun install to install the package globally to .dep
	// For claude, this will install @anthropic-ai/claude-code
	if version != "latest" {
		pkg = fmt.Sprintf("%s@%s", pkg, version)
	}

	// Install to .dep directory using bun
	// Create a package.json for the install
	packageJSON := fmt.Sprintf(`{"name": "%s-install", "version": "1.0.0"}`, name)
	if err := os.WriteFile(filepath.Join(depPath, "package.json"), []byte(packageJSON), 0644); err != nil {
		return fmt.Errorf("failed to create package.json: %w", err)
	}

	cmd := exec.Command(bunPath, "add", pkg)
	cmd.Dir = depPath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install %s via npm: %w", name, err)
	}

	// Find the binary in the bun node_modules/.bin directory
	binaryPath := filepath.Join(depPath, "node_modules", ".bin", name)
	if runtime.GOOS == "windows" {
		binaryPath += ".cmd"
	}

	// Ensure executable permissions
	if err := os.Chmod(binaryPath, 0755); err != nil {
		return fmt.Errorf("failed to set executable permissions: %w", err)
	}

	log.Info("Successfully installed NPM package", "name", name, "path", binaryPath)
	return nil
}

// ensureBun ensures bun binary is available and returns its path
func (i *NPMInstaller) ensureBun(debug bool) (string, error) {
	// Get the expected bun path from config
	bunPath := config.Get("bun")
	
	// Check if bun binary exists
	if _, err := os.Stat(bunPath); err == nil {
		// Binary exists, return the path
		return bunPath, nil
	}
	
	// Bun not found in .dep, check if it's available in PATH
	if bunInPath, err := exec.LookPath("bun"); err == nil {
		// Bun is available in PATH, use it
		log.Info("Using bun from PATH", "path", bunInPath)
		return bunInPath, nil
	}
	
	// Bun not found anywhere, we need to install it
	log.Info("Bun not found, installing bun first...")
	
	// We need to install bun, but this creates a circular dependency.
	// The solution is to use the binary installers interface to install bun
	// without going through the NPM installer.
	if err := i.installBunDependency(debug); err != nil {
		return "", fmt.Errorf("failed to install bun dependency: %w", err)
	}
	
	// Now check if bun is available
	if _, err := os.Stat(bunPath); err == nil {
		return bunPath, nil
	}
	
	return "", fmt.Errorf("bun installation completed but binary not found at expected path: %s", bunPath)
}

// installBunDependency installs bun using the github-release installer
// This avoids circular dependency with NPM installer
func (i *NPMInstaller) installBunDependency(_ bool) error {
	// Import is avoided to prevent circular dependency
	// Instead, we'll suggest the user to install bun manually first
	return fmt.Errorf("bun is required but not available. Please install bun first by running: go run . dep install bun")
}