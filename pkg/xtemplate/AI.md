# XTemplate / AI Collaboration Guide

## Web Admin

just for the xtemplate stuff.

- infra/web will import xtemplate web as it needs to.

## Plan

### Problem: CLI doesn’t bring up dependencies
- xtemplate CLI should run standalone but currently assumes NATS/Caddy are already alive. Fix by:
  - Detecting NATS via pkg/config; if missing, start the leaf using existing orchestrator code.
  - Ensuring Caddy is running (use pkg/caddy StartSupervised) before launching.
  - Reuse existing config paths (`config.GetXTemplatePath()` etc.) so data lands in `.data-test` when ENV=test.

### Problem: Admin UI only available via pkg/webapp
- Today `go run . tools xtemplate serve` only shows templates. We need the same process to expose the Datastar admin (same UX as full stack):
  - Run the xtemplate binary on an internal port.
  - Mount both template proxy and admin router on the public CLI port, and print both URLs.
  - Keep paths stable (`/xtemplate` inside main webapp, chosen path locally).

### Problem: pkg/webapp duplicates routing knowledge
- After unifying routing, export a helper (e.g. `xtemplate.RegisterAdminRoutes(router)`) so `pkg/webapp/service.go` stops hand-wiring xtemplate-specific paths.
- That keeps CLI and full stack on the same code path.

### Problem: No test for admin + template together
- Extend `xtemplate_integration_test.go` to hit both the template root and the admin endpoint, verifying dependencies auto-start and responses look sane.
- Later add Playwright coverage targeting the same URLs once routing is in place.

### Problem: Upstream fixtures manual
- `cli xtemplate upstream sync` exists but isn’t tied to the build; users forget to run it.
- Add `//go:generate go run . tools xtemplate upstream sync` so embedded assets stay current with the upstream repository.
