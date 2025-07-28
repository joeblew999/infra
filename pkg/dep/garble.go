package dep

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/joeblew999/infra/pkg/log"
	"github.com/joeblew999/infra/pkg/config"
)

type garbleInstaller struct{}

func (i *garbleInstaller) Install(binary DepBinary, debug bool) error {
	log.Info("Installing using go install", "binary", binary.Name, "version", binary.Version)

	installPath, err := Get(binary.Name)
	if err != nil {
		return fmt.Errorf("failed to get install path for %s: %w", binary.Name, err)
	}

	// Create a temporary GOBIN directory with absolute path
	tmpBin := filepath.Join(config.GetDepPath(), "tmp", "bin")
	tmpBinAbs, err := filepath.Abs(tmpBin)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for %s: %w", tmpBin, err)
	}
	if err := os.MkdirAll(tmpBinAbs, 0755); err != nil {
		return fmt.Errorf("failed to create temporary bin directory: %w", err)
	}
	defer os.RemoveAll(filepath.Dir(tmpBinAbs)) // Clean up tmp directory

	// Construct the go install command with version
	// Garble uses mvdan.cc/garble as its module path, not github.com/burrowers/garble
	var installTarget string
	if binary.Name == "garble" {
		installTarget = fmt.Sprintf("mvdan.cc/garble@%s", binary.Version)
	} else {
		installTarget = fmt.Sprintf("github.com/%s@%s", binary.Repo, binary.Version)
	}
	cmd := exec.Command("go", "install", installTarget)
	
	// Set GOBIN to our temporary directory
	cmd.Env = append(os.Environ(), fmt.Sprintf("GOBIN=%s", tmpBinAbs))
	
	log.Info("Running go install", "target", installTarget, "GOBIN", tmpBinAbs)
	
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to run go install %s: %w\nOutput: %s", installTarget, err, string(output))
	}

	// Determine the binary name with possible .exe extension
	binaryName := binary.Name
	if runtime.GOOS == "windows" {
		binaryName += ".exe"
	}
	
	srcPath := filepath.Join(tmpBinAbs, binaryName)
	
	// Verify the binary was created
	if _, err := os.Stat(srcPath); os.IsNotExist(err) {
		return fmt.Errorf("binary %s was not created in %s", binaryName, tmpBinAbs)
	}

	// Move the binary to its final destination
	if err := os.Rename(srcPath, installPath); err != nil {
		return fmt.Errorf("failed to move binary from %s to %s: %w", srcPath, installPath, err)
	}

	if err := os.Chmod(installPath, 0755); err != nil {
		return fmt.Errorf("failed to set executable permissions for %s: %w", installPath, err)
	}

	log.Info("Successfully installed binary", "binary", binary.Name, "path", installPath)
	return nil
}