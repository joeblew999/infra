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


