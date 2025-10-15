---
title: "DatastarUI Tooling Guide"
summary: "Workflow for the DatastarUI fork and Go helper commands."
draft: false
---

Use this guide when working on the DatastarUI frontend fork and its Go-based tooling wrappers inside the infra repo.

## Sources (do not remove)
- **Documentation**
  - https://datastarui.dev
- **Source Repositories**
  - https://github.com/coreybutler/datastarui
  - https://github.com/starfederation/datastar-go
- **Tooling**
  - https://bun.sh/docs/cli
  - https://templ.guide/
  - https://tailwindcss.com/docs
  - https://playwright.dev/docs/intro

---

## Part 1 – DatastarUI Fundamentals

### Topology
- **Upstream mirror**: Clean copy of Corey Butler’s DatastarUI repo; consume updates here.
- **Working fork**: Our editable fork where UI adjustments, Playwright specs, and Tailwind changes live.
- **Go wrapper**: Tooling that shells into Bun workflows, generates assets, and serves sample apps.

Keep fork changes downstream of upstream—sync before editing.

### Core Toolchain
- **Bun** for package management, scripts, and Playwright installs. Avoid npm/pnpm/just.
- **templ** generates Go templates from `.templ` files.
- **Tailwind CLI** compiles CSS; hashed outputs ensure cache busting.
- **Playwright** validates interactions across browsers.

Install prerequisites locally:
```bash
brew install oven-sh/bun/bun            # or follow Bun install guide
brew install templ                      # templ CLI
npm install -g tailwindcss              # or use Bun bundle
bun x playwright install                # installs browsers
```

### Daily Loop
1. Sync upstream → fork (git fetch/rebase) to stay current.
2. Run Bun workflows (`bun install`, `templ generate`, Tailwind build) when editing UI.
3. Produce hashed CSS (`out.<hash>.css`) so Go tooling can embed assets.
4. Execute `bun x playwright test` (headed or headless) to verify UI changes.
5. Commit fork updates separately from Go tooling modifications.

### Best Practices
- Track hashed CSS artifacts; do not rename them manually.
- Keep templ and Tailwind configs aligned across upstream and fork.
- Use Playwright tags to isolate slow suites when iterating.
- Document new components or utilities before exposing them to Go generators.

---

## Part 2 – Infra Repo Playbook

### Directory Map
| Area | Purpose | Notes |
| ---- | ------- | ----- |
| `.src/datastarui/upstream/datastarui` | Upstream mirror | No modifications; pull updates only. |
| `.src/datastarui/fork/datastarui` | Working fork | All frontend edits land here. |
| `pkg/datastarui/cmd/*` | Go wrappers (`codegen`, `playwright`) | Bridge Bun workflows with Go CLIs. |
| `pkg/datastarui/sampleapp` | Example Go app embedding DatastarUI | Uses generated assets + components. |
| `pkg/datastarui/run-*.sh` | Convenience scripts | Wrap Go commands for repetitive tasks. |

### Syncing the Fork
```bash
git -C .src/datastarui/upstream/datastarui pull
if ! git -C .src/datastarui/fork/datastarui remote | grep -q '^upstream$'; then
  git -C .src/datastarui/fork/datastarui remote add upstream ../../upstream/datastarui
fi
git -C .src/datastarui/fork/datastarui fetch upstream
git -C .src/datastarui/fork/datastarui rebase upstream/main
git -C .src/datastarui/fork/datastarui push origin main
```
- Run from repo root before changing the fork.
- Resolve conflicts in the fork; never modify the upstream mirror.

### Go Wrapper Commands (run inside `pkg/datastarui`)
- **Code generation & assets**
  ```bash
  GOWORK=off go run ./cmd/codegen --src $(pwd)/sampleapp
  ```
  - Accepts flags `--workflow`, `--tailwind-input`, `--tailwind-output`, `--binary`.
  - Defaults to Bun workflow; `--workflow=node` for upstream parity tests.

- **Sample app**
  ```bash
  GOWORK=off go run ./sampleapp
  ```
  - Serves at `http://localhost:4242`; requires hashed CSS from codegen.

- **Playwright helper**
  ```bash
  GOWORK=off go run ./cmd/playwright --src $(pwd)/sampleapp
  ```
  - Triggers codegen automatically; supports `--headed`, `--workflow`, `--base-url`, `--timeout`.

- **Tests**
  ```bash
  GOWORK=off go test ./...
  ```
  - Wraps Playwright execution during `go test` for CI compatibility.

Scripts `./run-datastarui-sample.sh` and `./run-datastarui-fork.sh` provide shortcuts for common loops.

### Editing Flow
1. Modify `.templ`, component, or test files inside the fork.
2. Run Bun build loop to regenerate assets.
3. Execute Go codegen to sync the sample app.
4. Validate with Playwright via Go helper or scripts.
5. Commit fork and Go changes separately; note hash changes in PR descriptions.

### MCP Automation Notes
- Navigate to local sample app, capture console logs/screenshots, and report timings via MCP Playwright server.
- Use `*_evaluate` calls for DOM snapshots or Datastar signal inspection.
- Clear logs between steps to highlight new errors.

### Coordination With Other Guides
- Backend SSE concepts live in the [Datastar backend guide](../datastar/).
- Deployment (Fly, ko) updates may require rebuilding the sample app for docs.
- Document workflow changes here whenever scripts, dependencies, or directory layouts shift.
