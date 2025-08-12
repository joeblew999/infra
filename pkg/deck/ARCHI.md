# deck

Deck is a comprehensive visualization system that transforms declarative `.dsh` markup into streaming SVG graphics via WASM and DataStar.

## Architecture Overview

### Core Pipeline
```
.dsh file → decksh → XML → WASM → SVG → DataStar → Browser
```

### Runtime Components
- **decksh.wasm**: Parser/compiler (dsh → XML)
- **svgdeck.wasm**: SVG generator (XML → SVG)  
- **dshfmt.wasm**: Code formatter
- **dshlint.wasm**: Syntax validator

## System Architecture

### 1. Source Management (Independent of pkg/dep)
**Source Code Handling:**
- **Git cloning**: Direct from GitHub repos
- **Version tracking**: Git tags and commit hashes
- **Local compilation**: Go toolchain required
- **WASM cross-compilation**: `GOOS=js GOARCH=wasm`
- **Cache management**: Store compiled binaries and WASM

**Tools to Compile:**
- `decksh` - dsh to XML compiler (from ajstarks/decksh)
- `svgdeck` - XML to SVG converter (from ajstarks/deck)
- `dshfmt` - dsh formatter (from ajstarks/decksh)
- `dshlint` - dsh linter (from ajstarks/decksh)

### 2. Build System
**pkg/deck handles:**
- Source code embedding via Go imports
- WASM cross-compilation at build time
- Version tracking via Go modules
- Build orchestration for WASM modules

**pkg/dep remains:**
- Binary-only dependency management
- No source compilation complexity

### 2. Storage Layout
```
.data/deck/
├── wasm/          # Compiled WASM modules (core runtime)
├── fonts/         # Google Fonts cache
├── templates/     # Example .dsh files
├── cache/         # SVG output cache
└── source/        # Cloned source repos for reference
```

### 3. API Design
- **POST /deck/compile** - Compile dsh → streaming SVG
- **POST /deck/format** - Format dsh code
- **POST /deck/lint** - Validate dsh syntax
- **GET /deck/examples** - Sample files
- **GET /deck/playground** - Interactive editor

### 4. Font Strategy
- **Primary**: Google Fonts API integration
- **Cache**: Local font storage in `.data/deck/fonts/`
- **Configuration**: `DECKFONTS` environment variable
- **Fallback**: System fonts

### 5. Runtime Architecture
**File-based Pipeline (Unidirectional):**
- **File Watcher**: Monitors `.dsh` file changes on filesystem
- **Automatic Pipeline**: `.dsh` change → decksh.wasm → XML → svgdeck.wasm → SVG
- **Output Generation**: Writes XML and SVG to `.data/deck/cache/` automatically
- **No Interactive Mode**: Pure file-based processing

**Direct Wazero Integration:**
- **decksh.wasm** → parses .dsh → XML (via Wazero) on file change
- **svgdeck.wasm** → converts XML → SVG (via Wazero) on file change
- **Wazero Runtime**: Built into pkg/deck directly
- **Memory Management**: Handled internally for file processing
- **Stream Processing**: For large .dsh files

### 6. Development Workflow
1. **Import**: Add deck tools as Go modules
2. **Build**: Cross-compile to WASM at build time
3. **Watch**: `go run . deck watch /path/to/*.dsh` (file monitoring)
4. **Test**: Manual file changes trigger automatic processing

### 7. File Processing Mode
**Unidirectional Pipeline:**
```
.dsh file change → decksh.wasm → XML file → svgdeck.wasm → SVG file
```

**Output Structure:**
```
.data/deck/cache/
├── example.dsh.xml    # Generated XML
├── example.dsh.svg    # Generated SVG
└── ...
```

**Watch Mode:**
- Monitors filesystem for `.dsh` file changes
- Processes immediately on change
- No interactive editing or web interfaces
- Pure file-based workflow
- Cache management for generated files

### 7. Real-time Features
- **DataStar streaming**: Live SVG updates
- **Interactive editing**: Real-time compilation
- **Font loading**: Dynamic Google Fonts
- **Caching**: Compiled WASM + SVG cache

### 8. Cross-format Support
- **SVG**: Primary browser format
- **PNG**: Raster fallback via pngdeck
- **PDF**: Document export via pdfdeck
- **Geo**: 2D maps via geodeck
- **Images**: giftsh transformations

## Tools Reference

### Core Commands
- **decksh** - [GitHub](https://github.com/ajstarks/decksh/tree/master/cmd)
- **svgdeck** - [GitHub](https://github.com/ajstarks/deck/tree/master/cmd/svgdeck)
- **dshfmt** - [GitHub](https://github.com/ajstarks/decksh/tree/master/cmd/dshfmt)
- **dshlint** - [GitHub](https://github.com/ajstarks/decksh/tree/master/cmd/dshlint)

### Controllers
- **deckd** - REST API server
- **deckweb** - Web interface

### Examples
- [dubois-data-portraits](https://github.com/ajstarks/dubois-data-portraits)
- [deckviz](https://github.com/ajstarks/deckviz)

## Environment Variables
- `DECKFONTS` - Font directory path
- `DECK_CACHE` - SVG cache directory
- `GOOGLE_FONTS_API` - Fonts API key

## Testing
```bash
# Check all tools
./infra deck test.dsh

# Interactive playground
./infra deck playground

# Format validation
go test ./pkg/deck/...
```