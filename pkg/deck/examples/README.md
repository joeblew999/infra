# Deck Examples

This directory contains example .dsh files demonstrating deck functionality.

## Examples

- **basic.dsh** - Simple shapes and text
- **markdown.dsh** - Native markdown integration with `content "markdown://file.md"`

## Testing

Use the deck CLI to test these examples:

```bash
# Test individual example
go run . tools deck watch pkg/deck/examples/basic.dsh

# Test all examples  
go run . tools deck watch pkg/deck/examples/ --formats=svg,png,pdf

# Manual generation
DECKFONTS=".data/font" pkg/deck/.build/bin/decksh pkg/deck/examples/basic.dsh > basic.xml
DECKFONTS=".data/font" pkg/deck/.build/bin/decksvg -outdir .data/deck/cache basic.xml
```