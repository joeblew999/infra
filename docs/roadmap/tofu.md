# Tofu

We use tofu, instead of Terraform.

Dep system ensures we have it and the right version.

Terraform configs are held in the terraform folder.

## STATUS

pkg/store knows the default Terraform configs folder.
pkg/dep ensures tofu binary is download.
pkg/tofu in place to run the tofu binary.
pkg/cmd is hooked up to pkg/tofu, so we can run tofu commands.

```sh
./infra --mode cli tofu providers

Providers required by configuration:
.
└── provider[registry.opentofu.org/lxc/incus] ~> 0.1

```

## Core Providers

*   **fly.io:**
    *   **Purpose:** To deploy and manage application servers on the Fly.io platform.
    *   **Documentation:** [https://registry.terraform.io/providers/fly-apps/fly/latest/docs](https://registry.terraform.io/providers/fly-apps/fly/latest/docs)
*   **cloudflare:**
    *   **Purpose:** For managing DNS records, CDN, and network tunnels.
    *   **Documentation:** [https://registry.terraform.io/providers/cloudflare/cloudflare/latest/docs](https://registry.terraform.io/providers/cloudflare/cloudflare/latest/docs)
*   **incus:**
    *   **Purpose:** For managing LXD/LXC containers or virtual machines, potentially for local development or specific bare-metal deployments.
    *   **Documentation:** [https://registry.terraform.io/providers/lxc/incus/latest/docs](https://registry.terraform.io/providers/lxc/incus/latest/docs)
*   **null:**
    *   **Purpose:** A no-op provider often used for running local commands, managing dependencies, or for testing purposes within OpenTofu configurations.
    *   **Documentation:** [https://registry.terraform.io/providers/hashicorp/null/latest/docs](https://registry.terraform.io/providers/hashicorp/null/latest/docs)

## Requirements & Architectural Approach

### Key Requirements:

1.  **Manage 100s of servers running `infra`:** Automated, scalable deployment and configuration across multiple clouds.
    *   **Multi-Cloud Deployment:** Deploy to approximately 4 clouds: Hetzner, OVH, Cloudflare (for edge services), and potentially others.
2.  **Metrics Collection:** Each server uses `gops` (or similar) to send memory, CPU, and disk metrics to NATS.
3.  **Autoscaling Trigger:** NATS acts as the trigger for OpenTofu-driven autoscaling actions.

### Infra Server Archetypes/Roles:

The `infra` server can run in different modes, taking on distinct roles:

*   **CLI Mode (`infra --mode cli`):** Used for manual OpenTofu operations, development, and debugging.
*   **Metric Agent Mode (`infra --mode service`):** Deployed on each server to collect and publish metrics to NATS.
*   **Autoscaling Orchestrator Mode (`infra --mode service`):** A central instance that listens to NATS metrics and triggers OpenTofu autoscaling actions.

### Server Deployment & Configuration (OpenTofu):

OpenTofu is ideal for defining server infrastructure. Each server (or groups of servers) would be represented by OpenTofu resources in `.tf` files.

*   **Cloud-Specific Abstraction with Modules:** OpenTofu configuration files will not be identical across different cloud providers (Hetzner, OVH, Cloudflare). Each cloud has unique resources and APIs. We will use OpenTofu modules to abstract these differences:
    *   **Provider-Specific Modules:** Create separate modules (e.g., `modules/server/hetzner`, `modules/server/incus-host`) to encapsulate cloud-specific details, exposing common input/output variables.
    *   **Root Configuration:** The main OpenTofu configuration will call these modules, allowing a consistent interface despite underlying cloud differences.
    *   **Variables:** Use variables to manage cloud-specific differences like regions, instance types, or image IDs.
*   **Incus Management:** OpenTofu will manage both the deployment and initial configuration of Incus hosts (e.g., installing Incus on a VM) and the lifecycle of containers/VMs *within* Incus using the `lxc/incus` provider.
*   **Scalability:** Use OpenTofu modules for reusable configurations, workspaces for distinct environments/instances, and potentially orchestration tools (like Terragrunt) for very large scale.
*   **`infra` Role:** The `infra` CLI will trigger OpenTofu actions (e.g., `infra tofu apply`) for specific servers or groups.

### Metrics Collection (`gops` & NATS):

The `infra` application (running in "service mode" on each server) will integrate `gops` to collect system metrics and use a NATS client to publish these to specific NATS topics (e.g., `metrics.server.<server_id>.cpu`). Each `infra` instance acts as a metric agent.

### Autoscaling Logic (NATS & OpenTofu):

A central `infra` instance (Autoscaling Orchestrator Mode) will subscribe to NATS metrics, analyze them against predefined rules, and trigger OpenTofu actions to adjust server counts. This can involve direct execution of `infra tofu apply` commands with dynamic variables. Future programmatic APIs via `pkg/tofu/runner.go` are a possibility for more granular control.

## State Management: OpenTofu Infrastructure State vs. Monitoring State

It is crucial to distinguish between two fundamentally different types of "state" in this architecture:

1.  **OpenTofu Infrastructure State (for Reconciliation):**
    *   **Purpose:** A record of *actual infrastructure resources* managed by OpenTofu, used for **reconciliation** (comparing desired vs. real-world state).
    *   **Characteristics:** Stored in a **remote backend** (e.g., S3/DynamoDB, Consul) with strong consistency and locking.
    *   **Granularity:** Managed per distinct deployment (core infra, autoscaling groups, individual servers) using separate root modules.
    *   **Module Encapsulation:** Module resources contribute to the root module's state; modules don't have independent state files.
    *   **Workspaces:** Ideal for isolating state when deploying the *same logical service* to different clouds or environments.
    *   **Backend Configuration:** Each root module has a unique backend configuration for state isolation.

2.  **Monitoring State (for Operational Awareness & Autoscaling Triggers):**
    *   **Purpose:** Dynamic data (CPU, memory, disk, etc.) collected from running servers, used for operational awareness and real-time autoscaling decisions.
    *   **Characteristics:** Typically stored in a **time-series database**, a **message queue (like NATS JetStream)**, or a dedicated monitoring system. Prioritizes high-volume ingestion and real-time access.

## Deployment Flow

This ordered list outlines the idempotent startup sequence for the `infra` system components:

1.  **Core Infrastructure Deployment:**
    *   **Action:** Deploy foundational shared infrastructure (e.g., NATS cluster, core networking, shared security groups) using OpenTofu.
    *   **Idempotence:** `infra tofu apply` ensures the desired state is achieved, creating resources if they don't exist or updating them if they've drifted.

2.  **Incus Host Deployment (if applicable):**
    *   **Action:** Provision and configure dedicated Incus hosts (VMs or bare metal) in your chosen cloud providers, including Incus installation and initial setup.
    *   **Idempotence:** `infra tofu apply` will ensure Incus is installed and configured as specified, skipping steps if already complete.

3.  **Metric Agent Deployment:**
    *   **Action:** Deploy `infra` servers configured as "Metric Agents" to your cloud providers. This includes provisioning the underlying compute instances and installing the `infra` binary.
    *   **Idempotence:** `infra tofu apply` ensures the correct number of agents are provisioned and the `infra` binary is present.

4.  **Autoscaling Orchestrator Deployment:**
    *   **Action:** Deploy the central `infra` server configured as the "Autoscaling Orchestrator."
    *   **Idempotence:** `infra tofu apply` ensures the orchestrator instance is provisioned and the `infra` binary is present.

5.  **Start Metric Agent Services:**
    *   **Action:** On each deployed Metric Agent server, start the `infra` application in "service mode" (`infra --mode service`) to begin collecting and publishing metrics to NATS.
    *   **Idempotence:** The service itself should be designed to handle restarts, reconnects to NATS, and resume metric collection without issues.

6.  **Start Autoscaling Orchestrator Service:**
    *   **Action:** On the deployed Autoscaling Orchestrator server, start the `infra` application in "service mode" (`infra --mode service`) to begin listening to NATS metrics and triggering autoscaling actions.
    *   **Idempotence:** The orchestrator service should be designed to handle restarts, reconnects to NATS, and resume its autoscaling logic.

