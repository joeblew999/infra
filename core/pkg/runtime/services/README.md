# Core Services

Built-in services that the orchestrator manages live here:

- `bus/` — embedded NATS/JetStream node wired into the event backbone
- `caddy/` — managed Caddy reverse proxy and module tooling
- `demo/*` — example workloads (including the JSON-driven demo) that prove reuse

Every service should depend only on `core/pkg/shared/*` helpers and register
through the controller/spec pipeline so the runtime can supervise it.
