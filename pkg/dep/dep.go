// Package dep provides a design-by-contract binary dependency management system.
//
// This package automatically downloads and manages GitHub-released binaries with
// version tracking and platform-specific asset selection. It follows design by
// contract principles to ensure API stability for consuming packages.
//
// # Public API Guarantees
//
// The following functions and types form the stable public API:
//   - Ensure(debug bool) error - Downloads and ensures all binaries are available
//   - Get(name string) (string, error) - Returns the path to a binary
//   - BinaryMeta, CoreBinary, AssetSelector structs - Data structures
//   - ErrBinaryNotFound, ErrInvalidInput, ErrInstallationFailed - Error types
//
// # API Stability Contract
//
//   - Function signatures will not change without major version bump
//   - Error types will remain consistent for error handling
//   - Struct field names and JSON tags are stable
//   - Binary name validation rules are stable
//
// # Usage Example
//
//	// Ensure all binaries are downloaded and up-to-date
//	if err := dep.Ensure(false); err != nil {
//	    log.Fatal(err)
//	}
//
//	// Get path to a specific binary
//	path, err := dep.Get("garble")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// # Supported Binaries
//
// Currently supported binaries: bento, task, tofu, caddy, ko, flyctl, garble
//
// Each binary is automatically selected based on runtime.GOOS and runtime.GOARCH
// using regex patterns to match GitHub release assets.
package dep

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/joeblew999/infra/pkg/log"
	"github.com/joeblew999/infra/pkg/config"
)

// Package errors for better error handling and API stability
var (
	// ErrBinaryNotFound is returned when a requested binary is not available
	ErrBinaryNotFound = fmt.Errorf("binary not found")
	// ErrInvalidInput is returned when input validation fails
	ErrInvalidInput = fmt.Errorf("invalid input")
	// ErrInstallationFailed is returned when binary installation fails
	ErrInstallationFailed = fmt.Errorf("installation failed")
)

// BinaryMeta stores metadata about an installed binary.
type BinaryMeta struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// getMetaPath returns the expected path for the metadata file.
func getMetaPath(binaryPath string) string {
	return binaryPath + "_meta.json"
}

// readMeta reads the metadata file for a given binary path.
func readMeta(binaryPath string) (*BinaryMeta, error) {
	metaPath := getMetaPath(binaryPath)
	data, err := os.ReadFile(metaPath)
	if err != nil {
		return nil, err
	}

	var meta BinaryMeta
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata for %s: %w", binaryPath, err)
	}
	return &meta, nil
}

// writeMeta writes the metadata file for a given binary path.
func writeMeta(binaryPath string, meta *BinaryMeta) error {
	metaPath := getMetaPath(binaryPath)
	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata for %s: %w", binaryPath, err)
	}

	if err := os.WriteFile(metaPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write metadata for %s: %w", binaryPath, err)
	}
	return nil
}

// CoreBinary represents a core bootstrapping binary.
type CoreBinary struct {
	Name       string
	Repo       string
	Version    string
	ReleaseURL string // Full URL to the GitHub release page
	Assets     []AssetSelector
}

// AssetSelector defines how to select a release asset.
type AssetSelector struct {
	OS    string
	Arch  string
	Match string // Regular expression to match the asset filename
}

// Installer defines the interface for installing a core binary.
type Installer interface {
	Install(binary CoreBinary, debug bool) error
}

// embeddedCoreBinaries will contain the manifest for core bootstrapping binaries.
// This will be embedded at compile time.
var embeddedCoreBinaries = []CoreBinary{
	{
		Name:       "bento",
		Repo:       "warpstreamlabs/bento",
		Version:    "v1.9.0",
		ReleaseURL: "https://github.com/warpstreamlabs/bento/releases/tag/v1.9.0",
		Assets: []AssetSelector{
			{OS: "darwin", Arch: "amd64", Match: `bento_.*_darwin_amd64\.tar\.gz$`},
			{OS: "darwin", Arch: "arm64", Match: `bento_.*_darwin_arm64\.tar\.gz$`},
			{OS: "linux", Arch: "amd64", Match: `bento_.*_linux_amd64\.tar\.gz$`},
			{OS: "linux", Arch: "arm64", Match: `bento_.*_linux_arm64\.tar\.gz$`},
			{OS: "windows", Arch: "amd64", Match: `bento_.*_windows_amd64\.zip$`},
		},
	},
	{
		Name:       "task",
		Repo:       "go-task/task",
		Version:    "v3.44.1", // Example version, update as needed
		ReleaseURL: "https://github.com/go-task/task/releases/tag/v3.44.1",
		Assets: []AssetSelector{
			{OS: "darwin", Arch: "amd64", Match: `task_darwin_amd64\.tar\.gz$`},
			{OS: "darwin", Arch: "arm64", Match: `task_darwin_arm64\.tar\.gz$`},
			{OS: "linux", Arch: "amd64", Match: `task_linux_amd64\.tar\.gz$`},
			{OS: "linux", Arch: "arm64", Match: `task_linux_arm64\.tar\.gz$`},
			{OS: "windows", Arch: "amd64", Match: `task_windows_amd64\.zip$`},
		},
	},
	{
		Name:       "tofu",
		Repo:       "opentofu/opentofu",
		Version:    "v1.7.2", // Example version, update as needed
		ReleaseURL: "https://github.com/opentofu/opentofu/releases/tag/v1.7.2",
		Assets: []AssetSelector{
			{OS: "darwin", Arch: "amd64", Match: `tofu_.*_darwin_amd64\.zip$`},
			{OS: "darwin", Arch: "arm64", Match: `tofu_.*_darwin_arm64\.zip$`},
			{OS: "linux", Arch: "amd64", Match: `tofu_.*_linux_amd64\.zip$`},
			{OS: "linux", Arch: "arm64", Match: `tofu_.*_linux_arm64\.zip$`},
			{OS: "windows", Arch: "amd64", Match: `tofu_.*_windows_amd64\.zip$`},
		},
	},
	{
		Name:       "caddy",
		Repo:       "caddyserver/caddy",
		Version:    "v2.10.0",                                                   // Updated version
		ReleaseURL: "https://github.com/caddyserver/caddy/releases/tag/v2.10.0", // Updated URL
		Assets: []AssetSelector{
			{OS: "darwin", Arch: "amd64", Match: `caddy_.*_darwin_amd64\.tar\.gz$`},
			{OS: "darwin", Arch: "arm64", Match: `caddy_.*_mac_arm64\.tar\.gz$`},
			{OS: "linux", Arch: "amd64", Match: `caddy_.*_linux_amd64\.tar\.gz$`},
			{OS: "linux", Arch: "arm64", Match: `caddy_.*_linux_arm64\.tar\.gz$`},
			{OS: "windows", Arch: "amd64", Match: `caddy_.*_windows_amd64\.zip$`},
		},
	},
	{
		Name:       "ko",
		Repo:       "ko-build/ko",
		Version:    "v0.18.0",
		ReleaseURL: "https://github.com/ko-build/ko/releases/tag/v0.18.0",
		Assets: []AssetSelector{
			{OS: "darwin", Arch: "amd64", Match: `ko_.*_Darwin_x86_64\.tar\.gz$`},
			{OS: "darwin", Arch: "arm64", Match: `ko_.*_Darwin_arm64\.tar\.gz$`},
			{OS: "linux", Arch: "amd64", Match: `ko_.*_Linux_x86_64\.tar\.gz$`},
			{OS: "linux", Arch: "arm64", Match: `ko_.*_Linux_arm64\.tar\.gz$`},
			{OS: "windows", Arch: "amd64", Match: `ko_.*_Windows_x86_64\.tar\.gz$`},
		},
	},
	{
		Name:       "flyctl",
		Repo:       "superfly/flyctl",
		Version:    "v0.3.159",
		ReleaseURL: "https://github.com/superfly/flyctl/releases/tag/v0.3.159",
		Assets: []AssetSelector{
			{OS: "darwin", Arch: "amd64", Match: `flyctl_.*_macOS_x86_64\.tar\.gz$`},
			{OS: "darwin", Arch: "arm64", Match: `flyctl_.*_macOS_arm64\.tar\.gz$`},
			{OS: "linux", Arch: "amd64", Match: `flyctl_.*_Linux_x86_64\.tar\.gz$`},
			{OS: "linux", Arch: "arm64", Match: `flyctl_.*_Linux_arm64\.tar\.gz$`},
			{OS: "windows", Arch: "amd64", Match: `flyctl_.*_Windows_x86_64\.zip$`},
		},
	},
	{
		Name:       "garble",
		Repo:       "burrowers/garble",
		Version:    "v0.14.2",
		ReleaseURL: "https://github.com/burrowers/garble/releases/tag/v0.14.2",
		Assets:     []AssetSelector{}, // Garble uses go install, no assets needed
	},
}

// Ensure downloads and prepares all binaries defined in the manifest.
// This function will handle both core bootstrapping binaries and generic ones.
func Ensure(debug bool) error {
	log.Info("Ensuring core binaries...")

	for _, binary := range embeddedCoreBinaries {
		log.Info("Checking binary", "name", binary.Name, "version", binary.Version, "repo", binary.Repo)

		installPath, err := Get(binary.Name)
		if err != nil {
			return fmt.Errorf("failed to get install path for %s: %w", binary.Name, err)
		}
		currentMeta, err := readMeta(installPath)
		if err == nil && currentMeta.Version == binary.Version {
			log.Info("Binary up to date", "name", binary.Name, "version", binary.Version)
			continue // Skip installation
		} else if err != nil && !os.IsNotExist(err) {
			log.Warn("Error reading metadata", "name", binary.Name, "error", err, "action", "attempting re-download")
		} else if currentMeta != nil && currentMeta.Version != binary.Version {
			log.Warn("Version mismatch", "name", binary.Name, "expected", binary.Version, "got", currentMeta.Version, "action", "attempting re-download")
		} else {
			log.Info("Binary not found or metadata missing", "name", binary.Name, "action", "attempting download and installation")
		}

		// Determine the correct installer based on binary name
		var installer Installer
		switch binary.Name {
		case "bento":
			installer = &bentoInstaller{}
		case "task":
			installer = &taskInstaller{}
		case "tofu":
			installer = &tofuInstaller{}
		case "caddy":
			installer = &caddyInstaller{}
		case "ko":
			installer = &koInstaller{}
		case "flyctl":
			installer = &flyctlInstaller{}
		case "garble":
			installer = &garbleInstaller{}
		default:
			return fmt.Errorf("no installer found for binary: %s", binary.Name)
		}

		if err := installer.Install(binary, debug); err != nil {
			return fmt.Errorf("failed to install %s: %w", binary.Name, err)
		}

		// Write metadata after successful installation
		if err := writeMeta(installPath, &BinaryMeta{Name: binary.Name, Version: binary.Version}); err != nil {
			return fmt.Errorf("failed to write metadata for %s: %w", binary.Name, err)
		}
	}

	log.Info("Core binaries ensured.")
	return nil
}

// Get returns the absolute path to the requested binary for the current platform.
// Returns an error if the binary name is invalid or not supported.
func Get(name string) (string, error) {
	if strings.TrimSpace(name) == "" {
		return "", fmt.Errorf("%w: binary name cannot be empty", ErrInvalidInput)
	}

	// Validate that the binary is in our supported list
	for _, binary := range embeddedCoreBinaries {
		if binary.Name == name {
			return config.Get(name), nil
		}
	}

	return "", fmt.Errorf("%w: binary '%s' is not supported", ErrBinaryNotFound, name)
}
