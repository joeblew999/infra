package cmd

import (
	"fmt"
	"log"
	"os"
	"os/exec"

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
	cmd := exec.Command(binary, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
