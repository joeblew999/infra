package dep

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/joeblew999/infra/pkg/store"
)

// CoreBinary represents a core bootstrapping binary.
type CoreBinary struct {
	Name    string
	Repo    string
	Version string
	Assets  []AssetSelector
}

// AssetSelector defines how to select a release asset.
type AssetSelector struct {
	OS    string
	Arch  string
	Match string // Regular expression to match the asset filename
}

// embeddedCoreBinaries will contain the manifest for core bootstrapping binaries.
// This will be embedded at compile time.
var embeddedCoreBinaries = []CoreBinary{
	{
		Name:    "task",
		Repo:    "go-task/task",
		Version: "v3.37.0", // Example version, update as needed
		Assets: []AssetSelector{
			{OS: "darwin", Arch: "amd64", Match: "task_darwin_amd64\\.tar\\.gz$"},
			{OS: "darwin", Arch: "arm64", Match: "task_darwin_arm64\\.tar\\.gz$"},
			{OS: "linux", Arch: "amd64", Match: "task_linux_amd64\\.tar\\.gz$"},
			{OS: "linux", Arch: "arm64", Match: "task_linux_arm64\\.tar\\.gz$"},
			{OS: "windows", Arch: "amd64", Match: "task_windows_amd64\\.zip$"},
		},
	},
	{
		Name:    "tofu",
		Repo:    "opentofu/opentofu",
		Version: "v1.7.2", // Example version, update as needed
		Assets: []AssetSelector{
			{OS: "darwin", Arch: "amd64", Match: "tofu_.*_darwin_amd64\\.zip$"},
			{OS: "darwin", Arch: "arm64", Match: "tofu_.*_darwin_arm64\\.zip$"},
			{OS: "linux", Arch: "amd64", Match: "tofu_.*_linux_amd64\\.zip$"},
			{OS: "linux", Arch: "arm64", Match: "tofu_.*_linux_arm64\\.zip$"},
			{OS: "windows", Arch: "amd64", Match: "tofu_.*_windows_amd64\\.zip$"},
		},
	},
	{
		Name:    "caddy",
		Repo:    "caddyserver/caddy",
		Version: "v2.8.4", // Example version, update as needed
		Assets: []AssetSelector{
			{OS: "darwin", Arch: "amd64", Match: "caddy_.*_darwin_amd64\\.tar\\.gz$"},
			{OS: "darwin", Arch: "arm64", Match: "caddy_.*_darwin_arm64\\.tar\\.gz$"},
			{OS: "linux", Arch: "amd64", Match: "caddy_.*_linux_amd64\\.tar\\.gz$"},
			{OS: "linux", Arch: "arm64", Match: "caddy_.*_linux_arm64\\.tar\\.gz$"},
			{OS: "windows", Arch: "amd64", Match: "caddy_.*_windows_amd64\\.zip$"},
		},
	},
}

// Ensure downloads and prepares all binaries defined in the manifest.
// This function will handle both core bootstrapping binaries and generic ones.
func Ensure() error {
	log.Println("Ensuring core binaries...")

	for _, binary := range embeddedCoreBinaries {
		log.Printf("Checking %s (version %s) from %s", binary.Name, binary.Version, binary.Repo)
		// TODO: Implement actual download, extraction, and installation logic here
		// For now, just simulate success if the target path exists

		// Simulate checking if binary exists
		installPath := filepath.Join(store.GetDepPath(), fmt.Sprintf("%s_%s_%s", binary.Name, os.GOOS, os.GOARCH))
		if _, err := os.Stat(installPath); os.IsNotExist(err) {
			log.Printf("  %s not found. Simulating download and installation...", binary.Name)
			// In a real implementation, this would involve HTTP requests, unarchiving, etc.
			// For now, we'll just create a dummy file.
			dummyFile, err := os.Create(installPath)
			if err != nil {
				return fmt.Errorf("failed to create dummy binary for %s: %w", binary.Name, err)
			}
			dummyFile.Close()
			log.Printf("  Simulated installation of %s to %s", binary.Name, installPath)
		} else if err != nil {
			return fmt.Errorf("error checking existence of %s: %w", binary.Name, err)
		} else {
			log.Printf("  %s already exists at %s. Skipping download.", binary.Name, installPath)
		}
	}

	log.Println("Core binaries ensured.")
	return nil
}

// Get returns the absolute path to the requested binary for the current platform.
func Get(name string) (string, error) {
	// TODO: Implement logic to find the binary in the .dep folder
	// For now, just return a simulated path
	for _, binary := range embeddedCoreBinaries {
		if binary.Name == name {
			return filepath.Join(store.GetDepPath(), fmt.Sprintf("%s_%s_%s", binary.Name, os.GOOS, os.GOARCH)), nil
		}
	}
	return "", fmt.Errorf("binary %s not found", name)
}
