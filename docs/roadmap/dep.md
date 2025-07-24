# Dependency Management

This document outlines the design for a Go package to manage external binary dependencies.

## STATUS

<!-- This section tracks, in a KISS way, what is still missing or needs attention. -->

caddy, task, tofu is done.



## What we dep

<!-- IMPORTANT: Do not delete any links in this section. They are intentionally added by the user. -->

task: https://github.com/go-task/task, https://github.com/go-task/task/releases/tag/v3.44.1

- so we can have task locally

gh: https://github.com/cli/cli, https://github.com/cli/cli/releases/tag/v2.76.1

- so the deps debugger works.

caddy: https://github.com/caddyserver/caddy, https://github.com/caddyserver/caddy/releases/tag/v2.10.0

- so we can wrap it all with caddy and have local https, so that things like Service works and A2HS can works

tofu: https://github.com/tofuutils, https://github.com/tofuutils/tenv

- SO we can do deployments.

digger: https://github.com/diggerhq/digger, https://github.com/diggerhq/digger/releases/tag/v0.6.110

- Later to see how we can use github CI to help with the tofu / terraform.

bento: https://github.com/warpstreamlabs/bento, https://github.com/warpstreamlabs/bento/releases/tag/v1.9.0

- So we can use it for reactive workflows

incus: https://github.com/lxc/incus, https://github.com/lxc/incus/releases/tag/v6.14.0

- MUCH later, so we can run dockers on dedciaetd hardware in the cloud. this needs linux.

### 1. Vision & Motivation
To provide a reliable, self-contained mechanism for `infra` to manage its essential external binary dependencies (like `tofu`, `task`, `caddy`). This ensures that these tools are always available and correctly versioned, simplifying development and deployment across different environments.

### 2. Core Functionality
*   **Automated Download & Installation:** Fetches specific, versioned binaries from their sources (e.g., GitHub releases).
*   **Idempotent:** Only downloads and installs a binary if it's missing or the version doesn't match, preventing unnecessary operations.
*   **Platform-Aware:** Handles different operating systems and architectures, including `.exe` extensions for Windows.
*   **Version Tracking:** Uses a metadata file (`_meta.json`) to track the version of each installed binary.

### 3. Storage
All managed binaries and their metadata are stored in the `./.dep/` directory. Binaries are named using a `{{name}}_{{os}}_{{arch}}` convention (e.g., `tofu_darwin_arm64`).

### 4. Go Package API (Conceptual)

```go
package dep

// Ensure checks for and installs all defined core binaries.
func Ensure(debug bool) error

// Get returns the absolute path to the requested binary for the current platform.
func Get(name string) string
```