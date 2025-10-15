---
title: "Fly.io Agent Guide"
summary: "Deploying and operating Fly-powered services in this repo."
draft: false
---

Reference for deploying and operating Fly-powered services in this repo. Use it alongside `pkg/fly/README.md` for CLI specifics.

## Sources (do not remove)
- **Documentation**
  - https://fly.io/docs/
  - https://fly.io/docs/machines/overview/
- **CLI & Releases**
  - https://github.com/superfly/flyctl

---

## Part 1 – Fly.io Essentials

### Responsibilities
- Keep Fly app manifests (`fly.toml`) aligned with desired services, regions, and health checks.
- Maintain authenticated `flyctl` access for deployers (local and CI) using scoped API tokens.
- Run deployments through documented workflows so Machines remain healthy after rollout.

### Install & Authenticate
```bash
curl -L https://fly.io/install.sh | sh    # or `brew install flyctl`
flyctl version
flyctl auth login                         # browser-based
```
- Use `flyctl auth token` when generating long-lived CI credentials.
- Export `FLY_API_TOKEN` to enable non-interactive commands like `flyctl auth whoami`.

### Core Commands
- `flyctl status` – high-level app health (machines, checks, releases).
- `flyctl machines list` – enumerate Machines in the active app.
- `flyctl machines run` / `update` / `destroy` – lifecycle controls.
- `flyctl logs --tail` – live log tailing; add `--instance` to inspect a particular Machine.
- `flyctl releases list` – identify candidate versions for rollback.

### Configuration & Secrets
- Define regions, services, and health checks in `fly.toml`; `checks` enforce rollout success.
- Store secrets with `flyctl secrets set KEY=VALUE`; they become environment variables on Machines.
- Volumes require region pinning—confirm `primary_region` matches volume availability.
- Use `flyctl certs list` and `flyctl certs show` to monitor TLS; rerun `flyctl certs add` when onboarding new domains.

### Production Considerations
- Prefer rolling strategy (`strategy = "rolling"`) to avoid downtime; ensure each Machine passes health checks before removing the previous version.
- Use `flyctl releases revert <version>` for rapid rollback if checks fail.
- Pair deployments with monitoring: `flyctl apps metrics` for CPU/memory and external dashboards for latency.
- When databases rely on Litestream, validate replication after each deploy.

---

## Part 2 – Infra Repo Playbook

### Repo Map
| Area | Purpose | Notes |
| ---- | ------- | ----- |
| `fly.toml` | Default app manifest at repo root | Used by `go run . fly deploy`; update when services or regions change. |
| `core/fly.toml` | Manifest for the `core` workspace | Consume image tags produced by ko; keep metadata mirrored with root manifest. |
| `pkg/fly` | Go helpers + Cobra commands (`fly deploy`, `fly status`, etc.) | Provides repo-standard wrappers over `flyctl`; inspect `deploy.go` before editing flags. |
| `processcompose/` | Tenant demos and orchestration docs | Some examples include Fly integration notes—update together with guide changes. |

### Daily Workflow
1. **Install or Update flyctl**
   ```bash
   go run . tools dep install flyctl
   flyctl version
   ```
   - Keeps dev machines aligned with the pinned version.
   - Confirm auth via `flyctl auth whoami` (uses `FLY_API_TOKEN`).

2. **Check App Health**
   ```bash
   go run . fly status
   go run . fly machines list
   go run . fly logs -- --tail
   ```
   - `go run . fly ...` commands wrap `flyctl` with repo defaults.
   - `-- --tail` passes through flags to `flyctl` for log streaming.

3. **Deploy**
   ```bash
   go run . fly deploy -- --build-only
   go run . fly deploy
   ```
   - `-- --build-only` lets you smoke-test the build without release.
   - For full deploys ensure the image tag matches the latest ko publish (see [ko Build Guide](../ko/)).

4. **Manage Machines**
   ```bash
   flyctl machines run ./core/fly.toml --app <app>
   flyctl machines clone <id>
   flyctl machines stop <id>
   ```
   - Prefer `flyctl machine update --restart` over delete/create to retain volumes.
   - Wrap repetitive flows inside `pkg/fly` if they become standard tasks.

### Configuration & Secrets (Repo Context)
- Store tokens in 1Password and expose through `.envrc` or CI secrets.
- Document domain changes in `docs/infra/domains.md` after updating TLS or DNS.
- When volumes are required, check `processcompose` examples for mount conventions.

### Gotchas
- `go run . fly ...` expects to run from repo root; the wrapper sets working dir before calling `flyctl`.
- Mac M-series hosts may need Docker running; otherwise use remote builders with `flyctl deploy --build-target`.
- Machines API responses lag—wait for `flyctl machines status` to show `started` before chaining more operations.
- CI must authenticate with scoped deploy tokens via `flyctl auth token`; never reuse personal tokens.

### When To Update This Guide
- New Fly manifests or environments come online.
- `pkg/fly` wrapper gains new toggles or command groups.
- Fly introduces new primitives (Volumes V2, image checks) that the repo adopts.
