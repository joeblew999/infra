✅ COMPLETED: Replaced kosho (carlsverre/kosho) with git-spice (abhinav/git-spice) in deps.

- Updated dep.json with gs binary (v0.16.1) from abhinav/git-spice
- Binary installs as `gs` command (shorter alias for git-spice)
- Installation verified: `.dep/gs --version` shows git-spice 0.16.1
- Documentation updated to use `go run . tools dep install gs`

https://github.com/abhinav/git-spice/blob/main/internal/forge/shamhub/README.md provides GitHub-independent DevOps workflows that work well with Soft Serve.