# Provider Code Consolidation Plan

✅ **STATUS: COMPLETED** - All refactoring steps have been successfully completed!

## Original Mess

## Current Mess

```
pkg/
├── fly/
│   ├── deploy.go           # Fly deployment operations
│   ├── logging.go          # Fly logger
│   ├── token_store.go      # Token storage (DUPLICATE!)
│   └── settings.go         # Settings storage (DUPLICATE!)
├── cloudflare/
│   ├── dns.go              # DNS operations
│   ├── bootstrap.go        # Bootstrap operations
│   ├── token_store.go      # Token storage (DUPLICATE!)
│   └── settings.go         # Settings storage (DUPLICATE!)
├── providers/
│   ├── fly.go              # DescribeFly() - read-model query
│   └── cloudflare.go       # DescribeCloudflare() - read-model query
├── auth/
│   ├── fly.go              # Fly authentication flow
│   ├── cloudflare.go       # Cloudflare authentication flow
│   └── cloudflare_bootstrap.go # Cloudflare bootstrap auth
└── storage/
    ├── fly.go              # DUPLICATE of pkg/fly/token_store.go + settings.go
    └── cloudflare.go       # DUPLICATE of pkg/cloudflare/token_store.go + settings.go
```


## Migration Steps

### Step 1: Consolidate Fly Package ✅
- [x] Create `pkg/fly/storage.go` merging `token_store.go` + `settings.go`
- [x] Move `pkg/auth/fly.go` → `pkg/fly/auth.go`
- [x] Move `pkg/providers/fly.go` → `pkg/fly/queries.go`
- [x] Update all imports

### Step 2: Consolidate Cloudflare Package ✅
- [x] Create `pkg/cloudflare/storage.go` merging `token_store.go` + `settings.go`
- [x] Move `pkg/auth/cloudflare.go` + `cloudflare_bootstrap.go` → `pkg/cloudflare/auth.go` + `bootstrap_impl.go`
- [x] Move `pkg/providers/cloudflare.go` → `pkg/cloudflare/queries.go`
- [x] Update all imports

### Step 3: Update Auth Service ✅
- [x] Update `pkg/auth/service.go` to call `fly.EnsureFlyToken()` and `cloudflare.EnsureCloudflareToken()`
- [x] Add wrapper functions in auth package for backwards compatibility
- [x] Keep only generic prompt/auth orchestration in `pkg/auth/`
- [x] Move Prompter interface to `pkg/types/` to avoid circular dependencies

### Step 4: Cleanup ✅
- [x] Delete `pkg/storage/` (was already deleted)
- [x] Delete `pkg/providers/` (was already deleted)
- [x] Remove backup files from `pkg/auth/`
- [x] Run tests - All passing
- [x] Build entire project - Successful

## Benefits Achieved

1. ✅ **Single source of truth** - all Fly code in `pkg/fly/`, all CF code in `pkg/cloudflare/`
2. ✅ **No duplication** - one storage implementation per provider
3. ✅ **Clear boundaries** - each provider package is self-contained
4. ✅ **Easier to extend** - add new providers by creating new packages
5. ✅ **Auth service simplified** - just orchestrates via wrapper functions, doesn't implement

## Final Structure

```
pkg/
├── fly/                    # ALL Fly.io code ✅
│   ├── auth.go            # Authentication (moved from pkg/auth/)
│   ├── deploy.go          # Deployment operations (existing)
│   ├── storage.go         # Token + settings storage (consolidated)
│   ├── logging.go         # Logger (existing)
│   └── queries.go         # DescribeFly (moved from pkg/providers/)
├── cloudflare/             # ALL Cloudflare code ✅
│   ├── auth.go            # Manual authentication (moved from pkg/auth/)
│   ├── bootstrap_impl.go  # Bootstrap authentication
│   ├── bootstrap.go       # Bootstrap helpers (existing)
│   ├── dns.go             # DNS operations (existing)
│   ├── storage.go         # Token + settings storage (consolidated)
│   └── queries.go         # DescribeCloudflare (moved from pkg/providers/)
├── auth/                   # Generic auth orchestration ✅
│   ├── service.go         # Auth service with wrapper functions
│   ├── prompt.go          # IOPrompter implementation
│   └── doc.go             # Documentation
├── types/                  # Shared types ✅
│   ├── prompt.go          # Prompter interface (moved to avoid cycles)
│   ├── deploy.go          # Deployment types
│   └── progress.go        # Progress types
└── [DELETED] storage/      # ✅ Removed
└── [DELETED] providers/    # ✅ Removed
```

## Key Architectural Decisions

1. **Prompter Interface in types/** - Moved to avoid circular dependency between auth and provider packages
2. **Wrapper Functions** - Auth package provides backwards-compatible wrappers that delegate to provider packages
3. **No Self-Imports** - Removed cloudflare package self-imports that were causing issues
4. **Type Safety** - Auth wrappers return proper types (cf.APITokenVerifyBody, flyapi.Client) instead of interface{}

## Verification

- ✅ All packages compile without errors
- ✅ Tests pass (cloudflare package has tests)
- ✅ Full project builds successfully (48MB binary created)
- ✅ No circular dependencies
- ✅ CLI commands updated and functional
