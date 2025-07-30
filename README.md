# Infrastructure Management System

[![GitHub](https://img.shields.io/badge/github-joeblew999/infra-blue)](https://github.com/joeblew999/infra)

A self-similar infrastructure management system that bridges development and production with identical workflows.

## 🚀 Quick Start

```bash
git clone https://github.com/joeblew999/infra.git
cd infra
go run .
```

Access the web interface at **http://localhost:1337**

## 🌐 Endpoints

| Endpoint | Description |
|----------|-------------|
| **http://localhost:1337/** | Main dashboard with NATS → DataStar demo |
| **http://localhost:1337/docs/** | Documentation |
| **http://localhost:1337/logs** | System logs |
| **http://localhost:1337/metrics** | Metrics |
| **http://localhost:1337/status** | Status |

## 🔧 Development

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

- **Development**: Use binaries and CLI tools directly
- **Runtime**: Use the same binaries and CLI tools, just orchestrated
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
go run .                    # Start web interface
go run . cli --help         # CLI commands
go run . api-check          # Verify API compatibility
```

### CI/CD
- **GitHub Actions**: Uses same binaries via Taskfiles
- **Terraform**: Provisions infrastructure via Taskfiles
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

Manage via `./pkg/dep/` - see [./roadmap/dep.md](./roadmap/dep.md) for details.

## 🌍 Deployment

### Primary
**Hetzner Cloud (Germany)** - European coverage

### Secondary
**OVH Cloud** - Global coverage
- [OVH Terraform Provider](https://github.com/ovh/terraform-provider-ovh)
- Supports: VMs, K8s, DNS, Load Balancers, Storage

## 📊 Monitoring

- **NATS JetStream** for event streaming
- **Multi-destination logging** for observability
- **Real-time web interface** for monitoring
- **Self-reflection** via NATS event publishing


