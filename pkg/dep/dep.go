package dep

import (
	"fmt"
	"log"

	"path/filepath"
	"runtime"

	"github.com/joeblew999/infra/pkg/store"
)

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
		Version:    "v3.37.0", // Example version, update as needed
		ReleaseURL: "https://github.com/go-task/task/releases/tag/v3.37.0",
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
	log.Println("Ensuring core binaries...")

	for _, binary := range embeddedCoreBinaries {
		log.Printf("Checking %s (version %s) from %s", binary.Name, binary.Version, binary.Repo)

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
	}

	log.Println("Core binaries ensured.")
	return nil
}

// Get returns the absolute path to the requested binary for the current platform.
func Get(name string) string {
	return filepath.Join(store.GetDepPath(), fmt.Sprintf(store.BinaryDepNameFormat, name, runtime.GOOS, runtime.GOARCH))
}
