# Beta Testing Guide

Thank you for participating in the beta testing of the Infrastructure Management System! This guide will help you test the system effectively and report useful feedback.

## 🎯 Testing Objectives

We're looking for feedback on:
1. **Ease of setup** - How smooth is the initial experience?
2. **Service reliability** - Do all services start and run consistently?
3. **Web interface usability** - Is the dashboard intuitive?
4. **Documentation clarity** - Are instructions clear and complete?
5. **Cross-platform compatibility** - Does it work on your OS?

## 🚀 Quick Setup Test

### Step 1: Clone and Start
```bash
git clone https://github.com/joeblew999/infra.git
cd infra
go run .
```

**⏱️ Expected time**: 30-60 seconds for first startup (downloads dependencies)  
**✅ Success indicator**: You see "🌐 Starting web server address=http://localhost:1337"

### Step 2: Verify Services
Open these URLs in your browser:

| Service | URL | Expected Result |
|---------|-----|----------------|
| **Main Dashboard** | http://localhost:1337 | Dashboard with service status |
| **PocketBase Admin** | http://localhost:8090 | Database admin login |
| **Bento Stream UI** | http://localhost:4195 | Stream processing interface |
| **System Status** | http://localhost:1337/status | JSON health status |

### Step 3: Test Shutdown
```bash
go run . shutdown
```
**✅ Success**: All services stop gracefully, no hanging processes

## 🧪 Detailed Test Scenarios

### Scenario 1: Multi-restart Test
Test service resilience:
```bash
go run .                    # Start services
# Wait 30 seconds
^C                          # CTRL+C to interrupt
go run .                    # Start again immediately
go run . shutdown           # Clean shutdown
go run .                    # Start one more time
```

**Expected**: Services should start cleanly each time without port conflicts.

### Scenario 2: CLI Commands Test
Test various CLI operations:
```bash
go run . -h                 # Help should show organized categories
go run . status             # Should show deployment status
go run . config             # Should print current configuration
go run . cli -h             # Should show CLI tools
go run . cli deck -h        # Should show deck commands
```

### Scenario 3: Web Interface Test
Navigate through the web interface:

1. **Dashboard** (http://localhost:1337):
   - Check service status display
   - Navigate to different sections
   - Test any interactive elements

2. **Logs** (http://localhost:1337/logs):
   - Verify log display works
   - Check for any errors in logs

3. **Metrics** (http://localhost:1337/metrics):
   - Confirm metrics are displayed
   - Look for performance indicators

### Scenario 4: Concurrent Usage
Test with multiple terminal windows:
```bash
# Terminal 1: Start services
go run .

# Terminal 2: Check status while running
go run . status

# Terminal 3: Try other commands
go run . config
go run . cli deck list
```

## 📝 What to Report

### Critical Issues (High Priority)
- **Services fail to start** - Specific error messages
- **Port conflicts** - Which ports are already in use  
- **Web interface not loading** - Browser console errors
- **Shutdown hangs** - Processes that won't stop

### Important Issues (Medium Priority)
- **Slow startup times** - How long it takes
- **UI/UX problems** - Confusing interface elements
- **Documentation gaps** - Missing or unclear instructions
- **CLI command failures** - Commands that don't work

### Nice-to-have Issues (Low Priority)
- **Performance observations** - Memory/CPU usage concerns
- **Feature suggestions** - What would make it better
- **UI improvements** - Design or usability suggestions

## 📊 Testing Checklist

Copy this checklist and mark ✅/❌ as you test:

### Setup & Installation
- [ ] Clone repository successfully
- [ ] First `go run .` completes without errors
- [ ] All 7 services start (check terminal output)
- [ ] Main dashboard loads at http://localhost:1337
- [ ] No port conflict errors

### Service Testing  
- [ ] NATS server accessible (check logs for "NATS server started")
- [ ] PocketBase admin loads at http://localhost:8090
- [ ] Bento UI loads at http://localhost:4195  
- [ ] Caddy proxy working (check for Caddy logs)
- [ ] Deck API responds at http://localhost:8888/api/v1/deck/

### CLI Testing
- [ ] `go run . -h` shows organized help
- [ ] `go run . status` works
- [ ] `go run . config` displays configuration
- [ ] `go run . shutdown` stops all services
- [ ] No hanging processes after shutdown

### Reliability Testing
- [ ] Services restart cleanly after shutdown
- [ ] Multiple start/stop cycles work
- [ ] CTRL+C stops services gracefully
- [ ] System recovers from interruption

### Cross-Platform Testing
**Platform**: ________________ (macOS 14, Ubuntu 22.04, Windows 11, etc.)
- [ ] All services start on your platform
- [ ] Web interfaces load correctly
- [ ] No platform-specific errors

## 🐛 Issue Reporting Template

When reporting issues, please use this template:

```markdown
## Issue Summary
Brief description of the problem

## Environment
- **OS**: macOS 14.5 / Ubuntu 22.04 / Windows 11
- **Go Version**: `go version` output
- **Hardware**: CPU/RAM specs if relevant

## Steps to Reproduce
1. Run `git clone https://github.com/joeblew999/infra.git`
2. Run `cd infra`
3. Run `go run .`
4. ...

## Expected Behavior
What should have happened

## Actual Behavior  
What actually happened

## Error Output
```
Paste full error output here
```

## Additional Context
- Browser used (if web issue): Chrome 118, Firefox 119, etc.
- First time setup or repeated run?
- Any other relevant details
```

## 🎉 Success Stories

If everything works perfectly, that's great feedback too! Please let us know:
- Your platform and setup
- How long startup took
- Overall experience rating
- What you liked most

## 🔄 Testing Updates

As we release updates during beta:

1. **Pull latest changes**:
   ```bash
   git pull origin main
   ```

2. **Stop and restart** to test the new version:
   ```bash
   go run . shutdown
   go run .
   ```

3. **Retest** any previously reported issues

## 💬 Communication Channels

- **GitHub Issues**: https://github.com/joeblew999/infra/issues
- **Discussions**: https://github.com/joeblew999/infra/discussions
- **Discord/Slack**: [Channel info if available]

---

**Thank you for helping make this system better!** 🙏

Your feedback is invaluable for identifying issues and improving the user experience before general release.