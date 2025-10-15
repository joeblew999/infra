# Scaling Strategy

This document captures how we scale the runtime today, the limits of Process Compose's
replica model, and the roadmap for multi-region deployments.

## 1. Local (Single Host) Scaling

Process Compose can launch multiple replicas of the same process inside one
project. The CLI and UI now expose `scale` controls, but they are deliberately
constrained:

- **Safe use cases** — worker-style processes (message consumers, background
  jobs) that do not bind to fixed TCP ports. These replicas can fan out work on a
  single host without clashing.
- **Unsafe use cases** — services such as Caddy, PocketBase, or NATS listening on
  well-known ports. Process Compose does not automatically re-map ports for each
  replica, so scaling to `N > 1` on one host will fail or result in port
  conflicts.

**Guardrails**

1. Each service manifest now carries `scalable` and `scale_strategy`
   metadata. They default to `false` / `infra`. Only services that opt in with
   `scalable: true` surface the scale UI controls.
2. The Ultraviolet TUI and Datastar web UI hide the scale prompt when the process
   is not marked scalable, and display a note explaining that scaling must be
   done via infrastructure automation instead.
3. `scale_strategy` currently supports:
   - `local` — safe to run replicas on the same host (worker queues, etc.).
   - `infra` — needs additional hosts/regions; UI disables local scaling and
     points operators to infra-level tooling.

## 2. Cross-Host / Multi-Region Scaling

Caddy, PocketBase, NATS, etc. require independent hosts. We plan to scale them
via infrastructure primitives instead of Process Compose replicas:

| Layer | Responsibility |
| --- | --- |
| Process Compose project | One stack per host. Each project keeps `replicas: 1` for port-bound services. |
| Orchestrator Controller *(planned)* | Watches desired replica counts per service/region and provisions additional Process Compose projects on demand. |
| Fly.io Deployments | Hosts the additional Process Compose stacks (one per region). Fly load-balancing routes external traffic to regional Caddy instances. |
| Datastar/TUI UI | Displays aggregated counts per region and triggers infra-level scaling through the controller (future). |

### Roadmap

1. **Controller stub** — new service responsible for deploying Process Compose
   stacks on Fly (one per region) using the existing tooling pipeline. It will
   expose an API the UI can call.
2. **Registry of scalable services** — manifest metadata so the controller knows
   whether a service can be replicated per host or requires infra scaling.
3. **UI enhancements** — show region-aware replica counts (e.g. Caddy@iad=2,
   Caddy@fra=1) and provide buttons to request additional regional capacity.
4. **DNS/load balancing** — ensure regional Caddy instances share certificate
   configuration and feed traffic to their local PocketBase/NATS peers.

## 3. Current Best Practices

- Use the built-in scale control only for stateless workers tied to message
  queues or background jobs.
- For HTTP services, prefer provisioning additional Process Compose stacks (or
  Fly apps) and adjust routing at the infrastructure layer.
- Keep manifests deterministic; if local replicas are necessary, ensure each
  replica receives unique ports via environment overrides before enabling the
  UI control.

This document will evolve as the controller scaffolding and cross-region
automation are implemented.

The controller spec lives in `controller/spec.yaml`. Use the CLI to inspect it:

```sh
go run ./cmd/core scale show --controller http://127.0.0.1:4400 --file controller/spec.yaml

# apply a service override from a dedicated YAML file
go run ./cmd/core scale set --controller http://127.0.0.1:4400 --file controller/service-pocketbase.yaml
```

### Runtime vs. Controller Responsibilities

- The **runtime binary** (`cmd/core`) never talks to Fly.io or Cloudflare
  directly. It focuses on local Process Compose control and the CLI/TUI/Web
  surfaces.
- Fly/Cloudflare integrations live in a separate **controller service** (part
  of the tooling plane) that owns the necessary API tokens, reads the desired
  state, and reconciles infrastructure.
- UI and CLI entry points that change global scale will call the controller
  API, not the runtime, keeping credentials and provisioning logic isolated.

## 4. Spec-Driven vs. Metrics-Driven Scaling

Two signals influence scaling decisions:

1. **Desired state (user spec).** A declarative configuration describing how
   many replicas to run, in which regions, and any minima/maxima. This spec is
   authoritative—metrics-driven scaling must converge back toward it.
2. **Observed load (metrics).** Telemetry that indicates demand (queue depth,
   CPU, latency). Metrics can temporarily push the system above the desired
   state, but only within bounds defined by the spec.

### Planned Controller Responsibilities

- Maintain the desired state (persisted spec per service/region).
- Reconcile the spec with reality by creating/destroying Process Compose stacks
  (via Fly API) and keeping per-host manifests at `replicas: 1` for port-bound
  services.
- Optionally respond to metrics by scaling within `[spec.min, spec.max]` and
  recording overrides for later reconciliation.
- Publish status back to the UI (current replicas, pending scale actions,
  deviations from desired state).

### UI/UX Implications

- Detail panes will display the declared spec (e.g., “desired: 2 in iad, 1 in
  fra; autoscale enabled up to 4”).
- Operators may request permanent spec changes (write to desired state) or
  temporary bursts (“scale out to 4 for 30 minutes”). Each control routes to the
  controller, not Process Compose directly.
- Local Process Compose scaling remains available only for development
  scenarios; production controls call into the controller.

### Fly.io Considerations

- Use Machines for rapid spin-up/down; they support scale-to-zero via idle
  shutdown. We'll explore mapping each Process Compose stack to a machine or VM.
- Networking: ensure regional Caddy instances share certificates via Fly’s
  global proxy while pointing to regional PocketBase/NATS peers.
- Secrets/config: decide whether specs are stored alongside controller state or
  injected via Fly config volumes.

Next steps include drafting the controller module (API shape, reconciliation
loop) and defining the schema for the desired state document.
