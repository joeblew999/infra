package playwright

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// EnsureDependencies checks if required tools are installed and optionally installs them.
// Returns the path to the tool or error if installation fails.
func EnsureDependencies(workflow WorkflowMode, depDir string, autoInstall bool, out io.Writer) error {
	if depDir == "" {
		depDir = ".dep"
	}

	if err := os.MkdirAll(depDir, 0755); err != nil {
		return fmt.Errorf("create dependency directory: %w", err)
	}

	switch workflow {
	case "", WorkflowBun:
		return ensureBun(depDir, autoInstall, out)
	case WorkflowNode:
		return ensurePnpm(depDir, autoInstall, out)
	default:
		return fmt.Errorf("unsupported workflow: %s", workflow)
	}
}

// ensureBun checks if bun is installed, optionally installing it.
func ensureBun(depDir string, autoInstall bool, out io.Writer) error {
	// Check if bun is already in PATH
	if _, err := exec.LookPath("bun"); err == nil {
		if out != nil {
			fmt.Fprintln(out, "✓ bun already installed (found in PATH)")
		}
		return nil
	}

	// Check if we have it in .dep/
	bunPath := filepath.Join(depDir, "bun")
	if _, err := os.Stat(bunPath); err == nil {
		if out != nil {
			fmt.Fprintf(out, "✓ bun found at %s\n", bunPath)
		}
		return nil
	}

	if !autoInstall {
		return fmt.Errorf("bun not found: install with 'brew install oven-sh/bun/bun' or enable auto-install")
	}

	if out != nil {
		fmt.Fprintln(out, "📦 Installing bun...")
	}

	// Download and install bun
	return installBunTo(depDir, out)
}

// ensurePnpm checks if pnpm is installed.
func ensurePnpm(depDir string, autoInstall bool, out io.Writer) error {
	// Check if pnpm is already in PATH
	if _, err := exec.LookPath("pnpm"); err == nil {
		if out != nil {
			fmt.Fprintln(out, "✓ pnpm already installed (found in PATH)")
		}
		return nil
	}

	if !autoInstall {
		return fmt.Errorf("pnpm not found: install with 'npm install -g pnpm' or enable auto-install")
	}

	if out != nil {
		fmt.Fprintln(out, "📦 Installing pnpm via npm...")
	}

	// Install pnpm globally via npm
	cmd := exec.Command("npm", "install", "-g", "pnpm")
	cmd.Stdout = out
	cmd.Stderr = out

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("install pnpm: %w", err)
	}

	if out != nil {
		fmt.Fprintln(out, "✅ pnpm installed successfully")
	}

	return nil
}

// installBunTo downloads and installs bun to the specified directory.
func installBunTo(depDir string, out io.Writer) error {
	var downloadURL string

	// Determine platform and architecture
	switch runtime.GOOS {
	case "darwin":
		switch runtime.GOARCH {
		case "amd64":
			downloadURL = "https://github.com/oven-sh/bun/releases/latest/download/bun-darwin-x64.zip"
		case "arm64":
			downloadURL = "https://github.com/oven-sh/bun/releases/latest/download/bun-darwin-aarch64.zip"
		default:
			return fmt.Errorf("unsupported darwin architecture: %s", runtime.GOARCH)
		}
	case "linux":
		switch runtime.GOARCH {
		case "amd64":
			downloadURL = "https://github.com/oven-sh/bun/releases/latest/download/bun-linux-x64.zip"
		case "arm64":
			downloadURL = "https://github.com/oven-sh/bun/releases/latest/download/bun-linux-aarch64.zip"
		default:
			return fmt.Errorf("unsupported linux architecture: %s", runtime.GOARCH)
		}
	default:
		return fmt.Errorf("unsupported OS: %s (use 'curl -fsSL https://bun.sh/install | bash')", runtime.GOOS)
	}

	if out != nil {
		fmt.Fprintf(out, "  Downloading from %s...\n", downloadURL)
	}

	// Download the zip file
	resp, err := http.Get(downloadURL)
	if err != nil {
		return fmt.Errorf("download bun: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download bun: HTTP %d", resp.StatusCode)
	}

	// Save to temporary file
	tmpFile, err := os.CreateTemp("", "bun-*.zip")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		return fmt.Errorf("download bun: %w", err)
	}
	tmpFile.Close()

	// Extract using unzip
	if out != nil {
		fmt.Fprintln(out, "  Extracting...")
	}

	extractCmd := exec.Command("unzip", "-o", tmpFile.Name(), "-d", depDir)
	extractCmd.Stdout = out
	extractCmd.Stderr = out

	if err := extractCmd.Run(); err != nil {
		return fmt.Errorf("extract bun: %w", err)
	}

	// Move binary to correct location
	// The zip contains bun-{os}-{arch}/bun
	var bunDirName string
	switch runtime.GOOS {
	case "darwin":
		if runtime.GOARCH == "arm64" {
			bunDirName = "bun-darwin-aarch64"
		} else {
			bunDirName = "bun-darwin-x64"
		}
	case "linux":
		if runtime.GOARCH == "arm64" {
			bunDirName = "bun-linux-aarch64"
		} else {
			bunDirName = "bun-linux-x64"
		}
	}

	srcPath := filepath.Join(depDir, bunDirName, "bun")
	dstPath := filepath.Join(depDir, "bun")

	if err := os.Rename(srcPath, dstPath); err != nil {
		return fmt.Errorf("move bun binary: %w", err)
	}

	// Clean up extracted directory
	os.RemoveAll(filepath.Join(depDir, bunDirName))

	// Make executable
	if err := os.Chmod(dstPath, 0755); err != nil {
		return fmt.Errorf("chmod bun: %w", err)
	}

	if out != nil {
		fmt.Fprintf(out, "✅ bun installed to %s\n", dstPath)
	}

	return nil
}
