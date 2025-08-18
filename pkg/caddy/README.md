# pkg/caddy

Configurable Caddy reverse proxy with automatic binary installation.

## Usage

```go
import "github.com/joeblew999/infra/pkg/caddy"

// Simple reverse proxy
config := caddy.NewPresetConfig(caddy.PresetSimple, 8080)
config.GenerateAndSave("Caddyfile")

runner := caddy.New()
runner.Run("run", "--config", ".data/caddy/Caddyfile")
```

## Presets

- `PresetSimple` - Single app reverse proxy
- `PresetDevelopment` - Main app + bento playground  
- `PresetFull` - Main app + bento + MCP server
- `PresetMicroservices` - Multi-service routing base

## Features

- Auto-downloads caddy binary via pkg/dep
- HTTPS with automatic certificates 
- Docker-ready .data/caddy/ configuration
- Fluent API for custom configurations




