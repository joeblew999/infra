# Core Agent Resources

# Agents
  
This project shares the central instructions in [../agents/AGENTS.md](../agents/AGENTS.md).
Project-specific canonical docs live in [docs/README.md](docs/README.md); update
those sources first and keep README pointers minimal.


ALWAYS run the code and tests and make sure the core and tooling works and deploy to local and fly on your own !

Recommended baseline before sending changes:
- `go test ./...` from repo root (catches shared package regressions).
- `go test ./tooling/...`, `go test ./controller/...`, `go test ./core/...` when
  touching module-specific code.
- Smoke the runtime (`go run ./cmd/core stack up && go run ./cmd/core stack down`)
  for orchestrator changes.
- Exercise tooling end-to-end with `go run ./tooling workflow deploy` (use a
  test app) to validate Fly + Cloudflare flows.

ALWAYS keep the main README up to date, so we can runs things !

Remember this repo is a Go workspace with three cooperating modules:
- `core/` — runtime CLI + shared packages (this directory)
- `controller/` — desired-state API (run with `GOWORK=off go run .`)
- `tooling/` — Fly/Cloudflare deployment CLI
Stay aware of cross-module impacts when making changes and run smoke checks in
each area as needed.

When touching the tooling CLI, treat `docs/tooling.md` as the canonical source.
Keep that guide in sync with behaviour and only mirror minimal pointers in
`tooling/README.md` and the root `README.md`.

Cloudflare authentication now supports both manual tokens and the
`auth cloudflare bootstrap` flow that creates scoped tokens from a global API
key. Document whichever path you extend and validate the CLI prompts end-to-end
before publishing changes. Tokens are snapshotted per-flow under
`.data/core/cloudflare/tokens/` while the active token pointer drives deploys—be
careful to keep the pointer aligned with the flow you expect to use.

Whenever you alter documentation:
- Update the canonical doc first (`docs/*`), then refresh the concise README
  pointer.
- Note any doc moves or new canonical locations in this file so future agents
  know where to extend guidance.
- Cross-check for duplicate guidance; consolidate instead of copying text.

ALWAYS ensure all services are working. Use Playwright MCP and curl to help you ensure things work.

When CLI auth flows require interactive login (Fly or Cloudflare), you may run the commands here and I can complete the browser prompts; do not skip auth checks.
