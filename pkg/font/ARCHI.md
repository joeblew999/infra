# pkg/font Architecture

## Purpose
Manage Google Fonts for the deck visualization system, ensuring consistent typography across SVG, PNG and PDF outputs.

## Core Components

https://fonts.google.com

https://developers.google.com/fonts/docs/developer_api

---

https://github.com/go-mods/gfonts/blob/master/gfonts.go ?

https://www.w3schools.com/css/css_font_google.asp shows how to directly acces them, which we might want too ?
- https://fonts.googleapis.com/css?family=Sofia


We will use HUGO later too ? 


### Font Discovery
- Search Google Fonts API for typefaces matching deck requirements
- Filter by SVG compatibility and licensing
- Cache metadata for offline access

### Font Acquisition
- Download selected fonts in required formats (TTF, WOFF, WOFF2)
- Verify checksums and integrity
- Store in local cache for fast access

### Font Management
- Track font versions and updates
- Handle font families and weights
- Maintain font-to-deck mapping configuration

## Cache Strategy
- **Location**: `.data/fonts/` (shared with deck system)
- **Structure**: `.data/fonts/{family}/{weight}.{format}`
- **Metadata**: JSON index file for font properties

## Integration Points

### Deck System Integration
- **Path**: `pkg/deck/build/fonts/` (symlinked to `.data/fonts/`)
- **Usage**: decksh and svgdeck tools reference cached fonts
- **Config**: Font selection via .dsh file directives

### API Interface
- `GetFont(family, weight string) (path string, error)`
- `ListAvailableFonts() []FontInfo`
- `CacheFont(family, weight string) error`

## Simple Flow
```
.dsh file → decksh → pkg/font cache → .data/fonts/ → SVG output
```

## Dependencies
- Google Fonts API access
- HTTP client for downloads
- File system caching
- JSON for metadata