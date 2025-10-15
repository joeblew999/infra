# Environment Variables

## Architecture Principle: Same Binary, Different Environment

The core runtime follows a key principle: **The exact same binary runs locally and on Fly.io**. The only difference is HOW environment variables are provided to the process.

## How It Works

### Service Code Design

Services read environment variables directly using `os.Getenv()`:

```go
// services/pocketbase/bootstrap.go
func BootstrapAuth(app *pocketbase.PocketBase) error {
    appURL := os.Getenv("CORE_POCKETBASE_APP_URL")
    smtpHost := os.Getenv("CORE_POCKETBASE_SMTP_HOST")
    // ...
}
```

### Local Development

**Environment variables are inherited from the parent shell:**

```bash
# 1. Load .env into shell
export $(cat .env | grep -v '^#' | xargs)

# 2. Start the stack
go run ./cmd/core stack up
```

**Flow:**
```
User Shell (with exported env vars)
  └─ go run ./cmd/core stack up
      └─ GenerateComposeConfig()
          └─ service.json: ${env.*} placeholders are IGNORED
          └─ Writes .core-stack/process-compose.yaml
      └─ process-compose up
          └─ Starts cmd/pocketbase (inherits shell environment)
              └─ services/pocketbase.Run()
                  └─ os.Getenv("CORE_POCKETBASE_APP_URL") ✓
```

### Fly.io Deployment

**Environment variables come from two sources:**

1. **Non-sensitive config** in `fly.toml`:
```toml
[env]
  CORE_ENVIRONMENT = 'production'
  CORE_POCKETBASE_APP_URL = 'https://infra-pocketbase-iad.fly.dev'
```

2. **Sensitive secrets** via `flyctl secrets set`:
```bash
flyctl secrets set \
  CORE_POCKETBASE_ADMIN_EMAIL=admin@example.com \
  CORE_POCKETBASE_ADMIN_PASSWORD=<password> \
  --app infra-pocketbase-iad
```

**Flow:**
```
Fly.io Platform
  ├─ fly.toml [env] → injected as environment variables
  ├─ flyctl secrets   → injected as environment variables
  └─ Starts container with cmd/pocketbase
      └─ services/pocketbase.Run()
          └─ os.Getenv("CORE_POCKETBASE_APP_URL") ✓
```

## Service Manifest (service.json)

The `service.json` files declare environment needs using `${env.*}` placeholders:

```json
{
  "process": {
    "env": {
      "POCKETBASE_DIR": "${data}/pocketbase",
      "CORE_POCKETBASE_APP_URL": "${env.CORE_POCKETBASE_APP_URL}",
      "CORE_POCKETBASE_ADMIN_EMAIL": "${env.CORE_POCKETBASE_ADMIN_EMAIL}"
    }
  }
}
```

**Important:** These `${env.*}` placeholders are **documentation only**. They are NOT resolved during process-compose generation. Services must read them via `os.Getenv()`.

## Generated process-compose.yaml

The generated YAML only includes **resolved internal placeholders**:

```yaml
processes:
  pocketbase:
    command: .dep/pocketbase
    environment:
      - POCKETBASE_DIR=.data/pocketbase  # ${data} resolved
    # ${env.*} vars are NOT in YAML - inherited from parent
```

## Variable Categories

### Runtime Paths (Resolved at Generation)
- `${data}` → `.data/` directory
- `${dep.binary}` → Path to built binary
- `${bin}`, `${logs}` → Other runtime directories

These are resolved during `GenerateComposeConfig()` and baked into the YAML.

### Application Config (Inherited at Runtime)
- `${env.VARIABLE}` → Read via `os.Getenv("VARIABLE")`

These are NOT resolved during generation. Services read them from the process environment at runtime.

## Environment Variable Conventions

All environment variables follow these patterns:

### Runtime Configuration
- `CORE_ENVIRONMENT` - Environment name (development, production)
- `CORE_APP_ROOT` - Application root directory

### Service Configuration
- `CORE_SERVICENAME_*` - Service-specific config
  - Example: `CORE_POCKETBASE_APP_URL`
  - Example: `CORE_POCKETBASE_ADMIN_EMAIL`

### Service-Specific Variables
- Services may use their own conventions
  - PocketBase-HA: `PB_NAME`, `PB_REPLICATION_URL`
  - NATS: `NATS_PILLOW_HUB_AND_SPOKE`

## Local Development Setup

### Option 1: Manual Export (Recommended)

```bash
# Load .env
export $(cat .env | grep -v '^#' | xargs)

# Verify
env | grep CORE_POCKETBASE

# Start stack
go run ./cmd/core stack up
```

### Option 2: Shell Script

```bash
# Create helper script: scripts/dev.sh
#!/bin/bash
set -a
source .env
set +a
exec "$@"

# Use it
./scripts/dev.sh go run ./cmd/core stack up
```

### Option 3: direnv (Advanced)

```bash
# Install direnv
brew install direnv

# Add to ~/.zshrc or ~/.bashrc
eval "$(direnv hook zsh)"

# Create .envrc
cat > .envrc << 'EOF'
dotenv
EOF

# Allow
direnv allow

# Now .env is automatically loaded in this directory
```

## Fly.io Deployment

### Setting Non-Sensitive Config

Edit `services/*/fly.toml`:

```toml
[env]
  CORE_ENVIRONMENT = 'production'
  CORE_POCKETBASE_APP_URL = 'https://your-app.fly.dev'
```

### Setting Sensitive Secrets

```bash
flyctl secrets set \
  CORE_POCKETBASE_ADMIN_EMAIL=admin@example.com \
  CORE_POCKETBASE_ADMIN_PASSWORD=<password> \
  CORE_POCKETBASE_SMTP_PASSWORD=<smtp-password> \
  --app your-app-name
```

### Viewing Secrets

```bash
# List secret names (not values)
flyctl secrets list --app your-app-name

# SSH into machine and check environment
flyctl ssh console --app your-app-name
env | grep CORE_
```

## .env File Structure

**Local development .env example:**

```bash
# Core Runtime
CORE_ENVIRONMENT=development
CORE_APP_ROOT=.

# PocketBase Application
CORE_POCKETBASE_APP_URL=http://localhost:8090

# PocketBase Admin (for bootstrap)
CORE_POCKETBASE_ADMIN_EMAIL=admin@localhost
CORE_POCKETBASE_ADMIN_PASSWORD=changeme123

# PocketBase SMTP (optional)
CORE_POCKETBASE_SMTP_HOST=
CORE_POCKETBASE_SMTP_PORT=
CORE_POCKETBASE_SMTP_USERNAME=
CORE_POCKETBASE_SMTP_PASSWORD=
CORE_POCKETBASE_SMTP_FROM=
CORE_POCKETBASE_SMTP_TLS=

# OAuth2 Providers (optional)
CORE_POCKETBASE_GOOGLE_CLIENT_ID=
CORE_POCKETBASE_GOOGLE_CLIENT_SECRET=
```

**Important:**
- `.env` is in `.gitignore` - never commit it
- `.env.example` is tracked - documents required variables
- Empty values (`=`) are acceptable for optional config

## Troubleshooting

### Service fails with "missing configuration"

**Cause:** Required environment variable not set.

**Solution:**
```bash
# Check if variable is set
env | grep CORE_POCKETBASE_APP_URL

# If missing, load .env
export $(cat .env | grep -v '^#' | xargs)

# Restart service
go run ./cmd/core stack down
go run ./cmd/core stack up
```

### Environment changes not taking effect

**Cause:** Shell environment not updated.

**Solution:**
```bash
# Re-export .env
export $(cat .env | grep -v '^#' | xargs)

# Restart stack (no need to regenerate YAML)
go run ./cmd/core stack down
go run ./cmd/core stack up
```

### Fly deployment has wrong values

**Cause:** fly.toml or secrets out of date.

**Solution:**
```bash
# Update fly.toml [env] section
vim services/pocketbase/fly.toml

# OR update secrets
flyctl secrets set VARIABLE=new-value --app your-app

# Redeploy
flyctl deploy
```

## Design Rationale

### Why Not Resolve ${env.*} at Generation Time?

**Considered:** Resolving `${env.VARIABLE}` during `GenerateComposeConfig()` and baking values into YAML.

**Rejected because:**
1. ❌ Creates stale YAML when .env changes
2. ❌ Doesn't work for Fly secrets (not available at build time)
3. ❌ Breaks the principle: "same binary, different environment"
4. ❌ Requires regenerating YAML for env changes

### Why Not Use process-compose env_file?

**Considered:** Using process-compose's `env_file` directive to load .env.

**Rejected because:**
1. ❌ Doesn't exist on Fly (no .env in container)
2. ❌ Requires different config for local vs. production
3. ❌ Breaks parity principle

### Current Approach: Environment Inheritance ✓

**Why it works:**
1. ✅ Same binary runs locally and on Fly
2. ✅ Local: Inherit from shell
3. ✅ Fly: Inherit from platform (fly.toml + secrets)
4. ✅ No stale configuration
5. ✅ Simple and predictable

## Future Improvements

### Automatic .env Loading

Could add `godotenv.Load()` to CLI:

```go
// cmd/core/main.go
import "github.com/joho/godotenv"

func main() {
    // Load .env if it exists (don't fail if missing)
    _ = godotenv.Load()

    // ... rest of main
}
```

**Tradeoff:** Makes behavior implicit. Current explicit export is clearer.

### Environment Validation

Could add startup validation:

```go
func Run(ctx context.Context, args []string) error {
    required := []string{
        "CORE_POCKETBASE_APP_URL",
        "CORE_POCKETBASE_ADMIN_EMAIL",
    }
    for _, key := range required {
        if os.Getenv(key) == "" {
            return fmt.Errorf("required: %s", key)
        }
    }
    // ...
}
```

**Benefit:** Fail fast with clear error messages.

### .env Templates

Could add `go run ./cmd/core env init` to generate .env from template:

```bash
# Generate .env from .env.example
go run ./cmd/core env init

# Validate .env has required vars
go run ./cmd/core env validate
```

**Benefit:** Better developer onboarding.
