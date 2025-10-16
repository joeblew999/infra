# CI/CD Setup Guide

## Overview

The core V2 repository includes a comprehensive CI/CD pipeline using GitHub Actions that automatically tests, builds, and deploys the application to Fly.io on every push to main.

## Workflow File

Location: `.github/workflows/core-ci.yml`

## Pipeline Stages

### 1. Test Job (Always Runs)

**Triggers**: Every push and pull request affecting `core/**`

**Steps**:
1. Checkout code
2. Setup Go (stable version)
3. Run tests (`SKIP_INTEGRATION_TESTS=1`)
4. Build core binary
5. Run `stack doctor` diagnostics

**Purpose**: Ensure code quality before deployment

---

### 2. Deploy Job (Main Branch Only)

**Triggers**: Only on push to `main` branch after tests pass

**Steps**:
1. Checkout code
2. Setup Go
3. Deploy to Fly.io using core CLI
4. Verify deployment health
5. Notify deployment status

**Purpose**: Automated production deployment

---

## GitHub Repository Setup

### Required Secrets

Add these secrets in GitHub repository settings (`Settings` → `Secrets and variables` → `Actions`):

#### 1. FLY_API_TOKEN

**Purpose**: Authenticate with Fly.io for deployments

**How to generate**:
```bash
# Login to Fly.io
flyctl auth login

# Generate deployment token
flyctl tokens create deploy

# Copy the token and add to GitHub secrets
```

**Add to GitHub**:
1. Go to repository Settings → Secrets → Actions
2. Click "New repository secret"
3. Name: `FLY_API_TOKEN`
4. Value: `<paste token from flyctl tokens create>`
5. Click "Add secret"

---

### Optional: Deployment Environments

For better control and approval workflows, configure a production environment:

**Setup Steps**:
1. Go to repository `Settings` → `Environments`
2. Click "New environment"
3. Name: `production`
4. Configure protection rules:
   - ✅ Required reviewers (select team members)
   - ✅ Wait timer (e.g., 5 minutes)
   - ✅ Deployment branches (only `main`)
5. Save

**Benefits**:
- Manual approval before production deploys
- Environment-specific secrets
- Deployment history per environment
- Wait timers for gradual rollouts

---

## Deployment Process

### Automatic Deployment (Main Branch)

When you push to `main`:

1. **Tests run** automatically
2. **If tests pass**, deployment job starts
3. **Deploy to Fly.io** using `go run ./cmd/core deploy`
4. **Health check** verifies deployment
5. **GitHub shows status** ✅ or ❌

### Manual Deployment Trigger

You can also trigger deployment manually:

```yaml
# Add to workflow file
on:
  workflow_dispatch:  # Enables manual trigger
```

Then trigger via:
- GitHub Actions tab → Select workflow → "Run workflow"

---

## Monitoring Deployments

### GitHub Actions Interface

View deployment status:
1. Go to repository → Actions tab
2. Select workflow run
3. View logs for test and deploy jobs

### Fly.io Dashboard

Monitor deployed app:
1. Visit https://fly.io/dashboard
2. Select `core-v2` app
3. View metrics, logs, and status

### Health Endpoint

Verify deployment health:
```bash
curl https://core-v2.fly.dev/api/health
```

---

## Troubleshooting

### Deployment Fails: "FLY_API_TOKEN not found"

**Problem**: Secret not configured

**Solution**:
```bash
# Generate new token
flyctl tokens create deploy

# Add to GitHub secrets (Settings → Secrets → Actions)
```

### Deployment Succeeds but Health Check Fails

**Problem**: App deployed but not responding

**Solution**:
```bash
# Check Fly.io logs
flyctl logs --app core-v2

# Check app status
flyctl status --app core-v2

# Restart if needed
flyctl apps restart core-v2
```

### Tests Pass Locally but Fail in CI

**Problem**: Environment differences

**Common causes**:
- Integration tests not skipped (`SKIP_INTEGRATION_TESTS` not set)
- Dependencies missing in CI
- Go version mismatch

**Solution**:
```bash
# Test with exact CI environment
docker run --rm -v $(pwd):/app -w /app golang:stable go test ./...

# Check CI logs for specific errors
# Update test configuration as needed
```

### Deployment Takes Too Long

**Problem**: Ko build is slow in CI

**Solution**: Add caching to workflow
```yaml
- name: Cache Go modules
  uses: actions/cache@v4
  with:
    path: |
      ~/.cache/go-build
      ~/go/pkg/mod
    key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
```

---

## Advanced Configuration

### Multi-Environment Deployment

Deploy to staging before production:

```yaml
jobs:
  deploy-staging:
    if: github.ref == 'refs/heads/develop'
    steps:
      - name: Deploy to Staging
        run: go run ./cmd/core deploy --app core-staging --env staging

  deploy-production:
    if: github.ref == 'refs/heads/main'
    needs: test
    steps:
      - name: Deploy to Production
        run: go run ./cmd/core deploy --app core-v2 --env production
```

### Deployment Notifications

Send Slack notifications on deployment:

```yaml
- name: Notify Slack
  if: always()
  uses: slackapi/slack-github-action@v1
  with:
    payload: |
      {
        "text": "Deployment ${{ job.status }}: core-v2",
        "blocks": [
          {
            "type": "section",
            "text": {
              "type": "mrkdwn",
              "text": "*Deployment Status:* ${{ job.status }}\n*App:* core-v2\n*Commit:* ${{ github.sha }}"
            }
          }
        ]
      }
  env:
    SLACK_WEBHOOK_URL: ${{ secrets.SLACK_WEBHOOK_URL }}
```

### Rollback on Failure

Automatically rollback failed deployments:

```yaml
- name: Deploy to Fly.io
  id: deploy
  run: go run ./cmd/core deploy --app core-v2

- name: Verify Deployment
  id: verify
  run: |
    curl -f https://core-v2.fly.dev/api/health

- name: Rollback on Failure
  if: failure() && steps.deploy.outcome == 'success'
  run: |
    flyctl releases rollback --app core-v2
    echo "⚠️ Deployment failed health check - rolled back"
```

### Gradual Rollout

Deploy to subset of regions first:

```yaml
- name: Deploy to Primary Region
  run: go run ./cmd/core deploy --app core-v2 --region syd

- name: Wait and Monitor
  run: |
    sleep 300  # Monitor for 5 minutes
    # Check error rates, metrics, etc.

- name: Deploy to All Regions
  if: success()
  run: |
    flyctl deploy --app core-v2 --strategy rolling
```

---

## Phase 2: Microservices CI/CD

For Phase 2 microservices deployment:

```yaml
jobs:
  deploy-microservices:
    if: github.ref == 'refs/heads/main'
    needs: test
    strategy:
      matrix:
        service: [nats, pocketbase, controller, caddy]
    steps:
      - name: Deploy ${{ matrix.service }}
        run: |
          cd deploy/phase2
          flyctl deploy \
            --config fly-${{ matrix.service }}.toml \
            --app core-${{ matrix.service }}
```

---

## Security Best Practices

### 1. Token Rotation

Rotate Fly.io tokens regularly:
```bash
# Revoke old token
flyctl tokens list
flyctl tokens revoke <token-id>

# Create new token
flyctl tokens create deploy

# Update GitHub secret
```

### 2. Least Privilege

Use deployment-specific tokens (not personal tokens):
```bash
# Create token with limited scope
flyctl tokens create deploy --expiry 90d
```

### 3. Secret Scanning

GitHub automatically scans for exposed secrets. If leaked:
1. Immediately revoke the token in Fly.io
2. Generate new token
3. Update GitHub secret
4. Investigate how it was exposed

### 4. Environment Protection

Configure environment protection rules:
- Require reviewers for production
- Restrict deployment branches
- Add wait timers
- Limit who can approve

---

## Monitoring and Alerts

### GitHub Actions Notifications

Enable email notifications:
1. GitHub Settings → Notifications
2. Enable "Actions" notifications
3. Choose email or web notifications

### Fly.io Metrics

Monitor deployed app:
- CPU usage
- Memory usage
- Request rates
- Error rates
- Response times

### Custom Health Checks

Add application-specific health checks:
```yaml
- name: Advanced Health Check
  run: |
    # Check specific endpoints
    curl -f https://core-v2.fly.dev/api/health
    curl -f https://core-v2.fly.dev/_/health

    # Check database connectivity
    # Check NATS connectivity
    # etc.
```

---

## Rollback Procedures

### Automatic Rollback

Already configured in workflow (rolls back on health check failure)

### Manual Rollback

If you need to manually rollback:

```bash
# View deployment history
flyctl releases --app core-v2

# Rollback to specific version
flyctl releases rollback v23 --app core-v2

# Verify rollback
curl https://core-v2.fly.dev/api/health
```

### Rollback from GitHub

Revert the problematic commit:
```bash
git revert <commit-sha>
git push origin main

# CI/CD will automatically deploy the reverted version
```

---

## Testing CI/CD Pipeline

### Test Locally

Simulate CI environment:
```bash
# Run tests as CI does
export SKIP_INTEGRATION_TESTS=1
go test ./...

# Build as CI does
go build -o /tmp/core ./cmd/core

# Test deploy command (dry-run)
go run ./cmd/core deploy --dry-run
```

### Test in Pull Request

Create a PR to test CI without deploying:
```bash
git checkout -b test-ci
git push origin test-ci

# Open PR in GitHub - tests run but deploy doesn't (only on main)
```

---

## Next Steps

1. **Add FLY_API_TOKEN secret** to GitHub repository
2. **Configure production environment** with protection rules
3. **Test deployment** by pushing to main
4. **Monitor first deployment** in GitHub Actions and Fly.io
5. **Set up notifications** (Slack, email, etc.)
6. **Document team runbooks** for handling failures

For detailed deployment documentation, see:
- [DEPLOYMENT.md](DEPLOYMENT.md) - Deployment methods and configuration
- [DEPLOYMENT_STATUS.md](DEPLOYMENT_STATUS.md) - Testing results
- [TROUBLESHOOTING.md](TROUBLESHOOTING.md) - Common issues and solutions
