package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/joeblew999/infra/pkg/dep"
	"github.com/spf13/cobra"
)

// NewCLICmd returns a cobra command that proxies directly to the upstream CLI.
func NewCLICmd() *cobra.Command {
	natsCmd := &cobra.Command{
		Use:   "nats",
		Short: "NATS CLI passthrough",
		Long: `Wrapper around the upstream NATS CLI.

Examples:
  infra cli nats server info
  infra cli nats stream list
  infra cli nats schema list
  infra cli nats --help

Cluster management commands live at the root level, e.g.
  go run . cluster local start`,
		SilenceUsage:       true,
		SilenceErrors:      true,
		DisableFlagParsing: true,
		Args:               cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCLI(args...)
		},
	}

	return natsCmd
}

func runCLI(args ...string) error {
	if err := dep.InstallBinary("nats", false); err != nil {
		return fmt.Errorf("failed to install nats binary: %w", err)
	}

	binaryPath, err := dep.Get("nats")
	if err != nil {
		return fmt.Errorf("failed to get nats binary path: %w", err)
	}

	absPath, err := filepath.Abs(binaryPath)
	if err != nil {
		return fmt.Errorf("failed to make binary path absolute: %w", err)
	}

	cmd := exec.Command(absPath, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}
