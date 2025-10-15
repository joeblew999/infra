---
title: "Process Compose Agent Guide"
summary: "Authoring and operating Process Compose projects for infra tenants."
draft: false
---

## Sources (do not remove)
- **Official docs**
  - https://f1bonacc1.github.io/process-compose/
- **Source & releases**
  - https://github.com/F1bonacc1/process-compose/tree/main/src
  - https://github.com/F1bonacc1/process-compose/releases
- **Recipes & examples**
  - https://github.com/F1bonacc1/process-compose-recipes

## Part 1 – Process Compose Basics

### 1.1 What Process Compose Provides
- Single binary supervisor that manages any CLI program (services, jobs, daemons).
- Declarative YAML with per-process restart policies, readiness probes, and logging.
- Built-in REST API, CLI, and TUI for lifecycle management.
- Works on macOS, Linux, Windows; ships prebuilt binaries at the releases link above.

### 1.2 Project Layout Expectations
```
project/
  process-compose.yml   # required root definition
  .pc-history/          # (optional) created by the CLI/TUI for state
  scripts/              # (optional) helper scripts invoked via exec probes
```
- Define one project per YAML. You can include external files via `${ENV}` variables.
- Keep secrets and dynamic values outside the file; load them with `env_file` or exported environment variables.

### 1.3 Authoring `process-compose.yml`
```yaml
version: "0.5"

import:
  - other-project.yml

processes:
  service_name:
    command: ["binary", "--flag"]
    working_dir: ./relative/path
    environment:
      EXAMPLE: "${EXAMPLE}"
    availability: { restart: "on_failure" }
    readiness_probe:
      http_get: { url: "http://127.0.0.1:8080/health" }
    shutdown:
      signal: "SIGINT"
      timeout: "10s"

log:
  level: info
  timestamps: true
```
- Arrays keep command arguments safely quoted; inline shell via `sh; -lc; |` for multi-line scripts.
- `availability.restart` accepts `always`, `on_failure`, `never`, or numeric retry counts.
- Probes support `http_get`, `tcp`, and `exec`. Pair with `grace_period` to avoid premature restarts.
- For dependencies, add `depends_on` with `condition: process_healthy` to gate startup order.

### 1.4 Running and Managing Projects
```
process-compose up
process-compose down
process-compose status
process-compose logs <name> --follow
process-compose scale <name>=N
```
- Run commands from the directory containing `process-compose.yml`, or pass `-f path/to/file.yml`.
- Use `process-compose tui` for the ncurses dashboard when you want interactive control.

### 1.5 REST API and Remote Control
- REST server defaults to `http://127.0.0.1:8080`; disable with `--no-server` or `PC_NO_SERVER=1`.
- Endpoints cover listing, starting, stopping, scaling, and updating the active project (`project update`).
- Remote CLI examples:
  - `process-compose process list`
  - `process-compose process start <NAME>`
  - `process-compose project update -f process-compose.yml`
- Switch to UNIX socket mode via `process-compose -U /tmp/pc.sock` for local automation without TCP.

### 1.6 Cross-Platform Notes
- macOS/Linux can run shell pipelines directly in `command` blocks.
- Windows defaults to `cmd.exe`; set `COMPOSE_SHELL=powershell.exe` to use PowerShell syntax.
- When sharing projects across platforms, prefer portable binaries and guard OS-specific commands with conditional shell logic.

---

## Part 2 – Tenant Stack Playbook (Caddy + PocketBase + Litestream + NATS)

### 2.1 Goal and Assumptions
- One process-compose project per tenant.
- PocketBase stores data under `/data` and can be restored from an object store via Litestream.
- Caddy is the only component exposed publicly; everything else binds to loopback.
- Optional NATS sidecar follows the same patterns if the tenant requires messaging.

### 2.2 Directory Layout
```
/app/
  process-compose.yml
  Caddyfile
  litestream.yml
/data/
```
- Ensure `/data` persists (volume or bind mount) if you need state beyond a single boot.

### 2.3 Baseline `process-compose.yml`
```yaml
processes:
  restore_db:
    command:
      - sh; -lc; |
        set -e
        DB=/data/pb_data.db
        if [ -f "$DB" ]; then
          echo "[restore] DB exists, skipping"
        else
          echo "[restore] restoring from replica..."
          mkdir -p /data /data/public
          litestream restore -if-replica-exists -config /app/litestream.yml "$DB" || true
        fi
        echo "[restore] done"
    availability: { restart: "never" }
    readiness_probe:
      exec: { command: ["sh","-lc","test -f /data/pb_data.db || exit 0"] }

  litestream:
    depends_on:
      restore_db: { condition: process_completed }
    command:
      - sh; -lc; |
        until [ -f /data/pb_data.db ]; do echo "[ls] waiting for DB..."; sleep 0.5; done
        exec litestream replicate -config /app/litestream.yml
    availability: { restart: "always" }
    readiness_probe:
      exec: { command: ["sh","-lc","pgrep -f 'litestream replicate' >/dev/null"] }

  pocketbase:
    depends_on:
      restore_db: { condition: process_completed }
      litestream: { condition: process_healthy }
    command:
      - sh; -lc; |
        [ -d /data/public ] || mkdir -p /data/public
        exec pocketbase serve --dir /data --publicDir /data/public --http 127.0.0.1:30081
    environment:
      PB_PUBLIC_URL: "${PB_PUBLIC_URL}"
      PB_ENCRYPTION_KEY: "${PB_ENCRYPTION_KEY}"
      GOMEMLIMIT: "192MiB"
    availability: { restart: "on_failure" }
    readiness_probe:
      http_get: { url: "http://127.0.0.1:30081/api/health" }

  caddy:
    depends_on:
      pocketbase: { condition: process_healthy }
    command: ["caddy","run","--config","/app/Caddyfile","--adapter","caddyfile"]
    availability: { restart: "on_failure" }
    readiness_probe:
      http_get: { url: "http://127.0.0.1:2019/config/" }
```
- Add a `nats` process mirroring the patterns above when a tenant needs messaging; bind it to loopback and register routes through Caddy.
- Export `PB_PUBLIC_URL` and `PB_ENCRYPTION_KEY` in the environment, or load them via `env_file`.

### 2.4 Operational Workflow
1. `process-compose up -f /app/process-compose.yml` – boots the tenant stack.
2. Confirm readiness probes via `process-compose status` or the REST API.
3. Use `process-compose logs pocketbase --follow` when debugging service startup.
4. Trigger replication failover by stopping `litestream`; it restarts automatically because of `restart: always`.
5. Rotate secrets by updating the environment (or `env_file`) and running `process-compose project update`.
6. Tear down with `process-compose down` when the tenant stack is no longer needed.

### 2.5 Quick Checklist For Agents
- [ ] Restore step idempotent? (No error if DB already present.)
- [ ] PocketBase binds to loopback only; routes exposed through Caddy.
- [ ] Litestream config points at tenant-specific bucket/credentials.
- [ ] Readiness probes match actual health endpoints.
- [ ] Optional services follow the same dependency graph.

---

## Additional Notes
- Synchronize this playbook with the upstream Process Compose docs when new versions land.
- Reuse Part 1 as a template when introducing new orchestrations; keep Part 2 updated for tenant-specific nuances.
