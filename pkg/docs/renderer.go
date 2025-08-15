package docs

import (
	"html"
	"strings"

	"github.com/yuin/goldmark"
	goldmark_parser "github.com/yuin/goldmark/parser"
	goldmark_renderer_html "github.com/yuin/goldmark/renderer/html"
)

// Renderer handles markdown to HTML conversion
type Renderer struct {
	md goldmark.Markdown
}

// NewRenderer creates a new markdown renderer
func NewRenderer() *Renderer {
	md := goldmark.New(
		goldmark.WithParserOptions(
			goldmark_parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			goldmark_renderer_html.WithUnsafe(),
		),
	)

	return &Renderer{md: md}
}

// RenderToHTML converts markdown content to HTML
func (r *Renderer) RenderToHTML(markdown []byte) (string, error) {
	var buf strings.Builder
	if err := r.md.Convert(markdown, &buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// RenderToHTMLPage wraps HTML content in a basic HTML page structure
func (r *Renderer) RenderToHTMLPage(title string, htmlContent string, nav []NavItem) string {
	navHTML := r.renderNavigation(nav)
	return `<!DOCTYPE html>
<html>
<head>
	<title>` + title + `</title>
	<meta charset="utf-8">
	<style>
		body { font-family: -apple-system, BlinkMacSystemFont, sans-serif; max-width: 1200px; margin: 0 auto; padding: 20px; display: flex; gap: 20px; }
		.nav { width: 200px; flex-shrink: 0; background: #f8f9fa; padding: 15px; border-radius: 5px; }
		.nav h3 { margin-top: 0; color: #333; }
		.nav ul { list-style: none; padding: 0; margin: 0; }
		.nav li { margin: 5px 0; }
		.nav a { color: #007acc; text-decoration: none; }
		.nav a:hover { text-decoration: underline; }
		.nav .bento-link { color: #dc2626; font-weight: bold; }
		.nav .bento-link:hover { color: #b91c1c; }
		.content { flex: 1; }
		pre, code { background: #f5f5f5; padding: 2px 4px; border-radius: 3px; }
		pre { padding: 10px; overflow-x: auto; }
		
		@media (max-width: 768px) {
			body { flex-direction: column; }
			.nav { width: 100%; }
		}
	</style>
</head>
<body>
	<div class="nav">` + navHTML + `</div>
	<div class="content">` + htmlContent + `</div>
</body>
</html>`
}

// renderNavigation renders the navigation menu HTML
func (r *Renderer) renderNavigation(nav []NavItem) string {
	var sb strings.Builder
	
	// Add bento playground link at the top
	sb.WriteString("<h3>Main Navigation</h3>")
	sb.WriteString("<ul>")
	sb.WriteString("<li><a href=\"/\">🏠 Home</a></li>")
	sb.WriteString("<li><a href=\"/bento-playground\" class=\"bento-link\">🎮 Bento Playground</a></li>")
	sb.WriteString("</ul>")
	
	// Add documentation navigation
	if len(nav) == 0 {
		sb.WriteString("<h3>Documentation</h3><p>No documents found.</p>")
		return sb.String()
	}

	sb.WriteString("<h3>Documentation</h3><ul>")

	for _, item := range nav {
		sb.WriteString("<li><a href=\"/docs/")
		sb.WriteString(html.EscapeString(item.Path))
		sb.WriteString("\">")
		sb.WriteString(html.EscapeString(item.Title))
		sb.WriteString("</a></li>")
	}

	sb.WriteString("</ul>")
	return sb.String()
}
