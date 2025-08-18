package collection

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// DepBinary represents a binary dependency (copied from dep package to avoid circular import)
type DepBinary struct {
	Name    string         `json:"name"`
	Repo    string         `json:"repo,omitempty"`
	Version string         `json:"version"`
	Assets  []AssetInfo    `json:"assets,omitempty"`
}

// AssetInfo represents a release asset for matching (copied from dep package)
type AssetInfo struct {
	OS    string `json:"os"`
	Arch  string `json:"arch"`
	Match string `json:"match"`
}

// loadDepConfig loads dependency configuration from dep.json
func loadDepConfig() ([]DepBinary, error) {
	// Check for local dep.json first
	if _, err := os.Stat("dep.json"); err == nil {
		return loadDepConfigFromFile("dep.json")
	}

	// Check in pkg/dep/dep.json
	depPath := filepath.Join("pkg", "dep", "dep.json")
	if _, err := os.Stat(depPath); err == nil {
		return loadDepConfigFromFile(depPath)
	}

	// Could also check embedded config here if needed
	return nil, fmt.Errorf("no dep.json found")
}

// loadDepConfigFromFile loads configuration from a specific file
func loadDepConfigFromFile(path string) ([]DepBinary, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open %s: %w", path, err)
	}
	defer file.Close()

	var binaries []DepBinary
	if err := json.NewDecoder(file).Decode(&binaries); err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", path, err)
	}

	return binaries, nil
}