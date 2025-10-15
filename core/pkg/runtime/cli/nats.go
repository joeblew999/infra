package cli

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	runtimeconfig "github.com/joeblew999/infra/core/pkg/runtime/config"
	servicenats "github.com/joeblew999/infra/core/services/nats"
)

// newNATSCommand creates the NATS management command tree scoped entirely to
// the deterministic core runtime. The legacy infra packages are intentionally
// not imported here so the core tree stays self-contained.
func newNATSCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "nats",
		Short: "Manage the embedded NATS service",
		Long: `Inspect and launch the NATS instance declared under core/services/nats.

Commands in this group operate strictly on the embedded manifest – no legacy
pkg/ dependencies are referenced.`,
	}

	cmd.AddCommand(newNATSRunCommand())
	cmd.AddCommand(newNATSEnsureCommand())
	cmd.AddCommand(newNATSSpecCommand())
	cmd.AddCommand(newNATSCommandCommand())

	return cmd
}

func newNATSRunCommand() *cobra.Command {
	run := &cobra.Command{
		Use:   "run [-- extra args]",
		Short: "Run the embedded NATS server",
		Long: `Ensure the manifest binaries exist, then execute the Pillow-managed
NATS process in the foreground. Additional arguments are forwarded to Pillow
after the manifest-defined args.`,
		Args: cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			if ctx == nil {
				ctx = context.Background()
			}
			return servicenats.Run(ctx, args)
		},
		Example: strings.TrimSpace(`
  # Run NATS using the embedded Pillow manifest
  core nats run

  # Run NATS with additional tracing flags passed to Pillow
  core nats run -- --trace`,
		),
	}

	run.DisableFlagsInUseLine = true
	return run
}

func newNATSEnsureCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "ensure",
		Short: "Ensure manifest binaries are installed",
		RunE: func(cmd *cobra.Command, args []string) error {
			spec, err := servicenats.LoadSpec()
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

func newNATSSpecCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "spec",
		Short: "Show the embedded NATS service spec",
		RunE: func(cmd *cobra.Command, args []string) error {
			spec, err := servicenats.LoadSpec()
			if err != nil {
				return err
			}
			cfg := runtimeconfig.Load()

			fmt.Fprintf(cmd.OutOrStdout(), "Environment: %s\n", cfg.Environment)
			fmt.Fprintln(cmd.OutOrStdout(), "\nProcess:")
			fmt.Fprintf(cmd.OutOrStdout(), "  Command: %s\n", spec.Process.Command)
			if len(spec.Process.Args) > 0 {
				for _, arg := range spec.Process.Args {
					fmt.Fprintf(cmd.OutOrStdout(), "  Arg: %s\n", arg)
				}
			}

			if len(spec.Process.Env) > 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "\nEnvironment:")
				keys := make([]string, 0, len(spec.Process.Env))
				for k := range spec.Process.Env {
					keys = append(keys, k)
				}
				sort.Strings(keys)
				for _, k := range keys {
					fmt.Fprintf(cmd.OutOrStdout(), "  %s=%s\n", k, spec.Process.Env[k])
				}
			}

			fmt.Fprintln(cmd.OutOrStdout(), "\nPorts:")
			w := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
			fmt.Fprintln(w, "NAME\tPORT\tPROTOCOL")
			fmt.Fprintf(w, "client\t%d\t%s\n", spec.Ports.Client.Port, spec.Ports.Client.Protocol)
			fmt.Fprintf(w, "cluster\t%d\t%s\n", spec.Ports.Cluster.Port, spec.Ports.Cluster.Protocol)
			fmt.Fprintf(w, "http\t%d\t%s\n", spec.Ports.HTTP.Port, spec.Ports.HTTP.Protocol)
			fmt.Fprintf(w, "leaf\t%d\t%s\n", spec.Ports.Leaf.Port, spec.Ports.Leaf.Protocol)
			_ = w.Flush()

			fmt.Fprintln(cmd.OutOrStdout(), "\nDeployment:")
			fmt.Fprintf(cmd.OutOrStdout(), "  Local Nodes: %d (%s mode)\n", spec.Config.Deployment.Local.Nodes, spec.Config.Deployment.Local.Mode)
			fmt.Fprintf(cmd.OutOrStdout(), "  Production Hub: %s (min %d nodes)\n", spec.Config.Deployment.Production.HubRegion, spec.Config.Deployment.Production.MinHubNodes)
			fmt.Fprintf(cmd.OutOrStdout(), "  Production Leaf Regions: %v\n", spec.Config.Deployment.Production.LeafRegions)
			fmt.Fprintf(cmd.OutOrStdout(), "  Leaf Nodes/Region: %d\n", spec.Config.Deployment.Production.LeafNodesPerRegion)

			fmt.Fprintln(cmd.OutOrStdout())
			fmt.Fprintf(cmd.OutOrStdout(), "JetStream: %v\n", spec.Config.JetStream)
			fmt.Fprintf(cmd.OutOrStdout(), "Backend: %s (auto-scale: %v)\n", spec.Config.Backend, spec.Config.AutoScale)
			fmt.Fprintf(cmd.OutOrStdout(), "Topology: %s\n", spec.Config.Topology)
			return nil
		},
	}
}

func newNATSCommandCommand() *cobra.Command {
	var includeEnv bool
	commandCmd := &cobra.Command{
		Use:   "command",
		Short: "Print the resolved launch command",
		RunE: func(cmd *cobra.Command, args []string) error {
			spec, err := servicenats.LoadSpec()
			if err != nil {
				return err
			}
			paths, err := spec.EnsureBinaries()
			if err != nil {
				return err
			}

			resolvedCmd := spec.ResolveCommand(paths)
			resolvedArgs := spec.ResolveArgs(paths)
			fmt.Fprintf(cmd.OutOrStdout(), "%s %s\n", resolvedCmd, strings.Join(resolvedArgs, " "))

			if includeEnv {
				fmt.Fprintln(cmd.OutOrStdout(), "\nEnvironment:")
				env := spec.ResolveEnv(paths)
				keys := make([]string, 0, len(env))
				for k := range env {
					keys = append(keys, k)
				}
				sort.Strings(keys)
				for _, k := range keys {
					fmt.Fprintf(cmd.OutOrStdout(), "  %s=%s\n", k, env[k])
				}
			}
			return nil
		},
	}
	commandCmd.Flags().BoolVar(&includeEnv, "env", false, "also print resolved environment variables")
	return commandCmd
}
