# Documentation Index

The `docs/` directory hosts the canonical references for each subsystem. Update
these sources first, then keep any README summaries pointed here.

| Area | Canonical doc | Highlights |
| ---- | ------------- | ---------- |
| Runtime CLI | [runtime.md](runtime.md) | Stack operations, process maintenance, TUI/Web UI commands. |
| Tooling CLI | [tooling.md](tooling.md) | Fly + Cloudflare workflow, configuration, deployment checklist. |
| Controller | [controller/design.md](controller/design.md) | Desired-state API design, reconciliation flow. |
| Scaling | [SCALING.md](SCALING.md) | Multi-region/replica strategies for the runtime stack. |

**Authoring guidelines**
- Keep examples runnable; prefer full command snippets or minimal Go programs.
- Include validation steps (tests, smoke commands) when documenting workflows.
- Note any required environment variables or external dependencies explicitly.
- When moving or renaming docs, update this index and the relevant agent
guidance.

**Quick maintenance checklist**
- `go test ./...` — baseline before publishing doc-led changes.
- `go run ./cmd/core stack up && go run ./cmd/core stack down` — runtime smoke.
- `go run ./tooling workflow deploy --profile fly --app <app> --org <org>` — tooling smoke (use a test app/Org).
- `docs/` links verified via `rg "\.md"` searches after renames.
