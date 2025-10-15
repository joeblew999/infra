# Core Runtime System

A deterministic, event-driven orchestration system for running distributed services locally and in production. Built on process-compose with full programmatic control.

## Quick Start

```bash
# Start the complete stack (NATS, PocketBase, Caddy)
go run . stack up

# Or use make
make run

# Check status
go run . stack status

# Stop services
go run . stack down
```

**Note**: Environment variables are auto-loaded from `.env` if it exists.

## What is This?

The **core** is a unified runtime system that:
- **Orchestrates services** using process-compose (embedded as a library)
- **Manages dependencies** between services with health checks
- **Abstracts service definitions** via `service.json` (works with any orchestrator)
- **Resolves placeholders** (`${env.*}`, `${dep.*}`, `${data}`) at generation time
- **Provides a single CLI** for all operations (local dev, deployment, debugging)

## Architecture

### Components

```
cmd/
├── core/           # Main CLI orchestrator
├── nats/           # NATS service binary
├── pocketbase/     # PocketBase service binary
├── caddy/          # Caddy service binary
└── processcompose/ # Wrapper around process-compose library

services/
├── nats/           # NATS service spec + implementation
├── pocketbase/     # PocketBase service spec + implementation
└── caddy/          # Caddy service spec + implementation

pkg/
├── runtime/        # Core runtime (CLI, process mgmt, config, UI)
└── shared/         # Shared packages (used by services + tooling)
```

### Service Abstraction

Each service has a `service.json` that describes:
- **Binaries** to build/download
- **Process** command, args, environment
- **Ports** exposed
- **Health checks** for orchestration
- **Compose overrides** for process-compose specific config

Example: `services/nats/service.json`
```json
{
  "binaries": [...],
  "process": {
    "command": "${dep.nats}",
    "env": {...},
    "compose": {
      "readiness_probe": {
        "http_get": {"url": "http://127.0.0.1:8222/healthz"},
        "initial_delay_seconds": 3,
        "failure_threshold": 5
      }
    }
  },
  "ports": {...}
}
```

### Placeholder Resolution

Placeholders are resolved during `process-compose.yaml` generation:

| Placeholder | Resolved To | Example |
|-------------|-------------|---------|
| `${dep.nats}` | Binary path | `.dep/nats` |
| `${data}` | Data directory | `.data` |
| `${env.VAR}` | Environment variable | Value of `$VAR` |

### Process Flow

1. **`go run ./cmd/core stack up`**
   ↓
2. **Generate** `.core-stack/process-compose.yaml`
   - Load `services/*/service.json`
   - Ensure binaries exist (build or download)
   - Resolve placeholders
   - Merge compose overrides
   ↓
3. **Execute** `go run ./cmd/processcompose up`
   - Start process-compose with generated config
   - Monitor health checks
   - Manage dependencies
   ↓
4. **Services run** with full control

## Key Files

| File | Purpose |
|------|---------|
| `.env` | Environment variables (auto-loaded by `cmd/core`) |
| `services/*/service.json` | Service manifest (orchestrator-agnostic) |
| `.core-stack/process-compose.yaml` | Generated orchestration config |
| `pkg/runtime/process/processcompose.go` | Config generator |
| `cmd/processcompose/main.go` | Our wrapper (full control) |

## Health Checks

Health checks prevent premature restarts:

```json
{
  "initial_delay_seconds": 3,    // Wait before first check
  "period_seconds": 5,            // Check interval
  "timeout_seconds": 3,           // Per-check timeout
  "failure_threshold": 5,         // Failures before restart
  "success_threshold": 1          // Successes to mark healthy
}
```

## Service Dependencies

Dependencies are declared in `buildComposeDefinition()`:

```go
// PocketBase depends on NATS being healthy
ensureDependsOn(pbEntry, map[string]map[string]any{
    "nats": {"condition": "process_healthy"},
})
```

## Stack Commands

```bash
# Lifecycle
go run ./cmd/core stack up        # Start all services
go run ./cmd/core stack down      # Stop all services
go run ./cmd/core stack status    # Show status

# Process management
go run ./cmd/core stack process list
go run ./cmd/core stack process logs <name>
go run ./cmd/core stack process restart <name>
go run ./cmd/core stack process stop <name>
go run ./cmd/core stack process start <name>

# Individual services
go run ./cmd/core nats run        # Run NATS standalone
go run ./cmd/core pocketbase run  # Run PocketBase standalone
go run ./cmd/core caddy run       # Run Caddy standalone
```

## Deployment

See [DEPLOYMENT.md](DEPLOYMENT.md) for Fly.io deployment instructions.

## Development

See [docs/DEVELOPMENT.md](docs/DEVELOPMENT.md) for:
- Adding new services
- Customizing process-compose behavior
- Debugging health checks
- Working with the codebase

## Troubleshooting

See [docs/TROUBLESHOOTING.md](docs/TROUBLESHOOTING.md)
