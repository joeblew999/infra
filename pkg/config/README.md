# Config

Central configuration package for cross-platform paths, environment detection, and binary naming.

## Usage

```go
import "github.com/joeblew999/infra/pkg/config"

// Get binary paths
config.Get("flyctl")          // .dep/flyctl or .dep/flyctl.exe
config.GetFlyctlBinPath()     // .dep/flyctl
config.GetDepPath()          // .dep

// Environment detection
config.IsProduction()        // true on Fly.io
config.IsDevelopment()       // true locally

// Platform utilities  
config.GetBinaryName("tool") // "tool.exe" on Windows
```

## Environment

- **Development**: Uses local `.dep`, `.bin`, `.data` folders
- **Production**: Uses standard OS paths (Fly.io deployment)

## Paths

- `.dep/` - External binaries (flyctl, ko, caddy, etc.)
- `.bin/` - Compiled project binaries
- `.data/` - Application data (databases, stores)
- `docs/` - Markdown documentation

## Future

Need to use github.com/adrg/xdg, so that we store data on dekstops in the right place ? 

Eventually configuration will be stored in NATS JetStream KV store. This centralized `pkg/config` design enables seamless migration from local files to distributed configuration without code changes.


