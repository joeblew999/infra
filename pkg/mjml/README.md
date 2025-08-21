# MJML Email Template Package

Package mjml provides email template rendering using MJML (Mailjet Markup Language) for generating responsive HTML emails at runtime.

## Features

- Convert MJML XML templates to responsive HTML emails
- Template variable substitution 
- Performance optimized with caching
- Debug mode for development
- Email client compatibility

## Usage

### Basic Template Rendering

```go
import "github.com/joeblew999/infra/pkg/mjml"

// Render MJML template to HTML
renderer := mjml.NewRenderer()
html, err := renderer.RenderTemplate("welcome", map[string]interface{}{
    "name": "John Doe",
    "email": "john@example.com",
})
```

### Template Management

```go
// Load templates from directory
err := renderer.LoadTemplatesFromDir("templates/")

// Load single template
err := renderer.LoadTemplate("welcome", mjmlContent)

// List available templates
templates := renderer.ListTemplates()
```

### Advanced Options

```go
// Create renderer with options
renderer := mjml.NewRenderer(
    mjml.WithCache(true),        // Enable caching
    mjml.WithDebug(true),        // Add debug attributes
    mjml.WithValidation(true),   // Validate MJML
)
```

## Template Structure

MJML templates are stored as XML files with Go template syntax for variables:

```xml
<mjml>
  <mj-head>
    <mj-title>{{.subject}}</mj-title>
  </mj-head>
  <mj-body>
    <mj-section>
      <mj-column>
        <mj-text>Hello {{.name}}!</mj-text>
      </mj-column>
    </mj-section>
  </mj-body>
</mjml>
```

## Email Templates

Common template patterns:
- Welcome emails
- Password reset
- Notifications  
- Newsletters
- Transactional emails

## Dependencies

- [gomjml](https://github.com/preslavrachev/gomjml) - MJML to HTML conversion
- Go html/template - Template variable substitution