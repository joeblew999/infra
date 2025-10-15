# Infrastructure Architecture

## Overview

This infrastructure provides a complete stack for building distributed applications with high availability, leveraging NATS messaging and PocketBase for data management.

## Core Components

### 1. NATS Cluster (Pillow-Managed)

**Location**: `services/nats/`, `pkg/nats/pillow.go`

**Purpose**: Provides distributed messaging infrastructure for service communication and data replication.

**Key Features**:
- **Pillow Integration**: Uses [Pillow](https://github.com/Nintron27/pillow) for simplified NATS embedding
- **Fly.io Aware**: Automatic multi-region clustering via `FlyioClustering` or `FlyioHubAndSpoke` adapters
- **NSC Authentication**: JWT-based authentication using NATS Security (NSC)
- **JetStream Enabled**: Persistent streams for reliable messaging and data sync

**Topology Options**:
- `FlyioClustering`: Full mesh topology across regions
- `FlyioHubAndSpoke`: Hub region with leaf nodes in other regions

**Configuration**:
```bash
# Enable hub-and-spoke instead of mesh
NATS_PILLOW_HUB_AND_SPOKE=true

# NATS will auto-discover Fly.io machines and cluster them
```

**Authentication**:
- Operator/Account/User JWTs generated via NSC
- System account for cluster operations
- Application account for service communication
- Credentials stored in `.data/nats-auth/`

### 2. PocketBase Service

**Location**: `services/pocketbase/`

**Purpose**: Embeds PocketBase with complete Datastar-based authentication UI.

**Key Features**:
- **28 API Endpoints**: Complete auth workflow (login, signup, OAuth2, OTP, password reset, account management)
- **Datastar Integration**: Reactive UI with Server-Sent Events (SSE)
- **OAuth2 Providers**: Google, GitHub, Microsoft, Apple
- **Bootstrap Configuration**: Auto-creates superuser from environment variables
- **Multi-Provider Auth**: Password, OAuth2, OTP, MFA

**Templates** (7 HTML files):
- `auth_index.html` - Login page
- `auth_signup.html` - Registration
- `auth_verify_email.html` - Email verification
- `auth_reset_request.html` - Password reset request
- `auth_reset_confirm.html` - Password reset confirmation
- `auth_settings.html` - Account management
- `auth_callback.html` - OAuth2 callback handler

**Configuration**:
```bash
# PocketBase settings
CORE_POCKETBASE_APP_URL=http://localhost:8090
CORE_POCKETBASE_ADMIN_EMAIL=admin@localhost
CORE_POCKETBASE_ADMIN_PASSWORD=changeme123

# OAuth2 providers (optional)
CORE_POCKETBASE_GOOGLE_CLIENT_ID=...
CORE_POCKETBASE_GOOGLE_CLIENT_SECRET=...
```

### 3. PocketBase-HA (High Availability)

**Location**: `services/pocketbase-ha/` (foundation laid)

**Purpose**: Provides distributed PocketBase with automatic replication across nodes.

**Architecture**:
```
┌─────────────────────────────────────────┐
│     PocketBase-HA Nodes (3+)           │
│  ┌──────┐  ┌──────┐  ┌──────┐         │
│  │ Node1│  │ Node2│  │ Node3│         │
│  │ SQLite│  │ SQLite│  │ SQLite│       │
│  └───┬──┘  └───┬──┘  └───┬──┘         │
│      └─────────┼─────────┘              │
│                ↓                         │
│     Pillow-Managed NATS Cluster        │
│       (Multi-Region Mesh)               │
└─────────────────────────────────────────┘
```

**Technology Stack**:
- [go-ha](https://github.com/litesql/go-ha): SQLite HA driver using NATS for replication
- **Leaderless**: All nodes can handle reads and writes
- **Conflict Resolution**: Configurable strategies via go-ha
- **Change Data Capture (CDC)**: Automatic replication of changes

**Integration**:
- Reuses Datastar auth templates from regular PocketBase
- Reuses bootstrap configuration
- Connects to Pillow-managed NATS cluster
- **CGO Required**: Must build with `CGO_ENABLED=1`

**Configuration**:
```bash
# Node identification
PB_NAME=pocketbase-node-1

# Connect to our Pillow NATS
PB_REPLICATION_URL=nats://localhost:4222

# Stream for PocketBase data
PB_REPLICATION_STREAM=pb
```

**Deployment Strategy**:
- Development: Single PocketBase instance
- Production: 3+ PocketBase-HA nodes across regions
- Each node connects to regional NATS via Pillow

## Architecture Decisions

### Why Pillow for NATS?

**Chosen**: Pillow + external NATS cluster

**Rationale**:
1. **Fly.io Optimization**: Pillow understands Fly.io regions and automatically configures clustering
2. **Separation of Concerns**: NATS infrastructure separate from application data
3. **Reusability**: One NATS cluster serves multiple purposes:
   - PocketBase replication
   - General pub/sub messaging
   - Log streaming
   - Future services
4. **Flexibility**: Can upgrade/replace components independently

**Alternative Considered**: litesql/ha with embedded NATS
- ❌ Less flexible (NATS only for SQLite sync)
- ❌ No Fly.io topology awareness
- ❌ Can't reuse NATS for other services

### Why go-ha over pocketbase-ha binary?

**Chosen**: Integrate go-ha driver into our own PocketBase service

**Rationale**:
1. **Control**: We control the PocketBase instance and can reuse our Datastar auth
2. **Flexibility**: Can customize bootstrap, routing, and features
3. **Library vs Binary**: go-ha is importable, pocketbase-ha is a standalone binary
4. **Maintainability**: Single service.go with HA toggle instead of duplicate services

## Deployment Topology

### Local Development
```
┌──────────────┐
│ PocketBase   │ (single node)
└──────┬───────┘
       ↓
┌──────────────┐
│ NATS         │ (single node, no Pillow)
└──────────────┘
```

### Production (Fly.io)
```
Region: iad (hub)              Region: lhr (leaf)          Region: nrt (leaf)
┌──────────────┐              ┌──────────────┐            ┌──────────────┐
│ PB-HA Node 1 │              │ PB-HA Node 4 │            │ PB-HA Node 6 │
│ PB-HA Node 2 │              │ PB-HA Node 5 │            │ PB-HA Node 7 │
│ PB-HA Node 3 │              └──────┬───────┘            └──────┬───────┘
└──────┬───────┘                     │                            │
       │                             │                            │
┌──────▼────────────────────────────▼────────────────────────────▼─────┐
│              Pillow NATS Cluster (Hub + Spoke Topology)              │
│   Hub: iad (3 nodes)   Leaf: lhr (1 node)   Leaf: nrt (1 node)     │
└──────────────────────────────────────────────────────────────────────┘
```

## File Structure

```
core/
├── services/
│   ├── nats/                    # NATS service with Pillow
│   │   ├── service.go          # Embedded NATS with NSC auth
│   │   └── service.json        # Manifest with Pillow config
│   ├── pocketbase/             # Single-node PocketBase
│   │   ├── auth.go             # 28 Datastar auth endpoints
│   │   ├── bootstrap.go        # Auto-config (exported)
│   │   ├── service.go          # Embedded PocketBase runner
│   │   ├── service.json        # Manifest
│   │   └── auth_*.html         # 7 Datastar UI templates
│   └── pocketbase-ha/          # HA PocketBase (foundation)
│       ├── service.go          # go-ha integrated service
│       ├── service.json        # HA manifest
│       └── auth_*.html         # Symlinked templates
├── pkg/
│   ├── nats/
│   │   ├── pillow.go           # Pillow NATS clustering
│   │   └── auth/               # NSC authentication
│   └── config/
│       ├── nats.go             # NATS configuration
│       └── pocketbase.go       # PocketBase configuration
└── cmd/
    ├── nats/main.go            # NATS CLI
    ├── pocketbase/main.go      # PocketBase CLI
    └── pocketbase-ha/main.go   # PocketBase-HA CLI
```

## Environment Variables

### NATS
- `NATS_PILLOW_HUB_AND_SPOKE`: Use hub-and-spoke topology (default: false = mesh)

### PocketBase
- `CORE_POCKETBASE_APP_URL`: Application URL
- `CORE_POCKETBASE_ADMIN_EMAIL`: Bootstrap admin email
- `CORE_POCKETBASE_ADMIN_PASSWORD`: Bootstrap admin password
- `CORE_POCKETBASE_*_CLIENT_ID`: OAuth2 provider IDs
- `CORE_POCKETBASE_*_CLIENT_SECRET`: OAuth2 provider secrets

### PocketBase-HA
- `PB_NAME`: Unique node name
- `PB_REPLICATION_URL`: NATS connection URL (default: from config.GetNATSURL())
- `PB_REPLICATION_STREAM`: NATS stream name (default: "pb")

## Build Requirements

### Standard Services (NATS, PocketBase)
```bash
go build ./cmd/nats
go build ./cmd/pocketbase
```

### PocketBase-HA (requires CGO)
```bash
CGO_ENABLED=1 go build ./cmd/pocketbase-ha
```

## Next Steps

1. **Complete PocketBase-HA Integration**: Wire go-ha driver into PocketBase service
2. **Testing**: Multi-node testing with conflict scenarios
3. **Monitoring**: Add metrics for replication lag and conflict resolution
4. **Documentation**: Deployment guides for Fly.io
5. **Fly.io Deployment**: Create fly.toml configurations for each region

## References

- [Pillow](https://github.com/Nintron27/pillow) - NATS embedding library
- [go-ha](https://github.com/litesql/go-ha) - SQLite HA driver
- [PocketBase](https://pocketbase.io/) - Backend-as-a-Service
- [Datastar](https://data-star.dev/) - Reactive UI framework
- [NATS](https://nats.io/) - Distributed messaging system
