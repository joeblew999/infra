# Font Package

## What
A comprehensive font management system that downloads, caches, and provides fonts for various output formats.

## Why
The system needs fonts for multiple use cases:
- **Deck tools**: Require TTF fonts for SVG/PNG/PDF generation
- **Web applications**: Need WOFF2 fonts for optimal loading
- **Email templates (MJML)**: Use web fonts for rich formatting

## Features
- **Dual Format Support**: Downloads both TTF and WOFF2 fonts from Google Fonts
- **Smart Caching**: Stores fonts in organized directory structure
- **Registry System**: Tracks downloaded fonts with metadata
- **Format-Specific APIs**: `GetTTF()`, `GetFormat()`, `CacheTTF()` methods
- **Auto-Download**: Fetches fonts on-demand from Google Fonts API

## Usage
```go
fm := font.NewManager()

// For web use (WOFF2)
path, err := fm.Get("Roboto", 400)

// For deck tools (TTF)
path, err := fm.GetTTF("Helvetica", 400)

// Cache specific format
err := fm.CacheTTF("Arial", 400)
```

## Architecture
- `Manager`: Main interface for font operations
- `Registry`: Tracks cached fonts and metadata
- `Google API`: Downloads fonts directly from Google Fonts Web API
- **Storage**: `.data/font/Family/weight.format` structure