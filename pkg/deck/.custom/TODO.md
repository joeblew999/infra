# TODO - pkg/deck Improvements

This package implements deck presentation generation based on the original [ajstarks/deck](https://github.com/ajstarks/deck) source found in `.source/deck/cmd/`. The renderers are adapted from the original implementations but still need several enhancements to achieve full feature parity.

## Completed ✅

- ✅ **Font Integration**: All renderers use pkg/font exclusively for portable font loading
- ✅ **Comprehensive Color Handling**: Full SVG color names, RGB, hex, and HSV support via colorlookup()  
- ✅ **PNG Renderer Rewrite**: Complete rewrite based on `.source/deck/cmd/pngdeck/pngdeck.go`
- ✅ **PDF Renderer Enhancement**: Based on `.source/deck/cmd/pdfdeck/pdfdeck.go` patterns
- ✅ **Shared Constants**: Centralized constants to match original deck source values
- ✅ **Layer-based Rendering**: Proper rendering order for all formats
- ✅ **Basic Shape Support**: Rectangles, ellipses, lines, basic text rendering
- ✅ **Source Update Command**: `infra deck update-source` command to keep `.source` repositories up-to-date

## High Priority 🔴

we have to work out how we want to do tests and the test data and goldens. It maybe there the 2 examples actualyl have inputs and outputs that match inside the .source code. That would be very luck if it were ??       

you mention "when pkg/font fails". Why should it ? google fonts used in the pkg/fonts is a huge corpus ? 

We have to ask the question. What IF we just used the deck code, and compiled it to Bianry and wasm, and pipelned the 2 togehter ? It might be a better apprach. This is what the .old code does !!! We need to decide based on effort. We might be reinventng the wheel too much ?

We could still use our pkg/fonts trick. We might have to parse the dekcsh and deckxml and see the fonts used. But i think in the original deck code, you pass the fonts in on the CLI ?

### PNG Renderer Enhancements (based on `.source/deck/cmd/pngdeck/`)

- [ ] **Image Loading & Processing**: Implement proper image loading with gift library for scaling, cropping, autoscale
- [ ] **Advanced Text Features**: 
  - [ ] Text rotation support  
  - [ ] Block text with proper word wrapping
  - [ ] Code block styling with background rectangles
  - [ ] Text alignment improvements (especially center/right)
- [ ] **Font Loading Improvements**: 
  - [ ] Better TTF font file handling and validation
  - [ ] Font fallback chain when pkg/font fails
  - [ ] Font size scaling and DPI handling
- [ ] **List Rendering**: 
  - [ ] Proper bullet positioning and sizing
  - [ ] Numbered lists with correct formatting
  - [ ] List item color and font overrides
- [ ] **Advanced Shapes**: 
  - [ ] Proper arc rendering with angles
  - [ ] Bezier curve implementation  
  - [ ] Polygon coordinate parsing and rendering
- [ ] **Gradient Support**: Rectangle and ellipse gradients with proper color stops

### PDF Renderer Enhancements (based on `.source/deck/cmd/pdfdeck/`)

- [ ] **Custom Font Integration**: 
  - [ ] Add pkg/font TTF files to PDF using AddFont()
  - [ ] Font embedding and subsetting
  - [ ] Unicode text support
- [ ] **Advanced Shapes**: 
  - [ ] Proper arc rendering (currently simplified as ellipse)
  - [ ] Bezier curves (currently simplified as lines)  
  - [ ] Polygon coordinate parsing and rendering
- [ ] **Image Support**: 
  - [ ] Image loading and embedding in PDF
  - [ ] Image scaling and positioning
  - [ ] Support for JPEG, PNG, GIF formats
- [ ] **Gradient Implementation**: 
  - [ ] PDF gradient support (fpdf limitations)
  - [ ] Background and shape gradients
- [ ] **Text Enhancements**:
  - [ ] Text rotation support
  - [ ] Better text alignment and positioning
  - [ ] Multi-line text handling

### SVG Renderer Rewrite (based on `.source/deck/cmd/svgdeck/`)

- [ ] **Complete SVG Renderer Rewrite**: Current implementation is basic, needs full rewrite based on `svgdeck.go`
- [ ] **Advanced Color Handling**: HSV color conversion and SVG-specific color processing  
- [ ] **Text Features**:
  - [ ] Text rotation and transformation
  - [ ] Text wrapping and block text  
  - [ ] Font lookup mapping system
  - [ ] Proper text alignment (start/middle/end)
- [ ] **Shape Enhancements**:
  - [ ] Arc rendering with proper polar coordinates
  - [ ] Bezier curve paths
  - [ ] Polygon rendering with coordinate parsing
- [ ] **Image Support**: 
  - [ ] SVG image embedding and linking
  - [ ] Image scaling and autoscale
  - [ ] Caption rendering
- [ ] **Advanced Features**:
  - [ ] Navigation links between slides
  - [ ] Grid with percentage labels
  - [ ] Gradient definitions and usage
  - [ ] Text along paths

## Medium Priority 🟡

### Core Infrastructure  

- [ ] **Page Size Support**: Implement full page size mapping (Letter, A4, Legal, etc.) from `.source`
- [ ] **Multi-slide Support**: 
  - [ ] Render all slides in deck, not just first slide
  - [ ] Slide range selection (pages flag)  
  - [ ] Slide navigation and linking
- [ ] **File I/O Improvements**:
  - [ ] Include file support for text content
  - [ ] Better error handling and reporting
  - [ ] Progress reporting for multi-slide operations
- [ ] **Command Line Parity**: Match original deck tool CLI interfaces exactly

### Rendering Pipeline

- [ ] **XML Processing**: Better XML parsing and validation matching deck library patterns
- [ ] **Canvas Management**: Proper canvas sizing and coordinate systems  
- [ ] **Performance**: Optimize rendering for large presentations
- [ ] **Memory Management**: Reduce memory usage for image-heavy presentations

## Low Priority 🟢

### Developer Experience

- [ ] **Testing**: 
  - [ ] Unit tests for all renderer functions
  - [ ] Integration tests comparing output to original tools
  - [ ] Golden file testing for regression prevention
- [ ] **Documentation**: 
  - [ ] Complete API documentation
  - [ ] Examples and tutorials
  - [ ] Performance benchmarks
- [ ] **Debugging**: 
  - [ ] Verbose rendering mode
  - [ ] Debug output for troubleshooting
  - [ ] Render pipeline visualization

### Advanced Features

- [ ] **Animation Support**: Basic SVG animations (if supported in original)
- [ ] **Interactive Elements**: Clickable areas and navigation (SVG)  
- [ ] **Accessibility**: Alt text and semantic markup
- [ ] **Optimization**: 
  - [ ] SVG optimization and minification
  - [ ] PDF compression and optimization
  - [ ] PNG optimization

## Technical Debt 🔧

- [ ] **Code Organization**: 
  - [ ] Reduce code duplication between renderers
  - [ ] Better separation of concerns
  - [ ] Consistent error handling patterns
- [ ] **Dependencies**: 
  - [ ] Minimize external dependencies where possible
  - [ ] Version pinning and compatibility testing
- [ ] **Configuration**: 
  - [ ] Centralized configuration management
  - [ ] Environment-specific defaults

## Notes

- All implementations should closely follow the patterns in `.source/deck/cmd/`
- Priority should be given to PNG renderer improvements as it's the most commonly used format
- SVG renderer needs the most work to match original svgdeck functionality  
- Font integration with pkg/font is working but needs better error handling
- Color system is now comprehensive and matches original deck source exactly