# Web Application Design

## Current State
- Single `server.go` file with embedded HTML strings
- Basic DataStar integration for SSE
- NATS JetStream messaging
- Manual HTML generation in handlers

## Refactoring Goals

### 1. Component-Based Architecture with DatastarUI
- Replace embedded HTML with Go/templ components
- Use DatastarUI component library for consistent, accessible UI
- Copy-paste approach: own the component code for customization

### 2. Proposed Structure
```
web/
├── components/           # DatastarUI components
│   ├── ui/              # Base UI components (button, card, etc.)
│   ├── layout/          # Layout components (nav, sidebar)
│   └── pages/           # Page-specific components
├── handlers/            # HTTP handlers
├── services/           # Business logic
├── templates/          # Go/templ template files
└── static/             # Static assets
```

### 3. Key Improvements
- **Type-safe templates**: Go/templ instead of string concatenation
- **Reusable components**: DatastarUI component system
- **Better separation**: Handlers focus on logic, templates on presentation
- **Consistent styling**: DatastarUI + Tailwind CSS
- **Accessibility**: Built into DatastarUI components

### 4. DatastarUI Integration Strategy

**Recommended Approach: Copy-Paste Pattern**
- DatastarUI is designed as a copy-paste library (like shadcn/ui)
- Copy only the components you need into `web/components/ui/`
- Customize directly in your codebase
- No dependency management or version conflicts

**Why Not Wrap/Fork?**
- ❌ **Wrapping**: Adds unnecessary abstraction layer
- ❌ **Forking**: Creates maintenance burden and update conflicts
- ✅ **Copy-paste**: Full ownership, easy customization, no dependencies

### 5. Implementation Plan
```
web/components/ui/
├── button/
│   ├── button.templ
│   ├── args.go
│   └── variants.go
├── card/
├── dialog/
└── ...
```

**Steps:**
1. Set up Go/templ tooling
2. Copy specific DatastarUI components as needed
3. Create layout components using copied UI components
4. Convert existing pages to use new component system
5. Extract handlers from main server file
