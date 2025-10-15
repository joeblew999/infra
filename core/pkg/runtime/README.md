# Runtime Modules

Packages under `core/pkg/runtime` compose the shared implementations into the deterministic orchestrator:

- `config`, `dep`, `log`, `events`, `state` — runtime facades that expose the behaviour defined in the matching shared packages.
- `services/nats`, `services/pocketbase`, `services/caddy`, `process`, `controller` — embedded service runners (Pillow-powered NATS, PocketBase, module-rich Caddy) plus orchestration glue now driven entirely by Process Compose. Use `core stack process <list|info|start|stop|restart|scale>` and `core stack project <state|update|reload>` to manage them at runtime.
- `cli`, `ui/web`, `ui/tui` — user interfaces that must stay in lockstep via shared reducers/components.
- `cli`, `ui/web`, `ui/tui` — user interfaces that must stay in lockstep via shared reducers/components. Live modes expose per-process controls (start/stop/restart/scale) and log tails sourced from Process Compose.
- `services/caddy` — managed reverse proxy tooling built from the custom module registry (no `xcaddy`).
- `pipeline` — orchestrated build/deploy flows replacing the old `pkg/workflows` helpers.
- `harness` — full-stack test harness that spins up the real orchestrator for automated tests.
- `services/demo/*` — real demo services (no mocks) proving the layering and feeding CLI/TUI/Web views.

Every runtime package should only import the corresponding shared module (plus unavoidable peer runtime packages).
