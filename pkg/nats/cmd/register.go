package cmd

import "github.com/spf13/cobra"

// RegisterCLI mounts the NATS CLI wrappers (nats, nsc) under the provided parent.
func RegisterCLI(parent *cobra.Command) {
	parent.AddCommand(NewCLICmd())
	parent.AddCommand(NewNSCCmd())
}

// RegisterWorkflows mounts cluster management commands under the provided parent.
func RegisterWorkflows(parent *cobra.Command) {
	parent.AddCommand(NewClusterCmd())
}
