# TODO

## 🔥 Critical - In Progress

- [ ] **Fix Caddy immediate exit - caddy.Run() doesn't block**
  - **Root Cause Found**: `caddy.Run()` returns nil immediately instead of blocking!
  - Debug output shows:
    ```
    [caddy] Calling caddy.Run()...
    {"level":"info","msg":"server running"...}
    [caddy] caddy.Run() returned: <nil>
    ```
  - Server briefly starts but exits immediately (TCP check passes then fails)
  - Caused 289 restarts in process-compose before discovery
  - Same pattern as PocketBase issue (returns instead of blocking)
  - **Solution needed**: Research Caddy v2 API for proper blocking server lifecycle
  - Possibilities:
    1. Use different Caddy API (caddy.Start() + wait?)
    2. Block on signal channel after Run()
    3. Use Caddy's built-in signal handling differently
  - Location: `services/caddy/service.go:99-124`
  - **Added**: Comprehensive debug logging to track execution flow

- [ ] **Test full stack orchestration with all three services**
  - Goal: NATS → PocketBase → Caddy all running and healthy
  - Status: NATS ✅ PocketBase ✅ Caddy ❌ (exits immediately, 289 crash loops)

## 🎯 High Priority - Stack Orchestration

## 🔧 Refactoring - Process-Compose Integration

- [ ] **Replace composecfg with upstream process-compose types**
  - Why: Reduce abstraction, use official types directly
  - Reference: `.src/devbox/` imports `github.com/f1bonacc1/process-compose/src/types`
  - Benefit: Less maintenance, better compatibility

- [ ] **Simplify HTTP client using devbox patterns**
  - Current: Custom HTTP client in our codebase
  - Reference: `.src/devbox/internal/services/client.go`
  - Benefit: Cleaner, proven implementation

- [ ] **Add ProcessManagerIsRunning check before operations**
  - Why: Prevent operations when process-compose not running
  - Reference: Devbox pattern for safety checks
  - Location: Add to `pkg/runtime/process/processcompose.go`

## 🛠️ Tooling - Missing Commands

- [ ] **Build `stack logs <service>` command**
  - Purpose: Stream logs for specific service
  - Should use: process-compose API or read log files
  - Location: `pkg/runtime/cli/stack.go`

- [ ] **Build `stack clean` command to remove zombies**
  - Purpose: Kill zombie processes, remove generated files
  - Should clean: `.core-stack/`, zombie processes, stale PIDs
  - Location: `pkg/runtime/cli/stack.go`

- [ ] **Build `stack doctor` command for diagnostics**
  - Purpose: Check for common issues automatically
  - Should check: Ports, health endpoints, config validity, version compatibility
  - Location: `pkg/runtime/cli/stack.go`

## 📝 Documentation - Future Work

- [ ] **Add examples to DEVELOPMENT.md**
  - Adding a new service walkthrough
  - Debugging process-compose issues
  - Writing health probes

- [ ] **Create migration guide for upstream types**
  - When we replace composecfg with upstream types
  - Breaking changes and how to adapt

## ✅ Completed (This Session)

- [x] **Backup and protect .data directory** - CRITICAL!
  - Created `.data/.BACKUP_TOKENS/` with Fly.io and Cloudflare API tokens
  - Added protection protocol to CLAUDE.md (⚠️ CRITICAL section)
  - Documented recovery procedures
  - Ensures autonomous deployment work can continue
- [x] **Diagnose Caddy crash loop**
  - Found 289 restarts in process-compose
  - Root cause: `caddy.Run()` returns nil immediately (doesn't block)
  - Added comprehensive debug logging to track execution
  - Identified need for different Caddy API approach
- [x] **Debug PocketBase startup** - FIXED! 🎉
  - Root causes found:
    1. `admin@localhost` not valid email → changed to `admin@example.com` in `.env`
    2. `auth.go:388` missing `se.Next()` call → added `return se.Next()`
    3. Used `os.Args` override instead of `SetArgs()` for proper CLI parsing
  - Result: Port 8090 listening, health endpoint returns `{"message":"API is healthy."}`
  - Files: `services/pocketbase/auth.go:388`, `.env:ADMIN_EMAIL`
- [x] Remove legacy READY messages from all services
- [x] Add structured logging with [service-name] prefixes
- [x] Add is_daemon: false to all service.json manifests
- [x] Fix NATS signal handling (NoSigs: true, removed SIGHUP)
- [x] Pin process-compose to v1.64.1 for stability
- [x] Document logging standards in CLAUDE.md
- [x] Document TODO.md tracking requirements in CLAUDE.md
- [x] Document troubleshooting patterns in docs/TROUBLESHOOTING.md
- [x] Add health probe best practices to docs
- [x] Clone reference repos (.src/devbox, .src/process-compose, .src/pocketbase)
- [x] Commit and push logging refactor changes

---

## Notes

**Current Stack Status**:
- ✅ NATS: Running on port 4222, healthy
- ✅ PocketBase: Fixed and working on port 8090, health endpoint responding
- ⏳ Caddy: Ready to test (was waiting for PocketBase)

**Key Learnings**:
- `is_daemon: false` is REQUIRED for long-running services
- Exec probes are more reliable than HTTP probes
- Legacy "READY" messages were misleading (removed Oct 2024)
- v1.64.1 is stable, v1.75.2+ has probe bugs
