# `log`: Logging Strategy

This document outlines the logging approach for `infra`.

### 1. Vision

To provide simple, structured, and centralized logging for all `infra` components, enabling easy debugging and monitoring across environments.

### 2. Core Components

*   **`slog`:** All Go code will use Go's built-in `slog` package for structured logging.
*   **`slog-nats`:** The `github.com/samber/slog-nats` adapter will send `slog` output to NATS. This allows logs from distributed `infra` services to be collected centrally.

### 3. Adapters

The [`github.com/samber`](https://github.com/samber) organization provides various `slog` adapters (e.g., for different output formats or destinations). These adapters are Go packages and are typically compiled directly into the `infra` binary. This means you select and include the desired adapters at build time, and they become part of the executable.

### 4. Centralized Monitoring (MCP Integration)

An MCP (Multi-Cloud Platform) system will subscribe to the NATS log stream. This enables real-time viewing and debugging of `infra` services running anywhere, providing live operational insights.