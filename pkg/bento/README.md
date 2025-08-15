# Bento Package

This package provides integration with [Bento](https://github.com/warpstreamlabs/bento), a stream processing system that enables real-time data processing and transformation.

## Offical Docs

https://warpstreamlabs.github.io/bento/docs/about

## Offical Playground

The bento binary has the WASM Playground embedded 

https://warpstreamlabs.github.io/bento/docs/guides/bloblang/playground/



## Overview

Bento is a lightweight, cloud-native stream processing engine that can handle various data sources and destinations. This package provides:

- **Service Management**: Start/stop bento as a managed service
- **Configuration Management**: Generate and manage bento configuration files
- **CLI Integration**: Command-line interface for managing bento
- **Environment Awareness**: Automatic configuration for development and production

## Usage

### Service Mode

When running in service mode, bento automatically starts alongside other services:

```bash
go run . service
```

This will:
- Start bento on port 4195
- Create default configuration if none exists
- Store configuration in `.data/bento/bento.yaml`

### CLI Commands

The following CLI commands are available:

#### Start bento service
```bash
go run . bento start
```

#### Check bento status
```bash
go run . bento status
```

#### Manage configuration
```bash
go run . bento config
```

### Configuration

The default configuration creates a simple generate→stdout pipeline:

```yaml
http:
  address: 0.0.0.0:4195
  enabled: true

input:
  generate:
    mapping: |
      root = { "message": "hello world", "timestamp": now() }
    interval: 5s

output:
  stdout: {}
```

## Directory Structure

```
.data/bento/
├── bento.yaml          # Main configuration file
├── logs/               # Bento logs
├── state/              # Stream state persistence
└── temp/               # Temporary processing files
```

## Development

### Testing

```bash
# Ensure bento is installed
go run . dep install bento

# Start bento service
go run . bento start

# Check status
go run . bento status
```





