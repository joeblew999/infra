---
title: "Task 010 — Markdown-Driven CLI & Web"
summary: "Implementation backlog for the Markdown-first realtime stack."
draft: false
---

# Task 010 — Markdown-Driven CLI & Web
summary: Build the implementation backlog for delivering the Markdown-first realtime stack (Task 009 spec).

---

## 1. Scope

Translate the architecture defined in Task 009 into actionable engineering work across backend, web, and CLI components. Focus on:

- Template registry service + tooling
- SSE view stream plumbing (snapshots, deltas, template events)
- Unified Markdown rendering pipeline (Goldmark for web, Glamour/Bubble Tea for CLI)
- Bidirectional interaction model (actions/mutations surfaced in both clients)
- Delivery milestones and testing strategy

---

## 2. Milestones

### M0 — Foundations
- [ ] Scaffold template registry HTTP API (stub responses).
- [ ] Expose `/api/views/{view}/snapshot` and `/stream` with static JSON.
- [ ] Build CLI skeleton command `infra cli view` with basic SSE connection + JSON logging.

### M1 — Shared Templates
- [ ] Move existing Datastar view(s) to Go Markdown templates in `/cli/templates`.
- [ ] Implement Goldmark renderer with `html.WithUnsafe(true)` for inline Datastar attributes.
- [ ] Add `go run . dev templates lint` command to validate helpers + schema tags.
- [ ] Publish templates into registry, update web to fetch from registry instead of embedded HTML.

### M2 — SSE-Backed CLI Renderer
- [ ] Implement view-model reducer (snapshot + JSON Patch deltas).
- [ ] Integrate Glamour for Markdown → ANSI rendering.
- [ ] Integrate Bubble Tea v2 + Ultraviolet viewports for diffed updates.
- [ ] Support `--no-style`, `--json`, `--auth-header`, `--resume-from` flags.
- [ ] Document Charm stack responsibilities:
  - Ultraviolet → cell-level diff renderer (Datastar-style morph for terminals).
  - Bubble Tea v2 → Elm-style event loop + Ultraviolet integration.
  - Bubbles Viewport → high-performance panel scrolling.
  - Glamour → Markdown → ANSI, matching web content.
  - Lip Gloss → layout/styling primitives shared across panels.

### M3 — Actions & Mutations
- [ ] Extend view-model contract with an `actions` collection (id, label, params, confirmation text).
- [ ] Implement backend action router (`POST /api/views/{view}/actions/{id}`) with auth + audit hooks.
- [ ] Emit SSE `action-result` events so web/CLI can reflect success, failure, or progress.
- [ ] Wire Datastar forms/buttons to call action endpoints and reconcile incoming deltas.
- [ ] Add CLI affordances: `infra cli view --run <action>` and interactive prompts for parameter capture.
- [ ] Support optimistic updates + rollback by tracking pending actions in reducer state.

### M4 — Template Hot Reload & Persistence
- [ ] Listen for SSE `template` events to reload Markdown + styles live.
- [ ] Cache latest snapshot + template version to disk (`~/.infra-cli/state`).
- [ ] Add CLI/web debug mode showing event lag, template version, render timing.

### M5 — Hardening & QA
- [ ] End-to-end tests: replay SSE transcripts, simulate template downgrade, schema mismatch, auth failures.
- [ ] Add rate limiting + idle timeouts to SSE broadcaster.
- [ ] Document operational runbooks (publishing, rollback, replay tooling).

---

## 3. Workstreams & Owners (draft)

| Stream | Lead | Notes |
|--------|------|-------|
| Template Registry | Backend | New service/module, ties into existing config/auth |
| View Streams | Backend | Bridges existing JetStream data to SSE endpoints |
| Web Renderer | Web | Moves Datastar views to Markdown pipeline |
| CLI Renderer | CLI | Bubble Tea + Glamour integration |
| Dev Tooling | Platform | Template lint/publish commands, replay tooling |

---

## 4. Key Decisions to Finalise

- Auth mechanism for template + SSE endpoints (bearer tokens vs signed cookies).
- JSON Patch library choice and merge strategy for deltas.
- Template publish workflow (manual command vs CI pipeline).
- Rollout order for existing views (which dashboards migrate first).
- Action contract: payload schema, optimistic update rules, acknowledgement semantics.

---

## 5. Risks & Mitigations

| Risk | Mitigation |
|------|------------|
| Large SSE payloads overwhelm CLI diffing | Segment views into panels; diff per viewport |
| Template version mismatch between web/CLI | Enforce SSE `template` events + cached version checks |
| Goldmark rendering deviates from Datastar expectations | Extensive snapshot testing + visual diffs |
| Registry downtime blocks rendering | Local cache fallback + document manual restore path |
| Action side effects race with view updates | Emit `action-result` events with correlation IDs; reducer drops stale optimistic state |

---

## 6. Deliverables Checklist

- [ ] Template registry service in production (+ docs).
- [ ] Web UI rendering via Markdown templates.
- [ ] CLI view command with Bubble Tea diffing.
- [ ] Shared action/mutation workflow (HTTP endpoints, SSE `action-result`, CLI/web UX).
- [ ] Automated tests + replay tooling.
- [ ] Operational runbook + developer onboarding guide.

---

## 7. Follow-ups

- Draft ADR on Markdown-first UI strategy.
- Identify additional views to port post-MVP.
- Explore open-sourcing CLI components once stable.
- Author design doc for action schema, optimistic update patterns, CLI UX.
- Research WYSIWYG decksh editor architecture (source-of-truth sync, live SVG editing).
