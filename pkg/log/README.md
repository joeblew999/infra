# pkg/log

Lightweight wrapper around Go's `slog` that centralizes configuration and fans logs out to multiple destinations (stdout, files, NATS, etc.) using the `samber/slog-multi` adaptors.

## Defaults

The active environment comes from either the `--env` flag or `ENVIRONMENT`. With no flag the CLI defaults to `production`, but internally we treat anything other than `production` as development (you can pass `--env=development` for clarity).

- **Development**: JSON logs to `stdout` at info level.
- **Production**: JSON logs to `stdout` **and** NATS (`nats://localhost:4222`, subject `config.NATSLogStreamSubject`).
- CLI overrides (level/format/destinations) are opt-in and only affect the current run:
  ```bash
  go run . --log-level=debug --log-format=text --log-output=stdout
  go run . --log-output=stdout --log-output=file=.data/logs/infra.log
  go run . --log-output=nats://localhost:4222?subject=logs.infra
  ```
  `--log-output` may be repeated. Supported forms: `stdout`, `stderr`, `file=/path/to.log`, `nats`, `nats:<subject>`, `nats=custom.subject`, or a full `nats://` URL with optional `?subject=` query. When the CLI specifies destinations, they replace environment defaults.

## Config File (optional)

Place `infra.log.json` at the repo root to declare destinations in JSON. The schema matches `log.MultiConfig`:

```json
{
  "destinations": [
    {"type": "stdout", "level": "info", "format": "json"},
    {"type": "file", "format": "json", "path": ".data/logs/infra.log"},
    {"type": "nats", "level": "info", "url": "nats://localhost:4222", "subject": "logs.infra"}
  ]
}
```

CLI flags take precedence over this file. When the file omits `level` or `format`, we fall back to the environment defaults above.

## Runtime Reconfiguration

`log.ReconfigureMultiLogger(config)` swaps handlers without restarting the process, so long-running services can re-read configuration or respond to admin commands.

## Package Layout

- `log.go` – thin wrapper exposing `InitLogger` and convenience helpers (`Info`, `Warn`, etc.).
- `multi.go` – multi-destination setup, CLI/config parsing, and NATS integration.
- `runtime.go` – thread-safe setter/getter for the active logger, used by the reconfigure path.

## Downstream Usage

Every package should import `github.com/joeblew999/infra/pkg/log` and use `log.Info`, `log.Warn`, etc. Supervised subprocesses (via `pkg/goreman`) stream their stdout/stderr back through this package, so all logs share the same routing and structure.
