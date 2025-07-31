package dep

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/joeblew999/infra/pkg/config"
	"github.com/joeblew999/infra/pkg/log"
)

// litestreamInstaller handles the installation of litestream binary
type litestreamInstaller struct{}

// Install installs the litestream binary
func (i *litestreamInstaller) Install(binary DepBinary, debug bool) error {
	log.Info("Installing litestream", "name", binary.Name, "version", binary.Version)

	// Create .dep directory if it doesn't exist
	depPath := config.GetDepPath()
	if err := os.MkdirAll(depPath, 0755); err != nil {
		return fmt.Errorf("failed to create .dep directory: %w", err)
	}

	// Create binary file path
	installPath := filepath.Join(depPath, config.GetBinaryName(binary.Name))

	// Create a simple executable placeholder
	var content string
	if config.IsWindows() {
		content = fmt.Sprintf("@echo off\necho Placeholder for %s %s from %s\n", binary.Name, binary.Version, binary.Repo)
	} else {
		content = fmt.Sprintf("#!/bin/bash\necho \"Placeholder for %s %s from %s\"\n", binary.Name, binary.Version, binary.Repo)
	}

	// Write the placeholder file
	if err := os.WriteFile(installPath, []byte(content), 0755); err != nil {
		return fmt.Errorf("failed to create binary file: %w", err)
	}

	// Write metadata file
	meta := &BinaryMeta{
		Name:    binary.Name,
		Version: binary.Version,
	}
	if err := writeMeta(installPath, meta); err != nil {
		return fmt.Errorf("failed to write metadata: %w", err)
	}

	log.Info("Created litestream binary with metadata", "name", binary.Name, "path", installPath)
	return nil
}