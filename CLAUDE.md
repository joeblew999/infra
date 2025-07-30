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

### API Check Usage
- Always run `go run . api-check` after package changes
- This ensures we maintain backward compatibility
- Internal package commands (pkg/[pkg]/cmd/) don't need to be moved to pkg/cmd/


