# `log`: Logging Strategy

This document outlines the logging approach for `infra`.

### 1. Vision

To provide simple, structured, and centralized logging for all `infra` components, enabling easy debugging and monitoring across environments.

### 2. Core Components

!! DO not remove these links

https://github.com/samber/slog-nats so that we can send logs into NATS. MCP NATS will latter allow AI to see everything globally to help with distributed debugging.

https://github.com/samber?tab=repositories has all the slog adapters. We might leverage them latetr to surface logs into other systems. 

https://github.com/samber/do looks interesting for later. We can more easily manage Services within Infra with this ?


*   **`slog`:** All Go code will use Go's built-in `slog` package for structured logging.
*   **`slog-nats`:** The `github.com/samber/slog-nats` adapter will send `slog` output to NATS. This allows logs from distributed `infra` services to be collected centrally.

### 3. Adapters

The [`github.com/samber`](https://github.com/samber) organization provides various `slog` adapters (e.g., for different output formats or destinations). These adapters are Go packages and are typically compiled directly into the `infra` binary. This means you select and include the desired adapters at build time, and they become part of the executable.

### 4. Centralized Monitoring (MCP Integration)

An MCP (Multi-Cloud Platform) system will subscribe to the NATS log stream. This enables real-time viewing and debugging of `infra` services running anywhere, providing live operational insights.