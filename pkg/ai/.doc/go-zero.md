# go-zero

## Setup

```sh
go install github.com/zeromicro/go-zero/tools/goctl@latest

# https://go-zero.dev/docs/tasks/installation/protoc

# proto gen tools
goctl env check --install --verbose --force

# check proto is there
goctl env check --verbose

# vscode plugin
# https://go-zero.dev/docs/tasks/installation/goctl-vscode
# https://github.com/zeromicro/goctl-vscode
```

## New Project

**Important:** When creating new go-zero projects in this workspace, follow these steps:

```sh
# 1. Create the project with goctl
goctl api new myproject

# 2. CRITICAL: Fix the module name in go.mod
cd myproject
# Edit go.mod and change:
#   FROM: module main
#   TO: module github.com/joeblew999/infra/pkg/ai/go-zero/myproject

# 3. Run go mod tidy to resolve dependencies
go mod tidy

# 4. Add to workspace
# Edit root go.work file and add:
#   use ./pkg/ai/go-zero/myproject
```

I have asked for this to be fixed at: https://github.com/zeromicro/go-zero/issues/5126


**Why this is needed:**
- `goctl` generates `go.mod` with `module main` by default
- This breaks internal imports when the code references itself
- The correct module name allows proper dependency resolution
