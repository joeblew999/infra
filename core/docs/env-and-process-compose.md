# Environment Variables and Process Compose Generation

This document explains how environment variables from `.env` files flow through the system and into the generated `process-compose.yaml` configuration.

## Overview

The core runtime uses a **service manifest pattern** where each service declares its configuration needs in a `service.json` file. These manifests use placeholder syntax to reference:

1. **Binary paths** (`${dep.binary_name}`)
2. **Runtime directories** (`${data}`, `${bin}`, `${dep}`, `${logs}`)
3. **Environment variables** (`${env.VARIABLE_NAME}`)

During process-compose generation, these placeholders are resolved and the resulting configuration is written to `.core-stack/process-compose.yaml`.

## The Flow

### 1. Environment Variables Source

Environment variables are typically defined in `.env` file at the repository root:

```bash
# .env
CORE_ENVIRONMENT=development
CORE_POCKETBASE_APP_URL=http://localhost:8090
CORE_POCKETBASE_ADMIN_EMAIL=admin@localhost
CORE_POCKETBASE_ADMIN_PASSWORD=changeme123
# ... more variables
```

These variables are loaded into the environment by:
- The shell (when manually sourcing the file)
- process-compose (when it starts)
- The Go application (via `os.Getenv()`)

**Note**: The codebase has `github.com/joho/godotenv` in go.mod as an indirect dependency, but currently does NOT explicitly call `godotenv.Load()`. Environment variables must be loaded externally (e.g., by the shell or process manager).

### 2. Service Manifest Declaration

Services declare their environment needs in `service.json`. Example from [services/pocketbase/service.json](../services/pocketbase/service.json):

```json
{
  "process": {
    "command": "${dep.pocketbase}",
    "env": {
      "POCKETBASE_DIR": "${data}/pocketbase",
      "CORE_POCKETBASE_APP_URL": "${env.CORE_POCKETBASE_APP_URL}",
      "CORE_POCKETBASE_ADMIN_EMAIL": "${env.CORE_POCKETBASE_ADMIN_EMAIL}",
      "CORE_POCKETBASE_ADMIN_PASSWORD": "${env.CORE_POCKETBASE_ADMIN_PASSWORD}"
    }
  }
}
```

**Placeholder Types**:
- `${dep.pocketbase}` - Path to the pocketbase binary (resolved from dependency manifest)
- `${data}` - Runtime data directory (e.g., `.data/`)
- `${env.VARIABLE_NAME}` - Environment variable value from `os.Getenv("VARIABLE_NAME")`

### 3. Placeholder Resolution

When `GenerateComposeConfig()` is called in [pkg/runtime/process/processcompose.go](../pkg/runtime/process/processcompose.go#L37-L61), it:

1. **Loads service specs** from each service's embedded `service.json`
2. **Ensures binaries** are built or available
3. **Calls `spec.ResolveEnv(paths)`** to resolve placeholders

The `ResolveEnv()` function in each service (e.g., [services/pocketbase/service.go:86-103](../services/pocketbase/service.go#L86-L103)) performs three types of substitution:

```go
func (s *Spec) ResolveEnv(paths map[string]string) map[string]string {
    result := make(map[string]string, len(s.Process.Env))
    runtime := runtimecfg.Load()
    for key, value := range s.Process.Env {
        resolved := value
        // 1. Replace ${dep.*} with binary paths
        for name, path := range paths {
            placeholder := fmt.Sprintf("${dep.%s}", name)
            resolved = strings.ReplaceAll(resolved, placeholder, path)
        }
        // 2. Replace ${data} with runtime data directory
        resolved = strings.ReplaceAll(resolved, "${data}", runtime.Paths.Data)
        // 3. Replace ${env.*} with environment variables
        resolved = replaceEnvPlaceholders(resolved)
        result[key] = resolved
    }
    return result
}
```

The `replaceEnvPlaceholders()` helper extracts the variable name from `${env.VARIABLE_NAME}` and calls `os.Getenv("VARIABLE_NAME")`:

```go
func replaceEnvPlaceholders(value string) string {
    for {
        start := strings.Index(value, "${env.")
        if start == -1 {
            break
        }
        end := strings.Index(value[start:], "}")
        if end == -1 {
            break
        }
        end += start
        placeholder := value[start : end+1]
        envVar := value[start+len("${env.") : end]
        replacement := os.Getenv(envVar)
        value = strings.ReplaceAll(value, placeholder, replacement)
    }
    return value
}
```

### 4. Process Compose Generation

The resolved environment map is converted to a slice of `KEY=VALUE` strings and written to `.core-stack/process-compose.yaml`:

```yaml
processes:
  pocketbase:
    command: .dep/pocketbase
    environment:
      - POCKETBASE_DIR=.data/pocketbase
      - CORE_POCKETBASE_APP_URL=http://localhost:8090
      - CORE_POCKETBASE_ADMIN_EMAIL=admin@localhost
      - CORE_POCKETBASE_ADMIN_PASSWORD=changeme123
```

**Note**: The generated file contains **resolved values**, not placeholders. If `.env` is updated, you must regenerate `process-compose.yaml` for changes to take effect.

### 5. Service Startup

When process-compose starts a service, it:

1. Sets the environment variables from the `environment:` section
2. Executes the command (e.g., `.dep/pocketbase`)

The service code can then read these variables using `os.Getenv()`. For example, [services/pocketbase/bootstrap.go:39-40](../services/pocketbase/bootstrap.go#L39-L40):

```go
appURL := os.Getenv("CORE_POCKETBASE_APP_URL")
smtpHost := os.Getenv("CORE_POCKETBASE_SMTP_HOST")
```

## Important Notes

### 1. Generation is Required

The `process-compose.yaml` file is **GENERATED**, not manually edited. If you:
- Update `.env` values
- Modify `service.json` manifests
- Change service dependencies

You **must regenerate** the compose file:

```bash
# Regenerate by running the orchestrator
go run ./cmd/core stack start
```

### 2. Missing Environment Variables

If a `${env.VARIABLE_NAME}` placeholder references a variable that doesn't exist in the environment, `os.Getenv()` returns an empty string. This means:

```json
{
  "CORE_POCKETBASE_SMTP_HOST": "${env.CORE_POCKETBASE_SMTP_HOST}"
}
```

Becomes:

```yaml
- CORE_POCKETBASE_SMTP_HOST=
```

If the variable is optional (like SMTP config), this is fine. If it's required, the service may fail to start or behave incorrectly.

### 3. No Automatic .env Loading

Currently, the Go codebase does **not** automatically load `.env` files using `godotenv.Load()`. Environment variables must be:

- Exported in your shell: `export CORE_POCKETBASE_APP_URL=http://localhost:8090`
- Loaded by process-compose (which inherits the parent process environment)
- Set when running commands: `CORE_ENVIRONMENT=production go run ./cmd/core`

If you want automatic loading, add this to the main() function:

```go
import "github.com/joho/godotenv"

func main() {
    // Load .env file if it exists (optional, doesn't error if missing)
    _ = godotenv.Load()

    // ... rest of main
}
```

### 4. Environment Variable Naming Convention

All environment variables follow the pattern:

- **Runtime config**: `CORE_*` (e.g., `CORE_ENVIRONMENT`, `CORE_APP_ROOT`)
- **Service config**: `CORE_SERVICENAME_*` (e.g., `CORE_POCKETBASE_APP_URL`)
- **Service-specific**: Service code may use its own conventions (e.g., `PB_NAME` for pocketbase-ha)

This ensures clear namespacing and avoids conflicts with system or third-party variables.

## Services Using This Pattern

All services in the `services/` directory use this pattern:

1. **NATS** ([services/nats/service.json](../services/nats/service.json))
   - Uses `${env.*}` for optional config like `NATS_PILLOW_HUB_AND_SPOKE`
   - Resolved by [replacePlaceholders()](../services/nats/service.go#L200-L240)

2. **PocketBase** ([services/pocketbase/service.json](../services/pocketbase/service.json))
   - Uses `${env.CORE_POCKETBASE_*}` for app config
   - Resolved by [ResolveEnv()](../services/pocketbase/service.go#L86-L103) and [replaceEnvPlaceholders()](../services/pocketbase/service.go#L105-L124)

3. **PocketBase-HA** ([services/pocketbase-ha/service.json](../services/pocketbase-ha/service.json))
   - Uses `${env.PB_*}` for HA-specific config (node name, replication URL, stream name)
   - Uses `${env.CORE_POCKETBASE_*}` for shared PocketBase config
   - Resolved by [replacePlaceholders()](../services/pocketbase-ha/service.go#L194-L234)

4. **Caddy** ([services/caddy/service.yaml](../services/caddy/service.yaml))
   - Currently uses direct env map (not placeholder-based)
   - Could be migrated to use `${env.*}` pattern

## Troubleshooting

### Problem: Environment variable not being set in service

**Symptom**: Service logs show empty or default values for configuration.

**Solution**:
1. Check `.env` file has the variable defined
2. Verify environment is loaded: `echo $VARIABLE_NAME`
3. Regenerate process-compose: `go run ./cmd/core stack start`
4. Check `.core-stack/process-compose.yaml` to see resolved value

### Problem: Changes to .env not reflected

**Symptom**: Updated `.env` values but service still uses old values.

**Solution**: Regenerate process-compose.yaml by restarting the orchestrator:
```bash
go run ./cmd/core stack down
go run ./cmd/core stack start
```

### Problem: Service fails with "missing required config"

**Symptom**: Service crashes or warns about missing required configuration.

**Solution**:
1. Check `service.json` to see which `${env.*}` variables are required
2. Add missing variables to `.env` file
3. Regenerate process-compose
4. Restart service

## Future Improvements

### Explicit .env Loading

Add `godotenv.Load()` to main() functions to automatically load `.env`:

```go
import "github.com/joho/godotenv"

func main() {
    // Load .env file from repository root
    if err := godotenv.Load(); err != nil {
        // Log but don't fail - .env might not exist in production
        log.Printf("No .env file found: %v", err)
    }

    // ... rest of main
}
```

### Environment Variable Validation

Add validation in `ResolveEnv()` to detect missing required variables:

```json
{
  "env": {
    "CORE_POCKETBASE_APP_URL": {
      "value": "${env.CORE_POCKETBASE_APP_URL}",
      "required": true
    }
  }
}
```

### Environment Templates

Support `.env.template` files that document all available variables:

```bash
# .env.template
# Copy to .env and fill in your values

# PocketBase Configuration
CORE_POCKETBASE_APP_URL=http://localhost:8090
CORE_POCKETBASE_ADMIN_EMAIL=admin@localhost
CORE_POCKETBASE_ADMIN_PASSWORD=changeme123
```

This would help new developers understand what needs to be configured.
