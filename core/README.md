# Core Platform

This tree hosts the greenfield deterministic core platform described in Task 015. All new runtime and shared modules live here, isolated from the legacy code that remains under `services/`.

Key goals:
- Deterministic startup pipeline built around an embedded JetStream bus and managed Caddy.
- Mirrored shared modules that services can import directly without touching the legacy helpers.
- No mocks: CLI, TUI, Web, and harness tests all speak to the real orchestrator components.

See `core/pkg/README.md` for the module layout and linkages.
