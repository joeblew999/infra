# cmd

Simple, idempotent service orchestration and CLI management.

## Philosophy

**One command to rule them all** - `infra` starts everything by default, with smart defaults and full control when needed.

## Usage

### Default (Start Everything)
```bash
go run .                    # Starts all services in PROD mode
go run . --env dev          # Starts all services in DEV mode
```

### CLI Tools (Always Available)
```bash
go run . --help             # Show all commands
go run . dep list           # Manage dependencies
go run . task              # Run Taskfiles
go run . build             # Build application
```

**Available tools**:
- `dep` - Binary dependency management
- `flyctl` - Fly.io deployment
- `task` - Task automation  
- `tofu` - Infrastructure as code
- `caddy` - Web server management
- `build/deploy/status` - Build/deploy workflows
- `ai/mcp/conduit` - Advanced debugging tools

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--env` | `production` | Environment: `production` or `development` |

## Behavior

- **Idempotent**: Safe to run multiple times
- **Always available**: All CLI tools accessible without mode switching
- **Smart defaults**: Uses pkg/config for all paths and ports
- **Environment-aware**: DEV mode enables migration tools, PROD mode is optimized
- **Always starts**: All services start by default, use CLI commands for control
- **Centralized config**: All settings come from pkg/config

## Examples

```bash
# Development setup
go run . --env dev

# Production mode
go run . --env prod

# CLI operations
go run . dep list
go run . task build
```

## Architecture

- **Unified interface**: All commands always available
- **Simple**: One entry point, clear defaults
- **Configurable**: All settings from pkg/config
- **Consistent**: Same pattern across all services







