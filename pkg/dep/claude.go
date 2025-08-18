package dep

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/joeblew999/infra/pkg/log"
)

// claudeInstaller handles installation of Claude Code CLI using the native install script
// This uses Anthropic's official installation method from claude.ai/install.sh
type claudeInstaller struct{}

// Install downloads and installs Claude Code using the native installation script
func (i *claudeInstaller) Install(binary DepBinary, debug bool) error {
	log.Info("Installing Claude Code using native installer", "platform", runtime.GOOS)

	// Get install path - Claude installs to a standard location but we'll symlink to our .dep directory
	installPath, err := Get(binary.Name)
	if err != nil {
		return fmt.Errorf("failed to get install path for %s: %w", binary.Name, err)
	}

	// Ensure .dep directory exists
	installDir := filepath.Dir(installPath)
	if err := os.MkdirAll(installDir, 0755); err != nil {
		return fmt.Errorf("failed to create install directory: %w", err)
	}

	// Use platform-specific installation method
	switch runtime.GOOS {
	case "windows":
		return i.installWindows(installPath, debug)
	case "darwin", "linux":
		return i.installUnix(installPath, debug)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

// installUnix installs Claude Code on Unix-like systems (macOS, Linux, WSL)
func (i *claudeInstaller) installUnix(installPath string, debug bool) error {
	// Create a temporary directory for installation
	tempDir, err := os.MkdirTemp("", "claude-install")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a custom installation script that installs to our temp directory
	tempScript := filepath.Join(tempDir, "claude_install.sh")
	script := fmt.Sprintf(`#!/bin/bash
set -e
export CLAUDE_INSTALL_DIR="%s"
curl -fsSL https://claude.ai/install.sh | bash
`, tempDir)

	if err := os.WriteFile(tempScript, []byte(script), 0755); err != nil {
		return fmt.Errorf("failed to create install script: %w", err)
	}

	// Run the installation script
	cmd := exec.Command("bash", tempScript)
	if debug {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	log.Info("Running Claude Code native installer...")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("native installation failed: %w", err)
	}

	// Find the claude binary in the temp directory and move it to our .dep directory
	return i.moveBinaryToDepDir(tempDir, installPath)
}

// installWindows installs Claude Code on Windows using PowerShell
func (i *claudeInstaller) installWindows(installPath string, debug bool) error {
	// Create a temporary directory for installation
	tempDir, err := os.MkdirTemp("", "claude-install")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Create PowerShell command with custom install directory
	psScript := fmt.Sprintf(`$env:CLAUDE_INSTALL_DIR="%s"; irm https://claude.ai/install.ps1 | iex`, tempDir)

	cmd := exec.Command("powershell", "-Command", psScript)
	if debug {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	log.Info("Running Claude Code native installer on Windows...")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("native installation failed: %w", err)
	}

	// Find the claude binary in the temp directory and move it to our .dep directory
	return i.moveBinaryToDepDir(tempDir, installPath)
}

// moveBinaryToDepDir finds the claude binary in the installation directory and moves it to our .dep folder
func (i *claudeInstaller) moveBinaryToDepDir(searchDir, installPath string) error {
	// Look for the claude binary in common locations within the search directory
	possiblePaths := []string{
		filepath.Join(searchDir, "claude"),
		filepath.Join(searchDir, "bin", "claude"),
		filepath.Join(searchDir, ".local", "bin", "claude"),
		filepath.Join(searchDir, "claude.exe"), // Windows
		filepath.Join(searchDir, "bin", "claude.exe"), // Windows
	}

	var foundPath string
	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			foundPath = path
			break
		}
	}

	// If not found in expected locations, search recursively
	if foundPath == "" {
		var err error
		foundPath, err = i.findClaudeBinary(searchDir)
		if err != nil {
			return fmt.Errorf("failed to find claude binary in installation directory: %w", err)
		}
	}

	if foundPath == "" {
		return fmt.Errorf("claude binary not found after installation")
	}

	// Copy the binary to our install path
	sourceFile, err := os.Open(foundPath)
	if err != nil {
		return fmt.Errorf("failed to open source binary: %w", err)
	}
	defer sourceFile.Close()

	destFile, err := os.Create(installPath)
	if err != nil {
		return fmt.Errorf("failed to create destination binary: %w", err)
	}
	defer destFile.Close()

	// Copy the file
	if _, err := io.Copy(destFile, sourceFile); err != nil {
		return fmt.Errorf("failed to copy binary: %w", err)
	}

	// Make it executable
	if runtime.GOOS != "windows" {
		if err := os.Chmod(installPath, 0755); err != nil {
			return fmt.Errorf("failed to make binary executable: %w", err)
		}
	}

	log.Info("Claude Code installed successfully", "path", installPath)
	return nil
}

// findClaudeBinary recursively searches for the claude binary
func (i *claudeInstaller) findClaudeBinary(searchDir string) (string, error) {
	var foundPath string
	
	err := filepath.Walk(searchDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Continue searching even if we hit an error
		}
		
		if info.IsDir() {
			return nil
		}
		
		filename := info.Name()
		if filename == "claude" || filename == "claude.exe" {
			// Verify it's executable
			if info.Mode().Perm()&0111 != 0 || runtime.GOOS == "windows" {
				foundPath = path
				return filepath.SkipDir // Found it, stop searching
			}
		}
		
		return nil
	})
	
	return foundPath, err
}
