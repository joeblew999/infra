# Core Runtime Architecture

## Overview

Core is a **deterministic, self-contained runtime stack** for local development and production deployment. It orchestrates NATS (message bus), PocketBase (database/backend), and Caddy (HTTP server) with health-based dependency management.

## Why Core Exists

**Architectural Independence:**
- ZERO dependencies on parent `infra/pkg/*` code
- Completely standalone Go module
- Can be extracted and run anywhere

**Goals:**
1. **Portability** - Move it anywhere, no external dependencies
2. **Determinism** - Same behavior every time, predictable startup
3. **Simplicity** - One command (`go run . up`), everything works
4. **Purity** - No leaky abstractions, use native APIs directly

## Architecture Layers

```
┌──────────────────────────────────────────────────────────┐
│  CLI Entry Point (cmd/core/main.go)                      │
│  Commands: up, down, status, nats, pocketbase, caddy     │
└────────────────┬─────────────────────────────────────────┘
                 │
                 ▼
┌──────────────────────────────────────────────────────────┐
│  Runtime Orchestration (pkg/runtime/)                    │
│  - Reads service.json manifests                          │
│  - Builds binaries to .dep/                              │
│  - Generates process-compose.yaml                        │
│  - Manages process-compose lifecycle                     │
└────────────────┬─────────────────────────────────────────┘
                 │
                 ▼
┌──────────────────────────────────────────────────────────┐
│  Process-Compose (external)                              │
│  - Starts processes with health checks                   │
│  - Enforces dependency ordering                          │
│  - Restarts on failure                                   │
└────────────────┬─────────────────────────────────────────┘
                 │
        ┌────────┴────────┬───────────┐
        ▼                 ▼           ▼
   ┌─────────┐      ┌──────────┐  ┌──────┐
   │  .dep/  │      │  .dep/   │  │ .dep/│
   │  nats   │      │pocketbase│  │caddy │
   └─────────┘      └──────────┘  └──────┘
        │                 │           │
        ▼                 ▼           ▼
   [NATS      ]    [PocketBase  ]  [Caddy  ]
   Port 4222        Port 8090       Port 2015
```

## Service Structure

Each service follows this pattern:

```
services/{service}/
├── service.json    # Manifest (binaries, ports, env, health)
├── service.go      # Implementation with Run(ctx) entry point
├── README.md       # Service-specific documentation
└── cmd/{service}/  # Thin binary wrapper in ../../cmd/{service}/
```

### Service Manifest (service.json)

```json
{
  "binaries": [
    {"name": "nats", "source": "go-build", "path": "./cmd/nats"}
  ],
  "process": {
    "command": "${dep.nats}",
    "env": {"NATS_DATA_DIR": "${data}/nats"},
    "compose": {
      "readiness_probe": {
        "http_get": {"url": "http://127.0.0.1:8222/healthz"},
        "initial_delay_seconds": 10,
        "period_seconds": 5
      }
    }
  },
  "ports": {
    "client": {"port": 4222, "protocol": "nats"},
    "http": {"port": 8222, "protocol": "http"}
  },
  "config": {"jetstream": true}
}
```

### Service Implementation (service.go)

```go
//go:embed service.json
var manifestFS embed.FS

func LoadSpec() (*Spec, error) {
    // Parse embedded service.json
}

func Run(ctx context.Context, args []string) error {
    // 1. Load spec
    // 2. Start service (e.g., NATS server)
    // 3. Wait for health
    // 4. Print "READY: ..."
    // 5. Block on <-ctx.Done()
    // 6. Graceful shutdown
}
```

### Command Binary (cmd/{service}/main.go)

```go
func main() {
    ctx, cancel := signal.NotifyContext(
        context.Background(),
        syscall.SIGINT,
        syscall.SIGTERM,
    )
    defer cancel()

    if err := service.Run(ctx, os.Args[1:]); err != nil {
        fmt.Fprintf(os.Stderr, "%s: %v\n", service, err)
        os.Exit(1)
    }
}
```

## Orchestration Flow

### Startup (`go run . up`)

1. **EnsureRuntime()** - Build all service binaries
   ```
   services/nats/service.json → go build cmd/nats → .dep/nats
   services/pocketbase/service.json → go build cmd/pocketbase → .dep/pocketbase
   services/caddy/service.json → go build cmd/caddy → .dep/caddy
   ```

2. **GenerateProcessComposeConfig()** - Create orchestration manifest
   ```yaml
   processes:
     nats:
       command: .dep/nats
       readiness_probe:
         http_get: {url: "http://127.0.0.1:8222/healthz"}

     pocketbase:
       command: .dep/pocketbase
       depends_on:
         nats: {condition: process_healthy}
       readiness_probe:
         http_get: {url: "http://127.0.0.1:8090/api/health"}

     caddy:
       command: .dep/caddy
       depends_on:
         pocketbase: {condition: process_healthy}
   ```

3. **StartProcessCompose()** - Execute orchestration
   ```
   → Start .dep/nats
   → Poll http://127.0.0.1:8222/healthz every 5s
   → When healthy: start .dep/pocketbase
   → Poll http://127.0.0.1:8090/api/health every 5s
   → When healthy: start .dep/caddy
   ```

### Health Check Flow

```
[NATS Binary]
    ↓
  ns.Start()  // NATS server goroutine
    ↓
  Wait for ports 4222, 8222 to be listening
    ↓
  Print "READY: nats tcp://127.0.0.1:4222"
    ↓
  <-ctx.Done()  // Block until SIGTERM
    ↓
  ns.Shutdown()

[Process-Compose]
    ↓
  Wait 10 seconds (initial_delay)
    ↓
  Poll GET http://127.0.0.1:8222/healthz
    ↓
  Success? → Mark NATS as healthy
    ↓
  Start PocketBase (depends on NATS healthy)
```

## Key Design Decisions

### 1. No Wrappers - Use Native APIs

**Before:**
```go
import "github.com/Nintron27/pillow"  // Wrapper library
server, err := pillow.Run(opts...)    // Hidden complexity
```

**After:**
```go
import "github.com/nats-io/nats-server/v2/server"
ns, err := server.NewServer(natsOpts)
go ns.Start()
```

**Why:** Direct control, no abstraction leaks, clear behavior.

### 2. No Cross-Module Dependencies

**Before:**
```go
import (
    "github.com/joeblew999/infra/pkg/config"      // Parent module
    "github.com/joeblew999/infra/pkg/nats/auth"   // Parent module
)
```

**After:**
```go
import (
    "github.com/joeblew999/infra/core/pkg/runtime/config"  // core internal
)
```

**Why:** Core is standalone, can be extracted without bringing parent code.

### 3. Signal Handling in Binary, Not Service

Command binaries (`cmd/*/main.go`) handle signals. Services (`services/*/service.go`) just block on `ctx.Done()`.

**Why:** Separation of concerns. Service doesn't care HOW context gets cancelled, binary controls process lifecycle.

### 4. Manifest-Driven Everything

Service behavior defined in `service.json`, not scattered through code.

**Why:** Single source of truth, easy to understand/modify, tooling can parse it.

## Architecture Purity (Recent Changes)

### Before (Impure)

**NATS service had:**
- `github.com/Nintron27/pillow` - Clustering wrapper (unnecessary)
- `github.com/joeblew999/infra/pkg/config` - Parent module config
- `github.com/joeblew999/infra/pkg/nats/auth` - NSC authentication (complex)
- `github.com/nats-io/jwt/v2` - JWT parsing (unused for local)

**service.json had:**
- `"backend": "pillow"`
- `"topology": "hub-spoke"`
- `nsc` binary for authentication
- `PILLOW_*` environment variables

### After (Pure)

**NATS service:**
- Direct `nats-server/v2` API only
- Simple standalone server
- No authentication (local development)
- No clustering (single node)

**service.json:**
- `"backend": "standalone"`
- `"topology": "single-node"`
- Only `nats` binary
- Only `NATS_DATA_DIR` environment variable

**Result:**
- 50 fewer lines of code
- No external dependencies outside core
- Clearer, simpler, maintainable

## Directory Structure

```
core/
├── cmd/                    # Command binaries
│   ├── core/               # Main CLI entry (go run . up)
│   ├── nats/               # NATS binary wrapper
│   ├── pocketbase/         # PocketBase binary wrapper
│   └── caddy/              # Caddy binary wrapper
├── services/               # Service implementations
│   ├── nats/
│   │   ├── service.json    # Manifest
│   │   ├── service.go      # Implementation
│   │   └── README.md
│   ├── pocketbase/
│   └── caddy/
├── pkg/
│   ├── runtime/            # Orchestration logic
│   │   ├── config/         # Runtime config
│   │   ├── dep/            # Binary building
│   │   └── process/        # Process-compose management
│   └── shared/             # Shared utilities
├── docs/
│   └── ARCHITECTURE.md     # This file
├── .dep/                   # Built binaries (gitignored)
├── .core-stack/            # Generated configs (gitignored)
└── .data/                  # Runtime data (gitignored)
```

## Commands

```bash
go run . up              # Start full stack
go run . down            # Stop stack
go run . status          # Show status

go run . nats spec       # Show NATS config
go run . nats run        # Run NATS standalone
go run . nats command    # Show NATS command

go run . pocketbase run  # Run PocketBase standalone
go run . caddy run       # Run Caddy standalone
```
