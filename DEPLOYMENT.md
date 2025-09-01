# Fly.io Deployment Guide

This guide walks you through deploying the infrastructure management system to Fly.io with goreman supervision and scaling capabilities.

## Prerequisites

1. **Fly.io Account**: Sign up at [fly.io](https://fly.io)
2. **Dependencies**: Automatically managed via `go run . dep`
3. **API Token**: Get from [Fly.io dashboard](https://fly.io/dashboard/personal/access_tokens)

## Quick Start

```bash
# 1. Clone and setup
git clone https://github.com/joeblew999/infra.git
cd infra

# 2. Test locally with goreman supervision
go run .                    # Starts all services locally

# 3. Deploy to Fly.io  
go run . deploy             # Idempotent deployment workflow
```

## Detailed Setup

### 1. Environment Setup

```bash
# Set your Fly.io API token
export FLY_API_TOKEN="your-token-here"

# Optional: Set custom app name and region
export FLY_APP_NAME="my-infra-mgmt"
export FLY_REGION="syd"
```

### 2. Deployment Commands

```bash
# Infrastructure management
go run . deploy              # Idempotent deployment workflow
go run . status              # Check deployment health
go run . shutdown            # Stop all services

# Fly.io specific commands
go run . cli fly status      # Fly.io machine status  
go run . cli fly logs        # View application logs
go run . cli fly ssh         # SSH into machine
go run . cli fly deploy      # Direct flyctl deploy

# Scaling (see scaling section below)
go run . cli fly scale       # Show current scaling
```

### 3. Local Development & Testing

Before deploying, test the full stack locally:

```bash
# Start all services with goreman supervision
go run .

# Check service status
go run . -h                  # See organized command help
curl http://localhost:1337/status    # Health endpoint
curl http://localhost:8090/         # PocketBase admin  
curl http://localhost:8888/api/v1/deck/health  # Deck API health
```

All services run under goreman supervision:
- **NATS Server** (4222) - Message streaming
- **PocketBase** (8090) - Database with admin UI
- **Caddy** (80/443) - Reverse proxy  
- **Bento** (4195) - Stream processing
- **Deck API** (8888) - Go-zero visualization API
- **Web Server** (1337) - Main dashboard

### 3. Database Setup with Litestream

```bash
# Start database replication on Fly.io
./.dep/flyctl ssh console

# Inside the machine:
/app # litestream replicate -config /app/litestream.yml

# Or use our CLI for local testing
go run . litestream start --db ./pb_data/data.db --backup ./backups/data.db
```

### 4. Using Terraform (Advanced)

```bash
# Initialize Terraform
cd terraform
./../.dep/tofu init

# Plan the deployment
./../.dep/tofu plan

# Apply the configuration
./../.dep/tofu apply

# Get the app URL
./../.dep/tofu output app_url
```

## Configuration

### fly.toml

The configuration is already set up in `fly.toml`:
- **App**: `infra-mgmt`
- **Region**: `syd`
- **Memory**: 512MB
- **CPU**: 1 shared CPU
- **Volume**: 1GB persistent storage at `/app/.data`

### Environment Variables

Set these in your fly.toml or via secrets:

```toml
[env]
  ENVIRONMENT = "production"
  PORT = "1337"
  LITESTREAM_REPLICA_URL = "s3://your-bucket/backups"
```

### Secrets Management

```bash
# Set AWS credentials for Tigris (S3-compatible storage)
./.dep/flyctl secrets set AWS_ACCESS_KEY_ID=your-key
./.dep/flyctl secrets set AWS_SECRET_ACCESS_KEY=your-secret
./.dep/flyctl secrets set LITESTREAM_REPLICA_URL=s3://your-bucket/backups
```

## Database Backup Strategy

### Automatic Backups with Litestream

1. **Continuous Replication**: Litestream continuously backs up your SQLite database
2. **Point-in-time Recovery**: Restore to any point in the last 24 hours
3. **Zero Downtime**: Backups happen without interrupting service

### Backup Configuration

Create `litestream.yml`:

```yaml
dbs:
  - path: /app/.data/data.db
    replicas:
      - type: s3
        bucket: your-backup-bucket
        path: backups/data.db
        region: auto
        endpoint: https://fly.storage.tigris.dev
        access-key-id: ${AWS_ACCESS_KEY_ID}
        secret-access-key: ${AWS_SECRET_ACCESS_KEY}
        retention: 24h
        sync-interval: 1s
```

### Manual Backup Operations

```bash
# Create backup locally
go run . litestream start --db ./pb_data/data.db --backup ./backups/

# Restore from backup
go run . litestream restore --db ./pb_data/data.db --backup ./backups/ --timestamp 2024-01-01T12:00:00Z

# Check backup status
go run . litestream status
```

## Troubleshooting

### Common Issues

1. **App Won't Start**
   ```bash
   ./.dep/flyctl logs
   ./.dep/flyctl status
   ```

2. **Database Issues**
   ```bash
   ./.dep/flyctl ssh console
   ls -la /app/.data/
   ```

3. **Backup Issues**
   ```bash
   ./.dep/flyctl ssh console
   litestream status -config /app/litestream.yml
   ```

## Scaling and Management

### Scaling Commands

The infrastructure supports comprehensive scaling options:

```bash
# Show current scaling configuration
go run . cli fly scale

# Horizontal scaling (add/remove machines)
go run . cli fly scale --count 2          # Scale to 2 machines
go run . cli fly scale --count 1          # Scale back to 1 machine

# Vertical scaling (resources per machine)  
go run . cli fly scale --memory 1024      # Scale memory to 1GB
go run . cli fly scale --memory 2048      # Scale memory to 2GB
go run . cli fly scale --cpu 2            # Scale to 2 CPU cores

# VM type scaling (machine performance)
go run . cli fly scale --vm shared-cpu-2x      # 2 shared CPUs
go run . cli fly scale --vm performance-2x     # 2 dedicated CPUs

# Combined scaling operations
go run . cli fly scale --count 2 --memory 2048 --cpu 2
```

### Scale Application

```bash
# Scale to multiple instances
go run . cli fly scale --count 2 -a your-app-name

# Scale memory  
go run . cli fly scale --memory 1024 -a your-app-name

# Scale CPU
go run . cli fly scale --vm shared-cpu-2x -a your-app-name
```

### Auto-scaling

The system supports manual scaling as shown above. For automatic scaling based on metrics, you can:

1. **Monitor metrics** via `/metrics` endpoint
2. **Set up alerts** based on CPU/memory usage
3. **Use external tools** to call scaling commands
4. **Implement custom logic** using the NATS event system

### Scaling Best Practices

- **Start small**: Begin with 1 machine and scale as needed
- **Monitor first**: Use `/metrics` and `/status` to understand usage
- **Scale gradually**: Increase resources incrementally  
- **Test scaling**: Verify application works correctly at different scales
- **Regional distribution**: Use `--region` for multi-region deployments

## Monitoring

### Health Checks

The app includes health checks at `/status`:
- **HTTP Check**: Every 10 seconds
- **TCP Check**: Every 15 seconds

### Metrics

Available at `/metrics` on port 9091:
- Application metrics
- Database metrics
- Backup status

## CI/CD Integration

### GitHub Actions

Create `.github/workflows/deploy.yml`:

```yaml
name: Deploy to Fly.io

on:
  push:
    branches: [main]

env:
  FLY_API_TOKEN: ${{ secrets.FLY_API_TOKEN }}

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: superfly/flyctl-actions/setup-flyctl@master
      - run: flyctl deploy --remote-only
```

## Security Considerations

1. **Secrets**: Never commit secrets to git
2. **Volumes**: Data is encrypted at rest
3. **Network**: HTTPS is enforced
4. **Access**: Use SSH keys for machine access

## Cost Optimization

- **Auto-stop**: Machines stop when idle
- **Volume**: Start with 1GB, scale as needed
- **Region**: Choose closest region for lower latency

## Next Steps

1. **Custom Domain**: Add your own domain
2. **Database Scaling**: Consider LiteFS for multi-region
3. **Monitoring**: Set up alerting
4. **Backup Testing**: Regularly test restore procedures