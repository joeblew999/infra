# Process Observability Usage Guide

This guide shows how to use the new process observability system to monitor process-compose events in real-time.

## Architecture

The observability system has three main components:

1. **Event Adapter** - Polls process-compose and publishes events to NATS JetStream
2. **NATS JetStream** - Message broker for pub/sub event distribution
3. **Consumers** - CLI tools, dashboards, or automated remediation that subscribe to events

```
process-compose (REST API)
         ↓
   Event Adapter (polls every 2s, detects changes)
         ↓
   NATS JetStream (pub/sub)
         ↓
    ┌────┴─────┬─────────┬──────────┐
    ↓          ↓         ↓          ↓
  watch    dashboard  alerts  auto-remediation
```

## Quick Start

### 1. Start the Stack

```bash
# Start process-compose with NATS, PocketBase, and Caddy
go run ./cmd/core stack up
```

### 2. Run the Event Adapter

In a new terminal:

```bash
# Start the event adapter (polls process-compose, publishes to NATS)
go run ./cmd/core stack observe adapter

# Output:
# Event adapter running...
#   Process Compose: http://127.0.0.1:28081
#   NATS: nats://127.0.0.1:4222
#   Poll interval: 2s
```

The adapter will immediately detect all running processes and publish `started` events for them.

### 3. Watch Events in Real-Time

In another terminal:

```bash
# Watch all process events
go run ./cmd/core stack observe watch

# Output (example):
# Watching events: core.process.>
# 09:53:31.123 ℹ️  default/caddy started
# 09:53:31.124 ℹ️  default/nats started
# 09:53:31.125 ℹ️  default/pocketbase started
```

## Event Types

The system publishes the following event types:

| Event Type | Description | Severity |
|------------|-------------|----------|
| `started` | Process started or became running | Info |
| `stopped` | Process stopped gracefully (exit=0) | Info |
| `crashed` | Process exited with non-zero code | Error |
| `restarted` | Process restart count increased | Info |
| `healthy` | Health probe changed to Ready | Info |
| `unhealthy` | Health probe changed to not Ready | Warning |
| `status_changed` | Process status changed | Debug |

## NATS Subject Hierarchy

Events are published to subjects following this pattern:

```
core.process.{namespace}.{name}.{event_type}
```

Examples:
- `core.process.default.nats.started`
- `core.process.default.pocketbase.crashed`
- `core.process.default.caddy.healthy`

## Watch Command Options

### Watch All Events

```bash
go run ./cmd/core stack observe watch
```

### Watch Specific Process

```bash
# Watch only NATS events
go run ./cmd/core stack observe watch --process nats

# Watch only default/pocketbase events
go run ./cmd/core stack observe watch --process default/pocketbase
```

### Watch Specific Event Type

```bash
# Watch only crash events across all processes
go run ./cmd/core stack observe watch --type crashed

# Watch only health events
go run ./cmd/core stack observe watch --type unhealthy
```

### Combine Filters

```bash
# Watch crashes for specific process
go run ./cmd/core stack observe watch --process pocketbase --type crashed
```

### JSON Output

```bash
# Get events as JSON for piping to other tools
go run ./cmd/core stack observe watch --json
```

## Adapter Configuration

The adapter can be configured with flags:

```bash
go run ./cmd/core stack observe adapter \
  --compose-port 28081 \            # Process-compose API port
  --nats-url nats://127.0.0.1:4222 \  # NATS server URL
  --poll-interval 2s                 # How often to poll for changes
```

## Testing the System

### 1. Generate Events by Restarting a Process

```bash
# In terminal 1: watch events
go run ./cmd/core stack observe watch

# In terminal 2: restart pocketbase
go run ./cmd/core stack process restart pocketbase

# Terminal 1 will show:
# 10:15:42.123 ℹ️  default/pocketbase restarted (count=1)
# 10:15:42.456 ℹ️  default/pocketbase stopped
# 10:15:43.789 ℹ️  default/pocketbase started
# 10:15:44.012 ℹ️  default/pocketbase healthy
```

### 2. Generate Crash Events

```bash
# Stop a process to generate stopped event
go run ./cmd/core stack process stop nats

# Watch will show:
# 10:20:15.123 ℹ️  default/nats stopped (exit=0)

# Start it again
go run ./cmd/core stack process start nats

# Watch will show:
# 10:20:20.456 ℹ️  default/nats started
```

## Event Schema

Each event contains:

```json
{
  "type": "started",
  "process": "nats",
  "namespace": "default",
  "timestamp": "2025-10-16T09:53:31.123Z",
  "subject": "core.process.default.nats.started",
  "severity": "info",
  "state": {
    "name": "nats",
    "namespace": "default",
    "status": "Running",
    "is_ready": "Ready",
    "has_ready_probe": true,
    "restarts": 0,
    "exit_code": 0,
    "is_running": true,
    "replicas": 1
  }
}
```

## Programmatic Usage

### Subscribe to Events in Go Code

```go
package main

import (
	"fmt"
	"github.com/joeblew999/infra/core/pkg/observability/events"
)

func main() {
	// Create consumer
	consumer, err := events.NewConsumer("nats://127.0.0.1:4222")
	if err != nil {
		panic(err)
	}
	defer consumer.Close()

	if err := consumer.Connect(); err != nil {
		panic(err)
	}

	// Subscribe to all crashes
	consumer.SubscribeEventType(events.EventTypeCrashed, func(evt events.Event) error {
		fmt.Printf("CRASH: %s exited with code %d\n", evt.Process, *evt.ExitCode)
		return nil
	})

	// Subscribe to specific process
	consumer.SubscribeProcess("pocketbase", func(evt events.Event) error {
		fmt.Printf("PocketBase event: %s\n", evt.String())
		return nil
	})

	consumer.Wait()
}
```

### Auto-Remediation Example

```go
// Auto-restart processes that crash (up to 3 times)
restartCount := make(map[string]int)

consumer.SubscribeEventType(events.EventTypeCrashed, func(evt events.Event) error {
	count := restartCount[evt.Process]

	if count < 3 {
		log.Printf("Auto-restarting %s (attempt %d/3)", evt.Process, count+1)

		cmd := exec.Command("go", "run", ".", "stack", "process", "restart", evt.Process)
		if err := cmd.Run(); err != nil {
			log.Printf("Failed to restart: %v", err)
			return err
		}

		restartCount[evt.Process] = count + 1
	} else {
		log.Printf("Process %s crashed too many times, giving up", evt.Process)
	}

	return nil
})
```

## Deployment Considerations

### Running Adapter as a Service

For production, run the adapter as a background service:

```bash
# Using process-compose (recommended)
# Add to process-compose.yaml:
processes:
  event-adapter:
    command: go run ./cmd/core stack observe adapter
    availability:
      restart: always

# Or run detached
nohup go run ./cmd/core stack observe adapter > adapter.log 2>&1 &
```

### NATS Retention

By default, events are retained for **24 hours** in JetStream. This is configurable in the adapter code:

```go
// In pkg/observability/events/adapter.go
MaxAge: 24 * time.Hour,  // Change this to adjust retention
```

### Performance

- **Poll interval**: Default 2s, adjustable via `--poll-interval`
- **Memory**: Minimal overhead, stores only last state snapshot
- **Network**: Polls process-compose REST API every 2s (very lightweight)
- **NATS**: Events are small (~1KB each), JetStream handles thousands/sec easily

## Troubleshooting

### Adapter Won't Start

```bash
# Check if NATS is running
go run ./cmd/core stack status

# Check if process-compose is accessible
curl http://127.0.0.1:28081/processes
```

### Events Not Appearing

```bash
# Check JetStream stream exists
# Install nats CLI: go install github.com/nats-io/natscli/nats@latest
nats stream ls

# Should show: PROCESS_EVENTS

# Check stream info
nats stream info PROCESS_EVENTS
```

### Watch Shows No Events

```bash
# Verify adapter is running and publishing
# Look for debug logs like:
# {"level":"debug","subject":"core.process.default.nats.started",...}

# Generate a test event by restarting something
go run ./cmd/core stack process restart nats
```

## Next Steps

See [OBSERVABILITY_ARCHITECTURE.md](OBSERVABILITY_ARCHITECTURE.md) for:
- Detailed architecture design
- Future phases (WebSocket logs, TUI dashboard)
- Auto-remediation patterns
- Implementation plan
