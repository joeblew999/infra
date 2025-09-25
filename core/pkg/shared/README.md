# Shared Modules

`core/pkg/shared` provides the concrete implementations reused by both the orchestrator runtime and any services built on top of it. Each subdirectory mirrors a runtime package under `core/pkg/runtime`:

- `config`, `dep`, `log`, `events`, `state`, `process`, `controller`, `bus` — primitives for configuration, binary management, logging, event envelopes, state reducers, supervision helpers, and JetStream connectivity.
- `cli`, `ui/web`, `ui/tui` — shared view models, reducers, and component templates that keep all interfaces in parity.
- `caddy` — configuration builders, module registry, and certificate distribution logic for the custom Caddy binary.
- `pipeline` — reusable pipeline steps including `ko` and `fly` integrations.
- `harness/testutil` — deterministic fixtures for tests and demos.

When introducing new behaviour, start here, document the contract, and then wrap it in the runtime tree.
