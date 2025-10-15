package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	corepocketbase "github.com/joeblew999/infra/core/services/pocketbase"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if err := corepocketbase.Run(ctx, os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "pocketbase: %v\n", err)
		os.Exit(1)
	}
}
