# Core Documentation

## Getting Started

- [Main README](../README.md) - Quick start and architecture overview
- [Development Guide](DEVELOPMENT.md) - Adding services, customizing, building
- [Troubleshooting](TROUBLESHOOTING.md) - Common issues and solutions

## Key Concepts

### Service Abstraction
Each service is described by a `service.json` file that is orchestrator-agnostic. This means you can swap out process-compose for docker-compose, systemd, or Kubernetes without changing service definitions.

### Placeholder Resolution
- `${dep.NAME}` → Binary path (e.g., `.dep/nats`)
- `${data}` → Data directory (`.data`)
- `${env.VAR}` → Environment variable value

### Health Checks
Services declare health checks in their `service.json`. Process-compose uses these to:
- Determine when a service is ready
- Manage dependencies (don't start B until A is healthy)
- Decide when to restart failed services

### Process Flow
1. Load service specs from `services/*/service.json`
2. Ensure binaries exist (build if needed)
3. Resolve all placeholders
4. Generate `.core-stack/process-compose.yaml`
5. Start process-compose with generated config
6. Monitor health checks and manage lifecycle

## Architecture

See [../README.md](../README.md) for complete architecture documentation.

## Services

- [NATS](../services/nats/README.md) - Message bus with JetStream
- [PocketBase](../services/pocketbase/README.md) - Database and auth
- [Caddy](../services/caddy/README.md) - Web server and reverse proxy
