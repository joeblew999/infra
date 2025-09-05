package toki

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/joeblew999/infra/pkg/dep"
)

// Runner provides toki CLI operations
type Runner struct {
	tokiPath string
}

// New creates a new toki runner
func New() (*Runner, error) {
	tokiPath, err := dep.Get("toki")
	if err != nil {
		return nil, fmt.Errorf("failed to get toki: %w", err)
	}
	return &Runner{tokiPath: tokiPath}, nil
}

// Generate runs toki generate command
func (r *Runner) Generate(modulePath string, sourceLang string, targetLangs []string) error {
	args := []string{
		"generate",
		"-m", modulePath,
		"-l", sourceLang,
	}
	
	for _, lang := range targetLangs {
		args = append(args, "-t", lang)
	}
	
	cmd := exec.Command(r.tokiPath, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	return cmd.Run()
}

// Lint runs toki lint command
func (r *Runner) Lint(modulePath string) error {
	cmd := exec.Command(r.tokiPath, "lint", "-m", modulePath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	return cmd.Run()
}

// WebEdit runs toki webedit command
func (r *Runner) WebEdit(modulePath string) error {
	cmd := exec.Command(r.tokiPath, "webedit", "-m", modulePath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	return cmd.Run()
}

// AvailableCommands returns the list of available toki commands
func (r *Runner) AvailableCommands() []string {
	return []string{"generate", "lint", "webedit"}
}

// Version returns the toki version
func (r *Runner) Version() (string, error) {
	// Toki doesn't have a --version flag, so we return the version from dep config
	return "v0.8.3", nil
}