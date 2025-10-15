package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	coresvc "github.com/joeblew999/infra/core/services/caddy"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if err := coresvc.Run(ctx, os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "caddy: %v\n", err)
		os.Exit(1)
	}
}
