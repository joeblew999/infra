package main

import (
	"context"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	wf "github.com/joeblew999/infra/pkg/datastarui/internal/workflow"
)

func TestPlaywrightSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("playwright suite skipped in short mode")
	}

	if _, err := exec.LookPath("bun"); err != nil {
		t.Skip("bun runtime not found; skipping playwright suite")
	}

	srcDir, err := filepath.Abs(".")
	if err != nil {
		t.Fatalf("resolve src path: %v", err)
	}

	cfg := wf.DefaultConfig()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	if err := wf.Run(ctx, srcDir, cfg); err != nil {
		t.Fatalf("playwright suite failed: %v", err)
	}
}
