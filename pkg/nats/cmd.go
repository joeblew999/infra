package nats

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/joeblew999/infra/pkg/dep"
	"github.com/spf13/cobra"
)

// NewNATSCmd creates a one-to-one wrapper for the NATS CLI
func NewNATSCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "nats",
		Short: "NATS CLI passthrough",
		Long: `One-to-one wrapper for the NATS CLI.

All NATS CLI commands and flags are passed through directly.
Examples:
  go run . nats server info
  go run . nats stream list
  go run . nats consumer list
  go run . nats schema list
  go run . nats --help`,
		Run: func(cmd *cobra.Command, args []string) {
			// Pass through all arguments to nats CLI
			if err := runNATSCommand(args...); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		},
		// Allow all flags to be passed through to nats CLI
		DisableFlagParsing: true,
	}
}

// runNATSCommand executes the nats CLI with the given arguments
func runNATSCommand(args ...string) error {
	// Ensure nats binary is installed
	if err := dep.InstallBinary("nats", false); err != nil {
		return fmt.Errorf("failed to install nats binary: %w", err)
	}

	// Get the nats binary path
	binaryPath, err := dep.Get("nats")
	if err != nil {
		return fmt.Errorf("failed to get nats binary path: %w", err)
	}

	// Convert to absolute path
	absBinaryPath, err := filepath.Abs(binaryPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Create and execute command - pure passthrough
	cmd := exec.Command(absBinaryPath, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}