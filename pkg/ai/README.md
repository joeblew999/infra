# pkg/ai - AI-Powered Infrastructure Management

Claude code 

claude is on the OS path.

bun is on the OS path, so you can use "bun x" to do "bunx".

## CLAUDE MCP Examples

Shows all the smart ways to configure claude.

https://github.com/github/github-mcp-server/blob/main/docs/installation-guides/install-claude.md

This is a golang mcp, so does not use node or bun.

mcp:
	# make the AI smarter :) 
	# https://github.com/go-delve/mcp-dap-server
	go install github.com/go-delve/mcp-dap-server@latest
mcp-start:
	# must start the delve MCP on port 8080
	mcp-dap-server
mcp-claude-list:
	claude mcp list
mcp-claude-add:
	claude mcp add --transport sse mcp-dap-server http://localhost:8080
mcp-claude-del:
	claude mcp remove mcp-dap-server

---

bun example that works , as a reference:

```sh
claude mcp list
claude mcp remove github
claude mcp add github -- bun x -y @modelcontextprotocol/server-github
claude mcp list
```

```sh
claude mcp list
claude mcp remove github
# 1) Register the Playwright MCP via Bun
claude mcp add playwright -- bunx -y @modelcontextprotocol/server-playwright
# 2) Patch the JSON in-place with any extra env vars
jq '.mcpServers.playwright.env.PLAYWRIGHT
claude mcp list
```



