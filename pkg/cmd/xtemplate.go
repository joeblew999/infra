package cmd

import (
	xtemplatecmd "github.com/joeblew999/infra/pkg/xtemplate/cmd"
	"github.com/spf13/cobra"
)

// AddXTemplateToCLI adds xtemplate commands to the CLI namespace
func AddXTemplateToCLI(cliParent *cobra.Command) {
	cliParent.AddCommand(xtemplatecmd.GetXTemplateCmd())
}