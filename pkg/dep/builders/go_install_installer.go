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

// GoInstallInstaller uses `go install` for packages that support it
type GoInstallInstaller struct{}

func (i *GoInstallInstaller) Install(name, repo, pkg, version string, debug bool) error {
	return i.InstallWithPlatforms(name, repo, pkg, version, debug, nil)
}

func (i *GoInstallInstaller) InstallWithPlatforms(name, repo, pkg, version string, debug bool, platforms []Platform) error {
	if len(platforms) > 1 {
		return fmt.Errorf("cross-platform builds not supported for go-install source type")
	}

	// Use go install to install the package
	packageURL := fmt.Sprintf("%s@%s", pkg, version)
	log.Info("Installing via go install", "package", packageURL)

	// Set GOBIN to our .dep directory so the binary is installed there
	installPath := config.GetDepPath()
	
	// Create temp GOBIN directory
	tempGoBin := filepath.Join(installPath, "temp-gobin")
	if err := os.MkdirAll(tempGoBin, 0755); err != nil {
		return fmt.Errorf("failed to create temp GOBIN directory: %w", err)
	}
	defer os.RemoveAll(tempGoBin)

	// Run go install with custom GOBIN
	cmd := exec.Command("go", "install", packageURL)
	cmd.Env = append(os.Environ(), "GOBIN="+tempGoBin)
	
	if debug {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install %s: %w", packageURL, err)
	}

	// Move the installed binary to the correct name and location
	installedBinary := filepath.Join(tempGoBin, "goctl") // The actual binary name
	if runtime.GOOS == "windows" {
		installedBinary += ".exe"
	}

	finalPath := filepath.Join(installPath, name)
	if runtime.GOOS == "windows" {
		finalPath += ".exe"
	}

	if err := os.Rename(installedBinary, finalPath); err != nil {
		return fmt.Errorf("failed to move binary to final location: %w", err)
	}

	log.Info("Binary installed successfully via go install", "name", name, "path", finalPath)
	return nil
}