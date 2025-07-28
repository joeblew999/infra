# goreman

A lightweight process manager inspired by [goreman](https://github.com/mattn/goreman) but implemented as a reusable Go package.



## Features

- **Process lifecycle management** - Start/stop/restart processes
- **Process groups** - Manage related processes as units
- **Cross-platform** - Works on Linux, macOS, and Windows
- **Health monitoring** - Built-in process health checks
- **Graceful shutdown** - Proper signal handling
- **Logging** - Aggregated process logs

## Usage

```go
import "github.com/joeblew999/infra/pkg/goreman"

// Create a new manager
manager := goreman.NewManager()

// Add processes
manager.AddProcess("conduit", &goreman.ProcessConfig{
    Command: "conduit",
    Args:    []string{"--config", "conduit.yml"},
    Env:     []string{"ENV=production"},
})

// Add process group
manager.AddGroup("connectors", []*goreman.ProcessConfig{
    {Name: "s3", Command: "conduit-connector-s3"},
    {Name: "postgres", Command: "conduit-connector-postgres"},
})

// Start all processes
if err := manager.Start(); err != nil {
    log.Fatal(err)
}

// Stop gracefully
if err := manager.Stop(); err != nil {
    log.Fatal(err)
}
```

## Process Management

```go
// Start individual process
manager.StartProcess("conduit")

// Start process group
manager.StartGroup("connectors")

// Get process status
status := manager.GetStatus("conduit")

// Restart process
manager.RestartProcess("conduit")
```

## Configuration

Define processes using a Procfile-like format:

```go
// Load from Procfile
manager.LoadProcFile("Procfile")

// Or define programmatically
manager.AddProcess("web", &goreman.ProcessConfig{
    Command: "conduit",
    Args:    []string{"--web"},
    Port:    8080,
    HealthCheck: &goreman.HealthCheck{
        URL: "http://localhost:8080/health",
        Timeout: 5 * time.Second,
    },
})
```


Missing pieces for production readiness:

  Process Management:
  - Restart policies - Auto-restart on failure (max attempts, backoff)
  - Dependency ordering - Start order enforcement (conduit before connectors)
  - Resource limits - CPU/memory constraints per process
  - Timeout handling - Startup/shutdown timeouts

  Error Handling:
  - Process failure detection - Exit code monitoring
  - Circuit breaker - Stop restarting if process keeps failing
  - Notification system - Webhooks/events on process state changes

  Configuration:
  - Procfile parsing - Load from external config files
  - Environment templating - Dynamic variable substitution
  - Validation - Config validation before starting

  Monitoring Integration:
  - pkg/gops integration - Process stats via existing monitoring
  - Health check endpoints - HTTP health probes
  - Metrics collection - Process metrics for monitoring

  Operational:
  - Log rotation - Prevent log file bloat
  - PID file management - Track running processes
  - Signal forwarding - Proper signal propagation to child processes
  - Graceful degradation - Partial startup if some processes fail

  CLI Integration:
  - Status dashboard - Human-readable process overview
  - Interactive commands - Start/stop/restart via CLI
  - Configuration reload - Hot-reload Procfile changes

  These would make it production-ready while keeping the simple, reusable design.