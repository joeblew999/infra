# 009 web-cli-realtime


summary. Realtime CLI via Markdown → ANSI (Glamour) — Specification

## 1. Purpose
Define a unified presentation system where both **Web (Datastar)** and **CLI** render from the same **JSON view-model** and **Markdown templates**. The CLI converts Markdown to ANSI using **Glamour** and updates live from **NATS** (preferred) or **SSE**.

## 2. Architecture
- **Model**: JSON view-model published over NATS (and/or exposed via SSE).
- **Templates**: Go `text/template` files outputting Markdown. Stored and versioned in NATS KV/Object.
- **Web**: Datastar consumes the view-model and renders HTML.
- **CLI**: Cobra consumes the view-model, renders Markdown templates, then pipes Markdown through Glamour → ANSI.
- **Hot Reload**: Both Web and CLI watch KV keys and reload templates live.

## 3. Transport
- **NATS (preferred)**:  
  - Subjects: snapshots (`state.<domain>.<view>`), events (`events.<domain>.<entity>.*`).  
  - Consumers: `DeliverPolicy=Last` for live only, `ByStartTime` for catch-up.  
  - Resilient reconnect, durable consumers per view.  
- **SSE**: Optional parity with Web. CLI can choose `--source sse:https://...` or `--source nats:nats://...`.

## 4. Template Versioning
- KV keys:  
  - `templates/cli/<view>.current` → `vX.Y.Z`  
  - `templates/cli/<view>/vX.Y.Z` → Markdown template blob  
- On startup: fetch current pointer, load versioned template, parse.  
- Watch `<view>.current` for hot reload.

## 5. View-Model Contract
```json
{
  "schema": "com.example.ui.global",
  "schema_version": "1.2.0",
  "generated_at": "2025-09-24T09:30:00Z",
  "filters": { "region": "apac", "tenant": "t-123" },
  "data": { "... per-view payload ..." }
}
````

Templates declare supported schema ranges in front-matter. CLI validates schema and warns or fails if incompatible.

## 6. Markdown Template Conventions

* Output pure Markdown: headings, tables, lists, emphasis, code, blockquotes.
* Compose with `{{ define "partial" }}` and `{{ template "partial" . }}`.
* Example:

```md
<!-- schema: com.example.ui.global min:1.0.0 max:2.0.0 -->

{{ define "badge" -}}
{{ if eq . "healthy" }} ✅ **healthy**
{{ else if eq . "degraded" }} ⚠️ *degraded*
{{ else }} ❌ {{ . }}
{{ end }}
{{- end }}

# {{ .Title }}

Filters: `region={{ .Filters.region }}` `tenant={{ .Filters.tenant }}`

| Name | Region | Status |
|------|--------|--------|
{{- range .Data.Servers }}
| **{{ .Name }}** | {{ .Region }} | {{ template "badge" .Status }} |
{{- end }}

> Updated live via NATS at {{ .GeneratedAt }}.
```

## 7. CLI Rendering Pipeline

1. **Source**: select NATS (default) or SSE.
2. **Template Load**: fetch and parse Markdown template.
3. **Data**: apply snapshot + event updates into view-model.
4. **Render**: execute template → Markdown → Glamour → ANSI if TTY.
5. **Fallback**: if not TTY or `--no-style`, output raw Markdown or NDJSON (`--json`).
6. **Debounce**: re-render every 100–200 ms max.
7. **Shutdown**: SIGINT/SIGTERM stops gracefully.

## 8. CLI Flags

* `--source`: `nats://` or `sse://`.
* `--view`: name of template/view.
* `--watch`: stay live.
* `--since`: seek JetStream history.
* `--json`: emit NDJSON view-models.
* `--no-style`: disable Glamour (print Markdown).
* `--strict`: fail on schema mismatch.
* `--filter`: apply filters (`--filter region=apac`).
* NATS auth: `--nats-creds`, `--nats-nkey`, `--nats-jwt`.

## 9. Styling

* Use `glamour.WithAutoStyle()` by default.
* Themes: `"dark"`, `"light"`, `"dracula"`, `"notty"`.
* Custom style JSON stored in KV (`templates/cli/_styles/brand/...`).

## 10. Template Helpers

Provide template funcs:

* `ago(ts)`, `timefmt(ts, layout)`, `truncate(s, n)`.
* `percent(f)`, `bar(curr, max)` → progress bar.
* `kv(map)` → Markdown rows.
* `plural(n, sing, plur)`.

## 11. Data Flow

* **Cold Start**: load snapshot (KV/Object).
* **Live Updates**: merge deltas.
* **Persist Last**: cache template + state for instant boot.

## 12. Error Modes

* **Template Missing**: show error, retry.
* **Schema Mismatch**: warn or fail (`--strict`).
* **Transport Down**: show last frame + “reconnecting…”.
* **Narrow Terminal**: wrap or stack rows.

## 13. Security

* Reuse NKEY/JWT/creds.
* Support `--nats-js-domain` for multi-domain JetStream.
* Allow nearest-region NATS URL.

## 14. Repo Layout

* `/cli/templates/` → dev Markdown templates.
* `/cli/styles/` → Glamour JSON.
* `/cli/cmd/` → Cobra commands.
* `/cli/internal/render/` → template + Glamour.
* `/cli/internal/model/` → view-models.
* `/cli/internal/nats/` → NATS client.
* `/cli/internal/sse/` → SSE client.

## 15. Rollout Checklist

* Expose NATS/SSE flags.
* Implement template loader + watcher.
* Add state reducer.
* Render Markdown → Glamour.
* Add debounce + resize detection.
* Add schema validation.
* Cache last frame (optional).

## 16. Ops Notes

* Version templates with semver.
* Deploy templates via atomic KV pointer update.
* Add `--debug` for event rates, lag, template version.
* Test by replaying recorded NDJSON.

## 17. KV Layout Example

* `templates/cli/global.current` → `v1.7.2`
* `templates/cli/global/v1.7.2` → template blob
* `templates/cli/_partials/badge/v1.3.0` → partial template
* `templates/cli/_styles/brand/v0.4.0` → Glamour JSON
* `state/global/snapshot` → snapshot blob
* `events/global/*` → JetStream events

## 18. User Stories

* Operator runs `myctl global --watch`, gets live ANSI dashboard.
* Developer adds field, bumps template version, publishes, both Web + CLI pick up live.
* SRE tails NDJSON: `myctl global --json | jq ...`.

## 19. Non-Goals

* Full HTML fidelity in terminal.
* Complex per-row interactivity (future Bubble Tea wrapper possible).

## 20. Summary

This spec defines a **Markdown-first** template system rendered as **HTML** for Web and **ANSI** for CLI, with **NATS-driven** realtime updates, live template versioning via KV/Object, and a CLI UX that degrades gracefully to Markdown or NDJSON for automation.

```
```
