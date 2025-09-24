# Litestream

SQLite replication and disaster recovery for stateless deployments.

## Overview

Litestream provides **continuous streaming replication** for SQLite databases, enabling near-zero data loss and fast recovery. This allows our infrastructure to achieve **stateless deployment** patterns while maintaining data durability.

## Core Benefits

### 1. Continuous Backup
- **Real-time streaming** to S3, filesystem, or other storage
- **Point-in-time recovery** with second-level granularity
- **No maintenance windows** - backups happen transparently

### 2. Stateless Deployment
- **Automatic restoration** on boot if database is missing
- **Ephemeral storage** compatible - data survives container restarts
- **Fast startup** - restore from S3 takes seconds, not minutes

### 3. Read Replicas
- **Lightweight read replicas** for scaling read operations
- **Cross-region replication** for global deployments
- **Zero-downtime failover** capabilities

## Usage Patterns

### Local Development (No S3 Required)
```yaml
# litestream.yml - Works without any cloud setup
dbs:
  - path: ./pb_data/data.db
    replicas:
      - type: file
        path: ./backups/pb_data.db
```

### Local Multi-Backup Strategy
```yaml
# Multiple local backups for safety
dbs:
  - path: ./pb_data/data.db
    replicas:
      - type: file
        path: ./backups/latest.db
      - type: file
        path: ./backups/daily/$(date +%Y-%m-%d).db
        retention: 7d
```

### Production (S3)
```yaml
# When you need cloud backup
dbs:
  - path: /app/pb_data/data.db
    replicas:
      - type: s3
        bucket: myapp-backups
        path: litestream/pb_data.db
        region: us-east-1
        access-key-id: ${AWS_ACCESS_KEY_ID}
        secret-access-key: ${AWS_SECRET_ACCESS_KEY}
```

## Integration Points

### PocketBase Integration
```bash
# Start Litestream with PocketBase
litestream replicate -config ./pkg/litestream/litestream.yml

# PocketBase will use the replicated database
./pocketbase serve --dir ./pb_data
```

### Docker Integration
```dockerfile
# Dockerfile addition
COPY pkg/litestream/litestream.yml /etc/litestream.yml
CMD ["litestream", "replicate", "-config", "/etc/litestream.yml"]
```

### Service Mode
```bash
# Start as background service
litestream replicate -config ./pkg/litestream/litestream.yml &

# Or use systemd service
systemctl start litestream
```

## Configuration Templates

### Production (Fly.io)
```yaml
dbs:
  - path: /app/pb_data/data.db
    replicas:
      - type: s3
        bucket: ${FLY_APP_NAME}-backups
        path: litestream/pb_data.db
        region: ${AWS_REGION:-us-east-1}
        access-key-id: ${AWS_ACCESS_KEY_ID}
        secret-access-key: ${AWS_SECRET_ACCESS_KEY}
        retention: 30d
        retention-check-interval: 1h
```

### Development
```yaml
dbs:
  - path: ./pb_data/data.db
    replicas:
      - type: file
        path: ./backups/$(date +%Y-%m-%d)/pb_data.db
        retention: 7d
```

### Testing
```yaml
dbs:
  - path: ./test_data/data.db
    replicas:
      - type: file
        path: /tmp/test_backup.db
        sync-interval: 1s  # Fast sync for testing
```

## Deployment Strategies

### 1. Ephemeral Storage Pattern
```bash
# Container starts, restores from S3 if needed
if [ ! -f ./pb_data/data.db ]; then
  litestream restore -config ./pkg/litestream/litestream.yml
fi

# Start replication
litestream replicate -config ./pkg/litestream/litestream.yml &

# Start application
./pocketbase serve --dir ./pb_data
```

### 2. Zero-Downtime Migrations
```bash
# Create new instance with restored data
litestream restore -o ./pb_data/new_data.db s3://bucket/path

# Switch over atomically
mv ./pb_data/new_data.db ./pb_data/data.db
```

### 3. Disaster Recovery
```bash
# Full database loss scenario
litestream restore -config ./pkg/litestream/litestream.yml
# Data restored to last known good state
```

## Monitoring & Health Checks

### Health Check Endpoint
```bash
# Check replication lag
curl -f http://localhost:9090/metrics | grep litestream_replica_lag_seconds

# Check backup age
curl -f http://localhost:9090/metrics | grep litestream_db_last_replica_time
```

### Alerting Rules
```yaml
# Prometheus alerting rules
- alert: LitestreamLag
  expr: litestream_replica_lag_seconds > 300
  for: 5m
  annotations:
    summary: "Litestream replication lag is high"
```

## File Structure

```
pkg/litestream/
├── README.md           # This documentation
├── litestream.yml      # Configuration templates
└── scripts/            # Deployment and utility scripts
    ├── start.sh        # Production startup script
    ├── restore.sh      # Disaster recovery script
    └── health-check.sh # Health monitoring script
```

## Quick Start (Local First)

1. **Install Litestream**
   ```bash
   go run . tools dep install litestream
   ```

2. **Start replication (no setup required)**
   ```bash
   litestream replicate -config ./pkg/litestream/litestream.yml
   ```

3. **Test it works**
   ```bash
   # Check backup is being created
   ls -la ./backups/pb_data.db
   
   # Make a change to test
   sqlite3 pb_data/data.db "CREATE TABLE test_backup (id INTEGER);"
   
   # Verify backup updated
   sqlite3 ./backups/pb_data.db ".tables"
   ```

4. **Test restoration**
   ```bash
   # Simulate data loss
   rm pb_data/data.db
   
   # Restore from local backup
   litestream restore -config ./pkg/litestream/litestream.yml
   ```

## Environment Variables

### Local Development (None Required)
No environment variables needed for filesystem replication.

### Production (S3 Required)
- `AWS_ACCESS_KEY_ID` - S3 access key
- `AWS_SECRET_ACCESS_KEY` - S3 secret key
- `AWS_REGION` - S3 region

### Optional
- `LITESTREAM_ACCESS_KEY_ID` - Alternative to AWS_ACCESS_KEY_ID
- `LITESTREAM_SECRET_ACCESS_KEY` - Alternative to AWS_SECRET_ACCESS_KEY
- `LITESTREAM_RETENTION` - Backup retention period (default: 24h)

## Security Considerations

- **Encryption** - S3 server-side encryption enabled by default
- **Access control** - Use IAM policies with least privilege
- **Network isolation** - S3 bucket policies for network restrictions
- **Key rotation** - Regular AWS key rotation recommended

## Performance Tuning

### Sync Frequency
- **Production**: 1s (default)
- **Development**: 10s (reduces API calls)
- **Testing**: 100ms (fast for CI/CD)

### Compression
- **Enabled by default** - Reduces storage costs
- **CPU vs storage tradeoff** - Consider for large databases

## Troubleshooting

### Common Issues

**High memory usage:**
```bash
# Reduce memory buffer size
export LITESTREAM_BUFFER_SIZE=32MB
```

**S3 permissions:**
```bash
# Test S3 access
aws s3 ls s3://your-bucket/litestream/
```

**Database locked:**
```bash
# Ensure exclusive WAL mode
sqlite3 pb_data/data.db "PRAGMA journal_mode=WAL;"
```

### Recovery Scenarios

**Partial corruption:**
```bash
# Restore to specific time
litestream restore -timestamp 2024-01-15T10:30:00Z s3://bucket/path
```

**Complete loss:**
```bash
# Full restore
litestream restore -config ./pkg/litestream/litestream.yml
```

## Next Steps

1. **Configure S3 bucket** with appropriate IAM policies
2. **Set up monitoring** with Prometheus metrics
3. **Test disaster recovery** procedures
4. **Document retention policies** for compliance
5. **Integrate with deployment** automation