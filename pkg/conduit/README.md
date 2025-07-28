# conduit

Binary management for [Conduit](https://github.com/ConduitIO/conduit) and its connectors.

## Quick Start

```bash
# Download all binaries
go test ./pkg/conduit -run TestPackageIntegration -v

# Verify binaries are ready
ls -la .dep/conduit*
```

## Usage

```go
// Download all binaries
err := conduit.Ensure(false)

// Get path to conduit binary
path := conduit.Get("conduit")
```

## Included Binaries

- **conduit** - Core Conduit binary (v0.12.1)
- **conduit-connector-s3** - S3 connector (v0.9.3)
- **conduit-connector-postgres** - Postgres connector (v0.14.0)
- **conduit-connector-kafka** - Kafka connector (v0.8.0)
- **conduit-connector-file** - File connector (v0.7.0)

## Running

### Process Management

Use the service layer to manage Conduit and connectors as processes:

```go
import "github.com/joeblew999/infra/pkg/conduit"

// Create service
service := conduit.NewService()

// Ensure binaries and start all processes
if err := service.EnsureAndStart(false); err != nil {
    log.Fatal(err)
}

// Start only core conduit
if err := service.StartCore(); err != nil {
    log.Fatal(err)
}

// Start only connectors
if err := service.StartConnectors(); err != nil {
    log.Fatal(err)
}

// Check status
status := service.Status()
for name, state := range status {
    fmt.Printf("%s: %s\n", name, state)
}

// Stop all processes
if err := service.Stop(); err != nil {
    log.Fatal(err)
}
```

### Service Methods
- `Start()` - Start all processes
- `Stop()` - Stop all processes gracefully
- `StartCore()` - Start only conduit core
- `StartConnectors()` - Start only connectors
- `StopCore()` - Stop only conduit core
- `StopConnectors()` - Stop only connectors
- `Restart()` - Restart all processes
- `Status()` - Get current process status

Configuration files are in `pkg/conduit/config/` and use the same format as `pkg/dep`.

  goreman - Ideal for:
  - Process supervision - Like foreman/systemd but cross-platform
  - Process groups - Start/stop conduit + connectors as a unit
  - Logging - Aggregated logs from all processes
  - Signal handling - Graceful shutdowns and restarts

  Combined Architecture:
  - goreman manages process lifecycle (start/stop/restart groups)
  - gopsutil provides health monitoring and resource usage
  - service.go orchestrates both for a complete solution

  Process Groups:
  conduit: conduit
  connectors: conduit-connector-s3, conduit-connector-postgres, etc.

  Service Flow:
  1. Start - goreman starts process group
  2. Monitor - gopsutil watches resource usage
  3. Health check - goreman status + gopsutil process info
  4. Restart - goreman restart on failure or config change