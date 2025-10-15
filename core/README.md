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

Core orchestrates NATS, PocketBase, and Caddy as a unified stack with health checks and dependency management.

**Key Features**:
- Single command to start/stop services
- Automatic binary building
- Health-based startup ordering
- Environment configuration via `.env` (optional)

See [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) for design details.

## Configuration

**`.env` file** (optional):
```bash
# PocketBase admin (auto-created on first run)
CORE_POCKETBASE_ADMIN_EMAIL=admin@localhost
CORE_POCKETBASE_ADMIN_PASSWORD=changeme123

# SMTP for email features (optional)
CORE_POCKETBASE_SMTP_HOST=smtp.gmail.com
CORE_POCKETBASE_SMTP_PORT=587
# ... more SMTP settings
```

See [.env.example](.env.example) for full configuration options.

**Generated Files**:
- `.dep/` - Built service binaries
- `.data/` - Service data (databases, logs)
- `.core-stack/process-compose.yaml` - Generated orchestration config

See [docs/DEVELOPMENT.md](docs/DEVELOPMENT.md) for service configuration details.

## Commands

### Stack Management
```bash
go run . stack up              # Start all services
go run . stack down            # Stop all services
go run . stack status          # Show service status
```

### Process Control
```bash
go run . stack process list              # List all processes
go run . stack process logs <name>       # View service logs
go run . stack process restart <name>    # Restart a service
go run . stack process stop <name>       # Stop a service
go run . stack process start <name>      # Start a service
```

### Individual Services
```bash
go run . nats run              # Run NATS standalone
go run . pocketbase run        # Run PocketBase standalone
go run . caddy run             # Run Caddy standalone
```

### Access Services
- **PocketBase**: http://localhost:8090
- **PocketBase Admin**: http://localhost:8090/_/
- **Caddy Proxy**: http://localhost:2015
- **NATS Client**: nats://localhost:4222
- **NATS Monitoring**: http://localhost:8222

## Documentation

- **[docs/ARCHITECTURE.md](docs/ARCHITECTURE.md)** - System design, generation flow, design decisions
- **[docs/DEVELOPMENT.md](docs/DEVELOPMENT.md)** - Adding services, debugging, extending
- **[docs/TROUBLESHOOTING.md](docs/TROUBLESHOOTING.md)** - Common issues and solutions
- **[DEPLOYMENT.md](DEPLOYMENT.md)** - Fly.io deployment instructions
