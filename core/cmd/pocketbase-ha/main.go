package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	pocketbaseha "github.com/joeblew999/infra/core/services/pocketbase-ha"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if err := pocketbaseha.Run(ctx, os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "pocketbase-ha: %v\n", err)
		os.Exit(1)
	}
}
