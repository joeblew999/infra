## NATS: Security and Federated System

This document outlines the role of NATS in providing a secure and federated system within `infra`.

## STATUS

<!-- This section tracks, in a KISS way, what is still missing or needs attention. -->

### 1. NATS Credentials for Access Control
NATS Security allows us to mint NATS Credentials for individual devices or users. This granular control enables us to precisely manage which NATS streams a particular user or device has access to. This capability is fundamental to building a federated system where access to data and services is strictly controlled and distributed.

### 2. Unified Authentication and Authorization
NATS Auth and NATS Auth Callout mechanisms provide a powerful way to implement a single, consistent authentication and authorization (AuthN/AuthZ) system across all `infra` components. This includes:
*   **Web Applications:** Authenticating and authorizing users accessing web interfaces.
*   **CLI Tools:** Controlling access for command-line users.
*   **Backend Services:** Securing inter-service communication.

By centralizing AuthN/AuthZ through NATS, we ensure a cohesive security posture and simplify management across the entire `infra` ecosystem.
