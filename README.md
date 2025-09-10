# Infrastructure Management System

[![GitHub](https://img.shields.io/badge/github-joeblew999/infra-blue)](https://github.com/joeblew999/infra)

A comprehensive infrastructure management system with goreman supervision, supporting everything from local development to production deployment.

## Demo

Running on Fly.io 

https://fly.io/apps/infra-mgmt

https://infra-mgmt.fly.dev


## 🚀 Quick Start

```bash
git clone https://github.com/joeblew999/infra.git
cd infra
go run .                    # Start all services with goreman supervision
```

### Local Docker Build image and run locally

```sh
go run . container 
```

builds with ko into .oci folder and runs locally.

### Local Docker and deploy to fly



Access the web interface at **http://localhost:1337**

### Infrastructure Commands
```bash
go run .                    # Start all services (NATS, Caddy, Bento, Deck API, Web Server)
go run . shutdown           # Stop all services
go run . status             # Check deployment status and health
go run . deploy             # Deploy to Fly.io with idempotent workflow
```

### Environment Modes
```bash
go run . --env development   # Development mode with debug features
go run . --env production    # Production mode (default, optimized)
```

## 🌐 Services & Endpoints

The infrastructure runs multiple supervised services:

| Service | Port | Description |
|---------|------|-------------|
| **Web Server** | 1337 | Main dashboard and API |
| **NATS Server** | 4222 | Messaging and event streaming |
| **PocketBase** | 8090 | Database and admin interface |
| **Bento** | 4195 | Stream processing |
| **Deck API** | 8888 | Visualization API (go-zero) |
| **Caddy** | 80/443 | Reverse proxy and HTTPS |

### Web Endpoints
| Endpoint | Description |
|----------|-------------|
| **http://localhost:1337/** | Main dashboard |
| **http://localhost:1337/docs/** | Documentation |
| **http://localhost:1337/logs** | System logs |
| **http://localhost:1337/metrics** | Metrics |
| **http://localhost:1337/status** | Health status |
| **http://localhost:8090/** | PocketBase admin |
| **http://localhost:4195/** | Bento UI |
| **http://localhost:8888/api/v1/deck/** | Deck API |

## 🔧 CLI Commands

### Infrastructure Management
```bash
go run . -h                  # Show organized help (Infrastructure/Development/Advanced)
go run . service             # Same as root command (start all services)
go run . shutdown            # Graceful shutdown of all services
go run . status              # Check deployment status and health
go run . deploy              # Deploy to Fly.io
```

### Scaling
```bash
go run . cli fly scale                    # Show current scaling
go run . cli fly scale --count 2          # Scale to 2 machines
go run . cli fly scale --memory 1024      # Scale memory to 1GB
go run . cli fly scale --cpu 2            # Scale to 2 CPU cores
go run . cli fly scale --vm shared-cpu-2x # Scale VM type
```

### Development Tools
```bash
go run . config              # Print current configuration  
go run . dep list            # Manage binary dependencies
go run . init                # Initialize new project
```

### CLI Tools
```bash
go run . cli deck            # Deck visualization tools
go run . cli gozero          # Go-zero microservices operations
go run . cli fly             # Fly.io operations and scaling
```

### Advanced Tools
```bash
go run . api-check           # Check API compatibility between commits
go run . cli                 # Access CLI tool wrappers
go run . completion          # Generate shell autocompletion
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

**Goreman Supervision**: All services run under goreman supervision for process management and graceful shutdown.

**Key Features:**
- ✅ **Goreman Process Supervision** - Automatic process management and restart
- ✅ **Go-zero Microservices** - Deck API built with go-zero framework  
- ✅ **NATS JetStream** - Real-time messaging and event streaming
- ✅ **PocketBase** - Embedded database with admin interface
- ✅ **Stream Processing** - Bento for data pipelines
- ✅ **Reverse Proxy** - Caddy for routing and HTTPS
- ✅ **Idempotent Workflows** - Safe to run multiple times
- ✅ **Graceful Shutdown** - SIGTERM with SIGKILL fallback
- ✅ **Auto-scaling** - Horizontal and vertical scaling on Fly.io

## 🔄 Workflows

### Local Development
```bash
go run .                    # Start all supervised services
go run . --env development  # Development mode with debug features
go run . shutdown           # Stop all services gracefully
```

All services start automatically with goreman supervision:
1. **NATS Server** (4222) - Message streaming
2. **PocketBase** (8090) - Database 
3. **Caddy** (80/443) - Reverse proxy
4. **Bento** (4195) - Stream processing
5. **Deck API** (8888) - Go-zero visualization API
6. **Deck Watcher** - File processing service
7. **Web Server** (1337) - Main interface

### Production Deployment
```bash
go run . deploy             # Idempotent Fly.io deployment
go run . status             # Check deployment health
go run . cli fly scale --count 2 --memory 1024  # Scale resources
```

### Multi-Environment
- **Local**: Goreman supervision with all services
- **CI/CD**: Idempotent deployment workflows  
- **Production**: Same architecture on Fly.io with scaling

## 🤖 AI Integration

**Built for Claude CLI and Gemini CLI**:
- Lightweight CLI tools instead of heavy IDE extensions
- Direct binary execution for AI workflows
- MCP server support available
- No VS Code slowdown

## 📦 Dependencies

Manage via `./pkg/dep/` - extend by juast matching hte dep.json from any pkg.

## 🌍 Deployment

### Fly.io Platform
The infrastructure is designed for **Fly.io** deployment with:

- **Multi-region support** - Deploy to 34+ regions globally
- **Auto-scaling** - Horizontal (machines) and vertical (CPU/memory) scaling
- **Volume persistence** - Persistent storage with automatic backups
- **Health checks** - Automatic health monitoring and restart
- **Edge networking** - Global anycast routing
- **Rolling deployments** - Zero-downtime updates

### Scaling Options
```bash
# Horizontal scaling (add machines)  
go run . cli fly scale --count 3

# Vertical scaling (resources per machine)
go run . cli fly scale --memory 2048 --cpu 2

# VM type scaling (machine performance)
go run . cli fly scale --vm performance-2x
``` 

## 📊 Monitoring & Observability

### Built-in Monitoring
- **Process Status** - Goreman supervision with health tracking
- **Service Health** - Individual service health checks
- **Real-time Logs** - Multi-destination logging system
- **Metrics Dashboard** - Available at `/metrics`
- **Status Endpoint** - Health checks at `/status`

### NATS Event Streaming
- **JetStream** - Persistent event streaming
- **Real-time messaging** - Inter-service communication  
- **Event publishing** - System events and metrics
- **Stream processing** - Bento integration for data pipelines

### External Monitoring
```bash
go run . status                    # Check deployment health
go run . cli fly status            # Fly.io machine status
go run . cli fly logs              # View application logs
```


