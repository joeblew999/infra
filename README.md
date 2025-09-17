# Infra

Everything here runs with one command:

```bash
go run .
```

That boots the supervised stack (web UI, NATS, PocketBase, Bento, Deck API, Caddy, XTemplate, optional mox). Stop it just as easily:

```bash
go run . shutdown
```

## Quick Start

```bash
git clone https://github.com/joeblew999/infra.git
cd infra
go run .              # start services
# ... hack ...
go run . shutdown     # stop services
```

## Everyday Commands

```bash
go run . status       # check health
go run . deploy       # deploy to Fly.io
go run . container    # build & run via ko + Docker
go run . --env development   # explicit dev mode
go run . service --no-mox=false   # include mox mail server when needed
go run . cli --help   # tooling & debugging namespace
```

## Need More?

This repo keeps deeper docs alongside the code:

- `docs/` – architecture notes, service guides, CLI details
- `pkg/` – package-level READMEs (goreman, nats, deck, etc.)
- `agents/` – instructions for automation agents working in this repo

If you ever forget what’s available, run `go run . --help`.
