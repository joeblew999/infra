package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
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
	opts := wf.RegisterFlags(flag.CommandLine, cfg, 2*time.Minute)
	flag.Parse()

	if opts.ShowVersion {
		fmt.Println("datastarui codegen", version)
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

	ctx, cancel := context.WithTimeout(context.Background(), opts.Timeout)
	defer cancel()

	opts.Report("datastarui codegen", version, srcDir, cfg)

	if err := wf.Prepare(ctx, srcDir, cfg); err != nil {
		log.Fatalf("code generation failed: %v", err)
	}

	log.Println("Code generation completed successfully")
}
