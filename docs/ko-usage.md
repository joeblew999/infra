# Ko Container Builds

This document covers using ko to build and deploy container images for the infrastructure management system.

## Required Configuration

### Environment Variables

**Essential:**
```bash
# Docker registry for pushing images
export KO_DOCKER_REPO=registry.fly.io/your-app-name

# Or for local development
export KO_DOCKER_REPO=ko.local
```

**Optional:**
```bash
# Override default platforms
export KO_DEFAULTPLATFORMS=linux/amd64,linux/arm64

# Override base image
export KO_DEFAULTBASEIMAGE=cgr.dev/chainguard/static:latest

# Environment detection
export ENVIRONMENT=production  # or development
```

## Common Usage

### Local Development
```bash
# Build for local Docker daemon
KO_DOCKER_REPO=ko.local go run . ko build --local github.com/joeblew999/infra

# Build to tarball (no Docker needed)
KO_DOCKER_REPO=test go run . ko build --tarball=infra.tar github.com/joeblew999/infra

# Build for specific platform
go run . ko build --platform=linux/arm64 --tarball=infra-arm64.tar github.com/joeblew999/infra
```

### Production Deployment
```bash
# Build and push to Fly.io registry
KO_DOCKER_REPO=registry.fly.io/your-app ENVIRONMENT=production go run . ko build github.com/joeblew999/infra

# Build multi-platform
KO_DOCKER_REPO=registry.fly.io/your-app go run . ko build --platform=linux/amd64,linux/arm64 github.com/joeblew999/infra
```

### CI/CD Pipeline
```bash
# In GitHub Actions or similar
export KO_DOCKER_REPO=registry.fly.io/${FLY_APP_NAME}
export ENVIRONMENT=production

# Build and push
go run . ko build github.com/joeblew999/infra
```

## Registry Setup

### Fly.io Registry
1. Create Fly.io app: `flyctl apps create your-app-name`
2. Login to registry: `go run . ko login registry.fly.io`
3. Set repository: `export KO_DOCKER_REPO=registry.fly.io/your-app-name`

### Other Registries
```bash
# Docker Hub
export KO_DOCKER_REPO=docker.io/username/repo

# GitHub Container Registry
export KO_DOCKER_REPO=ghcr.io/username/repo

# Google Container Registry
export KO_DOCKER_REPO=gcr.io/project-id/repo
```

## Configuration Files

- `.ko.yaml` - Build configuration (checked into git)
- Uses environment variables for runtime settings
- Automatically detects production vs development environment

## Troubleshooting

**"KO_DOCKER_REPO environment variable is unset"**
```bash
export KO_DOCKER_REPO=ko.local  # for local builds
```

**"Cannot connect to the Docker daemon"**
```bash
# Use tarball instead
go run . ko build --tarball=image.tar github.com/joeblew999/infra
```

**"importpath is not package main"**
```bash
# Use full import path
go run . ko build github.com/joeblew999/infra
```