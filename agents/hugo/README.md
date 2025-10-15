# Agents Docs

Everything in this folder is written for AI agents *and* humans from the start. Every guide or task lives as plain Markdown so tools can read it directly, and the same files can be served by Hugo when you want a web view.

## Layout
- `content/agents/<slug>/index.md` – agent playbooks (Part 1 fundamentals + Part 2 repo specifics).
- `content/tasks/*.md` – task backlog items, identical to the historic `tasks/` files.
- Legacy `AGENT_*.md` and `tasks/*.md` now point to the new locations for backwards compatibility.

## Runbook
```sh
brew install hugo

# from this directory
go run .              # hugo server on http://127.0.0.1:1414
go run . -mode build  # render static site to ./public
```

### Where To Read
- **Filesystem**: open the Markdown under `content/agents/...` or `content/tasks/...` (AI agents can parse the front matter + body directly).
- **HTTP**: after `go run . -mode build`, serve `./public` (or just the JSON manifest when we add it) so remote agents and humans fetch the same content.

### Quick Smoke Test
1. `go run . -mode build`
2. `python3 -m http.server 1414 --directory public` *(optional static preview)*
3. `go run .` and browse `http://127.0.0.1:1414/`

That’s it—add or edit guides/tasks inside `content/`, rebuild when you want the web view, and agents can always read the Markdown source.
