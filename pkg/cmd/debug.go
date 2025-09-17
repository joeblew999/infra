package cmd

import (
	debugcmd "github.com/joeblew999/infra/pkg/debug/cmd"
	"github.com/spf13/cobra"
)

// AddDebugToCLI adds debug commands to the CLI namespace
func AddDebugToCLI(cliParent *cobra.Command) {
	cliParent.AddCommand(debugcmd.GetDebugCmd())
}