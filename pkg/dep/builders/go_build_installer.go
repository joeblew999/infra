package builders

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/joeblew999/infra/pkg/config"
	"github.com/joeblew999/infra/pkg/dep/storage"
	"github.com/joeblew999/infra/pkg/log"
)

// GoBuildInstaller uses pkg/deck's build pattern for Go source compilation
type GoBuildInstaller struct{}

func (i *GoBuildInstaller) Install(name, repo, pkg, version string, debug bool) error {
	log.Info("Installing via 2-phase pipeline", "binary", name, "repo", repo, "package", pkg)

	// Parse owner/repo from repo string
	parts := strings.Split(repo, "/")
	if len(parts) != 2 {
		return fmt.Errorf("invalid repo format: %s (expected owner/repo)", repo)
	}
	owner, repoName := parts[0], parts[1]

	installPath := filepath.Join(config.GetDepPath(), name)
	if runtime.GOOS == "windows" {
		installPath += ".exe"
	}

	// Phase 1: Try to download from GitHub Packages
	githubStorage := storage.NewGitHub()
	if err := githubStorage.DownloadFromPackages(owner, repoName, name, version, installPath); err == nil {
		log.Info("Downloaded from GitHub Packages", "binary", name, "path", installPath)
		return nil
	}

	log.Info("Binary not in GitHub Packages, building from source", "binary", name)

	// Phase 2: Build from source and upload to GitHub Packages
	// Ensure absolute path for GOBIN
	absInstallPath, err := filepath.Abs(installPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}
	installDir := filepath.Dir(absInstallPath)

	// Build directory setup
	buildDir := filepath.Join(config.GetDepPath(), "build", name)
	if err := os.MkdirAll(buildDir, 0755); err != nil {
		return fmt.Errorf("failed to create build directory: %w", err)
	}
	defer os.RemoveAll(buildDir) // Clean up after build

	// Build the binary using go install
	log.Info("Building from source", "package", pkg, "version", version)

	var installCmd string
	if version == "latest" {
		installCmd = fmt.Sprintf("%s@latest", pkg)
	} else {
		installCmd = fmt.Sprintf("%s@%s", pkg, version)
	}

	// Use go install to build and install
	cmd := exec.Command("go", "install", installCmd)
	cmd.Env = append(os.Environ(), "GOBIN="+installDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to build %s: %w", name, err)
	}

	// Ensure executable permissions
	if err := os.Chmod(absInstallPath, 0755); err != nil {
		return fmt.Errorf("failed to set executable permissions: %w", err)
	}

	// Phase 3: Upload to GitHub Packages for future use
	uploadStorage := storage.NewGitHub()
	if err := uploadStorage.UploadToPackages(owner, repoName, name, version, absInstallPath); err != nil {
		log.Warn("Failed to upload to GitHub Packages, but binary is available locally", "error", err)
	} else {
		log.Info("Uploaded to GitHub Packages", "binary", name, "version", version)
	}

	log.Info("Successfully installed", "binary", name, "path", absInstallPath)
	return nil
}
