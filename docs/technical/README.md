# Technical Documentation

**Infrastructure Management System: Everything-as-Go-Import**

## 🎯 Core Philosophy: Complete Infrastructure in a Single Import

**The Problem**: Traditional projects require dozens of tools, configs, and deployment scripts  
**Our Solution**: Everything you need is embedded as Go code - import and run

### Revolutionary Approach
Instead of managing separate tools, everything is **embedded Go**:

```go
import "github.com/joeblew999/infra"

func main() {
    infra.Start() // Complete production infrastructure running
}
```

**What you get from a single import**:
- ✅ **Web Server** (embedded)
- ✅ **Database** (PocketBase embedded)
- ✅ **Message Queue** (NATS embedded)
- ✅ **Reverse Proxy** (Caddy embedded)
- ✅ **Stream Processing** (Bento embedded)
- ✅ **Container Building** (Ko embedded)
- ✅ **Deployment** (Fly.io integration embedded)
- ✅ **Process Supervision** (Goreman embedded)
- ✅ **Monitoring & Logging** (embedded)

## 🚀 Zero-Config Development Experience

```bash
# Traditional project setup
npm install && docker-compose up && kubectl apply -f k8s/ && ...

# Our approach  
go run . runtime up          # Everything starts, production-ready
```

**No external dependencies**:
- No Docker required (but supported)
- No Kubernetes required (but supported)
- No separate database installation
- No reverse proxy configuration
- No deployment pipeline setup

**Access Points:**
- **Main Dashboard**: http://localhost:1337
- **PocketBase Admin**: http://localhost:8090  
- **Bento Stream UI**: http://localhost:4195
- **Deck API**: http://localhost:8888/api/v1/deck/

## 📚 Technical Documentation

### Core Concepts
- **[System Specification](SYSTEM_SPEC.md)** - High-level architecture, goals, and operating model
- **[Everything-as-Go-Import](EVERYTHING_AS_GO.md)** - Revolutionary embedded infrastructure approach
- **[Beta Testing Guide](BETA_TESTING.md)** - Testing procedures and requirements  
- **[CLI Reference](CLI.md)** - Complete command-line interface documentation
- **[API Standards](api-standards.md)** - API design principles and standards

### Deployment & Operations  
- **[Deployment Guide](deployment.md)** - Fly.io deployment procedures
- **[Scaling Guide](SCALING.md)** - Horizontal and vertical scaling
- **[Ko Build System](ko-usage.md)** - Container building with Ko

### Development  
- **[Architecture Documentation](EVERYTHING_AS_GO.md)** - Detailed technical architecture

## 🏗️ System Architecture

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

## 🔧 Embedded Development & Deployment

### Prerequisites (Minimal)
- **Go 1.22+** (Only requirement)
- **Git** (For version control)
- **8GB+ RAM** (recommended)

**That's it!** No Docker, Kubernetes, databases, or external services needed.

### Complete Development Workflow (Embedded)
```bash
# Development (embedded servers, hot reload)
go run . runtime up --env development

# Testing (embedded test infrastructure)  
go test ./...

# API compatibility checking (embedded tooling)
go run . dev api-check

# Container building (embedded Ko builder)
go run . workflows build

# Production deployment (embedded Fly.io integration)  
go run . workflows deploy

# Scaling (embedded scaling commands)
go run . tools flyctl scale --count 3
```

### Why Everything is Embedded Go Code

**Traditional Approach** (Complex):
```bash
# Separate tools and configs everywhere
docker-compose.yml          # Container orchestration
Dockerfile                  # Container definition  
package.json               # Frontend dependencies
requirements.txt           # Python dependencies
nginx.conf                 # Reverse proxy config
k8s/                      # Kubernetes manifests
.github/workflows/        # CI/CD pipelines
terraform/                # Infrastructure as code
monitoring/               # Separate monitoring stack
```

**Our Approach** (Simple):
```go
// Everything embedded in Go
import "github.com/joeblew999/infra"

// All the above functionality available as Go functions:
infra.StartWebServer()
infra.StartDatabase() 
infra.StartMessageQueue()
infra.StartReverseProxy()
infra.Build()
infra.Deploy()
infra.Scale(3)
```

### CLI Development
```bash
# CLI tools and wrappers
go run . tools -h

# Dependency management
go run . tools dep list
go run . tools dep install <binary>

# Go-zero microservices
go run . tools gozero api create myservice
```

## 🧪 Testing & Quality

### Testing Framework
- **Unit Tests**: Standard Go testing with `go test`
- **Integration Tests**: Service startup and API testing
- **Beta Testing**: User acceptance testing procedures
- **API Compatibility**: Backward compatibility checking

### Quality Gates
- **Pre-commit Hooks**: Automated code quality checks
- **Linting**: Code style and best practice enforcement  
- **Type Checking**: Static analysis and type safety
- **Security Scanning**: Dependency and code vulnerability scanning

## 🔌 Integration Points

### External Services
- **NATS JetStream**: Message streaming and event handling
- **PocketBase**: Embedded database with REST API
- **Caddy**: HTTP server and reverse proxy
- **Bento**: Stream processing and data pipelines

### API Endpoints
- **Web Server**: http://localhost:1337/api/
- **Deck API**: http://localhost:8888/api/v1/deck/
- **PocketBase**: http://localhost:8090/api/
- **Health Checks**: http://localhost:1337/status

## 📦 Package Structure

```
pkg/
├── bento/          # Stream processing integration
├── caddy/          # Reverse proxy and web server
├── cmd/            # CLI command implementations  
├── config/         # Configuration management
├── deck/           # Visualization and presentation system
├── dep/            # Dependency management
├── fly/            # Fly.io deployment integration
├── gops/           # Process and port management
├── goreman/        # Process supervision
├── gozero/         # Go-zero microservice framework
├── log/            # Logging system
├── mjml/           # Email template system
├── nats/           # NATS messaging integration
└── pocketbase/     # Database integration
```

## 🚀 Deployment Options

### Local Development
```bash
go run .                    # All services on localhost
go run . runtime up --env development  # Development mode with debug features
```

### Production Deployment
```bash
go run . workflows deploy         # Idempotent Fly.io deployment
go run . workflows status         # Check deployment health
```

### Container Deployment  
```bash
go run . workflows build    # Build container with Ko
docker run -p 1337:1337 infra
```

## 📊 Monitoring & Observability

### Health Checks
- **System Health**: http://localhost:1337/status
- **Service Health**: Individual service health endpoints
- **Process Status**: `go run . runtime status`

### Logging
- **Structured Logging**: JSON format with correlation IDs
- **Log Aggregation**: Centralized logging with NATS
- **Log Levels**: DEBUG, INFO, WARN, ERROR with filtering

### Metrics
- **Application Metrics**: Performance and business metrics
- **System Metrics**: Resource usage and system health  
- **Custom Metrics**: Business-specific measurements

## 🔒 Security Considerations

### Authentication & Authorization
- **Role-Based Access Control**: User permissions and roles
- **Session Management**: Secure session handling
- **API Authentication**: Token-based API security

### Data Security
- **Encryption**: Data encryption at rest and in transit
- **Secret Management**: Secure handling of sensitive data
- **Audit Logging**: Complete audit trail for compliance

### Network Security  
- **HTTPS Only**: Automatic HTTPS with Caddy
- **Rate Limiting**: API and resource protection
- **CORS Configuration**: Cross-origin request security

## 🎓 Contributing

### Development Setup
1. Fork the repository
2. Set up development environment
3. Make your changes  
4. Run tests and quality checks
5. Submit pull request

### Coding Standards
- **Go Style Guide**: Follow standard Go conventions
- **API Design**: Follow established API patterns
- **Documentation**: Comprehensive code and API documentation
- **Testing**: Unit tests for all new functionality

---

**For business users and decision makers**, see the [Business Documentation](../business/) section.
