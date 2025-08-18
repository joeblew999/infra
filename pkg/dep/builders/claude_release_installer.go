package builders

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/joeblew999/infra/pkg/config"
	"github.com/joeblew999/infra/pkg/dep/util"
	"github.com/joeblew999/infra/pkg/log"
)

// ClaudeReleaseInstaller handles Claude Code installation from Google Cloud Storage
type ClaudeReleaseInstaller struct{}

// ClaudeManifest represents the manifest.json structure
type ClaudeManifest struct {
	Platforms map[string]ClaudePlatform `json:"platforms"`
}

// ClaudePlatform represents platform-specific information
type ClaudePlatform struct {
	Checksum string `json:"checksum"`
}

// Install downloads and installs Claude Code from Google Cloud Storage
func (i *ClaudeReleaseInstaller) Install(name, version string, debug bool) error {
	log.Info("Installing Claude Code from Google Cloud Storage", "name", name, "version", version)

	// Get the install path
	installPath := config.Get(name)
	
	// Ensure .dep directory exists
	installDir := filepath.Dir(installPath)
	if err := os.MkdirAll(installDir, 0755); err != nil {
		return fmt.Errorf("failed to create install directory: %w", err)
	}

	// Base URL for Claude releases
	baseURL := "https://storage.googleapis.com/claude-code-dist-86c565f3-f756-42ad-8dfa-d59b1c096819/claude-code-releases"

	// Get version (use stable if not specified or "latest")
	targetVersion := version
	if version == "latest" || version == "" {
		var err error
		targetVersion, err = i.getStableVersion(baseURL)
		if err != nil {
			return fmt.Errorf("failed to get stable version: %w", err)
		}
		log.Info("Using stable version", "version", targetVersion)
	}

	// Determine platform
	platform, err := i.getPlatform()
	if err != nil {
		return fmt.Errorf("failed to determine platform: %w", err)
	}
	log.Info("Detected platform", "platform", platform)

	// Download and verify manifest
	manifest, err := i.getManifest(baseURL, targetVersion)
	if err != nil {
		return fmt.Errorf("failed to get manifest: %w", err)
	}

	// Check if platform is supported
	platformInfo, exists := manifest.Platforms[platform]
	if !exists {
		// Try alternative platform names for Windows
		if strings.HasPrefix(platform, "windows") {
			alternativePlatforms := i.getAlternativeWindowsPlatforms(platform)
			for _, altPlatform := range alternativePlatforms {
				if info, exists := manifest.Platforms[altPlatform]; exists {
					platform = altPlatform
					platformInfo = info
					log.Info("Using alternative platform name", "platform", platform)
					break
				}
			}
		}
		
		if !exists {
			return fmt.Errorf("platform %s not supported, available platforms: %v", platform, i.getAvailablePlatforms(manifest))
		}
	}

	// Download binary
	binaryURL := fmt.Sprintf("%s/%s/%s/claude", baseURL, targetVersion, platform)
	if runtime.GOOS == "windows" {
		binaryURL += ".exe"
	}

	log.Info("Downloading Claude binary", "url", binaryURL)
	tempFile, err := util.DownloadToTemp(binaryURL, "claude-download-", true)
	if err != nil {
		return fmt.Errorf("failed to download binary: %w", err)
	}
	defer os.Remove(tempFile)

	// Verify checksum
	if err := i.verifyChecksum(tempFile, platformInfo.Checksum); err != nil {
		return fmt.Errorf("checksum verification failed: %w", err)
	}

	// Copy to final location
	if err := i.copyBinary(tempFile, installPath); err != nil {
		return fmt.Errorf("failed to copy binary: %w", err)
	}

	log.Info("Claude Code installed successfully", "path", installPath, "version", targetVersion)
	return nil
}

// getStableVersion fetches the stable version string
func (i *ClaudeReleaseInstaller) getStableVersion(baseURL string) (string, error) {
	resp, err := http.Get(baseURL + "/stable")
	if err != nil {
		return "", fmt.Errorf("failed to fetch stable version: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch stable version, status: %d", resp.StatusCode)
	}

	versionBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read stable version: %w", err)
	}

	return strings.TrimSpace(string(versionBytes)), nil
}

// getPlatform determines the platform string for Claude
func (i *ClaudeReleaseInstaller) getPlatform() (string, error) {
	var os, arch string

	switch runtime.GOOS {
	case "darwin":
		os = "darwin"
	case "linux":
		os = "linux"
	case "windows":
		os = "win32" // Claude uses win32 instead of windows
	default:
		return "", fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}

	switch runtime.GOARCH {
	case "amd64":
		arch = "x64"
	case "arm64":
		arch = "arm64"
	default:
		return "", fmt.Errorf("unsupported architecture: %s", runtime.GOARCH)
	}

	// Handle Linux musl detection (simplified)
	if runtime.GOOS == "linux" {
		// Check for musl indicators
		if i.isMuslLinux() {
			return fmt.Sprintf("%s-%s-musl", os, arch), nil
		}
	}

	return fmt.Sprintf("%s-%s", os, arch), nil
}

// isMuslLinux attempts to detect if we're running on musl Linux
func (i *ClaudeReleaseInstaller) isMuslLinux() bool {
	// Check for common musl indicators
	muslPaths := []string{
		"/lib/libc.musl-x86_64.so.1",
		"/lib/libc.musl-aarch64.so.1",
	}
	
	for _, path := range muslPaths {
		if _, err := os.Stat(path); err == nil {
			return true
		}
	}
	
	return false
}

// getAlternativeWindowsPlatforms returns alternative platform names to try for Windows
func (i *ClaudeReleaseInstaller) getAlternativeWindowsPlatforms(original string) []string {
	alternatives := []string{}
	
	if strings.Contains(original, "x64") {
		alternatives = append(alternatives, "win32-x64", "win-x64", "windows-amd64")
	}
	if strings.Contains(original, "arm64") {
		alternatives = append(alternatives, "win32-arm64", "win-arm64", "windows-arm64")
	}
	
	return alternatives
}

// getManifest downloads and parses the manifest.json
func (i *ClaudeReleaseInstaller) getManifest(baseURL, version string) (*ClaudeManifest, error) {
	manifestURL := fmt.Sprintf("%s/%s/manifest.json", baseURL, version)
	resp, err := http.Get(manifestURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch manifest: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch manifest, status: %d", resp.StatusCode)
	}

	var manifest ClaudeManifest
	if err := json.NewDecoder(resp.Body).Decode(&manifest); err != nil {
		return nil, fmt.Errorf("failed to parse manifest: %w", err)
	}

	return &manifest, nil
}

// getAvailablePlatforms returns a list of available platforms from manifest
func (i *ClaudeReleaseInstaller) getAvailablePlatforms(manifest *ClaudeManifest) []string {
	platforms := make([]string, 0, len(manifest.Platforms))
	for platform := range manifest.Platforms {
		platforms = append(platforms, platform)
	}
	return platforms
}



// verifyChecksum verifies the SHA256 checksum of the downloaded file
func (i *ClaudeReleaseInstaller) verifyChecksum(filePath, expectedChecksum string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file for checksum: %w", err)
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return fmt.Errorf("failed to calculate checksum: %w", err)
	}

	actualChecksum := hex.EncodeToString(hasher.Sum(nil))
	if actualChecksum != expectedChecksum {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expectedChecksum, actualChecksum)
	}

	return nil
}

// copyBinary copies the verified binary to the final location
func (i *ClaudeReleaseInstaller) copyBinary(srcPath, destPath string) error {
	src, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("failed to open source: %w", err)
	}
	defer src.Close()

	dest, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create destination: %w", err)
	}
	defer dest.Close()

	if _, err := io.Copy(dest, src); err != nil {
		return fmt.Errorf("failed to copy binary: %w", err)
	}

	// Make executable on Unix systems
	if runtime.GOOS != "windows" {
		if err := os.Chmod(destPath, 0755); err != nil {
			return fmt.Errorf("failed to make binary executable: %w", err)
		}
	}

	return nil
}