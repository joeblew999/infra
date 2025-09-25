# Demo Services

The demo services (`alpha`, `beta`) exercise the orchestrator end-to-end. Each service will eventually include:

- `config/`, `dep/`, `log/` wrappers built on shared modules.
- `process/` definitions that describe how to run the binary via the new process runner.
- `cli/`, `ui/` adapters so toggles and status views stay consistent across interfaces.

These services are the canonical examples for teams migrating onto the new platform.


also the json demo is designed to show a service using the json to wire up...