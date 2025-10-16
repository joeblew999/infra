# Health Monitoring Guide

## Overview

The core V2 stack includes comprehensive health monitoring through multiple commands that work together to provide real-time visibility into service health, events, and diagnostics.

## Monitoring Commands

### 1. Stack Status (Quick Check)

**Command**: `go run . stack status`

**Purpose**: Instant snapshot of stack health

**Output**:
```
Stack status: running (process-compose)
Ports in use: [4222 8090 2015]
Process Compose:
• default/caddy      status: Running      health: Ready    restarts: 0
• default/nats       status: Running      health: Ready    restarts: 0
• default/pocketbase status: Running      health: Ready    restarts: 0
Services:
• nats         port 4222  status: running     client → 4222
• pocketbase   port 8090  status: running     primary → 8090
• caddy        port 2015  status: running     http → 2015
```

**Use Case**: Quick "is everything running?" check

---

### 2. Stack Doctor (Comprehensive Diagnostics)

**Command**: `go run . stack doctor --verbose`

**Purpose**: Deep health check with actionable suggestions

**What It Checks**:
- ✅ Port availability (4222, 8090, 2015, 8222, 28081)
- ✅ Process-compose connectivity
- ✅ Individual process health and restart counts
- ✅ Health endpoints (HTTP checks for NATS, PocketBase, Caddy)
- ✅ .data directory existence
- ✅ Deployment tokens (Fly.io, Cloudflare)
- ✅ Zombie process detection

**Output**:
```
🔍 Running stack diagnostics...

→ Checking port availability...
  ✓ Port check complete

→ Checking process-compose...
  ✓ Process-compose running (3 processes)
    ✓ default/nats: Running (restarts: 0)
    ✓ default/pocketbase: Running (restarts: 0)
    ✓ default/caddy: Running (restarts: 0)

→ Checking health endpoints...
  ✓ NATS: healthy
  ✓ PocketBase: healthy
  ✓ Caddy: healthy

→ Checking .data directory...
  ✓ .data/core exists

→ Checking deployment tokens...
  ✓ Fly.io: token valid (gedw99@gmail.com, org: personal)
  ✓ Cloudflare: settings found

→ Checking for zombie processes...
  ✓ No zombie processes on stack ports

✅ Stack is healthy!
   0 errors, 0 warnings
```

**Use Case**: Pre-deployment checks, troubleshooting, CI/CD validation

**Flags**:
- `--verbose`: Show detailed output for each check

---

### 3. Stack Observe (Real-Time Events)

**Command**: `go run . stack observe watch`

**Purpose**: Live event stream from all services

**Features**:
- 🔴 Process lifecycle events (started, stopped, crashed)
- 🟢 Health state changes (healthy, unhealthy)
- 📊 Observability events from services
- 🔍 Filterable by process and event type

**Output**:
```
Watching events: process.*
Press Ctrl+C to stop

2025-10-16T18:30:45Z ✅ process.healthy: default/caddy
2025-10-16T18:30:46Z ✅ process.healthy: default/nats
2025-10-16T18:30:47Z ✅ process.healthy: default/pocketbase
2025-10-16T18:31:02Z ℹ️  service.request: default/caddy - GET /api/health → 200 OK
```

**Filtering**:
```bash
# Watch only nats events
go run . stack observe watch --process nats

# Watch only health changes
go run . stack observe watch --type healthy,unhealthy

# JSON output for parsing
go run . stack observe watch --json
```

**Use Case**: Debugging issues in real-time, monitoring during deployment, watching service interactions

**Flags**:
- `--process, -p`: Filter by process name
- `--type, -t`: Filter by event type
- `--json`: Output as JSON for automation
- `--nats-url`: Custom NATS server URL

---

## Health Monitoring Dashboard Workflow

### Local Development Monitoring

**Terminal 1** - Live event stream:
```bash
go run . stack observe watch
```

**Terminal 2** - Run operations:
```bash
# Make changes, restart services, etc.
go run . stack process restart default/caddy
```

**Terminal 3** - Periodic health checks:
```bash
# Every 30 seconds
watch -n 30 "go run . stack status"
```

---

### Production Monitoring Setup

#### 1. Continuous Health Checks

Create a monitoring script (`scripts/monitor.sh`):
```bash
#!/bin/bash

while true; do
  echo "=== Health Check $(date) ==="
  go run . stack doctor

  # Check exit code
  if [ $? -ne 0 ]; then
    echo "❌ Health check failed!"
    # Send alert (email, Slack, PagerDuty, etc.)
    ./scripts/alert.sh "Stack health check failed"
  fi

  sleep 300  # Check every 5 minutes
done
```

#### 2. Event Logging

Log all observability events:
```bash
# Log to file with timestamps
go run . stack observe watch --json >> /var/log/core-events.log

# Or send to logging service
go run . stack observe watch --json | ./scripts/send-to-loki.sh
```

#### 3. Alerting Rules

Monitor for specific patterns:
```bash
# Alert on any crashed processes
go run . stack observe watch --type crashed | while read event; do
  echo "🚨 ALERT: Process crashed - $event"
  ./scripts/alert.sh "Process crashed: $event"
done

# Alert on persistent unhealthy state
go run . stack observe watch --type unhealthy | while read event; do
  sleep 30  # Wait 30s
  if go run . stack status | grep -q "unhealthy"; then
    echo "🚨 ALERT: Service unhealthy for 30+ seconds"
    ./scripts/alert.sh "Persistent unhealthy state"
  fi
done
```

---

## Health Metrics

### Key Metrics to Monitor

1. **Process Restarts**
   - Source: `go run . stack status`
   - Threshold: Alert if restarts > 5 in 1 hour
   - Action: Check logs, investigate root cause

2. **Health Endpoint Response Time**
   - Source: Custom script with curl
   - Threshold: Alert if response > 2s
   - Action: Check service load, investigate slowness

3. **Port Availability**
   - Source: `go run . stack doctor`
   - Threshold: Alert if any port unavailable when stack should be running
   - Action: Check for zombie processes, port conflicts

4. **Event Rate**
   - Source: `go run . stack observe watch --json | jq .`
   - Threshold: Alert if crash events > 0 or unhealthy events > 3/min
   - Action: Investigate service instability

### Metrics Collection Script

```bash
#!/bin/bash
# collect-metrics.sh

OUTPUT_DIR="/var/log/core-metrics"
mkdir -p "$OUTPUT_DIR"

# Collect status every minute
while true; do
  TIMESTAMP=$(date +%s)
  go run . stack status --json > "$OUTPUT_DIR/status-$TIMESTAMP.json" 2>&1
  go run . stack doctor --json > "$OUTPUT_DIR/doctor-$TIMESTAMP.json" 2>&1

  # Cleanup old metrics (keep 24 hours)
  find "$OUTPUT_DIR" -name "*.json" -mtime +1 -delete

  sleep 60
done
```

---

## Integration with Observability Tools

### Prometheus Metrics

Expose metrics from health checks:
```bash
# Example: Convert stack status to Prometheus format
go run . stack status --json | jq -r '
  .processes[] |
  "process_restarts{name=\"\(.name)\"} \(.restarts)\n" +
  "process_running{name=\"\(.name)\"} \(if .status == \"Running\" then 1 else 0 end)"
' > /var/lib/prometheus/core-metrics.prom
```

### Grafana Dashboard

Create dashboards using:
- Process restart counts (line graph)
- Health status per service (gauge)
- Event rate over time (area graph)
- Port availability (status panel)

### Datadog/New Relic

Send events via API:
```bash
go run . stack observe watch --json | while read event; do
  curl -X POST https://api.datadoghq.com/api/v1/events \
    -H "DD-API-KEY: $DD_API_KEY" \
    -d "$event"
done
```

---

## Troubleshooting with Health Monitors

### Scenario 1: Service Won't Start

```bash
# Step 1: Check if anything is running
go run . stack status

# Step 2: Run full diagnostics
go run . stack doctor --verbose

# Step 3: Watch for events during startup
go run . stack observe watch &
go run . stack up

# Step 4: Check for specific errors
go run . stack observe watch --type crashed,unhealthy
```

### Scenario 2: Intermittent Failures

```bash
# Monitor continuously
go run . stack observe watch >> events.log &

# Let it run for a while, then analyze
grep "unhealthy\|crashed" events.log
```

### Scenario 3: Performance Degradation

```bash
# Check health endpoint response times
time curl http://127.0.0.1:8222/healthz  # NATS
time curl http://127.0.0.1:8090/api/health  # PocketBase
time curl http://127.0.0.1:2015/api/health  # Caddy

# Watch for slow requests
go run . stack observe watch | grep -i "slow\|timeout"
```

---

## Health Monitoring Best Practices

1. **Run `stack doctor` before every deployment**
   - Catch issues before they hit production
   - Validate environment is ready

2. **Keep `stack observe watch` running during development**
   - See real-time feedback on changes
   - Catch issues immediately

3. **Set up automated alerts for production**
   - Don't rely on manual checks
   - Get notified of issues immediately

4. **Log all events for post-mortem analysis**
   - Keep event logs for at least 7 days
   - Analyze patterns after incidents

5. **Monitor trends, not just current state**
   - Increasing restart counts indicate problems
   - Gradual response time degradation needs investigation

---

## Future Enhancements

Planned improvements to health monitoring:

1. **Web Dashboard** - Visual real-time dashboard accessible via browser
2. **Health Score** - Composite score (0-100) indicating overall stack health
3. **Predictive Alerts** - ML-based anomaly detection for early warning
4. **Auto-Remediation** - Automatic restart of unhealthy services
5. **Integration APIs** - REST/GraphQL API for custom monitoring tools
6. **Multi-Stack Monitoring** - Monitor multiple core deployments from one dashboard

---

## Commands Quick Reference

| Command | Purpose | When to Use |
|---------|---------|-------------|
| `stack status` | Quick snapshot | Every time you want to check "is it running?" |
| `stack doctor` | Deep health check | Before deployment, troubleshooting, CI/CD |
| `stack observe watch` | Live events | Debugging, monitoring during operations |
| `stack observe adapter` | Start event publisher | Run in background for observability |
| `stack up` | Start services | Beginning of development session |
| `stack down` | Stop services | End of development session |
| `stack clean` | Remove zombies | After crashes, before fresh start |
| `stack process restart <name>` | Restart service | After config changes, testing recovery |

---

## Next Steps

1. Try each command to understand the output
2. Set up a monitoring script for your environment
3. Integrate with your existing observability stack
4. Configure alerts based on your SLAs
5. Document your team's monitoring runbook

For questions or feature requests, see:
- [TROUBLESHOOTING.md](TROUBLESHOOTING.md) - Common issues
- [ARCHITECTURE.md](ARCHITECTURE.md) - System design
- [TODO.md](../TODO.md) - Planned improvements
