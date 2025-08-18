# pkg/dep

**Why**: Go binaries for external tools, without npm or brew.

**What**: Downloads tools like bun, claude, flyctl to `.dep/` folder.

```bash
go run . dep install   # install all
go run . dep install bun  # install one
go run . dep list     # see what's available
```

Uses `dep.json` for configuration. Binaries auto-detect your platform.

## Testing

```bash
# Quick tests (skip downloads)
go test -v -short ./...

# Full tests (with downloads)
go test -v ./...

# Coverage report
go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out -o coverage.html

# Update README with current binaries
go test -v -run TestUpdateREADME ./...
```

## Development Tools

```bash
# Check code quality (what VS Code Problems pane shows)
go vet ./...          # static analysis
gofmt -l .           # check formatting
gofmt -w .           # fix formatting
```





## Binaries
Currently configured: 16 binaries

- **flyctl**
- **ko**
- **caddy**
- **task**
- **tofu**
- **bento**
- **bun**
- **claude**
- **nats**
- **litestream**
- **deck-tools**
- **zig**
- **toki**
- **goose**
- **kosho**
- **gh**