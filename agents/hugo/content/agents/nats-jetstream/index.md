---
title: "NATS JetStream Agent Guide"
summary: "Streams, consumers, and infra-specific JetStream practices."
draft: false
---

## Sources (do not remove)
- **Official docs**
  - https://docs.nats.io/nats-concepts/jetstream
  - https://docs.nats.io/using-nats/jetstream/deploy
  - https://docs.nats.io/running-a-nats-service/configuration/jetstream
- **Server & core tooling**
  - https://github.com/nats-io/nats-server
  - https://github.com/nats-io/nsc
  - https://github.com/nats-io/natscli
- **Admin SDKs & examples**
  - https://github.com/nats-io/jsm.go
  - https://github.com/nats-io/nats.deno/tree/main/jetstream
- **Observability**
  - https://github.com/nats-io/nats-surveyor

## Part 1 – JetStream Fundamentals

### 1.1 Core Concepts
- **Streams** persist messages published to matching subjects. Configure subjects, storage (file/memory), replication, retention (limits|interest|workqueue), and max bytes/msgs/age.
- **Consumers** read from streams. Two flavors: *push* (server delivers to subject) and *pull* (client fetches batches). Acks (`Ack`, `AckSync`, `Nak`, `Term`, `InProgress`) determine redelivery.
- **Deduplication** uses `Msg-Id` header with a configurable window (`duplicate_window`).
- **Mirrors & Sources** let you aggregate or fan out streams for multi-region or multi-tenant setups.
- **Object Store & Key Value** build on JetStream streams; they behave like bucket APIs with chunked storage and metadata revisions.

### 1.2 Enabling JetStream on a NATS Server
```bash
nats-server --js --store_dir=/var/lib/nats/js --sd=/var/lib/nats/streams
```
```hcl
# server.conf
jetstream: {
  store_dir: "/var/lib/nats/js"
  max_mem:  4Gi
  max_file: 20Gi
}
```
- `store_dir` must sit on a persistent volume. Plan capacity for replication factor × max bytes.
- Clusters require routes and explicit replicas: configure `servers`, `cluster` block, and ensure `jetstream.domain` when federating across clusters.
- Keep `max_outstanding` and `mem_storage` in mind if you blend memory- and file-backed streams.

### 1.3 Lifecycle & Tooling
- **nats CLI** handles server status, stream/consumer CRUD, and message inspection.
  - `nats status jetstream`
  - `nats stream add --config stream.json`
  - `nats consumer next <stream> <consumer> --count 10`
- **nsc** provisions operators, accounts, users, and JWTs. Store material in git-friendly paths; rotate credentials via `nsc generate creds`.
- **jsm.go** (`jsm`) offers higher-level admin helpers for Go automation.
- **nats-surveyor** continuously scrapes JetStream metrics; deploy it when you need dashboards or alerting.

### 1.4 Operating Streams
- Choose retention based on workload: `limits` (bounded history), `interest` (delete after all consumers ack), `workqueue` (one-and-done tasks).
- Control redeliveries with `ack_wait`, `max_ack_pending`, and `max_deliver`.
- Use stream snapshots and restores for migrations (`nats stream snapshot`, `nats stream restore`).
- Monitor lag via `nats consumer info` and `nats server report jetstream`. Backpressure shows up as increasing pending or error counts.

### 1.5 Security & Tenancy
- Segment subjects by account using `nsc`—JetStream resources are scoped per account.
- Prefer strict limits: set `max_consumers`, `max_ack_pending`, and `allow_ack_all` to avoid runaway consumers.
- For multi-tenant setups, prefix subjects (e.g., `tenantA.events.*`). Pair with dedicated credentials and advertise only the required subjects via exports/imports.

### 1.6 Observability & Backups
- Expose the monitoring port (`http: 8222` by default) for Prometheus scrapes or manual inspection.
- `nats-surveyor` plus Grafana provides per-stream, per-consumer dashboards.
- Automate snapshots to S3 or another object store for disaster recovery. Validate restores regularly; JetStream refuses to restore into a different cluster name by default.

---

## Part 2 – JetStream Inside infra

### 2.1 Architecture Overview
- `pkg/nats` boots the embedded NATS server early in `go run . runtime up`. JetStream is enabled by default for local and single-node runs.
- Production mode can bridge to an external cluster when `NATS_CLUSTER_ENABLED=1`. Leaf nodes still mount JetStream storage locally while mirroring to the core cluster.
- We experiment with the NATS S3 gateway (`nats-s3` listener on `config.GetNatsS3Port()`), giving JetStream a cold tier for larger artifacts.

### 2.2 Configuration Entry Points
- `pkg/config/nats.go` centralizes ports, hosts, and paths:
  - `config.GetNATSURL()` → `nats://<host>:<port>`
  - `config.GetNATSClusterDataPath()` chooses `.data/` vs `.data-test/`
  - `config.GetNATSAuthStorePath()` stores `nsc` material under `nats-auth/`
  - `config.GetNATSLogStreamName()` defines the shared `LOGS` stream (`logs.app`).
- Respect environment knobs:
  - `NATS_HOST` overrides the default local hostname.
  - `NATS_CLUSTER_ENABLED` toggles multi-node clustering.
  - `APP_ROOT` adjusts data directories during tests or packaged runs.

### 2.3 Boot & Provisioning Workflow
1. `go run . runtime up` starts embedded NATS plus other services. Confirm with `go run . runtime status`.
2. `pkg/nats/auth` wraps `nsc` to generate operator/account/user JWTs in `.data/nats-auth/` (or `.data-test/...`).
3. Streams the platform sets up today:
   - `LOGS` (subjects `logs.app`) retains service logs for dashboards.
   - Feature branches add tenant-prefixed streams; ensure subjects live under `tenant.<id>.*`.
4. For object-store experiments, `pkg/nats/gateway` exposes an S3-compatible endpoint that fronts JetStream buckets. Ports come from `config.GetNatsS3Port()`; update Process Compose or goreman configs when enabling it.

### 2.4 Working With JetStream Data
- Use the bundled `nats` CLI: `go run . tools dep install nats` if missing, then `nats status jetstream --server $(config.GetNATSURL())`.
- When debugging stream consumers in Go, rely on `github.com/nats-io/nats.go` and `jsm.go`. Keep consumer names deterministic; follow `service` + feature identifiers.
- Follow repo patterns for logging events via `runtime/events` and persisting snapshots via JetStream → Datastar pipeline ([see the event-driven orchestrator task](../../tasks/014-event-driven-orchestrator/)).

### 2.5 Local vs Test Data Separation
- Local/dev data sits under `.data/nats-*`; tests redirect to `.data-test/nats-*`. Never hardcode filesystem paths—call the `pkg/config` helpers.
- Clean test artifacts after runs if they leak into `.data-test/`; they should be disposable.

### 2.6 Observability Stack
- Embedded NATS exposes the HTTP monitoring endpoint on `GetNATSClusterPortsForNode(index).http`.
- Attach `nats-surveyor` for richer metrics: point it at the monitoring URL and persist dashboards per tenant.
- Structured logs come via `pkg/nats/logger.go`, ensuring `go run . runtime watch` streams JetStream events alongside other service logs.

### 2.7 Operational Checklist For Agents
- [ ] Ensure JetStream storage directory exists and is writable before boot.
- [ ] Run `nsc` via `pkg/nats/auth`; never hand-roll JWT paths.
- [ ] Verify streams/consumers with the `nats` CLI before wiring new services.
- [ ] Mirror or source tenant streams when promoting from dev leaf → production cluster.
- [ ] Keep snapshots/backups scheduled if the stream holds authoritative data.
- [ ] Monitor pending counts and redeliveries; adjust `ack_wait` and `max_inflight` before scaling writers.

---

## Additional Notes
- Update this guide whenever we change default streams, add JetStream-backed services, or adjust authentication flows.
- Cross-reference [Process Compose](../process-compose/) when supervising NATS within multi-process projects.
