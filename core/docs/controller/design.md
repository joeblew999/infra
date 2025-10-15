# Scaling Controller Design Draft

This document sketches the controller that reconciles desired service state with
observed metrics and infrastructure capabilities.

## Goals

1. Keep Process Compose manifests deterministic (replicas per host = 1 for
   port-bound services).
2. Offer a declarative API for desired replicas per service/region.
3. Allow metrics-driven overrides within safe bounds.
4. Integrate with Fly.io for provisioning/teardown (scale-to-zero when idle).

## Core Concepts

### Desired State Schema

Defined in `controller/spec.yaml` and represented by `pkg/controller/spec`.

- `scale.strategy` — `infra` (default) or `local`.
- `scale.autoscale` — `manual`, `metrics`, or `disabled`.
- `scale.regions` — per-region min/desired/max, cooldown, burst TTL.
- `storage.provider` — e.g. `cloudflare-r2`; includes bucket names and
  credentials references for Litestream replicas and PocketBase assets.
- `routing.provider` — e.g. `cloudflare`; defines zone, DNS records, health
  checks, and load-balancing weights.

### Runtime State

- Active Process Compose stacks (Fly Machines) per service/region.
- Metrics: queue depth, latency, CPU (start with JetStream consumer lag).
- Outstanding scaling actions (create/destroy).

### Controller Loop

1. Load desired state.
2. Collect runtime state (Fly API, metrics backend, Cloudflare).
3. Diff desired vs actual; schedule create/destroy actions.
4. Apply actions (provision machines, update Cloudflare origins, configure
   Litestream targets, rotate credentials).
5. Persist status and emit events for UI/CLI.

### API Sketch

- `GET /v1/services` — current desired vs actual.
- `PATCH /v1/services/{id}` — update desired state.
- `POST /v1/services/{id}/burst` — temporary capacity increase with TTL.
- `POST /v1/services/{id}/reconcile` — manual reconcile trigger.

### Integration Points

- **UI** — call controller API instead of local Process Compose for infra
  scaling; display desired vs actual + metrics.
- **CLI** — `core scale show/set/burst` to inspect or modify desired state.
- **Litestream** — controller renders per-stack config pointing to Cloudflare R2
  buckets defined in the spec.
- **Cloudflare** — controller manages DNS records and load-balancing weights
  based on desired regional replicas.

## Open Questions

- Desired state storage backend (git repo, database, S3/R2).
- Credential management for Cloudflare and Fly (HashiCorp Vault vs. Fly secrets).
- Handling conflicting requests (manual overrides vs metrics autoscale).
- Observability: how to expose reconciliation events to operators.
