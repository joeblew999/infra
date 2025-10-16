# Phase 2: Microservices Deployment Configurations

This directory contains Fly.io configuration files for deploying core V2 as independent microservices.

## Configuration Files

| File | Service | Purpose | Resources |
|------|---------|---------|-----------|
| `fly-nats.toml` | core-nats | Event streaming with JetStream | 512MB, 1GB volume |
| `fly-pocketbase.toml` | core-pocketbase | Database and REST API | 1GB, 5GB volume |
| `fly-caddy.toml` | core-caddy | Edge proxy with TLS | 256MB, no volume |
| `fly-controller.toml` | core-controller | Control plane and observability | 512MB, 1GB volume |

## Deployment Order

Deploy services in this order to ensure dependencies are available:

### 1. Deploy NATS (Event Backbone)
```bash
cd deploy/phase2

# Create app
flyctl apps create core-nats --org personal

# Create volume
flyctl volumes create nats_data --region syd --size 1 --app core-nats

# Deploy
flyctl deploy --config fly-nats.toml --app core-nats
```

### 2. Deploy PocketBase (Database)
```bash
# Create app
flyctl apps create core-pocketbase --org personal

# Create volume
flyctl volumes create pb_data --region syd --size 5 --app core-pocketbase

# Deploy
flyctl deploy --config fly-pocketbase.toml --app core-pocketbase
```

### 3. Deploy Controller (Control Plane)
```bash
# Create app
flyctl apps create core-controller --org personal

# Create volume
flyctl volumes create controller_data --region syd --size 1 --app core-controller

# Deploy
flyctl deploy --config fly-controller.toml --app core-controller
```

### 4. Deploy Caddy (Edge Proxy)
```bash
# Create app
flyctl apps create core-caddy --org personal

# Deploy (no volume needed - stateless)
flyctl deploy --config fly-caddy.toml --app core-caddy
```

### 5. Configure DNS
```bash
# Add custom domain to Caddy
flyctl certs create your-domain.com --app core-caddy

# Or use Fly.io subdomain
# https://core-caddy.fly.dev
```

## Service Communication

All services communicate via Fly.io private network (6PN):

```
core-nats.internal:4222       → NATS client connections
core-nats.internal:8222       → NATS HTTP monitoring
core-pocketbase.internal:8090 → PocketBase API
core-controller.internal:8080 → Controller API
```

Services use these URLs via environment variables (already configured in toml files).

## Verification

After deployment, verify each service:

### Check NATS
```bash
flyctl logs --app core-nats
flyctl status --app core-nats

# Test from another app
curl http://core-nats.internal:8222/healthz
```

### Check PocketBase
```bash
flyctl logs --app core-pocketbase
flyctl status --app core-pocketbase

# Test from Caddy
curl http://core-pocketbase.internal:8090/api/health
```

### Check Controller
```bash
flyctl logs --app core-controller
flyctl status --app core-controller
```

### Check Caddy (Public)
```bash
flyctl logs --app core-caddy
flyctl status --app core-caddy

# Test public endpoint
curl https://core-caddy.fly.dev/api/health
```

## Monitoring

Monitor all services from one command:

```bash
# Status of all apps
flyctl status --app core-nats
flyctl status --app core-pocketbase
flyctl status --app core-controller
flyctl status --app core-caddy

# Logs from all services
flyctl logs --app core-nats &
flyctl logs --app core-pocketbase &
flyctl logs --app core-controller &
flyctl logs --app core-caddy &
```

## Scaling

Scale services independently:

```bash
# Scale edge proxy (most likely to need scaling)
flyctl scale count 3 --app core-caddy

# Scale memory if needed
flyctl scale vm shared-cpu-2x --memory 2048 --app core-pocketbase

# NATS clustering (requires configuration changes)
flyctl scale count 3 --app core-nats
# Then update fly-nats.toml with cluster routes
```

## Rollback

Rollback individual services without affecting others:

```bash
# View deployment history
flyctl releases --app core-caddy

# Rollback to previous version
flyctl releases rollback v2 --app core-caddy
```

## Secrets Management

Set secrets for services:

```bash
# PocketBase admin credentials
flyctl secrets set ADMIN_EMAIL=admin@example.com --app core-pocketbase
flyctl secrets set ADMIN_PASSWORD=<secure-password> --app core-pocketbase

# NATS authentication (if enabled)
flyctl secrets set NATS_USER=core --app core-nats
flyctl secrets set NATS_PASSWORD=<secure-password> --app core-nats

# Cloudflare tokens (for Caddy DNS challenges)
flyctl secrets set CLOUDFLARE_API_TOKEN=<token> --app core-caddy
```

## Cost Estimate

Based on Fly.io pricing:

| Service | Resources | Cost/Month |
|---------|-----------|------------|
| core-nats | 512MB, 1GB vol | ~$4 |
| core-pocketbase | 1GB, 5GB vol | ~$6 |
| core-controller | 512MB, 1GB vol | ~$4 |
| core-caddy | 256MB (×1-3) | ~$2-6 |
| **Total** | | **~$16-20/month** |

Compare to Phase 1 monolithic: ~$5-10/month

## Troubleshooting

### Service Can't Reach Another Service

Check private network connectivity:

```bash
# SSH into one service
flyctl ssh console --app core-caddy

# Test connectivity to another service
curl http://core-pocketbase.internal:8090/api/health
ping core-nats.internal
```

### Volume Mount Issues

```bash
# Check volume status
flyctl volumes list --app core-pocketbase

# Volume not mounted? Recreate it
flyctl volumes destroy <volume-id> --app core-pocketbase
flyctl volumes create pb_data --region syd --size 5 --app core-pocketbase
```

### Health Checks Failing

```bash
# View detailed health check results
flyctl status --app core-nats

# Check service logs for errors
flyctl logs --app core-nats

# Manually test health endpoint
flyctl ssh console --app core-nats
curl http://localhost:8222/healthz
```

## Migration from Phase 1

To migrate from monolithic (Phase 1) to microservices (Phase 2):

1. **Deploy Phase 2 services** (follow deployment order above)
2. **Verify all services healthy** (check logs and status)
3. **Update DNS** to point to core-caddy instead of core-v2
4. **Test thoroughly** with Phase 2 deployment
5. **Scale down Phase 1** once confident: `flyctl scale count 0 --app core-v2`
6. **Keep Phase 1 for rollback** for 1-2 weeks before deleting

## Rollback to Phase 1

If Phase 2 has issues:

```bash
# Scale down Phase 2
flyctl scale count 0 --app core-nats
flyctl scale count 0 --app core-pocketbase
flyctl scale count 0 --app core-controller
flyctl scale count 0 --app core-caddy

# Scale up Phase 1
flyctl scale count 1 --app core-v2

# Update DNS back to Phase 1
```

## Next Steps

1. Deploy to staging environment first
2. Load test Phase 2 architecture
3. Monitor for 1 week before production migration
4. Document any issues and optimizations
5. Create automated deployment scripts

For detailed architecture documentation, see [../../docs/MICROSERVICES_ARCHITECTURE.md](../../docs/MICROSERVICES_ARCHITECTURE.md)
