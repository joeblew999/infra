# dep

Binary dependency management with design-by-contract guarantees. Downloads, caches, and manages external tools required by the system.

## How it works

- **Configuration**: `dep.json` defines supported binaries with GitHub release patterns
- **Selection**: Automatic platform detection (`runtime.GOOS`, `runtime.GOARCH`) with regex matching
- **Caching**: Versioned downloads stored locally for idempotency
- **API**: Stable public interface (`Ensure()`, `Get()`) with guaranteed backward compatibility

## Supported sources

| Source | Binaries | Pattern |
|--------|----------|---------|
| GitHub releases | flyctl, ko, caddy, task, tofu, bento, garble, bun | Platform-specific asset matching |
| npm registry | claude | Node.js CLI package |

## claude

**Distribution**: npm registry

Downloads `@anthropic-ai/claude-code` npm package and creates wrapper scripts for the Node.js CLI tool. The package handles native binary downloads internally.

**Current**: 1.0.62  
**Registry**: `https://registry.npmjs.org/@anthropic-ai/claude-code/-/claude-code-{VERSION}.tgz`

## bun

**Distribution**: GitHub releases

Downloads Bun runtime from oven-sh/bun releases for platform-specific binaries.

**Current**: bun-v1.2.19
**Release**: `https://github.com/oven-sh/bun/releases`

## Development

### Release checking

Check latest versions for all configured binaries:

```bash
go test -run TestCheckAllReleases -v
```

Check specific binary:

```bash
go test -run TestCheckGitHubRelease -v
```

### Testing utilities

```go
// Remove specific binary for testing
err := dep.Remove("bun")  // Deletes bun and metadata
err := dep.Remove("claude")  // Also removes claude-code directory
```