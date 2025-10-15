# NATS

NATS message bus service for the core runtime, driven entirely by
[`pillow`](https://github.com/Nintron27/pillow). The same binary launched
locally is deployed on Fly, so behaviour stays identical across environments.

## Tools

The manifest ensures the following supporting binaries are available alongside
Pillow:

- `pillow` – orchestrates NATS locally and on Fly
- `nats` – upstream NATS CLI for diagnostics
- `nsc` – JWT/auth management CLI



## Features

- **Pillow Runtime**: Embedded Pillow runner orchestrates the bundled NATS server
- **Topology Flags**: Hook for hub-spoke vs mesh (future work via shared config)
- **JetStream Ready**: JetStream enabled by default with data rooted under `config.Paths.Data`
- **Monitoring**: HTTP monitoring exposed on the configured port (default 8222)

## Configuration

Backend selection is locked to Pillow for now. Topology fields remain in the spec
for future multi-region orchestration once the new process runner lands.

## Ports

- **4222** - NATS client connections
- **6222** - NATS cluster communication
- **8222** - HTTP monitoring interface
- **7422** - Leaf node connections

## CLI Usage

```bash
# Build/run the dedicated binary (same for local and Fly)
go run ./core/cmd/nats -- --help

# Show configuration
core nats spec

# Print the resolved command (with env)
core nats command --env

# Ensure binaries
core nats ensure

# Run locally with extra tracing flags
core nats run -- --trace
```
