# Demo Service: JSON

This demo proves the JSON-driven service onboarding flow. The manifest lives in
`service.json` and is loaded by the Go code to ensure binaries (via shared dep),
configure the process runner, and register with the controller.
