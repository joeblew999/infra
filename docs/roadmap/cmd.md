# `cmd`: Go Command and Service Entry Point

This document outlines the design for the `pkg/cmd` Go package, which will serve as the primary entry point for the `infra` system, providing both command-line interface (CLI) and long-running service functionalities.

### 1. Vision

To create a highly reusable and easily importable Go package that encapsulates the core `infra` application logic. This package will allow developers to seamlessly integrate `infra`'s capabilities into their own Go applications, whether as a standalone CLI tool or as an embedded service, minimizing boilerplate and ensuring consistent behavior.

### 2. Cross-Platform Accessibility for JavaScript Runtimes

To enable `npm`, `bun`, and `deno` developers to leverage `infra` within their projects, we will ensure `infra` is easily consumable by these JavaScript/TypeScript environments. This involves:

*   **Pre-compiled Binaries:** Distributing pre-compiled `infra` binaries for various operating systems and architectures, making them readily available for download and execution.
*   **Simplified Installation:** Providing clear, concise instructions and potentially helper scripts (e.g., a simple `install.js` or `install.sh`) that can be integrated into `package.json` scripts, `deno.json`, or `bunfig.toml` for automated setup.
*   **Idiomatic Integration:** Exploring options for a thin JavaScript/TypeScript wrapper or utility functions that can abstract away the direct execution of the Go binary, providing a more native feel for developers in these ecosystems.
*   **Path Resolution:** Ensuring that the `infra` binary, once installed via `dep` or other means, can be easily located and executed by Node.js, Bun, and Deno processes.

### 3. Motivation: Why `pkg/cmd` instead of `/cmd`?

The traditional `/cmd` directory structure in Go projects typically contains a `main` package that is not easily importable by other Go modules. By designing `infra`'s entry point as `pkg/cmd`, we achieve:

*   **Reusability:** Other Go projects can import `pkg/cmd` and leverage `infra`'s CLI commands or service components directly, without needing to fork or copy code.
*   **Flexibility:** It allows for custom `main` functions in consuming projects that can configure and run `infra` in specific ways (e.g., integrating with different logging frameworks or dependency injection systems).
*   **Testability:** Core command logic can be more easily tested in isolation.

### 4. Core Functionality

*   **Importable Entry Point:** `pkg/cmd` will expose functions or structs that allow external Go applications to initialize and run `infra`'s CLI or service modes programmatically.
*   **Mode Selection (CLI vs. Service):** The package will utilize command-line flags to determine whether `infra` should operate as a short-lived CLI tool (executing a command and exiting) or as a long-running service (e.g., a web server or background worker).
    *   **Default Behavior:** By default, if no specific mode flag is provided, `infra` will run as a service.
    *   **Flag-Driven:** A dedicated flag (e.g., `--mode=cli` or `--mode=service`) will explicitly control the operational mode.
*   **Dependency Integration:** `pkg/cmd` will internally leverage the `dep` package (as described in `dep.md`) to ensure that any external binary dependencies required for its operation (e.g., `task`, `tofu`) are automatically managed and available.

### 5. Benefits

*   **Reduced Boilerplate:** Consuming projects can import `pkg/cmd` and get a fully functional `infra` CLI or service with minimal setup.
*   **Consistent Behavior:** Ensures that `infra` behaves identically whether run as a standalone binary or embedded within another application.
*   **Easier Updates:** Updates to `infra`'s core logic or dependencies can be propagated to consuming projects simply by updating the `infra` module version.
*   **Enhanced Ecosystem:** Facilitates the creation of a richer ecosystem around `infra`, allowing developers to build custom tools and integrations on top of its core functionalities.
