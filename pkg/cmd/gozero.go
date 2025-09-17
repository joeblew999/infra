package cmd

import (
	gozerocmd "github.com/joeblew999/infra/pkg/gozero/cmd"
	"github.com/spf13/cobra"
)

// AddGoZeroToCLI adds gozero commands to the CLI namespace
func AddGoZeroToCLI(cliParent *cobra.Command) {
	cliParent.AddCommand(gozerocmd.GetGoZeroCmd())
}