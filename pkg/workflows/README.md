# Workflows

This package provides standardized, idempotent workflows for building and deploying applications.

## Available Workflows

### 1. Binary Build (`binary`)

Cross-platform binary compilation with standardized naming and output conventions.

**Usage:**
```bash
# Build for local platform only
go run . binary --local

# Build for all platforms and architectures
go run . binary --all

# Build specific platform/arch
go run . binary --platforms linux --arch amd64

# Custom binary name and output directory
go run . binary --name myapp --output ./build --all
```

**Features:**
- **Output Directory:** Uses `.bin` folder from `pkg/config.GetBinPath()`
- **Naming Convention:** `BINARYNAME_OS_ARCH` format
- **Platforms:** Windows, Darwin (macOS), Linux
- **Architectures:** arm64, amd64
- **Static linking** with `CGO_ENABLED=0` for portability
- **Optimized binaries** with `-ldflags="-s -w -trimpath"`
- **Proper file extensions:** `.exe` for Windows

**Built Binaries:**
```
.bin/
├── infra_darwin_amd64        # macOS Intel
├── infra_darwin_arm64        # macOS Apple Silicon
├── infra_linux_amd64         # Linux x86-64
├── infra_linux_arm64         # Linux ARM64
├── infra_windows_amd64.exe   # Windows x86-64
└── infra_windows_arm64.exe   # Windows ARM64
```

### 2. Container Build (`build`)

Container image builds using `ko` with standardized settings.

**Usage:**
```bash
# Build container image
go run . build

# Build and push to registry
go run . build --push --repo registry.fly.io/myapp

# Build for specific platform
go run . build --platform linux/arm64
```

**Features:**
- **Optimized for size and security**
- **Multi-platform support**
- **Consistent tagging and metadata**
- **Automatic registry authentication**
- **Production/development environment detection**

### 2a. Multi-Registry Build (`multiregistry`)

Build and push container images to multiple registries for redundancy and reliability.

**Usage:**
```bash
# Build and push to GitHub Container Registry (recommended)
go run . multiregistry --ghcr

# Build and push to both GHCR and Fly.io registry
go run . multiregistry --ghcr --fly

# Dry run to see what would happen
go run . multiregistry --ghcr --dry-run

# Build with custom app name and environment
go run . multiregistry --ghcr --app my-app --environment production
```

**Features:**
- **Primary Registry:** GitHub Container Registry (GHCR) - better reliability
- **Secondary Registry:** Fly.io registry - integrated with deployment
- **Secure Credential Management** - uses environment variables
- **Git Hash Injection** - automatic build metadata
- **Robust Error Handling** - continues on partial failures
- **Dry Run Support** - test without pushing

### 3. Container Deploy (`deploy`)

Idempotent deployment to Fly.io using container images.

**Usage:**
```bash
# Deploy to Fly.io
go run . deploy

# Deploy with custom app name
go run . deploy --app my-app --region syd

# Dry run to see what would happen
go run . deploy --dry-run
```

**Features:**
- **Idempotent** - safe to run multiple times
- **Automatic prerequisites check**
- **App and volume creation** if needed
- **Multi-registry container build** with GHCR primary
- **GHCR-based deployment** for better reliability
- **Deployment verification**

### 4. Status Check (`status`)

Check deployment status and health.

**Usage:**
```bash
# Check app status
go run . status --app my-app

# Check with logs
go run . status --app my-app --logs 100 --verbose
```

### 5. Project Init (`init`)

Initialize new projects with standard configuration.

**Usage:**
```bash
# Initialize new project
go run . init --name my-project --template web

# Force overwrite existing files
go run . init --name my-project --force
```

## Architecture

All workflows follow these principles:

- **Idempotent** - can be run multiple times safely
- **Standardized** - consistent behavior across environments
- **Configurable** - flexible via command-line flags
- **Observable** - comprehensive logging and status reporting
- **Modular** - can be used independently or together

## File Structure

```
pkg/workflows/
├── binary.go        # Cross-platform binary builds
├── container.go     # Container image builds with ko
├── multiregistry.go # Multi-registry container builds
├── deploy.go        # Fly.io deployment workflow
├── status.go        # Deployment status and health checks
└── init.go          # Project initialization
```