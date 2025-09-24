package cmd

import (
	workflowapi "github.com/joeblew999/infra/pkg/workflows/api"
	"github.com/spf13/cobra"
)

// RunDev mounts developer-focused utilities onto the root command.
func RunDev() {
	rootCmd.AddCommand(newDevCmd())
}

func newDevCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dev",
		Short: "Developer utilities and diagnostics",
	}

	cmd.AddCommand(newAPICheckCmd())
	return cmd
}

func newAPICheckCmd() *cobra.Command {
	var oldCommit string
	var newCommit string

	cmd := &cobra.Command{
		Use:   "api-check",
		Short: "Check Go API compatibility between commits",
		RunE: func(cmd *cobra.Command, args []string) error {
			return workflowapi.CheckCompatibility(oldCommit, newCommit)
		},
	}

	cmd.Flags().StringVar(&oldCommit, "old", "HEAD~1", "Old commit to compare against")
	cmd.Flags().StringVar(&newCommit, "new", "HEAD", "New commit to compare")
	return cmd
}
