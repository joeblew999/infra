package cmd

import (
	utmcmd "github.com/joeblew999/infra/pkg/utm/cmd"
	"github.com/spf13/cobra"
)

// AddUTMToCLI adds UTM commands to the CLI parent command
func AddUTMToCLI(parentCmd *cobra.Command) {
	parentCmd.AddCommand(utmcmd.GetUTMCmd())
}