# CLI Reference

This document covers the infrastructure management CLI, which is organized into logical command groups for better usability.

## CLI Organization

The CLI uses a hierarchical structure with organized help output:

```bash
go run . -h           # Show organized help with command categories
go run . runtime -h   # Service supervisor commands
go run . tools -h     # Tooling namespace
go run . workflows -h # Build & deploy workflows
go run . dev -h       # Developer utilities
```

### Command Categories

1. **Runtime** – Core service supervision
2. **Workflows** – Build, deploy, and maintenance flows
3. **Tools** – Wrapper commands for installed binaries and UI helpers
4. **Dev** – Developer-oriented diagnostics (API compatibility, etc.)
5. **Advanced** – Shell completion and other meta commands

## Runtime Commands

### Service Supervision

Start and stop all services under goreman supervision:

```bash
go run . runtime up                    # Start services
go run . runtime down                  # Graceful shutdown
go run . runtime status                # Summarize running services
go run . runtime list                  # Show configured services
go run . runtime watch --types status  # Stream live lifecycle events
```

Services started automatically:
- **NATS Server** (4222) - Message streaming
- **PocketBase** (8090) - Database with admin UI
- **Caddy** (80/443) - Reverse proxy
- **Bento** (4195) - Stream processing  
- **Deck API** (8888) - Go-zero visualization API
- **Deck Watcher** - File processing service
- **Web Server** (1337) - Main dashboard

### Containerized Runtime

```bash
go run . runtime container  # Build & run via ko + Docker
```

## Workflow Commands

```bash
go run . workflows deploy   # Deploy to Fly.io with idempotent workflow
go run . workflows status   # Inspect Fly.io deployment health
go run . workflows build    # Build production container image
```

## Tooling Namespace

```bash
go run . tools flyctl status   # Direct flyctl access
go run . tools deck watch      # Deck visualization helpers
go run . tools gozero api      # go-zero scaffolding
go run . tools ko version      # Invoke managed binaries
```

## Developer Utilities

```bash
go run . dev api-check --old v1.2.0 --new HEAD   # Compare Go API surfaces
```

## Advanced Commands

### API Compatibility

```bash
go run . dev api-check --old HEAD~2 --new HEAD  # Compare specific commits
```

### CLI Tool Wrappers

```bash
go run . tools flyctl status   # Fly.io commands
go run . tools deck watch      # Watch .dsh files and generate outputs
go run . tools deck build      # Build single deck file
go run . tools gozero api create  # Create new go-zero API service
```

### Shell Completion

```bash
go run . completion bash    # Generate bash completion
go run . completion zsh     # Generate zsh completion  
go run . completion fish    # Generate fish completion
```

## Scaling Commands

### Fly.io Scaling

All scaling operations are under the `cli fly` namespace:

```bash
# Show current scaling configuration
go run . tools fly scale

# Horizontal scaling (machines)
go run . tools fly scale --count 2          # Scale to 2 machines
go run . tools fly scale --count 1          # Scale back to 1 machine

# Vertical scaling (resources per machine)
go run . tools fly scale --memory 1024      # Scale memory to 1GB
go run . tools fly scale --memory 2048      # Scale memory to 2GB  
go run . tools fly scale --cpu 2            # Scale to 2 CPU cores

# VM type scaling (machine performance)
go run . tools fly scale --vm shared-cpu-2x      # 2 shared CPUs
go run . tools fly scale --vm performance-2x     # 2 dedicated CPUs

# Combined operations
go run . tools fly scale --count 2 --memory 2048 --cpu 2
```

### Scaling Options

**Machine Count (`--count`):**
- Horizontal scaling for increased capacity
- Each machine gets its own persistent volume
- Automatic load balancing across machines

**Memory (`--memory`):**
- 256, 512, 1024, 2048, 4096, 8192 MB options
- Scale based on application memory usage
- Monitor via `/metrics` endpoint

**CPU (`--cpu`):**
- 1, 2, 4, 8 CPU cores (depending on VM type)
- Scale for CPU-intensive workloads
- Shared vs dedicated CPU performance

**VM Type (`--vm`):**
- `shared-cpu-*x`: Shared CPU, lower cost, burstable
- `performance-*x`: Dedicated CPU, consistent performance

## Help System

### Organized Help Output

The CLI provides organized help to reduce command overwhelm:

#### Root Command (`go run . -h`)
```
Infrastructure Management System

QUICK START:
  infra          Start all services (NATS, Caddy, Bento, Deck API, Web Server)
  infra shutdown Stop all services

INFRASTRUCTURE COMMANDS:
  service        Run infrastructure services with goreman supervision
  shutdown       Kill running service processes
  status         Check deployment status and health
  deploy         Deploy application using idempotent workflow

DEVELOPMENT COMMANDS:
  config         Print current configuration
  dep            Manage binary dependencies
  init           Initialize new project

ADVANCED COMMANDS:
  api-check      Check API compatibility between commits
  cli            CLI tool wrappers
  completion     Generate shell autocompletion
```

#### CLI Tools (`go run . tools -h`)
```
SCALING & DEPLOYMENT:
  fly              Fly.io operations and scaling
  
DEVELOPMENT TOOLS:
  deck             Deck visualization tools  
  gozero           Go-zero microservices operations
  
BINARY TOOLS:
  tofu             OpenTofu infrastructure as code
  task             Task runner and build automation
  caddy            Web server with automatic HTTPS
  ko               Container image builder for Go
  flyctl           Direct flyctl commands
  nats             NATS messaging operations
  bento            Stream processing operations
```

### Individual Command Help

Each command provides detailed help:

```bash
go run . workflows deploy -h          # Deployment options and flags
go run . tools fly scale -h   # Scaling options and examples
go run . tools dep -h             # Dependency management help
```

## Environment Variables

### Global Flags

```bash
--env string     # Environment: production or development (default "production")  
--debug         # Enable debug mode
```

### Environment-specific Behavior

**Development Mode:**
```bash
go run . runtime up --env development
```
- Enhanced logging
- Debug endpoints enabled
- Development-specific configurations

**Production Mode (default):**
```bash
go run . runtime up --env production
```
- Optimized performance
- Production logging levels
- Security hardening

## Common Workflows

### Local Development

```bash
# Start all services for development
go run . runtime up --env development

# Check service health
curl http://localhost:1337/status

# View configuration
go run . config

# Stop all services
go run . runtime down
```

### Production Deployment

```bash
# Deploy to production
go run . workflows deploy

# Check deployment status
go run . workflows status

# Scale if needed
go run . tools fly scale --count 2 --memory 1024

# Monitor logs
go run . tools fly logs
```

### Debugging

```bash
# Check API compatibility
go run . dev api-check

# SSH into production machine
go run . tools fly ssh  

# Check deck service health
curl http://localhost:8888/api/v1/deck/health

# Build deck visualization
go run . tools deck build myfile.dsh

# Create go-zero service
go run . tools gozero api create myservice --output ./api/myservice
```

## Command Completion

The CLI supports shell completion for improved productivity:

```bash
# Bash
echo 'source <(go run . completion bash)' >> ~/.bashrc

# Zsh  
echo 'source <(go run . completion zsh)' >> ~/.zshrc

# Fish
go run . completion fish | source
```

## Error Handling

### Common Issues

1. **Port Already in Use**
   ```bash
   go run . runtime down    # Stop existing services
   go run . runtime up      # Restart clean
   ```

2. **Deployment Failures**
   ```bash
   go run . workflows status      # Check deployment status
   go run . tools fly logs # View error logs
   ```

3. **Scaling Issues**
   ```bash
   go run . tools fly status   # Check machine status
   go run . tools fly scale    # Verify current scaling
   ```

### Debug Mode

Enable debug mode for verbose output:

```bash
go run . --debug deploy     # Debug deployment
go run . --debug --env development  # Debug development mode
```
