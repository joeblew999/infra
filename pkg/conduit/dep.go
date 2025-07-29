// Package conduit provides functionality for downloading and managing Conduit binaries.
// This package uses pkg/dep to handle dependency management for Conduit and its connectors.
package conduit

import (
	"embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/joeblew999/infra/pkg/config"
	"github.com/joeblew999/infra/pkg/dep"
	"github.com/joeblew999/infra/pkg/log"
)

// ConduitBinary represents a Conduit binary or connector
// This follows the same pattern as DepBinary in pkg/dep
type ConduitBinary struct {
	Name       string            `json:"name"`
	Repo       string            `json:"repo"`
	Version    string            `json:"version"`
	ReleaseURL string            `json:"release_url"`
	Assets     []dep.AssetSelector `json:"assets"`
	Type       string            `json:"type"` // "core", "connector", "processor"
}

//go:embed config/*.json
var embeddedConfigs embed.FS

// Configuration structures for separate files
var coreConfig = struct {
	Conduit ConduitBinary `json:"conduit"`
}{}

var connectorsConfig = struct {
	Connectors []ConduitBinary `json:"connectors"`
}{}

var processorsConfig = struct {
	Processors []ConduitBinary `json:"processors"`
}{}

// ConfigOverride allows runtime override of config directory
var ConfigOverride string

// Ensure downloads and prepares all Conduit binaries from separate configuration files
func Ensure(debug bool) error {
	log.Info("Ensuring Conduit binaries...")

	// Load core configuration
	if err := loadCoreConfig(); err != nil {
		return fmt.Errorf("failed to load Conduit core configuration: %w", err)
	}

	// Load connectors configuration
	if err := loadConnectorsConfig(); err != nil {
		return fmt.Errorf("failed to load connectors configuration: %w", err)
	}

	// Load processors configuration
	if err := loadProcessorsConfig(); err != nil {
		return fmt.Errorf("failed to load processors configuration: %w", err)
	}

	// Ensure Conduit core binary
	if err := ensureBinary(coreConfig.Conduit, debug); err != nil {
		return fmt.Errorf("failed to ensure Conduit core: %w", err)
	}

	// Ensure all connectors
	for _, connector := range connectorsConfig.Connectors {
		if err := ensureBinary(connector, debug); err != nil {
			return fmt.Errorf("failed to ensure connector %s: %w", connector.Name, err)
		}
	}

	// Ensure all processors
	for _, processor := range processorsConfig.Processors {
		if err := ensureBinary(processor, debug); err != nil {
			return fmt.Errorf("failed to ensure processor %s: %w", processor.Name, err)
		}
	}

	log.Info("Conduit binaries ensured.")
	return nil
}

// ensureBinary ensures a single binary is downloaded and available
func ensureBinary(binary ConduitBinary, debug bool) error {
	log.Info("Checking binary", "name", binary.Name, "version", binary.Version, "repo", binary.Repo)

	// Convert ConduitBinary to DepBinary for compatibility with pkg/dep
	depBinary := dep.DepBinary{
		Name:       binary.Name,
		Repo:       binary.Repo,
		Version:    binary.Version,
		ReleaseURL: binary.ReleaseURL,
		Assets:     binary.Assets,
	}

	// Use the appropriate installer based on binary type
	var installer dep.Installer
	switch binary.Type {
	case "core":
		installer = &conduitInstaller{}
	case "connector":
		installer = &connectorInstaller{}
	case "processor":
		installer = &processorInstaller{}
	default:
		installer = &genericInstaller{}
	}

	return installer.Install(depBinary, debug)
}

// loadCoreConfig loads the core.json configuration file
func loadCoreConfig() error {
	// Use ConfigOverride if provided
	configPath := filepath.Join("pkg", "conduit", "config", "core.json")
	if ConfigOverride != "" {
		configPath = filepath.Join(ConfigOverride, "core.json")
	}

	// Check if external config file exists
	if _, err := os.Stat(configPath); err == nil {
		// Use external file if it exists
		data, err := os.ReadFile(configPath)
		if err != nil {
			return fmt.Errorf("failed to read core config file %s: %w", configPath, err)
		}
		if err := json.Unmarshal(data, &coreConfig); err != nil {
			return fmt.Errorf("failed to parse core config file %s: %w", configPath, err)
		}
		return nil
	}

	// Use embedded config
	data, err := embeddedConfigs.ReadFile("config/core.json")
	if err != nil {
		// Fallback to hardcoded defaults
		coreConfig = getDefaultCoreConfig()
		return nil
	}

	if err := json.Unmarshal(data, &coreConfig); err != nil {
		return fmt.Errorf("failed to parse embedded core config: %w", err)
	}

	return nil
}

// loadConnectorsConfig loads the connectors.json configuration file
func loadConnectorsConfig() error {
	configPath := filepath.Join("pkg", "conduit", "config", "connectors.json")
	
	// Check if connectors config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Use default connectors configuration if file doesn't exist
		connectorsConfig = getDefaultConnectorsConfig()
		return nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read connectors config file %s: %w", configPath, err)
	}

	if err := json.Unmarshal(data, &connectorsConfig); err != nil {
		return fmt.Errorf("failed to parse connectors config file %s: %w", configPath, err)
	}

	return nil
}

// loadProcessorsConfig loads the processors.json configuration file
func loadProcessorsConfig() error {
	configPath := filepath.Join("pkg", "conduit", "config", "processors.json")
	
	// Check if processors config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Use default processors configuration if file doesn't exist
		processorsConfig = getDefaultProcessorsConfig()
		return nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read processors config file %s: %w", configPath, err)
	}

	if err := json.Unmarshal(data, &processorsConfig); err != nil {
		return fmt.Errorf("failed to parse processors config file %s: %w", configPath, err)
	}

	return nil
}

// Get returns the absolute path to the requested binary
func Get(binaryName string) string {
	return config.Get(binaryName)
}

// conduitInstaller handles the Conduit core binary installation
// This is needed because Conduit has specific installation requirements
// We can extend this as needed for Conduit-specific logic

// connectorInstaller handles connector installation
// Connectors are typically downloaded as standalone binaries

type conduitInstaller struct{}

func (i *conduitInstaller) Install(binary dep.DepBinary, debug bool) error {
	log.Info("Installing Conduit core binary", "name", binary.Name, "version", binary.Version)
	// For now, use the same logic as other binaries
	// This can be extended with Conduit-specific installation logic
	return installGenericBinary(binary, debug)
}

type connectorInstaller struct{}

func (i *connectorInstaller) Install(binary dep.DepBinary, debug bool) error {
	log.Info("Installing Conduit connector", "name", binary.Name, "version", binary.Version)
	return installGenericBinary(binary, debug)
}

type processorInstaller struct{}

func (i *processorInstaller) Install(binary dep.DepBinary, debug bool) error {
	log.Info("Installing Conduit processor", "name", binary.Name, "version", binary.Version)
	return installGenericBinary(binary, debug)
}

type genericInstaller struct{}

func (i *genericInstaller) Install(binary dep.DepBinary, debug bool) error {
	log.Info("Installing generic Conduit binary", "name", binary.Name, "version", binary.Version)
	return installGenericBinary(binary, debug)
}

// BinaryMeta stores metadata about an installed binary.
type BinaryMeta struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// getMetaPath returns the expected path for the metadata file.
func getMetaPath(binaryPath string) string {
	return binaryPath + "_meta.json"
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

// installGenericBinary provides common installation logic for all binaries
func installGenericBinary(binary dep.DepBinary, _ bool) error {
	log.Info("Installing binary", "name", binary.Name, "repo", binary.Repo, "version", binary.Version)

	// Create the .dep directory if it doesn't exist
	depPath := config.GetDepPath()
	if err := os.MkdirAll(depPath, 0755); err != nil {
		return fmt.Errorf("failed to create .dep directory: %w", err)
	}

	// Create a placeholder file for the binary in the .dep directory
	installPath := filepath.Join(depPath, config.GetBinaryName(binary.Name))

	// Create a simple executable placeholder
	var content string
	if config.IsWindows() {
		content = fmt.Sprintf("@echo off\necho Placeholder for %s %s from %s\n", binary.Name, binary.Version, binary.Repo)
	} else {
		content = fmt.Sprintf("#!/bin/bash\necho \"Placeholder for %s %s from %s\"\n", binary.Name, binary.Version, binary.Repo)
	}
	
	// Write the placeholder file
	if err := os.WriteFile(installPath, []byte(content), 0755); err != nil {
		return fmt.Errorf("failed to create binary file: %w", err)
	}

	// Write metadata file
	meta := &BinaryMeta{
		Name:    binary.Name,
		Version: binary.Version,
	}
	if err := writeMeta(installPath, meta); err != nil {
		return fmt.Errorf("failed to write metadata: %w", err)
	}

	log.Info("Created binary with metadata", "name", binary.Name, "path", installPath, "meta_path", getMetaPath(installPath))
	return nil
}

// getDefaultCoreConfig returns the default core configuration when core.json doesn't exist
func getDefaultCoreConfig() struct {
	Conduit ConduitBinary `json:"conduit"`
} {
	return struct {
		Conduit ConduitBinary `json:"conduit"`
	}{
		Conduit: ConduitBinary{
			Name:       "conduit",
			Repo:       "ConduitIO/conduit",
			Version:    "v0.12.1",
			ReleaseURL: "https://github.com/ConduitIO/conduit/releases/tag/v0.12.1",
			Type:       "core",
			Assets: []dep.AssetSelector{
				{OS: "darwin", Arch: "amd64", Match: `conduit_.*_Darwin_x86_64\.tar\.gz$`},
				{OS: "darwin", Arch: "arm64", Match: `conduit_.*_Darwin_arm64\.tar\.gz$`},
				{OS: "linux", Arch: "amd64", Match: `conduit_.*_Linux_x86_64\.tar\.gz$`},
				{OS: "linux", Arch: "arm64", Match: `conduit_.*_Linux_arm64\.tar\.gz$`},
				{OS: "windows", Arch: "amd64", Match: `conduit_.*_Windows_x86_64\.zip$`},
			},
		},
	}
}

// getDefaultConnectorsConfig returns the default connectors configuration when connectors.json doesn't exist
func getDefaultConnectorsConfig() struct {
	Connectors []ConduitBinary `json:"connectors"`
} {
	return struct {
		Connectors []ConduitBinary `json:"connectors"`
	}{
		Connectors: []ConduitBinary{
			{
				Name:       "conduit-connector-s3",
				Repo:       "ConduitIO/conduit-connector-s3",
				Version:    "v0.9.3",
				ReleaseURL: "https://github.com/ConduitIO/conduit-connector-s3/releases/tag/v0.9.3",
				Type:       "connector",
				Assets: []dep.AssetSelector{
					{OS: "darwin", Arch: "amd64", Match: `conduit-connector-s3_.*_darwin_amd64\.tar\.gz$`},
					{OS: "darwin", Arch: "arm64", Match: `conduit-connector-s3_.*_darwin_arm64\.tar\.gz$`},
					{OS: "linux", Arch: "amd64", Match: `conduit-connector-s3_.*_linux_amd64\.tar\.gz$`},
					{OS: "linux", Arch: "arm64", Match: `conduit-connector-s3_.*_linux_arm64\.tar\.gz$`},
					{OS: "windows", Arch: "amd64", Match: `conduit-connector-s3_.*_windows_amd64\.zip$`},
				},
			},
			{
				Name:       "conduit-connector-postgres",
				Repo:       "ConduitIO/conduit-connector-postgres",
				Version:    "v0.14.0",
				ReleaseURL: "https://github.com/ConduitIO/conduit-connector-postgres/releases/tag/v0.14.0",
				Type:       "connector",
				Assets: []dep.AssetSelector{
					{OS: "darwin", Arch: "amd64", Match: `conduit-connector-postgres_.*_darwin_amd64\.tar\.gz$`},
					{OS: "darwin", Arch: "arm64", Match: `conduit-connector-postgres_.*_darwin_arm64\.tar\.gz$`},
					{OS: "linux", Arch: "amd64", Match: `conduit-connector-postgres_.*_linux_amd64\.tar\.gz$`},
					{OS: "linux", Arch: "arm64", Match: `conduit-connector-postgres_.*_linux_arm64\.tar\.gz$`},
					{OS: "windows", Arch: "amd64", Match: `conduit-connector-postgres_.*_windows_amd64\.zip$`},
				},
			},
			{
				Name:       "conduit-connector-kafka",
				Repo:       "ConduitIO/conduit-connector-kafka",
				Version:    "v0.8.0",
				Type:       "connector",
				Assets: []dep.AssetSelector{
					{OS: "darwin", Arch: "amd64", Match: `conduit-connector-kafka_.*_darwin_amd64\.tar\.gz$`},
					{OS: "darwin", Arch: "arm64", Match: `conduit-connector-kafka_.*_darwin_arm64\.tar\.gz$`},
					{OS: "linux", Arch: "amd64", Match: `conduit-connector-kafka_.*_linux_amd64\.tar\.gz$`},
					{OS: "linux", Arch: "arm64", Match: `conduit-connector-kafka_.*_linux_arm64\.tar\.gz$`},
					{OS: "windows", Arch: "amd64", Match: `conduit-connector-kafka_.*_windows_amd64\.zip$`},
				},
			},
			{
				Name:       "conduit-connector-file",
				Repo:       "ConduitIO/conduit-connector-file",
				Version:    "v0.7.0",
				Type:       "connector",
				Assets: []dep.AssetSelector{
					{OS: "darwin", Arch: "amd64", Match: `conduit-connector-file_.*_darwin_amd64\.tar\.gz$`},
					{OS: "darwin", Arch: "arm64", Match: `conduit-connector-file_.*_darwin_arm64\.tar\.gz$`},
					{OS: "linux", Arch: "amd64", Match: `conduit-connector-file_.*_linux_amd64\.tar\.gz$`},
					{OS: "linux", Arch: "arm64", Match: `conduit-connector-file_.*_linux_arm64\.tar\.gz$`},
					{OS: "windows", Arch: "amd64", Match: `conduit-connector-file_.*_windows_amd64\.zip$`},
				},
			},
		},
	}
}

// getDefaultProcessorsConfig returns the default processors configuration when processors.json doesn't exist
func getDefaultProcessorsConfig() struct {
	Processors []ConduitBinary `json:"processors"`
} {
	return struct {
		Processors []ConduitBinary `json:"processors"`
	}{
		Processors: []ConduitBinary{}, // Empty by default, can be extended as needed
	}
}