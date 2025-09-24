# Infra System Specification

## 1. Purpose & Audience
Infra delivers a complete production-grade infrastructure stack as a single Go import and CLI. This specification explains the system at a high level for architects, operators, and contributors so they can reason about capabilities, boundaries, and extension points without diving into code.

## 2. System Goals
- **Everything-as-Go-Import**: provide web, data, messaging, deployment, and supervision primitives as Go packages.
- **Single-command lifecycle**: `go run .` boots the full stack locally; `go run . workflows deploy` pushes the same stack to Fly.io.
- **Deterministic environments**: embed configuration defaults in `pkg/config` so development, CI, and production behave consistently.
- **Operational observability**: expose health, metrics, and logs through the supervised processes and CLI.
- **Extensibility**: allow new services, binaries, and workflows to be composed through packages under `pkg/` and the CLI surface in `pkg/cmd`.

## 3. Runtime Overview
Infra runs as a supervised constellation of long-lived processes plus supporting CLIs. By default `cmd.Execute()` coordinates everything. All components share a common configuration layer and write to environment-aware data directories.

```
go run .
└── goreman supervisor
    ├── webapp (http://localhost:1337)
    ├── nats (4222)
    ├── pocketbase (8090)
    ├── bento (4195)
    ├── deck api (8888)
    ├── caddy (80/443 proxy)
    ├── xtemplate renderer
    └── optional: mox, workflows, other pkg services
```

## 4. Core Subsystems
- **Process Supervision (`pkg/goreman`)**: registers and manages child processes, ensuring deterministic start/stop and graceful shutdown (`go run . shutdown`).
- **Configuration (`pkg/config`)**: single source for ports, paths, environment detection, and binary constants. All packages must read defaults through this API.
- **Service Packages**:
  - `pkg/webapp`: dashboard UI served on 1337 with Datastar-based front-end.
  - `pkg/nats`: embedded JetStream messaging.
  - `pkg/pocketbase`: application database and admin UI, storing data under `.data/` or `.data-test/`.
  - `pkg/bento`: stream processing and automation engine.
  - `pkg/deck`: visualization API & watcher services for go-zero projects.
  - `pkg/caddy`: reverse proxy + HTTPS termination, also fronting Fly deployments.
  - `pkg/mox`: optional mail server for local testing.
- **CLI Surface (`pkg/cmd`)**: unified entry point for service control, deployment, dependency installation, status checks, and developer tooling.
- **Binary Management (`pkg/dep`)**: installs external binaries declared in `pkg/dep/dep.json`; integrates with CLI (`go run . tools dep ...`).

## 5. Data & Configuration Management
- Configuration flows through `pkg/config`. New features must extend this package instead of hardcoding paths or ports.
- Persistent data lives in `.data/`; ephemeral test data lives in `.data-test/` and is gated by `config.IsTestEnvironment()`.
- Secrets and deployment credentials are supplied via environment variables or Fly.io secrets; `pkg/config` resolves them for consumers.

## 6. Deployment & Environments
- **Local Development**: `go run . runtime up --env development` enables developer-friendly defaults while still using the same supervised components.
- **Testing**: `go test ./...` and `go run . dev api-check` confirm API stability and component compatibility.
- **Containers**: `go run . runtime container` builds OCI images with Ko; no Dockerfile required.
- **Production**: `go run . workflows deploy` targets Fly.io using embedded configs in `pkg/fly` and `fly.toml`.
- **Future scale**: `go run . tools flyctl scale` and related commands manage horizontal scaling.

## 7. Observability & Operations
- **Status**: `go run . runtime status` queries supervised processes for health.
- **Metrics & Logs**: each managed service exposes endpoints (`/metrics`, `/logs` views) surfaced through the web dashboard and CLI helpers in `pkg/status` and `pkg/debug`.
- **Docs**: `docs/technical/` hosts deep-dives per component; `docs/business/` captures product context.
- **Automation Hooks**: githooks run `go run .` and linting to guard quality; CI should mirror the same commands.

## 8. Extension Patterns
- New services register via `goreman.RegisterAndStart` and must supply configuration using `pkg/config` getters.
- Additional binaries follow the five-step workflow defined in `CLAUDE.md` for `pkg/dep` updates.
- CLI extensions live under `pkg/cmd/<feature>` and are exposed through `cmd.Execute()`.
- Documentation for new capabilities belongs in `docs/technical/<topic>.md` with links from `docs/technical/README.md`.

## 9. Open Questions & Roadmap Items
- Align NATS orchestrator with both local and Fly workflows so supervision remains identical across environments (`TODO.md`).
- Expand observability coverage for Bento and Deck when multi-node deployments are introduced.
- Formalize security hardening guidelines for embedded PocketBase and Caddy before broader production rollout.

## 10. Related References
- `README.md` — quick start and command cheatsheet.
- `docs/technical/README.md` — core philosophy and architecture diagram.
- `docs/technical/CLI.md` — detailed CLI contract.
- `pkg/config` — configuration API definitions.
- `main.go` — entrypoint delegating to `pkg/cmd`.
