# Infrastructure Management System - Business Overview

A self-hosted platform for managing web services, document generation, workflow automation, and secure communications.

## What This System Does

This is a Go-based infrastructure management platform that orchestrates several services:

- **Web server** with embedded documentation and status pages
- **Document generation** using templates and brand consistency tools
- **Workflow automation** for business processes
- **Secure messaging** between system components
- **AI integration** via MCP (Model Context Protocol) for various assistants

The system runs as a single binary that manages multiple services through process supervision.

## AI Integration

The system includes MCP (Model Context Protocol) support, which allows AI assistants to:

- Access and modify system components
- Generate documents using existing templates
- Interact with workflow systems
- Query system status and logs

This works with Claude, ChatGPT, and other MCP-compatible AI tools. The integration allows iterative refinement of generated content rather than starting from scratch each time.

## Core Components

### 1. Document Generation (Deck)
A template system for generating consistent branded documents:

- **PDF generation** from templates with your brand assets
- **Web pages** using shared CSS and design components
- **Email templates** with MJML for consistent rendering
- **Presentations** and reports with embedded charts

The system uses a declarative syntax where you define content structure, and it handles the visual formatting automatically.

### 2. Workflow Automation (Bento)  
Stream processing and workflow automation using visual configurations:

- **Data transformation** - Process and route messages between systems
- **Event handling** - Respond to webhooks, file changes, timers
- **API integration** - Connect to external services and databases
- **Content processing** - Generate documents, send emails, update records

Based on Benthos stream processor with a configuration interface.

### 3. Infrastructure Services
Standard infrastructure components with process management:

- **HTTP server** with automatic HTTPS via Caddy
- **Message bus** using NATS for inter-service communication
- **Database** via embedded PocketBase
- **Process supervision** to keep services running
- **Monitoring** and logging for all components

## Use Cases

### When This Makes Sense
- You need consistent document generation without manual design work
- You have repetitive workflows that could be automated
- You want to self-host rather than use SaaS tools
- You need audit trails and process logging
- You're comfortable with technical setup and maintenance

### When It Doesn't
- You only need simple static websites
- Your document needs are minimal or highly custom
- You prefer managed SaaS solutions
- You don't have technical resources for setup/maintenance
- You need enterprise-grade support contracts

## How It Works

### Document Templates (Deck)
The system uses a template language for creating branded documents:

1. **Define your brand assets** - colors, fonts, logos in configuration files
2. **Write content** in a simple markup format
3. **Generate outputs** - PDF, HTML, SVG automatically styled

```
Content + Brand Config = Styled Document
example.dsh + brand.json = example.pdf
```

The template system handles:
- Consistent typography and color schemes
- Logo placement and sizing
- Layout and spacing rules
- Multi-format output (PDF, web, print)

### Email Templates
MJML integration for responsive email design:

- Templates use your brand configuration automatically  
- Responsive design for mobile and desktop
- Compatible with major email providers
- Version control for template changes

## Workflow Automation

### Stream Processing (Bento)
The system includes Benthos for data processing and workflow automation:

- **File processing** - Watch directories, process files as they arrive
- **API integration** - HTTP requests, webhooks, database operations  
- **Message routing** - Transform and route data between systems
- **Scheduled tasks** - Run processes on timers or intervals

Configuration is done through YAML files that define inputs, processors, and outputs. For example:

```yaml
input:
  file:
    paths: [ "./invoices/*.json" ]
pipeline:
  processors:
    - mapping: |
        root.customer = this.client_name
        root.amount = this.total.format_number()
output:
  http_client:
    url: "https://accounting-system.com/api/invoices"
```

### Common Patterns
- **Document generation** - Trigger PDF creation from form submissions
- **Email automation** - Send templated emails based on events
- **Data sync** - Keep systems in sync with periodic data transfers
- **File monitoring** - Process uploads, generate reports automatically

## Security and Infrastructure

### Message Bus (NATS)
Internal services communicate via NATS messaging:

- **Service coordination** - Services can send messages and events
- **Message persistence** - Optional message storage for reliability
- **Authentication** - Access control for message topics
- **Monitoring** - Built-in metrics and logging

### HTTP Server (Caddy)
Web interface and API access through Caddy:

- **Automatic HTTPS** - TLS certificates via Let's Encrypt or internal CA
- **Reverse proxy** - Routes requests to appropriate services
- **Access logging** - HTTP request logging for monitoring
- **Static files** - Serves documentation and web interface

### Data Storage  
- **PocketBase** - Embedded SQLite database for application data
- **File storage** - Local filesystem for templates and generated documents
- **Configuration** - YAML/JSON files for service configuration

## Getting Started

### Installation
1. **Download the binary** or build from source
2. **Run `./infra runtime up`** to start all services
3. **Access web interface** at http://localhost:1337
4. **Check service status** with `infra runtime status` or visit http://localhost:1337/status

### Initial Configuration
- **Brand assets** - Add your logos, colors, fonts to `.data/brand/`
- **Templates** - Create document templates in `.data/deck/`
- **Workflows** - Configure Bento YAML files in `.data/bento/`
- **Users** - Set up authentication via PocketBase admin panel

### Deployment Options
- **Local development** - Run directly on your machine
- **Docker** - Use included containerization
- **Cloud hosting** - Deploy to any VPS or cloud provider
- **On-premise** - Run on your own servers

The system is designed to be self-contained with minimal external dependencies.

## Monitoring and Maintenance

The system includes basic monitoring capabilities:

- **Service status** - Check if all components are running
- **Process logs** - View logs for each service
- **System metrics** - Basic performance and resource usage
- **Health checks** - HTTP endpoints for uptime monitoring

### What to Monitor
- **Document generation** - Check if templates compile without errors
- **Workflow execution** - Monitor Bento pipeline success/failure rates  
- **Storage usage** - Keep an eye on disk space for generated files
- **Service availability** - Ensure all components stay running

## Further Reading

- **[Technical Documentation](../technical/)** - Detailed setup and configuration
- **[API Reference](../api/)** - REST API endpoints and usage
- **[Template Documentation](../templates/)** - Creating and managing templates
- **[Workflow Examples](../workflows/)** - Common automation patterns

This system works best when you have technical resources available for setup and ongoing maintenance. It's designed for organizations that prefer self-hosting over SaaS solutions.
