# Gemini CLI 

Can we Used with Claude or without.

## extensions

https://github.com/google-gemini/gemini-cli/blob/main/docs/tools/mcp-server.md

The agent looks for extensions in two primary locations:
WORKSPACE/.gemini/extensions: This is within your current workspace or project directory. Extensions here are typically project-specific.
~/.gemini/extensions: This is in your home directory. Extensions here are global and available across all your Gemini CLI sessions.
If an extension with the same name exists in both locations, the one in your workspace directory will take precedence.

### Playwright

Playwright MCP: For browser automation, web scraping, UI testing. (You've already started this one!)

Installation: claude mcp add playwright npx '@playwright/mcp@latest'

How to use: In your Gemini CLI prompt, you'd refer to it, e.g., "Use playwright mcp to navigate to X and extract Y."