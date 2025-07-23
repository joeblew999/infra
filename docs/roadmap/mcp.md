# `mcp`: Multi-Cloud Platform Server Tools

This document outlines the design and integration of various Multi-Cloud Platform (MCP) server tools to support AI systems, focusing on testing, debugging, and operational diagnostics within the `infra` project.

Prefer golang based MCP Servers because they are small.


## Examples

Example of Claude being given a new MCP Server to help it debug the golang.


```sh
# Golang debuhgging: https://github.com/go-delve/mcp-dap-server
# https://github.com/go-delve/mcp-dap-server?tab=readme-ov-file#example-configuration-using-claude-code
go install github.com/go-delve/mcp-dap-server@latest
# start it
mcp-dap-server
'{{.CLAUDE__BINARY_NAME_NATIVE}} mcp add --transport sse mcp-dap-server http://localhost:8080'
```

There is along of dancing of processes required too with Claude. I have not tried Gemini cli and adding MCP yet.
- add the MCP
- start it
- start the Claude MCP server itself
- close and update claude.


### implementations

This is what we want to focus on:

https://github.com/google-gemini/gemini-cli/blob/main/docs/tools/mcp-server.md has the ones for gemini

https://github.com/modelcontextprotocol/go-sdk is the golang MCP SDK

https://github.com/modelcontextprotocol/go-sdk/network/dependents lists many great examples like

- https://github.com/go-delve/mcp-dap-server





### 1. Vision

To provide a robust and integrated suite of MCP server tools that enhance the development, testing, and operational capabilities of AI systems within the `infra` ecosystem. These tools will facilitate efficient debugging, automated testing, and real-time diagnostics across diverse cloud environments.

### 2. Core Functionality

*   **Automated Testing (Playwright):** Integrate Playwright as a headless browser automation tool for end-to-end testing of web interfaces and user flows, particularly for AI-driven applications.
*   **Remote Debugging (Golang Delve):** Provide seamless integration with Golang Delve for remote debugging of Go services, enabling developers to inspect runtime behavior and diagnose issues in distributed or containerized environments.
*   **Operational Diagnostics (Cloudflare MCP Server):** Utilize the Cloudflare MCP server for diagnosing failures and validating adjustments made to Terraform-managed infrastructure, offering insights into multi-cloud deployments.
*   **Centralized Management:** Explore mechanisms for managing and orchestrating these MCP tools from a central point within `infra`.

### 3. Tool-Specific Considerations

#### Playwright

*   **Purpose:** End-to-end testing of web UIs, especially those interacting with AI services.
*   **Integration:** Potentially run Playwright tests as part of CI/CD pipelines or on demand via `infra`'s CLI.
*   **Headless Execution:** Emphasize headless browser execution for efficiency in automated environments.

#### Golang Delve

*   **Purpose:** Remote debugging of Go services deployed across various environments (local, cloud, containers).
*   **Integration:** Provide a mechanism to attach Delve to running `infra` services or to launch services in debug mode.
*   **Security:** Ensure secure remote debugging connections, potentially through SSH tunneling or secure protocols.

#### Cloudflare MCP Server

*   **Purpose:** Diagnosing and validating Terraform adjustments across multi-cloud setups.
*   **Integration:** Leverage its capabilities to analyze Terraform state, plan outputs, and apply results for consistency and error detection.
*   **Observability:** Integrate its diagnostic outputs into `infra`'s logging or monitoring systems for better visibility.

### 4. Integration with `infra`

*   **`dep` Package:** The `dep` package will be responsible for managing the binaries for these MCP tools (e.g., Playwright browser binaries, Delve executable).
*   **`pkg/cmd`:** The `pkg/cmd` package will provide CLI commands to trigger Playwright tests, initiate Delve debugging sessions, or run Cloudflare MCP server diagnostics.
*   **Taskfiles:** Taskfiles will be used to orchestrate the execution of these tools, providing a consistent interface for developers.

### 5. Go Package API (Conceptual)

While these tools are external, `infra` might expose internal Go APIs to interact with them programmatically.

```go
package mcp

// RunPlaywrightTests executes Playwright tests.
func RunPlaywrightTests(testSuite string) error

// StartDelveSession starts a remote Delve debugging session.
func StartDelveSession(target string) error

// DiagnoseTerraform runs Cloudflare MCP server diagnostics on Terraform.
func DiagnoseTerraform(configPath string) error
```