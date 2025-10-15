package cli

import (
	"context"
	"fmt"
	"sort"

	"github.com/spf13/cobra"

	caddyservice "github.com/joeblew999/infra/core/services/caddy"
)

func newCaddyCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "caddy",
		Short: "Manage the embedded Caddy service",
	}

	cmd.AddCommand(newCaddyRunCommand())
	cmd.AddCommand(newCaddyEnsureCommand())
	cmd.AddCommand(newCaddySpecCommand())
	return cmd
}

func newCaddyRunCommand() *cobra.Command {
	run := &cobra.Command{
		Use:   "run",
		Short: "Run embedded Caddy with the generated config",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			if ctx == nil {
				ctx = context.Background()
			}
			return caddyservice.Run(ctx, args)
		},
	}
	run.DisableFlagsInUseLine = true
	return run
}

func newCaddyEnsureCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "ensure",
		Short: "Ensure the embedded Caddy binary is installed",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := caddyservice.LoadConfig()
			if err != nil {
				return err
			}
			paths, err := cfg.EnsureBinaries()
			if err != nil {
				return err
			}
			if len(paths) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No binaries declared in the manifest")
				return nil
			}
			keys := make([]string, 0, len(paths))
			for name := range paths {
				keys = append(keys, name)
			}
			sort.Strings(keys)
			fmt.Fprintln(cmd.OutOrStdout(), "Binaries installed:")
			for _, name := range keys {
				fmt.Fprintf(cmd.OutOrStdout(), "  %s -> %s\n", name, paths[name])
			}
			return nil
		},
	}
}

func newCaddySpecCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "spec",
		Short: "Show the embedded Caddy config manifest",
		RunE: func(cmd *cobra.Command, args []string) error {
			spec, err := caddyservice.LoadConfig()
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "HTTP port: %d (%s)\n", spec.Ports.HTTP.Port, spec.Ports.HTTP.Protocol)
			fmt.Fprintf(cmd.OutOrStdout(), "Target: %s\n", spec.Config.Target)
			if len(spec.Process.Env) > 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "Environment overrides:")
				keys := make([]string, 0, len(spec.Process.Env))
				for k := range spec.Process.Env {
					keys = append(keys, k)
				}
				sort.Strings(keys)
				for _, k := range keys {
					fmt.Fprintf(cmd.OutOrStdout(), "  %s=%s\n", k, spec.Process.Env[k])
				}
			}
			return nil
		},
	}
}
