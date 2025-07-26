# Go Development Agent

Everything can be done via "go run .", so that devs and agents can control everything, including code quality checks, compiling, packaging, etc 

This embraces the "infrastructure as code" for infra. 



## Quick Start

Enable automatic API protection:
```bash
task this:api:setup-hooks
```

## Daily Commands

```bash
# Check for breaking changes
task this:api:check

# View package documentation  
task this:api:docs -- ./pkg/packagename
go doc -all ./pkg/packagename

# Test before committing
task this:test:all
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

### golang development

Problem:

We need to formalise the goalng development flow also, to enforce good code quality.

We currently have it scatttered around from the github root at :

./.githooks/pre-commit

./.github/workflows/api-compatibility.yml

./script/check-api-compatibility.sh

./pkg/cmd

We want "go run ." to be the way we work always, so how do we incorporate that ? 

Solution:

Create a unified development CLI in main.go, using pkg/cmd

```
go run . pre-commit    # Run pre-commit checks
go run . ci            # Run CI checks  
go run . api-check     # API compatibility check
go run . dev           # Development mode
go run . help          # Show commands
go run .               # Runs essentiually what is production
```

Update existing files to call `go run .` commands. Move quality check logic from shell scripts into Go code.

./../pkg/cmd structure with CLI framework in place:

  - root.go - Root command setup
  - cli.go - CLI utilities
  - common.go - Shared functionality
  - service.go - Service commands
  - workflows.go - Workflow commands

  You can add your development commands (pre-commit, ci, api-check, dev) to this
  existing CLI structure rather than starting from scratch.

Looking at the existing pkg/cmd structure and the 3 files' functionality:

  Code Distribution:

  .githooks/pre-commit logic →
  - workflows.go - Pre-commit workflow command
  - common.go - Git staged files detection, documentation checks

  .github/workflows/api-compatibility.yml logic →
  - workflows.go - CI workflow command
  - service.go - API compatibility service functions

  scripts/check-api-compatibility.sh logic →
  - service.go - Core API diff functionality (git worktrees, apidiff calls)
  - common.go - Package discovery, error handling utilities

  This keeps workflow orchestration in workflows.go, reusable services in service.go,
  and shared utilities in common.go.







