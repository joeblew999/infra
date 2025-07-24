# `mcp`: Model Context Protocol (MCP) Servers

This document outlines the role of Model Context Protocol (MCP) servers within `infra`, focusing on their integration with AI agents and their management.

## STATUS

<!-- This section tracks, in a KISS way, what is still missing or needs attention. -->


### 1. Vision
To enable seamless interaction between AI agents (e.g., Claude CLI, Gemini CLI) and the `infra` ecosystem, allowing agents to query, manage, and extend `infra`'s capabilities through MCP servers. We prefer Go-based MCP Servers for their small footprint and efficiency.

### 2. Motivation & Benefits
*   **Automated Development Tasks:** Agents can automate routine development tasks, such as setting up and configuring development environments, reducing manual effort and token usage.
*   **Extensible System:** `infra` will support both core MCP servers (provided by `infra`) and user-defined MCP servers, making the system self-similar and highly extensible.
*   **Reduced Cognitive Load:** Agents can abstract complex operations, providing a simpler interface for users and developers.

### 3. Flow: Orchestration and Choreography
Managing MCP servers, especially when installing new ones, requires a specific flow of operations. `pkg/gops` will be instrumental in orchestrating and choreographing these steps:

*   **Orchestration (Centralized Control):** `pkg/gops` will provide the capabilities to:
    *   **Identify Running Processes:** Determine which MCP servers are currently active and on what ports.
    *   **Graceful Stop:** Temporarily halt existing MCP servers to avoid conflicts during installation or updates.
    *   **Start/Restart:** Launch new or restart updated MCP servers.
    *   This ensures a controlled sequence of operations, preventing conflicts and ensuring system stability.

*   **Choreography (Event-Driven with NATS):** For more dynamic and decoupled interactions, `infra` will leverage NATS JetStream:
    *   MCP servers can publish events (e.g., "I'm starting," "I'm ready," "I'm shutting down") to specific NATS streams.
    *   Other `infra` components or agents can subscribe to these events and react accordingly, enabling a more distributed and resilient management pattern.
    *   This combined approach of orchestration (via `pkg/gops`) and choreography (via NATS events) will streamline the management of MCP servers, making the "dance" of installation and updates seamless.

### 4. MCP Server Implementations
MCP servers act as the bridge between AI agents and specific functionalities or data sources. They expose well-defined interfaces that agents can interact with.

#### Suggested MCP Servers

<!-- IMPORTANT: Do not remove these links. They are intentionally added by the user. -->

*   **Base SDK:** https://github.com/modelcontextprotocol/go-sdk
*   **Learning Resources:** https://github.com/modelcontextprotocol/go-sdk/network/dependents (good examples)

#### Our Core MCPs

<!-- IMPORTANT: Do not remove these links. They are intentionally added by the user. -->

*   **GoDoc MCP Server:** https://github.com/yikakia/godoc-mcp-server
    *   Provides information from `pkg.go.dev`, useful for all Go programmers.

*   **Delve DAP MCP Server:** https://github.com/go-delve/mcp-dap-server

- uses https://github.com/google/go-dap 

#### User-Contributed MCPs

<!-- IMPORTANT: Do not remove these links. They are intentionally added by the user. -->

*   **Google Spreadsheet MCP:** https://github.com/kazz187/mcp-google-spreadsheet
    *   An example of a user-contributed MCP server.

*   **RSS Feeds MCP:** https://github.com/meinside/rss-feeds-go
    *   Another example of a user-contributed MCP server.


### 5. Integration with `infra`

*   **`dep` Package:** The `dep` package will be responsible for managing the binaries for these MCP tools.
*   **`pkg/cmd`:** The `pkg/cmd` package will provide CLI commands to interact with MCP servers.
*   **Taskfiles:** Taskfiles will be used to orchestrate the execution of these tools, providing a consistent interface for developers.



