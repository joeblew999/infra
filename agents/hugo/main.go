package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

func main() {
	cfg := parseFlags()

	if err := cfg.run(); err != nil {
		fmt.Fprintf(os.Stderr, "agents hugo runner: %v\n", err)
		os.Exit(1)
	}
}

type options struct {
	mode    string
	port    int
	drafts  bool
	watch   bool
	dryRun  bool
	hugoBin string
}

func parseFlags() options {
	var cfg options
	flag.StringVar(&cfg.mode, "mode", "serve", "Operation to run: serve or build")
	flag.IntVar(&cfg.port, "port", 1414, "Port for hugo server (serve mode only)")
	flag.BoolVar(&cfg.drafts, "drafts", true, "Include draft content")
	flag.BoolVar(&cfg.watch, "watch", true, "Enable file watching in serve mode")
	flag.BoolVar(&cfg.dryRun, "dry-run", false, "Print commands without executing them")
	flag.StringVar(&cfg.hugoBin, "hugo", "", "Path to the hugo binary (defaults to PATH lookup)")
	flag.Parse()

	cfg.mode = strings.ToLower(cfg.mode)
	return cfg
}

func (o options) run() error {
	if o.mode != "serve" && o.mode != "build" {
		return fmt.Errorf("unknown mode %q", o.mode)
	}

	hugoPath := o.hugoBin
	if hugoPath == "" {
		path, err := exec.LookPath("hugo")
		if err != nil {
			return errors.New("hugo binary not found; install with `brew install hugo` or provide --hugo path")
		}
		hugoPath = path
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("resolve working directory: %w", err)
	}

	agentsDir := cwd
	// ensure we are inside agents directory (no parent operations)
	if _, err := os.Stat(filepath.Join(agentsDir, "content")); err != nil {
		return fmt.Errorf("expected content directory inside %s: %w", agentsDir, err)
	}

	args := []string{"--source", agentsDir}

	if o.mode == "serve" {
		args = append([]string{"server"}, args...)
		args = append(args, "--bind", "127.0.0.1", "--port", fmt.Sprint(o.port))
		if !o.watch {
			args = append(args, "--watch", "false")
		}
		args = append(args, "--disableFastRender", "--renderToMemory")
	} else {
		args = append(args, "--minify", "--destination", filepath.Join(agentsDir, "public"))
	}

	if o.drafts {
		args = append(args, "--buildDrafts", "--buildFuture")
	}

	if o.dryRun {
		fmt.Printf("[dry-run] %s %s\n", hugoPath, strings.Join(args, " "))
		return nil
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cmd := exec.CommandContext(ctx, hugoPath, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start hugo: %w", err)
	}

	err = cmd.Wait()
	if ctx.Err() == context.Canceled {
		return nil
	}
	if err != nil {
		return fmt.Errorf("hugo finished with error: %w", err)
	}

	if o.mode == "build" {
		fmt.Printf("Build complete at %s\n", filepath.Join(agentsDir, "public"))
	} else {
		fmt.Printf("Hugo server running on http://127.0.0.1:%d (Ctrl+C to stop)\n", o.port)
		// give server a moment to stay alive if command returns unexpectedly
		time.Sleep(100 * time.Millisecond)
	}

	return nil
}
