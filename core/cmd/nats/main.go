package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	coreNATS "github.com/joeblew999/infra/core/services/nats"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if err := coreNATS.Run(ctx, os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "nats: %v\n", err)
		os.Exit(1)
	}
}
