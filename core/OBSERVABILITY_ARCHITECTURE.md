# Process-Compose Event Observability Architecture

**Status**: Phase 1 Implemented ✅
**Date**: 2025-10-16
**Version**: 2.0 (Polling-Based, No WebSockets)

## Design Principles

1. **Polling-based state detection** - No WebSocket streaming to clients
2. **NATS pub/sub for event distribution** - Decoupled, scalable event delivery
3. **Simple, reliable, observable** - Easy to understand and debug
4. **Integrates with existing control plane** - Works with `stack process` commands

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                    Process-Compose REST API                     │
│                      GET /processes (poll)                      │
└────────────────────────────────┬────────────────────────────────┘
                                 │
                                 │ HTTP Poll (every 2s)
                                 ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Event Adapter Service                        │
│  1. Poll /processes endpoint                                    │
│  2. Diff current vs previous state                              │
│  3. Detect state transitions (started/stopped/crashed/healthy)  │
│  4. Publish events to NATS JetStream                            │
└────────────────────────────────┬────────────────────────────────┘
                                 │
                                 │ NATS Pub/Sub
                                 ▼
┌─────────────────────────────────────────────────────────────────┐
│                   NATS JetStream (Message Broker)               │
│  Stream: PROCESS_EVENTS                                         │
│  Subjects: core.process.{namespace}.{name}.{event_type}         │
│  Retention: 24 hours                                            │
└────────┬────────────┬──────────────┬──────────────┬─────────────┘
         │            │              │              │
         ▼            ▼              ▼              ▼
    ┌────────┐  ┌─────────┐  ┌──────────┐  ┌──────────────┐
    │ watch  │  │   TUI   │  │ Web GUI  │  │ Auto-Remedy  │
    │  CLI   │  │Dashboard│  │Dashboard │  │  Policies    │
    └────────┘  └─────────┘  └──────────┘  └──────────────┘
```

## Event Types

| Event | Trigger | Example Subject |
|-------|---------|----------------|
| `started` | Process became running | `core.process.default.nats.started` |
| `stopped` | Process stopped (exit=0) | `core.process.default.pocketbase.stopped` |
| `crashed` | Process exited non-zero | `core.process.default.caddy.crashed` |
| `restarted` | Restart count increased | `core.process.default.nats.restarted` |
| `healthy` | Health probe → Ready | `core.process.default.pocketbase.healthy` |
| `unhealthy` | Health probe → Not Ready | `core.process.default.caddy.unhealthy` |
| `status_changed` | Status string changed | `core.process.default.nats.status_changed` |

## Implementation Status

### ✅ Phase 1: Core Event System (COMPLETED)

**Files Created**:
- `pkg/observability/events/adapter.go` - Event adapter (polls + publishes)
- `pkg/observability/events/types.go` - Event types and schemas
- `pkg/observability/events/consumer.go` - Consumer helpers
- `pkg/runtime/cli/observe.go` - CLI commands

**Commands Available**:
```bash
# Run event adapter
go run ./cmd/core stack observe adapter

# Watch events in real-time
go run ./cmd/core stack observe watch
go run ./cmd/core stack observe watch --process nats
go run ./cmd/core stack observe watch --type crashed
go run ./cmd/core stack observe watch --json
```

**Features**:
- Polls process-compose every 2s (configurable)
- Detects all state transitions
- Publishes to NATS JetStream
- CLI watch command with filtering
- JSON output for scripting
- Documented in OBSERVABILITY_USAGE.md

### 🔜 Phase 2: TUI Dashboard (PENDING)

Real-time terminal dashboard showing:
- Live process status grid
- Recent events stream
- Health indicators
- Resource usage

Uses NATS subscription (not WebSockets) to get live updates.

### 🔜 Phase 3: Auto-Remediation Policies (PENDING)

Automated recovery using existing `stack process` commands:

```go
// Example: Auto-restart crashed processes
consumer.SubscribeEventType(EventTypeCrashed, func(evt Event) error {
    if restarts < 3 {
        exec.Command("go", "run", ".", "stack", "process", "restart", evt.Process).Run()
    }
    return nil
})
```

Patterns:
- Auto-restart on crash (with backoff)
- Health-based recovery
- Resource-based scaling
- Dependency-aware restart chains

## Event Schema

```json
{
  "type": "started",
  "process": "nats",
  "namespace": "default",
  "timestamp": "2025-10-16T09:53:31Z",
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

## Control Plane Integration

The event system integrates with existing `stack process` commands for full control loop:

**Observe** (Event Adapter):
- Detects crashes, health issues, status changes
- Publishes events to NATS

**Decide** (Policies/Rules):
- Subscribe to events
- Apply business logic (e.g., "restart if crashed < 3 times")

**Act** (Existing Commands):
- `stack process restart <name>` - Restart process
- `stack process stop <name>` - Stop process
- `stack process scale <name> N` - Scale replicas
- `stack process start <name>` - Start process

## Performance Characteristics

- **Polling Overhead**: ~1 HTTP request per 2 seconds (minimal)
- **Memory**: Stores last state snapshot only (~10KB for typical stack)
- **Latency**: 0-2s to detect state change (configurable poll interval)
- **NATS Throughput**: Handles 1000s events/sec easily
- **Event Retention**: 24 hours in JetStream (configurable)

## Why Polling? (Not WebSockets)

✅ **Simple**: No connection management, reconnection logic
✅ **Reliable**: HTTP is stateless, no connection drops
✅ **Observable**: Easy to debug with curl/logs
✅ **Low Overhead**: 2s polling is negligible
✅ **Consistent**: Same pattern for GUI/TUI (poll or NATS events)
✅ **Proven**: Many production systems use polling for observability

❌ WebSockets add complexity:
- Connection management
- Reconnection logic
- State synchronization on reconnect
- Not needed for 2s update cadence

## Usage Examples

See [OBSERVABILITY_USAGE.md](OBSERVABILITY_USAGE.md) for complete guide.

**Quick Start**:
```bash
# Terminal 1: Start stack
go run ./cmd/core stack up

# Terminal 2: Start event adapter
go run ./cmd/core stack observe adapter

# Terminal 3: Watch events
go run ./cmd/core stack observe watch

# Terminal 4: Generate events
go run ./cmd/core stack process restart nats
```

## Future Enhancements

1. **Historical queries**: Query past events from JetStream
2. **Metrics aggregation**: Count crashes/restarts per service
3. **Alerting**: Send notifications on critical events
4. **Multi-stack**: Monitor multiple process-compose instances
5. **Event filtering**: Server-side filtering in event adapter
6. **Log aggregation**: Collect logs via polling `/process/logs/:name`

## Success Criteria

✅ Event adapter polls and detects changes
✅ Events published to NATS JetStream
✅ CLI watch command shows events in real-time
✅ Events contain full process state snapshot
✅ System is simple, reliable, observable
✅ Documentation complete

## References

- Implementation: `pkg/observability/events/`
- Usage Guide: `OBSERVABILITY_USAGE.md`
- CLI Commands: `go run ./cmd/core stack observe --help`
