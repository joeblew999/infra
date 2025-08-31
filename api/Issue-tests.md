# Issue: goctl should generate test scaffolding for API projects

## Summary
goctl generates comprehensive API service code but doesn't provide any test scaffolding, forcing developers to write all test boilerplate from scratch. Adding basic test generation would significantly improve developer experience and encourage testing best practices.

## Current Behavior
When running `goctl api go -api example.api -dir .`, no test files are generated:

❌ **Missing test files:**
- `internal/handler/*_test.go` - Handler unit tests
- `internal/logic/*_test.go` - Business logic tests
- `internal/svc/servicecontext_test.go` - Service context tests
- `mcpserver_test.go` - Integration tests

## Expected Behavior
goctl should generate basic test scaffolding with proper structure and examples to help developers get started with testing.

## Proposed Solution
Add a `--with-tests` flag to `goctl api go` command that generates test files alongside the service code:

```bash
goctl api go -api example.api -dir . --with-tests
```

## Suggested Test Files to Generate

### Handler Tests
`internal/handler/toolslisthandler_test.go`:
```go
// Code scaffolded by goctl. Safe to edit.
// goctl 1.8.5

package handler

import (
    "net/http"
    "net/http/httptest"
    "testing"
    
    "github.com/stretchr/testify/assert"
)

func TestToolsListHandler(t *testing.T) {
    // TODO: Add your test cases here
    t.Skip("TODO: Implement test cases")
}
```

### Logic Tests
`internal/logic/toolslistlogic_test.go`:
```go
// Code scaffolded by goctl. Safe to edit.
// goctl 1.8.5

package logic

import (
    "context"
    "testing"
    
    "github.com/stretchr/testify/assert"
)

func TestToolsListLogic_ToolsList(t *testing.T) {
    // TODO: Add your test cases here
    t.Skip("TODO: Implement test cases")
}
```

### Integration Tests
`mcpserver_test.go`:
```go
// Code scaffolded by goctl. Safe to edit.
// goctl 1.8.5

package main

import (
    "net/http"
    "net/http/httptest"
    "testing"
    
    "github.com/stretchr/testify/assert"
)

func TestMain(t *testing.T) {
    // TODO: Add integration test cases here
    t.Skip("TODO: Implement integration tests")
}
```

## Why This Matters

1. **Developer Productivity** - Reduces time spent writing test boilerplate
2. **Testing Best Practices** - Encourages developers to write tests by providing structure
3. **Consistency** - Ensures uniform test patterns across projects
4. **Go Standards** - Follows established Go testing conventions
5. **Complete Solution** - Other code generators (like gRPC) provide test scaffolding

## Alternative Implementation Options

### Option 1: Always Generate Tests
Generate basic test scaffolding by default with every `goctl api go` command.

### Option 2: Separate Command
Add a dedicated command: `goctl api test -api example.api -dir .`

### Option 3: Template System
Allow customizable test templates that teams can modify for their needs.

## Benefits

- **Lower barrier to entry** for testing go-zero APIs
- **Faster development cycle** with ready-to-use test structure
- **Better code coverage** when tests are easier to create
- **Team consistency** with standardized test patterns

## References
- [Go Testing Documentation](https://pkg.go.dev/testing)
- [Testify Framework](https://github.com/stretchr/testify)
- [Go HTTP Testing](https://pkg.go.dev/net/http/httptest)

## Version Info
- goctl version: 1.8.5
- go-zero version: 1.9.0

Adding test scaffolding would make goctl a more complete code generation solution and significantly improve the developer experience for go-zero API projects.

## Repository
Submit this issue to: https://github.com/zeromicro/go-zero/issues/new