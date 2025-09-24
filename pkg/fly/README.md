# Fly.io Integration

This package provides commands for managing Fly.io deployments and infrastructure.

## Usage

### CLI Commands

Add Fly.io commands to your CLI:

```go
import "github.com/joeblew999/infra/pkg/fly"

// In your root command setup
fly.AddCommands(rootCmd)
```

### Available Commands

- `go run . fly deploy` - Deploy to Fly.io
- `go run . fly status` - Check app status  
- `go run . fly logs` - Show logs
- `go run . fly ssh` - SSH into machine
- `go run . fly scale` - Scale resources

### Configuration

Ensure you have:
1. `fly.toml` configured for your app
2. `flyctl` installed (managed by pkg/dep)
3. Fly.io API token set via `FLY_API_TOKEN`

### Deployment Process

1. **Setup**: `go run . tools dep install flyctl`
2. **Deploy**: `go run . fly deploy`
3. **Monitor**: `go run . fly status`
4. **Debug**: `go run . fly logs` or `go run . fly ssh`

## Integration with Litestream

For Fly.io deployments with database backups:

1. Deploy with volume: `fly deploy`
2. Configure litestream: `go run . litestream start`
3. Monitor backups: `go run . litestream status`

## Environment Variables

- `FLY_API_TOKEN` - Fly.io API token
- `FLY_APP_NAME` - App name (defaults to 'infra-mgmt')
- `FLY_REGION` - Deployment region

## Storage

We use Tigis and wire Backblaze B2 (or any other S3-compatible store) as a shadow bucket for a Tigris bucket on Fly.io.

Below is the quickest, copy-paste-ready path to wire Backblaze B2 (or any other S3-compatible store) as a shadow bucket for a Tigris bucket on Fly.io.
────────────────────────────────────────
Gather the B2 credentials
• In Backblaze → Buckets → (your bucket) → App Keys
– KeyID  → B2_KEY_ID
– Secret → B2_APPLICATION_KEY
• Endpoint
– For most regions: https://s3.us-east-005.backblazeb2.com
(use the region-specific URL if your bucket is in another region)
Decide where the bucket lives
• New bucket – create it with shadow settings in one CLI call
• Existing bucket – add the shadow later
Create a new bucket with shadow (one-liner)
bash

Copy
flyctl storage create \
  -n my-cold-archive \
  --shadow-access-key  B2_KEY_ID \
  --shadow-secret-key  B2_APPLICATION_KEY \
  --shadow-endpoint    https://s3.us-east-005.backblazeb2.com \
  --shadow-region      us-east-005 \
  --shadow-write-through   # keeps B2 in sync for every upload
That’s it—Tigris will lazily pull objects from B2 on first request and cache them globally.
Add (or remove) a shadow on an existing bucket
bash

Copy
# add
flyctl storage update my-existing-bucket \
  --shadow-access-key  B2_KEY_ID \
  --shadow-secret-key  B2_APPLICATION_KEY \
  --shadow-endpoint    https://s3.us-east-005.backblazeb2.com \
  --shadow-region      us-east-005 \
  --shadow-write-through

# remove
flyctl storage update my-existing-bucket --clear-shadow
