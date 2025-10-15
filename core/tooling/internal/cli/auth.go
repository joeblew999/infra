package cli

import "github.com/spf13/cobra"

func newAuthCommand(profileFlag *string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Manage cached provider credentials",
	}
	cmd.AddCommand(newAuthFlyCommand(profileFlag))
	cmd.AddCommand(newAuthCloudflareCommand(profileFlag))
	cmd.AddCommand(newAuthStatusCommand(profileFlag))
	return cmd
}
