package docs

import (
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
func (r *Renderer) RenderToHTMLPage(title string, htmlContent string) string {
	return `<!DOCTYPE html>
<html>
<head>
	<title>` + title + `</title>
	<meta charset="utf-8">
	<style>
		body { font-family: -apple-system, BlinkMacSystemFont, sans-serif; max-width: 800px; margin: 0 auto; padding: 20px; }
		pre, code { background: #f5f5f5; padding: 2px 4px; border-radius: 3px; }
		pre { padding: 10px; overflow-x: auto; }
	</style>
</head>
<body>` + htmlContent + `</body>
</html>`
}
