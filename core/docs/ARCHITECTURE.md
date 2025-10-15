# Core Architecture

## Why

Core is **completely standalone** - ZERO dependencies on parent `infra/pkg/*`.

Goals: Portability, Simplicity, Clarity, Clean Slate

## How

1. **service.json** - manifests declare binaries, ports, health checks
2. **service.go** - implements Run(ctx) using native APIs
3. **cmd/*/main.go** - thin wrappers with signal handling
4. **process-compose** - orchestrates with health-based dependencies

## What Changed

Removed Pillow, NSC auth, cross-module imports. Pure standalone NATS.
