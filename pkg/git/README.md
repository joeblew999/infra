# git

Git workflow tools and integrations for systematic development.

## Overview

This package provides Git tooling integrations focused on **stacked branch development** workflows. Rather than managing Git complexity manually, we use specialized tools designed for systematic branch management.

## Tools

### git-spice

**git-spice** is our primary tool for managing stacked Git branches - essential when working on complex features that span multiple packages or require incremental delivery.

#### What it solves
- **Feature decomposition** - Break large changes into smaller, reviewable PRs
- **Dependency management** - Handle branches that depend on each other
- **Clean history** - Maintain logical commits while managing complex workflows
- **Parallel development** - Work on multiple related features simultaneously

#### Installation
```bash
# Available via our dep system
go run . dep install gs
```

#### Usage Patterns

**Basic stacked workflow:**
```bash
# Create feature stack
gs branch create feature/binary-build-base     # Core binary build logic
gs branch create feature/binary-meta           # meta.json generation
gs branch create feature/binary-cli            # CLI integration

# Submit entire stack
gs stack submit

# Navigate stack
gs log --stack
gs up/down
```

**Real-world examples:**
```bash
# Multi-package refactor
gs branch create refactor/config-loading
gs branch create feature/workflow-integration   # depends on config refactor
gs branch create feature/cli-enhancements       # depends on workflow

# Incremental feature delivery
gs branch create feature/darwin-arm64-support
gs branch create feature/windows-binaries       # depends on arm64 work
gs branch create feature/meta-generation        # depends on platform support
```

#### Key Commands
- `gs branch create <name>` - Create new branch in stack
- `gs stack submit` - Submit entire stack as separate PRs
- `gs stack restack` - Rebase all dependent branches
- `gs log --stack` - View stack structure
- `gs up/down` - Navigate between stacked branches

#### Integration Benefits
- **Idempotent workflows** - Each PR can be tested independently
- **Reduced review burden** - Smaller, focused changes for reviewers
- **Safe rebasing** - Automatic handling of branch dependencies
- **GitHub sync** - Automatic PR creation and updates

## Development Workflow

### Recommended Patterns

**For infrastructure changes:**
```bash
# Infrastructure refactor
gs branch create refactor/config-paths
gs branch create feature/platform-detection
gs branch create feature/binary-naming

# Testing improvements
gs branch create test/binary-build-tests
gs branch create test/cross-platform-validation
```

**For new features:**
```bash
# New tool integration
gs branch create feature/git-spice-support
gs branch create feature/ai-logging-integration
gs branch create feature/workflow-documentation
```

### Branch Naming Convention
- `feature/<component>-<description>` - New features
- `refactor/<area>-<description>` - Code restructuring
- `test/<component>-<description>` - Testing improvements
- `fix/<component>-<description>` - Bug fixes

## File Structure

```
pkg/git/
├── README.md           # This documentation
└── [Integration files] # Future git-specific utilities
```

## Quick Start

1. **Install git-spice**
   ```bash
   go run . dep install git-spice
   ```

2. **Initialize repository**
   ```bash
   gs repo init
   ```

3. **Create feature stack**
   ```bash
   gs branch create feature/your-change
   # Make changes, commit
   gs branch create feature/next-change
   # Continue development
   gs stack submit
   ```

## Best Practices

- **Keep stacks focused** - Related changes that can be reviewed independently
- **Use descriptive names** - Clear branch names help reviewers understand context
- **Test each branch** - Ensure each PR in the stack is independently functional
- **Document dependencies** - Use PR descriptions to explain the stack relationship

## Troubleshooting

**Common issues:**
- **Rebase conflicts**: Use `gs stack restack` to automatically handle dependencies
- **Lost track**: `gs log --stack` shows current stack structure
- **Need to reorder**: `gs branch onto` can move branches within the stack

**Getting help:**
```bash
gs help
```