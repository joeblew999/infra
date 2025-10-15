# DatastarUI Agent & Operator Guide

This README doubles as the Codex agent playbook and the human operator quick reference for `pkg/datastarui`. Follow the instructions exactly so the Go helpers, Bun workflow, and Playwright specs stay in sync with the DatastarUI fork.

---

## Mission Checklist

1. **Read the agents docs first.** Pair this file with `agents/AGENT_datastarui.md` (primary) and `agents/AGENT_datastar.md` for backend context.
2. **Work from the fork checkout.** Shared components and layouts live in `.src/datastarui/fork/datastarui`; do not fork the sample app into its own design system.
3. **Keep sample code lean.** Pages under `sampleapp/` compose shared components only. If you need a new primitive, lift it from the fork.
4. **Use the Go command wrappers.** Run `cmd/codegen` and `cmd/playwright` rather than ad‑hoc Bun scripts; the wrappers wire up templ, Tailwind, Bun, and server lifecycles for you.
5. **Respect hashed CSS assets.** The Tailwind build must write `static/css/out.css` plus a hashed copy; pages should resolve the hash at runtime (see `pages/dashboard.templ`).

---

## Tooling Requirements

- Go 1.24.x (matches the DatastarUI upstream go.mod)
- Bun runtime (`brew install oven-sh/bun/bun` or equivalent)
- templ CLI (`brew install templ`)
- Tailwind CLI (`brew install tailwindcss/tap/tailwindcss`)
- Playwright browsers (installed via `bun x playwright install`)

> **Agent tip:** `pkg/dep` automation is not wired up yet—assume these binaries are already present. If they are missing, surface an explicit error rather than trying to install them silently.

---

## Codegen Workflow (Sample App)

Run from `pkg/datastarui/` unless you are intentionally targeting another DatastarUI checkout.

```sh
GOWORK=off go run ./cmd/codegen --src $(pwd)/sampleapp
```

- Generates templ Go files.
- Rebuilds Tailwind CSS and writes the hashed asset.
- Produces the Go binary (`sampleapp`) when `--binary` is supplied (default: `datastarui` target).

**Alternate workflow:** if you must mirror Corey’s pnpm/docker setup, add `--workflow=node` so the helper emits the pnpm commands instead of driving Bun directly.

---

## Running the Sample App

```sh
GOWORK=off go run ./sampleapp
```

- Boots the local DatastarUI sample server.
- Serves on `http://localhost:4242` by default (override with `--addr` in the helper flags).
- Uses the hashed CSS file resolved at runtime—ensure the Tailwind build ran first.

For scripted usage, `./run-datastarui-sample.sh serve` wraps this command.

---

## Playwright & Tests

### Headless via Go helper

```sh
GOWORK=off go run ./cmd/playwright --src $(pwd)/sampleapp
```

- Performs `bun install`, templ generation, Tailwind rebuild, and Go build before launching tests.
- Runs Playwright headless. Pass `--headed` to open the browser.
- Respects `PLAYWRIGHT_BASE_URL` and the CLI overrides (`--base-url`, `--workflow`, `--timeout`, etc.).

### Go test wrapper

```sh
GOWORK=off go test ./...
```

- Invokes the same Playwright workflow in-process.
- Use `GOWORK=off go test ./sampleapp` for a narrower scope.

Helper scripts mirror these commands: `./run-datastarui-sample.sh <codegen|playwright|test|serve-test>` for the sample app; `./run-datastarui-fork.sh <...>` for the fork checkout.

---

## Layout & Asset Expectations

- **Components**: import from `.src/datastarui/fork/datastarui/components/...`.
- **Layouts**: the sample app should reuse the fork’s layout templates (e.g. `layouts/root.templ`). Only page bodies live locally.
- **Tailwind config**: `cmd/codegen` loads `tailwind.config.js`, which itself imports the fork’s Tailwind config and appends the sample page globs. Keep both configs updated together.
- **Hashed CSS**: use the `out.<hash>.css` convention. The code checks `static/css/out.*.css` at runtime and falls back to `out.css` if no hash is present.

---

## When Editing or Adding Sample Pages

1. Update the `.templ` source (not the generated `.go`).
2. Re-run `GOWORK=off go run ./cmd/codegen --src $(pwd)/sampleapp`.
3. Verify the hashed CSS reference still resolves.
4. Run Playwright tests (`GOWORK=off go run ./cmd/playwright --src $(pwd)/sampleapp`).
5. If you add new components, pull them from the fork and ensure the Tailwind purge globs include the new paths.

---

## Troubleshooting Notes

- **Missing Bun/Tailwind/templ**: fail fast with a clear error message; do not auto-install.
- **Playwright install**: the Go helper runs `bun x playwright install` automatically. If browsers look stale, delete `playwright/.cache` and rerun the helper.
- **Breaking API in sample app**: align page updates with the fork’s components to avoid drift. Re-run `go run . api-check` at repo root when touching shared APIs.

---

## Reference Links

- Primary agent brief: `agents/AGENT_datastarui.md`
- Datastar backend brief: `agents/AGENT_datastar.md`
- Upstream repo mirror: `.src/datastarui/upstream/datastarui`
- Fork staging area: `.src/datastarui/fork/datastarui`

Keep this README synced with any workflow changes—Codex agents rely on the exact commands here to operate autonomously, and human operators expect the same guidance.
