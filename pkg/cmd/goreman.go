package cmd

import (
	goremanCmds "github.com/joeblew999/infra/pkg/goreman/cmd"
	"github.com/spf13/cobra"
)

// RunGoreman adds goreman commands to the root command for direct usage.
func RunGoreman() {
	rootCmd.AddCommand(goremanCmds.GetGoremanCmd())
}

// AddGoremanToCLI adds goreman commands to the CLI namespace.
func AddGoremanToCLI(cliParent *cobra.Command) {
	cliParent.AddCommand(goremanCmds.GetGoremanCmd())
}
