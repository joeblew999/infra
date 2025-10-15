# Tooling CLI

A unified, idempotent CLI for Fly.io + Cloudflare deployments.

This is really about making is easy for me ( from the CLI ), and end users via a Datastar GUI ( later ), to signup to fly and cloudflare and then for the system to get the tokens, so that the system can provisions things on FLY and CF.

This pattern can then later be extented to help me and the users to sign up to other Services and for our system to get the TOKENS. 

Its solves a tricky problem, where its hard for users to give us the tokens.

## What Makes This CLI Different

- **Idempotent**: Run auth commands anytime - they check existing credentials first
- **Clear Guidance**: Helpful messages explain what's happening and what to do next
- **Two-Path Cloudflare Auth**: Choose between manual token or automated bootstrap
- **Visual Feedback**: Uses ✓ and ⚠ symbols to show status clearly

## Quick Start

### 1. Build the CLI

```bash
go build -o core-tool .
```

### 2. Authenticate with Fly.io

```bash
./core-tool auth fly
```

This command will:
- ✓ Check if you already have valid credentials (safe to re-run!)
- ✓ Open your browser for Fly.io login if needed
- ✓ Verify the token works
- ✓ Save credentials for future use

**Note**: If you already have credentials, it just verifies them and exits.

### 3. Authenticate with Cloudflare

#### Option A: API Token (Recommended)

```bash
./core-tool auth cloudflare
```

You'll need to create a scoped API token at: https://dash.cloudflare.com/profile/api-tokens

Required permissions:
- Zone:DNS:Edit
- Zone:Zone:Read
- Account:R2:Edit (if using R2 storage)

#### Option B: Bootstrap (One-time Setup)

```bash
./core-tool auth cloudflare bootstrap
```

This uses your Global API Key once to create a scoped token automatically.
After bootstrap, use `./core-tool auth cloudflare` to verify.

### 4. Verify Your Setup

```bash
# Check both Fly and Cloudflare at once
./core-tool auth status

# Or verify individually
./core-tool auth fly verify
./core-tool auth cloudflare verify
```

### 5. Deploy

```bash
./core-tool workflow deploy \
  --app <fly-app-name> \
  --org <fly-org-slug> \
  --repo registry.fly.io/<fly-app-name>
```

## Authentication is Idempotent

You can safely run these commands anytime:

```bash
# These check existing credentials first and only prompt if needed
./core-tool auth fly
./core-tool auth cloudflare
```

If credentials are already valid, you'll see:
```
✓ Fly.io token is valid for user@example.com
✓ All checks passed - you're ready to deploy!
```

## Common Workflows

### First Time Setup

```bash
# 1. Build
go build -o core-tool .

# 2. Authenticate Fly
./core-tool auth fly

# 3. Authenticate Cloudflare (choose one method)
./core-tool auth cloudflare          # Token method
# OR
./core-tool auth cloudflare bootstrap # Bootstrap method

# 4. Check everything
./core-tool auth status

# 5. Deploy
./core-tool workflow deploy --app myapp --org myorg --repo registry.fly.io/myapp
```

### Re-authenticating

```bash
# Just run the auth commands again - they're idempotent!
./core-tool auth fly
./core-tool auth cloudflare
```

### Checking Credentials

```bash
# See all cached settings
./core-tool auth status

# Verify just Fly
./core-tool auth fly verify

# Verify just Cloudflare  
./core-tool auth cloudflare verify
```

## Advanced Usage

### Using Profiles

```bash
# Use a specific profile
./core-tool --profile production auth fly
./core-tool --profile staging workflow deploy --app staging-app
```

### Non-Interactive Mode

```bash
# Provide token directly (for CI/CD)
./core-tool auth fly --token $FLY_API_TOKEN
./core-tool auth cloudflare --token $CF_API_TOKEN

# Don't open browser
./core-tool auth fly --no-browser
./core-tool auth cloudflare --no-browser
```

### Override Token Paths

```bash
# Store tokens in custom locations
./core-tool auth fly --path /custom/path/fly-token.txt
./core-tool auth cloudflare --path /custom/path/cf-token.txt
```

## Command Reference

### Authentication Commands

| Command | Purpose | Idempotent? |
|---------|---------|-------------|
| `auth fly` | Authenticate with Fly.io | ✓ Yes |
| `auth fly verify` | Check Fly credentials | ✓ Yes |
| `auth cloudflare` | Authenticate with Cloudflare (token) | ✓ Yes |
| `auth cloudflare bootstrap` | Create Cloudflare token (one-time) | ✓ Yes |
| `auth cloudflare verify` | Check Cloudflare credentials | ✓ Yes |
| `auth status` | Show all cached settings | ✓ Yes |

### Deployment Commands

| Command | Purpose |
|---------|---------|
| `workflow deploy` | Build and deploy to Fly.io |

### Getting Help

```bash
# See all commands
./core-tool --help

# Help for specific command
./core-tool auth --help
./core-tool auth cloudflare --help
./core-tool auth fly --help
./core-tool workflow deploy --help
```

## Troubleshooting

### "Token invalid" message

Just run the auth command again - it will prompt for new credentials:
```bash
./core-tool auth fly
./core-tool auth cloudflare
```

### "No Fly app configured"

This is OK! It just means the app check was skipped. Specify the app name:
```bash
./core-tool workflow deploy --app myapp --org myorg
```

### "No Cloudflare zone configured"

This is OK on first run! The zone will be configured during your first deployment.

### Need to switch Cloudflare accounts?

Just re-run the auth command with a new token:
```bash
./core-tool auth cloudflare --token <new-token>
```

## Documentation

The canonical, always-up-to-date playbook lives in [docs/tooling.md](../docs/tooling.md).

## Development

```bash
# Run tests
go test ./...

# Build
go build -o core-tool .

# Format code
go fmt ./...
```

## Why This Design?

1. **Idempotent by default**: You can run auth commands anytime without breaking existing setups
2. **Clear feedback**: Visual indicators (✓ ⚠) make status obvious
3. **Helpful guidance**: Commands explain what they do and what you need
4. **Two Cloudflare paths**: Flexibility for different security requirements
5. **Verify commands**: Check credentials without changing anything
6. **Profile support**: Manage multiple environments easily
