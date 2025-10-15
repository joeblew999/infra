# Runtime CLI Guide

This document collects the runtime (core) commands so the top-level README can
stay focused on quick-start workflows.

## Local Stack Basics
```sh
# Inspect top-level commands (returns immediately)
go run ./cmd/core --help

# Start the full stack (blocks until you interrupt; alias: core stack run)
go run ./cmd/core stack up

# Stop the stack gracefully
go run ./cmd/core stack down

# Check stack status without starting it
go run ./cmd/core stack status
```

### Process Compose Interaction
```sh
# List managed processes without restarting the stack
go run ./cmd/core stack process list --json | jq

# Inspect a single process entry
go run ./cmd/core stack process info pocketbase --json | jq

# Start/stop/restart individual processes
go run ./cmd/core stack process start pocketbase
go run ./cmd/core stack process stop pocketbase --json
go run ./cmd/core stack process restart nats

# Tail logs or truncate them
go run ./cmd/core stack process logs pocketbase --lines 50
go run ./cmd/core stack process truncate pocketbase
```

### Binary/Config Maintenance
```sh
# Ensure embedded service binaries are staged
go run ./cmd/core caddy ensure
go run ./cmd/core pocketbase ensure
go run ./cmd/core nats ensure

# Inspect or update the Process Compose project
go run ./cmd/core stack project state --json | jq
go run ./cmd/core stack project update --file overrides.json
go run ./cmd/core stack project reload --json
```

### UI Surfaces
```sh
# Render the live TUI snapshot (returns immediately)
go run ./cmd/core tui
# Controls on service detail pages: p=start, o=stop, r=restart, s=scale

# Serve the web UI shell (blocks until Ctrl+C)
go run ./cmd/core web
```

### Scaling & Controller Interaction
```sh
# Inspect desired scaling spec (controller API first, fallback to local file)
go run ./cmd/core scale show --controller http://127.0.0.1:4400 --file controller/spec.yaml

# Apply a service scaling update (expects a single-service YAML)
go run ./cmd/core scale set --controller http://127.0.0.1:4400 --file controller/service-pocketbase.yaml

# Stream controller desired-state events (SSE endpoint)
go run ./cmd/core controller watch --controller http://127.0.0.1:4400
```

> **Note**: scaling commands require the controller API to be running. See
> `controller/README.md` for details on starting it locally.

### Ports & Environment
- Process Compose listens on port `28081` by default. Override with
  `core stack up -- --port <port>` or by exporting `PC_PORT_NUM` before running
  the CLI.
- The same port is used for status queries, so keep it reachable if you
  customise it.

For additional runtime notes (multi-region scaling, deterministic startup
pipeline) consult `docs/SCALING.md` and `pkg/runtime/README.md`.
