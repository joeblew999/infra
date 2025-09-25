# Test Project — Embedded Runtime Playground

This project exists as a sandbox for experimenting with the service runtime outside of the main CLI entry point. It boots the same supervised stack (web, NATS, Caddy, PocketBase, Bento, Deck API, XTemplate, Hugo) via `runtime.Start`, making it easy to layer project-specific experiments on top.

## Why this folder?

- **Service runtime experiments** – Validate how flexible the `pkg/service/runtime` package is without touching `main.go`.
- **Future JSON specs** – Prototype loading custom service specs or overrides before promoting them into the core runtime.
- **PocketBase replication** – Explore Litestream + NATS object store replication flows in isolation.

## Getting Started

```bash
go run ./projects/test
```

You’ll see the same supervised service startup sequence as `go run . runtime up`, but driven from this isolated entry point.

## Litestream + PocketBase + NATS Object Store

The long-term experiment is to prototype replicated PocketBase storage using Litestream with NATS JetStream as the backing object store. Suggested flow:

1. **Bootstrap services** – Run `go run ./projects/test` to launch the core stack.
2. **Enable NATS object store** – Use the embedded NATS JetStream instance to provision an object store bucket for Litestream.
3. **Configure Litestream** – Point Litestream at the PocketBase SQLite database paths given by `config.GetPocketBaseDataPath()` and target the NATS object store endpoint.
4. **Simulate writes** – Insert data via the PocketBase admin UI (`http://localhost:8090`) or REST API.
5. **Replay from backup** – Stand up a second PocketBase instance (or restart with an empty data directory) and restore state from the Litestream-managed backups.
6. **Capture learnings** – Record configuration files, environment variables, and failure modes back in this README or accompanying docs.

None of the Litestream wiring is added yet—this directory is the staging area for that work.

## Next Ideas

- Try overriding `runtime.Options` (e.g., `OnlyServices`, `SkipServices`) to validate partial-stack boot.
- Experiment with custom `Preflight` hooks to seed configuration or secrets before services start.
- Prototype a JSON/YAML service-spec loader and feed it into `runtime.Start` from this entry point.

Contributions or notes from experiments should live here so the main repo stays stable while we explore.
