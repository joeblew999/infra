# Playwright Testing Framework

Go package for orchestrating Playwright browser tests with automatic server lifecycle management.

## Features

- **Server Lifecycle**: Automatically starts/stops test servers
- **Multiple Workflows**: Support for Bun and Node/pnpm
- **Headless or Headed**: Run tests in headless mode or with visible browser
- **Flexible Configuration**: Custom prepare hooks, server commands, timeouts
- **HTTP Readiness**: Waits for server to be ready before running tests
- **Graceful Shutdown**: SIGINT then SIGKILL for clean server stops

## Quick Start

### Directory Structure

```
your-project/
  tests/
    example.spec.ts          # Playwright TypeScript tests
  playwright.config.ts       # Playwright configuration
  main.go                    # Server to test (optional)
```

### Basic Usage

```go
package main

import (
    "context"
    "time"

    playwright "github.com/joeblew999/infra/core/pkg/testing/playwright"
)

func main() {
    cfg := playwright.Config{
        SourceDir: "./",
        BaseURL:   "http://localhost:8080",
        Workflow:  playwright.WorkflowBun,
        Headed:    false,
        Timeout:   5 * time.Minute,

        ServerConfig: playwright.ServerConfig{
            Command: []string{"go", "run", "."},
        },
    }

    ctx := context.Background()
    if err := playwright.Run(ctx, cfg); err != nil {
        log.Fatal(err)
    }
}
```

### Go Test Integration

```go
package main

import (
    "context"
    "testing"
    "time"

    playwright "github.com/joeblew999/infra/core/pkg/testing/playwright"
)

func TestPlaywright(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping playwright tests in short mode")
    }

    cfg := playwright.DefaultConfig()
    cfg.SourceDir = "."

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
    defer cancel()

    if err := playwright.Run(ctx, cfg); err != nil {
        t.Fatalf("playwright tests failed: %v", err)
    }
}
```

## Configuration

### Config Options

```go
type Config struct {
    // SourceDir with tests/ and playwright.config.ts
    SourceDir string

    // BaseURL where server is running
    BaseURL string

    // Workflow: WorkflowBun or WorkflowNode
    Workflow WorkflowMode

    // Headed: true shows browser, false is headless
    Headed bool

    // Timeout for entire test run
    Timeout time.Duration

    // Prepare: optional pre-test hook
    Prepare func(ctx context.Context, dir string) error

    // ServerConfig controls server startup
    ServerConfig ServerConfig
}
```

### Server Configuration

```go
type ServerConfig struct {
    // Command to start server (e.g., ["go", "run", "."])
    Command []string

    // Binary path instead of command
    Binary string

    // SkipServer for testing external sites
    SkipServer bool

    // StartTimeout to wait for readiness
    StartTimeout time.Duration
}
```

## Advanced Usage

### Custom Prepare Hook

```go
cfg := playwright.Config{
    SourceDir: "./",

    // Run custom setup before tests
    Prepare: func(ctx context.Context, dir string) error {
        // Generate templates, build CSS, etc.
        cmd := exec.CommandContext(ctx, "make", "build")
        cmd.Dir = dir
        return cmd.Run()
    },
}
```

### Testing External Sites

```go
cfg := playwright.Config{
    SourceDir: "./tests",
    BaseURL:   "https://example.com",

    ServerConfig: playwright.ServerConfig{
        SkipServer: true,  // Don't start a server
    },
}
```

### Using a Pre-built Binary

```go
cfg := playwright.Config{
    SourceDir: "./",
    BaseURL:   "http://localhost:8080",

    ServerConfig: playwright.ServerConfig{
        Binary: "./bin/myapp",  // Use compiled binary
    },
}
```

### Headed Mode (Show Browser)

```go
cfg := playwright.Config{
    SourceDir: "./",
    Headed:    true,  // Open browser window
}
```

## Workflow Modes

### Bun (Default)

```go
cfg.Workflow = playwright.WorkflowBun
```

Requires: `bun` runtime installed
- Install: `brew install oven-sh/bun/bun`
- Runs: `bun x playwright test`

### Node/pnpm

```go
cfg.Workflow = playwright.WorkflowNode
```

Requires: `pnpm` installed
- Install: `npm install -g pnpm`
- Runs: `pnpm exec playwright test`

## Example Projects

### DatastarUI Tests

```go
// pkg/datastarui/cmd/playwright/main.go
cfg := playwright.Config{
    SourceDir: "./sampleapp",
    BaseURL:   "http://localhost:4242",
    Workflow:  playwright.WorkflowBun,

    // Custom prepare for templ + Tailwind
    Prepare: prepareDatastarUI,

    ServerConfig: playwright.ServerConfig{
        Binary: "./sampleapp/datastarui-sample",
    },
}
```

### Deployment Auth Tests

```go
// core/tooling/tests/playwright_test.go
cfg := playwright.Config{
    SourceDir: "./tests",

    ServerConfig: playwright.ServerConfig{
        SkipServer: true,  // Testing external OAuth flows
    },
}
```

## Troubleshooting

### Bun not found

```bash
brew install oven-sh/bun/bun
```

### pnpm not found

```bash
npm install -g pnpm
```

### Server not ready

Increase `ServerConfig.StartTimeout`:

```go
cfg.ServerConfig.StartTimeout = 60 * time.Second
```

### Tests timeout

Increase overall timeout:

```go
cfg.Timeout = 10 * time.Minute
```

## API Reference

### Functions

- `Run(ctx, cfg) error` - Execute complete test workflow
- `InstallPlaywright(ctx, dir, workflow) error` - Install browsers
- `RunPlaywrightTests(ctx, dir, url, workflow, headed) error` - Run tests
- `StartServer(ctx, dir, cfg) (*exec.Cmd, error)` - Start server
- `StopServer(cmd)` - Stop server gracefully
- `WaitForHTTP(url, timeout) error` - Wait for readiness
- `VerifyWorkflow(workflow) error` - Check tools available

### Types

- `Config` - Complete test configuration
- `ServerConfig` - Server lifecycle configuration
- `WorkflowMode` - Bun or Node execution mode
