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
- [ ] Validation notes: `infra runtime up --only nats`, `infra runtime up --only hugo`, `infra runtime list`, and `infra runtime status` all succeed (or produce clear guidance if something else is already bound to the port).

## Open Questions / Needs Input
- [ ] Any additional services we should include in the default “start individually” list (e.g. docs-hugo variants)?
- [ ] How should the web UI present service status (new page vs. existing dashboard)?

## Next Steps (draft)
- [x] Execute Task 000 (port stability) and verify behaviour.
- [x] Integrate ensures + status updates across services.
- [ ] Validate runtime commands in dev environment (docs updated in README + CLI usage sections).


## Logging

- [x] Route embedded NATS server output through the structured logger so console lines match other services (`pkg/nats/logger.go`).
- [ ] Confirm goreman-managed services also present consistent log metadata once other binaries are wrapped.
