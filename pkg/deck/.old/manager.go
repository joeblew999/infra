package deck

import (
	"fmt"
	"os"

	"github.com/joeblew999/infra/pkg/log"
)

// Manager coordinates deck tool building and usage
type Manager struct {
	Builder *Builder
}

// NewManager creates a new deck manager
func NewManager() *Manager {
	return &Manager{
		Builder: NewBuilder(),
	}
}

// Install builds all deck tools from source
func (m *Manager) Install() error {
	log.Info("Installing deck tools from source...")
	return m.Builder.BuildAll()
}

// InstallTool builds a specific tool
func (m *Manager) InstallTool(name string) error {
	log.Info("Installing specific deck tool", "tool", name)
	return m.Builder.BuildTool(name)
}

// GetBinary returns path to native binary
func (m *Manager) GetBinary(name string) (string, error) {
	binaries, _ := m.Builder.GetPaths()
	if path, ok := binaries[name]; ok {
		return path, nil
	}
	return "", fmt.Errorf("binary %s not found, run install first", name)
}

// GetWASM returns path to WASM module
func (m *Manager) GetWASM(name string) (string, error) {
	_, wasm := m.Builder.GetPaths()
	if path, ok := wasm[name]; ok {
		return path, nil
	}
	return "", fmt.Errorf("WASM %s not found, run install first", name)
}

// ListTools returns available tools
func (m *Manager) ListTools() []string {
	var tools []string
	for _, tool := range Tools {
		tools = append(tools, tool.Name)
	}
	return tools
}

// Status shows build status of all tools
func (m *Manager) Status() map[string]map[string]string {
	binaries, wasm := m.Builder.GetPaths()
	
	status := make(map[string]map[string]string)
	for _, tool := range Tools {
		toolStatus := make(map[string]string)
		
		if path, ok := binaries[tool.Name]; ok {
			toolStatus["binary"] = path
		} else {
			toolStatus["binary"] = "not built"
		}
		
		if path, ok := wasm[tool.Name]; ok {
			toolStatus["wasm"] = path
		} else {
			toolStatus["wasm"] = "not built"
		}
		
		status[tool.Name] = toolStatus
	}
	
	return status
}

// Clean removes all built binaries and WASM
func (m *Manager) Clean() error {
	log.Info("Cleaning deck build artifacts...")
	
	buildDir := m.Builder.BuildDir
	wasmDir := m.Builder.WASMDir
	
	if err := os.RemoveAll(buildDir); err != nil {
		return fmt.Errorf("failed to clean binaries: %w", err)
	}
	
	if err := os.RemoveAll(wasmDir); err != nil {
		return fmt.Errorf("failed to clean WASM: %w", err)
	}
	
	// Recreate directories
	return m.Builder.EnsureDirectories()
}

// Update pulls latest source code and rebuilds
func (m *Manager) Update() error {
	log.Info("Updating deck tools...")
	
	// Clean and rebuild
	if err := m.Clean(); err != nil {
		return fmt.Errorf("failed to clean: %w", err)
	}
	
	// Remove source directories to force fresh clones
	sourceDir := m.Builder.SourceDir
	if err := os.RemoveAll(sourceDir); err != nil {
		return fmt.Errorf("failed to clean source: %w", err)
	}
	
	return m.Install()
}