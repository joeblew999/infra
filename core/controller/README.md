# Controller Service

The controller exposes a small HTTP API that keeps desired state in sync with
infrastructure providers (Fly, Cloudflare, etc.). It persists the spec to disk
and publishes Server-Sent Events so CLIs and UIs can follow reconciliation in
real time.

## Quickstart
```sh
cd controller

# Run with default spec + address
GOWORK=off go run .

# Custom spec and bind address
GOWORK=off go run . --spec ./spec.yaml --addr 0.0.0.0:4400 \
  --cloudflare-token-file ~/.config/core-cloudflare/api_token

# Stream state changes (SSE) from another terminal
curl -N http://127.0.0.1:4400/v1/events
# or
go run ../cmd/core controller watch --controller 127.0.0.1:4400
```

## API
- `GET /v1/services` — current desired state (JSON)
- `PATCH /v1/services/update` — apply a service update `{ "service": {...} }`
- `GET /v1/events` — SSE stream with JSON payloads: reason, time, desired state

## Configuration
The desired state spec lives in `spec.yaml`. It defines:
- Services with scale/storage/routing definitions
- Cloudflare DNS records (`routing.dns_records`) and load-balancing metadata

When the process exits cleanly it writes the in-memory desired state back to the
same spec file.

## Providers
- **Machines**: pluggable interface for Fly Machines. The current build still
  uses a `NullMachines` stub, to be replaced with a real Fly provider.
- **Routing**: `controller/pkg/providers/cloudflare` integrates with
  `cloudflare-go` to ensure DNS records match the spec when `routing.provider`
  is `cloudflare`. Provide credentials via `--cloudflare-token`,
  `--cloudflare-token-file`, or environment variables
  `CLOUDFLARE_API_TOKEN` / `CLOUDFLARE_API_TOKEN_FILE`.

## Deployment Notes
- The service is compiled as its own Go module; run with `GOWORK=off` to avoid
  picking up the core workspace overrides during local testing.
- Expose the desired spec via volume mount or config map in production so
  updates persist across restarts.
- Credentials (Fly API token, Cloudflare API token, etc.) should be injected via
  environment variables or secrets management; do not commit them to the repo.

Future work includes wiring the Fly Machines provider, emitting richer status
events, and embedding the stream into the TUI/web shells.
