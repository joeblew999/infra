# Task 000 — Service Port Stability & Conflict Recovery

## Context
- [x] `go run .` and `infra runtime up --only <name>` currently fail if a service’s port is already in use (e.g., Hugo on 1313). The supervisor exits instead of recovering.
- [x] Orphaned goreman-managed processes (from previous runs) and manual service starts create port conflicts that are hard to diagnose.
- [x] When the port is owned by a process outside our supervisor (e.g., a user ran `go run .` in another terminal or another application already bound the port), our startup attempt fails with a generic “address already in use” and offers no guidance.
- [x] We need a consistent mechanism to detect ownership, reclaim (when safe), and report port usage so subsequent runs stay idempotent.

## Goals
- [x] Detect port ownership before launching each service and decide whether to reclaim or warn.
- [x] Automatically stop goreman-managed processes for the same service (developer convenience) when in dev mode.
- [x] Surface clear CLI/web messages when a port is owned by an unrelated process—include PID, process name, and suggested remediation (stop the process or change the port).
- [x] Update the service status API/UI to show running processes, ports, and conflicts.
- [x] Document the policy so it’s obvious when we auto-stop vs. when we warn.

## Port Conflict Policy
- [x] **Reclaim:** If the port belongs to a goreman-managed process for the same service/project, stop it automatically and continue (log the action).
- [x] **Warn (other infra session):** If the port belongs to another `infra` supervisor (different workspace/terminal), warn with PID/command and instruct the user to run `infra shutdown` or stop it manually.
- [x] **Warn (external process):** If the port belongs to something outside infra, warn/fail with PID/command and guidance to free the port or change config.

## Deliverables
- [x] Port inspection helper that identifies goreman vs. external processes (with PID logging and policy classification).
- [x] Integration within the service supervisor (and CLI) so each service checks its port during the ensure phase and applies the policy.
- [x] Enhanced status data (for `infra runtime status` and `/status` page) reflecting running/stopped/conflict states and the last action taken.
- [x] Validation notes covering `go run .`, `infra runtime up --only hugo`, and conflict scenarios (same session crash, different infra session, unrelated process).

## Open Questions / Needs Input
- [x] Scope creep beyond services that bind TCP ports (e.g., background workers without sockets) is deferred (port-bound services only).
- [x] How aggressive should auto-reclaim be in production environments (likely warn-only by default)? (Keeping warn-only policy).

## Next Steps (draft)
- [x] Prototype port ownership detection and conflict classification.
- [x] Wire detection into the service ensure/start flow.
- [x] Update status reporting and validate workflows.

### Validation Notes
- ✅ `go test ./pkg/service/... ./pkg/status/...` – compilation passes for updated packages.
- 🔄 Manual: `go run .` after killing the process to leave goreman PIDs behind – expect auto reclamation log and "Reclaimed stale process…; Service running…" as last action.
- 🔄 Manual: `infra runtime up --only hugo` with another infra session running – expect startup block with PID/command guidance.
- 🔄 Manual: External port holder (`python -m http.server 1313`) before starting Hugo – expect failure with external process remediation guidance.
