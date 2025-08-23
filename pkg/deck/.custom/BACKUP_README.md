# .custom - Custom Renderer Implementation Backup

This folder contains our custom Go rewrite of the deck rendering system that was developed before switching to the binary pipeline approach.

## What's Here

- **Complete PNG renderer** - Rewritten based on `.source/deck/cmd/pngdeck/pngdeck.go`
- **Enhanced PDF renderer** - Based on `.source/deck/cmd/pdfdeck/pdfdeck.go` patterns  
- **Comprehensive color system** - Full SVG color names, RGB, hex, HSV support
- **pkg/font integration** - Platform-independent font loading
- **Layer-based rendering** - Proper rendering order for all formats
- **Command integration** - Cobra commands for CLI interface
- **Shared constants** - Centralized constants matching deck source values

## Architecture Decision

We switched from this custom rewrite approach to the binary pipeline approach (`.old`) because:

1. **Effort**: Binary pipeline requires weeks, not months
2. **Compatibility**: 100% feature parity with original deck tools
3. **Maintenance**: Updates via git pull, not code rewrites
4. **Features**: All advanced features work immediately

## Value of This Work

This custom implementation was valuable for:
- **Domain understanding** - Deep knowledge of deck rendering pipeline
- **Font system integration** - Proved pkg/font approach works
- **Architecture insights** - Understanding of layer-based rendering
- **Reference implementation** - Fallback if binary approach has issues

## Files

- `renderer.go` - Core rendering interface and XML parsing
- `png_renderer.go` - Complete PNG renderer with gg graphics library
- `pdf_renderer.go` - PDF renderer with fpdf integration  
- `colors.go` - Comprehensive color handling system
- `consts.go` - Shared constants from original deck source
- `service.go` - HTTP service integration
- `cmd/` - Cobra command implementations
- `example/` - Example modules and test cases

## Usage (if needed)

If the binary pipeline approach encounters issues, this custom implementation can be restored by:

1. Moving files back to `pkg/deck/`
2. Updating imports in `pkg/cmd/deck.go`
3. Running `go mod tidy` to restore dependencies

The work done here significantly informed the final binary pipeline approach and validates our understanding of the deck ecosystem.