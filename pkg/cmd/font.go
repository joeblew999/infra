package cmd

import (
	fontcmd "github.com/joeblew999/infra/pkg/font/cmd"
	"github.com/spf13/cobra"
)

// AddFontToCLI adds font commands to the CLI namespace
func AddFontToCLI(cliParent *cobra.Command) {
	cliParent.AddCommand(NewFontCmd())
}

func NewFontCmd() *cobra.Command {
	return fontcmd.GetFontCmd()
}