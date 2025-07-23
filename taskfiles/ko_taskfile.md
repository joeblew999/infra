# ko

https://ko.build

https://github.com/ko-build/ko

https://github.com/ko-build/ko/releases/tag/v0.18.0

Ko allows building docker images without any need for docker.

## Goal

Build and publish Go binaries as OCI/Docker images using [ko](https://github.com/ko-build/ko), with minimal config and automation, ready for use with Cloudflare R2 or any registry.


## Steps

1. **Install ko**
   - Use the Taskfile: `task ko-bin`
   - Installs the correct version to `.bin/ko`

2. **Set KO_DOCKER_REPO**
   - This is the Docker registry/repo prefix for your images (e.g. `docker.io/youruser/yourrepo`)
   - Can be set as an environment variable, Taskfile var, or CLI override

3. **Build & Publish**
   - Use the Taskfile: `task ko-build`
   - This will build and publish all Go binaries in the project as images to your KO_DOCKER_REPO

4. **Go Project Structure**
   - Each Go binary should have its own `main.go` and `go.mod` in its directory (e.g. `cmd/nats-secrets/`)
   - ko will build from these entrypoints

5. **Cloudflare R2**
   - Use as a registry backend or for storing artifacts
   - (See Cloudflare docs for setup)

## Example

```sh
export KO_DOCKER_REPO=docker.io/youruser/yourrepo
# or use the default in Taskfile

# Build and publish all images
./.bin/ko build ./cmd/your-binary
```

## Tips
- No Dockerfile needed for simple Go binaries
- ko auto-detects Go entrypoints
- Use Taskfile for repeatable, cross-platform automation
- For multi-arch, see ko docs

---

_This approach: fast, simple, no Dockerfile, works with any OCI registry (including R2)._
