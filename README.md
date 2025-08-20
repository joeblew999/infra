# Infrastructure Management System

[![GitHub](https://img.shields.io/badge/github-joeblew999/infra-blue)](https://github.com/joeblew999/infra)

A self-similar infrastructure management system that bridges development and production with identical workflows.

## 🚀 Quick Start

```bash
git clone https://github.com/joeblew999/infra.git
cd infra
go run .                    # Start all services
```

Access the web interface at **http://localhost:1337**

### Environment Modes
```bash
go run . --env dev          # Development mode
go run . --env production   # Production mode (default)
```

## 🌐 Endpoints

| Endpoint | Description |
|----------|-------------|
| **http://localhost:1337/** | Main dashboard with NATS → DataStar demo |
| **http://localhost:1337/docs/** | Documentation |
| **http://localhost:1337/logs** | System logs |
| **http://localhost:1337/metrics** | Metrics |
| **http://localhost:1337/status** | Status |

## 🔧 Development

### CLI Tools (Always Available)
All debugging and provisioning tools are always available:
```bash
go run . --help               # Show all commands
go run . dep list             # Manage dependencies
go run . task                # Run Taskfiles
go run . build               # Build application
```

### Web GUI Debugging
Use Claude Code's built-in Playwright tools:
```
mcp__playwright__browser_navigate → http://localhost:1337
mcp__playwright__browser_click → click buttons
mcp__playwright__browser_type → input text
mcp__playwright__browser_evaluate → check DOM state
mcp__playwright__browser_console_messages → debug errors
```

### Architecture
**Self-Similar Design**: The same patterns work at development time and runtime.

- **Unified Interface**: `go run .` - starts all services automatically
- **Always Available**: All CLI tools accessible without mode switching
- **No mental model shift** between dev and prod

**Key Features:**
- ✅ **NATS JetStream** for real-time messaging and logging
- ✅ **Multi-destination logging** (stdout, files, NATS)
- ✅ **Runtime configuration** without restart
- ✅ **Idempotent workflows** across environments
- ✅ **Cross-platform** (laptop, CI, CD, production)

## 🔄 Workflows

### Local Development

```bash
go run .                    # Always starts all services
go run . --env dev          # Development mode with migration tools
go run . --env production   # Production mode (optimized)
```

### CI/CD
- **GitHub Actions**: All IAC based, so DRY, with github actions being close to empty.
- **Terraform**: Provisions infrastructure via workflows as IAC.
- **Versioned**: Use git hashes/tags for reproducible builds

### Multi-Environment
- **Laptop**: Direct binary execution
- **CI**: Same binaries via GitHub Actions
- **CD**: Same binaries via Terraform
- **Production**: Same binaries via orchestration

## 🤖 AI Integration

**Built for Claude CLI and Gemini CLI**:
- Lightweight CLI tools instead of heavy IDE extensions
- Direct binary execution for AI workflows
- MCP server support available
- No VS Code slowdown

## 📦 Dependencies

Manage via `./pkg/dep/` - extend by juast matching hte dep.json from any pkg.

## 🌍 Deployment

### Primary
**Hetzner Cloud (Germany)** - European coverage


### Secondary
**Fl.io Cloud** - Global coverage
- 6 NATS Servers protected internally.
- 22 regions, with autoscaling in each. 

## 📊 Monitoring

- **NATS JetStream** for event streaming
- **Multi-destination logging** for observability
- **Real-time web interface** for monitoring
- **Self-reflection** via NATS event publishing


