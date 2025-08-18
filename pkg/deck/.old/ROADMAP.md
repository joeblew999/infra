# Deck System Roadmap

## Phase 1: Foundation (Week 1)
### Basic Source Management
- [ ] Create `pkg/deck/manager.go` for Git source handling
- [ ] Add `GetDeckPath()` and `GetDeckWASMPath()` to `pkg/config`
- [ ] Implement `.data/deck/` directory structure
- [ ] Basic Git clone functionality for repos

### Repositories to Target
- `ajstarks/decksh` - dsh compiler
- `ajstarks/deck` - SVG generator
- https://github.com/ajstarks/dubois-data-portraits - examples
- https://github.com/ajstarks/deckviz - examples

## Phase 2: Build System (Week 2)
### Go Module Integration
- [ ] Import decksh, svgdeck, dshfmt, dshlint as Go modules
- [ ] Version tracking via go.mod
- [ ] Source code embedding (no external binaries)
- [ ] Build-time WASM compilation

### WASM Cross-Compilation
- [ ] `GOOS=js GOARCH=wasm` build setup
- [ ] WASM module validation
- [ ] Binary size optimization
- [ ] Embedded WASM assets

## Phase 3: CLI Integration (Week 3)
### Basic Commands
- [ ] `go run . deck install` - Download & compile tools
- [ ] `go run . deck build` - Compile to WASM
- [ ] `go run . deck update` - Update source repos
- [ ] `go run . deck status` - Show versions & cache

### Version Management
- [ ] Git tag resolution
- [ ] Commit hash pinning
- [ ] Build metadata storage
- [ ] Rollback capability

## Phase 4: Direct Wazero Integration (Week 4)
### WASM Runtime (No Bento)
- [ ] Wazero integration directly in pkg/deck
- [ ] WASM module loading and initialization
- [ ] Memory management for large files
- [ ] Function call interface design
- [ ] Error handling and recovery

### Basic API
- [ ] `POST /deck/compile` - Basic dsh→SVG
- [ ] Error handling and validation
- [ ] Simple caching layer

## Phase 5: File Watcher (Week 5)
### Filesystem Monitoring
- [ ] File watcher implementation
- [ ] .dsh file change detection
- [ ] Automatic pipeline triggering
- [ ] Output file generation
- [ ] Cache management for generated files

### Unidirectional Processing
- [ ] Change → XML → SVG flow
- [ ] Error handling for malformed .dsh
- [ ] Performance optimization for large files

## Phase 6: Advanced Features (Week 6)
### Font System
- [ ] Google Fonts API integration
- [ ] Local font caching
- [ ] Font fallback system
- [ ] Configuration management

### Format Support
- [ ] PNG export via pngdeck
- [ ] PDF export via pdfdeck
- [ ] 2D maps via geodeck
- [ ] Image transformations via giftsh

## Phase 7: Production Ready (Week 7)
### Optimization
- [ ] Build caching performance
- [ ] WASM bundle size optimization
- [ ] Memory management
- [ ] Error reporting

### Documentation & Examples
- [ ] Complete API documentation
- [ ] Interactive tutorials
- [ ] Sample gallery
- [ ] Integration tests

## Development Workflow
```bash
# Phase 1-2: Foundation
go run . deck install        # Clone & compile
go run . deck build --wasm   # Cross-compile

# Phase 3-4: Testing
go run . deck test example.dsh
go run . deck playground    # Interactive testing

# Phase 5-7: Production
go run . deck serve         # Full service mode
```

## Success Criteria
- [ ] Zero external Go dependencies beyond stdlib
- [ ] Clean separation from pkg/dep
- [ ] Cross-platform (macOS, Linux, Windows)
- [ ] WASM modules under 2MB each
- [ ] Real-time streaming working
- [ ] Interactive playground functional

## Risk Mitigation
- **Build failures**: Graceful degradation to native binaries
- **Large WASM**: Progressive loading and caching
- **Git dependencies**: Mirror fallback for offline builds
- **Font issues**: System font fallbacks