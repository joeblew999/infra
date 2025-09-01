# Infrastructure Management System Documentation

**Choose your path based on your role and interests:**

## 👔 For Business Users & Decision Makers

**Are you evaluating this system for your business?** 

➡️ **[Business Documentation](business/)** - ROI, features, and business impact

**Focus Areas:**
- **Corporate Branding** - Web, PDF, Email branding automation
- **Workflow Automation** - Visual business process designer  
- **Enterprise Security** - Compliance and audit capabilities
- **Implementation Planning** - Rollout strategy and training

---

## 👨‍💻 For Developers & Technical Teams

**Are you implementing, customizing, or contributing to the system?**

➡️ **[Technical Documentation](technical/)** - Setup, architecture, and development

**Focus Areas:**
- **System Architecture** - Service components and integration
- **Development Environment** - Local setup and workflows
- **Deployment & Operations** - Production deployment and scaling
- **API Integration** - Technical integration details

---

## 🚀 Quick Start (All Users)

**Prerequisites:**
- Go 1.22+ installed
- Git installed  
- 8GB+ RAM recommended

**Get Started:**
```bash
git clone https://github.com/joeblew999/infra.git
cd infra
go run .                    # Starts all services
```

**Access Points:**
- **Main Dashboard**: http://localhost:1337
- **PocketBase Admin**: http://localhost:8090  
- **Bento Stream UI**: http://localhost:4195
- **Deck API**: http://localhost:8888/api/v1/deck/

**Stop Services:**
```bash
go run . shutdown           # Graceful shutdown
```

## 📚 Documentation Structure

### Essential Reading
- **[Beta Testing Guide](BETA_TESTING.md)** - Start here for beta testers
- **[CLI Reference](CLI.md)** - Complete command reference
- **[Troubleshooting](TROUBLESHOOTING.md)** - Common issues and solutions

### Deployment & Operations  
- **[Fly.io Deployment](deployment.md)** - Production deployment guide
- **[Scaling Guide](SCALING.md)** - Horizontal and vertical scaling
- **[API Standards](api-standards.md)** - API design principles

### Advanced Topics
- **[Ko Build System](ko-usage.md)** - Container building with Ko
- **[Roadmap](roadmap/ROADMAP.md)** - Development roadmap and features

### Development Tools
- **[Deck Visualization](deck/)** - Presentation generation system
- **[Games Examples](games/)** - Example implementations

## 🏗️ Architecture Overview

The system runs multiple supervised services:

```
┌─────────────────────────────────────────────────────────────┐
│                    Goreman Supervision                      │
├─────────────────────────────────────────────────────────────┤
│ NATS Server (4222)     │ Message streaming & events       │
│ PocketBase (8090)      │ Database with admin UI           │ 
│ Caddy (80/443)         │ Reverse proxy & HTTPS            │
│ Bento (4195)           │ Stream processing                │
│ Deck API (8888)        │ Go-zero visualization API        │
│ Deck Watcher           │ File processing service          │
│ Web Server (1337)      │ Main dashboard & interface       │
└─────────────────────────────────────────────────────────────┘
```

## 🎯 Key Features

- **Goreman Process Supervision** - Automatic process management and restart
- **Go-zero Microservices** - Modern microservice architecture
- **NATS JetStream** - Real-time messaging and event streaming  
- **PocketBase** - Embedded database with admin interface
- **Stream Processing** - Bento for data pipelines
- **Reverse Proxy** - Caddy for routing and HTTPS
- **Idempotent Workflows** - Safe to run multiple times
- **Auto-scaling** - Horizontal and vertical scaling on Fly.io

## 🧪 Beta Testing Focus Areas

1. **Service Startup** - Test `go run .` on different platforms
2. **Web Interface** - Navigate http://localhost:1337 and report UX issues  
3. **CLI Commands** - Try various `go run . [command]` operations
4. **Process Management** - Test shutdown/restart scenarios
5. **Documentation** - Report unclear or missing documentation

## 🐛 Reporting Issues

When reporting issues, please include:

1. **Platform**: OS and version (macOS/Linux/Windows)
2. **Go Version**: `go version` output
3. **Command**: Exact command that failed
4. **Logs**: Full error output from terminal
5. **Expected vs Actual**: What you expected vs what happened

**Report Issues**: [GitHub Issues](https://github.com/joeblew999/infra/issues)

## 💡 Getting Help

- **Troubleshooting**: Check [TROUBLESHOOTING.md](TROUBLESHOOTING.md)
- **CLI Reference**: See [CLI.md](CLI.md)  
- **Discord/Slack**: [Community channels if available]
- **GitHub Discussions**: [Project discussions]

## 🚦 System Status

Beta testers can check system health:

```bash
go run . status             # Overall deployment status
```

**Health Endpoints:**
- http://localhost:1337/status - Web server health
- http://localhost:1337/metrics - System metrics
- http://localhost:8090/_health - PocketBase health

---

**Note**: This is beta software. Expect rough edges and please report issues to help us improve the system.