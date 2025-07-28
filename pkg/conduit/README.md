# conduit

Binary management for [Conduit](https://github.com/ConduitIO/conduit) and its connectors.

## Quick Start

```bash
# Download all binaries
go test ./pkg/conduit -run TestPackageIntegration -v

# Verify binaries are ready
ls -la .dep/conduit*
```

## Usage

```go
// Download all binaries
err := conduit.Ensure(false)

// Get path to conduit binary
path := conduit.Get("conduit")
```

## Included Binaries

- **conduit** - Core Conduit binary (v0.12.1)
- **conduit-connector-s3** - S3 connector (v0.9.3)
- **conduit-connector-postgres** - Postgres connector (v0.14.0)
- **conduit-connector-kafka** - Kafka connector (v0.8.0)
- **conduit-connector-file** - File connector (v0.7.0)

Configuration files are in `pkg/conduit/config/` and use the same format as `pkg/dep`.