# Task 003 — Unified Path & Environment Layout

## Context
- [ ] Runtime assets currently live under repo-local dot folders (e.g. `.data`, `.logs`), while recent changes introduced an `app/` root mirroring Fly’s `/app` but only for some paths.
- [ ] We need a consistent story for three environments: running from source, running an installed CLI on a developer laptop (respecting XDG/LOCALAPPDATA), and running inside Fly.io (where `/app` is the mounted volume).
- [ ] Toolchain downloads (`.dep`) should remain global, but project-specific runtime data must live in a per-project workspace.

## Goals
- [ ] Define a clear separation between global caches (tooling) and per-project runtime directories, with environment-aware path helpers in `pkg/config`.
- [ ] Implement automatic detection (without new env vars) for repo mode, installed-CLI mode (XDG/LOCALAPPDATA), and Fly.io mode.
- [ ] Update all `Get*Path` helpers to resolve through a new `GetAppRoot()` (or equivalent) so services/tests hit the correct directories for each environment.
- [ ] Ensure tests continue to use `.data-test` regardless of environment, preserving isolation.
- [ ] Document the layout so contributors know where data/logs/configs live in each environment.

## Deliverables
- [ ] Updated `pkg/config` path helpers with consolidated `GetAppRoot()` logic and environment detection.
- [ ] Migration of existing runtime paths (NATS, PocketBase, Hugo, etc.) to the new helper, removing hardcoded repo-relative constants where possible.
- [ ] Documentation update (README/AGENT) describing the global vs project folder structure and how it maps to dev, CLI install, and Fly.
- [ ] Validation notes showing `go run .`, `go test ./...`, and `infra runtime up --only <name>` work across the environments (source checkout vs synthetic installed path).

## Open Questions / Needs Input
- [ ] Default location for global caches when running as an installed CLI (e.g. `~/.infra/.dep` vs `$XDG_DATA_HOME/infra/.dep`).
- [ ] Naming/location for per-project workspaces (e.g. `~/infra-projects/<name>` vs user-selected path).
- [ ] How to expose overrides (e.g. `INFRA_HOME`, `INFRA_PROJECT_ROOT`) without complicating the default path logic.

## Next Steps (draft)
- [ ] Sketch detection flow (repo vs installed vs Fly) and confirm defaults with the team.
- [ ] Implement `GetAppRoot()` + helper refactors in `pkg/config`, update call sites.
- [ ] Update docs and validate commands/tests.
