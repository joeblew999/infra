package deck

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/joeblew999/infra/pkg/config"
	"github.com/joeblew999/infra/pkg/log"
)

// HealthIssue represents a health check problem
type HealthIssue struct {
	Type        string `json:"type"`        // tool, pipeline, dependency, fonts
	Tool        string `json:"tool,omitempty"`
	Message     string `json:"message"`
	Severity    string `json:"severity"`    // error, warning, info
	Suggestion  string `json:"suggestion,omitempty"`
	Timestamp   string `json:"timestamp"`
}

// HealthReport contains the results of all health checks
type HealthReport struct {
	Overall     string         `json:"overall"`     // healthy, degraded, unhealthy
	Timestamp   string         `json:"timestamp"`
	Duration    string         `json:"duration"`
	ToolsOK     int           `json:"tools_ok"`
	ToolsTotal  int           `json:"tools_total"`
	PipelineOK  bool          `json:"pipeline_ok"`
	Issues      []HealthIssue `json:"issues"`
}

// HealthChecker performs health checks on the deck system
type HealthChecker struct {
	manager   *Manager
	builder   *Builder
	tempDir   string
	verbose   bool
}

// NewHealthChecker creates a new health checker
func NewHealthChecker() *HealthChecker {
	return &HealthChecker{
		manager: NewManager(),
		builder: NewBuilder(),
		verbose: false,
	}
}

// SetVerbose enables verbose logging for health checks
func (h *HealthChecker) SetVerbose(verbose bool) {
	h.verbose = verbose
}

// RunFullHealthCheck performs all health checks and returns a comprehensive report
func (h *HealthChecker) RunFullHealthCheck() *HealthReport {
	start := time.Now()
	report := &HealthReport{
		Timestamp:  start.Format(time.RFC3339),
		ToolsTotal: len(Tools),
		Issues:     []HealthIssue{},
	}

	// Create temp directory for testing
	var err error
	h.tempDir, err = os.MkdirTemp("", TempDirPrefix+"*")
	if err != nil {
		report.Issues = append(report.Issues, HealthIssue{
			Type:      "system",
			Message:   fmt.Sprintf("Failed to create temp directory: %v", err),
			Severity:  "error",
			Timestamp: time.Now().Format(time.RFC3339),
		})
		report.Overall = "unhealthy"
		return report
	}
	defer os.RemoveAll(h.tempDir)

	h.logf("🏥 Starting comprehensive deck health check...")

	// Check system dependencies
	depIssues := h.checkSystemDependencies()
	report.Issues = append(report.Issues, depIssues...)

	// Check all tools
	toolIssues := h.checkAllTools()
	report.Issues = append(report.Issues, toolIssues...)
	
	// Count healthy tools
	report.ToolsOK = report.ToolsTotal - len(toolIssues)

	// Test pipeline if tools are available
	if len(toolIssues) == 0 {
		pipelineOK, pipelineIssues := h.TestPipeline()
		report.PipelineOK = pipelineOK
		report.Issues = append(report.Issues, pipelineIssues...)
	}

	// Check fonts and assets
	assetIssues := h.checkAssets()
	report.Issues = append(report.Issues, assetIssues...)

	// Determine overall health
	report.Overall = h.determineOverallHealth(report)
	report.Duration = time.Since(start).String()

	h.logf("✅ Health check completed in %s", report.Duration)
	return report
}

// ValidateTool checks if a specific tool is built and functional
func (h *HealthChecker) ValidateTool(toolName string) error {
	h.logf("🔧 Validating tool: %s", toolName)

	// Get tool path from deck build directory
	toolPath := h.getToolPath(toolName)
	
	// Check if file exists and is executable
	if err := h.checkFileExecutable(toolPath, toolName); err != nil {
		return err
	}

	// Test tool responds to version check
	return h.testToolVersion(toolPath, toolName)
}

// getToolPath returns the path to a deck tool binary
func (h *HealthChecker) getToolPath(toolName string) string {
	// Map tool names to actual binaries using constants
	binaryName := toolName
	switch toolName {
	case "decksh":
		binaryName = DeckshBinary
	case "deckfmt", "dshfmt":
		binaryName = DeckfmtBinary 
	case "decklint", "dshlint":
		binaryName = DecklintBinary
	case "decksvg", "svgdeck":
		binaryName = DecksvgBinary
	case "deckpng", "pngdeck":
		binaryName = DeckpngBinary
	case "deckpdf", "pdfdeck":
		binaryName = DeckpdfBinary
	}
	
	// Return absolute path using constants
	path := filepath.Join(BuildRoot, "bin", binaryName)
	absPath, _ := filepath.Abs(path)
	return absPath
}

// checkSystemDependencies verifies required system tools are available
func (h *HealthChecker) checkSystemDependencies() []HealthIssue {
	var issues []HealthIssue
	
	dependencies := []struct {
		command    string
		name       string
		required   bool
		suggestion string
	}{
		{GitCommand, "Git", true, "Install Git to enable source updates"},
		{GoCommand, "Go compiler", true, "Install Go 1.21+ to build tools from source"},
	}

	for _, dep := range dependencies {
		if _, err := exec.LookPath(dep.command); err != nil {
			severity := "warning"
			if dep.required {
				severity = "error"
			}
			
			issues = append(issues, HealthIssue{
				Type:       "dependency",
				Message:    fmt.Sprintf("%s not found in PATH", dep.name),
				Severity:   severity,
				Suggestion: dep.suggestion,
				Timestamp:  time.Now().Format(time.RFC3339),
			})
		} else {
			h.logf("✅ %s: Available", dep.name)
		}
	}

	return issues
}

// checkAllTools validates all deck tools
func (h *HealthChecker) checkAllTools() []HealthIssue {
	var issues []HealthIssue

	for _, tool := range Tools {
		if err := h.ValidateTool(tool.Name); err != nil {
			issues = append(issues, HealthIssue{
				Type:       "tool",
				Tool:       tool.Name,
				Message:    fmt.Sprintf("Tool validation failed: %v", err),
				Severity:   "error",
				Suggestion: fmt.Sprintf("Run: ./infra deck build install to rebuild %s", tool.Name),
				Timestamp:  time.Now().Format(time.RFC3339),
			})
		} else {
			h.logf("✅ %s: OK", tool.Name)
		}
	}

	return issues
}

// TestPipeline runs a complete .dsh → SVG pipeline test
func (h *HealthChecker) TestPipeline() (bool, []HealthIssue) {
	h.logf("🔄 Testing complete .dsh → SVG pipeline...")
	
	var issues []HealthIssue

	// Create test .dsh content
	testDSH := `deck
	text "Health Check Test" 50 50 2
	circle 25 75 5 "blue"
	rect 75 75 10 5 "red" 0.8
edeck`

	// Write test file
	dshFile := filepath.Join(h.tempDir, "health-test.dsh")
	if err := os.WriteFile(dshFile, []byte(testDSH), 0644); err != nil {
		issues = append(issues, HealthIssue{
			Type:      "pipeline",
			Message:   fmt.Sprintf("Failed to create test .dsh file: %v", err),
			Severity:  "error",
			Timestamp: time.Now().Format(time.RFC3339),
		})
		return false, issues
	}

	// Test .dsh → XML
	xmlFile := filepath.Join(h.tempDir, "health-test.xml")
	if err := h.testDSHToXML(dshFile, xmlFile); err != nil {
		issues = append(issues, HealthIssue{
			Type:      "pipeline",
			Message:   fmt.Sprintf("DSH→XML conversion failed: %v", err),
			Severity:  "error",
			Suggestion: "Check decksh tool and font configuration",
			Timestamp: time.Now().Format(time.RFC3339),
		})
		return false, issues
	}

	// Test XML → SVG
	svgFile := filepath.Join(h.tempDir, "health-test.svg")
	if err := h.testXMLToSVG(xmlFile, svgFile); err != nil {
		issues = append(issues, HealthIssue{
			Type:      "pipeline",
			Message:   fmt.Sprintf("XML→SVG conversion failed: %v", err),
			Severity:  "error",
			Suggestion: "Check svgdeck tool configuration",
			Timestamp: time.Now().Format(time.RFC3339),
		})
		return false, issues
	}

	// Validate SVG output
	if err := h.validateSVGOutput(svgFile); err != nil {
		issues = append(issues, HealthIssue{
			Type:      "pipeline",
			Message:   fmt.Sprintf("SVG output validation failed: %v", err),
			Severity:  "warning",
			Timestamp: time.Now().Format(time.RFC3339),
		})
	}

	h.logf("✅ Pipeline test: Complete .dsh → XML → SVG")
	return len(issues) == 0, issues
}

// checkAssets verifies fonts and other required assets
func (h *HealthChecker) checkAssets() []HealthIssue {
	var issues []HealthIssue

	// Check font directory
	fontDir := filepath.Join(config.GetDataPath(), FontsDirPath)
	if _, err := os.Stat(fontDir); os.IsNotExist(err) {
		issues = append(issues, HealthIssue{
			Type:       "assets",
			Message:    fmt.Sprintf("Font directory not found: %s", fontDir),
			Severity:   "warning",
			Suggestion: "Font directory will be created automatically, but custom fonts won't be available",
			Timestamp:  time.Now().Format(time.RFC3339),
		})
	} else {
		h.logf("✅ Font directory: %s", fontDir)
	}

	// Check output directories are writable
	outputDir := filepath.Join(config.GetDataPath(), CacheDirPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		issues = append(issues, HealthIssue{
			Type:       "assets",
			Message:    fmt.Sprintf("Cannot create output directory: %v", err),
			Severity:   "error",
			Suggestion: "Check file permissions for data directory",
			Timestamp:  time.Now().Format(time.RFC3339),
		})
	} else {
		h.logf("✅ Output directory: %s", outputDir)
	}

	return issues
}

// Helper methods

func (h *HealthChecker) checkFileExecutable(path, toolName string) error {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return fmt.Errorf("binary not found at: %s", path)
	}
	if err != nil {
		return fmt.Errorf("cannot access binary: %w", err)
	}

	// Check if file is executable
	if info.Mode()&0111 == 0 {
		return fmt.Errorf("binary not executable: %s", path)
	}

	return nil
}

func (h *HealthChecker) testToolVersion(path, toolName string) error {
	// Try --version first
	cmd := exec.Command(path, "--version")
	output, err := cmd.CombinedOutput()
	if err == nil && len(output) > 0 {
		h.logf("  Version: %s", strings.TrimSpace(string(output)))
		return nil
	}

	// Fallback to --help
	cmd = exec.Command(path, "--help")
	output, err = cmd.CombinedOutput()
	if err == nil && len(output) > 0 {
		return nil
	}

	// Final test - just run the tool
	cmd = exec.Command(path)
	err = cmd.Start()
	if err != nil {
		return fmt.Errorf("tool cannot be executed: %w", err)
	}
	
	// Kill it immediately
	cmd.Process.Kill()
	cmd.Wait()
	
	return nil
}

func (h *HealthChecker) testDSHToXML(dshFile, xmlFile string) error {
	toolPath := h.getToolPath("decksh")
	
	// Check if tool exists
	if _, err := os.Stat(toolPath); os.IsNotExist(err) {
		return fmt.Errorf("decksh not built: %s", toolPath)
	}

	cmd := exec.Command(toolPath, dshFile)
	cmd.Env = append(os.Environ(), "DECKFONTS="+filepath.Join(config.GetDataPath(), FontsDirPath))
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("decksh execution failed: %w, output: %s", err, string(output))
	}

	// Write XML output
	return os.WriteFile(xmlFile, output, 0644)
}

func (h *HealthChecker) testXMLToSVG(xmlFile, svgFile string) error {
	toolPath := h.getToolPath("decksvg")
	
	// Check if tool exists
	if _, err := os.Stat(toolPath); os.IsNotExist(err) {
		return fmt.Errorf("decksvg not built: %s", toolPath)
	}

	cmd := exec.Command(toolPath, xmlFile)
	cmd.Env = append(os.Environ(), "DECKFONTS="+filepath.Join(config.GetDataPath(), FontsDirPath))
	cmd.Dir = filepath.Dir(svgFile)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("svgdeck execution failed: %w, output: %s", err, string(output))
	}

	// Check if SVG was created
	expectedSVG := strings.TrimSuffix(xmlFile, ".xml") + ".svg"
	if _, err := os.Stat(expectedSVG); err == nil {
		return os.Rename(expectedSVG, svgFile)
	}

	return nil
}

func (h *HealthChecker) validateSVGOutput(svgFile string) error {
	content, err := os.ReadFile(svgFile)
	if err != nil {
		return fmt.Errorf("cannot read SVG file: %w", err)
	}

	svgStr := string(content)
	if !strings.Contains(svgStr, "<svg") || !strings.Contains(svgStr, "</svg>") {
		return fmt.Errorf("invalid SVG structure")
	}

	if !strings.Contains(svgStr, "Health Check Test") {
		return fmt.Errorf("test content not found in SVG")
	}

	h.logf("  SVG size: %d bytes", len(content))
	return nil
}

func (h *HealthChecker) determineOverallHealth(report *HealthReport) string {
	errorCount := 0
	warningCount := 0

	for _, issue := range report.Issues {
		switch issue.Severity {
		case "error":
			errorCount++
		case "warning":
			warningCount++
		}
	}

	if errorCount > 0 {
		return "unhealthy"
	}
	if warningCount > 0 {
		return "degraded"  
	}
	if report.ToolsOK == report.ToolsTotal && report.PipelineOK {
		return "healthy"
	}
	
	return "degraded"
}

func (h *HealthChecker) logf(format string, args ...interface{}) {
	if h.verbose {
		log.Info(fmt.Sprintf(format, args...))
	}
}