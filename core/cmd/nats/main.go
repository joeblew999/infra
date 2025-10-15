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
	// Set up signal handling - only SIGINT and SIGTERM, not SIGHUP
	// This allows process-compose to manage the process lifecycle properly
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		fmt.Fprintf(os.Stderr, "[nats/main] Received signal: %v\n", sig)
		cancel()
	}()

	if err := coreNATS.Run(ctx, os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "nats: %v\n", err)
		os.Exit(1)
	}
}
