package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/joeblew999/infra/pkg/log"
	"github.com/joeblew999/infra/pkg/config"
)

func EnsureInfraDirectories() error {
	// Create .dep directory
	if err := os.MkdirAll(config.GetDepPath(), 0755); err != nil {
		return fmt.Errorf("failed to create .dep directory: %w", err)
	}
	log.Info("Ensured directory exists", "path", config.GetDepPath())

	// Create .bin directory
	if err := os.MkdirAll(config.GetBinPath(), 0755); err != nil {
		return fmt.Errorf("failed to create .bin directory: %w", err)
	}
	log.Info("Ensured directory exists", "path", config.GetBinPath())

	// Create .data directory
	if err := os.MkdirAll(config.GetDataPath(), 0755); err != nil {
		return fmt.Errorf("failed to create .data directory: %w", err)
	}
	log.Info("Ensured directory exists", "path", config.GetDataPath())

	// Create taskfiles directory
	if err := os.MkdirAll(config.GetTaskfilesPath(), 0755); err != nil {
		return fmt.Errorf("failed to create taskfiles directory: %w", err)
	}
	log.Info("Ensured directory exists", "path", config.GetTaskfilesPath())

	return nil
}

func ExecuteBinary(binary string, args ...string) error {
	// Save current working directory
	oldDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory: %w", err)
	}

	// Get absolute path of the binary before changing directory
	absoluteBinaryPath, err := filepath.Abs(binary)
	if err != nil {
		return fmt.Errorf("failed to get absolute path of binary: %w", err)
	}

	// Change to the terraform directory
	if err := os.Chdir(config.GetTerraformPath()); err != nil {
		return fmt.Errorf("failed to change directory to terraform: %w", err)
	}

	cmd := exec.Command(absoluteBinaryPath, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()

	// Change back to the original working directory
	if err := os.Chdir(oldDir); err != nil {
		return fmt.Errorf("failed to change back to original directory: %w", err)
	}

	return err
}
