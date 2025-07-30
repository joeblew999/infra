package dep

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/joeblew999/infra/pkg/log"
)

// bunInstaller handles installation of Bun from GitHub releases
// Downloads platform-specific zip files from oven-sh/bun releases
type bunInstaller struct{}

// Install downloads and installs the Bun binary
func (i *bunInstaller) Install(binary DepBinary, debug bool) error {
	// Determine platform and architecture
	platform := runtime.GOOS
	arch := runtime.GOARCH

	// Map Go platform/architecture to Bun's naming convention
	bunPlatform, bunArch, ext := i.getBunPlatformInfo(platform, arch)
	if bunPlatform == "" || bunArch == "" {
		return fmt.Errorf("unsupported platform/architecture: %s/%s", platform, arch)
	}

	// Use specific version from binary.Version
	actualVersion := binary.Version
	if actualVersion == "" || actualVersion == "latest" {
		actualVersion = "v1.2.19" // Default to known working version
	}

	// Construct download URL using GitHub release pattern
	downloadURL := fmt.Sprintf("https://github.com/oven-sh/bun/releases/download/%s/bun-%s-%s.%s",
		actualVersion, bunPlatform, bunArch, ext)

	log.Info("Downloading Bun", "url", downloadURL, "platform", platform, "arch", arch)

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
		return fmt.Errorf("failed to download and install Bun: %w", err)
	}

	// Make executable on Unix systems
	if platform != "windows" {
		if err := os.Chmod(installPath, 0755); err != nil {
			return fmt.Errorf("failed to make Bun executable: %w", err)
		}
	}

	log.Info("Bun installed successfully", "path", installPath)
	return nil
}

// getBunPlatformInfo maps Go platform/architecture to Bun's naming convention
func (i *bunInstaller) getBunPlatformInfo(goOS, goArch string) (platform, arch, ext string) {
	// Map Go OS to Bun platform names
	platformMap := map[string]string{
		"darwin":  "darwin",
		"linux":   "linux",
		"windows": "windows",
	}

	// Map Go arch to Bun arch names
	archMap := map[string]string{
		"amd64": "x64",
		"arm64": "aarch64",
	}

	platform, ok1 := platformMap[goOS]
	arch, ok2 := archMap[goArch]
	ext = "zip"

	if !ok1 || !ok2 {
		return "", "", ""
	}

	return platform, arch, ext
}

// downloadAndExtract handles the actual download and extraction
func (i *bunInstaller) downloadAndExtract(url, installPath, _ string, _ bool) error {
	// Create a temporary directory for extraction
	tempDir, err := os.MkdirTemp("", "bun-download")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Download to temp directory
	archivePath := filepath.Join(tempDir, "bun-archive")
	if err := downloadFileDirect(url, archivePath); err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}

	// Extract the zip file
	if err := unzip(archivePath, tempDir); err != nil {
		return fmt.Errorf("failed to extract zip: %w", err)
	}

	// Find the binary in extracted files
	binaryName := "bun"
	if runtime.GOOS == "windows" {
		binaryName = "bun.exe"
	}

	// Check for platform-specific subdirectory - determine platform/arch names
	platform := runtime.GOOS
	arch := runtime.GOARCH
	
	// Map to Bun naming convention
	bunPlatform, bunArch, _ := (&bunInstaller{}).getBunPlatformInfo(platform, arch)
	platformArch := fmt.Sprintf("bun-%s-%s", bunPlatform, bunArch)
	
	// Check possible locations for the binary
	possiblePaths := []string{
		filepath.Join(tempDir, platformArch, binaryName),      // Platform subdirectory
		filepath.Join(tempDir, binaryName),                    // Root
		filepath.Join(tempDir, "bun", binaryName),            // Generic subdirectory
	}
	
	found := false
	binaryPath := ""
	
	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			binaryPath = path
			found = true
			break
		}
	}
	
	if !found {
		return fmt.Errorf("could not find %s in extracted archive", binaryName)
	}

	// Move binary to final location
	if err := os.Rename(binaryPath, installPath); err != nil {
		return fmt.Errorf("failed to move binary to install location: %w", err)
	}

	return nil
}