# 🚀 Deployment Ready Checklist

**Status**: ✅ **READY FOR PRODUCTION DEPLOYMENT**

Last verified: 2025-10-16

---

## Current System Status

### Local Stack
- ✅ All services running (NATS, PocketBase, Caddy)
- ✅ 0 restarts on all services
- ✅ Health checks: All Ready
- ✅ Ports: 4222, 8090, 2015 all active
- ✅ Observability system operational

### Deployment Infrastructure
- ✅ Core CLI deploy command implemented
- ✅ Tooling system deployment workflow working
- ✅ Tool installation automated (`ensure` command)
- ✅ CI/CD pipeline configured
- ✅ Phase 2 microservices architecture designed

---

## Deployment Options

### Option 1: Manual Deployment (Fastest)

**Using Core CLI** (Recommended for first deployment):
```bash
# Install tools
go run . ensure all

# Set Fly.io token
export FLY_API_TOKEN=<your_token>

# Deploy to staging first
go run . deploy --app core-v2-staging --region syd

# Verify staging
curl https://core-v2-staging.fly.dev/api/health

# Deploy to production
go run . deploy --app core-v2 --region syd

# Verify production
curl https://core-v2.fly.dev/api/health
```

**Using Tooling System** (Interactive with Cloudflare):
```bash
cd tooling
./tooling workflow deploy --app core-v2 --verbose

# Follow prompts for Fly.io and Cloudflare auth
```

---

### Option 2: Automated CI/CD

**Setup** (One-time, 5 minutes):
```bash
# 1. Generate Fly.io deploy token
flyctl tokens create deploy

# 2. Add to GitHub secrets
#    GitHub repo → Settings → Secrets → Actions → New secret
#    Name: FLY_API_TOKEN
#    Value: <paste token>

# 3. (Optional) Configure production environment protection
#    GitHub repo → Settings → Environments → New environment
#    Name: production
#    Enable: Required reviewers, Wait timer, Deployment branches
```

**Deploy** (Automatic on push to main):
```bash
git push origin main

# GitHub Actions will:
# 1. Run all tests
# 2. Build binary
# 3. Run stack doctor
# 4. Deploy to Fly.io (if tests pass)
# 5. Verify health check
# 6. Notify status
```

**Monitor**:
- GitHub Actions: https://github.com/joeblew999/infra/actions
- Fly.io Dashboard: https://fly.io/dashboard
- Health endpoint: https://core-v2.fly.dev/api/health

---

## Pre-Deployment Checklist

### Required (Before First Deploy)

- [ ] **Fly.io account created** (https://fly.io/app/sign-up)
- [ ] **Fly.io CLI authenticated** (`flyctl auth login`)
- [ ] **Deployment token generated** (`flyctl tokens create deploy`)
- [ ] **GitHub secret added** (for CI/CD) or environment variable set (for manual)
- [ ] **App name decided** (default: `core-v2`)
- [ ] **Primary region selected** (default: `syd` Sydney)

### Recommended (Best Practices)

- [ ] **Deploy to staging first** (`core-v2-staging`)
- [ ] **Test staging deployment thoroughly**
- [ ] **Set up monitoring/alerting** (Fly.io dashboard, health checks)
- [ ] **Configure DNS** (if using custom domain)
- [ ] **Review fly.toml configuration** (adjust resources if needed)
- [ ] **Plan rollback strategy** (document steps)

### Optional (Nice to Have)

- [ ] **GitHub environment protection** (for approval workflows)
- [ ] **Slack/Discord notifications** (for deployment status)
- [ ] **Log aggregation** (Datadog, Grafana, etc.)
- [ ] **Cloudflare DNS** (for tooling system deployment)
- [ ] **Load testing results** (before production launch)

---

## Deployment Commands Reference

### Manual Deployment

```bash
# Dry run (test without deploying)
go run . deploy --dry-run

# Deploy to specific app and region
go run . deploy --app core-v2 --region syd

# Deploy to different environment
go run . deploy --app core-v2 --env production

# Using tooling system (interactive)
cd tooling && ./tooling workflow deploy --app core-v2
```

### CI/CD Deployment

```bash
# Trigger deployment (push to main)
git push origin main

# Manual trigger via GitHub Actions
# GitHub repo → Actions → Core CI/CD → Run workflow

# Check deployment status
gh run list --workflow=core-ci.yml
gh run view <run-id>
```

### Tool Installation

```bash
# Install all tools
go run . ensure all

# Install individually
go run . ensure ko
go run . ensure flyctl

# Force reinstall
go run . ensure all --force

# Check installed tools
.dep/ko version
.dep/flyctl version
```

---

## First Deployment Steps

### Step 1: Prepare (5 minutes)

```bash
# Verify local stack is healthy
go run . stack doctor

# Install deployment tools
go run . ensure all

# Authenticate with Fly.io
flyctl auth login

# Generate deployment token
flyctl tokens create deploy
# Save this token - you'll need it!
```

### Step 2: Create Staging App (2 minutes)

```bash
# Create Fly.io app
flyctl apps create core-v2-staging --org personal

# Create persistent volume
flyctl volumes create core_data --region syd --size 1 --app core-v2-staging
```

### Step 3: Deploy to Staging (5 minutes)

```bash
# Set token
export FLY_API_TOKEN=<your_token>

# Deploy
go run . deploy --app core-v2-staging --region syd

# Monitor deployment
flyctl logs --app core-v2-staging
```

### Step 4: Verify Staging (3 minutes)

```bash
# Check status
flyctl status --app core-v2-staging

# Test health endpoint
curl https://core-v2-staging.fly.dev/api/health

# Test PocketBase API
curl https://core-v2-staging.fly.dev/_/health

# Check NATS monitoring (via Caddy proxy)
# Configure Caddy to expose NATS monitoring if needed
```

### Step 5: Deploy to Production (5 minutes)

Once staging is verified:

```bash
# Create production app
flyctl apps create core-v2 --org personal

# Create production volume
flyctl volumes create core_data --region syd --size 1 --app core-v2

# Deploy
go run . deploy --app core-v2 --region syd

# Verify
curl https://core-v2.fly.dev/api/health
```

### Step 6: Configure CI/CD (5 minutes)

```bash
# Add GitHub secret
# GitHub repo → Settings → Secrets → Actions → New secret
# Name: FLY_API_TOKEN
# Value: <your deployment token>

# Test CI/CD
git commit --allow-empty -m "test: trigger CI/CD"
git push origin main

# Watch deployment
# GitHub repo → Actions tab
```

---

## Health Verification

After deployment, verify all systems are operational:

### Automated Checks

```bash
# Run full diagnostics remotely
flyctl ssh console --app core-v2 -C "curl http://localhost:2015/api/health"

# Check process-compose status
flyctl ssh console --app core-v2 -C "curl http://localhost:28081/processes"
```

### Manual Checks

1. **API Health**: `curl https://core-v2.fly.dev/api/health`
2. **PocketBase**: `curl https://core-v2.fly.dev/_/health`
3. **Application logs**: `flyctl logs --app core-v2`
4. **Metrics**: Check Fly.io dashboard for CPU, memory, requests

### Expected Responses

```json
// API Health (should return 200)
{"status": "ok"}

// PocketBase Health (should return 200)
{"message": "API is healthy."}
```

---

## Rollback Procedures

### If Deployment Fails

```bash
# View deployment history
flyctl releases --app core-v2

# Rollback to previous version
flyctl releases rollback <version> --app core-v2

# Verify rollback
curl https://core-v2.fly.dev/api/health
```

### If Health Checks Fail

```bash
# Check logs for errors
flyctl logs --app core-v2

# Restart app
flyctl apps restart core-v2

# If still failing, rollback (see above)
```

### Emergency: Revert Git Commit

```bash
# Revert problematic commit
git revert <commit-sha>
git push origin main

# CI/CD will auto-deploy the reverted version
```

---

## Monitoring Post-Deployment

### First 24 Hours

Monitor closely:
- ✅ Error rates (should be 0%)
- ✅ Response times (< 200ms for health checks)
- ✅ Memory usage (should be stable)
- ✅ Restart count (should be 0)

### Ongoing Monitoring

Set up alerts for:
- ❌ Health check failures
- ⚠️  High error rates (> 1%)
- ⚠️  Slow response times (> 2s)
- ⚠️  Memory growth (indicates leak)
- ⚠️  Frequent restarts (> 3 per hour)

---

## Cost Estimates

### Phase 1 (Current - Monolithic)

- **Compute**: ~$5-10/month (shared CPU, 1GB RAM)
- **Storage**: ~$0.15/month (1GB volume)
- **Bandwidth**: First 100GB free
- **Total**: ~$5-10/month for light-moderate usage

### Phase 2 (Future - Microservices)

- **Total**: ~$16-20/month (see `docs/MICROSERVICES_ARCHITECTURE.md`)
- **Benefits**: Independent scaling, better isolation
- **When to migrate**: When single instance becomes bottleneck

---

## Support and Documentation

### If Issues Arise

1. **Check documentation**:
   - `docs/DEPLOYMENT.md` - Deployment methods
   - `docs/TROUBLESHOOTING.md` - Common issues
   - `docs/CI_CD_SETUP.md` - CI/CD configuration

2. **Check logs**:
   - Local: `go run . stack status`
   - Remote: `flyctl logs --app core-v2`

3. **Run diagnostics**:
   - Local: `go run . stack doctor --verbose`
   - Remote: `flyctl ssh console --app core-v2`

4. **Community resources**:
   - Fly.io docs: https://fly.io/docs
   - Fly.io community: https://community.fly.io

---

## Next Steps After Deployment

### Immediate (Day 1)
- [ ] Verify all health checks passing
- [ ] Set up uptime monitoring (UptimeRobot, Pingdom, etc.)
- [ ] Configure DNS (if using custom domain)
- [ ] Document deployment in team wiki

### Short-term (Week 1)
- [ ] Monitor resource usage
- [ ] Adjust VM size if needed
- [ ] Set up log aggregation
- [ ] Configure backup strategy (Litestream to R2)

### Medium-term (Month 1)
- [ ] Analyze usage patterns
- [ ] Plan for Phase 2 (microservices) if needed
- [ ] Implement auto-scaling rules
- [ ] Set up comprehensive monitoring

---

## Deployment Success Criteria

Deployment is successful when:

- ✅ Health endpoint returns 200 OK
- ✅ PocketBase API is accessible
- ✅ No errors in logs for 1 hour
- ✅ Memory usage is stable (not growing)
- ✅ Response times < 500ms (p95)
- ✅ Zero restarts in first hour
- ✅ All services reporting Ready status

---

## Summary

**You are ready to deploy!** 🚀

Choose your deployment method:
- **Quick start**: Manual deployment with core CLI
- **Production setup**: CI/CD with GitHub Actions
- **Full featured**: Tooling system with Cloudflare DNS

All infrastructure is in place. Documentation is comprehensive. Code is tested and working locally.

**Recommendation**: Deploy to staging first, verify thoroughly, then promote to production.

---

*For detailed deployment walkthrough, see: `docs/DEPLOYMENT.md`*
*For CI/CD setup, see: `docs/CI_CD_SETUP.md`*
*For troubleshooting, see: `docs/TROUBLESHOOTING.md`*
