package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/joeblew999/infra/pkg/config"
	"github.com/spf13/cobra"
)

func registerToolCommands(cliParent *cobra.Command) {
	cliParent.AddCommand(newBinaryProxyCmd("tofu", config.GetTofuBinPath))
	cliParent.AddCommand(newBinaryProxyCmd("task", config.GetTaskBinPath))
	cliParent.AddCommand(newBinaryProxyCmd("ko", config.GetKoBinPath))
	cliParent.AddCommand(newBinaryProxyCmd("flyctl", config.GetFlyctlBinPath))
}

func newBinaryProxyCmd(name string, pathFunc func() string) *cobra.Command {
	return &cobra.Command{
		Use:                name,
		Short:              fmt.Sprintf("Run %s commands", name),
		DisableFlagParsing: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return executeBinary(pathFunc(), args...)
		},
	}
}

func executeBinary(binary string, args ...string) error {
	oldDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory: %w", err)
	}

	absBinary, err := filepath.Abs(binary)
	if err != nil {
		return fmt.Errorf("failed to get absolute path of binary: %w", err)
	}

	if err := os.Chdir(config.GetTerraformPath()); err != nil {
		return fmt.Errorf("failed to change directory to terraform: %w", err)
	}

	cmd := exec.Command(absBinary, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	runErr := cmd.Run()

	if err := os.Chdir(oldDir); err != nil {
		return fmt.Errorf("failed to change back to original directory: %w", err)
	}

	return runErr
}
