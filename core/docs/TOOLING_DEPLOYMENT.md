# Tooling System Deployment Details

## Overview

The tooling system (`./core/tooling`) provides a comprehensive deployment workflow with automatic verification and DNS management. Here's how it works and what domains it uses.

---

## Current Configuration

### Fly.io Settings
**Location**: `.data/core/fly/settings.json`

```json
{
  "org_slug": "personal",
  "region_code": "syd",
  "region_name": "Sydney, Australia"
}
```

- **Organization**: `personal` (your Fly.io org)
- **Primary Region**: `syd` (Sydney, Australia)
- **App Name**: Defaults to profile setting or can be overridden with `--app` flag

### Cloudflare Settings
**Location**: `.data/core/cloudflare/settings.json`

```json
{
  "zone_name": "amplify-cms.com",
  "zone_id": "8df8800d531257e33bca551186c0df44",
  "account_id": "7384af54e33b8a54ff240371ea368440",
  "r2_bucket": "test",
  "app_domain": "1.amplify-cms.com"
}
```

- **Domain**: `amplify-cms.com`
- **Subdomain**: `1.amplify-cms.com` (configured as app_domain)
- **Zone ID**: `8df8800d531257e33bca551186c0df44`
- **R2 Bucket**: `test` (for backups/assets)

---

## Deployment Workflow

### Step 1: Authentication

The tooling verifies credentials before deployment:

```go
// tooling/pkg/fly/verify.go
func VerifyFlyToken(ctx context.Context, token string) (string, *Client, error) {
    client := flyapi.NewClientFromOptions(...)

    // Verify token by getting current user
    user, err := client.GetCurrentUser(ctx)

    // Returns user identity (email or name)
    identity = fmt.Sprintf("Fly user %s", user.Email)

    return identity, client, nil
}
```

**What it checks**:
- ✅ Token is valid
- ✅ Can authenticate with Fly.io API
- ✅ User identity (email: gedw99@gmail.com)

### Step 2: Configuration Generation

**Location**: `tooling/pkg/configinit/`

Generates two configuration files:

1. **`.ko.yaml`** - Ko builder configuration
   ```yaml
   defaultBaseImage: cgr.dev/chainguard/static:latest
   builds:
   - id: core
     dir: ./cmd/core
   ```

2. **`fly.toml`** - Fly.io deployment configuration
   - Uses template from profile or default
   - Injects app name, region, org
   - Sets up volumes, services, VM resources

### Step 3: Build & Push

**Location**: `tooling/pkg/release/pipeline.go`

1. **Build with Ko**:
   ```bash
   ko build --bare --platform=linux/amd64,linux/arm64 ./cmd/core
   ```
   - Builds multi-platform container images
   - Pushes to `registry.fly.io/<app-name>`
   - Returns image reference (SHA)

2. **Deploy to Fly.io**:
   ```bash
   flyctl deploy --app <app-name> --image <image-ref>
   ```

### Step 4: Verification

**Location**: `tooling/pkg/release/pipeline.go:320-340`

After deployment, the tooling verifies the app is running:

```go
func verifyApp(ctx context.Context, cl *flyapi.Client, appName string) error {
    deadline := time.Now().Add(60 * time.Second)

    for {
        // Try to get app info from Fly.io API
        app, err := cl.GetApp(ctx, appName)

        if err == nil && app != nil {
            // App exists and is accessible
            time.Sleep(5 * time.Second)  // Wait for stability
            return nil
        }

        if time.Now().After(deadline) {
            return fmt.Errorf("timed out waiting for app")
        }

        time.Sleep(2 * time.Second)  // Retry
    }
}
```

**Verification checks**:
- ✅ App exists in Fly.io
- ✅ App is accessible via API
- ✅ Waits 5 seconds for stability
- ✅ Timeout: 60 seconds

**Note**: This only verifies the app exists in Fly.io, not that health checks are passing. For health verification, you need to:

```bash
# Check app status
flyctl status --app <app-name>

# Check health endpoint
curl https://<app-name>.fly.dev/api/health
```

---

## DNS Management (Cloudflare)

### Automatic DNS Setup

When deploying with the tooling system, it can optionally configure Cloudflare DNS:

**Location**: `tooling/pkg/cloudflare/dns.go`

1. **Authenticates with Cloudflare**:
   - Uses stored token from `.data/core/cloudflare/`
   - Verifies zone access

2. **Creates/Updates DNS Records**:
   ```
   Type: CNAME
   Name: 1.amplify-cms.com
   Target: <app-name>.fly.dev
   Proxied: Yes (Orange cloud)
   ```

3. **Result**:
   - `https://1.amplify-cms.com` → points to Fly.io app
   - Cloudflare proxies traffic (DDoS protection, CDN)
   - TLS certificate managed by Cloudflare

### Domain Configuration

**Current Setup**:
- **Fly.io app URL**: `https://<app-name>.fly.dev`
- **Custom domain**: `https://1.amplify-cms.com`
- **Base domain**: `amplify-cms.com`

**To change the subdomain**:

Edit `.data/core/cloudflare/settings.json`:
```json
{
  "app_domain": "core.amplify-cms.com"  // Change this
}
```

Or use different domain entirely:
```json
{
  "zone_name": "your-domain.com",
  "app_domain": "core.your-domain.com"
}
```

---

## Deployment URLs

Based on your current configuration:

### Phase 1 (Monolithic)

**Fly.io Direct**:
- `https://core-v2.fly.dev` (default app name)
- Or `https://<your-app-name>.fly.dev`

**Cloudflare Custom Domain** (if configured):
- `https://1.amplify-cms.com` (current setting)
- Points to Fly.io app via CNAME

### Phase 2 (Microservices)

Each service gets its own Fly.io app:

**Fly.io URLs**:
- `https://core-caddy.fly.dev` (edge proxy)
- `https://core-nats.fly.dev` (NATS - internal only)
- `https://core-pocketbase.fly.dev` (PocketBase - internal only)
- `https://core-controller.fly.dev` (controller - internal only)

**Cloudflare Custom Domains**:
- `https://core.amplify-cms.com` → `core-caddy.fly.dev` (main)
- `https://nats.amplify-cms.com` → `core-nats.fly.dev` (monitoring)
- `https://api.amplify-cms.com` → `core-pocketbase.fly.dev` (API)

---

## Verification Steps

### After Deployment

The tooling performs these verifications automatically:

1. **Fly.io App Exists** ✅
   ```go
   app, err := client.GetApp(ctx, appName)
   // Verifies app is created and accessible
   ```

2. **Image Deployed** ✅
   ```
   Release ID: v1, v2, v3...
   Image: registry.fly.io/<app>/core@sha256:...
   ```

3. **App Running** (manual check):
   ```bash
   flyctl status --app core-v2
   ```

4. **Health Endpoint** (manual check):
   ```bash
   curl https://core-v2.fly.dev/api/health
   ```

5. **DNS Propagation** (if using Cloudflare):
   ```bash
   dig 1.amplify-cms.com
   # Should show CNAME to core-v2.fly.dev
   ```

---

## Testing Deployment

### Dry Run (Core CLI)

```bash
go run . deploy --dry-run --app core-v2-test
```

Shows what would happen without actually deploying.

### Full Deployment (Tooling)

```bash
cd tooling
./tooling workflow deploy --app core-v2-test --verbose
```

**What happens**:
1. Prompts for Fly.io authentication (opens browser)
2. Prompts for Cloudflare authentication (opens browser)
3. Generates `.ko.yaml` and `fly.toml`
4. Builds container with ko
5. Pushes to Fly.io registry
6. Deploys to Fly.io
7. **Verifies app exists** (60s timeout)
8. (Optional) Configures Cloudflare DNS
9. Returns deployment summary

### Manual Verification After Deployment

```bash
# 1. Check Fly.io app status
flyctl status --app core-v2-test

# 2. Check health endpoint
curl -i https://core-v2-test.fly.dev/api/health

# 3. Check logs
flyctl logs --app core-v2-test

# 4. Check DNS (if using Cloudflare)
dig 1.amplify-cms.com

# 5. Test custom domain
curl -i https://1.amplify-cms.com/api/health
```

---

## Troubleshooting

### Verification Fails: "Timed out waiting for app"

**Cause**: App deployment is slow or failed

**Fix**:
```bash
# Check deployment status
flyctl status --app <app-name>

# Check logs for errors
flyctl logs --app <app-name>

# Check if volume is mounted
flyctl volumes list --app <app-name>

# Restart if needed
flyctl apps restart <app-name>
```

### DNS Not Resolving

**Cause**: Cloudflare DNS not configured or propagation delay

**Fix**:
```bash
# Check DNS records in Cloudflare
# Visit: https://dash.cloudflare.com/

# Or use API to check
curl -X GET "https://api.cloudflare.com/client/v4/zones/<zone-id>/dns_records" \
  -H "Authorization: Bearer <token>"

# Wait 1-5 minutes for propagation
```

### Health Check Returns 502

**Cause**: App deployed but not healthy

**Fix**:
```bash
# Check if services are running
flyctl ssh console --app <app-name>
curl http://localhost:2015/api/health

# Check process-compose status
curl http://localhost:28081/processes

# Check individual services
curl http://localhost:4222/
curl http://localhost:8090/api/health
```

---

## Configuration Reference

### Environment Variables (Fly.io)

Set in `fly.toml`:
```toml
[env]
  ENVIRONMENT = 'production'
  CORE_NATS_PORT = '4222'
  CORE_POCKETBASE_PORT = '8090'
  CORE_CADDY_PORT = '2015'
```

### Secrets (Fly.io)

Set via flyctl:
```bash
flyctl secrets set ADMIN_PASSWORD=<password> --app core-v2
flyctl secrets set DATABASE_URL=<url> --app core-v2
```

### DNS Records (Cloudflare)

Current configuration:
```
Type: CNAME
Name: 1
Value: core-v2.fly.dev
Proxied: Yes
TTL: Auto
```

---

## Summary

### How Tooling Verifies Deployment

1. **Before deployment**: Verifies Fly.io token by calling `GetCurrentUser()`
2. **During deployment**: Builds and pushes image, deploys to Fly.io
3. **After deployment**: Polls `GetApp()` API for 60 seconds until app is accessible
4. **Returns**: Release ID, image reference, elapsed time

**Note**: Tooling does NOT check health endpoints. You must verify manually:
```bash
curl https://<app-name>.fly.dev/api/health
```

### Domains Used

**Fly.io**:
- Default: `https://<app-name>.fly.dev`
- Example: `https://core-v2.fly.dev`

**Cloudflare** (optional):
- Current: `https://1.amplify-cms.com`
- Configurable in `.data/core/cloudflare/settings.json`

**Registry**:
- `registry.fly.io/<app-name>` (container images)

---

## Next Steps

1. **Review domain configuration**:
   ```bash
   cat .data/core/cloudflare/settings.json
   ```

2. **Decide on custom domain** or use Fly.io default

3. **Test deployment to staging**:
   ```bash
   cd tooling
   ./tooling workflow deploy --app core-v2-staging
   ```

4. **Verify health manually**:
   ```bash
   curl https://core-v2-staging.fly.dev/api/health
   ```

5. **Deploy to production** when ready

---

*For complete deployment guide, see: `docs/DEPLOYMENT_READY.md`*
