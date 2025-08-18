# Authentication Design

NOTE: at the stage of setting up example/main.go to use the caddy pkg, so we run under https.

## Overview

This package implements a passkey-based authentication system with NATS integration for distributed session management and user-scoped messaging.

---

https://github.com/go-webauthn/webauthn

module github.com/go-webauthn/webauthn

---

https://github.com/starfederation/datastar-go

module github.com/starfederation/datastar-go

https://cdn.jsdelivr.net/gh/starfederation/datastar@main/bundles/datastar.js

---

https://github.com/CoreyCole/datastarui

module github.com/coreycole/datastarui


## Architecture

```
Browser (WebAuthn) � Go Server � NATS Cluster � CLI Access
     �                  �            �           �
  Passkey Auth    Session Storage  KV Buckets  NKey Creds
```

## Components

### 1. WebAuthn/Passkeys
- **Device-bound** passkeys stored in browser/OS keychain
- Uses native biometric prompts (Touch ID, Face ID, Windows Hello)
- No cloud sync required - stays local to device
- Implemented with `github.com/go-webauthn/webauthn`

### 2. Datastar Integration
- Reactive HTML-over-SSE for real-time UI updates
- No JWT tokens needed - stateful server-side sessions
- Session state: `connectionID � userID` mapping
- Distributed session storage via NATS KV

### 3. NATS Cluster
- **Global mesh** across Fly.io regions using jeffh/nats-cluster
- **JetStream KV** for session storage with TTL expiration
- **Per-user namespaces** for isolated streams/subjects
- **Embedded leaf nodes** in each application server

### 4. User Management
- **Long-term users** in database/KV bucket
- **Short-lived NATS credentials** generated per session
- **Ephemeral NKey pairs** for CLI access
- **Auto-expiring permissions** tied to login sessions

## Authentication Flow

### Browser Authentication
1. User enters username in Datastar UI
2. WebAuthn registration/login ceremony
3. Browser shows native biometric prompt
4. Server validates cryptographic assertion
5. Session stored in NATS KV: `sessions[connectionID] = userID`
6. Subsequent SSE fragments authenticated via session lookup

### CLI Access
1. After successful browser auth, server generates ephemeral NKey pair
2. Private key seed delivered to browser via authenticated SSE
3. User saves seed and configures NATS context:
   ```bash
   nats context add myapp \
     --server wss://nats.cluster.fly.dev:443 \
     --nkey $(cat user.seed)
   ```
4. CLI has scoped access to user's namespace: `user.<userID>.>`

## NATS Permissions

Each authenticated user gets:

```go
Permissions{
    Publish: []string{
        "user.<userID>.>",                              // User subjects
        "$JS.API.STREAM.CREATE.user_<userID>",         // Create streams
        "$JS.API.STREAM.INFO.user_<userID>",           // Stream info
        "$JS.API.CONSUMER.CREATE.user_<userID>.>",     // Create consumers
    },
    Subscribe: []string{
        "user.<userID>.>",                              // User subjects
        "$JS.EVENT.ADVISORY.>",                         // Optional events
    },
}
```

## Session Storage

### NATS KV Configuration
```go
kv, _ := js.CreateKeyValue(&nats.KeyValueConfig{
    Bucket:   "sessions",
    Replicas: 3,           // Multi-region redundancy
    TTL:      30*time.Minute,  // Auto-cleanup idle sessions
})
```

### Multi-App Isolation
Each app gets its own KV bucket:
```go
bucket := fmt.Sprintf("sess-%s", os.Getenv("FLY_APP_NAME"))
```

## Deployment

### NATS Cluster (Fly.io)
- Deploy jeffh/nats-cluster to 6 regions: `ord fra sin sjc lhr iad`
- One VM per region: `fly scale count 6 --max-per-region 1`
- Automatic mesh discovery via Fly's private IPv6 network
- No auth required on private network (6PN isolation)

### Application Servers
- Embedded NATS leaf nodes in each app
- Connect to cluster: `nats://<cluster-app>.internal:4222`
- Auto-discover all regions for session replication

## Security Model

### Trust Boundaries
- **Fly 6PN**: Network-level isolation between organizations
- **WebAuthn**: Cryptographic proof of key possession
- **NATS Namespaces**: Subject-level isolation per user
- **Session TTL**: Time-bounded access without re-auth

### No Complex JWT
- No NATS operator/account JWT chains
- No credential files to manage
- Ephemeral NKeys generated per session
- Simple username+password fallback if needed

## User Experience

### Browser
- Click "Register/Login" � native biometric prompt
- Immediate access to authenticated UI via SSE
- No passwords, no 2FA codes, no email verification

### CLI
- Download seed file from authenticated browser session
- One-time NATS context setup
- Full JetStream access within user namespace
- Automatic credential expiry requires re-auth

## Advantages

1. **Passwordless**: Phishing-resistant WebAuthn
2. **Distributed**: Global session replication via NATS
3. **Stateful**: No JWT complexity, server-side sessions
4. **Scoped**: Per-user NATS namespaces for isolation  
5. **Real-time**: Datastar SSE for instant UI updates
6. **CLI-friendly**: Bridge browser auth to command-line tools
7. **Auto-cleanup**: TTL expiration prevents credential sprawl

## Implementation Notes

- Use pkg/config for all NATS connection strings
- Follow package boundary rules (work within pkg/auth)
- Leverage existing Datastar patterns in codebase
- Consider rate limiting for WebAuthn ceremonies
- Monitor NATS cluster health across regions