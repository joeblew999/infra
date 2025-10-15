package cli

import (
	"github.com/spf13/cobra"

	sharedcli "github.com/joeblew999/infra/core/pkg/shared/cli"
)

// NewCommand constructs the tooling CLI entry point.
func NewCommand() *cobra.Command {
	root := sharedcli.NewRootCommand(sharedcli.BuilderOptions{
		Use:   "core-tool",
		Short: "Core release tooling",
		Long:  "Profile-driven tooling for local smoke tests and Fly deployments.",
	})
	root.SilenceErrors = true
	root.SilenceUsage = false

	var profile string
	root.PersistentFlags().StringVar(&profile, "profile", "", "Tooling profile to use (defaults to shared configuration)")

	root.AddCommand(newAuthCommand(&profile))
	root.AddCommand(newWorkflowCommand(&profile))

	return root
}
