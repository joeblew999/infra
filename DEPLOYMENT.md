# Fly.io Deployment Guide

This guide walks you through deploying the infrastructure management system to Fly.io with automated database backups.

## Prerequisites

1. **Fly.io Account**: Sign up at [fly.io](https://fly.io)
2. **Fly CLI**: Install with `go run . dep install flyctl`
3. **API Token**: Get from [Fly.io dashboard](https://fly.io/dashboard/personal/access_tokens)

## Quick Start

```bash
# 1. Install flyctl
./.dep/flyctl version

# 2. Login to Fly.io
./.dep/flyctl auth login

# 3. Deploy the application
./.dep/flyctl deploy
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

### 2. Using Our CLI

```bash
# Deploy with our custom commands
go run . fly deploy

# Check status
go run . fly status

# View logs
go run . fly logs

# SSH into the machine
go run . fly ssh

# Scale resources
go run . fly scale
```

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

### Scaling

```bash
# Scale memory
./.dep/flyctl scale memory 1GB

# Scale CPU
./.dep/flyctl scale cpu 2

# Scale machines
./.dep/flyctl scale count 2
```

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