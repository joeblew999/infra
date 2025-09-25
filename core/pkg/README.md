# Core Packages

Every Go package for the deterministic platform lives beneath this directory. The structure mirrors the architecture in Task 015:

- `runtime/` — runtime glue that boots the orchestrator (`cmd`, controller, process runner, bus, UI shells, etc.).
- `shared/` — reusable implementations consumed by runtime code and demo services.

When adding a new capability, build the shared package first, then wire it into the runtime tree. This keeps the dependency direction clean and makes it obvious when shared/runtime drift.
