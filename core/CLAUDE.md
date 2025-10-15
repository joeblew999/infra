# CLAUDE.md

## Agent Guides

Use the AGENTS.md at `/Users/apple/workspace/go/src/github.com/joeblew999/infra/agents/AGENTS.md`, following everything it says and its links.

The agent guides are located at `/Users/apple/workspace/go/src/github.com/joeblew999/infra/agents/hugo/content/agents/`:
- `golang/index.md` - Go service development patterns
- `process-compose/index.md` - Process orchestration patterns
- `datastar/index.md` - Datastar backend
- `datastarui-tool/index.md` - DatastarUI tooling
- `litestream/index.md` - Litestream backups
- `nats-jetstream/index.md` - NATS JetStream
- `fly/index.md` - Fly deployment
- `ko/index.md` - ko build

These guides provide context for working with this codebase.

## ⚠️ CRITICAL: .data Directory Protection

**NEVER delete or modify without extreme caution!**

### What's in .data

`.data/` contains **irreplaceable API tokens** for deployment:
- `core/fly/settings.json` - Fly.io API token and organization settings
- `core/cloudflare/settings.json` - Cloudflare API tokens and zone configs
- `core/cloudflare/tokens/` - Active and backup tokens
- `core/secrets/` - Encrypted secrets storage

**These tokens enable autonomous deployment work without user intervention.**

### Mandatory Protection Protocol

**Before ANY code changes that touch .data:**

1. **Check the operation**:
   ```bash
   # Does this code modify .data/core/fly or .data/core/cloudflare?
   grep -r "\.data/core" <files>
   ```

2. **Create timestamped backup**:
   ```bash
   cp -r .data/core .data/.BACKUP_TOKENS_$(date +%Y%m%d_%H%M%S)/
   ```

3. **Verify backup exists**:
   ```bash
   ls -lah .data/.BACKUP_TOKENS/
   ```

4. **After changes, test tokens**:
   ```bash
   # Verify Fly token works
   go run . tooling fly auth whoami

   # Verify Cloudflare token works
   go run . tooling cloudflare zones list
   ```

### Recovery

If tokens are accidentally deleted:
```bash
cp -r .data/.BACKUP_TOKENS/core/* .data/core/
```

### Why This Matters

- **Autonomous work**: Without these tokens, deployment work requires user intervention
- **Irreplaceable**: User must manually re-enter tokens (breaks automation)
- **Session continuity**: Lost tokens mean stopping all deployment work

### Existing Backup

A baseline backup exists at `.data/.BACKUP_TOKENS/` created 2025-10-15.

## Task Tracking - CRITICAL

**TODO.md**: Living task list - MUST be kept up to date during work sessions

**Why This Matters**:
- Claude Code sessions can lock out or timeout unexpectedly
- Context gets lost between sessions
- Future agents need to know what was being worked on and why
- Prevents duplicate work and losing progress

**Requirements**:
1. **Update TODO.md immediately when**:
   - Starting a new task (mark as in progress)
   - Completing a task (mark as done, move to ✅ Completed section)
   - Discovering new issues (add to appropriate priority section)
   - Changing priorities or blocking issues

2. **TODO.md Structure**:
   - 🔥 Critical - In Progress (1-2 items max, currently active)
   - 🎯 High Priority (next up, blocked items)
   - 🔧 Refactoring (improvements, tech debt)
   - 🛠️ Tooling (missing commands, dev experience)
   - ✅ Completed (recent wins, for context)

3. **Each Task Should Include**:
   - Clear description of the problem
   - Current status/blockers
   - File locations (e.g., `services/pocketbase/service.go:217`)
   - Next steps or investigation notes

**Example Update Pattern**:
```markdown
- [x] **Debug PocketBase startup** - FIXED
  - Root cause: app.Execute() returns immediately, doesn't block
  - Solution: Use app.Start() + app.Serve() instead
  - Fixed in: services/pocketbase/service.go:213-217
  - Tested: Port 8090 now listening ✅
```

## Documentation Structure

**README.md**: Quick start and operational commands only
- How to run services
- Basic commands (`go run . stack up`, etc.)
- What commands do (not why or how they work internally)
- Quick reference for developers who just want to use the system

**docs/**: In-depth architectural and design documentation
- `docs/ARCHITECTURE.md` - System design, generation flow, key decisions
- `docs/DEVELOPMENT.md` - Adding services, debugging, extending
- `docs/TROUBLESHOOTING.md` - Common issues and solutions

**Rule**: README is for operators. docs/ is for architects and maintainers.

## Reference Repositories

Located at `../.src/` (parent infra directory, git-ignored):

### `.src/devbox/` - Jetify DevBox
**Repository**: https://github.com/jetify-com/devbox
**Version Used**: Uses process-compose v1.64.1 (see their go.mod)
**Why Important**: Production-proven patterns for wrapping process-compose

**Key Files to Reference**:
- `internal/services/client.go` - Clean HTTP API client for process-compose
- `internal/services/manager.go` - Global instance registry and lifecycle management
- `internal/services/config.go` - YAML parsing and service discovery
- `plugins/*/process-compose.yaml` - Real-world service configurations

**Patterns We Use**:
1. **Direct upstream types**: They import `github.com/f1bonacc1/process-compose/src/types`
2. **Simple HTTP client**: Clean API wrappers without complex retry logic
3. **Process manager checks**: `ProcessManagerIsRunning()` before operations
4. **is_daemon usage**: PostgreSQL uses `true`, MySQL uses `false`

**When to Reference**:
- Updating HTTP client code
- Adding new process management features
- Debugging process-compose integration issues
- Understanding version compatibility

### `.src/process-compose/` - Process Compose Source
**Repository**: https://github.com/F1bonacc1/process-compose
**Current Version**: v1.64.1 (matches devbox for stability)
**Why Important**: Source of truth for behavior, types, and documentation

**Key Directories**:
- `src/types/` - Official type definitions (use these!)
- `src/app/` - Process lifecycle, daemon handling, signal management
- `www/docs/` - Official documentation (more detailed than GitHub README)
- `fixtures-code/` - Test configurations showing edge cases

**Critical Docs**:
- `www/docs/launcher.md` - Process lifecycle, is_daemon flag (line 186+)
- `www/docs/health.md` - Health probes, readiness vs liveness (line 74+)
- `src/types/process.go` - ProcessConfig struct definition

**Key Findings**:
1. **is_daemon flag is required** for proper lifecycle management (line 194-196 in launcher.md)
2. **HTTP probes have bugs** in v1.75.2, stable in v1.64.1
3. **Exec probes are most reliable** - use shell commands with exit codes
4. **Version matters**: 74 commits difference between v1.64.1 and v1.75.2

**When to Reference**:
- Debugging process state transitions
- Understanding why processes mark as "Completed"
- Checking signal handling behavior
- Verifying YAML configuration syntax
- Investigating version-specific bugs

## Process-Compose Integration - Critical Lessons

### The `is_daemon` Flag
**Location**: Service manifests at `services/*/service.json` → `compose.is_daemon`
**Values**:
- `false` - Foreground process that blocks (our NATS, PocketBase, Caddy)
- `true` - Backgrounded daemon (like `docker run -d`, needs shutdown command)
- `omitted` - Treated as task that completes (WRONG for services!)

**Without this flag**: Process-compose thinks services are tasks, marks them "Completed", may send SIGTERM on health check events.

**Rule**: ALL long-running services MUST explicitly set `is_daemon: false` or `true`.

### Health Probes
**Exec probes are most reliable**:
```json
"readiness_probe": {
  "exec": {
    "command": "nc -z 127.0.0.1 4222"  // String, not array!
  }
}
```

**HTTP probes have URL parsing bugs** in some versions:
- v1.75.2: Connects to port 80 instead of specified port
- v1.64.1: More stable but still prefer exec probes

**Best Practice**: Use exec probe with curl for HTTP endpoints:
```json
"exec": {
  "command": "curl -f http://127.0.0.1:8090/api/health"
}
```

### Version Pinning
**We use v1.64.1** (not latest v1.75.2) because:
- Matches devbox's proven version
- Avoids 74 commits of potential regressions
- HTTP probe bugs in newer versions

**Update go.mod**:
```go
github.com/f1bonacc1/process-compose v1.64.1
```

## Abstraction Trade-offs

**Our Approach**: Generate process-compose.yaml from service.json manifests
**Devbox Approach**: Static process-compose.yaml files with variable substitution

**Why We Abstract**:
- Control via real-time TUI and Web GUI
- Service discovery and dynamic orchestration
- Unified configuration format across services
- Type-safe Go structs for validation

**Cost of Abstraction**:
- Config regeneration required for changes
- Complexity in composecfg package
- Version compatibility risks
- Debugging requires checking generated YAML

**When Abstraction Helps**:
- Programmatic service management
- Dynamic port allocation
- Environment-specific configurations
- Integration with our controller system

**When It Hurts**:
- Quick YAML debugging (must check generated file)
- Version updates (need to track upstream types)
- Non-idempotent builds (stale configs)
- Learning curve for contributors

**Mitigation Strategies**:
1. Use upstream types directly (reduce drift)
2. Cache generated configs (hash-based)
3. Show "using cached config" messages
4. Add `stack clean` command
5. Keep examples in `.src/` for reference

## Service Logging Standards

### Legacy "READY" Messages - DO NOT USE

**Historical Context**: Before process-compose integration, services used stdout messages like:
```go
fmt.Printf("READY: nats tcp://127.0.0.1:%d\n", port)
```

**Problems**:
1. **Misleading**: Service printed "READY" but wasn't actually functional
2. **Wrong stream**: Used stdout instead of stderr
3. **No context**: Doesn't indicate what's actually ready
4. **Conflicts**: Process-compose has its own health check system

**These were removed** in October 2024 session after discovering they caused debugging confusion.

### Proper Logging Pattern

**Always use stderr with structured prefixes**:

```go
// Good - Informative startup logging
fmt.Fprintf(os.Stderr, "[service-name] Server listening on tcp://127.0.0.1:%d\n", port)
fmt.Fprintf(os.Stderr, "[service-name] HTTP monitoring on http://127.0.0.1:%d\n", httpPort)
fmt.Fprintf(os.Stderr, "[service-name] Waiting for shutdown signal...\n")

// Good - Error logging
case err := <-errCh:
    if err != nil {
        fmt.Fprintf(os.Stderr, "[service-name] Error: %v\n", err)
    }
    return err
```

**Pattern Requirements**:
- ✅ Use `fmt.Fprintf(os.Stderr, ...)` not `fmt.Printf(...)`
- ✅ Include `[service-name]` prefix for log identification
- ✅ Log specific ports and configuration details
- ✅ Log lifecycle events (startup, waiting, shutdown)
- ✅ Log errors with context
- ❌ Never use "READY:" prefix (legacy pattern)
- ❌ Don't use stdout for service logs

### Health Checks vs Logging

**Separation of Concerns**:
- **Process-compose health probes**: Determine service readiness
- **Service logging**: Human-readable diagnostic information

**Don't conflate the two**:
```json
// Health check (in service.json)
"readiness_probe": {
  "exec": {
    "command": "nc -z 127.0.0.1 4222"
  }
}
```

```go
// Separate diagnostic logging (in service.go)
fmt.Fprintf(os.Stderr, "[nats] Server listening on tcp://127.0.0.1:4222\n")
```

**Why Separate**: Health checks are for automation, logs are for humans. They serve different purposes and should not be tightly coupled.

### Current Service Implementations

**Examples to Follow**:
- [services/nats/service.go:291-296](services/nats/service.go#L291-L296) - NATS startup logging
- [services/caddy/service.go:107-109](services/caddy/service.go#L107-L109) - Caddy startup logging
- [services/caddy/service.go:113-115](services/caddy/service.go#L113-L115) - Error handling

**When Adding New Services**:
1. Copy logging pattern from existing services
2. Use service-specific prefix `[service-name]`
3. Log meaningful configuration (ports, targets, etc.)
4. Don't rely on logs for health checks - use probes
