package dep

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	"github.com/joeblew999/infra/pkg/log"
)

// claudeInstaller handles installation of Claude Code CLI from Anthropic's npm package
// This downloads the npm package and creates a wrapper script to execute the Node.js CLI
type claudeInstaller struct{}

// Install downloads and installs the Claude Code binary
func (i *claudeInstaller) Install(binary DepBinary, debug bool) error {
	// Use "latest" version or specific version from binary.Version
	version := binary.Version
	if version == "" || version == "latest" {
		version = "latest"
	}

	// Determine platform and architecture
	platform := runtime.GOOS
	arch := runtime.GOARCH

	// Map Go platform/arch to Claude's naming convention
	claudePlatform, claudeArch, ext := i.getClaudePlatformInfo(platform, arch)
	if claudePlatform == "" || claudeArch == "" {
		return fmt.Errorf("unsupported platform/architecture: %s/%s", platform, arch)
	}

	// Use specific version instead of "latest"
	actualVersion := "1.0.62" // Latest version from npm
	if version != "latest" {
		actualVersion = version
	}
	
	// Construct download URL using npm registry for Claude Code package
	downloadURL := fmt.Sprintf("https://registry.npmjs.org/@anthropic-ai/claude-code/-/claude-code-%s.tgz", 
		actualVersion)

	log.Info("Downloading Claude Code", "url", downloadURL, "platform", platform, "arch", arch)

	// Get install path
	installPath, err := Get(binary.Name)
	if err != nil {
		return fmt.Errorf("failed to get install path for %s: %w", binary.Name, err)
	}

	// Ensure directory exists
	installDir := filepath.Dir(installPath)
	if err := os.MkdirAll(installDir, 0755); err != nil {
		return fmt.Errorf("failed to create install directory: %w", err)
	}

	// Download and extract the binary
	if err := i.downloadAndExtract(downloadURL, installPath, ext, debug); err != nil {
		return fmt.Errorf("failed to download and install Claude: %w", err)
	}

	// Make binary executable on Unix systems
	if platform != "windows" {
		if err := os.Chmod(installPath, 0755); err != nil {
			return fmt.Errorf("failed to make Claude executable: %w", err)
		}
	}

	log.Info("Claude Code installed successfully", "path", installPath)
	return nil
}

// getClaudePlatformInfo maps Go platform/architecture to Claude's naming convention
func (i *claudeInstaller) getClaudePlatformInfo(goOS, goArch string) (platform, arch, ext string) {
	// Map Go OS to Claude platform names
	platformMap := map[string]string{
		"darwin":  "darwin",
		"linux":   "linux", 
		"windows": "windows",
	}

	// Map Go arch to Claude arch names
	archMap := map[string]string{
		"amd64": "amd64",
		"arm64": "arm64",
		"386":   "386",
	}

	platform, ok1 := platformMap[goOS]
	arch, ok2 := archMap[goArch]
	ext = "tar.gz"
	
	if goOS == "windows" {
		ext = "zip"
	}

	if !ok1 || !ok2 {
		return "", "", ""
	}

	return platform, arch, ext
}

// downloadAndExtract handles the actual download and extraction of the binary
func (i *claudeInstaller) downloadAndExtract(url, installPath, ext string, debug bool) error {
	// Create a temporary directory for extraction
	tempDir, err := os.MkdirTemp("", "claude-download")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Download to temp directory
	archivePath := filepath.Join(tempDir, "claude-archive")
	if err := downloadFileDirect(url, archivePath); err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}

	// Extract the npm package (tgz format)
	if err := untarGz(archivePath, tempDir); err != nil {
		return fmt.Errorf("failed to extract tgz: %w", err)
	}

	// The npm package contains a Node.js CLI tool, not a native binary
	// We'll install the entire package and create a wrapper script
	packageDir := filepath.Join(tempDir, "package")
	cliPath := filepath.Join(packageDir, "cli.js")
	
	if _, err := os.Stat(cliPath); os.IsNotExist(err) {
		return fmt.Errorf("could not find cli.js in npm package")
	}

	// Create install directory
	installDir := filepath.Dir(installPath)
	if err := os.MkdirAll(installDir, 0755); err != nil {
		return fmt.Errorf("failed to create install directory: %w", err)
	}

	// Create a dedicated directory for the Claude package
	claudeDir := filepath.Join(installDir, "claude-code")
	if err := os.RemoveAll(claudeDir); err != nil {
		return fmt.Errorf("failed to remove existing claude directory: %w", err)
	}
	
	// Move the entire package to the install location
	if err := os.Rename(packageDir, claudeDir); err != nil {
		return fmt.Errorf("failed to move package to install location: %w", err)
	}

	// Create a wrapper script to execute the CLI with Bun (fallback to Node.js)
	wrapperScript := `#!/bin/bash
if command -v bun >/dev/null 2>&1; then
    exec bun "$(dirname "$0")/claude-code/cli.js" "$@"
else
    exec node "$(dirname "$0")/claude-code/cli.js" "$@"
fi
`

	wrapperPath := installPath
	if runtime.GOOS == "windows" {
		wrapperScript = `@echo off
where bun >nul 2>nul
if %errorlevel% == 0 (
    bun "%~dp0claude-code\cli.js" %*
) else (
    node "%~dp0claude-code\cli.js" %*
)
`
		wrapperPath = installPath + ".bat"
	}

	if err := os.WriteFile(wrapperPath, []byte(wrapperScript), 0755); err != nil {
		return fmt.Errorf("failed to create wrapper script: %w", err)
	}

	return nil
}

// Helper functions for downloading and extracting using existing dep infrastructure
func downloadFileDirect(url, destPath string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(destPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Download file
	out, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", destPath, err)
	}
	defer out.Close()

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download from %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status %d for %s", resp.StatusCode, url)
	}

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to copy downloaded content to file: %w", err)
	}

	return nil
}