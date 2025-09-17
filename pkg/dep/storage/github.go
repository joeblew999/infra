package storage

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/joeblew999/infra/pkg/config"
	"github.com/joeblew999/infra/pkg/log"
)

// GitHub provides unified storage for GitHub Packages and Releases
// Handles both uploads to GitHub Packages and downloads from GitHub Releases

type GitHub struct {
	// HTTP client for API calls
	client *http.Client
}

// NewGitHub creates a new GitHub storage client
func NewGitHub() *GitHub {
	return &GitHub{
		client: &http.Client{},
	}
}

// UploadToPackages uploads a binary to GitHub Packages using GitHub CLI
// Creates a release if it doesn't exist and uploads the binary as an asset
func (g *GitHub) UploadToPackages(owner, repo, binaryName, version, sourcePath string) error {
	platform := fmt.Sprintf("%s-%s", runtime.GOOS, runtime.GOARCH)
	assetName := fmt.Sprintf("%s-%s-%s", binaryName, version, platform)
	if runtime.GOOS == "windows" {
		assetName += ".exe"
	}
	
	// Ensure GitHub CLI is available
	if err := g.ensureGitHubCLI(); err != nil {
		return fmt.Errorf("GitHub CLI not available: %w", err)
	}
	
	// Ensure the release exists
	releaseTag := fmt.Sprintf("%s-%s", binaryName, version)
	createCmd := exec.Command("gh", "release", "create", releaseTag, "--title", fmt.Sprintf("%s %s", binaryName, version), "--notes", fmt.Sprintf("Automated release of %s %s", binaryName, version), "--repo", fmt.Sprintf("%s/%s", owner, repo))
	createCmd.Stdout = os.Stdout
	createCmd.Stderr = os.Stderr
	
	// Create release (ignore if exists)
	_ = createCmd.Run()
	
	// Upload the asset
	uploadCmd := exec.Command("gh", "release", "upload", releaseTag, sourcePath, "--clobber", "--repo", fmt.Sprintf("%s/%s", owner, repo))
	uploadCmd.Stdout = os.Stdout
	uploadCmd.Stderr = os.Stderr
	
	if err := uploadCmd.Run(); err != nil {
		return fmt.Errorf("failed to upload %s: %w", assetName, err)
	}
	
	fmt.Printf("[GitHub Packages] Successfully uploaded %s to release %s\n", assetName, releaseTag)
	return nil
}

// DownloadFromReleases downloads a binary from GitHub Releases using direct HTTP
// No authentication required - public releases only
func (g *GitHub) DownloadFromReleases(owner, repo, releaseURL, destPath string, assets []AssetInfo) error {
	// TODO: Implement direct HTTP download from GitHub releases
	// This should download public release assets without requiring GitHub CLI
	return fmt.Errorf("direct GitHub releases download not implemented yet for %s/%s", owner, repo)
}

// DownloadFromPackages downloads a binary from GitHub Releases using GitHub CLI
// First tries the specific release tag, then falls back to standard releases
func (g *GitHub) DownloadFromPackages(owner, repo, binaryName, version, destPath string) error {
	platform := fmt.Sprintf("%s-%s", runtime.GOOS, runtime.GOARCH)
	assetName := fmt.Sprintf("%s-%s-%s", binaryName, version, platform)
	if runtime.GOOS == "windows" {
		assetName += ".exe"
	}

	// Ensure GitHub CLI is available
	if err := g.ensureGitHubCLI(); err != nil {
		return fmt.Errorf("GitHub CLI not available: %w", err)
	}

	releaseTag := fmt.Sprintf("%s-%s", binaryName, version)

	// Create destination directory
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Use GitHub CLI to download the release asset
	downloadCmd := exec.Command("gh", "release", "download", releaseTag, 
		"--repo", fmt.Sprintf("%s/%s", owner, repo),
		"--pattern", assetName,
		"--output", destPath,
		"--clobber")
	
	if err := downloadCmd.Run(); err != nil {
		return fmt.Errorf("failed to download %s via CLI: %w", assetName, err)
	}

	fmt.Printf("[GitHub Packages] Successfully downloaded %s from release %s\n", assetName, releaseTag)
	return nil
}

// AssetInfo represents a GitHub release asset for download
// This is used by the download system to match the correct asset

type AssetInfo struct {
	OS    string
	Arch  string
	Match string // regex pattern to match filename
}

// GetBinaryName returns the platform-specific binary name
func (g *GitHub) GetBinaryName(baseName string) string {
	return config.GetBinaryName(baseName)
}


// IsGitHubCLIReady checks if GitHub CLI is available and authenticated
func (g *GitHub) IsGitHubCLIReady() bool {
	// Check if gh command exists
	if _, err := exec.LookPath("gh"); err != nil {
		return false
	}
	
	// Check if authenticated
	cmd := exec.Command("gh", "auth", "status")
	if err := cmd.Run(); err != nil {
		return false
	}
	
	return true
}

// GetPlatform returns the current platform string
func (g *GitHub) GetPlatform() string {
	return fmt.Sprintf("%s-%s", runtime.GOOS, runtime.GOARCH)
}

// BuildPackageName builds the GitHub Packages naming convention
func (g *GitHub) BuildPackageName(owner, repo, binaryName, version string) string {
	platform := fmt.Sprintf("%s-%s", runtime.GOOS, runtime.GOARCH)
	return fmt.Sprintf("ghcr.io/%s/%s/%s:%s-%s", owner, repo, binaryName, version, platform)
}

// ensureGitHubCLI ensures GitHub CLI is available, installing it if necessary
func (g *GitHub) ensureGitHubCLI() error {
	if g.IsGitHubCLIReady() {
		return nil // Already available
	}

	// Check if we can find the gh binary at the configured path first
	ghPath := config.Get("gh")
	if _, err := os.Stat(ghPath); err == nil {
		log.Info("GitHub CLI found at configured path", "path", ghPath)
		return nil // Binary exists at expected location
	}

	// During dep installation or preflight, we can't auto-install to avoid circular dependency
	// Just return a helpful error message
	return fmt.Errorf("GitHub CLI not available at %s - run 'go run . dep install gh' first", ghPath)
}