package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// BuilderOptions configures the shared root command builder.
type BuilderOptions struct {
	Use     string
	Short   string
	Long    string
	Version string
}

// NewRootCommand constructs a Cobra command with consistent defaults used by
// runtime and services.
func NewRootCommand(opts BuilderOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:           defaultString(opts.Use, "core"),
		Short:         defaultString(opts.Short, "Deterministic core platform"),
		Long:          opts.Long,
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	if opts.Version != "" {
		cmd.Version = opts.Version
		cmd.SetVersionTemplate("{{.Version}}\n")
	}
	cmd.SetHelpTemplate(helpTemplate)
	return cmd
}

const helpTemplate = `{{with (or .Long .Short)}}{{.}}

{{end}}Usage:
  {{.UseLine}}
{{if .HasAvailableSubCommands}}
Available Commands:
{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}
{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

Global Flags:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}`

func defaultString(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

// AddCommand ensures deterministic ordering when attaching sub-commands.
func AddCommand(parent *cobra.Command, children ...*cobra.Command) {
	for _, child := range children {
		parent.AddCommand(child)
	}
	parent.Commands() // trigger sorting inside cobra
}

// MarkRequired marks a command flag as required and panics on configuration
// mistakes during wiring so failures surface early in tests.
func MarkRequired(cmd *cobra.Command, name string) {
	if err := cmd.MarkFlagRequired(name); err != nil {
		panic(fmt.Sprintf("mark required flag %s: %v", name, err))
	}
}
