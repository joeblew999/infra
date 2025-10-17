package playwright

import (
	"context"
	"time"
)

// WorkflowMode describes how Playwright tests should be executed.
type WorkflowMode string

const (
	// WorkflowBun executes Playwright via Bun runtime
	WorkflowBun WorkflowMode = "bun"
	// WorkflowNode executes Playwright via pnpm/Node.js
	WorkflowNode WorkflowMode = "node"
)

// Config captures the complete configuration for running Playwright tests.
type Config struct {
	// SourceDir is the directory containing tests/ and playwright.config.ts
	SourceDir string

	// BaseURL is the URL where the server is running (passed to Playwright)
	BaseURL string

	// Workflow determines how to run Playwright (bun or node/pnpm)
	Workflow WorkflowMode

	// Headed controls whether to show the browser (false = headless)
	Headed bool

	// Timeout is the overall timeout for the entire test run
	Timeout time.Duration

	// Prepare is an optional hook for project-specific preparation
	// (e.g., running templ generate, Tailwind builds, etc.)
	// Called before starting the server.
	Prepare func(ctx context.Context, sourceDir string) error

	// ServerConfig controls how the test server is started
	ServerConfig ServerConfig
}

// ServerConfig controls how the test server is started and managed.
type ServerConfig struct {
	// Command is a custom command to start the server (e.g., ["go", "run", "."])
	// If empty, uses Binary instead
	Command []string

	// Binary is the path to a compiled binary to run as the server
	// Ignored if Command is set
	Binary string

	// SkipServer disables server startup (for testing external sites)
	SkipServer bool

	// StartTimeout is how long to wait for the server to be ready
	StartTimeout time.Duration
}

// DefaultConfig returns sensible defaults for Playwright testing.
func DefaultConfig() Config {
	return Config{
		BaseURL:      "http://localhost:4242",
		Workflow:     WorkflowBun,
		Headed:       false,
		Timeout:      5 * time.Minute,
		ServerConfig: DefaultServerConfig(),
	}
}

// DefaultServerConfig returns sensible defaults for server management.
func DefaultServerConfig() ServerConfig {
	return ServerConfig{
		SkipServer:   false,
		StartTimeout: 30 * time.Second,
	}
}
