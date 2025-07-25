# Infrastructure Management Repository

All Agent instructions live in this agents folder.

This repository manages infrastructure using Taskfiles and Terraform/OpenTofu for container orchestration and cloud deployments, plus a DataStar-based web interface.




Use the installed MCP Servers as much as possible.

## Golang Debugging

The main.go runs everything. There is no Config to manage. So it is easy for you to always debug from their.

When coding, always keep going until the ```go run .``` works and the tests pass.

Use golang tools to help ensure good code quality:

- gofmt: This built-in tool automatically formats Go source code according to the official Go style guide. Consistent formatting improves readability and reduces merge conflicts.
- goimports: Similar to gofmt, goimports also manages imports, adding missing ones and removing unused ones, further enhancing code cleanliness.
- go vet: Another built-in tool, go vet examines Go source code for suspicious constructs that are likely to be errors, such as incorrect function calls or variable misuse.





## AI Context Files

This repository contains specialized Claude AI context files for different components:

- **[AGENT_taskfile.md](./AGENT_taskfile.md)** - Taskfile development patterns, component architecture, and best practices

- **[AGENT_datastar.md](./AGENT_datastar.md)** - DataStar development patterns, component architecture, and best practices for the web interface
- **[AGENT_datastarui.md](./AGENT_datastarui.md)** - DataStarUI component library guidance (referenced by web/README.md)

When working on all development and operations tasks, automatically reference these files for detailed guidance on DataStar patterns, component development, and UI best practices.