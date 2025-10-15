package web

import "strings"

// Page represents a minimal shared layout for web shells.
type Page struct {
	Title   string
	Content string
}

// Render returns a deterministic HTML fragment used by runtime/services.
func Render(p Page) string {
	var b strings.Builder
	b.WriteString("<section class=\"core-page\">")
	if p.Title != "" {
		b.WriteString("<h1>")
		b.WriteString(p.Title)
		b.WriteString("</h1>")
	}
	if p.Content != "" {
		b.WriteString("<div class=\"core-content\">")
		b.WriteString(p.Content)
		b.WriteString("</div>")
	}
	b.WriteString("</section>")
	return b.String()
}
