---
title: "Datastar Backend Agent Guide"
summary: "SSE producers and web shells that power Datastar dashboards."
draft: false
---

Primary reference for developers and agents working on Datastar-powered back ends inside `infra`. Pair this with the [DatastarUI tooling guide](../datastarui-tool/) when touching front-end shells.

## Sources (do not remove)
- **Documentation**
  - https://data-star.dev/reference/attributes
- **SDK**
  - https://github.com/starfederation/datastar-go
- **Examples & Community**
  - https://github.com/starfederation/datastar-go/network/dependents

---

## Part 1 – Datastar Fundamentals

### Why Datastar
- Datastar streams HTML fragments over Server-Sent Events (SSE), enabling progressive enhancement without rebuilding entire pages.
- State changes become `morph` events that patch the DOM, keeping client interactions lightweight and resilient to disconnects.
- Works best when producers output deterministic snapshots, letting the frontend reconcile without client-side state containers.

### Core Concepts
- **SSE Streams**: `datastar.NewSSE(w, r)` negotiates the connection, keeps heartbeats alive, and retries automatically.
- **Snapshots & Morphs**: Send a full HTML snapshot on connect, then incremental diffs via `sse.PatchElements(html)` or `sse.Send("event", payload)`.
- **Identity Attributes**: Use `data-datastar-key` and deterministic IDs to let morphs target the correct DOM nodes.
- **Backpressure**: Throttle outbound updates or batch them; Datastar queues on the server side but the client may fall behind if you emit at high frequency.

### Generic Handler Blueprint
```go
func handleStatus(w http.ResponseWriter, r *http.Request) {
    sse := datastar.NewSSE(w, r)
    defer sse.Close()

    if err := sse.PatchElements(renderSnapshot()); err != nil {
        return
    }

    updates := subscribeUpdates()
    for {
        select {
        case <-r.Context().Done():
            return
        case data := <-updates:
            html := renderPartial(data)
            if err := sse.PatchElements(html); err != nil {
                return
            }
        }
    }
}
```
- Always respect `r.Context()` to terminate cleanly.
- Keep rendering pure (no side effects) so you can reuse the same helpers in tests.

### Template & Morph Guidelines
- Compile templates once (`sync.Once`) and keep helper functions returning strings (easier to embed).
- Apply deterministic ordering to lists before rendering; morph diffing expects stable element positions.
- Avoid inline scripts inside morph fragments—deliver JavaScript via shared layout so reconnections do not break.
- Use semantic HTML and Tailwind-compatible classes; Datastar only cares about consistent structure, not styling choices.

### Verification Patterns
- `curl -N http://localhost:<port>/<path>/api/stream` to watch raw events.
- When debugging patches, diff rendered HTML strings locally before emitting to clients.
- Pair SSE integration tests with standard Go unit tests; Datastar handlers are regular HTTP handlers.
- Simulate reconnects by cancelling the request context and ensuring the first snapshot rehydrates the full UI.

---

## Part 2 – Infra Repo Playbook

### Surfaces You Own
1. **Data producers (Go back end)** – packages that gather runtime/config state and publish events: `pkg/status`, `pkg/config`, `pkg/runtime/events`, `pkg/service/state`, plus any process hooks that call `events.Publish`.
2. **Datastar web shells (HTML + SSE)** – dashboards under `pkg/status/web`, `pkg/config/web`, and the shared chrome in `pkg/webapp/templates`. These ship HTML fragments over Datastar SSE and mount inside the main web app (`pkg/webapp`).

Keep these paths in lockstep: producer structs define the JSON/state shape, web shells map that shape into Datastar-friendly HTML, and both must evolve together.

### Repository Map

| Area | Purpose | Key Files |
| ---- | ------- | --------- |
| `pkg/webapp` | Web server that hosts all dashboards | `service.go`, `templates/base.go`, `templates/nav.go` |
| `pkg/status` | Collects runtime/service stats | `status.go`, `runtime.go`, `web/handler.go`, `web/datastar.go` |
| `pkg/config` | Central configuration + live config UI | `config.go`, `web/service.go`, `web/datastar.go` |
| `pkg/runtime/events` | Event bus feeding Datastar streams | `events.go` |
| `pkg/service/state` | Event-sourced snapshot used by status UI | `snapshot.go` |
| `docs/technical/SYSTEM_SPEC.md` | High-level system overview highlighting Datastar webapp | — |

Templates live under each package’s `web/templates/` directory (e.g., `pkg/status/web/templates/status_cards.html`). They are embedded with `//go:embed` and rendered via helper functions in the same package.

### Development Workflow

#### Run the Web UI locally
```bash
go run . --mode=cli web serve
# defaults to http://localhost:1337
```
Flags:
- `--port` (default from `config.GetWebServerPort()`)
- `--nats` for realtime demos (optional for Datastar but required for NATS-backed features)
- `--docs-dev` keep `true` in dev so docs reload from filesystem.

#### Validate producers
- `go test ./pkg/status/...`
- `go test ./pkg/config/...`
- `go test ./pkg/runtime/events`
- `go test ./pkg/service/state`

When producers change exported structs, also run `go run . api-check` from repo root to keep API contracts aligned.

#### Exercise Datastar streams
- Navigate to `http://localhost:1337/status` and `http://localhost:1337/config` for manual QA.
- For automated checks use MCP Playwright steps from the [DatastarUI tooling guide](../datastarui-tool/) (navigate → console logs → interactions → report).
- To inspect SSE manually: `curl -N http://localhost:1337/status/api/stream` (expect `event: morph` payloads).

### Patterns Specific to This Repo

#### Datastar SSE Handlers
All SSE endpoints follow the same structure:
1. Create generator: `sse := datastar.NewSSE(w, r)` from `github.com/starfederation/datastar-go/datastar`.
2. Send an initial snapshot (`sendStatusSnapshot`, `sendConfigSnapshot`).
3. Subscribe to updates (`runtimeevents.Subscribe` or `time.NewTicker`).
4. Emit HTML fragments with `sse.PatchElements(html)` or custom events via `sse.Send(eventName, payload)`.

Throttle updates if the producer can fire rapidly (see `statusStreamInterval` + `lastUpdate` guard in `pkg/status/web/handler.go`).

#### HTML Partials
- Render via helper (`RenderStatusCards`, `RenderConfigCards`). Each uses `sync.Once` and cached `template.Template` instances.
- Templates produce Tailwind classes and Datastar attributes (`data-datastar-morph`, etc.) expected by the frontend. Keep attribute naming consistent with upstream docs.
- Editing flow:
  1. Change template under `web/templates/*.html`.
  2. Adjust mapping struct in `datastar.go` (or equivalent) to supply any new data.
  3. Update snapshot mappers (`mapStatusToTemplate`, `mapConfigToTemplate`).
  4. Regenerate by rebuilding (`go run . --mode=cli web serve`) or rerunning relevant tests.

#### Event Source Lifecycle
- Producers publish events through `runtime/events`. Consumers (status/config streams) subscribe and render.
- `pkg/service/state` listens to events and keeps an in-memory snapshot used by the status dashboard. If you introduce new event fields, update `RuntimeState` and `apply*` helpers accordingly.
- Any supervision changes (e.g., new managed service) must call `events.Publish` so the status UI reflects it.

#### Shared Chrome
- The base layout is defined in `pkg/webapp/templates/base.go`. It injects Tailwind via `https://unpkg.com/@tailwindcss/browser@4` and Datastar via `GetDataStarScript()` pointing to the CDN bundle (`https://cdn.jsdelivr.net/.../datastar.js`). Stay on the CDN unless security requirements change.
- Navigation items are registered in `init()` blocks (see `pkg/status/web/handler.go` and `pkg/config/web/service.go`). Additions must go through `templates.RegisterNavItem`.

### Adding or Modifying a Datastar Dashboard

1. **Define the data contract** in the producer package (structs + mapper functions).
2. **Expose update events** using `runtime/events.Publish` or a ticker.
3. **Create HTML partial** in `web/templates/` with the Datastar attributes needed for morphing or data binding.
4. **Embed & render** the partial with `//go:embed`, `template.New(...).Parse`, and a `RenderXCards` helper returning the HTML string.
5. **Stream via SSE** using the established handler pattern.
6. **Mount routes** in the relevant web service and register the nav item.
7. **Verify locally** (`go run . --mode=cli web serve` + Playwright/MCP snapshot).
8. **Run tests** for both producer and web packages.

For SERP-like or list UIs consider using Datastar events (`sse.Send`) instead of whole-card morphs; follow `HandleServiceEvents` for an example envelope.

### Tooling Expectations

- Go 1.24.x (project standard).
- No Node tooling required for backend work; Bun/templ/Tailwind live in the DatastarUI fork.
- Rely on `go vet`, `go test`, and `go run . api-check` to catch regressions.
- Pre-commit hook runs `go run .`, so make sure `go run .` succeeds after your changes.

### Reference Workflows & Docs

- `docs/technical/SYSTEM_SPEC.md` – big-picture placement of the Datastar web app.
- `docs/README.md` – user-facing health endpoints list.
- `pkg/bento/README.md` – mentions Datastar-backed admin UI; useful for style alignment.
- External Datastar docs and SDK links listed above – consult for attribute naming, SSE event usage, and Go helpers.

Keep this guide updated whenever Datastar endpoints or event contracts change, and coordinate with the DatastarUI tooling guide when UI assets or Tailwind config evolve.
