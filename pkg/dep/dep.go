package dep

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/joeblew999/infra/pkg/log"
	"github.com/joeblew999/infra/pkg/config"
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
}

// Ensure downloads and prepares all binaries defined in the manifest.
// This function will handle both core bootstrapping binaries and generic ones.
func Ensure(debug bool) error {
	log.Info("Ensuring core binaries...")

	for _, binary := range embeddedCoreBinaries {
		log.Info("Checking binary", "name", binary.Name, "version", binary.Version, "repo", binary.Repo)

		installPath := Get(binary.Name)
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
		case "task":
			installer = &taskInstaller{}
		case "tofu":
			installer = &tofuInstaller{}
		case "caddy":
			installer = &caddyInstaller{}
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
func Get(name string) string {
	return config.Get(name)
}
