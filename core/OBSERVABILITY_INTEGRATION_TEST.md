# Observability Integration Testing Guide

This document describes how to test the integration between the observability event system and the GUI/TUI.

## What Was Integrated

**Before**:
- GUI/TUI polled process-compose for current state
- Event log showed generic sync messages and user actions
- No visibility into state transitions or crash history

**After**:
- GUI/TUI still polls for current state (unchanged)
- **PLUS** subscribes to observability events from NATS
- Rich event log with icons, severity levels, and transition history
- 100 event retention (up from 10)

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│         Event Adapter (background process)              │
│  Polls process-compose → Detects changes → NATS        │
└────────────────────┬────────────────────────────────────┘
                     │ NATS pub/sub
                     ▼
┌─────────────────────────────────────────────────────────┐
│              GUI/TUI Live Store                         │
│  ┌──────────────────┐     ┌─────────────────────┐      │
│  │ Compose Polling  │     │  Event Subscription │      │
│  │ (current state)  │     │  (transitions)      │      │
│  └──────────────────┘     └─────────────────────┘      │
│              │                       │                  │
│              └───────┬───────────────┘                  │
│                      ▼                                  │
│              Unified Event Log                          │
│         (polling + observability events)                │
└─────────────────────────────────────────────────────────┘
```

## Prerequisites

Ensure you have:
1. **Stack running**: `go run ./cmd/core stack up`
2. **Event adapter running**: `go run ./cmd/core stack observe adapter`
3. NATS, PocketBase, Caddy all healthy

## Test Scenarios

### Test 1: Basic Integration - Events Appear in GUI

**Goal**: Verify observability events flow into GUI event log

**Steps**:
```bash
# Terminal 1: Start stack
go run ./cmd/core stack up

# Terminal 2: Start event adapter
go run ./cmd/core stack observe adapter

# Terminal 3: Start web GUI with live mode
go run ./cmd/core web --live

# Open browser to http://127.0.0.1:3400
```

**Expected Results**:
- ✅ Event log shows icons: ▶️ ⏹️ ❌ 🔄 ✅ ⚠️
- ✅ Events like: "▶️ default/nats started"
- ✅ Events like: "✅ default/pocketbase healthy"
- ✅ Severity levels: info, warning, error
- ✅ Timestamps for each event

**Failure Modes**:
- ❌ "Warning: failed to start event stream" → NATS not running
- ❌ No events with icons → Event adapter not running
- ❌ Only sync messages → Integration not working

---

### Test 2: User Actions Generate Events

**Goal**: Verify user mutations trigger observability events showing completion

**Steps**:
```bash
# With GUI running from Test 1...

# In browser or via CLI:
go run ./cmd/core stack process restart nats
```

**Expected Event Timeline**:
```
15:04:07 ✅ default/nats healthy
15:04:06 ▶️ default/nats started
15:04:05 🔄 default/nats restarted (count=1)
15:04:03 ℹ️ process nats restart triggered  ← User action
15:04:00 ℹ️ process-compose sync @ 15:04:00
```

**What to Verify**:
- ✅ User action logged: "process nats restart triggered"
- ✅ Restart event appears: "🔄 default/nats restarted"
- ✅ Started event appears: "▶️ default/nats started"
- ✅ Healthy event appears: "✅ default/nats healthy"
- ✅ Timeline shows complete story

---

### Test 3: Crash Detection

**Goal**: Verify crashes are detected and logged with proper icons

**Steps**:
```bash
# Force a crash by stopping a process
go run ./cmd/core stack process stop pocketbase

# Then start it again
go run ./cmd/core stack process start pocketbase
```

**Expected Events**:
```
15:05:12 ▶️ default/pocketbase started
15:05:10 ⏹️ default/pocketbase stopped (exit=0)
15:05:08 ℹ️ process pocketbase stopping  ← User action
```

**Or simulate a real crash** (if possible):
```bash
# Kill process directly (if you can get PID)
kill -9 <pid>
```

**Expected Events**:
```
15:06:15 ▶️ default/pocketbase started    ← Automatic restart
15:06:14 🔄 default/pocketbase restarted (count=1)
15:06:13 ❌ default/pocketbase crashed (exit=9)
```

---

### Test 4: Health Transitions

**Goal**: Verify health probe transitions are captured

**Steps**:
```bash
# Restart a service with health probes
go run ./cmd/core stack process restart caddy
```

**Expected Events**:
```
15:07:10 ✅ default/caddy healthy        ← Health probe passed
15:07:08 ▶️ default/caddy started
15:07:06 🔄 default/caddy restarted (count=1)
```

**If health probe fails** (misconfigured service):
```
15:08:05 ⚠️ default/someservice unhealthy (health=NotReady)
```

---

### Test 5: Event Log Retention

**Goal**: Verify 100 event retention (up from 10)

**Steps**:
```bash
# Generate many events by restarting multiple services
for i in {1..20}; do
  go run ./cmd/core stack process restart nats
  sleep 3
done
```

**Expected Results**:
- ✅ Event log scrolls with many events
- ✅ At least 60-80 events visible (20 restarts × 3-4 events each)
- ✅ Oldest events truncated at 100
- ✅ No performance degradation

---

### Test 6: TUI Integration

**Goal**: Verify TUI also shows observability events

**Steps**:
```bash
# With stack and adapter running...
go run ./cmd/core tui --live

# In another terminal, trigger events:
go run ./cmd/core stack process restart pocketbase
```

**Expected Results**:
- ✅ TUI event log updates in real-time
- ✅ Icons displayed (if terminal supports emojis)
- ✅ Same events as web GUI
- ✅ Timestamps synchronized

---

### Test 7: Event Adapter Failure

**Goal**: Verify graceful degradation when adapter not running

**Steps**:
```bash
# Start GUI WITHOUT event adapter running
pkill -f "observe adapter"

# Start web GUI
go run ./cmd/core web --live
```

**Expected Results**:
- ✅ Warning message: "failed to start event stream"
- ✅ GUI still works (polling continues)
- ✅ Event log shows sync messages
- ✅ No observability events (adapter not running)

**Then start adapter**:
```bash
# Start adapter
go run ./cmd/core stack observe adapter

# Refresh browser or wait
```

**Expected Results**:
- ❌ Events don't appear (connection already failed)
- ℹ️ Must restart GUI to reconnect

**Improvement Needed**: Add reconnection logic (future enhancement)

---

### Test 8: Multiple Event Types

**Goal**: Verify all 7 event types appear correctly

**Event Type Checklist**:
- ✅ `started` → ▶️ icon, info severity
- ✅ `stopped` → ⏹️ icon, info severity
- ✅ `crashed` → ❌ icon, error severity
- ✅ `restarted` → 🔄 icon, info severity
- ✅ `healthy` → ✅ icon, info severity
- ✅ `unhealthy` → ⚠️ icon, warning severity
- ✅ `status_changed` → 📊 icon, debug severity

**How to Trigger Each**:
```bash
# started
go run ./cmd/core stack process start <name>

# stopped
go run ./cmd/core stack process stop <name>

# crashed (simulate)
kill -9 <pid>

# restarted
go run ./cmd/core stack process restart <name>

# healthy (automatic after health probe)
# Just wait for health probe to pass

# unhealthy (misconfigure health probe)
# Modify process-compose.yaml temporarily

# status_changed (automatic)
# Happens during state transitions
```

---

## Event Format Examples

### User Actions (current behavior, unchanged)
```
15:04:03 [info] process nats restart triggered
15:04:01 [info] process pocketbase stopping
15:03:58 [info] process caddy scaled to 2
```

### Observability Events (new behavior)
```
15:04:07 [info] ✅ default/nats healthy
15:04:06 [info] ▶️ default/nats started
15:04:05 [info] 🔄 default/nats restarted (count=1)
15:04:03 [error] ❌ default/pocketbase crashed (exit=1)
15:04:01 [warning] ⚠️ default/caddy unhealthy (health=NotReady)
```

### Combined Timeline
```
15:05:10 [info] ✅ default/nats healthy          ← Observability
15:05:08 [info] ▶️ default/nats started          ← Observability
15:05:06 [info] 🔄 default/nats restarted (count=1) ← Observability
15:05:05 [info] process nats restart triggered  ← User action
15:05:03 [info] process-compose sync @ 15:05:03 ← Polling
```

---

## Troubleshooting

### Events Not Appearing

**Symptom**: No observability events with icons

**Checks**:
1. Is event adapter running? `ps aux | grep "observe adapter"`
2. Is NATS running? `curl http://127.0.0.1:8222/healthz`
3. Check adapter logs for errors
4. Check GUI startup for warnings

**Fix**:
```bash
# Ensure NATS is in the stack
go run ./cmd/core stack status

# Start event adapter
go run ./cmd/core stack observe adapter

# Restart GUI
go run ./cmd/core web --live
```

---

### Connection Errors

**Symptom**: "Warning: failed to start event stream: connect to nats..."

**Cause**: NATS not running or wrong URL

**Fix**:
```bash
# Check NATS is running
go run ./cmd/core stack processes | grep nats

# If not running, start stack
go run ./cmd/core stack up
```

---

### Duplicate Events

**Symptom**: Same event appears multiple times

**Cause**: Multiple event adapters running

**Fix**:
```bash
# Kill all adapters
pkill -f "observe adapter"

# Start only one
go run ./cmd/core stack observe adapter
```

---

### No Icons in TUI

**Symptom**: TUI shows `?` instead of emojis

**Cause**: Terminal doesn't support Unicode emojis

**Fix**: Use a modern terminal (iTerm2, Alacritty, Windows Terminal)

---

## Performance Checks

### Memory Usage

**Before**:
- GUI/TUI: ~50MB
- Event log: 10 events × ~100 bytes = 1KB

**After**:
- GUI/TUI: ~52MB (minimal increase)
- Event log: 100 events × ~200 bytes = 20KB
- NATS consumer: ~5MB

**Acceptable**: < 10MB increase total

---

### CPU Usage

**Expected**:
- Event adapter: ~1-2% (polling + publishing)
- NATS: ~1% (message routing)
- GUI/TUI: Same as before (event handling is async)

**Unacceptable**: > 10% CPU sustained

---

## Success Criteria

✅ **Integration Complete** when:
1. GUI/TUI event log shows observability events with icons
2. User actions followed by completion events
3. All 7 event types display correctly
4. 100 event retention works
5. Graceful degradation when adapter not running
6. No performance regression
7. Both TUI and Web GUI work

---

## Known Limitations

1. **No reconnection**: If adapter starts after GUI, must restart GUI
2. **No filtering**: Event log shows all events (can't filter by process yet)
3. **No historical query**: Only shows events since GUI started
4. **Icon support**: Requires Unicode-capable terminal for TUI

---

## Future Enhancements

1. **Auto-reconnect**: Retry NATS connection if it fails
2. **Event filtering**: Filter by process name, event type, severity
3. **Historical replay**: Query JetStream for past events on startup
4. **Timeline view**: Visual timeline showing state transitions
5. **Alert indicators**: Flash/highlight on crashes

---

## Files Modified

- `pkg/runtime/ui/live/store.go` - Added `StartEventStream()`, event retention 100
- `pkg/runtime/cli/execute.go` - Start event stream in web/TUI commands
- Event icons: ▶️ ⏹️ ❌ 🔄 ✅ ⚠️ 📊 ℹ️

