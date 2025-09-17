# Repository Guidelines

## Project Structure & Module Organization
`main.go` orchestrates goreman-managed services (NATS, PocketBase, Deck API, Datastar web). Reusable packages live in `pkg/` (`pkg/nats`, `pkg/pocketbase`, `pkg/deck`, `pkg/config`); keep new code single-purpose and export only minimal APIs. Web UI assets stay in `web/` (`handlers.go`, `templates/`, `static/`), while go-zero APIs and experiments live in `api/`. Infrastructure definitions sit in `terraform/` and `nats-cluster/`, and onboarding notes for agents are in `agents/`.

## Build, Test, and Development Commands
- `go run .` boots the full stack; add `--env=development` for local tweaks.
- `go run . shutdown` stops goreman processes cleanly.
- `go run . status` surfaces runtime health checks.
- `go run . deploy` runs the Fly.io deployment workflow (idempotent).
- `go run . container` builds the OCI image with `ko` for local Docker testing.
- `go build -o infra .` produces a standalone binary for packaging or CI.

## Coding Style & Naming Conventions
Target Go 1.25.1 Always run `gofmt` and `goimports`; the repo hooks expect both. Keep package names lower-case, exported symbols in `PascalCase`, and follow the existing logging helpers in `pkg/log` instead of bespoke prints. Run `go vet ./...` when you add a new package or touch critical initialization paths.

## Testing Guidelines
`go test ./...` should pass before every push. Free that port before running or use `go test -run TestGitHashInjection` while iterating. Capture coverage for service-heavy changes with `go test ./pkg/... -cover` and drop the summary in the PR description.

## Commit & Pull Request Guidelines
Use the conventional-commit prefixes that already exist (`feat:`, `fix:`, `chore:`) so automation stays reliable. Keep commits focused and messages in imperative mood. PRs need a clear summary, linked issues (`Fixes #123`), and proof of successful tests. Include screenshots for `web/` changes and highlight modifications to `.env`, Fly.io secrets, or Terraform resources.


