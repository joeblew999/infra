# Task 014 — Event-Driven Orchestrator (JetStream Backbone)
summary: Design and implement a brand-new orchestrator core that uses embedded NATS JetStream as the sole event bus, driving service orchestration, observability, and testing via events. No legacy supervisor migration—greenfield build.

> **Working tree layout** (initial scaffold under `pkg/orchestrator/`):
> - `shared/` – shared utilities: `shared/config` (config façade), `shared/ui` (common templates/components), logging/test helpers reused by orchestrator + services
> - `bus/` – JetStream connection/bootstrap, publish/subscribe helpers, snapshot emission
> - `controller/` – orchestrator loop: consumes spec/events, drives processes, publishes lifecycle events
> - `process/` – process runner, restart policy, port management, log metadata events
> - `spec/` – spec structs, registry publishing `orchestrator.service.register`, validation rules
> - `state/` – reducers to rebuild snapshots, optional persisted checkpoints, query APIs
> - `harness/` – Go harness + CLI wrapper for deterministic tests, event replay utilities
> - `cli/` – Cobra commands (`up`, `list`, `status`, `watch`) wrapping orchestrator APIs
> - `ui/web/` – Datastar `/status` handlers + HTML templates (`ui/web/templates/`) consuming orchestrator events
> - `ui/tui/` – Bubble Tea + Ultraviolet TUI (`ui/tui/templates/`) mirroring the web layout using the same reducer patterns
> - `demo/alpha/`, `demo/beta/` – demo services with `main.go`, `cli/`, `web/` for integration exercises (compiled into `.dep` binaries)
> - `docs/` – orchestrator-specific Markdown docs (architecture, onboarding, demos)

---

## 1. Scope

Build a new orchestrator module that:
- Boots (or connects to) the embedded JetStream leaf for single-node operation, and can join the external NATS cluster when coordinating multiple servers.
- Publishes every lifecycle transition as a JetStream event (JSON envelope with versioning).
- Consumes service registration/config from events; no hard-coded spec slices.
- Provides a first-class harness for deterministic Go tests (in-memory JetStream) without mocks (tests use real demo services and in-memory JetStream).
- Hosts the orchestrator CLI + web status/SSE handlers within the same orchestrator package tree.
- Provides a Bubble Tea/Ultraviolet TUI and Datastar web UI with identical capabilities, both reusing the same event reducer patterns.
- Ships with two real demo services (e.g. `alpha` and `beta`) under `pkg/orchestrator/demo/alpha` and `pkg/orchestrator/demo/beta`, each mirroring production layout (`main.go`, `cli/`, `web/`), compiled into `.dep` demo binaries for realistic orchestration tests.
- Shared orchestrator utilities (`pkg/orchestrator/shared`) cover config façade, event constants, shared UI components, and helper functions; both orchestrator and service UIs reuse the same visual system.
- Includes Datastar HTML templates (`ui/web/templates/`) and Ultraviolet templates (`ui/tui/templates/`) that render identical layouts driven by shared reducers.
- Exposes Prometheus-friendly metrics endpoints and emits broker stats while leaving all NATS process management inside `pkg/nats` (or external tooling).
- Enforces a deterministic startup sequence: `cmd/orchestrator/main.go` performs the same orchestrator + demo boot routine on every run (no runtime toggles).
- Replaces the legacy supervisor entirely once CLI/UI and services move over.

### Existing Infra Package Map

**Direct imports (use as-is)**
- `pkg/config` — single source of truth for data directories, broker flags (`config.ShouldEnsureNATSCluster`), CLI feature flags. The orchestrator config façade wraps these APIs instead of inventing new path logic.
- `pkg/log` — structured logging plus log streaming helpers (`pkg/log/web/streaming.go`) for event metadata feeds.
- `pkg/dep` — binary packaging flow that produces demo service executables under `.dep/`.

**Extended / adjusted in-place**
- `pkg/nats` — single home for spinning up embedded leaves/clusters and exposing broker stats; orchestration code calls into it rather than starting NATS processes itself.
- Demo services — create `pkg/orchestrator/demo/*/config` façades that wrap `pkg/config` helpers so each demo stays environment-aware without reimplementing global defaults.

**Reference-only (copy patterns, no coupling)**
- `pkg/goreman` — supervision patterns inspire the new event-driven runner, but no compatibility layer remains.
- `pkg/metrics` & `pkg/metrics/web` — Prometheus + Datastar wiring references for our own metrics code (including Surveyor exposure) while keeping orchestrator metrics self-contained.
- `pkg/status/web` and `pkg/config/web` — Datastar SSE patterns we mirror inside the orchestrator UI without importing their handlers directly.

---

## 2. Milestones

### M0 — Module Skeleton & Broker Boot
- [ ] Scaffold the new orchestrator package tree under `pkg/orchestrator/` (initial folders: `bus/`, `controller/`, `process/`, `spec/`, `state/`, `harness/`).
- [ ] Implement `pkg/orchestrator/shared/config` façade exposing orchestrator-specific settings (broker mode, data paths, feature flags).
- [ ] Use the `pkg/nats` bootstrap helpers to start the embedded NATS leaf for dev/test and production (honouring `.data-test` vs `.data`) and to bridge to external clusters when configured.
- [ ] Fail-fast broker startup (short retry loop, then fatal error) with clear CLI log.
- [ ] Hook up bootstrap stubs so both UI shells (Datastar + Ultraviolet) can render initial connectivity status.
- [ ] Ensure the startup pipeline is fully deterministic (single path through pkg/nats + orchestrator core) and document the sequence.

### M1 — Event Schema & State Reducers
- [ ] Define JSON event envelope structure (`version`, `type`, `timestamp`, `origin`, `data`). Document canonical subjects: `orchestrator.service.register`, `orchestrator.service.ensure`, `orchestrator.service.port.conflict`, `orchestrator.service.port.reclaimed`, `orchestrator.service.start`, `orchestrator.service.ready`, `orchestrator.service.stop`, `orchestrator.service.start.failed`, `orchestrator.command.*`.
- [ ] Implement reducers that rebuild orchestrator state from the event stream with optional snapshot events for fast bootstrap. Expose query APIs for current snapshot + streaming.
- [ ] Wire reducers into both UI shells so they render live state directly from the orchestrator event stream and publish service toggle commands (command subjects exist even before the runner is wired).

### M2 — Initial UI Shells (Web & TUI)
- [ ] Implement Datastar `/status` shell under `pkg/orchestrator/ui/web` that subscribes directly to orchestrator broker events from the start, including UI toggles for enabling/disabling services.
- [ ] Implement Bubble Tea/Ultraviolet TUI shell under `pkg/orchestrator/ui/tui` driven by the same reducers, with service enable/disable controls mirrored from the web shell.
- [ ] Document shared template structure (`ui/web/templates/`, `ui/tui/templates/`) and list feature parity expectations.
- [ ] Verify both UI shells remain functional against the live broker stream after this milestone.

### M3 — Process Runner & Spec Registration
- [ ] Implement spec registration via Go packages publishing `orchestrator.service.register` events (spec declares logical `portKey`, `dataKey`, `env` keys, `portStrategy`, `restartPolicy`).
- [ ] Build a new event-driven `ProcessRunner` (goroutine + context) that fully replaces goreman; publish events for ensure → start → ready, port conflicts, restarts, exit codes.
- [ ] Support restart policies (`never`, `on-failure`, `always`) with exponential backoff and event notifications when retries stop.
- [ ] Add two demo services (`alpha`, `beta`) under `pkg/orchestrator/demo/alpha` and `pkg/orchestrator/demo/beta`, each with `main.go`, `cli/`, and `web/` folders to provide real binaries and endpoints for orchestration tests.
- [ ] Ensure UI shells surface demo service lifecycles using real events produced by the runner, their enable/disable controls send commands successfully, and both UIs expose the same feature set.

### M4 — Harness & Tests
- [ ] Provide a Go harness (`orchestratorharness.New(t)`): spins up in-memory JetStream, injects fake specs, offers event recorder + cleanup of `.data-test` dirs. Optional CLI harness command for manual replay.
- [ ] Convert orchestrator CLI tests (up/list/status/watch) to use the harness (assert on event payloads instead of logs/globals).
- [ ] Document harness usage, including how to replay captured event logs and skip preflight/codegen in tests.
- [ ] Confirm harness exercises both UI layers (web/TUI) in automated smoke tests.

### M5 — CLI & UI Integration
- [ ] Provide a dedicated `cmd/orchestrator/main.go` entrypoint that boots the new orchestrator core, demo services, CLI, web handlers, and the TUI (no dependency on legacy supervisor) and automatically ensures demo binaries exist (invoking `pkg/dep` install when needed).
- [ ] Move orchestrator CLI code into `pkg/orchestrator/cli` (wrapping the core module, gated behind `INFRA_ORCHESTRATOR_V2`).
- [ ] Move web/SSE handlers into `pkg/orchestrator/ui/web` subscribing to JetStream events.
- [ ] Implement Bubble Tea/Ultraviolet TUI in `pkg/orchestrator/ui/tui`, reusing the same event reducers/state as the web layer.
- [ ] Validate final feature parity checklist (TUI vs web) before flipping orchestrator default.
- [ ] Remove legacy supervisor references once V2 becomes default.
- [ ] Finalize TUI/web parity checklist before flipping the default orchestrator.

#### UI/TUI Readiness Checklist (playable at every milestone)
- **M0**: Both shells (`ui/web`, `ui/tui`) connect to the live broker stream immediately so `go run ./pkg/orchestrator/ui/web` and `go run ./pkg/orchestrator/ui/tui` display broker status and basic service listings.
- **M1**: Reducer wiring lets each shell process the live broker stream and flip command toggles (publishing command subjects even before the runner executes them); parity review confirms identical controls.
- **M2**: Datastar and Bubble Tea shells expose matching enable/disable interactions with shared templates and reducers; manual smoke verifies they render side by side from the same live event stream.
- **M3**: Real ProcessRunner events drive both shells; toggles dispatch identical commands and reflect demo service lifecycle state in lockstep.
- **M4**: Harness smoke tests launch web/TUI shells under test to guarantee parity stays intact; snapshot recordings live under `.data-test` as part of regression tests.
- **M5**: Release checklist signs off on feature parity (all orchestrator operations visible and controllable in both shells) before making orchestrator V2 the default.

---

## 3. Design Decisions — Options & Recommendations

| Topic | Options Considered | Recommendation |
|-------|--------------------|----------------|
| Broker mode | 1) Always embedded leaf using `pkg/config`; 2) Feature-flag permutations; 3) Auto-detect external broker | Run the same embedded leaf bootstrap every time via `pkg/nats`; optional external URLs are read from `pkg/config` so behaviour stays deterministic |
| Broker failure handling | 1) Fail fast; 2) Retry indefinitely; 3) Silent fallback to memory | Fail fast after a short retry loop and surface fatal error |
| Event schema format | 1) JSON envelope; 2) Protobuf; 3) CloudEvents | JSON envelope (`version`, `type`, `timestamp`, `origin`, `data`) for inspectability |
| Schema evolution | 1) Versioned payload, ignore unknown fields; 2) Versioned subjects; 3) Migration layer | Version field in payload, consumers ignore unknown fields; add snapshot events if replay slows |
| Spec registration | 1) Go packages publish events; 2) Config files published at boot; 3) JetStream-backed registry first | Go packages publish `orchestrator.service.register` events (extensible later to external publishers) |
| Spec metadata | 1) Full resource block in event; 2) Logical keys resolved via `pkg/config`; 3) Minimal spec, runner fetches config | Use logical keys (`portKey`, `dataKey`, etc.) plus `portStrategy` for runner resolution |
| Process runner implementation | 1) Fork goreman internals; 2) Build fresh supervisor; 3) Pluggable interface | Build new event-driven runner, lifting ideas from goreman only by copying/reference—no compatibility layer |
| Logs & exit visibility | 1) Emit inline log fragments; 2) Metadata-only events; 3) Filesystem-only logging | Publish metadata events (attachments + exit code); keep full logs on disk; optional tail fragments for tests |
| Restart policy | 1) Spec-defined policies; 2) External command-driven; 3) Manual only | Spec field (`never`, `on-failure`, `always`) with exponential backoff and event notifications |
| Port allocation | 1) Spec-level strategy; 2) Global randomization in tests; 3) OS allocated with follow-up event | Spec declares `portStrategy`; tests choose `random`, dev/prod stay `fixed`; ready events include actual port |
| Demo binaries refresh | 1) Manual `go build`; 2) Custom scripts; 3) Leverage `pkg/dep` | Orchestrator startup automatically invokes `pkg/dep` installers when binaries are missing, keeping behaviour identical each run |
| Harness design | 1) Go API; 2) CLI tool; 3) Both | Provide both: Go harness for automated tests, CLI wrapper for manual replay |
| CLI/UI rollout | 1) Feature flag; 2) Hard cutover; 3) Auto-detect | Launch behind `INFRA_ORCHESTRATOR_V2`, then flip default and remove legacy code |
| Security/auth | 1) Reuse existing creds; 2) Always require creds; 3) Middleware validation only | Reuse existing prod creds; dev/test default to anonymous but allow opt-in creds; middleware planned later |
| Data retention | 1) Rolling window; 2) Unlimited; 3) Dual streams | Configure JetStream rolling window limits; supply snapshot/replay tooling |
| Failure/conflict signalling | 1) Dedicated event types; 2) Generic error event; 3) Logs only | Emit dedicated typed events with full metadata |
| Package layout | 1) `pkg/orchestrator`; 2) New module elsewhere; 3) Rename old package | Use `pkg/orchestrator` with subpackages (`bus`, `controller`, `process`, `spec`, `state`, `harness`, `cli`, `ui/web`, `ui/tui`); legacy supervisor removed after adoption |
| Config access | 1) Call `pkg/config` everywhere; 2) Add `pkg/orchestrator/shared/config` façade; 3) New dedicated config module | Provide `pkg/orchestrator/shared/config` façade wrapping `pkg/config` for orchestrator-specific settings (broker mode, data paths, feature flags), reused by demo services |
| UI architecture | 1) Separate patterns for web/TUI; 2) Shared reducers with different renderers; 3) Web-only | Share reducers/state between web SSE and Bubble Tea/Ultraviolet TUI; renderers receive the same event-fed data |
| Observability | 1) Prometheus metrics; 2) Metrics as events; 3) Logging only | Instrument Prometheus metrics (event throughput, restarts) via `pkg/nats` helpers or external Surveyor; optionally publish aggregated metrics events |
| Multi-node coordination | 1) Support multiple controllers via shared cluster; 2) Single controller only; 3) Partitioned responsibilities | Support multiple controllers joining cluster; include node ID in events; defer locking/leader election to follow-up |

---

## 4. Risks & Mitigations
| Risk | Mitigation |
|------|------------|
| Greenfield build scope creep | Define clear acceptance criteria per milestone; reuse existing config/helpers where possible. |
| Embedded JetStream overhead in dev | Default to memory store, random ports; document cleanup; allow persistent mode for long-lived dev runs. |
| CLI/UI regressions | Develop new CLI/UI directly against orchestrator V2, use flag for gradual rollout; legacy supervisor kept only as read-only reference until removal. |
| Harness flakiness | Randomize ports, isolate `.data-test`, ensure harness teardown cleans resources; add stress tests. |
| External cluster coordination complexity | Keep cluster integration simple (publish/subscribe); defer advanced coordination (locks) to follow-up task if needed. |

---

## 5. Deliverables Checklist
- [ ] New orchestrator module booting JetStream, with configuration overrides.
- [ ] Event schema docs + helper library for publishing/subscribing.
- [ ] Event-driven ProcessRunner + spec registration flow.
- [ ] Harness package (Go + optional CLI) with deterministic tests covering orchestrator CLI/status flows.
- [ ] Updated CLI/UI consuming new orchestrator APIs; flag for rollout; legacy supervisor slated for deletion.
- [ ] Bubble Tea/Ultraviolet TUI sharing reducers with web SSE implementation, with documented feature parity.
- [ ] Demo services (`alpha`, `beta`) built with orchestrator integration.
- [ ] Orchestrator entrypoint automatically (re)installs demo binaries into `.dep/` via `pkg/dep` so every run has identical prerequisites.
- [ ] Shared config façade (`pkg/orchestrator/shared/config`) exposing orchestrator flags/paths.
- [ ] Observability hooks (metrics, log attachments) and documentation.
- [ ] Metrics wiring documented, making it trivial to hook `pkg/nats` helpers or external Surveyor into orchestrator builds.
- [ ] Strategy for external cluster connectivity (config, credentials, node IDs).
- [ ] Update dep.json / build scripts to include demo binaries.
- [ ] Initial orchestrator docs published in `pkg/orchestrator/docs/` (architecture, onboarding, demos) **after** all other deliverables land.

---

## 6. Follow-ups (Post-V2)
- Integrate orchestrator events into observability dashboards and alerting (Prometheus/Grafana).
- Build CLI tooling for inspecting/replaying event streams, debugging port conflicts, and tracing restarts.
- Explore multi-instance orchestration primitives (leader election, distributed locks) using JetStream.
- Delete the legacy supervisor module once all dependencies run on orchestrator V2.
