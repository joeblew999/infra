# TODO: Upload/Download System

## Current State
Found the storage/ directory with upload/download functionality:

### Files
- `pkg/dep/storage/github.go` - GitHub Packages upload/download
- `pkg/dep/storage/nats.go` - NATS storage system

### GitHub Storage Features
- Upload binaries to GitHub Packages/Releases via GitHub CLI
- Download binaries from GitHub Releases
- Platform-specific asset naming (OS-arch)
- Automatic release creation
- GitHub CLI bootstrapping

## Questions for User
Please explain what you want me to work on with the upload/download system:

1. What specific improvements or features do you need?
2. Are there issues with the current implementation?
3. Should I integrate this with the existing dep system?
4. Do you want to expand/refactor the storage backends?
5. Is there missing functionality you need?

## Next Steps
Waiting for user clarification on the specific upload/download work needed.

do you have a thing that can check to see if there are new releases ? 

## ZIG

https://ziglang.org/download/ says there are darwin amr64 releases, but go run . dep upgrade says there are not.... might need debugging ? 

## Litestream

https://github.com/benbjohnson/litestream/releases/tag/v0.5.0-beta1 only has linux releases, so i asked for windows and mac too at https://github.com/benbjohnson/litestream/issues/716 so that we can test locally. Maybe we should change to a source based dep for now ? 
- they also have NATS Object Store integration, which is great for us. We need to get NATS running on fly properly in 6 regions.
- https://github.com/corylanou/litestream-test might be useful for us. I made an issue to add NATS testing too.

