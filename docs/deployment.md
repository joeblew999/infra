# Fly.io Deployment Guide

This guide covers deploying the Infrastructure Management System to Fly.io using multiple deployment methods.

## Prerequisites

1. **Fly.io Account**: Sign up at [fly.io](https://fly.io)
2. **Go 1.22+**: For building the application
3. **Git**: For version control

## Quick Start Deployment

### Option 1: Idempotent Workflow (Recommended)
```bash
# Test deployment without executing
go run . deploy --dry-run

# Run full idempotent deployment
go run . deploy

# Deploy to specific app and region
go run . deploy --app my-app --region syd
```

### Option 2: Manual Step-by-Step

#### Step 1: Install and Configure Flyctl
```bash
# Install flyctl via our dependency manager
go run . flyctl version

# Authenticate with Fly.io
go run . flyctl auth login
```

#### Step 2: Create Fly.io App
```bash
# Create new app (change app name as needed)
go run . flyctl apps create infra-mgmt

# Or use a generated name
go run . flyctl apps create --generate-name
```

#### Step 3: Deploy with Ko (Containerless)
```bash
# Set environment variables
export FLY_APP_NAME=your-app-name
export KO_DOCKER_REPO=registry.fly.io/${FLY_APP_NAME}
export ENVIRONMENT=production

# Build and deploy
go run . ko build github.com/joeblew999/infra

# Deploy the built image
go run . flyctl deploy --app ${FLY_APP_NAME}
```

#### Step 4: Configure Secrets
```bash
# Set production environment
go run . flyctl secrets set ENVIRONMENT=production -a ${FLY_APP_NAME}

# Add any other required secrets
go run . flyctl secrets set DATABASE_URL=your-db-url -a ${FLY_APP_NAME}
```

### Option 3: Traditional Docker Build
```bash
# Deploy using Dockerfile
go run . flyctl deploy --app your-app-name
```

## Configuration

### Environment Variables
- `ENVIRONMENT=production` - Enables production mode
- `PORT=1337` - Web server port (configured in fly.toml)
- `FLY_APP_NAME` - Automatically set by Fly.io

### Persistent Storage
The application uses a persistent volume mounted at `/app/.data` for:
- NATS data storage
- Application state
- Logs and metrics

### Health Checks
- **HTTP Check**: `GET /status` - Application health
- **TCP Check**: Port 1337 - Service availability

## Scaling and Management

### Scale Application
```bash
# Scale to multiple instances
go run . flyctl scale count 2 -a your-app-name

# Scale memory
go run . flyctl scale memory 1024 -a your-app-name

# Scale CPU
go run . flyctl scale vm shared-cpu-2x -a your-app-name
```

### Monitor Application
```bash
# View logs
go run . flyctl logs -a your-app-name

# Real-time logs
go run . flyctl logs -f -a your-app-name

# Application status
go run . flyctl status -a your-app-name

# SSH into running instance
go run . flyctl ssh console -a your-app-name
```

### Manage Volumes
```bash
# List volumes
go run . flyctl volumes list -a your-app-name

# Create additional volume
go run . flyctl volumes create infra_data --size 2 -a your-app-name

# Backup volume
go run . flyctl volumes snapshots create vol_xyz -a your-app-name
```

## CI/CD with GitHub Actions

### Setup
1. Add repository secrets in GitHub:
   - `FLY_API_TOKEN`: Get from `flyctl auth token`
   - `FLY_APP_NAME`: Your Fly.io app name

2. Push to main branch triggers automatic deployment

### Workflow Features
- ✅ Automated testing
- ✅ Ko-based container builds
- ✅ Production deployment
- ✅ Deployment verification

## Advanced Configuration

### Custom Domains
```bash
# Add custom domain
go run . flyctl certs create your-domain.com -a your-app-name

# Check certificate status
go run . flyctl certs show your-domain.com -a your-app-name
```

### Database Integration
```bash
# Create Fly.io Postgres
go run . flyctl postgres create --name infra-db

# Attach to app
go run . flyctl postgres attach infra-db -a your-app-name
```

### Redis Cache
```bash
# Create Redis instance
go run . flyctl redis create --name infra-redis

# Get connection details
go run . flyctl redis status infra-redis
```

## Troubleshooting

### Common Issues

**Build Failures:**
```bash
# Check build logs
go run . flyctl logs -a your-app-name

# Verify ko configuration
go run . ko build --push=false --oci-layout-path=./build github.com/joeblew999/infra
```

**Connection Issues:**
```bash
# Check app status
go run . flyctl status -a your-app-name

# Verify health checks
curl https://your-app-name.fly.dev/status
```

**Performance Issues:**
```bash
# Monitor metrics
go run . flyctl metrics -a your-app-name

# Check resource usage
go run . flyctl ssh console -a your-app-name
top
```

### Debug Commands
```bash
# Enter debug mode
go run . flyctl ssh console -a your-app-name

# Check service status
systemctl status your-service

# View application logs
tail -f /var/log/your-app.log
```

## Cost Optimization

### Resource Allocation
- **Development**: shared-cpu-1x, 256MB RAM
- **Production**: shared-cpu-2x, 512MB RAM
- **High Traffic**: dedicated-cpu-1x, 1GB RAM

### Auto-scaling
```toml
# In fly.toml
[services.concurrency]
  type = "connections"
  hard_limit = 1000
  soft_limit = 1000

auto_stop_machines = true
auto_start_machines = true
min_machines_running = 1
```

## Security

### Best Practices
- ✅ Secrets via Fly.io secrets (not environment variables)
- ✅ HTTPS enforced by default
- ✅ Non-root user in container
- ✅ Minimal base image (Chainguard Static)
- ✅ Regular security updates

### Monitoring
- Application logs via `flyctl logs`
- Metrics via `/metrics` endpoint
- Health checks via `/status` endpoint

## Support

- **Fly.io Docs**: https://fly.io/docs/
- **Ko Documentation**: https://ko.build/
- **Application Issues**: Check logs and status endpoints