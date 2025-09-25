# Task 015 — Deterministic Core Platform (JetStream Backbone)
summary: Build a deterministic core platform with mirrored shared modules so every runtime package has a 1:1 shared counterpart reusable by example services.

## Problem / Solution
**Problems**
- Legacy packages (`pkg/config`, `pkg/dep`, `pkg/log`, `pkg/nats`, `pkg/caddy`) own critical behaviour, so we effectively have two layers (old supervisor helpers and the new core); every service wires the old layer differently and deterministic startup never emerges.
- Process supervision, lifecycle events, and logging are inconsistent across services, CLIs, and UIs.
- CLI/TUI/Web layers duplicate reducers and components, so parity constantly drifts.
- Deterministic testing is hard because harness utilities are scattered and services cannot reuse a common boot pipeline.

**Solution**
- Replace the legacy helpers (including the existing NATS/Caddy wrappers) with mirrored shared modules under `core/pkg/shared/*` and have runtime code in `core/pkg/runtime/*` compose them.
- Move JetStream lifecycle, metrics, and supervision into the new platform (`core/pkg/runtime/services/bus`, `core/pkg/runtime/process`) and emit consistent events.
- Build shared CLI/TUI/Web primitives so core shells and services render the same state/reducers.
- Provide a deterministic harness plus example services that prove the layering before retiring the legacy supervisor.

## Why Core vs Shared?
- **Predictability**: Runtime packages depend only on their shared counterparts, keeping dependency direction clear and simplifying reviews.
- **Service reuse**: Services import shared modules directly, guaranteeing they match core behaviour without duplicating logic.
- **Deterministic testing**: Shared harness/testutil modules let both runtime and services replay event logs and assert the same projections.
- **Legacy isolation**: Legacy packages stay reference-only; all new development lives inside `core/pkg`.

## Architecture Overview
- `core/pkg/runtime/<module>` contains runtime glue; `core/pkg/shared/<module>` holds the reusable implementation (config, dep, log, events, state, process, controller helpers, bus primitives, CLI/TUI/Web components, harness/test utilities).
- Services under `core/pkg/runtime/services/demo/*` mirror the shared structure (`config/`, `dep/`, `log/`, `cli/`, `ui/`, `process/`), delegating to shared modules only.
- The legacy `pkg/nats` and `pkg/caddy` packages are reference-only; they exist because we currently have dual layers (legacy vs core), and the goal is to migrate completely to the new core. `core/pkg/runtime/services/bus` re-implements embedded JetStream startup, clustering, and metrics using shared config/log primitives, and exposes configuration via `core/pkg/shared/bus`.

### Core vs Shared Principles
1. **Mirror every module**: For each runtime package `core/pkg/runtime/<module>` maintain a matching `core/pkg/shared/<module>` that exports reusable types, builders, and helpers.
2. **Runtime depends on shared only**: Runtime packages may only import shared modules (plus sibling runtime modules where unavoidable); legacy helpers stay unimported.
3. **Shared defines contracts**: Event envelopes, state projections, CLI flags, TUI/Web components, and harness APIs live in `core/pkg/shared/*`, guaranteeing a single source of truth.
4. **Services prove reuse**: Example services under `core/pkg/runtime/services/demo/*` mirror the shared structure (`config`, `dep`, `log`, `cli`, `ui`, `process`) while delegating to `core/pkg/shared/*`.
5. **No legacy wrappers**: Shared modules reimplement behaviour; they only treat existing packages (`pkg/config`, `pkg/dep`, `pkg/log`, etc.) as references during the rewrite and never import them.

### Package Mapping
| Runtime (`core/pkg/runtime/...`) | Shared (`core/pkg/shared/...`) | Responsibility |
| --- | --- | --- |
| config | shared/config | Config surface (paths, env defaults)
| dep | shared/dep | Binary install helpers (`.dep/` management)
| log | shared/log | Structured logging fields/adapters
| events | shared/events | Event envelopes, subjects, command DTOs
| state | shared/state | Reducers and state projections
| process | shared/process | Restart policies, telemetry helpers
| controller | shared/controller | Controller DTOs/hooks
| bus | shared/bus | JetStream bootstrap, connection helpers, metrics
| cli | shared/cli | Cobra command builders, flags
| ui/web | shared/ui/web | Datastar fragments/layouts sourced from shared templates
| ui/tui | shared/ui/tui | Bubble Tea primitives/layout helpers that reuse the same templates/state reducers
| harness | shared/harness/testutil | In-memory start/stop + fixtures
| services/demo/* | shared/* | Service wrappers around shared modules
| ko tasks | shared/pipeline/ko | Build/push container images with ko (env + registry helpers)


### Plan for Caddy
- Replace the current dual Caddy layers (legacy supervisor and new core) by treating Caddy as part of the core platform. All configuration lives in `core/pkg/shared/caddy`, which provides global, regional, and node-level templates plus health checks and deploy helpers.
- Build a custom Caddy binary via standard Go builds (no `xcaddy`) driven by `core/pkg/runtime/services/caddy` commands; the binary is composed from the module registry declared in `core/pkg/shared/caddy/modules`.
- Ensure every node boots the managed Caddy instance as part of the deterministic startup (config → bus → caddy → controller → services) so there is exactly one startup mode.
- Persist issued/renewed certificates in JetStream via `core/pkg/shared/bus`, emitting lifecycle events so other nodes can replicate material without shared filesystems.
- Provide shared config fragments so services can opt into HTTP exposure, reuse reverse-proxy snippets, and publish health endpoints consistently across CLI/TUI/Web.
- Document global (edge) vs regional (per-datacenter) deployment patterns that ride on the same binary and configuration primitives while still being orchestrated in Go.

### Candidate Caddy Modules
We maintain a module registry (`core/pkg/shared/caddy/modules`) listing the plugins to include when building the custom Caddy binary. Useful starting modules:
- `http.reverse_proxy`, `http.handlers.rewrite`, `http.handlers.redirect` (core primitives)
- ACME helpers (`http.handlers.acme_server`) and DNS challenge providers (`github.com/caddy-dns/cloudflare`, `github.com/caddy-dns/route53`, etc.)
- `github.com/greenpau/caddy-auth-portal`, `github.com/greenpau/caddy-auth-jwt` for richer auth flows
- `github.com/mholt/caddy-ratelimit` for rate limiting
- `github.com/aksdb/caddy-crowdsec-bouncer` for CrowdSec integration
- `github.com/SchumacherFM/caddy-prometheus` for Prometheus metrics export
- `github.com/caddyserver/forwardproxy` if we need forward-proxy capabilities
The registry must be mirrored in shared config so `core caddy build` (standard Go build, no `xcaddy`) can compose these modules consistently across environments.


### Pipeline Platform (replacement for `pkg/workflows`)
- Replace the legacy `pkg/workflows` package with a new mirrored design: runtime lives in `core/pkg/runtime/pipeline`, reusable primitives in `core/pkg/shared/pipeline`.
- Workflows (“pipelines”) stay in Go so they’re invokable via `go run . pipeline ...` and can compose shared steps directly.
- Shared pipeline steps wrap core functionality (`config`, `dep`, `log`, `bus`, `process`, `caddy`, `artifact`, `deploy`, `notify`, `ko`, `fly`). Each step is idempotent and emits events/logs through shared modules.
- Runtime pipeline coordinator orchestrates runs, publishes statuses on the bus, and exposes CLI/TUI/Web commands for list/run/status/cancel.
- Provide a DSL/registry for defining pipelines in Go; optional exports (JSON/YAML) for the UI to render pipeline graphs.
- Integrate with the harness so pipelines can run in-memory for testing and developers can execute them locally with shared fixtures.
- Example pipelines: build core binaries, build the custom Caddy binary, publish artifacts to storage, deploy to environments (including ko and Fly builds), and roll back deterministically.


### Service Binary Specification
To make third-party binaries plug-and-play, the new platform reads JSON specs describing each service. Every `core/pkg/runtime/services/demo/*` package must provide (and the runtime must honour) the following fields:

```jsonc
{
  "id": "redis",
  "displayName": "Redis",
  "description": "In-memory cache used by the demo stack.",
  "icon": "🧠",
  "required": false,

  "dep": {
    "name": "redis",
    "version": "7.2.4",
    "source": "github-release",
    "asset": "redis-7.2.4-linux-amd64.tar.gz"
  },

  "process": {
    "command": "{shared.dep.bin}/redis-server",
    "args": ["{shared.config.dir}/redis.conf"],
    "env": {
      "REDIS_PORT": "{ports.primary}",
      "REDIS_DATA": "{shared.config.dataDir}"
    },
    "workingDir": "{shared.config.dataDir}",
    "restartPolicy": "on-failure",
    "ensure": [
      {"type": "template", "template": "redis.conf.tmpl", "target": "{shared.config.dir}/redis.conf"},
      {"type": "mkdir", "path": "{shared.config.dataDir}"}
    ]
  },

  "ports": {
    "primary": {"port": 6379, "protocol": "tcp"},
    "additional": [
      {"name": "metrics", "port": 9121, "protocol": "http"}
    ]
  },

  "logging": {
    "provider": "shared/log",
    "stdout": "{shared.log.dir}/redis.log",
    "format": "json",
    "retention": "7d"
  },

  "health": {
    "type": "http",
    "interval": "5s",
    "timeout": "3s",
    "endpoint": "http://127.0.0.1:9121/health",
    "expectedStatus": [200]
  },

  "caddy": {
    "expose": true,
    "routes": [
      {
        "match": ["host(`redis.internal`)"] ,
        "handle": [
          {
            "handler": "reverse_proxy",
            "upstreams": ["127.0.0.1:6379"],
            "health_check": {"path": "/health", "interval": "5s", "timeout": "2s"}
          }
        ]
      }
    ]
  },

  "dependencies": ["core.bus"],
  "profiles": ["default", "demo"]
}
```

**Key fields**
- `dep`: mirrors `core/pkg/shared/dep` entries (name/version/source/asset).
- `process`: command, args, env, working dir, restart policy, plus `ensure` actions (template rendering, mkdir, etc.).
- `ports`: primary + additional named ports (used for ownership, monitoring, Caddy exposure).
- `logging`: instructs `core/pkg/shared/log` where/how to emit logs.
- `health`: liveness/readiness probe definition for the process runner and Caddy.
- `caddy`: optional HTTP exposure, referencing shared Caddy fragments.
- `dependencies`: other services that must start first (e.g., `core.bus`).
- `profiles`: which runtime profiles auto-enable this service.

Runtime code in `core/pkg/runtime/process`, `core/pkg/shared/config`, `core/pkg/shared/dep`, and `core/pkg/shared/caddy` must be able to consume this schema. Example services must load the JSON at startup and register themselves with the controller.

### Plan for Legacy `pkg/nats`
- Treat `pkg/nats` as reference-only (we currently have two NATS layers: legacy and core); no new code imports the legacy package directly.
- Re-create the embedded JetStream lifecycle inside `core/pkg/runtime/services/bus` (startup, shutdown, leaf-node bridging back into a larger cluster, metrics, certificate storage for Caddy) using shared config/log primitives.
- Default the embedded broker to an in-memory JetStream store with deterministic clean-up for harness/tests while still supporting disk-backed stores for real nodes via shared config.
- Expose bus configuration and connection helpers via `core/pkg/shared/bus` (and supporting config/events modules) so runtime packages consume a stable API.
- Identify useful helpers (auth setup, gateways) from `pkg/nats` and either port them into shared modules or defer as follow-up work.
- Deprecate `pkg/nats` in documentation once `core/pkg/runtime/services/bus` covers all required behaviour.

## Phased Build Plan
- **Phase 1 (Week 1-2)**: Build shared config/dep/log/events/state/process/controller/bus/testutil modules with unit tests.
- **Phase 2 (Week 3)**: Add shared CLI/TUI/Web primitives and wire the runtime skeleton (`cmd/core/main.go`, controller stubs) to shared modules.
- **Phase 3 (Week 4)**: Implement `core/pkg/runtime/services/bus` (built on shared/bus primitives), finalise the event pipeline, and add replay tests.
- **Phase 4 (Week 5)**: Implement the process runner and demo services using shared modules; verify enable/disable flows.
- **Phase 5 (Week 6)**: Harden CLI/TUI/Web parity, metrics, external connectivity, and implement global/regional Caddy configs (`core/pkg/shared/caddy`) with bus-backed certificate storage.
- **Phase 6 (Week 7+)**: Deliver the harness, port tests, publish docs/migration notes, and remove or archive the legacy NATS/Caddy helpers once shared/runtime equivalents are proven.


---

## 1. Scope
Deliver a deterministic core platform that:
- Boots the same pipeline on every run (`cmd/core/main.go` → shared config → core bus → core caddy → controller → services → CLI/TUI/Web), with deterministic teardown on every node.
- Provides mirrored shared packages so runtime modules and services share the same primitives while legacy helpers remain reference-only.
- Publishes lifecycle transitions over JetStream (`core.*`) using JSON envelopes from shared modules; reducers rebuild state only from live streams/snapshots.
- Ships example services (`demo/alpha`, `demo/beta`) that depend exclusively on shared modules and are orchestrated end-to-end.
- Ensures CLI/TUI/Web shells share reducers, components, and state projections, maintaining parity.
- Emits Prometheus metrics and broker stats via the core bus.
- Operates without mocks: both shells and harness-driven tests talk to the live orchestrator components started by the deterministic pipeline.

---

## 2. Milestones
### M0 — Mirrored Shared Foundations
- Scaffold the `core/pkg/runtime/` tree with placeholder runtime packages alongside empty shared companions (`core/pkg/shared/<module>`).
- Implement shared config/dep/log/events/state/process/controller/bus/CLI/TUI/Web/harness/testutil modules with unit tests and documentation explaining runtime mirroring.
- Document the deterministic startup graph (shared config → bus placeholder → controller placeholder) in `cmd/core/main.go`.

- Implement minimal runtime packages (`core/pkg/runtime/config`, `core/pkg/runtime/dep`, `core/pkg/runtime/log`, etc.) that compose their shared counterparts without JetStream yet.
- Provide controller scaffolding (`core/pkg/runtime/controller`) consuming shared events/state types and exposing hooks for bus/process integration.
- Build stub CLI/TUI/Web shells using shared primitives to render startup configuration/state, and ensure the per-node Caddy process starts/stops with the skeleton.

- Implement `core/pkg/runtime/services/bus` (built on shared/bus primitives) to start embedded JetStream deterministically, expose publish/subscribe helpers, and surface Prometheus endpoints.
- Finalise shared event envelopes/subjects and integrate `core/pkg/runtime/state` reducers with shared state projections.
- Connect controller and shells to the live bus so command subjects and state updates flow through shared types end-to-end; the TUI/Web toggles must already dispatch enable/disable commands (even while the process runner is stubbed) to lock in parity and command semantics.
- Add replay tests proving deterministic rebuilds from the event log.

- Implement `core/pkg/runtime/process` (restart policies, backoff, port strategies) using shared log/config modules, exposing shared/process helpers.
- Implement `core/pkg/runtime/spec` and lifecycle event publishing, ensuring shared event constants/types are used everywhere.
- Build services (`core/pkg/runtime/services/demo/alpha`, `core/pkg/runtime/services/demo/beta`) with `config/dep/log/cli/ui/process` subpackages that wrap `core/pkg/shared/*` modules and ship binaries via shared/dep.
- Verify enable/disable flows, lifecycle event propagation, and logging parity across all shells.

- Finish runtime CLI/TUI/Web presenters (`core/pkg/runtime/cli`, `core/pkg/runtime/ui/*`) using shared primitives; publish a parity checklist covering toggles, restarts, logs, metrics.
- Harden metrics exposure via `core/pkg/runtime/services/bus` (event throughput, restart counts, lag) and document integration with external Surveyor.
- Provide configuration for external NATS connectivity through shared config while keeping the startup path deterministic.

- Deliver `core/pkg/runtime/harness` (mirrored by shared/harness) to spin up the full stack in-memory and expose helpers for CLI/UI/service testing.
- Port CLI tests and add UI smoke tests that run through the harness using shared state projections.
- Produce migration guidance and a retirement plan for the legacy supervisor, then publish final docs in `core/pkg/runtime/docs` once parity and determinism are signed off.


---

## 3. Design Decisions — Options & Recommendations
| Topic | Options Considered | Recommendation |
|-------|--------------------|----------------|
| Startup sequence | 1) Deterministic pipeline; 2) Flag permutations; 3) Env-specific branches | Single deterministic pipeline; variations flow through shared config without altering code path |
| Module mirroring | 1) Ad-hoc sharing; 2) Selective mirroring; 3) Full 1:1 mirroring | Maintain 1:1 `core/pkg/runtime/<module>` ↔ `core/pkg/shared/<module>` to simplify dependencies and reuse |
| Event model | 1) JSON envelope; 2) Protobuf; 3) CloudEvents | JSON envelopes in shared/events for transparency and easy replay |
| Process supervision | 1) Reuse goreman; 2) Thin wrappers; 3) New runner | New runner in `core/pkg/runtime/process` with shared/process helpers; goreman used only for ideas |
| UI layer | 1) Duplicate shells; 2) Partial sharing; 3) Shared primitives | Shared CLI/TUI/Web primitives in `core/pkg/shared` underpin all shells |
| Metrics exposure | 1) Embedded Surveyor; 2) External Surveyor; 3) Logs only | Expose metrics via `core/pkg/runtime/services/bus`, rely on external Surveyor |
| Example binaries | 1) Manual build; 2) Optional flag; 3) Auto ensure | Auto install via shared/dep during startup |
| Harness strategy | 1) Runtime-only tests; 2) Custom service harness; 3) Shared harness | Provide `core/pkg/shared/harness` mirrored by runtime harness to standardize testing |


---

## 4. Risks & Mitigations
| Risk | Mitigation |
|------|------------|
| Shared/core divergence | Enforce code reviews ensuring every runtime change updates its shared counterpart |
| Scope creep | Lock milestone exit criteria; review weekly |
| UI parity regressions | Share state projections, run parity smoke tests via harness |
| JetStream resource overhead | Default to memory store, random ports, clean teardown |
| Harness flakiness | Isolate `.data-test/`, provide robust teardown helpers |
| External cluster complexity | Treat external URLs as bus inputs without altering startup flow |


---

## 5. Deliverables Checklist
- [ ] Deterministic `cmd/core/main.go` booting the core stack via `core/pkg/runtime/services/bus`
- [ ] Full suite of mirrored shared modules (`core/pkg/shared/*`) used by all runtime packages and services
- [ ] Event schema package + helper library for publishing/subscribing
- [ ] Event-driven process runner with restart policies and port strategies
- [ ] Example services (`alpha`, `beta`) structured with `config/dep/log/cli/ui/process` packages that wrap shared modules and are orchestrated end-to-end
- [ ] CLI/TUI/Web shells sharing reducers and shared UI primitives
- [ ] Caddy integration: shared/global/regional configs (`core/pkg/shared/caddy`), module registry (`core/pkg/shared/caddy/modules`), go-based custom builds (`core caddy build`), bus-backed certificate storage, and runtime commands (`core/pkg/runtime/services/caddy`) built/tested without `xcaddy`; every node boots with the managed Caddy process.
- [ ] Harness package supporting Go + UI smoke tests
- [ ] Harness defaults to in-memory JetStream and real process supervision (no mocks) so tests exercise the exact orchestrator pipeline
- [ ] Prometheus metrics endpoints + integration guidance for external Surveyor
- [ ] Migration notes + retirement plan for the legacy supervisor
- [ ] Final documentation in `core/pkg/runtime/docs` after parity and determinism are validated


---

## 6. Follow-ups
- Hook core event streams into centralized observability dashboards (Grafana/Prometheus)
- Build tooling for event replay, restart tracing, and port-conflict diagnostics
- Explore multi-instance orchestration (leader election, distributed locking) once the deterministic single-node design is proven stable
