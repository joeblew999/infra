# DatastarUI Sample App

A minimal DatastarUI application used by the infra automation helpers. It ships with:

- A templ-rendered home page with a Datastar counter
- Tailwind CSS configuration and `static/css/index.css`
- Playwright config and test (`tests/button.spec.ts`)

Run the automated workflows from the repo root:

```bash
# Rebuild templ + Tailwind assets
GOWORK=off go run ./pkg/datastarui/cmd/codegen

# Execute the Playwright suite (Bun workflow by default)
GOWORK=off go run ./pkg/datastarui/cmd/playwright
```

Pass `--workflow=node` to reuse the pnpm/Tailwind/Playwright toolchain instead of Bun.
