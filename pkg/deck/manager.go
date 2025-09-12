package deck

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

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

// ConvertToPDF converts an XML file to PDF using decksvg then external tools
func (m *Manager) ConvertToPDF(xmlFile, outputPDF string) error {
	// Get the decksvg binary path first  
	decksvgPath, err := m.GetBinary("decksvg")
	if err != nil {
		return fmt.Errorf("decksvg binary not available: %w", err)
	}

	// Check if input file exists
	if _, err := os.Stat(xmlFile); os.IsNotExist(err) {
		return fmt.Errorf("input XML file not found: %s", xmlFile)
	}

	// First convert to SVG - decksvg creates files with pattern: filename-00001.svg
	outputDir := filepath.Dir(xmlFile)
	xmlBaseName := strings.TrimSuffix(filepath.Base(xmlFile), filepath.Ext(xmlFile))
	expectedSVG := filepath.Join(outputDir, xmlBaseName+"-00001.svg")
	
	cmd := exec.Command(decksvgPath, "-outdir", outputDir, xmlFile)
	
	// Set DECKFONTS environment for font access
	env := os.Environ()
	env = append(env, "DECKFONTS=.data/font")
	cmd.Env = env

	log.Info("Converting XML to SVG", "input", xmlFile, "output", expectedSVG)
	
	// Capture both stdout and stderr for debugging
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to convert XML to SVG: %w, stderr: %s", err, stderr.String())
	}
	
	// Log stderr if there are warnings
	if stderr.Len() > 0 {
		log.Info("decksvg warnings/errors", "stderr", stderr.String())
	}
	
	// Check if the expected SVG file was created
	if _, err := os.Stat(expectedSVG); os.IsNotExist(err) {
		return fmt.Errorf("expected SVG file not created: %s", expectedSVG)
	}
	
	svgFile := expectedSVG

	// Keep the SVG file next to the original markdown
	finalSVG := strings.TrimSuffix(outputPDF, filepath.Ext(outputPDF)) + ".svg"
	
	log.Info("Attempting to rename SVG file", "from", svgFile, "to", finalSVG)
	
	// Check if the source SVG file exists
	if _, err := os.Stat(svgFile); err != nil {
		return fmt.Errorf("source SVG file does not exist: %s - %w", svgFile, err)
	}
	
	if err := os.Rename(svgFile, finalSVG); err != nil {
		log.Info("Rename failed, attempting copy instead", "error", err.Error())
		
		// If rename fails, try copy instead
		sourceData, readErr := os.ReadFile(svgFile)
		if readErr != nil {
			return fmt.Errorf("failed to read source SVG file: %w", readErr)
		}
		
		if writeErr := os.WriteFile(finalSVG, sourceData, 0644); writeErr != nil {
			return fmt.Errorf("failed to write final SVG file: %w", writeErr)
		}
		
		// Clean up temp file after successful copy
		os.Remove(svgFile)
	}
	
	log.Info("Successfully created SVG output", "path", finalSVG)
	return nil
}