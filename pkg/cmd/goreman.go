package cmd

import (
	"github.com/spf13/cobra"
	goremanCmds "github.com/joeblew999/infra/pkg/goreman/cmd"
)

var goremanCmd = &cobra.Command{
	Use:   "goreman",
	Short: "Process management and monitoring",
	Long:  `Monitor and manage supervised processes via goreman`,
}

func init() {
	// Add subcommands from goreman package
	goremanCmd.AddCommand(goremanCmds.PsCmd)
	goremanCmd.AddCommand(goremanCmds.StartCmd)
	goremanCmd.AddCommand(goremanCmds.StopCmd)
	goremanCmd.AddCommand(goremanCmds.RestartCmd)
	goremanCmd.AddCommand(goremanCmds.RegisterCmd)
	goremanCmd.AddCommand(goremanCmds.ServicesCmd)
	
	// Add to root command
	// rootCmd.AddCommand(goremanCmd)  // Disabled - use web interface at :1337/goreman instead
}