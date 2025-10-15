# Agents Directory Index

The canonical guides now live under `content/agents/<slug>/index.md`. This file sticks around so older links still land here, but new references should point straight at the Markdown bundles (or the Hugo site once built).

## How To Reference From Subprojects
```markdown
# Agents

This project shares the central instructions in [../agents/content/agents](../agents/content/agents).
```
Adjust the relative path as needed.

## Guide Catalog
| Guide | Path |
| ----- | ---- |
| Go Service Guide | `content/agents/golang/index.md` |
| Datastar Backend | `content/agents/datastar/index.md` |
| DatastarUI Tooling | `content/agents/datastarui-tool/index.md` |
| Process Compose | `content/agents/process-compose/index.md` |
| Litestream | `content/agents/litestream/index.md` |
| NATS JetStream | `content/agents/nats-jetstream/index.md` |
| Fly Deployment | `content/agents/fly/index.md` |
| ko Build | `content/agents/ko/index.md` |

## Tasks
Tasks moved to `content/tasks/*.md`; the legacy `tasks/` files now link there.

## Tooling Expectations
- `go run .` from this directory wraps `hugo` so you can preview the docs locally.
- Keep guides AI-friendly: front matter + Markdown, no build steps needed to read them.
