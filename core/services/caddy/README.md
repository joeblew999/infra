# Caddy Service

Embedded Caddy runner with the standard HTTP modules plus our extras
(`caddy-dns/acmedns`, `caddy-l4`). The embedded stack uses this binary the same way locally and on Fly.

Artifacts:
- `service.json` — manifest storing default ports, env overrides, and proxy target
- `service.go` — embedded runner that loads the manifest and starts Caddy via the
  Go API
- `core/cmd/caddy/main.go` — dedicated binary entrypoint

## CLI Usage

```bash
# Dedicated binary (same locally and on Fly)
go run ./core/cmd/caddy -- --help

# Run via CLI wrapper
core caddy run

# Inspect manifest metadata
core caddy spec
```

## Modules Included

- Standard HTTP modules (`github.com/caddyserver/caddy/v2/modules/standard`)
- `github.com/caddy-dns/acmedns`
- `github.com/mholt/caddy-l4`

Adjust `service.go` imports to add or remove modules as needed.


