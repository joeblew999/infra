package ai

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/joeblew999/infra/pkg/dep"
	"github.com/joeblew999/infra/pkg/log"
)

// GooseRunner executes goose commands with proper binary path resolution
type GooseRunner struct {
	binaryPath string
}

// NewGooseRunner creates a new goose runner
func NewGooseRunner() *GooseRunner {
	// Get the goose binary path from dep system
	binaryPath, err := dep.Get("goose")
	if err != nil {
		log.Warn("Could not get goose binary path", "error", err)
		// Fallback to system goose if available
		binaryPath = "goose"
	}
	
	return &GooseRunner{
		binaryPath: binaryPath,
	}
}

// Run executes a goose command with the given arguments
func (r *GooseRunner) Run(args ...string) error {
	cmd := exec.Command(r.binaryPath, args...)
	cmd.Stdout = nil
	cmd.Stderr = nil

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("goose command failed: %w", err)
	}
	return nil
}

// RunWithOutput executes a goose command and returns the output
func (r *GooseRunner) RunWithOutput(args ...string) ([]byte, error) {
	cmd := exec.Command(r.binaryPath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return output, fmt.Errorf("goose command failed: %w", err)
	}
	return output, nil
}

// RunInteractive executes a goose command with interactive input/output
func (r *GooseRunner) RunInteractive(args ...string) error {
	cmd := exec.Command(r.binaryPath, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("goose interactive command failed: %w", err)
	}
	return nil
}

// Session starts or resumes a Goose session
func (r *GooseRunner) Session(sessionName string) error {
	args := []string{"session"}
	if sessionName != "" {
		args = append(args, sessionName)
	}
	
	log.Info("Starting Goose session", "session", sessionName)
	return r.RunInteractive(args...)
}

// RunFile executes Goose commands from a file
func (r *GooseRunner) RunFile(filename string) error {
	args := []string{"run", filename}
	
	log.Info("Running Goose from file", "file", filename)
	return r.RunInteractive(args...)
}

// RunStdin executes Goose commands from stdin
func (r *GooseRunner) RunStdin() error {
	args := []string{"run"}
	
	log.Info("Running Goose from stdin")
	return r.RunInteractive(args...)
}

// Configure runs Goose configuration setup
func (r *GooseRunner) Configure() error {
	args := []string{"configure"}
	
	log.Info("Configuring Goose")
	return r.RunInteractive(args...)
}

// Info displays Goose information
func (r *GooseRunner) Info() error {
	args := []string{"info"}
	
	return r.RunInteractive(args...)
}

// Web starts the Goose web interface
func (r *GooseRunner) Web() error {
	args := []string{"web"}
	
	log.Info("Starting Goose web interface")
	fmt.Println("🌐 Starting Goose web interface...")
	fmt.Println("   This will start a local web server for browser-based interaction")
	return r.RunInteractive(args...)
}

// Version gets the Goose version
func (r *GooseRunner) Version() (string, error) {
	output, err := r.RunWithOutput("--version")
	if err != nil {
		return "", err
	}
	return string(output), nil
}

// ListSessions lists available Goose sessions
func (r *GooseRunner) ListSessions() error {
	// This would use a sessions command if available
	// For now, we'll use the projects command as a proxy
	args := []string{"projects"}
	
	log.Info("Listing Goose sessions/projects")
	return r.RunInteractive(args...)
}

// Schedule manages scheduled Goose jobs
func (r *GooseRunner) Schedule(action string, args ...string) error {
	schedArgs := []string{"schedule", action}
	schedArgs = append(schedArgs, args...)
	
	log.Info("Managing Goose schedule", "action", action)
	return r.RunInteractive(schedArgs...)
}

// Benchmark runs Goose system benchmarks
func (r *GooseRunner) Benchmark() error {
	args := []string{"bench"}
	
	log.Info("Running Goose benchmarks")
	fmt.Println("🏃 Running Goose system benchmarks...")
	fmt.Println("   This will evaluate system configuration across practical tasks")
	return r.RunInteractive(args...)
}

// MCP manages MCP servers bundled with Goose
func (r *GooseRunner) MCP(serverName string, args ...string) error {
	mcpArgs := []string{"mcp", serverName}
	mcpArgs = append(mcpArgs, args...)
	
	log.Info("Running Goose MCP server", "server", serverName)
	return r.RunInteractive(mcpArgs...)
}

// Recipe manages Goose recipes
func (r *GooseRunner) Recipe(action string, args ...string) error {
	recipeArgs := []string{"recipe", action}
	recipeArgs = append(recipeArgs, args...)
	
	log.Info("Managing Goose recipe", "action", action)
	return r.RunInteractive(recipeArgs...)
}

// Update updates the Goose CLI version
func (r *GooseRunner) Update() error {
	args := []string{"update"}
	
	log.Info("Updating Goose CLI")
	fmt.Println("🔄 Updating Goose CLI...")
	return r.RunInteractive(args...)
}