# Process Observability System

Real-time process lifecycle monitoring through NATS JetStream event streaming.

## Overview

The observability system monitors process-compose and publishes lifecycle events (starts, stops, crashes, health changes) to NATS JetStream for real-time monitoring and historical analysis.

## Architecture

```
┌─────────────────┐       ┌──────────────┐       ┌─────────────┐
│ Process-Compose │ ◄───┤ Event Adapter │──────►│ NATS Stream │
│   (Port 28081)  │ poll  │ (2s interval)│ pub   │PROCESS_EVENTS│
└─────────────────┘       └──────────────┘       └──────┬──────┘
                                                         │ sub
                                                         ▼
                                              ┌──────────────────┐
                                              │ Event Consumers  │
                                              │ - Watch CLI      │
                                              │ - GUI/TUI        │
                                              │ - Custom Apps    │
                                              └──────────────────┘
```

## Quick Start

```bash
# Terminal 1: Start the event adapter
go run . stack observe adapter

# Terminal 2: Watch events in real-time
go run . stack observe watch
```

## Commands

### Run Event Adapter

Polls process-compose and publishes events to NATS:

```bash
go run . stack observe adapter

# Options:
--compose-port 28081        # Process Compose API port
--nats-url nats://...      # NATS server URL
--poll-interval 2s          # Poll frequency
```

### Watch Events

Subscribe to events in real-time:

```bash
# Watch all events
go run . stack observe watch

# Watch specific process
go run . stack observe watch --process default/caddy

# Watch specific event type
go run . stack observe watch --type crashed

# JSON output
go run . stack observe watch --json
```

## Event Types

- `started` - Process started or detected running
- `stopped` - Process stopped gracefully
- `crashed` - Process exited with non-zero code
- `restarted` - Process restart detected
- `healthy` - Health probe passed
- `unhealthy` - Health probe failed
- `status_changed` - Process status transition

## NATS Subjects

Events use hierarchical NATS subjects:

```
core.process.{namespace}.{name}.{type}
```

**Examples:**
- `core.process.default.nats.started`
- `core.process.default.caddy.crashed`
- `core.process.default.pocketbase.healthy`

## Event Structure

```json
{
  "type": "started",
  "process": "caddy",
  "namespace": "default",
  "timestamp": "2025-10-16T16:26:18Z",
  "subject": "core.process.default.caddy.started",
  "severity": "info",
  "state": {
    "name": "caddy",
    "namespace": "default",
    "status": "Running",
    "is_running": true,
    "is_ready": "Ready",
    "pid": 12345,
    "restarts": 0
  }
}
```

## Code Structure

Follows the shared/runtime pattern from [pkg/README.md](../pkg/README.md):

```
pkg/shared/observability/     # Implementation
├── adapter.go               # Event adapter (polls & publishes)
├── consumer.go              # Event consumer (subscribe & handle)
└── types.go                 # Event types & schemas

pkg/runtime/observability/   # Re-exports for convenience
└── observability.go         # Re-exports shared types
```

## Integration Example

```go
import "github.com/joeblew999/infra/core/pkg/runtime/observability"

// Create consumer
consumer, err := observability.NewConsumer("nats://127.0.0.1:4222")
if err != nil {
    log.Fatal(err)
}
defer consumer.Close()

// Connect
if err := consumer.Connect(); err != nil {
    log.Fatal(err)
}

// Subscribe to crashes
err = consumer.SubscribeEventType(observability.EventTypeCrashed,
    func(evt observability.Event) error {
        fmt.Printf("CRASH: %s (exit=%d)\n", evt.Process, *evt.ExitCode)
        // Send alert, trigger restart, etc.
        return nil
    })

// Wait for events
consumer.Wait()
```

## JetStream Configuration

- **Stream Name**: `PROCESS_EVENTS`
- **Subjects**: `core.process.>`
- **Retention**: 24 hours
- **Storage**: File-based
- **Delivery**: Last-per-subject (replays recent state)

## References

- **Implementation**: [pkg/shared/observability/](../pkg/shared/observability/)
- **CLI Commands**: [pkg/runtime/cli/stack.go](../pkg/runtime/cli/stack.go) (search "observe")
- **Package Pattern**: [pkg/README.md](../pkg/README.md)
