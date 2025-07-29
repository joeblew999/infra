package tofu

import (
	"fmt"
	"os/exec"
	"path/filepath"

	"github.com/joeblew999/infra/pkg/config"
)

// Runner executes tofu commands
type Runner struct {
	binaryPath string
	workingDir string
}

// New creates a new tofu runner
func New() *Runner {
	// Convert to absolute paths
	binaryPath, _ := filepath.Abs(config.GetTofuBinPath())
	workingDir, _ := filepath.Abs(config.GetTerraformPath())

	return &Runner{
		binaryPath: binaryPath,
		workingDir: workingDir,
	}
}

// Run executes a tofu command with the given arguments
func (r *Runner) Run(args ...string) error {
	cmd := exec.Command(r.binaryPath, args...)
	cmd.Dir = r.workingDir

	// Inherit stdout/stderr so we can see output
	cmd.Stdout = nil
	cmd.Stderr = nil

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("tofu command failed: %w", err)
	}

	return nil
}

// RunTofu executes the tofu command with the given arguments
func RunTofu(args []string) error {
	runner := New()
	return runner.Run(args...)
}

// RunWithOutput executes a tofu command and returns the output
func (r *Runner) RunWithOutput(args ...string) ([]byte, error) {
	cmd := exec.Command(r.binaryPath, args...)
	cmd.Dir = r.workingDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return output, fmt.Errorf("tofu command failed: %w", err)
	}

	return output, nil
}
