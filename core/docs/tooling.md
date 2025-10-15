# Tooling Workflow Guide

> Canonical reference for the `core-tool` CLI. Keep this document in sync with
> behaviour so agents know exactly how to run Fly + Cloudflare deployments.

The tooling module provides an idempotent deployment workflow plus lower-level
commands for advanced scripting. Start here for end-to-end instructions, then
refer to package docs when extending the CLI or embedding the orchestrator in
other surfaces.

## Workflows

### One-Command Deploy
```sh
go run ./tooling workflow deploy \
  --profile fly \
  --app <fly-app> \
  --org <fly-org> \
  --repo registry.fly.io/<fly-app>
```
This single command:
1. Runs `workflow setup` (render configs, ensure Fly & Cloudflare tokens).
2. Executes the local smoke build (`release --profile local --no-run`).
3. Publishes & deploys to Fly (`release --profile fly`).

Re-run it anytime; existing configs/tokens are reused and missing pieces are
created on the fly.

### Supporting Workflows
```sh
# Prepare configs and tokens without building/deploying
# Optional flags: --skip-fly, --skip-cloudflare, --no-browser
go run ./tooling workflow setup --profile fly --app <fly-app> --org <fly-org>

# Verify credentials against providers (Fly app/org + Cloudflare zone)
go run ./tooling workflow verify --profile fly \
  --app <fly-app> --org <fly-org> --zone <cloudflare-zone>

# Run Fly-only or Cloudflare-only auth
go run ./tooling auth fly [--no-browser]
go run ./tooling auth fly verify
go run ./tooling auth cloudflare [--no-browser]
go run ./tooling auth cloudflare verify
```

## Configuration

### Environment Variables
- `CORE_FLY_APP`, `CORE_FLY_ORG`, `CORE_FLY_REGION` – override Fly settings
  supplied by the active profile.
- `CORE_KO_REPOSITORY` – change the container registry target.
- `CORE_TAG_TEMPLATE` – customise ko image tags.
- `CORE_FLY_TOKEN_PATH`, `CORE_CLOUDFLARE_TOKEN_PATH` – relocate cached tokens.
- `CORE_FLY_API_BASE` – point at an alternative Fly API (staging, etc.).

### Profiles
Two profiles ship with the repo:
- `local` – Docker-friendly local runs (`SupportsDocker=true`).
- `fly` – Production Fly.io deploys.

Select with `--profile` or `CORE_TOOLING_PROFILE`. Extend profile definitions in
`pkg/shared/config/tooling.go` when new environments are introduced.

### Tokens & Secrets
- Tokens are written into `.data/core/secrets/{fly,cloudflare}/` at the repo root
  so every surface (CLI, GUIs, background services) reads the same cache.
- Provider preferences live alongside the secrets: `.data/core/fly/settings.json`
  and `.data/core/cloudflare/settings.json`.
- Override locations via environment variables listed above or the shared
  `--path` flag on the `auth` subcommands.
- Delete the files under `.data/core/` and rerun `workflow setup` (or the
  individual `auth` commands) to rotate credentials.
- The Cloudflare tooling also snapshots tokens per-origin under
  `.data/core/cloudflare/tokens/{manual,bootstrap}` while tracking the active
  token via `.data/core/cloudflare/active_token`, so you can switch between
  manual and bootstrapped credentials without losing the previous one.

### Cloudflare Token Requirements
- `Zone:DNS:Edit` – manage DNS records.
- `Zone:Zone:Read` – list zones for selection.
- `Account:R2:Edit` – manage R2 buckets (optional but recommended).
Generate tokens at <https://dash.cloudflare.com/profile/api-tokens>.

### Cloudflare Authentication Options
- `go run ./tooling auth cloudflare --token <api-token>` – skip prompts by
  pasting a pre-generated token. Validation checks ensure required scopes are
  present.
- `go run ./tooling auth cloudflare bootstrap --global-key <key> --email <email>` –
  hand the CLI a Cloudflare global API key once and it will:
  1. Prompt for the zone/hostname (same as the manual flow).
  2. Discover the required permission groups (Zone Read, DNS Edit, optional R2).
  3. Create a scoped API token named `core-tooling-<timestamp>` with just those
     permissions.
  4. Save the scoped token to `.data/core/secrets/cloudflare/api_token` and
     discard the global key.
  Keeping `--no-browser` unset opens the Cloudflare Global API Key page in your
  browser so you can copy the value directly into the CLI prompt.
- Add `--include-r2` to request R2 permissions or `--path` to direct the token to an
  alternate location.

## Manual Commands (Advanced)
The underlying subcommands are still available when you need fine-grained
control:

```sh
# Render configuration templates only
go run ./tooling config init --profile fly --app <fly-app> --org <fly-org>

# Cache Fly token (launches browser unless --no-browser is set)
go run ./tooling auth --profile fly

# Cache Cloudflare token (prompts for paste)
go run ./tooling auth cloudflare --profile fly

# Run local release flow without Docker
go run ./tooling release --profile local --no-run

# Publish & deploy using explicit flags
go run ./tooling release --profile fly --app <fly-app> --org <fly-org> --repo registry.fly.io/<fly-app>
```

Use these when scripting CI or experimenting with new profiles; otherwise the
`workflow` wrapper should be your go-to.

## Architecture & Integration
- `tooling/pkg/orchestrator` coordinates auth, deploy, and Cloudflare DNS.
  Construct it via `pkg/app.New()` for CLI, GUI, or TUI surfaces.
- Use `ProgressEmitter` or the provided `StreamAdapter` to stream updates to
  logs, SSE endpoints, or WebSockets.
- Read-only provider snapshots are exposed via `tooling/pkg/providers` so other
  services can show Fly/Cloudflare status using the cached credentials.
- Keep presentation layers thin—delegate authentication and release work to the
  orchestrator rather than re-implementing flows.

## Troubleshooting
- **Fly deploy fails with "You must be authenticated"**: rerun `workflow setup`
  (or `auth --profile fly`) to refresh the token. Ensure the token has access to
  the target organization/app.
- **Cloudflare zone access denied**: confirm the token has DNS edit rights for
  the zone and rerun `workflow setup --skip-fly`.
- **Dirty repository errors from ko**: commit or stash generated files, then
  rerun `workflow deploy`. You can pass `--force` to `workflow setup` if you need
  templates rewritten.

For more advanced scenarios (custom profiles, CI usage) see the inline comments
in `pkg/shared/config/tooling.go` and the source under `tooling/internal/cli`.

## Development Checklist
- `go test ./tooling/...`
- `go build -o core-tool .`
- Update this document alongside behaviour changes; keep `tooling/README.md`
  limited to build instructions and a pointer back here.
