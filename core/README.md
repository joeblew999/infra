# Core Platform

This repository contains the deterministic core platform (Task 015). All new
runtime modules and shared packages live here.

## Goals
- Deterministic startup pipeline around JetStream, PocketBase, and Caddy.
- Shared packages that downstream services can import directly.
- Real integration tests—CLI, TUI, web, and harness code exercise the actual
  orchestrator components.

See `pkg/runtime/README.md` for the detailed module layout. For subsystem
playbooks, consult [docs/README.md](docs/README.md) which indexes the
canonical guides.

## Quick Commands
- Start stack: `go run ./cmd/core stack up`
- Stop stack: `go run ./cmd/core stack down`
- Deploy to Fly: `go run ./tooling workflow deploy --profile fly --app <fly-app> --org <fly-org> --repo registry.fly.io/<fly-app>`


## Repository Layout
This workspace hosts three cooperating Go modules managed via `../go.work`:
- `core/` (this directory) — runtime CLI, shared packages, and binaries under `cmd/` (e.g. `go run ./cmd/core`).
- `controller/` — desired-state API and reconciler service. Run it from that directory with `GOWORK=off go run .`.
- `tooling/` — release and Fly deployment CLI (`go run ./tooling`).

Wrapper binaries such as `cmd/processcompose` stay in `cmd/` so each executable can import only the dependencies it needs without pulling deploy-time tooling into the runtime build.

## Runtime
- Start: `go run ./cmd/core stack up`
- UI: `go run ./cmd/core tui` or `go run ./cmd/core web`
- Stop: `go run ./cmd/core stack down`

Full command reference: [docs/runtime.md](docs/runtime.md).

## Controller Quickstart
```sh
# From controller/ (separate module).
cd controller

# Run the controller API with an explicit desired-state spec and bind address.
GOWORK=off go run . --spec spec.yaml --addr 127.0.0.1:4400 \
  --cloudflare-token-file ~/.config/core/cloudflare/api_token

# Persist state to disk before exiting (Ctrl+C to stop); the spec file is updated on shutdown.
```

The controller exposes `/v1/services` (read) and `/v1/services/update` (PATCH) endpoints. Set
`CONTROLLER_ADDR` or pass `--controller` to `core scale` commands so the runtime CLI can fetch and
apply desired state updates against the running service. Cloudflare credentials can be supplied via
`--cloudflare-token`, `--cloudflare-token-file`, or environment variables
`CLOUDFLARE_API_TOKEN` / `CLOUDFLARE_API_TOKEN_FILE`.

## Tooling
- `go run ./tooling workflow deploy --profile fly --app <fly-app> --org <fly-org> --repo registry.fly.io/<fly-app>`

For the complete Fly + Cloudflare deployment playbook—including setup, auth,
and integration patterns—see the canonical guide at [docs/tooling.md](docs/tooling.md).

## Config Templates
Templates for `.ko.yaml` and `fly.toml` live in `config/templates`. The workflow
commands regenerate them automatically; run `go run ./tooling config init`
manually only if you need custom overrides.

## Release Checklist
- ✅ `go test ./...`
- ✅ `git status --short`
- 🚀 `go run ./tooling workflow deploy --profile fly --app <fly-app> --org <org-slug> --repo <registry>`

If ko still reports a dirty state, double-check for generated files (e.g. `core/.ko.yaml`, `core/fly.toml`) and either commit them or pass `--force` to regenerate after cleaning the tree.

## tooling

this does deployment and setsup Cloudflare and Fly with everything. 
