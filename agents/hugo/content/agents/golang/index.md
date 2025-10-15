---
title: "Go Service Agent Guide"
summary: "How we structure, build, and test Go services across the repo."
draft: false
---

Reference for engineers building and maintaining Go services across projects. Pair this with [Process Compose](../process-compose/) and other playbooks when orchestrating multi-service stacks.

## Sources (do not remove)
- **Documentation**
  - https://go.dev/doc/
  - https://go.dev/doc/effective_go
- **Style & Tooling**
  - https://google.github.io/styleguide/go/guide
  - https://pkg.go.dev/cmd/go
- **Concurrency & Runtime**
  - https://go.dev/doc/articles/race_detector
  - https://go.dev/blog/context

---

## Part 1 – Go Fundamentals

### Project Structure
- Keep packages small, lower-case, and focused. Expose only what downstream consumers need.
- Separate binaries from libraries: commands under `cmd/<name>` import reusable code from `pkg/...` or internal packages.
- Favor composition over global state. Share dependencies through constructors or context injection.

### Toolchain & Commands
```bash
go build ./...
go test ./...
go fmt ./...
go vet ./...
```
- `go build` validates all packages; pass `-o <binary>` to produce an executable.
- `go test` runs unit tests; use `-run`, `-bench`, and `-race` as needed.
- `go fmt` and `goimports` keep formatting and imports aligned.
- `go vet` uncovers common mistakes (printf mismatches, unreachable code).

### Modules & Dependency Hygiene
- Maintain a single module (`go.mod`) per repository unless isolation is required.
- Use `go get example.com/lib@version` to upgrade dependencies; commit the updated `go.sum`.
- Keep replace directives temporary—document why they exist and remove when upstream fixes land.

### Testing & Quality Gates
- Target deterministic tests. Stub time, random sources, and external calls.
- Add table-driven tests for pure logic and integration tests for side effects.
- Run `go test -race ./...` before shipping concurrency-heavy changes.
- Linting: integrate `staticcheck`, `golangci-lint`, or repo-specific linters when available.

### Code Review Expectations
- Use clear naming, short functions, and keep error handling explicit.
- Return wrapped errors via `fmt.Errorf("context: %w", err)` to aid debugging.
- Document exported symbols with `// Name ...` comments; align with `godoc` tone.

---

## Part 2 – Infra Repo Playbook

### Repo Layout
| Area | Purpose | Notes |
| ---- | ------- | ----- |
| `main.go` | Entry for the root CLI | Wires Cobra commands and global configuration. |
| `pkg/service/runtime` | Supervisor & orchestration logic | Drives Caddy, PocketBase, NATS, and supporting services. |
| `pkg/status`, `pkg/config` | Runtime state surfaces | Feed Datastar dashboards and CLI status output. |
| `pkg/caddy` | Reverse proxy integration | Generates routes and reloads Caddy from Go. |
| `pkg/dep`, `pkg/config` | Managed binaries & configuration registry | Extend these instead of baking absolute paths or env lookups. |

### CLI Namespaces
- `go run . runtime <subcommand>` – start, stop, inspect the core supervisor.
  - `runtime up` drains prior runs, regenerates configuration (Caddyfile, env), and boots managed services.
  - `runtime watch` follows structured events; filters via `--service`, `--types`, `--json`.
- `go run . workflows ...` – pipelines for build/deploy flows (deploy, status, build).
- `go run . tools ...` – curated wrappers around managed binaries (flyctl, nats, deck, ko, etc.).
- `go run . dev ...` – developer diagnostics (`dev api-check` guards API compatibility).

### Build & Test Workflow
- Primary build: `go build -o infra .` when distributing the CLI.
- Supervisor smoke test: `go run . runtime up`; verify via `go run . runtime status`; shut down using `go run . runtime down`.
- Container path: `go run . runtime container` (delegates to ko + Docker; see [ko Build Guide](../ko/)).
- Required tests before PRs:
  - `go test ./...`
  - `go test ./pkg/service/runtime`
  - `go test ./pkg/status`
  - `go test ./pkg/caddy`
  - `go test ./pkg/nats`
- Run `go run . api-check` whenever exported structs or API responses change.

### Runtime Patterns To Follow
- Service adapters own startup logic. Add new integrations by extending adapter packages rather than editing the orchestrator core.
- Publish lifecycle events through `pkg/runtime/events` so CLI, status dashboards, and Datastar streams stay synchronized.
- Use `pkg/log` for structured logging; avoid `fmt.Println` in production paths.
- Regenerate proxy routes through `pkg/caddy` helpers; Caddy reloads automatically when route definitions change.

### Tooling & Environment
- Target Go `1.25.1`; ensure local toolchains match CI.
- Hooks expect `gofmt` + `goimports`; install via `go install golang.org/x/tools/cmd/goimports@latest`.
- Monitor for race conditions with `go test -race ./pkg/...` when touching concurrency-heavy code.
- Keep managed dependencies defined in `pkg/dep/dep.json`; rerun `go run . tools dep list` after updates.

### Commit & PR Checklist
- [ ] Conventional commit prefix (`feat:`, `fix:`, `chore:`, ...).
- [ ] Focused commits with testing evidence in the message or description.
- [ ] Mention configuration changes (`fly.toml`, `.env`, Terraform) explicitly.
- [ ] Attach screenshots or terminal captures for user-facing updates (dashboards, CLI output).

Keep this guide aligned with new services, CLI namespaces, or runtime features as they land.
