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

### Symptom: Services restart continuously

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
