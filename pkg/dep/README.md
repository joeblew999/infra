# dep

downloads binaries to .dep, ensuring the version is correct.

core binaries are used by the system to do core things.

secondary binaries can be setup using the json also. this is more for extensions later that developers may needs.




### Dep

Uses Design by Contract principles to ensure we do not break other code.

This Downloads binaries from github and records their version for idempotency.  These binaries are used by the systems core functionality, and stored locally.

The system automatically selects the correct asset based on the runtime platform and matches the filename using regex patterns.

Because each github repository releases their binaries in different forms, we have a golang file for each that is aware of the specifics of each.

The getGitHubReleaseDebug function  calls the gh CLI tool to get GitHub release information so that a developer or AI agent can get this information. 

### bento

For workflows at runtime.
https://github.com/warpstreamlabs/bento
https://github.com/warpstreamlabs/bento/releases/tag/v1.9.0

---

NOTE: For later running conduit look at https://github.com/warpstreamlabs/bento/discussions/396

- https://github.com/gregfurman/bento/tree/add/conduit/internal/impl/conduit

---

NOTE: For later running WASM, we need a golang example for it to run.


### caddy

Web server with automatic HTTPS.
https://github.com/caddyserver/caddy
https://github.com/caddyserver/caddy/releases/tag/v2.10.0

### flyctl

Fly.io CLI tool.
https://github.com/superfly/flyctl
https://github.com/superfly/flyctl/releases/tag/v0.3.159

### garble

https://github.com/burrowers/garble
https://github.com/burrowers/garble/releases/tag/v0.14.2

### ko

Container image builder for Go.
https://github.com/ko-build/ko
https://github.com/ko-build/ko/releases/tag/v0.18.0

### task

Task runner / build tool.
https://github.com/go-task/task
https://github.com/go-task/task/releases/tag/v3.44.1

### tofu

OpenTofu infrastructure as code.
https://github.com/opentofu/opentofu
https://github.com/opentofu/opentofu/releases/tag/v1.7.2


