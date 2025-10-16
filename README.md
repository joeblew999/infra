# Infra

**Monorepo with V1 (legacy) and V2 (core) architectures:**
- **V1 (this directory)**: Legacy runtime with goreman-based orchestration
- **V2 ([core/](core/README.md))**: Modern event-driven orchestration with process-compose ← **Active development**

## V1 - Legacy Runtime

Manage the whole stack with the `infra runtime` namespace:

```bash
# start the supervised stack (web UI, NATS, PocketBase, Bento, Deck API, Caddy, etc.)
go run . runtime up

# stop everything
go run . runtime down

# see what's running
go run . runtime status

# stream live lifecycle events
go run . runtime watch --service web --types status

# https://localhost:1337
# http://localhost:1337
```

## Quick Start

```bash
git clone https://github.com/joeblew999/infra.git
cd infra
go run . runtime up       # start the supervised services
# ... hack ...
go run . runtime down     # stop the services
```

## Everyday Commands

```bash
go run . runtime list         # list available services
go run . runtime status       # check local health
go run . runtime container    # build & run via ko + Docker
go run . workflows deploy     # deploy to Fly.io
go run . tools flyctl status  # access supporting tooling
go run . dev api-check        # compare Go API surfaces
```

## V2 - Core Runtime (Active Development)

See **[core/README.md](core/README.md)** for the modern architecture with:
- Process-compose orchestration
- Real-time observability via NATS JetStream
- Event-driven TUI/GUI
- Health-based dependency management
- Single binary deployment

## Need More?

This repo keeps deeper docs alongside the code:

- `docs/` – V1 architecture notes, service guides, CLI details
- `core/docs/` – V2 architecture, observability, development guides
- `pkg/` – package-level READMEs (goreman, nats, deck, etc.)
- `agents/` – instructions for automation agents working in this repo

If you ever forget what's available, run `go run . --help`.
