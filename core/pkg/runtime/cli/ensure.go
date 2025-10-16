package cli

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"
)

func newEnsureCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ensure [tool]",
		Short: "Install deployment tools (ko, flyctl)",
		Long: `Install deployment tools to .dep/ directory.

Available tools:
  ko      - Container image builder for Go
  flyctl  - Fly.io CLI tool
  all     - Install all tools

Examples:
  go run . ensure ko      # Install ko only
  go run . ensure flyctl  # Install flyctl only
  go run . ensure all     # Install both tools`,
		Args: cobra.ExactArgs(1),
		RunE: ensureRun,
	}

	cmd.Flags().Bool("force", false, "Force reinstall even if tool exists")

	return cmd
}

func ensureRun(cmd *cobra.Command, args []string) error {
	tool := args[0]
	force, _ := cmd.Flags().GetBool("force")
	out := cmd.OutOrStdout()

	// Create .dep directory if it doesn't exist
	depDir := ".dep"
	if err := os.MkdirAll(depDir, 0755); err != nil {
		return fmt.Errorf("create .dep directory: %w", err)
	}

	switch tool {
	case "ko":
		return installKo(out, depDir, force)
	case "flyctl":
		return installFlyctl(out, depDir, force)
	case "all":
		fmt.Fprintln(out, "📦 Installing all deployment tools...\n")
		if err := installKo(out, depDir, force); err != nil {
			return err
		}
		fmt.Fprintln(out)
		if err := installFlyctl(out, depDir, force); err != nil {
			return err
		}
		fmt.Fprintln(out, "\n✅ All tools installed successfully!")
		return nil
	default:
		return fmt.Errorf("unknown tool: %s (available: ko, flyctl, all)", tool)
	}
}

func installKo(out io.Writer, depDir string, force bool) error {
	koPath := filepath.Join(depDir, "ko")

	// Check if already installed
	if !force {
		if _, err := os.Stat(koPath); err == nil {
			// Verify it works
			if err := exec.Command(koPath, "version").Run(); err == nil {
				fmt.Fprintf(out, "✓ ko already installed at %s\n", koPath)
				fmt.Fprintln(out, "  Use --force to reinstall")
				return nil
			}
		}
	}

	fmt.Fprintln(out, "📦 Installing ko...")

	// Use go install to build ko
	fmt.Fprintln(out, "  Building ko from source (this may take a minute)...")

	installCmd := exec.Command("go", "install", "github.com/google/ko@latest")
	installCmd.Stdout = out
	installCmd.Stderr = out

	if err := installCmd.Run(); err != nil {
		return fmt.Errorf("install ko: %w", err)
	}

	// Find where go install put it
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("get home directory: %w", err)
		}
		gopath = filepath.Join(homeDir, "go")
	}

	goBinPath := filepath.Join(gopath, "bin", "ko")

	// Copy to .dep/
	if err := copyFile(goBinPath, koPath); err != nil {
		return fmt.Errorf("copy ko to .dep: %w", err)
	}

	// Make executable
	if err := os.Chmod(koPath, 0755); err != nil {
		return fmt.Errorf("make ko executable: %w", err)
	}

	fmt.Fprintf(out, "✅ ko installed successfully to %s\n", koPath)

	// Show version
	versionCmd := exec.Command(koPath, "version")
	versionCmd.Stdout = out
	versionCmd.Run()

	return nil
}

func installFlyctl(out io.Writer, depDir string, force bool) error {
	flyctlPath := filepath.Join(depDir, "flyctl")

	// Check if already installed
	if !force {
		if _, err := os.Stat(flyctlPath); err == nil {
			// Verify it works
			if err := exec.Command(flyctlPath, "version").Run(); err == nil {
				fmt.Fprintf(out, "✓ flyctl already installed at %s\n", flyctlPath)
				fmt.Fprintln(out, "  Use --force to reinstall")
				return nil
			}
		}
	}

	fmt.Fprintln(out, "📦 Installing flyctl...")

	// Determine download URL based on platform
	var downloadURL string
	switch runtime.GOOS {
	case "darwin":
		switch runtime.GOARCH {
		case "amd64":
			downloadURL = "https://github.com/superfly/flyctl/releases/latest/download/flyctl_Darwin_x86_64.tar.gz"
		case "arm64":
			downloadURL = "https://github.com/superfly/flyctl/releases/latest/download/flyctl_Darwin_arm64.tar.gz"
		default:
			return fmt.Errorf("unsupported macOS architecture: %s", runtime.GOARCH)
		}
	case "linux":
		switch runtime.GOARCH {
		case "amd64":
			downloadURL = "https://github.com/superfly/flyctl/releases/latest/download/flyctl_Linux_x86_64.tar.gz"
		case "arm64":
			downloadURL = "https://github.com/superfly/flyctl/releases/latest/download/flyctl_Linux_arm64.tar.gz"
		default:
			return fmt.Errorf("unsupported Linux architecture: %s", runtime.GOARCH)
		}
	case "windows":
		return fmt.Errorf("Windows not yet supported - please download flyctl manually from https://fly.io/docs/hands-on/install-flyctl/")
	default:
		return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}

	fmt.Fprintf(out, "  Downloading from %s...\n", downloadURL)

	// Download tar.gz
	resp, err := http.Get(downloadURL)
	if err != nil {
		return fmt.Errorf("download flyctl: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed: HTTP %d", resp.StatusCode)
	}

	// Save to temporary file
	tmpFile, err := os.CreateTemp("", "flyctl-*.tar.gz")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		return fmt.Errorf("save download: %w", err)
	}
	tmpFile.Close()

	fmt.Fprintln(out, "  Extracting...")

	// Extract tar.gz
	extractCmd := exec.Command("tar", "-xzf", tmpFile.Name(), "-C", depDir, "flyctl")
	if err := extractCmd.Run(); err != nil {
		return fmt.Errorf("extract flyctl: %w", err)
	}

	// Make executable
	if err := os.Chmod(flyctlPath, 0755); err != nil {
		return fmt.Errorf("make flyctl executable: %w", err)
	}

	fmt.Fprintf(out, "✅ flyctl installed successfully to %s\n", flyctlPath)

	// Show version
	versionCmd := exec.Command(flyctlPath, "version")
	versionCmd.Stdout = out
	versionCmd.Run()

	return nil
}

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, sourceFile); err != nil {
		return err
	}

	return destFile.Sync()
}
