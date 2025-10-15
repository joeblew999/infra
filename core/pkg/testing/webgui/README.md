# Web GUI Integration Tests

This package provides integration tests for the core stack's Web GUI components:
- NATS monitoring endpoint (`:8222`)
- PocketBase admin UI and API (`:8090`)
- Caddy reverse proxy (`:2015`)

## Running Tests

### Prerequisites

The core stack must be running:
```bash
go run ./cmd/core stack up
```

### Run All Integration Tests

```bash
# Run all Web GUI tests
go test -v ./pkg/testing/webgui/...

# Run specific test
go test -v ./pkg/testing/webgui/... -run TestPocketBaseAdmin

# Skip integration tests (useful for CI without stack running)
SKIP_INTEGRATION_TESTS=1 go test ./pkg/testing/webgui/...
```

## Test Coverage

### TestStackHealth
- **PocketBase Direct**: Verifies PocketBase health endpoint at `:8090/api/health`
- **Caddy Proxy**: Verifies Caddy proxy at `:2015/api/health`

### TestNATSHealth
- Verifies NATS monitoring endpoint at `:8222/healthz`

### TestPocketBaseAdmin
- Verifies PocketBase admin UI is accessible at `:8090/_/`

### TestCaddyProxy
- Verifies Caddy correctly proxies requests to PocketBase
- Compares responses from both direct and proxied endpoints

## Architecture

### Client (`client.go`)
```go
client := webgui.NewClient("http://127.0.0.1:8090")

// Wait for service to be ready
ctx := context.Background()
err := client.WaitForReady(ctx)

// Check health
err = client.CheckHealth(ctx)

// Make custom GET request
resp, err := client.Get(ctx, "/api/collections")
```

### Test Patterns

All integration tests:
1. Check `SKIP_INTEGRATION_TESTS` environment variable
2. Create context with timeout (30s default)
3. Use `webgui.Client` for HTTP requests
4. Log success with `✓` prefix

## Browser Testing with Playwright

For visual/browser testing, use the MCP Playwright tools available in Claude Code sessions:

1. **Manual Testing** (via Claude Code):
   - Open browser to http://127.0.0.1:8090/_/
   - Test login flows
   - Verify collection CRUD operations
   - Test proxy routing through Caddy

2. **Automated Browser Tests** (future enhancement):
   - Integrate with `github.com/playwright-community/playwright-go`
   - Add browser tests to this package
   - Support headless mode for CI/CD

## CI/CD Integration

### GitHub Actions Example

```yaml
name: Integration Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Start Stack
        run: |
          go run ./cmd/core stack up &
          sleep 10  # Wait for services to start

      - name: Run Integration Tests
        run: go test -v ./pkg/testing/webgui/...

      - name: Stop Stack
        if: always()
        run: go run ./cmd/core stack down
```

## Troubleshooting

### Tests Fail: "Service not ready"

**Solution**: Ensure stack is running
```bash
go run ./cmd/core stack status
```

### Tests Fail: "Connection refused"

**Solution**: Check ports are not in use
```bash
lsof -i :4222  # NATS
lsof -i :8090  # PocketBase
lsof -i :2015  # Caddy
```

### Skip Tests Locally

```bash
export SKIP_INTEGRATION_TESTS=1
go test ./...
```
