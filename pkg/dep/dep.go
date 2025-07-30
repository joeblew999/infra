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
//   - BinaryMeta, DepBinary, AssetSelector structs - Data structures
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
// Currently supported binaries: bento, task, tofu, caddy, ko, flyctl, garble, claude
//
// Each binary is automatically selected based on runtime.GOOS and runtime.GOARCH
// using regex patterns to match GitHub release assets.
package dep

import (
	"embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
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

// DepBinary represents a dependency binary.
type DepBinary struct {
	Name       string          `json:"name"`
	Repo       string          `json:"repo"`
	Version    string          `json:"version"`
	ReleaseURL string          `json:"release_url"` // Full URL to the GitHub release page
	Assets     []AssetSelector `json:"assets"`
}

// AssetSelector defines how to select a release asset.
type AssetSelector struct {
	OS    string `json:"os"`
	Arch  string `json:"arch"`
	Match string `json:"match"` // Regular expression to match the asset filename
}

// Installer defines the interface for installing a dependency binary.
type Installer interface {
	Install(binary DepBinary, debug bool) error
}

//go:embed dep.json
var embeddedConfig embed.FS

// depBinaries contains the loaded dependency binaries from JSON configuration
var depBinaries []DepBinary

// loadConfig loads the dependency configuration from embedded JSON or external file
func loadConfig() ([]DepBinary, error) {
	// If depBinaries is already loaded, return it
	if len(depBinaries) > 0 {
		return depBinaries, nil
	}

	// Try to load from external config file first
	configPath := filepath.Join("pkg", "dep", "dep.json")
	if _, err := os.Stat(configPath); err == nil {
		data, err := os.ReadFile(configPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read external config file %s: %w", configPath, err)
		}
		var binaries []DepBinary
		if err := json.Unmarshal(data, &binaries); err != nil {
			return nil, fmt.Errorf("failed to parse external config file %s: %w", configPath, err)
		}
		depBinaries = binaries
		return depBinaries, nil
	}

	// Fall back to embedded configuration
	data, err := embeddedConfig.ReadFile("dep.json")
	if err != nil {
		return nil, fmt.Errorf("failed to read embedded config: %w", err)
	}

	var binaries []DepBinary
	if err := json.Unmarshal(data, &binaries); err != nil {
		return nil, fmt.Errorf("failed to parse embedded config: %w", err)
	}

	depBinaries = binaries
	return depBinaries, nil
}

// Ensure downloads and prepares all binaries defined in the manifest.
// This function will handle both core bootstrapping binaries and generic ones.
func Ensure(debug bool) error {
	log.Info("Ensuring core binaries...")

	// Load configuration from embedded JSON
	binaries, err := loadConfig()
	if err != nil {
		return fmt.Errorf("failed to load dependency configuration: %w", err)
	}

	for _, binary := range binaries {
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
		case "bun":
			installer = &bunInstaller{}
		case "claude":
			installer = &claudeInstaller{}
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

	// Load configuration to validate binary exists
	binaries, err := loadConfig()
	if err != nil {
		return "", fmt.Errorf("failed to load dependency configuration: %w", err)
	}

	// Validate that the binary is in our supported list
	for _, binary := range binaries {
		if binary.Name == name {
			return config.Get(name), nil
		}
	}

	return "", fmt.Errorf("%w: binary '%s' is not supported", ErrBinaryNotFound, name)
}

// Remove deletes a specific binary and its metadata file from the .dep directory.
// Useful for testing/debugging to force reinstallation of a specific binary.
func Remove(name string) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("%w: binary name cannot be empty", ErrInvalidInput)
	}

	// Validate that the binary is supported
	binaries, err := loadConfig()
	if err != nil {
		return fmt.Errorf("failed to load dependency configuration: %w", err)
	}

	found := false
	for _, binary := range binaries {
		if binary.Name == name {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("%w: binary '%s' is not supported", ErrBinaryNotFound, name)
	}

	// Get the binary path
	binaryPath := config.Get(name)
	metaPath := getMetaPath(binaryPath)

	// Remove binary if it exists
	if _, err := os.Stat(binaryPath); err == nil {
		if err := os.Remove(binaryPath); err != nil {
			return fmt.Errorf("failed to remove binary: %w", err)
		}
	}

	// Remove metadata file if it exists
	if _, err := os.Stat(metaPath); err == nil {
		if err := os.Remove(metaPath); err != nil {
			return fmt.Errorf("failed to remove metadata: %w", err)
		}
	}

	// Remove the claude-code directory for claude (npm package)
	if name == "claude" {
		claudeDir := filepath.Join(filepath.Dir(binaryPath), "claude-code")
		if _, err := os.Stat(claudeDir); err == nil {
			if err := os.RemoveAll(claudeDir); err != nil {
				return fmt.Errorf("failed to remove claude package directory: %w", err)
			}
		}
	}

	log.Info("Binary removed successfully", "name", name)
	return nil
}
