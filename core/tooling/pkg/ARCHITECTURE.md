# Package Architecture

## Design Principles

1. **Single Responsibility** - Each file has one clear purpose
2. **Separation of Concerns** - Auth, verification, configuration, and storage are separate
3. **Consistent Patterns** - Fly and Cloudflare follow the same structure
4. **No Duplication** - Shared logic is extracted, not duplicated

## Package Structure

### Provider Packages (fly/, cloudflare/)

Each provider package follows this pattern:

```
pkg/fly/                      pkg/cloudflare/
├── auth.go        (137 LOC)  ├── auth.go         (241 LOC)
│   - EnsureToken()           │   - EnsureToken()
│   - RunAuth()               │   - RunAuth()
│   - InteractiveAuth()       │   - RunBootstrap()
│                             │
├── verify.go      (44 LOC)   ├── verify.go       (73 LOC)
│   - VerifyToken()           │   - verifyToken()
│                             │   - VerifyToken()
│                             │   - verifyPermissions()
│                             │
├── config.go      (103 LOC)  ├── config.go       (106 LOC)
│   - ConfigurePreferences()  │   - ConfigurePreferences()
│                             │
├── storage.go     (129 LOC)  ├── storage.go      (179 LOC)
│   - DefaultTokenPath()      │   - DefaultTokenPath()
│   - SaveToken()             │   - SaveToken()
│   - LoadToken()             │   - LoadToken()
│   - SaveSettings()          │   - SaveTokenForKind()
│   - LoadSettings()          │   - SaveSettings()
│                             │   - LoadSettings()
│                             │
├── queries.go     (70 LOC)   ├── queries.go      (102 LOC)
│   - DescribeFly()           │   - DescribeCloudflare()
│                             │
├── deploy.go      (123 LOC)  ├── dns.go          (94 LOC)
│   - Deploy()                │   - EnsureAppHostname()
│                             │
└── logging.go     (52 LOC)   ├── bootstrap.go    (116 LOC)
    - NewLogger()             │   - SelectPermissionGroups()
                              │   - CreateScopedToken()
                              └── bootstrap_test.go
```

### Auth Package (auth/)

Orchestration layer that delegates to provider packages:

```
pkg/auth/
├── service.go      (101 LOC)
│   - Service.EnsureFly()
│   - Service.EnsureCloudflare()
│   - RunFlyAuth()           # wrapper → fly.RunFlyAuth()
│   - RunCloudflareAuth()    # wrapper → cloudflare.RunCloudflareAuth()
│   - RunCloudflareBootstrap() # wrapper → cloudflare.RunCloudflareBootstrap()
│   - VerifyFlyToken()       # wrapper → fly.VerifyFlyToken()
│   - VerifyCloudflareToken() # wrapper → cloudflare.VerifyCloudflareToken()
│
├── prompt.go       (66 LOC)
│   - IOPrompter
│   - Notify()
│   - PromptSecret()
│
└── doc.go          (74 LOC)
    - Package documentation
```

### Types Package (types/)

Shared interfaces and data structures:

```
pkg/types/
├── prompt.go
│   - Prompter interface   # Shared by all packages
│   - PromptMessage
│   - PromptResponse
│
├── deploy.go
│   - DeploymentInfo
│   - FlyLiveInfo
│   - CloudflareLiveInfo
│
└── progress.go
    - ProgressTracker
```

## File Size Guidelines

- **< 150 LOC**: Perfect - single focused responsibility
- **150-250 LOC**: Good - well-organized with clear sections
- **> 250 LOC**: Consider splitting into focused modules

## Current Metrics

### Cloudflare Package
- auth.go: 241 LOC (auth flows: ensure, manual, bootstrap)
- storage.go: 179 LOC (token + settings management)
- bootstrap.go: 116 LOC (permission selection, token creation)
- config.go: 106 LOC (zone/account configuration)
- queries.go: 102 LOC (live info queries)
- dns.go: 94 LOC (DNS operations)
- verify.go: 73 LOC (token verification, permission checks)

### Fly Package  
- auth.go: 137 LOC (auth flows: ensure, run, interactive)
- storage.go: 129 LOC (token + settings management)
- deploy.go: 123 LOC (deployment operations)
- config.go: 103 LOC (org/region configuration)
- queries.go: 70 LOC (live info queries)
- logging.go: 52 LOC (logger adapter)
- verify.go: 44 LOC (token verification)

## Benefits of This Structure

1. **Easy Navigation** - File names clearly indicate purpose
2. **Maintainable** - Small, focused files are easier to understand and modify
3. **Testable** - Each module can be tested independently
4. **Reusable** - Verification and configuration logic is isolated
5. **Consistent** - Both providers follow the same pattern
6. **No Circular Dependencies** - Clear hierarchy with types at the base

## Adding a New Provider

To add a new provider (e.g., AWS, Azure):

1. Create `pkg/newprovider/` directory
2. Implement these files following the pattern:
   - `auth.go` - EnsureToken(), RunAuth() functions
   - `verify.go` - VerifyToken() function
   - `config.go` - ConfigurePreferences() function  
   - `storage.go` - Token and settings persistence
   - `queries.go` - Describe() function for live info
3. Add wrapper functions to `pkg/auth/service.go`
4. Update CLI commands in `internal/cli/`

The consistent pattern makes adding providers straightforward.
