# `pkg/gops`: Process Management and System Introspection

This document outlines the design for the `pkg/gops` Go package, which will provide utilities for interacting with and monitoring system processes.

## STATUS

<!-- This section tracks, in a KISS way, what is still missing or needs attention. -->




### 1. Vision
To build a more resilient and user-friendly `infra` system by providing robust, cross-platform tools for managing and inspecting running processes, leveraging `gopsutil`. This package will complement process runners like `goreman` by providing deeper introspection and control capabilities.

## Usage Flow

1.  **Pre-flight Checks:** Before starting a service, `pkg/gops` can verify port availability and check for existing process conflicts.
2.  **Runtime Monitoring:** Once services are running, `pkg/gops` can monitor their health and resource usage.
3.  **Controlled Shutdowns:** Facilitate graceful termination of processes.


<!-- IMPORTANT: Do not delete any links in this section. They are intentionally added by the user. -->

https://github.com/shirou/gopsutil

https://github.com/mattn/goreman


### 2. Motivation & Benefits
*   **Reliable Service Startup:** Proactively check for port availability and existing process conflicts to prevent "address already in use" errors and ensure services start cleanly.
*   **Enhanced Process Control:** Facilitate graceful shutdowns and restarts of `infra` components, improving system stability.
*   **Runtime Verification:** Complement `pkg/dep` by verifying the presence and executability of required binaries at runtime.
*   **Basic Health Monitoring:** Quickly determine if `infra` services are running as expected.

### 3. Core Functionality
*   **Port Status:** Check if a specific network port is open or available.
*   **Process Status:** Identify if a process is running, by name or PID.
*   **Process Termination:** Safely stop running processes.
*   **Binary Verification:** Confirm a binary exists and is executable.

### 4. Service Discovery
`pkg/gops` will leverage NATS JetStream for dynamic service discovery. Services will register their presence and exposed ports/endpoints with JetStream upon startup. Other components can then query JetStream to discover available services and their locations, enabling flexible and resilient communication within the `infra` ecosystem.
