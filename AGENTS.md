# CLAUDE.md

Use the ./agents/AGENT.md, following everything it says and its links.

## Development Principles

### Package Boundaries & API Contracts
- **Work within one package at a time** to maintain speed and clarity
- **Use `go run . api-check`** to verify API contracts before changes
- **pkg/[package]/cmd/** contains internal commands for that package
- **pkg/cmd/** contains public CLI commands exposed to users
- **Public APIs**: Only use exported functions/types from other packages
- **Internal vs External**: Keep package internals separate from public CLI

### Configuration Must-Check Rule
- **ALWAYS check pkg/config FIRST** for any configuration needs
- **No hardcoded paths or URLs** - use pkg/config getters
- **pkg/config is the single source of truth** for defaults
- **File system paths, URLs, volume names** → pkg/config
- **Environment-aware defaults** → pkg/config
- **Future XDG/Docker paths** → pkg/config

### Test vs Production Data Separation
- **Tests use**: `.data-test/` directory (isolated, can be deleted safely)
- **Production uses**: `.data/` directory (persistent, backed up)
- **Environment detection**: `config.IsTestEnvironment()` automatically detects test runs
- **Path functions**: `config.GetFontPath()`, `config.GetCaddyPath()`, etc. are environment-aware
- **Pattern**: All data-heavy packages should use `config.Get*Path()` functions
- **Benefits**: Clean test artifacts, no production data pollution, easy debugging

### Web GUI Debugging
- **Always use Playwright MCP tools** for web GUI debugging
- **No manual browser interaction** - use:
  - `mcp__playwright__browser_navigate` to load pages
  - `mcp__playwright__browser_snapshot` to see current state
  - `mcp__playwright__browser_click` to interact with elements
  - `mcp__playwright__browser_evaluate` to inspect DOM
  - `mcp__playwright__browser_console_messages` to check errors
- **Test URLs**: http://localhost:1337 (service mode)
- **Debug routes**: /, /logs, /metrics, /docs, /status

### API Check Usage
- Always run `go run . api-check` after package changes
- This ensures we maintain backward compatibility
- Internal package commands (pkg/[pkg]/cmd/) don't need to be moved to pkg/cmd/

### Examples Structure
- **pkg/[package]/example/** contains standalone example modules for each package
- **Each example has its own go.mod** making it a separate module
- **Root go.work** includes all example modules as workspace members
- **Examples should be simple** and focused on demonstrating package API usage
- **Complex implementation code** belongs in the main package, not examples
- **No examples in central examples/ directory** - keep examples with their packages

### pkg/dep Package Rules
**When editing pkg/dep, follow these exact rules:**

#### 1. Configuration Format (dep.json)
- **Structure**: Use exact JSON format with `name`, `repo`, `version`, `release_url`, `assets[]`
- **Asset matching**: Use regex patterns in `match` field, escape dots (\.) and dollars ($)
- **Platform naming**: Use `"darwin"`, `"linux"`, `"windows"` for OS, `"amd64"`, `"arm64"` for arch
- **Archive handling**: Support `.zip` and `.tar.gz` formats

#### 2. Adding New Binaries (5 steps required)
1. **Add to dep.json**: Configure the binary with proper asset patterns
2. **Create installer file**: `<binary>.go` with `<binary>Installer` struct
3. **Add switch case**: In dep.go switch statement for installer selection
4. **Update CLI list**: In pkg/cmd/dep.go add to depListCmd hardcoded list
5. **Update documentation**: Add to dep.go supported binaries comment

#### 3. Installer Implementation Pattern
- **Struct**: `type <binary>Installer struct{}`
- **Method**: `func (i *<binary>Installer) Install(binary DepBinary, debug bool) error`
- **Archive handling**: Use `unzip()` for .zip, `untarGz()` for .tar.gz
- **Binary path**: Handle nested directories (e.g., `binary-name-version-os-arch/binary-name`)
- **Windows**: Add `.exe` extension for Windows binaries

#### 4. Testing Requirements
- Run `go run . dep list` to verify CLI integration
- Run `go run . api-check` to verify API compatibility
- Test installation with `go run . dep install <binary>`
- Verify binary works: `.dep/<binary> --version`

#### 5. Asset Pattern Examples
```json
{
  "name": "toolname",
  "repo": "owner/repo",
  "version": "v1.0.0",
  "release_url": "https://github.com/owner/repo/releases",
  "assets": [
    {
      "match": "toolname-.*-darwin-arm64\\.zip$",
      "os": "darwin",
      "arch": "arm64"
    },
    {
      "match": "toolname-.*-linux-amd64\\.tar\\.gz$",
      "os": "linux",
      "arch": "amd64"
    }
  ]
}
```

#### 6. Cross-Platform Build Support
- **`--cross-platform` flag**: Available for `go run . dep local install` command
- **Supported binaries**: Only works with `"source": "go-build"` binaries
- **Target platforms**: darwin/amd64, darwin/arm64, linux/amd64, linux/arm64, windows/amd64, windows/arm64
- **Binary naming**: Multi-platform builds create `binary-name-os-arch` files in `.dep/`
- **CGO limitations**: Cross-compilation automatically sets `CGO_ENABLED=0` and provides helpful error messages for CGO-dependent binaries
- **Usage examples**:
  ```bash
  # Install single binary for all platforms
  go run . dep local install garble --cross-platform
  
  # Install all binaries for all platforms  
  go run . dep local install --cross-platform
  ```
- **CGO workarounds**: For CGO-dependent binaries like litestream, use CGO-free forks or switch to `github-release` source

#### 7. Code Generation for Binary Constants
- **Garble-proof constants**: Binary names are auto-generated from `dep.json` to prevent obfuscation
- **Source of truth**: `pkg/dep/dep.json` drives both installation and Go constants
- **Generation**: Run `go generate ./pkg/config` to regenerate `binaries_gen.go`
- **Usage**: Use `config.BinaryLitestream` instead of `"litestream"` strings
- **Automatic**: Constants update automatically when `dep.json` changes
- **Type-safe**: Compile-time verification of binary references

#### 8. Dynamic Process Supervision with Goreman
- **Import-based**: No Procfiles needed - packages register themselves
- **Idempotent**: `goreman.RegisterAndStart()` handles registration + startup
- **Global registry**: Singleton pattern for automatic process management
- **Graceful shutdown**: SIGTERM with SIGKILL fallback after timeout
- **Status monitoring**: Centralized process health checking

**Usage pattern for packages:**
```go
// In pkg/myservice/service.go
func StartSupervised() error {
    return goreman.RegisterAndStart("myservice", &goreman.ProcessConfig{
        Command:    config.Get(config.BinaryMyService),
        Args:       []string{"--config", "./config.yml"},
        WorkingDir: ".",
        Env:        os.Environ(),
    })
}
```

**Service orchestration:**
```go
// Packages decide what to start - no central orchestration needed
litestream.StartSupervised("", "", "", false)
caddy.StartSupervised()
bento.StartSupervised()

// Graceful shutdown of all processes
goreman.StopAll()
```


