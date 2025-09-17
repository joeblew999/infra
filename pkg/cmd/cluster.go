package cmd

import (
	natscmd "github.com/joeblew999/infra/pkg/nats/cmd"
	"github.com/spf13/cobra"
)

// addClusterCommands adds the cluster management commands to the root command
func addClusterCommands(rootCmd *cobra.Command) {
	rootCmd.AddCommand(natscmd.GetClusterCmd())
}