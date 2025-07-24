package cmd

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/joeblew999/infra/pkg/store"
)

func EnsureInfraDirectories() error {
	// Create .dep directory
	if err := os.MkdirAll(store.GetDepPath(), 0755); err != nil {
		return fmt.Errorf("failed to create .dep directory: %w", err)
	}
	log.Printf("Ensured directory exists: %s", store.GetDepPath())

	// Create .bin directory
	if err := os.MkdirAll(store.GetBinPath(), 0755); err != nil {
		return fmt.Errorf("failed to create .bin directory: %w", err)
	}
	log.Printf("Ensured directory exists: %s", store.GetBinPath())

	// Create .data directory
	if err := os.MkdirAll(store.GetDataPath(), 0755); err != nil {
		return fmt.Errorf("failed to create .data directory: %w", err)
	}
	log.Printf("Ensured directory exists: %s", store.GetDataPath())

	// Create taskfiles directory
	if err := os.MkdirAll(store.GetTaskfilesPath(), 0755); err != nil {
		return fmt.Errorf("failed to create taskfiles directory: %w", err)
	}
	log.Printf("Ensured directory exists: %s", store.GetTaskfilesPath())

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
	if err := os.Chdir(store.GetTerraformPath()); err != nil {
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
