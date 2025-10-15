package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	toolcli "github.com/joeblew999/infra/core/tooling/internal/cli"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	cmd := toolcli.NewCommand()
	cmd.SetContext(ctx)
	cmd.SetArgs(os.Args[1:])

	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "core-tool: %v\n", err)
		os.Exit(1)
	}
}
