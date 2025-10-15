# Tooling Architecture

## Overview

The tooling system provides deployment and infrastructure management through multiple interfaces (CLI, TUI, Web GUI) backed by a unified service layer.

## Architecture Layers

```
┌─────────────────────────────────────────────────────────┐
│  Interfaces (CLI / TUI / Web GUI)                       │
│  - CLI: cobra commands                                  │
│  - TUI: bubbletea UI                                    │
│  - Web: DataStar + SSE                                  │
└────────────────┬────────────────────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────────────────────┐
│  Service Layer (CQRS + SSE Events)                      │
│  - Commands: Deploy, Auth, Configure                    │
│  - Queries: GetStatus, ListApps                         │
│  - Events: Progress, Completion, Errors                 │
└────────────────┬────────────────────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────────────────────┐
│  Providers (pkg/providers/)                             │
│  - fly/      (Fly.io client, auth, deploy)              │
│  - cloudflare/ (CF client, auth, DNS)                   │
│  - Future providers can be added                        │
└─────────────────────────────────────────────────────────┘
```

## Provider Structure

Each provider is self-contained in `pkg/providers/{name}/`:

```
pkg/providers/fly/
├── client.go       # Fly.io API client wrapper
├── auth.go         # Authentication flow
├── deploy.go       # Deployment operations
├── storage.go      # Token and settings persistence
└── types.go        # Provider-specific types

pkg/providers/cloudflare/
├── client.go       # Cloudflare API client wrapper
├── auth.go         # Authentication flow
├── dns.go          # DNS operations
├── storage.go      # Token and settings persistence
└── types.go        # Provider-specific types
```

## Service Layer Pattern

### CQRS (Command Query Responsibility Segregation)

**Commands** (write operations):
- `DeployCommand` - deploy an app to a provider
- `AuthCommand` - authenticate with a provider
- `ConfigureCommand` - update provider settings

**Queries** (read operations):
- `GetDeploymentStatusQuery` - check deployment progress
- `ListAppsQuery` - list apps for a provider
- `GetAuthStatusQuery` - check if authenticated

### SSE (Server-Sent Events)

All long-running operations emit progress events:

```go
type Event struct {
    Type    string      // "progress", "complete", "error"
    Step    string      // current operation step
    Message string      // human-readable message
    Data    interface{} // structured data
}
```

Example flow:
1. Client issues `DeployCommand`
2. Service validates and starts deployment
3. Service emits events: `{"type":"progress","step":"building",...}`
4. Service emits events: `{"type":"progress","step":"pushing",...}`
5. Service emits final: `{"type":"complete","step":"deployed",...}`

## Browser Auth Handling

### CLI/TUI Mode
```go
// Open browser directly
url := startAuthFlow()
openBrowser(url)
pollForToken()
```

### Web GUI Mode
```go
// Return redirect URL in event
event := Event{
    Type: "auth_required",
    Data: map[string]string{
        "redirect_url": authURL,
    },
}
// Frontend handles redirect
```

## Interface Integration

### CLI
```bash
$ core-tool deploy --app myapp
Deploying to Fly.io...
▸ Building image... ✓
▸ Pushing image... ✓
▸ Creating release... ✓
Deployed: v42
```

### TUI
```
┌─ Deploy Status ──────────────────┐
│ App: myapp                       │
│ Provider: Fly.io                 │
│                                  │
│ [████████████░░░░] 75%           │
│ ▸ Creating release...            │
└──────────────────────────────────┘
```

### Web GUI (DataStar SSE)
```html
<div data-on-sse="/deploy/stream">
  <div data-text="$step"></div>
  <progress data-value="$progress"></progress>
</div>
```

## Service Implementation

```go
type Service struct {
    providers map[string]Provider
    events    chan Event
}

type Provider interface {
    Name() string
    // Providers are NOT required to implement common methods
    // Each provider exposes its own specific API
}

// Commands return event stream
func (s *Service) Deploy(ctx context.Context, cmd DeployCommand) (<-chan Event, error) {
    events := make(chan Event)

    go func() {
        defer close(events)

        events <- Event{Type: "progress", Step: "building"}
        // ... deployment logic
        events <- Event{Type: "complete", Step: "deployed"}
    }()

    return events, nil
}
```

## Migration Plan

1. Create `pkg/providers/` structure
2. Move existing code:
   - `pkg/fly/*` → `pkg/providers/fly/`
   - `pkg/auth/fly.go` → `pkg/providers/fly/auth.go`
   - `pkg/storage/fly.go` → `pkg/providers/fly/storage.go`
   - Same for Cloudflare
3. Build service layer in `pkg/service/`
4. Update interfaces (CLI/TUI/Web) to use service layer
5. Remove old scattered code

## Key Design Decisions

1. **No unified provider interface** - Fly and Cloudflare are too different
2. **Provider-specific APIs** - each provider exposes what it needs
3. **Service layer handles orchestration** - not providers
4. **Events for all interfaces** - CLI polls, TUI updates, Web streams SSE
5. **Browser auth is provider-specific** - handled in each provider's auth flow
