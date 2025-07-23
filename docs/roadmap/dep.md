# `dep`: Go Binary Dependency Manager

This document outlines the design for a Go package to manage external binary dependencies. We envision a two-phased approach to dependency management:

1.  **Core Bootstrapping Binaries:** A small, fixed set of essential binaries required for initial developer laptop setup. These will be handled by built-in code within the `dep` package itself for maximum reliability during bootstrapping.
2.  **Generic Downloader and Runner:** A flexible, manifest-driven system for managing a wider range of project-specific CLI tools. This system will allow developers to easily integrate binary dependency management directly within their Taskfiles. For instance, a `hetzner_taskfile` could ensure the `hetzner` CLI is present in the `.dep` folder before executing commands that rely on it.

For both phases, the goal is to easily download, manage, and run these binaries.

## Architectural Foundations

### Constants and Paths

To ensure reusability and consistency across the codebase, all core file paths and types will be defined as constants within a dedicated `pkg/store` Go package. This includes:

*   `.dep` folder: The designated location for all downloaded and managed external binary dependencies.
*   `.bin` folder: The location for the project's own compiled binaries, which will also adhere to the `name_OS_ARCH` naming pattern.
*   `taskfiles` folder: The directory containing Taskfiles for various project automation tasks.

### Manifest Management

*   **Core Binaries Manifest:** The manifest for core bootstrapping binaries will be **embedded directly within the `dep` package**. This ensures that the `dep` binary is self-contained and portable, allowing it to function correctly across local development, CI/CD pipelines, and production environments without external configuration files.

    The core binaries included in this embedded manifest are:
    *   `caddy`
    *   `tofu`
    *   `task`

*   **Generic Binaries Manifest:** The manifest for generic (secondary) binaries will be an external `dependencies.json` file. This allows other developers to easily extend and manage project-specific tools by importing and configuring this system.

### 1. Vision

To create a zero-dependency, self-contained Go package that manages the download, extraction, and execution of versioned binary dependencies from GitHub releases. This tool will standardize how external CLI tools are fetched and used within the project, removing reliance on system-wide installations (`brew`, `apt`) or manual downloads. It will be driven by a clear and simple manifest file for generic dependencies, and built-in logic for core bootstrapping tools, leveraging the defined constants and manifest embedding strategies.

### 2. Core Features

*   **Manifest-Driven (Generic):** All generic dependencies will be defined in a single `dependencies.json` file in the project root, as described in the Manifest Management section.
*   **Built-in Logic (Core):** A small set of critical bootstrapping binaries will have their download and installation logic hardcoded within the `dep` package, utilizing the embedded manifest.
*   **GitHub Release Downloads:** Fetch specific, versioned tools directly from their public GitHub releases.
*   **Cross-Platform Archive Support:** Natively handle common archive formats (`.tar.gz`, `.zip`) for different operating systems and architectures.
*   **Idempotent by Design:** Before fetching any dependency, the tool will first check if the final target binary (e.g., `./.dep/tofu_darwin_arm64`) already exists. If the file is present, the entire download and extraction process for that dependency is skipped. This ensures that running the tool repeatedly has no side effects.
*   **Standardized File System Layout:** Use a predictable directory structure for storing temporary downloads and final binaries, adhering to the constants defined in `pkg/store`.
*   **Platform-Aware Execution:** Provide a simple way to get the path to the correct binary for the current OS and architecture, handling details like the `.exe` extension on Windows automatically.

### 3. Core Bootstrapping Binaries

These are the foundational tools necessary for the initial setup of a developer's environment. Their installation logic will be embedded directly within the `dep` package to ensure minimal external dependencies during the bootstrapping phase. The specific binaries included are listed under the "Core Binaries Manifest" section.

### 4. Configuration (`dependencies.json` Manifest) - For Generic Binaries

The tool is configured via a `dependencies.json` file. To handle the diverse and unpredictable naming conventions of release assets, the manifest uses a list of **asset selectors**.

For a given dependency, the `dep` tool determines the user's current OS and architecture. It then iterates through the `assets` list, looking for the first entry that matches the platform. That entry's `match` regular expression is then used to find and select the correct downloadable file from the GitHub release.

**Example Manifest:**
```json
[
  {
    "name": "tofu",
    "repo": "opentofu/opentofu",
    "version": "v1.7.2",
    "assets": [
      { "os": "linux", "arch": "amd64", "match": "tofu_.*_linux_amd64\\.tar\\.gz$" },
      { "os": "darwin", "arch": "arm64", "match": "tofu_.*_darwin_arm64\\.tar\\.gz$" },
      { "os": "windows", "arch": "amd64", "match": "tofu_.*_windows_amd64\\.zip$" }
    ]
  },
  {
    "name": "flux",
    "repo": "fluxcd/flux2",
    "version": "v2.2.3",
    "assets": [
      { "os": "linux", "arch": "amd64", "match": "flux_.*_linux_amd64\\.tar\\.gz$" },
      { "os": "darwin", "arch": "arm64", "match": "flux_.*_darwin_arm64\\.tar\\.gz$" }
    ],
    "in_archive_path": "flux"
  },
  {
    "name": "my-cli",
    "repo": "user/my-cli",
    "version": "v1.1.0",
    "assets": [
      { "os": "windows", "arch": "amd64", "match": "my-cli-windows\\.exe$" }
    ]
  }
]
```

**Manifest Fields:**

*   **`name`** (required): The local alias for the binary.
*   **`repo`** (required): The source GitHub repository.
*   **`version`** (required): The exact release tag.
*   **`assets`** (required): An array of asset selector objects.
    *   **`os`** (required): The target OS (`darwin`, `linux`, `windows`).
    *   **`arch`** (required): The target architecture (`amd64`, `arm64`).
    *   **`match`** (required): A **regular expression** used to match the filename of the asset in the GitHub release.
*   **`in_archive_path`** (optional): The path to the binary *inside* the archive. Useful if the binary is in a subdirectory (like `bin/` or `dist/`).
*   **`post_install_script`** (optional): A shell script to run after download and extraction for complex setup.

### 5. Filesystem and Naming

*   **Cache Directory:** All binaries will be stored in the `./.dep/` directory.
*   **Temporary Directory:** Downloads and extractions will occur in `./.dep/tmp/`, which is cleaned up after each successful operation.
*   **Binary Naming Convention:** The final binaries will be stored in `./.dep/` using the format: `{{name}}_{{os}}_{{arch}}{{extension}}`.
    *   Example on an M1 Mac: `.dep/tofu_darwin_arm64`
    *   Example on Windows: `.dep/tofu_windows_amd64.exe`

### 6. Go Package API (Conceptual)

The Go package should expose a simple API to be used by other parts of the project.

```go
package dep

// Ensure downloads and prepares all binaries defined in the manifest.
// This function will handle both core bootstrapping binaries and generic ones.
func Ensure() error

// Get returns the absolute path to the requested binary for the current platform.
func Get(name string) (string, error)
```