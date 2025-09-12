# Consistent Branding Across Multiple Channels

This example demonstrates how to maintain brand consistency across different output formats using a single source of brand configuration.

## The Problem

Most businesses struggle with brand consistency because:
- Different teams use different design tools
- Brand assets get scattered across various locations
- No single source of truth for colors, fonts, and layouts
- Manual design work for each document type
- Inconsistent application of brand guidelines

## How This System Helps

The template system allows you to:
1. **Define brand assets once** in configuration files
2. **Apply automatically** across all document types
3. **Generate consistent outputs** without manual design work
4. **Support multiple languages** using the same visual identity

## Brand Configuration

All brand information lives in `.data/brand/` directory:

```
.data/brand/
├── config.json          # Colors, fonts, sizing rules
├── logo.svg             # Primary logo
├── logo-dark.svg        # Logo for dark backgrounds
├── fonts/               # Custom font files
└── images/              # Brand imagery and icons
```

### Example Brand Config

```json
{
  "colors": {
    "primary": "#2563eb",
    "secondary": "#64748b", 
    "accent": "#f59e0b",
    "text": "#1f2937",
    "background": "#ffffff"
  },
  "fonts": {
    "heading": "Inter Bold",
    "body": "Inter Regular",
    "monospace": "JetBrains Mono"
  },
  "logo": {
    "primary": "logo.svg",
    "height": "40px",
    "dark_variant": "logo-dark.svg"
  }
}
```

## Output Formats

The same brand configuration generates consistent results across multiple channels:

### 1. PDF Documents
Business reports, invoices, and marketing materials with:
- Consistent color scheme throughout
- Branded headers and footers
- Logo placement following brand guidelines
- Typography matching brand fonts

```bash
# Generate branded invoice
./infra deck generate invoice.dsh --format=pdf
# Output: branded invoice with your colors, fonts, and logo
```

### 2. Web Pages
Landing pages and web content with:
- CSS generated from brand colors
- Web-safe font configurations
- Responsive logo sizing
- Consistent styling across pages

```bash
# Generate branded web page
./infra deck generate landing-page.dsh --format=html
# Output: HTML + CSS with brand styling applied
```

### 3. Email Templates
MJML email templates with:
- Brand colors in email design
- Logo optimized for email clients
- Consistent typography and spacing
- Mobile-responsive design

```bash
# Generate branded email template
./infra deck generate newsletter.dsh --format=mjml
# Output: MJML template ready for email campaigns
```

### 4. Presentations
Slide decks and presentations with:
- Branded slide templates
- Consistent chart and graph colors
- Logo placement on master slides
- Typography following brand guidelines

## Multi-Language Support

The same visual brand can support multiple languages:

```json
{
  "languages": {
    "en": {
      "company_name": "Your Company",
      "tagline": "Building better solutions"
    },
    "es": {
      "company_name": "Su Empresa", 
      "tagline": "Construyendo mejores soluciones"
    },
    "fr": {
      "company_name": "Votre Entreprise",
      "tagline": "Construire de meilleures solutions"
    }
  }
}
```

Generate the same document in different languages:

```bash
# English version
./infra deck generate report.dsh --lang=en --format=pdf

# Spanish version  
./infra deck generate report.dsh --lang=es --format=pdf

# French version
./infra deck generate report.dsh --lang=fr --format=pdf
```

All versions maintain the same visual identity while adapting content language.

## Real-World Example

A typical workflow might look like:

1. **Marketing team** needs a new product flyer
2. **Design team** has already set up brand configuration
3. **Content team** writes copy in template format
4. **System generates** PDF, web, and email versions automatically
5. **All outputs** maintain perfect brand consistency

No design review needed for brand compliance - it's built into the system.

## Getting Started

1. **Set up your brand assets** in `.data/brand/`
2. **Create templates** that reference brand variables
3. **Generate outputs** in multiple formats
4. **Iterate and refine** templates as needed

The initial setup takes some time, but once configured, generating new branded materials becomes a matter of minutes rather than hours or days. 