package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	runtimecli "github.com/joeblew999/infra/core/pkg/runtime/cli"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if err := runtimecli.Execute(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "core: %v\n", err)
		os.Exit(1)
	}
}
