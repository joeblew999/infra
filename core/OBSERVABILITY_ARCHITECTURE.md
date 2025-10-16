# Process-Compose Event Observability Architecture

**Status**: Design Phase
**Date**: 2025-10-15
**Version**: 1.0

## Research Findings

### Process-Compose Capabilities (v1.64.1)

After examining the process-compose source code, we have:

✅ **WebSocket Support** (`/process/logs/ws`)
- Real-time log streaming
- Multiple process subscription
- Follow mode for continuous streaming
- Offset support for historical logs

✅ **REST API Endpoints**:
```
GET  /processes              → All process states
GET  /process/:name          → Single process state
GET  /project/state          → Project-wide state
GET  /process/logs/:name     → Historical logs
POST /process/start/:name    → Start process
PATCH /process/stop/:name    → Stop process
POST /process/restart/:name  → Restart process
```

✅ **Rich State Information** (`ProcessState`):
```go
- Name, Namespace, Status
- Health (is_ready), HasHealthProbe
- Restarts, ExitCode, Pid
- SystemTime, Age
- CPU, Memory usage
- IsRunning, IsElevated
```

❌ **No Native Event Stream**:
- No SSE endpoint for state changes
- No pub/sub for process events
- Must poll `/processes` for state changes

### **Our Existing Control Plane** ✅

We already have comprehensive process mutation via `stack process`:
```bash
go run . stack process start <name>     # Start stopped process
go run . stack process stop <name>      # Stop running process
go run . stack process restart <name>   # Restart process
go run . stack process scale <name> N   # Scale to N replicas
go run . stack process logs <name>      # View historical logs
go run . stack process info <name>      # Get detailed state
```

**Key Insight**: We have **CONTROL** (write operations). We need **OBSERVABILITY** (read/events).

This enables powerful patterns:
- **Auto-remediation**: Event adapter can call `stack process restart` on crash
- **Health-based scaling**: Scale up/down based on health events
- **Automatic recovery**: Restart unhealthy processes
- **Policy enforcement**: Stop processes that exceed resource limits

## Refined Architecture

### **Complete Control Loop: Observe → Decide → Act**

```
┌─────────────────────────────────────────────────────────────────┐
│                    Process-Compose API                          │
│  ┌────────────────┐              ┌──────────────────┐          │
│  │ WebSocket      │              │ REST API         │          │
│  │ /process/logs  │              │ /processes       │          │
│  │ /ws            │              │ /project/state   │          │
│  └────────────────┘              └──────────────────┘          │
└────────┬────────────────────────────────────┬──────────────────┘
         │                                     │
         │ (log stream)                        │ (poll 2-5s)
         ▼                                     ▼
┌─────────────────────────────────────────────────────────────────┐
│              Event Adapter Service                              │
│  ┌──────────────────────┐      ┌─────────────────────────┐     │
│  │ Log Aggregator       │      │ State Change Detector   │     │
│  │ - Connects to WS     │      │ - Polls /processes      │     │
│  │ - Buffers logs       │      │ - Diffs previous state  │     │
│  │ - Enriches metadata  │      │ - Detects transitions   │     │
│  └──────────────────────┘      └─────────────────────────┘     │
│                    │                        │                   │
│                    └────────┬───────────────┘                   │
│                             ▼                                   │
│              ┌─────────────────────────────┐                    │
│              │   Event Publisher           │                    │
│              │   - Transforms to events    │                    │
│              │   - Enriches with context   │                    │
│              │   - Publishes to NATS       │                    │
│              └─────────────────────────────┘                    │
└──────────────────────────────┬──────────────────────────────────┘
                               │
                               ▼
┌─────────────────────────────────────────────────────────────────┐
│                     NATS JetStream                              │
│  Subject Hierarchy:                                             │
│  • core.process.{name}.started                                  │
│  • core.process.{name}.stopped                                  │
│  • core.process.{name}.healthy                                  │
│  • core.process.{name}.unhealthy                                │
│  • core.process.{name}.restarted                                │
│  • core.process.{name}.crashed                                  │
│  • core.process.{name}.log                                      │
│  • core.project.updated                                         │
│                                                                  │
│  Streams:                                                        │
│  • PROCESS_EVENTS (retention: 1 hour, max: 10k msgs)           │
│  • PROCESS_LOGS (retention: 1 hour, max: 100k msgs)            │
└─────────────────────────────┬───────────────────────────────────┘
                              │
                              │ (subscribe)
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Event Consumers                              │
│  ┌─────────────┐  ┌──────────────┐  ┌────────────────────┐    │
│  │ CLI         │  │ TUI          │  │ Web UI             │    │
│  │ stack watch │  │ Dashboard    │  │ (SSE to browser)   │    │
│  └─────────────┘  └──────────────┘  └────────────────────┘    │
│  ┌─────────────┐  ┌──────────────┐  ┌────────────────────┐    │
│  │ Alerts      │  │ Metrics      │  │ Auto-remediation   │    │
│  │ (webhooks)  │  │ (Prometheus) │  │ (restart policies) │    │
│  └─────────────┘  └──────────────┘  └────────────────────┘    │
└─────────────────────────────────────────────────────────────────┘
```

## Event Schema

### State Change Events

```json
{
  "type": "process.started|stopped|healthy|unhealthy|restarted|crashed",
  "process": "nats",
  "namespace": "default",
  "timestamp": "2025-10-15T20:30:01Z",
  "previous_state": {
    "status": "Stopped",
    "is_running": false,
    "restarts": 0
  },
  "current_state": {
    "status": "Running",
    "is_running": true,
    "pid": 12345,
    "health": "Ready",
    "restarts": 0,
    "cpu": 0.5,
    "mem": 45678912
  },
  "metadata": {
    "duration_since_start": "5s",
    "environment": "production"
  }
}
```

### Log Events

```json
{
  "type": "process.log",
  "process": "pocketbase",
  "namespace": "default",
  "timestamp": "2025-10-15T20:30:05Z",
  "log": {
    "level": "info",
    "message": "[pocketbase] Server listening on :8090",
    "raw": "[pocketbase] Server listening on :8090\n"
  }
}
```

### Project Events

```json
{
  "type": "project.updated",
  "timestamp": "2025-10-15T20:30:10Z",
  "state": {
    "process_num": 3,
    "running_process_num": 3,
    "uptime": "5m30s"
  }
}
```

## Implementation Plan

### Phase 1: Event Adapter Core (Week 1)

**Goal**: Get basic state change detection working

**Components**:
1. **State Poller**
   ```go
   // pkg/runtime/events/poller.go
   type StatePoller struct {
       client    *process.ComposeClient
       interval  time.Duration
       lastState map[string]ProcessState
   }

   func (p *StatePoller) Poll(ctx context.Context) ([]StateChange, error)
   ```

2. **State Differ**
   ```go
   // pkg/runtime/events/differ.go
   func DetectChanges(prev, curr map[string]ProcessState) []StateChange
   ```

3. **Event Publisher**
   ```go
   // pkg/runtime/events/publisher.go
   type NATSPublisher struct {
       nc *nats.Conn
   }

   func (p *NATSPublisher) Publish(event Event) error
   ```

**Files to Create**:
- `pkg/runtime/events/adapter.go` - Main adapter service
- `pkg/runtime/events/poller.go` - State polling
- `pkg/runtime/events/differ.go` - Change detection
- `pkg/runtime/events/publisher.go` - NATS publishing
- `pkg/runtime/events/types.go` - Event types

**Testing**:
- Unit tests for differ logic
- Integration test with mock process-compose
- Verify event publishing to NATS

### Phase 2: Log Streaming (Week 2)

**Goal**: Stream logs via WebSocket and publish to NATS

**Components**:
1. **WebSocket Client**
   ```go
   // pkg/runtime/events/logstream.go
   type LogStreamer struct {
       wsURL    string
       processes []string
   }

   func (ls *LogStreamer) Stream(ctx context.Context) (<-chan LogEvent, error)
   ```

2. **Log Parser**
   ```go
   // Parse log levels from messages
   func ParseLogLevel(message string) (level string, parsed string)
   ```

**Files to Create**:
- `pkg/runtime/events/logstream.go` - WebSocket log streaming
- `pkg/runtime/events/logparser.go` - Log parsing and enrichment

**Testing**:
- Connect to real process-compose WebSocket
- Verify log streaming works
- Test reconnection on disconnect

### Phase 3: CLI Consumer (Week 2-3)

**Goal**: Build `stack watch` command as first consumer

**Command**:
```bash
$ go run ./cmd/core stack watch
→ Watching stack events (Ctrl+C to exit)...

[20:30:01] nats: started (pid 12345)
[20:30:01] nats: healthy
[20:30:05] pocketbase: started (pid 12346)
[20:30:05] pocketbase: healthy
[20:30:12] caddy: restarted (1 restart)
[20:30:13] caddy: unhealthy (connection refused)
[20:30:18] caddy: healthy

$ go run ./cmd/core stack watch --service nats
→ Watching nats events only...

$ go run ./cmd/core stack watch --type health
→ Watching health events only...
```

**Files to Create**:
- `pkg/runtime/cli/stack_watch.go` - Watch command
- `pkg/runtime/events/consumer.go` - NATS consumer helper

**Features**:
- Filter by service name
- Filter by event type
- Colored output
- Follow mode (tail -f style)

### Phase 4: TUI Dashboard (Week 3-4)

**Goal**: Real-time visual dashboard

**UI Layout**:
```
┌─ Process Status ──────────────────────────────────────────────┐
│ nats        ✓ Running  (0 restarts) │ CPU: 0.5%  Mem: 45 MB  │
│ pocketbase  ✓ Healthy  (0 restarts) │ CPU: 1.2%  Mem: 128 MB │
│ caddy       ⚠ Starting (1 restart)  │ CPU: 0.1%  Mem: 32 MB  │
└───────────────────────────────────────────────────────────────┘
┌─ Recent Events ───────────────────────────────────────────────┐
│ [20:30:18] caddy: healthy                                     │
│ [20:30:13] caddy: unhealthy (connection refused)              │
│ [20:30:12] caddy: restarted                                   │
│ [20:30:05] pocketbase: healthy                                │
│ [20:30:05] pocketbase: started                                │
└───────────────────────────────────────────────────────────────┘
┌─ Live Logs (caddy) ───────────────────────────────────────────┐
│ [caddy] Server listening on :2015                             │
│ [caddy] Proxying to http://127.0.0.1:8090                     │
│ [caddy] Health check passed                                   │
│                                                                │
└───────────────────────────────────────────────────────────────┘
```

**Files to Create**:
- `pkg/runtime/cli/stack_dashboard.go` - Dashboard command
- `pkg/runtime/ui/dashboard/` - TUI components

## Key Design Decisions

### 1. **Polling Interval: 2 seconds**
- ✅ Low enough latency for most use cases
- ✅ Doesn't overwhelm process-compose API
- ✅ Configurable via flag

### 2. **Use NATS JetStream for Persistence**
- ✅ Already running in our stack
- ✅ Replay events on consumer restart
- ✅ Multiple consumers don't duplicate load
- ✅ Retention: 1 hour or 10k events (configurable)

### 3. **Hybrid WebSocket + Polling**
- ✅ WebSocket for logs (real-time, already supported)
- ✅ Polling for state (no native event stream)
- ✅ Best of both worlds

### 4. **Event Adapter as Separate Service**
- ✅ Can run standalone or embedded
- ✅ Single point of integration with process-compose
- ✅ Easier to test and maintain

### 5. **Subject Hierarchy**
```
core.process.{name}.{event_type}
core.project.{event_type}
core.health.{name}
core.logs.{name}
```
- ✅ Easy to filter/subscribe
- ✅ Follows NATS best practices
- ✅ Wildcards work: `core.process.*.healthy`

## Questions to Resolve

### 1. **Event Adapter Lifecycle**

**Option A: Embedded in stack up**
- Starts/stops with stack
- Simple deployment
- Coupled to stack

**Option B: Standalone service**
- Independent lifecycle
- Can restart without affecting stack
- Additional process to manage

**Recommendation**: Option A (embedded) for simplicity

### 2. **Event Retention**

**Options**:
- Memory only (ephemeral, fast)
- JetStream (persistent, slower)
- Both (memory + optional persistence)

**Recommendation**: Start with JetStream (1hr retention)
- Can replay on consumer restart
- Good for debugging
- Not too much overhead

### 3. **Performance Tuning**

**Polling interval**:
- 1s = Very responsive, higher load
- 2s = Good balance (recommended)
- 5s = Lower load, might miss quick changes

**Batch size**:
- Process all changes immediately
- Or batch events every N seconds?

**Recommendation**: 2s polling, immediate publishing

### 4. **Error Handling**

What if process-compose API is down?
- Keep trying with exponential backoff
- Publish `adapter.disconnected` event
- Consumers show "stale" indicator

### 5. **Log Parsing**

Do we parse log levels from messages?
```
[nats] INFO: Server listening → level=INFO
[caddy] ERROR: Failed to start → level=ERROR
```

**Recommendation**:
- Yes, parse common patterns
- Fall back to `level=UNKNOWN` if can't parse
- Allow custom parsers per service

## Auto-Remediation Patterns (Using Existing Commands)

Since we already have `stack process` commands, we can build powerful auto-remediation:

### Pattern 1: Auto-Restart on Crash

```go
// Subscribe to crash events
nc.Subscribe("core.process.*.crashed", func(msg *nats.Msg) {
    evt := parseEvent(msg.Data)

    // Policy: Auto-restart critical services
    if isCriticalService(evt.Process) && evt.CurrentState.Restarts < 3 {
        log.Info("Auto-restarting crashed process", "process", evt.Process)

        // Use existing command!
        exec.Command("go", "run", ".", "stack", "process", "restart", evt.Process).Run()

        // Publish remediation event
        publishEvent("core.remediation.restart", evt.Process)
    } else {
        // Too many restarts, alert instead
        sendAlert("Process %s crashed too many times", evt.Process)
    }
})
```

### Pattern 2: Health-Based Recovery

```go
// Subscribe to unhealthy events
nc.Subscribe("core.process.*.unhealthy", func(msg *nats.Msg) {
    evt := parseEvent(msg.Data)

    // Wait 30s to see if it recovers
    time.Sleep(30 * time.Second)

    // Check if still unhealthy
    state := getCurrentState(evt.Process)
    if state.Health != "Ready" {
        log.Warn("Process unhealthy for 30s, restarting", "process", evt.Process)
        exec.Command("go", "run", ".", "stack", "process", "restart", evt.Process).Run()
    }
})
```

### Pattern 3: Resource-Based Scaling

```go
// Subscribe to all process events
nc.Subscribe("core.process.*.*", func(msg *nats.Msg) {
    evt := parseEvent(msg.Data)

    // Check resource usage
    if evt.CurrentState.CPU > 80.0 {
        log.Info("High CPU detected", "process", evt.Process, "cpu", evt.CurrentState.CPU)

        // Scale up if possible
        exec.Command("go", "run", ".", "stack", "process", "scale", evt.Process, "2").Run()
    }
})
```

### Pattern 4: Dependency-Based Restart

```go
// If PocketBase crashes, restart Caddy (proxy depends on it)
nc.Subscribe("core.process.pocketbase.restarted", func(msg *nats.Msg) {
    log.Info("PocketBase restarted, restarting dependent services")

    // Restart Caddy to re-establish proxy connection
    exec.Command("go", "run", ".", "stack", "process", "restart", "caddy").Run()
})
```

### Pattern 5: Policy Enforcement

```go
// Prevent non-critical services from restarting too often
nc.Subscribe("core.process.*.started", func(msg *nats.Msg) {
    evt := parseEvent(msg.Data)

    if !isCriticalService(evt.Process) && evt.CurrentState.Restarts > 5 {
        log.Warn("Non-critical service flapping, stopping it", "process", evt.Process)

        // Stop the flapping service
        exec.Command("go", "run", ".", "stack", "process", "stop", evt.Process).Run()

        // Alert operator
        sendAlert("Stopped flapping service: %s", evt.Process)
    }
})
```

**Key Advantage**: All remediation logic can use our existing, tested `stack process` commands!

## Success Criteria

✅ **Phase 1 Complete When**:
- State changes detected within 2s
- Events published to NATS correctly
- No memory leaks after 24h run

✅ **Phase 2 Complete When**:
- Logs stream in real-time via WebSocket
- Log events published to NATS
- Reconnects automatically on disconnect

✅ **Phase 3 Complete When**:
- `stack watch` command works
- Shows events in real-time
- Filtering works correctly

✅ **Phase 4 Complete When**:
- TUI dashboard displays live state
- Updates smoothly without flicker
- Handles 100+ events/second

## Next Steps

1. **User Approval**: Get feedback on this design
2. **Prototype**: Build Phase 1 minimal viable adapter
3. **Test**: Verify with real stack
4. **Iterate**: Refine based on learnings
5. **Expand**: Add Phase 2-4 features

## Open Questions for Discussion

1. Is 2s polling interval acceptable, or do we need faster?
2. Should event adapter be embedded or standalone?
3. What event retention is needed? (1hr, 24hr, 7d?)
4. Which consumers are highest priority? (CLI, TUI, Web UI?)
5. Do we need log parsing, or raw logs are fine?
6. Should we support event webhooks to external systems?

---

**Status**: Ready for review and refinement
