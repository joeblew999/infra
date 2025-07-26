# API Design Standards

## Package Documentation Requirements

Every package MUST include:

```go
// Package example provides [brief description].
//
// This package [detailed description of purpose].
//
// # Public API Guarantees
//
// The following functions and types form the stable public API:
//   - Function1(args) return - Description
//   - Type1 struct - Description
//   - ErrorType1 - Description
//
// # API Stability Contract
//
//   - Function signatures will not change without major version bump
//   - Error types will remain consistent for error handling
//   - Struct field names and JSON tags are stable
//   - Behavior contracts are documented and stable
//
// # Usage Example
//
//	// Example usage
//	result, err := example.Function1("input")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// # Breaking Change Policy
//
// - Major version bump required for breaking changes
// - Deprecation warnings for 2+ versions before removal
// - New functionality added alongside existing (not replacing)
package example
```

## API Design Rules

### 1. Error Handling
- Define package-specific error variables
- Use `fmt.Errorf` with `%w` for error wrapping
- Document all possible error conditions

### 2. Function Design
- Return `(result, error)` pattern for fallible operations
- Validate all inputs at public API boundaries
- Document preconditions and postconditions

### 3. Type Stability
- Never change exported struct field types
- Use JSON tags to decouple internal names from API
- Version embedded structs if needed

### 4. Backward Compatibility
- Add new fields to end of structs
- Use functional options for complex constructors
- Deprecate instead of removing

## Enforcement

1. Run `go doc -all .` to verify documentation quality
2. Run `./scripts/check-api-compatibility.sh` before merging
3. Use semantic versioning strictly
4. Code review focuses on API stability