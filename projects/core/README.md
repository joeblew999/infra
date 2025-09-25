# Core Orchestrator Project

`go run ./projects/core` will become the primary entry point for the new deterministic platform. For now it is a stub; once the runtime packages are ready this binary should:

1. Load configuration via `core/pkg/runtime/config`.
2. Start the embedded bus, Caddy, and controller.
3. Expose CLI/TUI/Web shells powered by the shared reducers.

Keep legacy `main.go` untouched until the new stack reaches feature parity.
