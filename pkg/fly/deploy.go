package fly

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/joeblew999/infra/pkg/config"
	"github.com/joeblew999/infra/pkg/dep"
)

// DeployNATSCluster deploys a NATS cluster node to Fly.io
func DeployNATSCluster(appName, region string) error {
	// Ensure flyctl is available
	if err := dep.InstallBinary(config.BinaryFlyctl, false); err != nil {
		return fmt.Errorf("failed to ensure flyctl: %w", err)
	}

	// Create app if it doesn't exist
	if err := ensureApp(appName); err != nil {
		return fmt.Errorf("failed to ensure app %s: %w", appName, err)
	}

	// Generate and write fly.toml
	flyToml := GetNATSClusterTemplate(appName, region)
	flyTomlPath := filepath.Join(os.TempDir(), fmt.Sprintf("fly-%s.toml", appName))

	if err := os.WriteFile(flyTomlPath, []byte(flyToml), 0644); err != nil {
		return fmt.Errorf("failed to write fly.toml: %w", err)
	}
	defer os.Remove(flyTomlPath)

	// Deploy using flyctl
	cmd := exec.Command(config.GetFlyctlBinPath(), "deploy", "--config", flyTomlPath, "--app", appName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// DeployAppServer deploys an application server to Fly.io
func DeployAppServer(appName, region string) error {
	// Ensure flyctl is available
	if err := dep.InstallBinary(config.BinaryFlyctl, false); err != nil {
		return fmt.Errorf("failed to ensure flyctl: %w", err)
	}

	// Create app if it doesn't exist
	if err := ensureApp(appName); err != nil {
		return fmt.Errorf("failed to ensure app %s: %w", appName, err)
	}

	// Generate and write fly.toml
	flyToml := GetAppServerTemplate(appName, region)
	flyTomlPath := filepath.Join(os.TempDir(), fmt.Sprintf("fly-%s.toml", appName))

	if err := os.WriteFile(flyTomlPath, []byte(flyToml), 0644); err != nil {
		return fmt.Errorf("failed to write fly.toml: %w", err)
	}
	defer os.Remove(flyTomlPath)

	// Deploy using flyctl
	cmd := exec.Command(config.GetFlyctlBinPath(), "deploy", "--config", flyTomlPath, "--app", appName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// ensureApp creates a Fly app if it doesn't exist
func ensureApp(appName string) error {
	// Check if app exists
	checkCmd := exec.Command(config.GetFlyctlBinPath(), "status", "--app", appName)
	if err := checkCmd.Run(); err == nil {
		// App exists
		return nil
	}

	// App doesn't exist, create it
	createCmd := exec.Command(config.GetFlyctlBinPath(), "apps", "create", appName, "--org", "personal")
	createCmd.Stdout = os.Stdout
	createCmd.Stderr = os.Stderr
	return createCmd.Run()
}

// GetAppStatus returns the status of a Fly app
func GetAppStatus(appName string) (string, error) {
	if err := dep.InstallBinary(config.BinaryFlyctl, false); err != nil {
		return "unknown", fmt.Errorf("failed to ensure flyctl: %w", err)
	}

	cmd := exec.Command(config.GetFlyctlBinPath(), "status", "--app", appName, "--json")
	_, err := cmd.Output()
	if err != nil {
		return "unknown", fmt.Errorf("flyctl status failed: %w", err)
	}

	// TODO: Parse JSON output to determine actual status
	// For now, just return success if command succeeded
	return "running", nil
}