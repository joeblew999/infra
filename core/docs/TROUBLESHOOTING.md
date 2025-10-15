# Troubleshooting Guide

## Stack Won't Start

### Symptom: "address already in use"
```
Error: listen tcp :28081: bind: address already in use
```

**Solution**: Kill existing processes
```bash
# Find processes on ports
lsof -i :4222   # NATS
lsof -i :8090   # PocketBase
lsof -i :2015   # Caddy
lsof -i :28081  # Process-compose

# Kill specific process
kill -9 <PID>

# Or stop the stack cleanly
go run ./cmd/core stack down
```

### Symptom: Services restart continuously or marked as "Completed"

**Most Common Cause**: Missing `is_daemon` flag in service.json

**Solution**: Add to `services/*/service.json`:
```json
{
  "compose": {
    "is_daemon": false,  // REQUIRED for long-running services!
    "readiness_probe": { ... }
  }
}
```

**Why This Matters**: Without `is_daemon`, process-compose treats services as tasks that complete, causing:
- Service marked as "Completed" immediately
- SIGTERM sent when health checks run
- Continuous restart loops

**Check health check configuration**:
```bash
# View generated config
cat .core-stack/process-compose.yaml | grep -A 10 "readiness_probe"
```

**Increase delays and thresholds** in `services/*/service.json`:
```json
{
  "readiness_probe": {
    "initial_delay_seconds": 10,  // Increase if service is slow to start
    "failure_threshold": 10        // Increase tolerance
  }
}
```

### Symptom: Service stuck in "Pending"

**Cause**: Dependency not healthy

**Solution**:
```bash
# Check which service is blocking
go run ./cmd/core stack status

# Check dependency logs
go run ./cmd/core stack process logs nats

# Test health endpoint manually
curl http://127.0.0.1:8222/healthz
```

### Symptom: "TUI startup error: open /dev/tty"

**Cause**: Process-compose trying to start TUI in non-interactive environment

**Solution**: Already handled by adding `--tui=false` in `ExecuteCompose()`. If you see this, check that `pkg/runtime/process/processcompose.go` includes:
```go
if command == "up" {
    if !hasFlag(tail, "--tui", "-t") {
        tail = append([]string{"--tui=false"}, tail...)
    }
}
```

## Service-Specific Issues

### NATS

**Port 8222 not listening**:
- Check that `HTTPPort` is set in `services/nats/service.go`:
```go
natsOpts.HTTPPort = spec.Ports.HTTP.Port  // Not 0!
```

**Authentication errors**:
- NSC credentials are auto-generated in `.data/nats/nsc`
- Delete `.data/nats` to regenerate

### PocketBase

**"must be a valid email address"**:
- Check environment variables are set:
```bash
echo $CORE_POCKETBASE_ADMIN_EMAIL
echo $CORE_POCKETBASE_ADMIN_PASSWORD
```
- Ensure `.env` is loaded (auto-loaded by `cmd/core`)

**Service doesn't start**:
- Changed `app.Execute()` to `app.Start()` in `services/pocketbase/service.go`
- `Execute()` doesn't register default commands

### Caddy

**Service won't start**:
- Check Caddyfile exists
- Verify port 2015 is not in use

## Environment Variables

### Not being loaded

**.env file**:
```bash
# Check it exists
cat .env

# Manually export for testing
export $(cat .env | xargs)
```

**Placeholders not resolved**:
- ${env.*} placeholders are resolved in `ResolveEnv()` functions
- Check `services/*/service.go` includes `replaceEnvPlaceholders()`

## Misleading Symptoms

### Service logs say "started" but port not responding

**Symptom**: Service prints startup message but health checks fail or port is not listening

**Example**:
```
[pocketbase] Server started on http://127.0.0.1:8090
# But curl http://127.0.0.1:8090 fails
```

**Common Causes**:

1. **Race condition**: TCP check passed but service not fully initialized
   - Solution: Increase `initial_delay_seconds` in health probe
   - Use health endpoint (e.g., `/api/health`) instead of TCP check

2. **Service crashed after logging**: Error occurred after log message
   - Solution: Check for error logs after startup message
   - Look for panic traces or error messages in process logs

3. **Wrong port logged**: Service bound to different port
   - Solution: Verify port configuration in service.json
   - Test with `lsof -i :PORT` to see what's actually listening

4. **Process exited immediately**: Wrapper spawned process but it died
   - Solution: Run service standalone: `go run ./cmd/core <service> run`
   - Check if embedded mode is compatible with the service

**Debugging Pattern**:
```bash
# 1. Check if port is actually listening
lsof -i :8090

# 2. Check process status
go run ./cmd/core stack status

# 3. View full logs
go run ./cmd/core stack process logs pocketbase

# 4. Test standalone
go run ./cmd/core pocketbase run
```

### Legacy "READY:" messages

**If you see**:
```
READY: nats tcp://127.0.0.1:4222
READY: caddy http://127.0.0.1:2015
```

**Problem**: You're running old code. These messages were removed because they were misleading (printed before service was actually ready).

**Solution**:
- Pull latest code
- Rebuild binaries: `rm -rf .dep/ && go run ./cmd/core stack up`
- New format: `[service-name] Server listening on...`

## Health Probe Issues

### HTTP probes not working / connecting to wrong port

**Symptom**: Health check tries to connect to port 80 or wrong address

**Cause**: Bug in process-compose HTTP probe URL parsing (especially in v1.75.2+)

**Solution**: Use exec probe with curl instead:
```json
// Bad - HTTP probe can have URL parsing bugs
"readiness_probe": {
  "http_get": {
    "host": "127.0.0.1",
    "port": 8090,
    "path": "/api/health"
  }
}

// Good - Exec probe is most reliable
"readiness_probe": {
  "exec": {
    "command": "curl -f http://127.0.0.1:8090/api/health"
  }
}
```

### Exec probe command format errors

**Symptom**: `json: cannot unmarshal array into Go struct field`

**Cause**: `exec.command` must be a string, not an array

**Solution**:
```json
// Wrong - Array format
"exec": {
  "command": ["nc", "-z", "127.0.0.1", "4222"]
}

// Correct - String format
"exec": {
  "command": "nc -z 127.0.0.1 4222"
}
```

### Best Practices for Health Probes

**TCP Check** (simplest, for any port):
```json
"exec": {
  "command": "nc -z 127.0.0.1 4222"
}
```

**HTTP Check** (for health endpoints):
```json
"exec": {
  "command": "curl -f http://127.0.0.1:8090/api/health"
}
```

**Timing Recommendations**:
```json
{
  "initial_delay_seconds": 5,   // Wait before first check
  "period_seconds": 5,          // Time between checks
  "timeout_seconds": 3,         // Max time for check to complete
  "success_threshold": 1,       // Checks to mark healthy
  "failure_threshold": 5        // Checks to mark unhealthy
}
```

## Process-Compose Issues

### Can't connect to process-compose

```bash
# Check if running
lsof -i :28081

# Check status
go run ./cmd/core stack status

# Restart
go run ./cmd/core stack down
go run ./cmd/core stack up
```

### Process-compose version mismatch

**Symptom**: Unexpected behavior, unknown YAML keys, probe issues

**Check version**:
```bash
# Check go.mod
grep process-compose go.mod
```

**Expected**: `github.com/f1bonacc1/process-compose v1.64.1`

**If different**: We pin to v1.64.1 for stability (matches devbox). Newer versions (v1.75.2+) have 74 commits of changes and known issues with HTTP probes.

### Logs not showing

```bash
# Use process logs command
go run ./cmd/core stack process logs <service-name>

# Or check .core-stack/ directory
ls -la .core-stack/
```

## Debugging

### Enable verbose logging

```bash
# Set log level
export LOG_LEVEL=debug

# Run with output
go run ./cmd/core stack up 2>&1 | tee stack.log
```

### Check generated config

```bash
# View full config
cat .core-stack/process-compose.yaml

# Check specific service
cat .core-stack/process-compose.yaml | grep -A 30 "pocketbase:"
```

### Test services individually

```bash
# Run service standalone (bypasses process-compose)
go run ./cmd/core nats run
go run ./cmd/core pocketbase run
go run ./cmd/core caddy run
```

### Inspect binary paths

```bash
# Check binaries were built
ls -la .dep/

# Manually build if needed
go build -o .dep/nats ./cmd/nats
go build -o .dep/pocketbase ./cmd/pocketbase
go build -o .dep/caddy ./cmd/caddy
```

## Clean Reset

```bash
# Stop everything
go run ./cmd/core stack down

# Kill any zombie processes
pkill -9 nats
pkill -9 pocketbase
pkill -9 caddy
pkill -9 processcompose

# Remove generated files
rm -rf .core-stack/

# Remove data (caution: deletes databases!)
rm -rf .data/

# Remove binaries (will rebuild on next run)
rm -rf .dep/

# Start fresh
go run ./cmd/core stack up
```

## Getting Help

1. Check this guide
2. Review [DEVELOPMENT.md](DEVELOPMENT.md)
3. Inspect logs: `go run ./cmd/core stack process logs <service>`
4. Test health endpoints manually
5. Run services individually to isolate issues
