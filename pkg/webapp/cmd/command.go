package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/joeblew999/infra/pkg/config"
	"github.com/joeblew999/infra/pkg/webapp"
)

// Register mounts the web app commands under the provided parent.
func Register(parent *cobra.Command) {
	parent.AddCommand(NewCommand())
}

// NewCommand builds the root web command hierarchy.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "web",
		Short: "Run the web management UI",
		Long:  "Control the embedded infrastructure web UI with live metrics, docs, and orchestration helpers.",
	}

	cmd.AddCommand(newServeCommand())
	return cmd
}

func newServeCommand() *cobra.Command {
	serveCmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the web UI server",
		RunE:  runServe,
	}

	serveCmd.Flags().String("port", config.GetWebServerPort(), "Port to listen on")
	serveCmd.Flags().String("nats", config.GetNATSURL(), "NATS URL for realtime features")
	serveCmd.Flags().Bool("docs-dev", true, "Serve docs from local filesystem instead of embedded copy")

	return serveCmd
}

func runServe(cmd *cobra.Command, args []string) error {
	port, _ := cmd.Flags().GetString("port")
	natsURL, _ := cmd.Flags().GetString("nats")
	docsDev, _ := cmd.Flags().GetBool("docs-dev")

	svc := webapp.NewService(
		webapp.WithPort(port),
		webapp.WithNATSURL(natsURL),
		webapp.WithDocsDevMode(docsDev),
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	go func() {
		<-sigCh
		fmt.Println("\nStopping web server...")
		cancel()
	}()

	if err := svc.Start(ctx); err != nil {
		return fmt.Errorf("web server exited: %w", err)
	}

	return nil
}
