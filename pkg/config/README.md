# config

Centralized configuration for infra. Provides strongly-typed defaults for file system paths, URLs, and environment settings.

ALWAYS use const for strings, so that obfuscation works fine.

## Design Intent

**No config files needed** - just use the defaults. Other packages get their configuration needs from this single source of truth.

## What Goes Here

- **File system root paths** (`.dep`, `.data`, etc.)
- **URLs and endpoints** 
- **Environment-aware defaults** (dev vs prod)
- **Future XDG/Docker paths**
- **Volume names for host disks**

## Usage

```go
// Get any config value
configFile := config.GetLoggingConfigFile()
level := config.GetLoggingLevel()
path := config.GetDepPath()
```

## Environment Support

- **Development**: Local paths, debug logs
- **Production**: Optimized paths, warn logs  
- **Docker**: Volume paths, container defaults
- **XDG**: User-specific paths when available

## Structure

All configuration is accessed through strongly-typed getter functions. No JSON parsing needed by consuming packages.