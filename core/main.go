package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	runtimecli "github.com/joeblew999/infra/core/pkg/runtime/cli"
)

func main() {
	// Auto-load .env if it exists (fail silently if not - allows production without .env)
	_ = godotenv.Load()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if err := runtimecli.Execute(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "infra: %v\n", err)
		os.Exit(1)
	}
}
