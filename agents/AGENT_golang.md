# Repository Guidelines (Go Agent)

## Project Structure
- `main.go` wires the CLI; runtime orchestration lives under `pkg/service/runtime`, `pkg/status`, and `pkg/goreman`.
- Domain packages (bento, deck, pocketbase, etc.) expose adapters that the runtime consumes. Keep logic in those packages; the runtime layer should only orchestrate.
- Reverse proxy support sits in `pkg/caddy`; the runtime regenerates and reloads Caddy as services publish routes.
- Configuration and binaries are centralized in `pkg/config` and `pkg/dep`; extend those instead of hardcoding paths.

## CLI Namespaces
- `go run . runtime up|down|status|list|watch|container`
  - `runtime up` always shuts down any previous supervisor, regenerates the Caddyfile, and boots the core stack (Caddy, web, NATS, PocketBase, etc.).
  - `runtime watch` streams live service events; use `--service`, `--types`, `--json` for filtering.
- `go run . workflows ...` – build/deploy pipelines (`deploy`, `status`, `build`, etc.).
- `go run . tools ...` – wrappers for managed binaries (flyctl, deck, gozero, dep, ko, etc.).
- `go run . dev ...` – developer diagnostics such as `dev api-check` (Go API compatibility).

## Build, Test, and Local Development
- Regular build: `go build -o infra .`
- Runtime smoke test: `go run . runtime up` then `go run . runtime status`; shut down with `go run . runtime down`.
- Container workflow: `go run . runtime container` (uses ko + Docker).
- Tests:
  - `go test ./...` before every push. The NATS integration tests now bind to an ephemeral port when 4222 is unavailable, so you no longer need to free the port manually.
  - Capture coverage with `go test ./pkg/... -cover` for service-heavy changes.
  - Run specific suites on demand: `go test ./pkg/service/runtime`, `go test ./pkg/status`, `go test ./pkg/caddy`, `go test ./pkg/nats`.

## Coding Style & Tooling
- Target Go 1.25.1. Always run `gofmt`/`goimports`; hooks expect both.
- Keep packages lower-case and focused; export the minimal API surface.
- Use `pkg/log` for structured logging instead of ad-hoc prints.
- Run `go vet ./...` when adding packages or touching initialization paths.

## Runtime Patterns
- Each service adapter owns its start/ensure logic and Caddy routes. Add new services by exposing an adapter, not by editing the orchestrator directly.
- Caddy is part of the core runtime (required) and is reloaded automatically when services register routes.
- The status subsystem (`pkg/service/state` + `pkg/status`) is event-sourced; publish lifecycle events to keep CLI and web dashboards in sync.

## Commit & PR Expectations
- Use conventional commit prefixes (`feat:`, `fix:`, `chore:`…).
- Keep commits focused, messages imperative, and include proof of testing.
- In PRs, highlight configuration changes (`.env`, Fly secrets, Terraform) and attach screenshots for web-facing updates.
