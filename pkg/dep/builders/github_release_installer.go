package builders

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"

	"github.com/joeblew999/infra/pkg/config"
	"github.com/joeblew999/infra/pkg/dep/internal"
	"github.com/joeblew999/infra/pkg/dep/util"
	"github.com/joeblew999/infra/pkg/log"
)


// GitHubReleaseInstaller handles GitHub release-based binary installations
type GitHubReleaseInstaller struct{}

// Note: We'll use the types from the main dep package to avoid duplication

// AssetSelector defines how to select a release asset
type AssetSelector struct {
	OS    string `json:"os"`
	Arch  string `json:"arch"`
	Match string `json:"match"` // Regular expression to match the asset filename
}

// Install downloads and installs a binary from GitHub releases
func (i *GitHubReleaseInstaller) Install(name, repo, version string, assets []AssetSelector, debug bool) error {
	log.Info("Installing from GitHub release", "binary", name, "repo", repo, "version", version)

	// Get the install path
	installPath := config.Get(name)
	
	// Ensure .dep directory exists
	installDir := filepath.Dir(installPath)
	if err := os.MkdirAll(installDir, 0755); err != nil {
		return fmt.Errorf("failed to create install directory: %w", err)
	}

	// Get GitHub release information
	release, err := i.getGitHubRelease(repo, version)
	if err != nil {
		return fmt.Errorf("failed to get release info: %w", err)
	}

	// Select the appropriate asset for current platform
	asset, err := i.selectAsset(release, assets)
	if err != nil {
		return fmt.Errorf("failed to select asset for %s: %w", name, err)
	}

	log.Info("Downloading asset", "asset_name", asset.Name, "url", asset.BrowserDownloadURL)

	// Create temporary directory for download and extraction
	tempDir, err := os.MkdirTemp("", fmt.Sprintf("%s-download", name))
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Download the asset
	archivePath := filepath.Join(tempDir, asset.Name)
	if err := util.DownloadFile(asset.BrowserDownloadURL, archivePath, true); err != nil {
		return fmt.Errorf("failed to download asset: %w", err)
	}

	// Extract the archive
	extractDir := filepath.Join(tempDir, "extracted")
	if err := os.MkdirAll(extractDir, 0755); err != nil {
		return fmt.Errorf("failed to create extraction directory: %w", err)
	}

	if err := internal.ExtractArchive(archivePath, extractDir); err != nil {
		return fmt.Errorf("failed to extract archive: %w", err)
	}

	// Find and copy the binary to the install location
	if err := i.installBinary(extractDir, installPath, name); err != nil {
		return fmt.Errorf("failed to install binary: %w", err)
	}

	log.Info("Successfully installed binary", "binary", name, "path", installPath)
	return nil
}

// getGitHubRelease fetches release information from GitHub API
func (i *GitHubReleaseInstaller) getGitHubRelease(repo, version string) (*GitHubRelease, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/tags/%s", repo, version)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch GitHub release from %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d for %s", resp.StatusCode, url)
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("failed to decode GitHub release response: %w", err)
	}

	return &release, nil
}

// selectAsset selects the appropriate asset for the current platform
func (i *GitHubReleaseInstaller) selectAsset(release *GitHubRelease, selectors []AssetSelector) (*GitHubReleaseAsset, error) {
	for _, selector := range selectors {
		if selector.OS == runtime.GOOS && selector.Arch == runtime.GOARCH {
			for _, asset := range release.Assets {
				if matched, _ := regexp.MatchString(selector.Match, asset.Name); matched {
					return &asset, nil
				}
			}
		}
	}
	return nil, fmt.Errorf("no matching asset found for %s/%s", runtime.GOOS, runtime.GOARCH)
}



// installBinary finds the binary in the extracted directory and copies it to install path
func (i *GitHubReleaseInstaller) installBinary(extractDir, installPath, binaryName string) error {
	// Look for the binary in common locations
	possiblePaths := []string{
		filepath.Join(extractDir, binaryName),
		filepath.Join(extractDir, "bin", binaryName),
		filepath.Join(extractDir, binaryName, "bin", binaryName), // TinyGo pattern: tinygo/bin/tinygo
		filepath.Join(extractDir, binaryName+".exe"), // Windows
		filepath.Join(extractDir, "bin", binaryName+".exe"), // Windows
		filepath.Join(extractDir, binaryName, "bin", binaryName+".exe"), // TinyGo pattern Windows
	}

	// Also look for binaries in subdirectories (common pattern: binary-name-version-os-arch/binary-name)
	if err := filepath.Walk(extractDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Continue walking even if we hit an error
		}
		
		if info.IsDir() {
			return nil
		}
		
		filename := info.Name()
		if filename == binaryName || filename == binaryName+".exe" {
			// Verify it's executable (or on Windows)
			if info.Mode().Perm()&0111 != 0 || runtime.GOOS == "windows" {
				possiblePaths = append(possiblePaths, path)
			}
		}
		
		return nil
	}); err != nil {
		log.Warn("Error walking extraction directory", "error", err)
	}

	// Find the first existing binary
	var foundPath string
	for _, path := range possiblePaths {
		if stat, err := os.Stat(path); err == nil {
			if stat.IsDir() {
				log.Debug("Skipping directory", "path", path)
				continue // Skip directories
			}
			foundPath = path
			break
		}
	}

	if foundPath == "" {
		return fmt.Errorf("binary %s not found in extracted archive", binaryName)
	}

	// Copy the binary to the install location
	sourceFile, err := os.Open(foundPath)
	if err != nil {
		return fmt.Errorf("failed to open source binary: %w", err)
	}
	defer sourceFile.Close()

	destFile, err := os.Create(installPath)
	if err != nil {
		return fmt.Errorf("failed to create destination binary: %w", err)
	}
	defer destFile.Close()

	// Copy the file
	if _, err := io.Copy(destFile, sourceFile); err != nil {
		return fmt.Errorf("failed to copy binary: %w", err)
	}

	// Make it executable on Unix systems
	if runtime.GOOS != "windows" {
		if err := os.Chmod(installPath, 0755); err != nil {
			return fmt.Errorf("failed to make binary executable: %w", err)
		}
	}

	return nil
}


// GitHub API types (duplicated from util.go to avoid circular imports)
type GitHubReleaseAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

type GitHubRelease struct {
	Assets []GitHubReleaseAsset `json:"assets"`
}