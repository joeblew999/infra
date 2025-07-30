# log

We have wrapped the golang slog using the same interface.

This is so that from here, so can use the slog adapters from: https://github.com/samber?tab=repositories

## Smart Defaults

`go run .` uses **JSON to stdout, info level** as the smart default. This provides structured logs that work everywhere without configuration.

## Configuration Override

Create `infra.log.json` in the same directory to override the default. The file uses the same JSON structure as the multi-destination configuration.

## Multi-Destination Logging

Instead of choosing just one destination for your logs, you can send them to several places at once.

https://github.com/samber/slog-multi lets you write the same log to multiple destinations simultaneously! This is perfect for your multi-device monitoring setup.

## Runtime Configuration Support

Logs can be reconfigured at runtime without restarting the application. Use `ReconfigureMultiLogger(config)` to change destinations, formats, and levels on-the-fly.

## Architecture

The package is split into three layers:

- **log.go**: Original simple logger (backward compatible)
- **multi.go**: Multi-destination logging with slog-multi + config loading
- **runtime.go**: Runtime reconfiguration support

## Adapters

### NATS

https://github.com/samber/slog-nats is for writing logs to Nats Jetstream, so we get a global view of everything in real time. Can even just use Nats Jetstream with no storage to make it fast.

example: https://github.com/samber/slog-nats/blob/main/example/example.go

### Quickwit

https://github.com/samber/slog-quickwit is for writing logs to Quickwit, so we can search logs over time. Logs are stored in FS or S3 apparently.

example: https://github.com/samber/slog-quickwit/blob/main/example/example.go

## Configuration

Configuration is JSON-based and supports multiple destination types with individual settings for each destination.

Supported destination types:
- stdout
- stderr  
- file
- nats (planned)
- quickwit (planned)

Each destination can have its own:
- Log level (debug, info, warn, error)
- Output format (json, text)
- Custom settings per destination type

## Files

- **infra.log.json**: Optional configuration file for multi-destination logging
- **No file**: Uses smart default (JSON to stdout, info level)

## Usage Patterns

**Default**: `go run .` → JSON to stdout, info level
**Configured**: Create `infra.log.json` → Multi-destination logging
**Dynamic**: Use `ReconfigureMultiLogger()` for runtime changes

All approaches maintain the same logging interface - no code changes needed when switching between configurations.