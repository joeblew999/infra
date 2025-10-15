package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"time"

	wf "github.com/joeblew999/infra/pkg/datastarui/internal/workflow"
)

const version = "v0.1.0"

func main() {
	wd, err := os.Getwd()
	if err != nil {
		log.Fatalf("locate working directory: %v", err)
	}

	defaultSrc := filepath.Join(wd, "pkg/datastarui/sampleapp")

	cfg := wf.DefaultConfig()
	opts := wf.RegisterFlags(flag.CommandLine, cfg, 5*time.Minute)
	flag.Parse()

	if opts.ShowVersion {
		fmt.Println("datastarui playwright", version)
		return
	}

	srcDir := opts.Src
	if strings.TrimSpace(srcDir) == "" {
		srcDir = defaultSrc
	}
	srcDir, err = filepath.Abs(srcDir)
	if err != nil {
		log.Fatalf("resolve src path: %v", err)
	}

	if _, err := os.Stat(srcDir); err != nil {
		log.Fatalf("src path invalid: %v", err)
	}

	opts.Apply(&cfg)

	switch cfg.Workflow {
	case wf.WorkflowBun, "":
		if _, err := exec.LookPath("bun"); err != nil {
			log.Fatalf("bun runtime not found: %v", err)
		}
	case wf.WorkflowNode:
		if _, err := exec.LookPath("pnpm"); err != nil {
			log.Fatalf("pnpm binary not found: %v", err)
		}
	default:
		log.Fatalf("unsupported workflow: %s", cfg.Workflow)
	}

	baseCtx, cancel := context.WithTimeout(context.Background(), opts.Timeout)
	defer cancel()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt)
	go func() {
		select {
		case <-sigs:
			cancel()
		case <-baseCtx.Done():
		}
	}()

	opts.Report("datastarui playwright", version, srcDir, cfg)

	if err := wf.Run(baseCtx, srcDir, cfg); err != nil {
		log.Fatalf("playwright run failed: %v", err)
	}

	log.Println("Playwright suite completed successfully")
}
