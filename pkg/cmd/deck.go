package cmd

import (
	deckcmd "github.com/joeblew999/infra/pkg/deck/cmd"
	"github.com/spf13/cobra"
)

// RunDeck registers deck commands
func RunDeck() {
	rootCmd.AddCommand(NewDeckCmd())
}

// AddDeckToCLI adds deck commands to the CLI namespace
func AddDeckToCLI(cliParent *cobra.Command) {
	cliParent.AddCommand(NewDeckCmd())
}

func NewDeckCmd() *cobra.Command {
	// Return the restored binary pipeline command
	return deckcmd.GetDeckCmd()
}

