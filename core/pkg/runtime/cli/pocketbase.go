package cli

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"sort"

	pocketbasesvc "github.com/joeblew999/infra/core/services/pocketbase"
)

func newPocketBaseCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pocketbase",
		Short: "Manage the embedded PocketBase service",
		Long:  "Inspect and run the PocketBase instance declared under core/services/pocketbase.",
	}

	cmd.AddCommand(newPocketBaseRunCommand())
	cmd.AddCommand(newPocketBaseEnsureCommand())
	cmd.AddCommand(newPocketBaseSpecCommand())
	return cmd
}

func newPocketBaseRunCommand() *cobra.Command {
	run := &cobra.Command{
		Use:   "run",
		Short: "Run PocketBase using the embedded runner",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			if ctx == nil {
				ctx = context.Background()
			}
			return pocketbasesvc.Run(ctx, args)
		},
	}
	run.DisableFlagsInUseLine = true
	return run
}

func newPocketBaseEnsureCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "ensure",
		Short: "Ensure PocketBase tooling binaries are installed",
		RunE: func(cmd *cobra.Command, args []string) error {
			spec, err := pocketbasesvc.LoadSpec()
			if err != nil {
				return err
			}
			paths, err := spec.EnsureBinaries()
			if err != nil {
				return err
			}

			keys := make([]string, 0, len(paths))
			for k := range paths {
				keys = append(keys, k)
			}
			sort.Strings(keys)

			fmt.Fprintln(cmd.OutOrStdout(), "Binaries installed:")
			for _, k := range keys {
				fmt.Fprintf(cmd.OutOrStdout(), "  %s -> %s\n", k, paths[k])
			}
			return nil
		},
	}
}

func newPocketBaseSpecCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "spec",
		Short: "Show the embedded PocketBase service spec",
		RunE: func(cmd *cobra.Command, args []string) error {
			spec, err := pocketbasesvc.LoadSpec()
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), "Process: embedded")
			fmt.Fprintf(cmd.OutOrStdout(), "Primary port: %d (%s)\n", spec.Ports.Primary.Port, spec.Ports.Primary.Protocol)
			if len(spec.Process.Env) > 0 {
				keys := make([]string, 0, len(spec.Process.Env))
				for k := range spec.Process.Env {
					keys = append(keys, k)
				}
				sort.Strings(keys)
				fmt.Fprintln(cmd.OutOrStdout(), "Environment overrides:")
				for _, k := range keys {
					fmt.Fprintf(cmd.OutOrStdout(), "  %s=%s\n", k, spec.Process.Env[k])
				}
			}
			return nil
		},
	}
}
