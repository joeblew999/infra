package main

// This file provides the main entry point for the infra application.
// It is intentionally minimal to ensure that any Go project can easily
// import and use the infra packages without conflicts or complexity.
// 
// Usage modes:
//   - Library usage: import "github.com/joeblew999/infra/pkg/[package]"
//     Examples: pkg/dep, pkg/workflows, pkg/cmd (for CLI library)
//   - CLI usage: go run . --mode=cli [command] [args]
//     Examples: go run . --mode=cli dep install
//   - Service mode: go run . [runs service by default]
// 
// The pkg/cmd package provides the unified CLI interface and can be
// imported by other projects as a library for consistent CLI behavior.

import (
	"github.com/joeblew999/infra/pkg/cmd"
)

func main() {
	// Logging will be initialized by cmd.Execute() based on mode
	cmd.Execute()
}
