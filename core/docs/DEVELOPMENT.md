# Development Guide

## Adding a New Service

1. **Create service directory**:
```bash
mkdir -p services/myservice
cd services/myservice
```

2. **Create `service.json`**:
```json
{
  "binaries": [
    {
      "name": "myservice",
      "version": "1.0.0",
      "source": "go-build",
      "path": "./cmd/myservice"
    }
  ],
  "process": {
    "command": "${dep.myservice}",
    "args": ["--config", "${data}/myservice/config.yml"],
    "env": {
      "MY_VAR": "${env.MY_VAR}"
    },
    "compose": {
      "availability": {
        "restart": "always"
      },
      "readiness_probe": {
        "http_get": {"url": "http://127.0.0.1:9000/health"},
        "initial_delay_seconds": 5,
        "period_seconds": 5,
        "timeout_seconds": 3,
        "failure_threshold": 5,
        "success_threshold": 1
      }
    }
  },
  "ports": {
    "http": {"port": 9000, "protocol": "http"}
  }
}
```

3. **Create service implementation** (`service.go`):
```go
package myservice

import (
    "context"
    "embed"
    "encoding/json"
    
    runtimedep "github.com/joeblew999/infra/core/pkg/runtime/dep"
)

//go:embed service.json
var manifestFS embed.FS

type Spec struct {
    Binaries []runtimedep.BinarySpec `json:"binaries"`
    Process  ProcessSpec             `json:"process"`
    Ports    PortsSpec               `json:"ports"`
}

func LoadSpec() (*Spec, error) {
    data, err := manifestFS.ReadFile("service.json")
    if err != nil {
        return nil, err
    }
    var spec Spec
    if err := json.Unmarshal(data, &spec); err != nil {
        return nil, err
    }
    return &spec, nil
}

func Run(ctx context.Context, args []string) error {
    // Implementation
    return nil
}
```

4. **Create cmd entry** (`cmd/myservice/main.go`):
```go
package main

import (
    "context"
    "fmt"
    "os"
    "os/signal"
    "syscall"
    
    "github.com/joeblew999/infra/core/services/myservice"
)

func main() {
    ctx, cancel := signal.NotifyContext(
        context.Background(),
        syscall.SIGINT,
        syscall.SIGTERM,
    )
    defer cancel()
    
    if err := myservice.Run(ctx, os.Args[1:]); err != nil {
        fmt.Fprintf(os.Stderr, "myservice: %v\n", err)
        os.Exit(1)
    }
}
```

5. **Register in process generator** (`pkg/runtime/process/processcompose.go`):
```go
import myservicesvc "github.com/joeblew999/infra/core/services/myservice"

func buildComposeDefinition(root string) (map[string]any, error) {
    // ... existing code ...
    
    myserviceSpec, err := myservicesvc.LoadSpec()
    if err != nil {
        return nil, fmt.Errorf("myservice spec: %w", err)
    }
    myservicePaths, err := myserviceSpec.EnsureBinaries()
    if err != nil {
        return nil, fmt.Errorf("myservice ensure binaries: %w", err)
    }
    
    myserviceEnv := myserviceSpec.ResolveEnv(myservicePaths)
    myserviceArgs := relativeArgs(root, myserviceSpec.ResolveArgs(myservicePaths))
    myserviceEntry := composeProcessEntry(
        root,
        relativeCommand(root, myservicePaths["myservice"]),
        myserviceArgs,
        myserviceEnv,
        myserviceSpec.ComposeOverrides(),
    )
    processes["myservice"] = myserviceEntry
    
    // ... rest of code ...
}
```

## Customizing Process-Compose

Since we embed process-compose as a library, we have full control. Our wrapper is at `cmd/processcompose/main.go`:

```go
package main

import (
    "fmt"
    "os"
    pccmd "github.com/f1bonacc1/process-compose/src/cmd"
)

func main() {
    // Add pre-startup hooks here
    
    // Delegate to upstream
    pccmd.Execute()
    
    // Add post-shutdown hooks here
}
```

## Debugging Health Checks

### View Generated Config
```bash
cat .core-stack/process-compose.yaml
```

### Test Health Endpoint Manually
```bash
# NATS
curl http://127.0.0.1:8222/healthz

# PocketBase
curl http://127.0.0.1:8090/api/health
```

### Watch Process Status
```bash
watch -n 1 'go run ./cmd/core stack status'
```

### View Process Logs
```bash
go run ./cmd/core stack process logs nats
go run ./cmd/core stack process logs pocketbase
```

### Common Issues

**Service restarts immediately**:
- Check `initial_delay_seconds` is sufficient
- Increase `failure_threshold`
- Verify health endpoint actually works

**Service stuck in "Pending"**:
- Dependency not healthy
- Check dependent service health check

**Service shows "Skipped"**:
- Dependency failed
- Check logs of dependency

## Project Structure

```
core/
├── cmd/                    # Binaries
│   ├── core/              # Main CLI
│   ├── nats/              # NATS service
│   ├── pocketbase/        # PocketBase service
│   ├── caddy/             # Caddy service
│   └── processcompose/    # Process-compose wrapper
├── services/              # Service definitions
│   ├── nats/
│   │   ├── service.json   # Service manifest
│   │   ├── service.go     # Implementation
│   │   └── README.md
│   ├── pocketbase/
│   └── caddy/
├── pkg/
│   ├── runtime/           # Core runtime packages
│   │   ├── cli/          # CLI commands
│   │   ├── config/       # Config loading
│   │   ├── process/      # Process management
│   │   └── ui/           # TUI components
│   └── shared/           # Shared utilities
├── .env                   # Environment variables
├── .core-stack/          # Generated files
│   └── process-compose.yaml
└── .data/                # Service data
    ├── nats/
    └── pocketbase/
```

## Environment Variables

Auto-loaded from `.env` by `cmd/core/main.go`:

```go
import "github.com/joho/godotenv"

func main() {
    _ = godotenv.Load()  // Silent fail in production
    // ...
}
```

## Building

```bash
# Build all binaries
go build -o .dep/core ./cmd/core
go build -o .dep/nats ./cmd/nats
go build -o .dep/pocketbase ./cmd/pocketbase
go build -o .dep/caddy ./cmd/caddy

# Or let the system build them on demand
go run ./cmd/core stack up
```
