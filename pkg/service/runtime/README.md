# Service Runtime

This package coordinates the supervised service startup orchestrated by `infra runtime`.

## Port conflict handling

The runtime inspects each declared service port before startup and records what happened so the CLI and `/status` dashboard can show the "last action" taken.

- **Reclaim:** In development (`config.IsDevelopment()`), if the PID belongs to the same goreman-managed service we send it a shutdown, wait up to three seconds for the socket to clear, then continue. Successful reclaims are reported as the last action (e.g. "Reclaimed stale process …; Service running on port 1313").
- **Warn (other infra session):** When another infra supervisor owns the port we abort startup, surface the PID/command, and tell the user to run `infra runtime down` in that session.
- **Warn (external process):** For non-infra processes we abort with the owning PID/command and the guidance to stop the process or change infra's port configuration.

If the runtime cannot inspect a port or a start/ensure hook fails, the error is also captured as the last action so it shows up immediately in `infra runtime status` and on the status page.

## Dynamic reverse proxy configuration

Every service spec can publish HTTP routes (path → target). During `infra runtime up` the orchestrator aggregates these descriptors, generates a Caddyfile, and launches Caddy under goreman supervision. When services start later (for example Bento or PocketBase) they notify the runtime, which regenerates the Caddyfile and issues a zero-downtime `caddy reload`. This keeps HTTPS and proxy routes in sync without manual edits.
