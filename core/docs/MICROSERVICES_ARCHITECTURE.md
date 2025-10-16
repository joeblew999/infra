# Microservices Architecture (Phase 2)

## Overview

Phase 2 transforms the monolithic deployment into a microservices architecture where each service runs as a separate Fly.io app with independent scaling, health management, and persistent storage.

## Architecture Design

### Service Breakdown

```
┌─────────────────────────────────────────────────────────────┐
│                      Fly.io Platform                        │
│                                                             │
│  ┌──────────────┐                                          │
│  │  core-caddy  │  ← Public edge (*.core-v2.fly.dev)      │
│  │  (Proxy)     │     Handles TLS, routing, load balancing │
│  └──────┬───────┘                                          │
│         │ Private Network (6PN)                            │
│    ┌────┼─────────────────┬─────────────────┐             │
│    │    │                 │                 │             │
│ ┌──▼────▼───┐  ┌─────────▼──────┐  ┌──────▼────────┐    │
│ │ core-nats │  │ core-pocketbase │  │ core-controller│    │
│ │ (Events)  │  │ (Database/API)  │  │ (Orchestration)│    │
│ └───────────┘  └─────────────────┘  └───────────────┘    │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

### Service Definitions

#### 1. core-caddy (Edge Proxy)
**Purpose**: Public-facing reverse proxy and TLS termination

**Configuration**:
- **Exposed**: HTTPS (443) → Internet
- **Internal**: HTTP (2015) → Private network
- **Scaling**: 1-3 instances (auto-scale on traffic)
- **Volume**: None (stateless)
- **Memory**: 256MB per instance

**Routes**:
```
https://core-v2.fly.dev/         → core-pocketbase:8090 (Web UI)
https://core-v2.fly.dev/api/*    → core-pocketbase:8090 (API)
https://core-v2.fly.dev/_/*      → core-pocketbase:8090 (Admin)
https://nats.core-v2.fly.dev/    → core-nats:8222      (Monitoring)
```

**Environment**:
```env
POCKETBASE_URL=http://core-pocketbase.internal:8090
NATS_URL=http://core-nats.internal:8222
```

---

#### 2. core-nats (Event Streaming)
**Purpose**: NATS JetStream for event-driven communication

**Configuration**:
- **Exposed**: None (private network only)
- **Internal**:
  - Client: `nats://core-nats.internal:4222`
  - HTTP Monitoring: `http://core-nats.internal:8222`
- **Scaling**: 1 instance initially, 3 for HA cluster
- **Volume**: 1GB persistent (JetStream storage)
- **Memory**: 512MB per instance

**Clustering** (Future):
```
core-nats-1.internal:4222  ←→  core-nats-2.internal:4222
       ↕
core-nats-3.internal:4222
```

**Environment**:
```env
NATS_CLUSTER_NAME=core-v2-cluster
NATS_JETSTREAM_ENABLED=true
NATS_STORE_DIR=/app/.data/jetstream
```

---

#### 3. core-pocketbase (Database & API)
**Purpose**: Application database, authentication, and REST API

**Configuration**:
- **Exposed**: None (accessed via core-caddy proxy)
- **Internal**: `http://core-pocketbase.internal:8090`
- **Scaling**: 1 instance (SQLite single-writer)
- **Volume**: 5GB persistent (database + uploads)
- **Memory**: 1GB per instance

**Data**:
- SQLite database at `/app/.data/pb_data/`
- File uploads at `/app/.data/pb_public/`
- Backups via Litestream to R2

**Environment**:
```env
POCKETBASE_DATA_DIR=/app/.data/pb_data
NATS_URL=nats://core-nats.internal:4222
OBSERVABILITY_ENABLED=true
```

---

#### 4. core-controller (Orchestration)
**Purpose**: Central orchestration, observability aggregation, and control plane

**Configuration**:
- **Exposed**: None (internal control plane)
- **Internal**: `http://core-controller.internal:8080`
- **Scaling**: 1-2 instances (active/standby)
- **Volume**: 1GB persistent (state + logs)
- **Memory**: 512MB per instance

**Responsibilities**:
- Aggregate observability events from all services
- Health monitoring and alerting
- Configuration management
- Service discovery coordination

**Environment**:
```env
NATS_URL=nats://core-nats.internal:4222
POCKETBASE_URL=http://core-pocketbase.internal:8090
CONTROLLER_MODE=production
```

---

## Service Communication

### Network Architecture

All services communicate via **Fly.io Private Network (6PN)**:
- DNS: `<app-name>.internal` resolves to all instances
- Encrypted: Automatic WireGuard encryption
- Zero-trust: Services can only reach others in same org

### Event Flow

```
1. User Request
   └─> Caddy (HTTPS)
       └─> PocketBase (HTTP)
           └─> NATS (publish event)
               └─> Controller (consume + aggregate)

2. Internal Events
   └─> Any Service → NATS JetStream
       └─> Subscribers (ephemeral consumers)
           └─> Process events asynchronously
```

### Service Discovery

Each service discovers others via environment variables:
```bash
# Set in fly.toml for each service
NATS_URL=nats://core-nats.internal:4222
POCKETBASE_URL=http://core-pocketbase.internal:8090
CONTROLLER_URL=http://core-controller.internal:8080
```

---

## Deployment Strategy

### Phase 2 Migration Plan

**Step 1: Deploy Shared Infrastructure**
```bash
# Create NATS cluster
flyctl apps create core-nats --org personal
flyctl volumes create nats_data --region syd --size 1 --app core-nats
flyctl deploy --config fly-nats.toml --app core-nats

# Create PocketBase instance
flyctl apps create core-pocketbase --org personal
flyctl volumes create pb_data --region syd --size 5 --app core-pocketbase
flyctl deploy --config fly-pocketbase.toml --app core-pocketbase
```

**Step 2: Deploy Caddy Edge**
```bash
flyctl apps create core-caddy --org personal
flyctl deploy --config fly-caddy.toml --app core-caddy
```

**Step 3: Deploy Controller**
```bash
flyctl apps create core-controller --org personal
flyctl volumes create controller_data --region syd --size 1 --app core-controller
flyctl deploy --config fly-controller.toml --app core-controller
```

**Step 4: Configure DNS**
```bash
# Point domain to Caddy
flyctl certs create core-v2.fly.dev --app core-caddy
flyctl certs create nats.core-v2.fly.dev --app core-caddy
```

### Rollback Strategy

If Phase 2 fails, roll back to Phase 1:
```bash
# Scale down microservices
flyctl scale count 0 --app core-nats
flyctl scale count 0 --app core-pocketbase
flyctl scale count 0 --app core-caddy
flyctl scale count 0 --app core-controller

# Scale up monolithic
flyctl scale count 1 --app core-v2
```

---

## Configuration Files

Each service needs its own `fly-<service>.toml`:

### fly-nats.toml
```toml
app = 'core-nats'
primary_region = 'syd'

[env]
  NATS_CLUSTER_NAME = 'core-v2-cluster'
  NATS_JETSTREAM_ENABLED = 'true'
  NATS_STORE_DIR = '/app/.data/jetstream'

[[mounts]]
  source = 'nats_data'
  destination = '/app/.data'

[[services]]
  internal_port = 4222
  protocol = 'tcp'

  [[services.ports]]
    port = 4222

[[services]]
  internal_port = 8222
  protocol = 'tcp'

  [[services.ports]]
    port = 8222

[[vm]]
  memory = '512mb'
  cpu_kind = 'shared'
  cpus = 1
```

### fly-pocketbase.toml
```toml
app = 'core-pocketbase'
primary_region = 'syd'

[env]
  POCKETBASE_DATA_DIR = '/app/.data/pb_data'
  NATS_URL = 'nats://core-nats.internal:4222'
  OBSERVABILITY_ENABLED = 'true'

[[mounts]]
  source = 'pb_data'
  destination = '/app/.data'
  initial_size = '5gb'

[[services]]
  internal_port = 8090
  protocol = 'tcp'

  [[services.ports]]
    port = 8090

[[vm]]
  memory = '1gb'
  cpu_kind = 'shared'
  cpus = 1
```

### fly-caddy.toml
```toml
app = 'core-caddy'
primary_region = 'syd'

[env]
  POCKETBASE_URL = 'http://core-pocketbase.internal:8090'
  NATS_URL = 'http://core-nats.internal:8222'

[[services]]
  protocol = 'tcp'
  internal_port = 2015
  auto_stop_machines = 'stop'
  auto_start_machines = true
  min_machines_running = 1

  [[services.ports]]
    port = 80
    handlers = ['http']

  [[services.ports]]
    port = 443
    handlers = ['tls', 'http']

  [services.concurrency]
    type = 'connections'
    hard_limit = 25
    soft_limit = 20

[[vm]]
  memory = '256mb'
  cpu_kind = 'shared'
  cpus = 1
```

### fly-controller.toml
```toml
app = 'core-controller'
primary_region = 'syd'

[env]
  NATS_URL = 'nats://core-nats.internal:4222'
  POCKETBASE_URL = 'http://core-pocketbase.internal:8090'
  CONTROLLER_MODE = 'production'

[[mounts]]
  source = 'controller_data'
  destination = '/app/.data'

[[services]]
  internal_port = 8080
  protocol = 'tcp'

  [[services.ports]]
    port = 8080

[[vm]]
  memory = '512mb'
  cpu_kind = 'shared'
  cpus = 1
```

---

## Build Strategy

### Multi-Service Container Build

Each service needs its own container image:

**Option 1: Separate Binaries**
```bash
# Build separate binaries for each service
ko build --bare --platform=linux/amd64 ./cmd/nats --tags=core-nats
ko build --bare --platform=linux/amd64 ./cmd/pocketbase --tags=core-pocketbase
ko build --bare --platform=linux/amd64 ./cmd/caddy --tags=core-caddy
ko build --bare --platform=linux/amd64 ./cmd/controller --tags=core-controller
```

**Option 2: Single Binary with Subcommands** (Recommended)
```bash
# Build once, deploy to all services with different commands
ko build --bare --platform=linux/amd64 ./cmd/core

# Each fly.toml specifies which subcommand to run
[processes]
  nats = "/app/core nats serve"
  pocketbase = "/app/core pocketbase serve"
  caddy = "/app/core caddy serve"
  controller = "/app/core controller serve"
```

**Benefits of Single Binary**:
- ✅ Faster builds (one container for all services)
- ✅ Easier version management (same version everywhere)
- ✅ Shared code automatically in sync
- ✅ Simpler CI/CD pipeline

---

## Scaling Considerations

### Vertical Scaling (Per-Service)
```bash
# Scale memory/CPU for specific service
flyctl scale vm shared-cpu-2x --memory 2048 --app core-pocketbase
```

### Horizontal Scaling
```bash
# Add more instances
flyctl scale count 3 --app core-caddy       # Edge proxy
flyctl scale count 3 --app core-nats        # NATS cluster (requires config)
flyctl scale count 1 --app core-pocketbase  # SQLite = single writer
flyctl scale count 2 --app core-controller  # Active/standby
```

### Auto-Scaling Rules
```toml
# In fly.toml for services that can auto-scale
[services]
  auto_stop_machines = 'stop'
  auto_start_machines = true
  min_machines_running = 1

  [services.concurrency]
    type = 'connections'
    hard_limit = 100
    soft_limit = 80
```

---

## Monitoring & Observability

### Health Checks

Each service exposes health endpoints:
```
GET /health  → Basic liveness check
GET /ready   → Readiness check (dependencies healthy)
GET /metrics → Prometheus metrics
```

### Distributed Tracing

NATS events include trace context:
```json
{
  "trace_id": "abc123",
  "span_id": "xyz789",
  "service": "core-pocketbase",
  "event": "user.created",
  "timestamp": "2025-10-16T18:00:00Z"
}
```

### Centralized Logging

Controller aggregates logs from all services:
```bash
# View aggregated logs
flyctl logs --app core-controller

# View service-specific logs
flyctl logs --app core-nats
flyctl logs --app core-pocketbase
```

---

## Cost Analysis

### Phase 1 (Monolithic)
- 1 VM: 1GB RAM, shared CPU = **~$5-10/month**
- 1 Volume: 1GB = **~$0.15/month**
- **Total: ~$5-10/month**

### Phase 2 (Microservices)
- core-caddy: 256MB × 1 = **~$2/month**
- core-nats: 512MB × 1 = **~$4/month** (×3 for cluster = $12/month)
- core-pocketbase: 1GB × 1 = **~$6/month**
- core-controller: 512MB × 1 = **~$4/month**
- Volumes: 4 × 1-5GB = **~$1/month**
- **Total: ~$17/month** (single instances) or **~$25/month** (with NATS cluster)

**Cost Increase**: 2-3x for improved isolation, scaling, and reliability

---

## Security Improvements

### Phase 1 Security
- ✅ All services in one container
- ✅ Only Caddy exposed to internet
- ⚠️ Services share memory space (process isolation only)

### Phase 2 Security
- ✅ Network isolation (services in separate VMs)
- ✅ Private network encryption (WireGuard)
- ✅ Zero-trust between services
- ✅ Independent security updates per service
- ✅ Blast radius containment (compromise of one service doesn't affect others)

---

## Implementation Checklist

- [ ] Create fly.toml files for each service
- [ ] Update Caddyfile to proxy to internal services
- [ ] Add service discovery via DNS (*.internal)
- [ ] Implement health check endpoints for all services
- [ ] Test NATS communication across private network
- [ ] Configure persistent volumes for stateful services
- [ ] Set up monitoring and alerting
- [ ] Document deployment procedure
- [ ] Create rollback procedure
- [ ] Load test multi-service setup
- [ ] Implement auto-scaling rules
- [ ] Set up distributed tracing
- [ ] Configure log aggregation

---

## Next Steps

1. **Validate design** with stakeholders
2. **Create fly.toml files** for each service
3. **Test locally** with separate containers
4. **Deploy to staging** environment first
5. **Load test** and tune scaling parameters
6. **Deploy to production** with monitoring
7. **Document learnings** for future reference
