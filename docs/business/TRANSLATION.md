# Multi-Language Translation System

This document covers the current translation capabilities and future plans for supporting multiple languages across the platform.

## Current State

### What Works Today

**Go Applications (via Toki)**
The system currently supports translation for Go-based applications using the toki translation system:

- **Web interface** - All user-facing text can be translated
- **API responses** - Error messages and status text in multiple languages
- **CLI commands** - Help text and command output translation
- **Email notifications** - System-generated emails in user's preferred language

**Supported Languages:**
- English (en) - Primary language
- Spanish (es) - Full support
- French (fr) - Full support
- German (de) - Full support
- Japanese (ja) - Basic support
- Chinese (zh) - Basic support

### Current Limitations

**Markup and Templates**
- **Deck templates** - No translation support yet
- **Markdown documents** - Static content only
- **Email templates (MJML)** - Manual translation required
- **Web content** - Mixed support (interface translated, content not)

**Document Generation**
- PDF documents generate in single language per run
- No automatic content translation for user-generated content
- Brand assets (logos, images) not language-aware

## How Toki Translation Works

### For Go Applications

The toki system uses key-based translation with fallback support:

```go
// In Go code
import "ubuntusoftware/infra/pkg/toki"

// Translate a message
msg := toki.T("welcome_message", map[string]interface{}{
    "Name": user.Name,
})

// Output: "Welcome, John!" (English)
// Output: "Bienvenido, John!" (Spanish)
```

### Translation Files

Translation data lives in `.data/i18n/` directory:

```
.data/i18n/
├── en.json          # English (primary)
├── es.json          # Spanish
├── fr.json          # French
├── de.json          # German
├── ja.json          # Japanese
└── zh.json          # Chinese
```

**Example translation file (en.json):**
```json
{
  "welcome_message": "Welcome, {{.Name}}!",
  "error_not_found": "Resource not found",
  "button_submit": "Submit",
  "email_subject_welcome": "Welcome to {{.CompanyName}}",
  "status_processing": "Processing your request..."
}
```

**Spanish version (es.json):**
```json
{
  "welcome_message": "¡Bienvenido, {{.Name}}!",
  "error_not_found": "Recurso no encontrado",
  "button_submit": "Enviar",
  "email_subject_welcome": "Bienvenido a {{.CompanyName}}",
  "status_processing": "Procesando tu solicitud..."
}
```

## What Needs to Be Built

### 1. Deck Template Translation

**Current Problem:**
Deck templates are static - they don't support dynamic language switching.

**Proposed Solution:**
Extend deck templates to support toki translation keys:

```
// Current deck template (static)
title "Welcome to Our Company"
text "We provide excellent services"

// Proposed deck template (translatable)
title {{T "company_welcome_title"}}
text {{T "company_services_description"}}
```

**Implementation Requirements:**
- Extend deck parser to recognize `{{T "key"}}` syntax
- Integrate toki translation lookup during document generation
- Support language parameter in deck CLI commands

### 2. Markdown Translation Support

**Current Problem:**
Markdown documents are static files with no translation mechanism.

**Proposed Solutions:**

**Option A: Separate Files Per Language**
```
docs/
├── business/
│   ├── README.en.md
│   ├── README.es.md
│   ├── README.fr.md
│   └── README.de.md
```

**Option B: Embedded Translation Keys**
```markdown
# {{T "business_overview_title"}}

{{T "business_overview_intro"}}

## {{T "features_heading"}}

- {{T "feature_document_generation"}}
- {{T "feature_workflow_automation"}}
```

**Option C: Hybrid Approach**
- Static structural content in separate files per language
- Dynamic content (user data, system status) uses translation keys
- Best of both worlds for different content types

### 3. Email Template Translation

**Current Problem:**
MJML email templates require manual duplication for each language.

**Proposed Solution:**
Template inheritance with translation support:

```mjml
<!-- Base template -->
<mj-text>{{T "email_welcome" .UserName}}</mj-text>
<mj-button>{{T "button_get_started"}}</mj-button>

<!-- Generates language-specific versions -->
<!-- en: "Welcome, John!" / "Get Started" -->
<!-- es: "¡Bienvenido, John!" / "Comenzar" -->
```

## Implementation Roadmap

### Phase 1: Deck Translation (Priority: High)
- [ ] Extend deck template parser to support translation keys
- [ ] Add language parameter to deck CLI commands
- [ ] Test with existing brand templates
- [ ] Update documentation with translation examples

### Phase 2: Markdown Translation (Priority: Medium)  
- [ ] Choose translation approach (separate files vs embedded keys)
- [ ] Implement markdown preprocessor for translation
- [ ] Convert existing documentation to translatable format
- [ ] Add language switching to web interface

### Phase 3: Email Template Translation (Priority: Medium)
- [ ] Extend MJML processor to support translation keys
- [ ] Create base email templates with translation support
- [ ] Update email sending system to use user's preferred language
- [ ] Test with common email workflows

### Phase 4: Advanced Features (Priority: Low)
- [ ] Automatic content translation via external APIs
- [ ] Language-specific brand assets (logos, images)
- [ ] Right-to-left (RTL) language support
- [ ] Pluralization and complex grammar rules

## Language Selection Strategy

### User Language Detection
1. **Explicit parameter** - User specifies language in CLI/API call
2. **User preference** - Stored in user profile/session
3. **Browser language** - Accept-Language header for web interface
4. **System default** - Fallback to English if no preference found

### URL Structure for Web Interface
- `example.com/en/docs` - English documentation
- `example.com/es/docs` - Spanish documentation  
- `example.com/fr/docs` - French documentation

### CLI Language Support
```bash
# Generate document in Spanish
./infra deck generate report.dsh --lang=es

# Set default language for session
./infra config set language es

# Generate all supported languages
./infra deck generate report.dsh --all-languages
```

## Content Management Strategy

### Who Manages Translations?

**Technical Content (UI, errors, system messages):**
- Developers maintain translation keys in code
- Translation files managed in version control
- Automated testing for missing translations

**Business Content (documentation, marketing):**
- Content teams provide source text in English
- Professional translation services for high-quality content
- Community contributions for less critical content

**User-Generated Content:**
- Users responsible for their own content translation
- Optional integration with translation services (Google Translate, DeepL)
- Machine translation for preview only, not production

### Translation Workflow

1. **Content Creation** - Written in English with translation keys
2. **Translation Request** - Send keys to translation service/team
3. **Integration** - Add translations to appropriate files
4. **Testing** - Verify translation rendering across formats
5. **Deployment** - Release with new language support

## Technical Considerations

### Performance
- Translation files loaded at startup, cached in memory
- Fallback chain: user language → English → key name
- Lazy loading for large translation files

### Maintenance
- Automated detection of missing translation keys
- Version control integration for translation updates
- Translation validation in CI/CD pipeline

### Extensibility
- Plugin system for custom translation providers
- Support for regional variants (en-US vs en-GB)
- Integration with external translation management systems

## Getting Started with Translation

### For Developers
1. **Add translation keys** to Go code using toki
2. **Update translation files** with new keys
3. **Test multiple languages** during development
4. **Follow naming conventions** for translation keys

### For Content Teams  
1. **Use translation keys** in templates instead of hardcoded text
2. **Provide context** for translators (where/how text is used)
3. **Test translated content** in actual application context
4. **Consider cultural differences** beyond just language

### For System Administrators
1. **Configure default language** for system
2. **Set up translation file management** process
3. **Monitor for missing translations** in logs
4. **Plan storage** for multiple document versions

This system is designed to scale from simple single-language deployments to complex multi-language, multi-regional implementations as your needs grow.