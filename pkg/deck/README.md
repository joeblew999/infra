# Deck Visualization System

A complete pipeline for transforming declarative `.dsh` markup into SVG graphics.

Deck Org does not make binary or wasm releaes, so we do it oursouce in cmd/build

pkg/dep then downloads them off our github releases.

## Quick Start

```bash
# Install all deck tools
./infra deck install

# Watch .dsh files for changes
./infra deck watch ./slides/

# List available tools
./infra deck list

# Clean build artifacts
./infra deck clean
```

## Core Pipeline

```
.dsh file → decksh → XML → svgdeck → SVG
```

## Available Tools

| Tool | Purpose | Usage |
|------|---------|--------|
| `decksh` | Compile .dsh to XML | `decksh input.dsh > output.xml` |
| `svgdeck` | Convert XML to SVG | `svgdeck input.xml` |
| `pngdeck` | Convert XML to PNG | `pngdeck input.xml` |
| `pdfdeck` | Convert XML to PDF | `pdfdeck input.xml` |
| `dshfmt` | Format .dsh files | `dshfmt input.dsh` |
| `dshlint` | Validate .dsh syntax | `dshlint input.dsh` |

## File Formats

### .dsh File Structure
```
deck 10 5
  text "Hello World" 5 2.5 2
  circle 5 2.5 1 "red" 0.5
  rect 1 1 9 4 "blue" 0.3
edeck
```

### File Locations
- **Binaries**: `.data/deck/bin/`
- **WASM**: `.data/deck/wasm/`
- **Output**: `.data/deck/cache/`

## Usage Examples

### Create and Process Files
```bash
# Format a .dsh file
./.data/deck/bin/dshfmt slides/example.dsh

# Validate syntax
./.data/deck/bin/dshlint slides/example.dsh

# Manual pipeline
./.data/deck/bin/decksh slides/example.dsh > slides/example.xml
./.data/deck/bin/svgdeck slides/example.xml
```

### Auto-Watch Mode
```bash
# Watch directory for changes
./infra deck watch ./slides/

# Watch multiple directories
./infra deck watch ./slides/ ./templates/
```

### Build System Commands
```bash
./infra deck install    # Build all tools
./infra deck status     # Show build status
./infra deck clean      # Clean artifacts
./infra deck list       # List available tools
```

## Environment
- **DECKFONTS**: Font directory path
- **DECK_CACHE**: SVG cache directory

## Examples
SEE tests in pkg/deck

We need to finish grabbign the other 2 examples repos also.