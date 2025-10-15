# Contributing Guidelines

## Documentation Maintenance

**IMPORTANT**: Keep documentation synchronized with code changes.

### When Adding a New Command

1. **Update the Makefile** ([Makefile](Makefile))
   - Add a new `.PHONY` target
   - Add the command to the `help` target output
   - Example:
     ```makefile
     .PHONY: new-command
     new-command:
         @go run . new-command
     ```

2. **Update README.md** ([README.md](README.md))
   - Add the command to the Quick Start section if it's commonly used
   - Add it to the Stack Commands section with description
   - Example:
     ```markdown
     ## Stack Commands

     - `go run . new-command` - Description of what it does
     ```

3. **Update DEVELOPMENT.md** ([docs/DEVELOPMENT.md](docs/DEVELOPMENT.md))
   - Add detailed usage examples
   - Document any flags or options
   - Include common use cases

### When Adding a New Service

1. **Create service directory** in `services/`
2. **Add service.json** with complete spec
3. **Update README.md** to list the new service
4. **Update docs/DEVELOPMENT.md** with integration steps
5. **Add Make target** if the service needs special build steps

### When Changing Environment Variables

1. **Update .env.example** with new variables and descriptions
2. **Update docs/TROUBLESHOOTING.md** with any new error cases
3. **Search codebase** for old variable names and update all references

### When Modifying Architecture

1. **Update README.md** Architecture section
2. **Update docs/DEVELOPMENT.md** with new patterns
3. **Consider adding architectural decision record** in `docs/adr/` (if we create that)

## Documentation Structure

```
core/
├── README.md              # Main entry point - Quick Start, Architecture, Commands
├── CONTRIBUTING.md        # This file - How to contribute
├── Makefile               # Quick commands - always keep help target updated
└── docs/
    ├── DEVELOPMENT.md     # Developer guide - detailed workflows
    └── TROUBLESHOOTING.md # Problem solving - common issues and fixes
```

## Documentation Rules

1. **All code changes MUST include documentation updates**
2. **Examples must be tested and working**
3. **Keep language simple and direct**
4. **Use relative links** for navigation between docs
5. **Update the help command** when adding CLI commands

## Testing Documentation

Before committing, verify:

```bash
# Test all Makefile targets
make help
make build
make clean

# Test all commands in README Quick Start
go run . stack up
go run . stack status
go run . stack down

# Verify links work (relative paths)
# Check all markdown files for broken links
```

## Git Workflow

```bash
# 1. Make your changes (code + docs)
git add .

# 2. Ensure docs are updated
# - Makefile help target?
# - README.md commands?
# - DEVELOPMENT.md examples?

# 3. Commit with clear message
git commit -m "feat: add new-command with full documentation"

# 4. Push
git push
```

## Questions?

If unsure about documentation placement:
- **Quick reference?** → README.md
- **Detailed how-to?** → docs/DEVELOPMENT.md
- **Error resolution?** → docs/TROUBLESHOOTING.md
- **Build command?** → Makefile
