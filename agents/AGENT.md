# Infrastructure Management Repository

All Agent instructions live in this agents folder.

Use installed MCP Servers.

## Development


main.go runs everything. Keep coding until `go run .` works and all test pass.  

Also the githook is important because it forces code quality to be checked.  It calls go run actually..

Use gofmt, goimports, go vet.

## Agent Files


- AGENT_datastar.md - DataStar web interface
- AGENT_datastarui.md - DataStar UI components

Reference these files when working on respective components.