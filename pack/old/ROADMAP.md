# Development and Release Roadmap

This document outlines the development, testing, and release strategy for this tool. The primary goal is to create a reliable, cross-platform CLI that is easily distributed to developers via npm.

Our core principle is automation via the `Taskfile.yml` to ensure all steps are repeatable and less prone to error.

---

## Phase 1: Local Development & Testing (Unsigned Binaries)

This is the day-to-day development loop. The focus is on fast iteration and ensuring the core logic of the Go program is working as expected within the Node.js wrapper environment.

**Testing "Without Signing":** Binaries in this phase are **not** code-signed. This is standard for local development as it avoids the overhead of the signing process for every small change.

### Workflow:

1.  **Code:** Make changes to the Go source code (`main.go`).
2.  **Build:** Run `task install`. This task will:
    *   Compile the `main.go` file into a platform-specific binary (e.g., `your-tool-darwin-arm64`).
    *   Place the binary in the `bin/` directory, making it ready for execution.
3.  **Test:** Run `task test`. This is the crucial local verification step. It will:
    *   Run `npm install --ignore-scripts` to set up the `node_modules` directory and create the necessary symlinks for `npx` and `bunx` to find the local executable.
    *   Execute the tool using the Node.js wrappers (`npx`, `bunx`) to confirm it runs correctly.
    *   Execute the tool with `deno` (if installed).

---

## Phase 2: Pre-Release Staging & Testing (Signed Binaries)

This phase is the final quality gate before a public release. Its purpose is to test the *actual user experience* from installation to execution, including code signing.

**Testing "With Signing":** Binaries in this phase **are** code-signed for macOS and Windows to prevent security warnings and build user trust.

### Workflow:

1.  **Sign Binaries:** A new `task sign` (to be created) will be run. This task will:
    *   First, run `task build` to compile the binaries for all target platforms (macOS, Windows, Linux).
    *   Then, use a code-signing utility (like `gon` for macOS) to sign the relevant binaries.
2.  **Create a Draft GitHub Release:**
    *   Create a new tag (e.g., `v1.0.1-beta`).
    *   Create a *draft* (private) release on GitHub associated with this tag.
    *   Upload the *signed* binaries to this draft release.
3.  **Test the "Real" Installation:** A new `task test-release` (to be created) will be run. This task will:
    *   Temporarily modify `package.json` to point to the draft GitHub release URL.
    *   Run `npm install` (with the `postinstall` script enabled).
    *   Verify that the correct signed binary is downloaded and runs successfully.
    *   This test will be performed on all target platforms (macOS, Windows, Linux) via CI/CD.

---

## Phase 3: Production Release

Once Phase 2 is successfully completed, the tool is ready for public release.

### Workflow:

1.  **Publish GitHub Release:** Convert the draft release from Phase 2 into a public release.
2.  **Update `package.json`:** Set the final version number (e.g., `1.0.1`).
3.  **Publish to npm:** Run `npm publish`.

---
