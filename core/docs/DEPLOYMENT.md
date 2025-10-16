# Deployment Guide

This guide covers deploying core V2 to Fly.io, both as a monolithic container (Phase 1) and as microservices (Phase 2).

## Deployment Tools

The infra repository includes a comprehensive tooling system for deployment automation. There are two ways to deploy:

### Option 1: Tooling System (Recommended)

The **tooling binary** (`./core/tooling`) provides a complete deployment workflow:
- Handles Fly.io and Cloudflare authentication automatically
- Manages container building with ko
- Deploys to Fly.io with full orchestration
- Stores credentials securely in `.data/core/`

**Quick start:**
```bash
cd core
./tooling workflow deploy --app core-v2 --verbose
```

The tooling will:
1. Prompt for Fly.io authentication (opens browser)
2. Prompt for Cloudflare authentication (if DNS management needed)
3. Build container image with ko
4. Deploy to Fly.io with your configuration

### Option 2: Core CLI Deploy Command

The **core CLI** (`go run . deploy`) provides a simpler, manual deployment:
- Requires pre-installed ko and flyctl tools
- Uses environment variable for Fly.io token
- No Cloudflare integration
- Good for CI/CD pipelines or manual deploys

**Quick start:**
```bash
# Install tools first
go install github.com/google/ko@latest
curl -L https://fly.io/install.sh | sh

# Deploy
FLY_API_TOKEN=<token> go run . deploy --app core-v2
```

## Prerequisites

### For Tooling System (Option 1)

No additional tools needed! The tooling binary includes everything:
- ✅ ko builder integrated
- ✅ Fly.io API client built-in
- ✅ Cloudflare DNS management (optional)
- ✅ Credential storage in `.data/core/`

**Requirements:**
1. Fly.io account: https://fly.io/app/sign-up
2. (Optional) Cloudflare account for DNS: https://www.cloudflare.com/

### For Core CLI (Option 2)

You need to install deployment tools manually:

1. **ko** - Container image builder for Go applications
   - Installation: `go install github.com/google/ko@latest`
   - Or download from: https://github.com/google/ko/releases
   - Place in `.dep/ko` or ensure it's in your PATH

2. **flyctl** - Fly.io CLI tool
   - Installation: `curl -L https://fly.io/install.sh | sh`
   - Or download from: https://fly.io/docs/hands-on/install-flyctl/
   - Place in `.dep/flyctl` or ensure it's in your PATH

3. **Fly.io API token**
   - Generate at: https://fly.io/user/personal_access_tokens
   - Set environment variable: `export FLY_API_TOKEN=<your_token>`

## Phase 1: Monolithic Deployment

Deploy the entire core stack as a single container.

### What Gets Deployed

The monolithic deployment includes all services in one container:
- **Process-compose** - Orchestrates all services
- **NATS JetStream** - Event streaming
- **PocketBase** - Database and admin API
- **Caddy** - Reverse proxy and TLS termination

All services communicate via localhost within the container, with only Caddy exposed publicly.

### Deployment Configuration

The deployment is configured in `fly.toml`:
- **App name**: `core-v2` (customize with `--app` flag)
- **Region**: `syd` Sydney (customize with `--region` flag)
- **Resources**: 1GB RAM, shared CPU
- **Persistent volume**: 1GB at `/app/.data`
- **Exposed ports**: 80 (HTTP) → 443 (HTTPS)

### Deploy Commands

#### Using Tooling System (Recommended)

```bash
# Interactive deployment with authentication prompts
./tooling workflow deploy --app core-v2 --verbose

# Custom region
./tooling workflow deploy --app core-v2 --region lax

# Non-interactive (for CI/CD)
./tooling workflow deploy --app core-v2 --no-browser --json
```

The tooling system will:
- Open browser for Fly.io authentication
- Open browser for Cloudflare authentication (optional)
- Build and push container automatically
- Deploy to Fly.io with full orchestration

#### Using Core CLI

```bash
# Dry run (test without deploying)
go run . deploy --dry-run

# Deploy to production
FLY_API_TOKEN=<token> go run . deploy

# Custom app name and region
FLY_API_TOKEN=<token> go run . deploy --app my-core-app --region lax

# Deploy to staging
FLY_API_TOKEN=<token> go run . deploy --env staging --app core-staging
```

### Deployment Process

When you run `go run . deploy`, it performs these steps:

1. **Check tools** - Verifies ko and flyctl are available
2. **Build container** - Uses ko to build multi-platform image (amd64/arm64)
3. **Push to registry** - Pushes to `registry.fly.io/<app-name>`
4. **Deploy** - Uses flyctl to deploy with fly.toml configuration

### First-Time Setup

If deploying to a new app for the first time:

```bash
# Create the Fly.io app
flyctl apps create core-v2 --org personal

# Create the persistent volume
flyctl volumes create core_data --region syd --size 1

# Deploy
FLY_API_TOKEN=<token> go run . deploy
```

### Monitoring Deployment

After deployment completes, monitor your app:

```bash
# Check app status
flyctl status --app core-v2

# View logs
flyctl logs --app core-v2

# Access the app
open https://core-v2.fly.dev
```

### Troubleshooting

#### Error: "ko build tool not found"
Install ko:
```bash
go install github.com/google/ko@latest
```

Or place the ko binary in `.dep/ko`.

#### Error: "flyctl not found"
Install flyctl:
```bash
curl -L https://fly.io/install.sh | sh
```

Or place the flyctl binary in `.dep/flyctl`.

#### Error: "FLY_API_TOKEN environment variable is required"
Set your Fly.io API token:
```bash
export FLY_API_TOKEN=<your_token>
```

#### Error: "app not found"
Create the app first:
```bash
flyctl apps create <app-name>
```

#### Deployment Fails During Health Checks
Check the logs to see which service is failing:
```bash
flyctl logs --app core-v2
```

Common issues:
- **Volume not mounted**: Create the volume with `flyctl volumes create`
- **Insufficient memory**: Increase VM memory in fly.toml
- **Port conflicts**: Ensure services use configured ports from environment

## Phase 2: Microservices Deployment (Future)

Phase 2 will deploy each service as a separate Fly.io app:
- `core-nats` - NATS JetStream
- `core-pocketbase` - PocketBase database
- `core-caddy` - Caddy proxy
- `core-controller` - Orchestration controller

Each service will:
- Scale independently
- Have its own persistent volume
- Communicate via Fly.io private network
- Use Fly.io DNS for service discovery

Documentation for Phase 2 will be added when implementation begins.

## Deployment Architecture

### Monolithic (Phase 1)

```
                ┌─────────────────────────────────┐
                │   Fly.io VM (core-v2)           │
Internet ───────┤   ┌─────────────────────────┐   │
 (HTTPS)        │   │  Process-Compose        │   │
                │   │  ┌────┬──────┬───────┐  │   │
                │   │  │NATS│PB    │Caddy  │  │   │
                │   │  │:4222:8090 │:2015  │  │   │
                │   │  └────┴──────┴───────┘  │   │
                │   └─────────────────────────┘   │
                │                                 │
                │   Volume: /app/.data            │
                └─────────────────────────────────┘
```

### Microservices (Phase 2 - Future)

```
                ┌─────────────┐
                │ core-caddy  │
Internet ───────┤ (Proxy)     │
 (HTTPS)        └──────┬──────┘
                       │ Private Network
          ┌────────────┼────────────┐
          │            │            │
     ┌────▼────┐  ┌───▼─────┐ ┌───▼──────┐
     │core-nats│  │core-pb  │ │core-ctrl │
     │(Events) │  │(DB)     │ │(Orchestr)│
     └─────────┘  └─────────┘ └──────────┘
         Each with own volume and scaling
```

## Security Considerations

### Secrets Management

Do NOT commit secrets to git. Use environment variables:

```bash
# In production (Fly.io secrets)
flyctl secrets set DATABASE_PASSWORD=<password> --app core-v2
flyctl secrets set ADMIN_PASSWORD=<password> --app core-v2

# In development (local .env)
cp .env.example .env
# Edit .env with your local secrets
```

### TLS Certificates

Fly.io automatically provisions and renews TLS certificates for your apps.
- Certificates are issued by Let's Encrypt
- Automatic renewal 30 days before expiry
- Custom domains supported with DNS verification

### Network Security

Monolithic deployment (Phase 1):
- Only Caddy (port 2015) is exposed to internet
- Internal services (NATS, PocketBase) only accessible within container
- All external traffic goes through Caddy reverse proxy

Microservices deployment (Phase 2):
- Services communicate via Fly.io private network (6PN)
- Only Caddy proxy exposed to internet
- Internal service-to-service communication encrypted

## Cost Estimates

### Monolithic Deployment (Phase 1)

Based on Fly.io pricing (as of 2024):
- **Compute**: ~$5-10/month (shared CPU, 1GB RAM)
- **Storage**: ~$0.15/GB/month (1GB volume)
- **Bandwidth**: First 100GB free, then $0.02/GB

**Total estimated cost**: ~$5-10/month for light usage

### Microservices Deployment (Phase 2)

- **Per service**: ~$5/month (4 services = $20/month)
- **Storage**: ~$0.60/month (4x 1GB volumes)
- **Benefits**: Independent scaling, better isolation

**Total estimated cost**: ~$20-30/month

## Rollback Strategy

If a deployment fails or causes issues:

```bash
# List deployment history
flyctl releases --app core-v2

# Rollback to previous version
flyctl releases rollback <version> --app core-v2
```

## Continuous Deployment

The CI/CD workflow (`.github/workflows/core-ci.yml`) can be extended to automatically deploy on merge to main:

```yaml
# Add to .github/workflows/core-ci.yml
- name: Deploy to Fly.io
  if: github.ref == 'refs/heads/main'
  env:
    FLY_API_TOKEN: ${{ secrets.FLY_API_TOKEN }}
  run: |
    cd core
    go run . deploy --app core-v2-prod
```

## Next Steps

1. **Install required tools** (ko, flyctl)
2. **Set up Fly.io account** and generate API token
3. **Test with dry-run**: `go run . deploy --dry-run`
4. **Deploy to staging**: `go run . deploy --env staging --app core-staging`
5. **Monitor deployment**: `flyctl logs --app core-staging`
6. **Deploy to production**: `go run . deploy --app core-v2`

For questions or issues, see:
- [TROUBLESHOOTING.md](TROUBLESHOOTING.md) - Common issues and solutions
- [ARCHITECTURE.md](ARCHITECTURE.md) - System design and architecture
- [DEVELOPMENT.md](DEVELOPMENT.md) - Development workflow
