## `proto`: Protocol Buffers for Version Compatibility

This document outlines the strategic use of Protocol Buffers (Protobufs) within the `infra` system.

We eventually MUST use this approach, rather than just goalng types.

## STATUS

<!-- This section tracks, in a KISS way, what is still missing or needs attention. -->


### 1. Motivation
In a distributed system with multiple components running on various devices, maintaining compatibility across different binary versions is crucial. Protobufs provide a language-agnostic, backward-compatible, and forward-compatible serialization format that allows us to evolve our data structures and APIs without breaking existing deployments.

### 2. Vision
To leverage Protobufs as the primary data interchange format for inter-service communication within `infra`, ensuring seamless operation and version tolerance across diverse environments.

### 3. Benefits
*   **Backward and Forward Compatibility:** Enables different versions of services to communicate effectively, simplifying deployments and upgrades.
*   **Language Agnostic:** Supports communication between components written in different programming languages.
*   **Efficient Serialization:** Provides a compact and efficient wire format, reducing network overhead.
*   **Clear API Definitions:** Protobuf schema files serve as clear, versioned definitions of our data structures and service interfaces.

### 4. Integration (Future Considerations)

<!-- IMPORTANT: Do not delete any links in this section. They are intentionally added by the user. -->
*   **NATS:** Protobuf messages will be used for publishing and subscribing data over NATS streams. https://github.com/delaneyj/toolbelt/tree/main/natsrpc
*   **Huma:** Protobuf definitions will inform and potentially generate API endpoints for Huma-based services. https://github.com/danielgtaylor/huma

### 5. Huma aspects


Huma is great because you write golang and not specs. This makes it so that if the code compiles, it probable works.

We use Chi (https://github.com/danielgtaylor/huma/tree/main/adapters/humachi)  and SSE ( https://github.com/danielgtaylor/huma/tree/main/sse ) with this, so that it can work well with DataStar and MCP systems. 





