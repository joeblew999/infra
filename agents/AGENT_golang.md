# Go Development Agent

Everything can be done via "go run .", so that devs and agents can control everything, including code quality checks, compiling, packaging, etc 

This embraces the "infrastructure as code" for infra. 



## Quick Start

Everything is controlled via `go run .` - no task files needed.

## Daily Commands

```bash
# Check for breaking changes
go run . api-check

# Run pre-commit checks (API + docs)
go run . pre-commit

# Run full CI checks
go run . ci

# View package documentation  
go doc -all ./pkg/packagename

# Development mode (when implemented)
go run . dev
```

## Package Requirements

Every Go package must have:

1. **Package documentation** with API contract:
```go
// Package example provides [brief description].
//
// # Public API Guarantees
// - Function1(args) return - Description
// - Type1 struct - Description
//
// # API Stability Contract
// - Function signatures will not change without major version bump
// - Error types remain consistent
//
// # Usage Example
// code example here
package example
```

2. **Package-specific error types**:
```go
var (
    ErrNotFound = fmt.Errorf("not found")
    ErrInvalid = fmt.Errorf("invalid input")
)
```

3. **Input validation** on public functions
4. **Tests** covering public API

## Automatic Protection

- Pre-commit hooks catch breaking changes
- CI/CD blocks breaking PRs  
- Documentation standards enforced

## Breaking Change Handling

**Safe changes:**
- Add new functions
- Add fields to end of structs
- Add new error types

**Breaking changes (require major version bump):**
- Change function signatures
- Remove functions/types
- Change struct field types

**If you get a breaking change warning:**
1. Add new functions instead of changing existing ones
2. Deprecate old functions rather than removing them
3. Add fields to structs instead of changing existing ones
4. If you must break: bump major version and coordinate with affected developers

**Common patterns:**
```go
// ✅ Safe - adding new functions
func NewFunction() error { ... }

// ✅ Safe - adding fields to end of structs  
type Config struct {
    ExistingField string
    NewField      int  // Safe to add
}

// ❌ Breaking - changing function signatures
func ExistingFunc(old string) string              // Before
func ExistingFunc(old string, new int) string     // Breaking
```

Reference implementation: `pkg/dep`

## TODO

### Fix API Compatibility Checker

Problem:

The API compatibility checker is reporting false positive breaking changes across all packages when running:

```bash
go run . api-check
go run . pre-commit  
```

Current behavior: Reports breaking changes in every package (docs, pkg/log, pkg/nats, etc.) even when those packages weren't modified.

Root cause: Bug in `runAPICompatibilityCheck()` function in `pkg/cmd/common.go` around lines 105-180.

Likely issues:
- Package discovery logic finding wrong directories
- apidiff command failing silently but being treated as breaking changes
- Git worktree setup not working correctly  
- Error handling interpreting failed apidiff runs as API breaks

Impact:
- Pre-commit hooks failing with false positives
- Developers will lose trust in API checking system
- Real breaking changes might be missed in the noise
- CI pipeline blocked by unreliable checks

Solution needed:
1. Debug the `runAPICompatibilityCheck()` function step by step
2. Add proper error handling and logging to see what apidiff is actually doing
3. Verify git worktree setup is working correctly
4. Test package discovery logic to ensure it finds the right directories
5. Verify apidiff is running correctly and interpreting results properly

Test commands:
```bash
# Should work without false positives
go run . api-check --old HEAD~1 --new HEAD
go run . pre-commit
```







