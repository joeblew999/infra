---
title: "Litestream Agent Guide"
summary: "Replication, restore, and operations for Litestream-managed SQLite."
draft: false
---

## Sources (do not remove)
- **Official docs**
  - https://litestream.io
  - https://litestream.io/reference/config
  - https://litestream.io/reference/cli
- **Source & tooling**
  - https://github.com/benbjohnson/litestream
  - https://github.com/benbjohnson/litestream/issues
- **Community recipes**
  - https://github.com/benbjohnson/litestream/tree/master/examples
  - https://fly.io/docs/litestream/getting-started/

## Part 1 – Litestream Fundamentals

### 1.1 Core Concepts
- **Purpose**: Litestream continuously replicates SQLite databases to object storage (S3-compatible). It provides point-in-time recovery and live replicas.
- **Deployment modes**:
  - *Sidecar process* (`litestream replicate`) tails the SQLite WAL while your app runs the native SQLite library.
  - *SQLite VFS* integration loads `liblitestream` so the database itself ships WAL updates; enables in-process failover logic and read replicas.
- **Replica types**: `s3`, `file`, `gcs`, `abs`, and `sqlite` (for on-disk read replicas). Configure storage credentials, retention, sync intervals per replica.
- **Snapshots vs WAL**: Litestream ships periodic snapshots and streams WAL deltas between snapshots so restores can replay to a specific timestamp.
- **Read replicas**: Latest releases maintain downstream SQLite files, either via the VFS or standalone `litestream replicate --exec` workers.

### 1.2 Configuration Files
```yaml
dbs:
  - path: /data/pb_data.db
    access-key-id: ${LITESTREAM_ACCESS_KEY_ID}
    secret-access-key: ${LITESTREAM_SECRET_ACCESS_KEY}
    replicas:
      - type: s3
        bucket: my-tenant-backups
        path: pocketbase/pb_data.db
        endpoint: https://s3.us-east-1.amazonaws.com
        retention: 168h
    read-replicas:
      - name: analytics-reporting
        type: sqlite
        path: /var/replicas/pb_reporting.db
```
- Each `dbs` entry describes one SQLite file and its replicas. Credentials can live at the DB level or individual replica level.
- `read-replicas` keep warm copies nearby for low-latency reads. On VFS deployments, Litestream automatically routes read queries there when the primary is unavailable.
- Optional settings: `sync-interval`, `snapshot-interval`, `min-wal-checkpoint`, `concurrency`, and `retention` per replica.
- VFS deployments add a `vfs:` block (e.g., `vfs: { shared: true }`) or use connection strings such as `file:/data/pb_data.db?vfs=litestream`.

### 1.3 Command Basics
- `litestream replicate -config litestream.yml` – primary replication process for sidecar deployments.
- `litestream restore -if-replica-exists -config litestream.yml /path/to/db` – idempotent restore step.
- `litestream run -config litestream.yml -- sqlite-app --flags` – launches your app with the Litestream VFS preloaded (latest releases).
- `litestream snapshots list <replica-url>` – inspect stored snapshots.
- `litestream wal list <replica-url>` – view WAL segments for point-in-time recovery.
- `litestream read-replica sync <name>` – force synchronization of a configured read replica (CLI now exposes per-replica commands).

### 1.4 Health & Monitoring
- Litestream logs replication status to stdout. Watch for `stream resumed`, `sync complete`, and error messages.
- Supervisors should rate-limit restarts; repeated non-zero exits often indicate credential or connectivity failures.
- VFS deployments expose replication state via pragma queries (`PRAGMA litestream.checkpoint;`). Surface those in health checks if you embed the VFS.
- For metrics, wrap Litestream with your own exporter or rely on object-store audit logs to confirm writes.

### 1.5 Restore & Disaster Recovery
- Restores pull the most recent snapshot plus WAL segments up to an optional `-timestamp`.
- Typical workflow:
  1. Ensure the target DB path is absent (or use `-overwrite` carefully).
  2. `litestream restore -if-replica-exists -config litestream.yml /data/pb_data.db`.
  3. Start your application (sidecar mode) or reopen via VFS connection string.
- For VFS deployments, Litestream can automatically promote a read replica if the primary path is unavailable—test failover periodically.

### 1.6 Multi-Replica & Multi-Region
- Configure multiple replicas for redundancy (e.g., S3 primary + secondary in another region; local `sqlite` replica for reads).
- Align `snapshot-interval` and `retention` to RPO/RTO requirements; longer retention increases storage cost but improves recovery options.
- Use new `lag-alert` settings (CLI flags) to audit replication drift; failover read replicas that exceed acceptable lag.

### 1.7 Upgrading & Compatibility
- Latest builds require SQLite 3.40+ when using VFS mode. Ensure target environments ship the matching dynamic library.
- Litestream remains backward compatible with previous `litestream.yml`; new fields (`read-replicas`, `vfs`) are optional.
- Re-run acceptance tests after upgrading: replicate to a test bucket, restore, validate schema, and exercise read replica queries.

---

## Part 2 – Litestream Inside infra

### 2.1 Default Layout
- `/app/litestream.yml` lives alongside the tenant’s `process-compose.yml` and `Caddyfile`.
- SQLite data (`pb_data.db`) resides under `/data/`, matching PocketBase expectations.
- Environment variables such as `LITESTREAM_ACCESS_KEY_ID`, `LITESTREAM_SECRET_ACCESS_KEY`, and tenant-specific bucket names are injected by runtime tooling (Fly secrets, local `.env`, Process Compose env blocks).

### 2.2 Restore → Replicate → Serve Flow
1. **Restore**: `restore_db` process runs `litestream restore -if-replica-exists` to hydrate `/data/pb_data.db` before PocketBase starts.
2. **Replicate**: `litestream` process tails WAL changes via `litestream replicate -config /app/litestream.yml`.
3. **Serve**: PocketBase starts only after the restore process reports success. Litestream continues streaming in the background or, in future VFS mode, runs in-process through PocketBase’s SQLite connection.

### 2.3 Configuration Patterns
- Template `litestream.yml` per tenant with `${TENANT}` placeholders for bucket paths and optional read replica definitions.
- Default retention: `168h` (7 days). Adjust per tenant SLA.
- When introducing read replicas, place them on fast local disks (e.g., `/var/replicas/<tenant>.db`) and document routing logic in the application layer.
- Ensure tenant environments can reach the object store (Fly Machines networking, MinIO in dev, etc.).

### 2.4 Local Development Setup
- Provide a MinIO container or S3-compatible endpoint to test replication end-to-end. Example env:
  - `LITESTREAM_ENDPOINT=http://127.0.0.1:9000`
  - `LITESTREAM_BUCKET=dev-tenant`
  - `AWS_ACCESS_KEY_ID` / `AWS_SECRET_ACCESS_KEY`
- To test VFS mode locally: build the latest Litestream, load the shared library (`LD_PRELOAD=liblitestream.so`) or use `litestream run -- sqlite3 app.db`.
- Clean up `.data-test` directories after tests; they should not persist between runs.

### 2.5 Monitoring & Alerts
- Pipe Litestream stdout into structured logs (Process Compose or goreman). Search for `error` or `lag` messages.
- Object store dashboards should alert on absent uploads or failed deletions when retention expires.
- When using read replicas, track lag via `litestream read-replica status` and expose it to Prometheus.

### 2.6 Operational Checklist For Agents
- [ ] Confirm credentials, bucket, and network reachability before booting Litestream.
- [ ] Run the restore step during deployments; never start PocketBase without a restored DB.
- [ ] Verify replication status after deploy (`process-compose logs litestream --follow`).
- [ ] Exercise read replica failover tests when enabling the feature; document promotion steps.
- [ ] Keep the `litestream` binary updated in `dep.json` and rerun `go run . tools dep list` after bumps.
- [ ] Align retention and replica settings with tenant contracts; update configs when SLAs change.

### 2.7 Roadmap Notes (VFS Adoption)
- Evaluate embedding Litestream VFS inside PocketBase once we confirm compatibility with the app’s SQLite driver.
- Plan migration strategy: dual-run sidecar + VFS, compare replication lag, then flip to VFS-only if reliable.
- Update Process Compose definitions when we remove the external `litestream` process (replace with env vars/flags on PocketBase).

---

## Additional Notes
- Coordinate Litestream upgrades with PocketBase schema migrations—new schema releases may require restore testing.
- Combine with NATS JetStream (see the [NATS guide](../nats-jetstream/)) if you need event notifications on successful snapshots or read-replica lag.
- Update this playbook when new workflows (Process Compose revisions, Fly secrets rotation, VFS rollout) change how Litestream is managed.
