---
title: "Task 002 — Evaluate Process Compose Adoption"
summary: "Assess replacing goreman with Process Compose locally and on Fly.io."
draft: false
---

# Task 002 — Evaluate Process Compose Adoption

IS THIS REALLY RELEVNET ANYMORE ????

## Context
- [ ] Current process supervision is handled with goreman + custom helpers; we’re layering on status views, per-service start/stop, and port management manually.
- [ ] Process Compose (latest release v1.75.2) ships built-in status dashboards, declarative service definitions, health checks, and CLI controls that overlap with our roadmap.
- [ ] Need to determine whether Process Compose can replace/augment our supervisor both locally and on Fly.io.
- [ ] NATS orchestration must integrate cleanly—Process Compose would need to expose what’s running so our embedded leaf/cluster logic stays in sync.
- [ ] Regardless of TUI/web shipped by Process Compose, we still require our own DataStar-based dashboard (fed from NATS) to keep the existing UX and event stream consistent.

## Goals
- [ ] Confirm Process Compose runs cleanly in our local development environment (within `./app`).
- [ ] Validate deployment on Fly.io—ensure the binary boots, services run, and logs/health endpoints behave.
- [ ] Compare feature coverage: service start/stop, status reporting (CLI & web), port/env handling, restarting, dependency chains.
- [ ] Outline a migration path (or reasons to defer) that dovetails with Task 001’s service isolation work and keeps NATS/DataStar orchestration coherent.

## Deliverables
- [ ] Experiment notes (commands, configs) showing Process Compose running locally with at least two services.
- [ ] Fly.io proof-of-life: minimal app using Process Compose v1.75.2, with deployment/log output captured.
- [ ] Summary doc (or README update) evaluating pros/cons versus goreman, including any blockers.
- [ ] Recommendation: adopt now, adopt later (with prerequisites), or stay on goreman; call out NATS + DataStar dashboard implications explicitly.

## Open Questions / Needs Input
- [ ] What minimum feature parity do we require before switching (e.g., DataStar integration, service metadata)?
- [ ] How do we want to expose Process Compose’s config (single YAML, package-specific fragments, generated file)?
- [ ] Should Process Compose replace goreman entirely, or run alongside as an optional mode initially?
- [ ] How will Process Compose share service state with our NATS + DataStar pipeline (e.g., status API, custom hooks)?

## Next Steps (draft)
- [ ] Fetch v1.75.2 binary, run a local sample config alongside our existing services.
- [ ] Deploy the sample to a Fly.io test app and collect logs.
- [ ] Document findings and review with the team.
