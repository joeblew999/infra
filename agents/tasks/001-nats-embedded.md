# Task 001 — Embedded NATS & Service Isolation

## Context
- [x] `go run .` currently stalls when multiple services start, especially when they claim ports already in use (e.g. Hugo on 1313).
- [x] We recently wired `ShouldEnsureNATSCluster()` and ensured the embedded leaf can boot, but other services still assume shared startup and don’t check/reclaim their own resources.
- [x] Goal is to be able to start each service (and the full stack) idempotently from the new `./app` layout while keeping production-ready cluster orchestration available when explicitly enabled.
- [x] Task 000 handles the port-conflict detection/recovery; this task builds on that foundation to integrate per-service ensures and status wiring.

## Goals
- [x] Embedded NATS starts reliably in development without cluster orchestration.
- [x] Each service package exposes an `Ensure…` helper (dirs + port checks) that can be invoked independently.
- [x] `infra runtime list` and `infra runtime up --only <name>` provide the canonical CLI for listing and starting services; `go run .` reuses the same paths.
- [x] `infra runtime status` (CLI) and the web UI expose which services are running, using the existing `pkg/service` utilities and the port-conflict data from Task 000.
- [x] Document how to use `infra runtime list|up|status`, `go run .`, and the environment flags (`NATS_CLUSTER_ENABLED`, `APP_ROOT`).

## Deliverables
- [x] Integration of Task 000’s port detection into the per-service ensure helpers.
- [x] Updated CLI (`infra runtime list`, `infra runtime up --only <name>`, `infra runtime status`, plus `go run .` alias updates).
- [x] Documentation update summarising dev vs prod startup expectations, service inspection commands, and the new environment flags.
- [x] Validation notes: `infra runtime up --only nats`, `infra runtime up --only hugo`, `infra runtime list`, and `infra runtime status` all succeed (or produce clear guidance if something else is already bound to the port).
  - ✅ Plan: run each command manually, capture output snippets/guidance, and paste summary back here.
  - ✅ `go run . runtime up --only nats` (SIGINT after `runtime start complete`) → embedded leaf booted with `runtime.ready` event; `nats-s3` status surfaced as structured JSON and shutdown reclaimed 4222/5222.
  - ✅ `go run . runtime up --only hugo` (SIGINT post-ready) → docs server ensured dirs, reclaimed 1313, and exited cleanly with structured shutdown.
  - ✅ `go run . runtime list` → confirms required vs optional split; `deck-watcher` remains optional/disabled.
  - ✅ `go run . runtime status` → reports all services `stopped`/`free` when supervisor idle.

## Startup Classes
- **Always-on services (Required=true)**
  - Web Server (`web`) — control panel on port 1337.
  - Embedded NATS (`nats`) — leaf node with JetStream enabled (client port 4222, S3 gateway on 5222) that can bridge to the multi-node cluster when `NATS_CLUSTER_ENABLED` is true.
  - Caddy Reverse Proxy (`caddy`) — TLS/front door, defaults to port 80.

- **Default optional services (Required=false, Enabled by default)**
  - PocketBase (`pocketbase`) — database + admin UI on port 8090.
  - Bento (`bento`) — stream processor on port 4195.
  - Deck API (`deck-api`) — presentation API on port 8888.
  - XTemplate (`xtemplate`) — template dev server on port 8080.
  - Hugo Docs (`hugo`) — documentation site on port 1313.

- **Disabled until integration lands**
  - Deck Watcher (`deck-watch`).

- [x] Hugo should be part of the default “start individually” list so docs stay in sync.
- [x] We will surface service status on the existing dashboard as a dedicated panel.

## Next Steps (draft)
- [x] Execute Task 000 (port stability) and verify behaviour.
- [x] Integrate ensures + status updates across services.
- [ ] Validate runtime commands in dev environment (docs updated in README + CLI usage sections).
  - ✅ Action: manual validation session (no automation yet) + update README/CLI docs with any gotchas.


## Logging

- [x] Route embedded NATS server output through the structured logger so console lines match other services (`pkg/nats/logger.go`).
- [x] Confirm goreman-managed services also present consistent log metadata once other binaries are wrapped.
  - ✅ First step: tail current goreman logs during runtime start to verify existing output before adding wrappers.
  - ✅ Observed structured JSON lifecycle lines for `nats-s3`/others (`External process status`, `✅ Stopped process …`) during NATS/Hugo runs; no stray plain-text output from supervised binaries.
