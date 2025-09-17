package cmd

import (
	tokicmd "github.com/joeblew999/infra/pkg/toki/cmd"
	"github.com/spf13/cobra"
)

// AddTokiToCLI adds toki commands to the CLI namespace
func AddTokiToCLI(cliParent *cobra.Command) {
	cliParent.AddCommand(tokicmd.GetTokiCmd())
}