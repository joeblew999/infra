# Deployment Guide

This guide explains how to deploy the infrastructure services to Fly.io using the `tooling/core-tool` CLI.

## Prerequisites

1. **Fly.io Account**: Sign up at https://fly.io
2. **Authentication**: Run `tooling/core-tool auth fly` to authenticate
3. **Organization**: Have a Fly.io organization created

## Build Configuration

Services are built using [ko](https://github.com/ko-build/ko) which creates minimal container images from Go binaries.

**Build Configuration**: [.ko.yaml](.ko.yaml)

Configured services:
- `core-app` - Main application (cmd/core)
- `nats` - NATS message broker with Pillow clustering (cmd/nats)
- `pocketbase` - PocketBase with Datastar auth (cmd/pocketbase)
- `pocketbase-ha` - PocketBase HA with go-ha driver (cmd/pocketbase-ha) **[CGO_ENABLED=1]**

## Deployment Architecture

### Single Region (Development/Testing)

```
Region: iad (Virginia)
┌────────────────────────────────┐
│  NATS (infra-nats-iad)         │
│  - Port 4222 (client)          │
│  - Port 6222 (cluster)         │
│  - Port 7422 (leaf)            │
│  - Port 8222 (HTTP monitoring) │
└────────────┬───────────────────┘
             │
             ↓
┌────────────────────────────────┐
│  PocketBase (infra-pb-iad)     │
│  - Port 8090 (HTTP/HTTPS)      │
│  - Connects to NATS for events │
└────────────────────────────────┘
```

### Multi-Region (Production HA)

```
Region: iad (Hub)              Region: lhr (Leaf)          Region: nrt (Leaf)
┌─────────────────┐           ┌─────────────────┐         ┌─────────────────┐
│ NATS (3 nodes)  │◄─────────►│ NATS (1 node)   │◄───────►│ NATS (1 node)   │
│ Hub cluster     │           │ Leaf node       │         │ Leaf node       │
└────────┬────────┘           └────────┬────────┘         └────────┬────────┘
         │                             │                           │
         ↓                             ↓                           ↓
┌─────────────────┐           ┌─────────────────┐         ┌─────────────────┐
│ PB-HA (3 nodes) │           │ PB-HA (2 nodes) │         │ PB-HA (2 nodes) │
│ go-ha + NATS    │           │ go-ha + NATS    │         │ go-ha + NATS    │
└─────────────────┘           └─────────────────┘         └─────────────────┘
```

## Deployment Steps

### 1. Authenticate

```bash
cd tooling
./core-tool auth fly
./core-tool auth status  # Verify authentication
```

### 2. Deploy NATS (Foundation)

NATS must be deployed first as other services depend on it.

```bash
# Deploy to primary region (iad)
./core-tool workflow deploy \
  --app infra-nats-iad \
  --org personal \
  --region iad \
  --repo registry.fly.io/infra-nats-iad

# For multi-region, deploy to additional regions
./core-tool workflow deploy \
  --app infra-nats-lhr \
  --org personal \
  --region lhr \
  --repo registry.fly.io/infra-nats-lhr
```

**Pillow Auto-Clustering**: NATS nodes will automatically discover each other via Fly.io's private network and form a cluster.

**Verify NATS**:
```bash
flyctl ssh console -a infra-nats-iad
# Inside the container:
curl http://localhost:8222/varz  # Check server status
```

### 3. Deploy PocketBase

Once NATS is running, deploy PocketBase:

```bash
./core-tool workflow deploy \
  --app infra-pocketbase-iad \
  --org personal \
  --region iad \
  --repo registry.fly.io/infra-pocketbase-iad
```

**Set Secrets** (for production):
```bash
flyctl secrets set \
  CORE_POCKETBASE_ADMIN_EMAIL=admin@example.com \
  CORE_POCKETBASE_ADMIN_PASSWORD=<strong-password> \
  --app infra-pocketbase-iad
```

**Access PocketBase**:
- Web UI: `https://infra-pocketbase-iad.fly.dev/ds`
- Admin UI: `https://infra-pocketbase-iad.fly.dev/_/`
- API: `https://infra-pocketbase-iad.fly.dev/api/`

### 4. Deploy PocketBase-HA (Optional)

For high availability with automatic replication:

```bash
# Note: Requires CGO_ENABLED=1 (handled by .ko.yaml)
./core-tool workflow deploy \
  --app infra-pocketbase-ha-iad \
  --org personal \
  --region iad \
  --repo registry.fly.io/infra-pocketbase-ha-iad
```

**Scale to multiple nodes**:
```bash
flyctl scale count 3 --app infra-pocketbase-ha-iad
```

Each node will:
1. Connect to NATS cluster
2. Sync SQLite database via go-ha driver
3. Handle requests independently
4. Resolve conflicts automatically

## Service Configuration

### NATS Service

**fly.toml**: [services/nats/fly.toml](services/nats/fly.toml)

**Ports**:
- 4222: Client connections
- 6222: Cluster routing
- 7422: Leaf node connections
- 8222: HTTP monitoring

**Environment**:
- `NATS_PILLOW_HUB_AND_SPOKE=false`: Use mesh topology (default)
- `NATS_PILLOW_HUB_AND_SPOKE=true`: Use hub-and-spoke topology

**Resources**:
- Memory: 1GB
- CPU: 1 shared vCPU
- Storage: 10GB persistent volume

### PocketBase Service

**fly.toml**: [services/pocketbase/fly.toml](services/pocketbase/fly.toml)

**Ports**:
- 80/443: HTTP/HTTPS (Datastar UI + API)

**Environment**:
- `CORE_POCKETBASE_APP_URL`: Public URL
- `PB_REPLICATION_URL`: NATS connection (optional)
- `CORE_POCKETBASE_ADMIN_EMAIL`: Bootstrap admin email
- `CORE_POCKETBASE_ADMIN_PASSWORD`: Bootstrap admin password

**Resources**:
- Memory: 512MB
- CPU: 1 shared vCPU
- Storage: 10GB persistent volume

### PocketBase-HA Service

**fly.toml**: [services/pocketbase-ha/fly.toml](services/pocketbase-ha/fly.toml) *(to be created)*

**Ports**: Same as PocketBase

**Environment**:
- `PB_NAME`: Unique node identifier
- `PB_REPLICATION_URL`: NATS URL for sync
- `PB_REPLICATION_STREAM`: Stream name (default: "pb")
- All PocketBase env vars

**Resources**:
- Memory: 1GB (more than single-node for replication)
- CPU: 1 shared vCPU
- Storage: 10GB persistent volume per node

## Scaling

### NATS

**Horizontal Scaling**:
```bash
# Scale within region
flyctl scale count 3 --app infra-nats-iad

# Add new region
flyctl launch --copy-config --region lhr --app infra-nats-lhr
```

Pillow will automatically:
- Cluster nodes within the same region
- Create supercluster across regions
- Configure leaf nodes appropriately

### PocketBase (Single-Node)

**Vertical Scaling**:
```bash
# Increase memory
flyctl scale memory 1024 --app infra-pocketbase-iad

# Increase CPUs
flyctl scale vm dedicated-cpu-1x --app infra-pocketbase-iad
```

**Note**: Single-node PocketBase does not support horizontal scaling. Use PocketBase-HA for multi-node.

### PocketBase-HA

**Horizontal Scaling**:
```bash
# Scale to 3 nodes in region
flyctl scale count 3 --app infra-pocketbase-ha-iad

# Deploy to additional regions
flyctl launch --copy-config --region lhr --app infra-pocketbase-ha-lhr
flyctl scale count 2 --app infra-pocketbase-ha-lhr
```

All nodes will sync via NATS automatically.

## Monitoring

### NATS Monitoring

**HTTP Monitoring Endpoint**:
```bash
curl https://infra-nats-iad.fly.dev:8222/varz
curl https://infra-nats-iad.fly.dev:8222/connz
curl https://infra-nats-iad.fly.dev:8222/routez
```

**Fly.io Metrics**:
```bash
flyctl logs -a infra-nats-iad
flyctl status -a infra-nats-iad
```

### PocketBase Monitoring

**Health Check**:
```bash
curl https://infra-pocketbase-iad.fly.dev/api/health
```

**Logs**:
```bash
flyctl logs -a infra-pocketbase-iad
```

**Admin Dashboard**: Access via `https://infra-pocketbase-iad.fly.dev/_/`

## Troubleshooting

### NATS Not Clustering

1. Check Fly.io private network connectivity:
   ```bash
   flyctl ssh console -a infra-nats-iad
   ping infra-nats-lhr.internal
   ```

2. Check NATS cluster status:
   ```bash
   curl http://localhost:8222/routez
   ```

3. Verify Pillow environment:
   ```bash
   flyctl ssh console -a infra-nats-iad
   env | grep NATS
   env | grep FLY
   ```

### PocketBase Connection Issues

1. Check if NATS is reachable:
   ```bash
   flyctl ssh console -a infra-pocketbase-iad
   nc -zv infra-nats-iad.internal 4222
   ```

2. Check PocketBase logs:
   ```bash
   flyctl logs -a infra-pocketbase-iad
   ```

3. Verify environment variables:
   ```bash
   flyctl ssh console -a infra-pocketbase-iad
   env | grep CORE_POCKETBASE
   ```

### PocketBase-HA Replication Issues

1. Check go-ha driver connection:
   ```bash
   flyctl logs -a infra-pocketbase-ha-iad | grep "replication"
   ```

2. Verify all nodes see each other:
   ```bash
   # On each node
   curl http://localhost:8090/api/health
   ```

3. Check NATS stream:
   ```bash
   flyctl ssh console -a infra-nats-iad
   # Use NATS CLI to inspect stream "pb"
   ```

## Rollback

If a deployment fails:

```bash
# List releases
flyctl releases -a infra-pocketbase-iad

# Rollback to previous version
flyctl releases rollback <version> -a infra-pocketbase-iad
```

## Maintenance

### Update Services

```bash
# Rebuild and redeploy
cd tooling
./core-tool workflow deploy --app <app-name> --org personal
```

### Backup PocketBase Data

1. **Using Fly.io Volumes**:
   ```bash
   flyctl volumes list -a infra-pocketbase-iad
   flyctl volumes snapshot create <volume-id>
   ```

2. **Using Litestream** (recommended for production):
   - Configure in `.env` or secrets
   - Automatic continuous backups to S3

### Database Migrations

PocketBase handles migrations automatically on startup. For HA setups:
1. Stop all nodes
2. Run migration on one node
3. Start all nodes

## Cost Estimation

**Development (Single Region)**:
- NATS: ~$5-10/month (1 machine, 1GB RAM)
- PocketBase: ~$5/month (1 machine, 512MB RAM)
- **Total**: ~$10-15/month

**Production (3 Regions, HA)**:
- NATS: ~$45/month (5 machines across regions)
- PocketBase-HA: ~$60/month (7 machines across regions)
- **Total**: ~$105/month

*Prices based on Fly.io shared-cpu-1x pricing as of 2025*

## Additional Resources

- [Fly.io Documentation](https://fly.io/docs/)
- [ko Build Tool](https://github.com/ko-build/ko)
- [Pillow NATS](https://github.com/Nintron27/pillow)
- [PocketBase Docs](https://pocketbase.io/docs/)
- [go-ha Driver](https://github.com/litesql/go-ha)
- [NATS Documentation](https://docs.nats.io/)
