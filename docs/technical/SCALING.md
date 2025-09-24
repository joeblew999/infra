# Scaling Guide

This guide covers scaling the infrastructure management system on Fly.io, including horizontal scaling (adding machines) and vertical scaling (increasing resources per machine).

## Overview

The infrastructure supports three types of scaling:

1. **Horizontal Scaling**: Add or remove machines (`--count`)
2. **Vertical Scaling**: Increase CPU/memory per machine (`--cpu`, `--memory`)
3. **VM Type Scaling**: Change machine performance tiers (`--vm`)

## Scaling Commands

### Show Current Configuration

```bash
go run . tools fly scale
```

Output shows:
- **COUNT**: Number of machines
- **KIND**: VM type (shared, performance, etc.)
- **CPUS**: Number of CPU cores
- **MEMORY**: RAM allocation
- **REGIONS**: Geographic distribution

### Horizontal Scaling

Add or remove machines for increased capacity and redundancy:

```bash
# Scale to 2 machines
go run . tools fly scale --count 2

# Scale to 3 machines  
go run . tools fly scale --count 3

# Scale back to 1 machine
go run . tools fly scale --count 1
```

**Benefits:**
- Increased capacity and throughput
- Better fault tolerance (if one machine fails, others continue)
- Load distribution across multiple instances

**Considerations:**
- Each machine gets its own persistent volume
- Applications must handle distributed state properly
- Network latency between machines

### Vertical Scaling

Increase resources per machine for CPU/memory-intensive workloads:

```bash
# Scale memory to 1GB
go run . tools fly scale --memory 1024

# Scale memory to 2GB  
go run . tools fly scale --memory 2048

# Scale to 2 CPU cores
go run . tools fly scale --cpu 2

# Scale to 4 CPU cores
go run . tools fly scale --cpu 4
```

**Memory Options:**
- 256 MB, 512 MB, 1024 MB (1GB), 2048 MB (2GB), 4096 MB (4GB), 8192 MB (8GB)

**CPU Options:**
- 1, 2, 4, 8 cores (depending on VM type)

### VM Type Scaling  

Switch between different machine performance tiers:

```bash
# Shared CPU machines (burstable)
go run . tools fly scale --vm shared-cpu-1x    # 1 shared CPU
go run . tools fly scale --vm shared-cpu-2x    # 2 shared CPUs  
go run . tools fly scale --vm shared-cpu-4x    # 4 shared CPUs
go run . tools fly scale --vm shared-cpu-8x    # 8 shared CPUs

# Dedicated CPU machines (consistent performance)
go run . tools fly scale --vm performance-1x   # 1 dedicated CPU
go run . tools fly scale --vm performance-2x   # 2 dedicated CPUs
go run . tools fly scale --vm performance-4x   # 4 dedicated CPUs
go run . tools fly scale --vm performance-8x   # 8 dedicated CPUs
```

**Shared CPUs**: 
- Lower cost, good for most workloads
- CPU time shared with other applications
- Burstable performance

**Dedicated CPUs**: 
- Higher cost, consistent performance  
- Guaranteed CPU allocation
- Better for CPU-intensive applications

### Combined Scaling

Perform multiple scaling operations in one command:

```bash
# Scale to 2 machines with 2GB RAM and 2 CPUs each
go run . tools fly scale --count 2 --memory 2048 --cpu 2

# Scale to high-performance setup
go run . tools fly scale --count 3 --vm performance-4x --memory 4096
```

## Scaling Strategies

### Development/Testing
```bash
# Start small for development
go run . tools fly scale --count 1 --memory 512 --vm shared-cpu-1x
```

### Production - Low Traffic
```bash  
# Single machine with decent resources
go run . tools fly scale --count 1 --memory 1024 --cpu 2
```

### Production - Medium Traffic
```bash
# Multiple machines for redundancy
go run . tools fly scale --count 2 --memory 2048 --vm shared-cpu-2x
```

### Production - High Traffic
```bash
# High-performance cluster
go run . tools fly scale --count 3 --memory 4096 --vm performance-2x
```

### CPU-Intensive Workloads
```bash
# More CPUs for processing
go run . tools fly scale --count 2 --cpu 4 --vm performance-4x
```

### Memory-Intensive Workloads  
```bash
# More RAM for data processing
go run . tools fly scale --count 1 --memory 8192 --vm performance-2x
```

## Monitoring and Scaling Decisions

### Check Resource Usage

```bash
# Check application status
go run . tools fly status

# View logs for performance issues
go run . tools fly logs

# SSH in to check resource usage
go run . tools fly ssh
# Inside machine: htop, df -h, free -m
```

### Key Metrics to Monitor

1. **CPU Usage**: Scale up CPU if consistently >80%
2. **Memory Usage**: Scale up memory if >85% used
3. **Response Time**: Scale horizontally if response times increase
4. **Error Rate**: Scale up if errors due to resource constraints
5. **Disk Usage**: Monitor volume usage (/app/.data)

### Application Endpoints

Monitor these endpoints to make scaling decisions:

```bash
curl https://your-app.fly.dev/status   # Health check
curl https://your-app.fly.dev/metrics  # Application metrics
```

## Scaling Best Practices

### Gradual Scaling
- Start with minimum viable resources
- Scale up gradually based on actual usage
- Test application at each scale level

### Redundancy vs Cost
- Use multiple smaller machines for redundancy
- Use single larger machine for cost optimization
- Consider geographic distribution

### Resource Planning
- CPU: Scale when sustained >80% usage
- Memory: Scale when >85% usage (leave buffer for spikes)
- Network: Monitor latency between regions

### Application Architecture
- Ensure stateless application design for horizontal scaling
- Use external databases/storage for persistent state
- Implement proper health checks
- Handle graceful shutdowns

## Troubleshooting Scaling Issues

### Scaling Command Fails
```bash
# Check machine status
go run . tools fly status

# View recent logs
go run . tools fly logs

# Verify sufficient resources available
go run . tools fly scale  # Check current allocation
```

### Application Not Responding After Scale
```bash
# Check health checks
curl https://your-app.fly.dev/status

# SSH into machine to debug
go run . tools fly ssh

# Check goreman process status
ps aux | grep goreman
```

### Performance Issues After Scaling
- Verify application can handle multiple instances
- Check database connection pooling
- Monitor inter-service communication
- Review resource allocation balance (CPU vs memory)

## Cost Optimization

### Resource Right-sizing
- Monitor actual usage vs allocated resources
- Scale down unused resources
- Use shared CPUs for variable workloads

### Regional Considerations
- Deploy closer to users for lower latency
- Consider data transfer costs between regions
- Use volume replication for data locality

### Scheduled Scaling
Consider implementing scheduled scaling for predictable traffic patterns:
- Scale up during business hours
- Scale down during off-peak times
- Use external tools to automate based on time/metrics

## Advanced Scaling

### Multi-region Deployment
```bash  
# Deploy to multiple regions
go run . workflows deploy --region syd   # Sydney
go run . workflows deploy --region dfw   # Dallas  
go run . workflows deploy --region ams   # Amsterdam
```

### Volume Management
Each machine gets its own persistent volume. For scaling:
- Data replication between volumes
- Shared storage solutions
- Backup and restore procedures

### Load Balancing
Fly.io provides automatic load balancing across machines in the same app. Consider:
- Session affinity requirements
- Database connection patterns
- File upload/download handling