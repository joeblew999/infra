# CLAUDE.md

## Agent Guides

Use the AGENTS.md at `/Users/apple/workspace/go/src/github.com/joeblew999/infra/agents/AGENTS.md`, following everything it says and its links.

The agent guides are located at `/Users/apple/workspace/go/src/github.com/joeblew999/infra/agents/hugo/content/agents/`:
- `golang/index.md` - Go service development patterns
- `process-compose/index.md` - Process orchestration patterns
- `datastar/index.md` - Datastar backend
- `datastarui-tool/index.md` - DatastarUI tooling
- `litestream/index.md` - Litestream backups
- `nats-jetstream/index.md` - NATS JetStream
- `fly/index.md` - Fly deployment
- `ko/index.md` - ko build

These guides provide context for working with this codebase.

## Documentation Structure

**README.md**: Quick start and operational commands only
- How to run services
- Basic commands (`go run . stack up`, etc.)
- What commands do (not why or how they work internally)
- Quick reference for developers who just want to use the system

**docs/**: In-depth architectural and design documentation
- `docs/ARCHITECTURE.md` - System design, generation flow, key decisions
- `docs/DEVELOPMENT.md` - Adding services, debugging, extending
- `docs/TROUBLESHOOTING.md` - Common issues and solutions

**Rule**: README is for operators. docs/ is for architects and maintainers.
