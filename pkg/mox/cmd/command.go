package cmd

import (
	"github.com/joeblew999/infra/pkg/mox"
	"github.com/spf13/cobra"
)

// Register mounts the mox CLI commands under the provided parent.
func Register(parent *cobra.Command) {
	parent.AddCommand(NewMoxCmd())
}

// NewMoxCmd returns the root mox command with subcommands attached.
func NewMoxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mox",
		Short: "Manage the mox mail server",
		Long:  "Commands for managing the mox mail server lifecycle and initialization",
	}

	cmd.AddCommand(newStartCmd())
	cmd.AddCommand(newInitCmd())
	return cmd
}

func newStartCmd() *cobra.Command {
	start := &cobra.Command{
		Use:   "start",
		Short: "Start the mox mail server",
		RunE: func(cmd *cobra.Command, args []string) error {
			domain, _ := cmd.Flags().GetString("domain")
			adminEmail, _ := cmd.Flags().GetString("admin-email")
			return mox.StartSupervised(domain, adminEmail)
		},
	}

	start.Flags().String("domain", "localhost", "Domain for the mail server")
	start.Flags().String("admin-email", "admin@localhost", "Admin email for the mail server")
	return start
}

func newInitCmd() *cobra.Command {
	initCmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize the mox mail server",
		RunE: func(cmd *cobra.Command, args []string) error {
			domain, _ := cmd.Flags().GetString("domain")
			adminEmail, _ := cmd.Flags().GetString("admin-email")
			server := mox.NewServer(domain, adminEmail)
			return server.Init()
		},
	}

	initCmd.Flags().String("domain", "localhost", "Domain for the mail server")
	initCmd.Flags().String("admin-email", "admin@localhost", "Admin email for the mail server")
	return initCmd
}
