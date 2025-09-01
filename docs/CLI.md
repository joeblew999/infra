# CLI Reference

This document covers the infrastructure management CLI, which is organized into logical command groups for better usability.

## CLI Organization

The CLI uses a hierarchical structure with organized help output:

```bash
go run . -h      # Show organized help with command categories
go run . cli -h  # Show all available CLI tools organized by purpose
```

### Command Categories

1. **Infrastructure Commands** - Core service management
2. **Development Commands** - Essential development tools  
3. **Advanced Commands** - API compatibility and completion
4. **CLI Tools** - Specialized tools organized by purpose:
   - **Scaling & Deployment** - Fly.io operations
   - **Development Tools** - Deck and go-zero 
   - **Binary Tools** - Direct access to managed binaries

## Infrastructure Commands

### Service Management

Start and stop all services with goreman supervision:

```bash
go run .                    # Start all services (default command)
go run . service             # Same as root command
go run . shutdown            # Graceful shutdown of all services
```

Services started automatically:
- **NATS Server** (4222) - Message streaming
- **PocketBase** (8090) - Database with admin UI
- **Caddy** (80/443) - Reverse proxy
- **Bento** (4195) - Stream processing  
- **Deck API** (8888) - Go-zero visualization API
- **Deck Watcher** - File processing service
- **Web Server** (1337) - Main dashboard

### Deployment & Status

```bash
go run . deploy             # Deploy to Fly.io with idempotent workflow
go run . status             # Check deployment status and health
```

The deploy command provides:
- Idempotent deployment (safe to run multiple times)
- Automatic app and volume creation
- Container image building with ko
- Health check verification

## Development Commands

### Configuration

```bash
go run . config             # Print current configuration paths and settings
```

Shows:
- Binary dependency paths
- Data directory locations  
- Service configuration
- Environment settings

### Dependency Management

```bash
go run . dep                # Manage binary dependencies
go run . dep list           # List all configured dependencies
go run . dep install <name> # Install specific dependency
go run . dep status         # Show installation status
```

Supported dependencies include flyctl, ko, caddy, bento, and more.


### Project Initialization

```bash
go run . init               # Initialize new project with standard configuration
```

## Advanced Commands

### API Compatibility

```bash
go run . api-check          # Check API compatibility between commits
go run . api-check --old HEAD~2 --new HEAD  # Compare specific commits
```

### CLI Tool Wrappers

```bash
go run . cli                # Access CLI tool wrappers namespace
go run . cli fly            # Fly.io commands
go run . cli fly scale      # Scaling operations
go run . cli fly status     # Fly.io machine status
go run . cli fly logs       # Application logs
go run . cli fly ssh        # SSH into machine
go run . cli deck           # Deck visualization tools
go run . cli deck watch     # Watch .dsh files and generate outputs
go run . cli deck build     # Build single deck file
go run . cli gozero         # Go-zero microservices operations
go run . cli gozero api create  # Create new go-zero API service
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
go run . cli fly scale

# Horizontal scaling (machines)
go run . cli fly scale --count 2          # Scale to 2 machines
go run . cli fly scale --count 1          # Scale back to 1 machine

# Vertical scaling (resources per machine)
go run . cli fly scale --memory 1024      # Scale memory to 1GB
go run . cli fly scale --memory 2048      # Scale memory to 2GB  
go run . cli fly scale --cpu 2            # Scale to 2 CPU cores

# VM type scaling (machine performance)
go run . cli fly scale --vm shared-cpu-2x      # 2 shared CPUs
go run . cli fly scale --vm performance-2x     # 2 dedicated CPUs

# Combined operations
go run . cli fly scale --count 2 --memory 2048 --cpu 2
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

#### CLI Tools (`go run . cli -h`)
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
go run . deploy -h          # Deployment options and flags
go run . cli fly scale -h   # Scaling options and examples
go run . dep -h             # Dependency management help
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
go run . --env development
```
- Enhanced logging
- Debug endpoints enabled
- Development-specific configurations

**Production Mode (default):**
```bash
go run . --env production
```
- Optimized performance
- Production logging levels
- Security hardening

## Common Workflows

### Local Development

```bash
# Start all services for development
go run . --env development

# Check service health
curl http://localhost:1337/status

# View configuration
go run . config

# Stop all services
go run . shutdown
```

### Production Deployment

```bash
# Deploy to production
go run . deploy

# Check deployment status
go run . status

# Scale if needed
go run . cli fly scale --count 2 --memory 1024

# Monitor logs
go run . cli fly logs
```

### Debugging

```bash
# Check API compatibility
go run . api-check

# SSH into production machine
go run . cli fly ssh  

# Check deck service health
curl http://localhost:8888/api/v1/deck/health

# Build deck visualization
go run . cli deck build myfile.dsh

# Create go-zero service
go run . cli gozero api create myservice --output ./api/myservice
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
   go run . shutdown    # Stop existing services
   go run .             # Restart clean
   ```

2. **Deployment Failures**
   ```bash
   go run . status      # Check deployment status
   go run . cli fly logs # View error logs
   ```

3. **Scaling Issues**
   ```bash
   go run . cli fly status   # Check machine status
   go run . cli fly scale    # Verify current scaling
   ```

### Debug Mode

Enable debug mode for verbose output:

```bash
go run . --debug deploy     # Debug deployment
go run . --debug --env development  # Debug development mode
```