package cli

import (
	"context"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	runtimeconfig "github.com/joeblew999/infra/core/pkg/runtime/config"
	runtimecontroller "github.com/joeblew999/infra/core/pkg/runtime/controller"
	runtimeprocess "github.com/joeblew999/infra/core/pkg/runtime/process"
	runtimeui "github.com/joeblew999/infra/core/pkg/runtime/ui"
	"github.com/joeblew999/infra/core/pkg/runtime/ui/live"
	runtimeuitui "github.com/joeblew999/infra/core/pkg/runtime/ui/tui"
	runtimeuiweb "github.com/joeblew999/infra/core/pkg/runtime/ui/web"
	sharedbuild "github.com/joeblew999/infra/core/pkg/shared/build"
	sharedcli "github.com/joeblew999/infra/core/pkg/shared/cli"
)

// Execute constructs the core CLI and runs it using the provided context.
func Execute(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}
	root := newRootCommand()
	root.SetContext(ctx)
	root.SetOut(os.Stdout)
	root.SetErr(os.Stderr)
	return root.ExecuteContext(ctx)
}

func newRootCommand() *cobra.Command {
	cmd := sharedcli.NewRootCommand(sharedcli.BuilderOptions{
		Use:   "core",
		Short: "Deterministic core orchestrator",
		Long: strings.TrimSpace(`
Core manages the deterministic runtime stack described in Task 015.

It provides a single entry point for mirrored shared modules, manifest-driven
services, and the static UI shells while the process runner and event pipeline
are still being wired in.
`),
	})
	cmd.SilenceUsage = true
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	}
	cmd.Example = strings.TrimSpace(`
  # List the services registered with the deterministic controller
  core services

  # Render the Ultraviolet-based terminal snapshot
  core tui

  # Launch the Datastar-powered web dashboard on a custom port
  core web --addr 127.0.0.1:3435

  # Display build metadata for the current binary
  core version
`)

	sharedcli.AddCommand(cmd,
		newUpCommand(),
		newDownCommand(),
		newStatusCommand(),
		newCaddyCommand(),
		newNATSCommand(),
		newPocketBaseCommand(),
		newStackCommand(),
		newSecretsCommand(),
		newServicesCommand(),
		newTUICommand(),
		newWebCommand(),
		newVersionCommand(),
		newDeployCommand(),
		newEnsureCommand(),
	)
	return cmd
}

func newServicesCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "services",
		Short: "List services registered with the core controller",
		RunE: func(cmd *cobra.Command, args []string) error {
			registry, err := runtimecontroller.LoadBuiltIn()
			if err != nil {
				return err
			}
			services := registry.List()
			if len(services) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "no services registered")
				return nil
			}
			for _, svc := range services {
				fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s\n", svc.ID, svc.Process.Command)
			}
			return nil
		},
	}
}

func newTUICommand() *cobra.Command {
	var (
		page        string
		liveMode    bool
		composePort int
		controller  string
	)
	cmd := &cobra.Command{
		Use:   "tui",
		Short: "Render the terminal UI snapshot",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !liveMode {
				return runtimeuitui.Run(cmd.Context(), cmd.OutOrStdout(), page, nil)
			}

			port := composePort
			if port <= 0 {
				port = runtimeprocess.ComposePort(nil)
			}
			snapshot := runtimeui.LoadTestSnapshot()
			if states, err := runtimeprocess.FetchComposeProcesses(cmd.Context(), port); err == nil {
				serviceStates := runtimeui.ServiceStatusesFromCompose(states)
				snapshot = runtimeui.BuildSnapshotFromServiceStatus(serviceStates)
			}
			store := live.NewStore(snapshot)
			store.StartComposeSync(cmd.Context(), port, 2*time.Second)

			// Start observability event stream
			cfg := runtimeconfig.Load()
			if err := store.StartEventStream(cmd.Context(), cfg.Services.NATS); err != nil {
				// Don't fail TUI startup if event stream fails, just warn
				fmt.Fprintf(cmd.ErrOrStderr(), "Warning: failed to start event stream: %v\n", err)
			}

			return runtimeuitui.Run(cmd.Context(), cmd.OutOrStdout(), page, store)
		},
	}
	cmd.Flags().StringVar(&page, "page", "overview", "page route to render (overview or service/<id>)")
	cmd.Flags().BoolVar(&liveMode, "live", false, "enable live data simulation")
	cmd.Flags().IntVar(&composePort, "compose-port", 0, "Process Compose port to query for live data (defaults to PC_PORT_NUM or 28081)")
	cmd.Flags().StringVar(&controller, "controller", os.Getenv("CONTROLLER_ADDR"), "controller API address for live events")
	return cmd
}

func newWebCommand() *cobra.Command {
	var (
		addr        string
		page        string
		liveMode    bool
		composePort int
		controller  string
	)
	cmd := &cobra.Command{
		Use:   "web",
		Short: "Serve the web UI",
		RunE: func(cmd *cobra.Command, args []string) error {
			listener, actualAddr, err := listenAddress(addr)
			if err != nil {
				return err
			}
			defer listener.Close()

			opts := runtimeuiweb.Options{Page: page}
			if liveMode {
				port := composePort
				if port <= 0 {
					port = runtimeprocess.ComposePort(nil)
				}
				snapshot := runtimeui.LoadTestSnapshot()
				if states, err := runtimeprocess.FetchComposeProcesses(cmd.Context(), port); err == nil {
					serviceStates := runtimeui.ServiceStatusesFromCompose(states)
					snapshot = runtimeui.BuildSnapshotFromServiceStatus(serviceStates)
				}
				store := live.NewStore(snapshot)
				store.StartComposeSync(cmd.Context(), port, 2*time.Second)

				// Start observability event stream
				cfg := runtimeconfig.Load()
				if err := store.StartEventStream(cmd.Context(), cfg.Services.NATS); err != nil {
					fmt.Fprintf(cmd.ErrOrStderr(), "Warning: failed to start event stream: %v\n", err)
				}

				opts.Store = store
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Starting web UI on http://%s (press Ctrl+C to stop)\n", actualAddr)
			if err := runtimeuiweb.Run(cmd.Context(), listener, cmd.OutOrStdout(), opts); err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), "web UI stopped")
			return nil
		},
	}
	cmd.Flags().StringVar(&addr, "addr", "127.0.0.1:3400", "address to bind the web UI")
	cmd.Flags().StringVar(&page, "page", "overview", "default page route to render")
	cmd.Flags().BoolVar(&liveMode, "live", false, "enable live data simulation")
	cmd.Flags().IntVar(&composePort, "compose-port", 0, "Process Compose port to query for live data (defaults to PC_PORT_NUM or 28081)")
	cmd.Flags().StringVar(&controller, "controller", os.Getenv("CONTROLLER_ADDR"), "controller API address for live events")
	return cmd
}

func listenAddress(addr string) (net.Listener, string, error) {
	listener, err := net.Listen("tcp", addr)
	if err == nil {
		return listener, listener.Addr().String(), nil
	}
	fallback, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, "", err
	}
	return fallback, fallback.Addr().String(), nil
}

func newVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show build information",
		RunE: func(cmd *cobra.Command, args []string) error {
			info := sharedbuild.Get()
			if !info.Available {
				fmt.Fprintln(cmd.OutOrStdout(), "core version: unknown (no build info)")
				return nil
			}

			fmt.Fprintf(cmd.OutOrStdout(), "core version: %s\n", info.Version)
			if info.GoVersion != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "go version: %s\n", info.GoVersion)
			}
			if info.Revision != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "git commit: %s\n", info.Revision)
			}
			if info.BuildTime != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "build time: %s\n", info.BuildTime)
			}
			if info.ModifiedSet {
				fmt.Fprintf(cmd.OutOrStdout(), "dirty tree: %t\n", info.Modified)
			}
			return nil
		},
	}
}
