# Managed Binary Distribution System Design

## Overview
Create a complete pipeline for managing binary dependencies with full control over releases and distribution.

## Workflow Phases

### Phase 1: Collection
```
Original Sources → Local Collection → Validation → Standardization
```
- Download from original sources (GitHub releases, etc.)
- Collect all platforms in parallel
- Validate checksums/signatures
- Standardize naming and structure

### Phase 2: Managed Release
```
Local Collection → GitHub Release → Asset Upload → Manifest Generation
```
- Create versioned GitHub releases in your repo
- Upload all platform binaries as release assets
- Generate manifest with checksums and metadata
- Atomic release publishing

### Phase 3: Runtime Distribution
```
Runtime Request → Managed Release → Fast Download → Local Cache
```
- Download from your managed releases first
- Fallback to original sources if needed
- Local caching for performance
- Version consistency guarantees

## Storage Format

### Directory Structure
```
.collection/
├── binaries/
│   ├── flyctl/
│   │   ├── v0.3.162/
│   │   │   ├── darwin-amd64/flyctl
│   │   │   ├── darwin-arm64/flyctl
│   │   │   ├── linux-amd64/flyctl
│   │   │   ├── linux-arm64/flyctl
│   │   │   ├── windows-amd64/flyctl.exe
│   │   │   └── windows-arm64/flyctl.exe
│   │   └── manifest.json
│   └── caddy/
│       ├── v2.10.0/
│       └── manifest.json
└── metadata/
    ├── collection-report.json
    └── release-status.json
```

### Manifest Format
```json
{
  "binary": "flyctl",
  "version": "v0.3.162",
  "collection_date": "2025-08-17T10:00:00Z",
  "source": {
    "repo": "superfly/flyctl",
    "release_url": "https://github.com/superfly/flyctl/releases/tag/v0.3.162"
  },
  "platforms": {
    "darwin-amd64": {
      "filename": "flyctl",
      "size": 12345678,
      "sha256": "abc123...",
      "executable": true
    },
    "windows-amd64": {
      "filename": "flyctl.exe", 
      "size": 12345678,
      "sha256": "def456...",
      "executable": true
    }
  },
  "managed_release": {
    "repo": "joeblew999/infra-binaries",
    "tag": "flyctl-v0.3.162",
    "published": true,
    "url": "https://github.com/joeblew999/infra-binaries/releases/tag/flyctl-v0.3.162"
  }
}
```

## Commands Design

### Collection Commands
```bash
# Collect a specific binary for all platforms
go run . tools dep collect flyctl v0.3.162

# Collect all configured binaries
go run . tools dep collect-all

# Collect only missing platforms
go run . tools dep collect flyctl --missing-only
```

### Release Management Commands  
```bash
# Upload collected binaries to managed release
go run . tools dep release flyctl v0.3.162

# Release all collected binaries
go run . tools dep release-all

# Check release status
go run . tools dep release-status
```

### Runtime Commands (Enhanced)
```bash
# Install from managed releases (default)
go run . tools dep install flyctl

# Force install from original source
go run . tools dep install flyctl --source=original

# Show source preference
go run . tools dep install flyctl --dry-run
```

## Implementation Plan

### 1. Collection System
- Multi-platform downloader using existing installers
- Parallel collection with progress tracking
- Checksum validation and metadata generation
- Local storage management

### 2. Release Management
- GitHub API integration for release creation
- Batch asset upload with progress
- Manifest generation and upload
- Atomic publishing with rollback

### 3. Runtime Enhancement
- Managed release detection and preference
- Fallback chain: managed → original → cache
- Version consistency checks
- Performance optimizations

### 4. Configuration
- Repository settings for managed releases
- Platform matrix configuration
- Source preference policies
- Caching and cleanup settings

## Benefits

1. **Control**: Full control over binary versions and availability
2. **Performance**: Faster downloads from your CDN/releases
3. **Reliability**: Fallback to original sources if needed
4. **Security**: Checksums and signatures validation
5. **Consistency**: Same binary versions across all environments
6. **Offline**: Local caching for offline development

## Integration Points

- Extends existing dep.json configuration
- Uses existing installer framework
- Enhances existing CLI commands
- Leverages existing progress/download utilities