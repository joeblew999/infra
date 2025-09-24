package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/joeblew999/infra/pkg/config"
	"github.com/joeblew999/infra/pkg/dep"
	"github.com/spf13/cobra"
)

// NewNSCCmd returns a cobra command that proxies to the nsc binary.
func NewNSCCmd() *cobra.Command {
	return &cobra.Command{
		Use:                "nsc",
		Short:              "NATS credentials management CLI",
		DisableFlagParsing: true,
		SilenceUsage:       true,
		SilenceErrors:      true,
		Args:               cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := dep.InstallBinary(config.BinaryNsc, false); err != nil {
				return fmt.Errorf("failed to install nsc binary: %w", err)
			}

			path, err := dep.Get(config.BinaryNsc)
			if err != nil {
				return fmt.Errorf("failed to resolve nsc binary path: %w", err)
			}

			absPath, err := filepath.Abs(path)
			if err != nil {
				return fmt.Errorf("failed to make nsc path absolute: %w", err)
			}

			execCmd := exec.Command(absPath, args...)
			execCmd.Stdout = os.Stdout
			execCmd.Stderr = os.Stderr
			execCmd.Stdin = os.Stdin
			execCmd.Env = append(os.Environ(),
				fmt.Sprintf("NSC_STORE_DIR=%s", config.GetNATSAuthStorePath()),
				"NSC_NO_GITHUB_UPDATES=1",
			)
			return execCmd.Run()
		},
	}
}
