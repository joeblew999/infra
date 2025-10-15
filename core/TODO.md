# TODO

## 🔥 Critical - In Progress

- [ ] **Fix Caddy not listening on port 2015**
  - Issue: Caddy logs "Server started" but port 2015 not actually listening
  - Process status: Running in process-compose, but `lsof -i :2015` shows nothing
  - Same symptom as PocketBase had, but different root cause (no OnServe hooks)
  - Caddy uses `caddy.Run(&config)` which should block and listen
  - Next: Add more debug logging, verify waitForTCP is working correctly
  - Location: `services/caddy/service.go:99-103`

- [ ] **Test full stack orchestration with all three services**
  - Goal: NATS → PocketBase → Caddy all running and healthy
  - Status: NATS ✅ PocketBase ✅ Caddy ⚠️ (process running but port not listening)

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

## ✅ Completed

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
