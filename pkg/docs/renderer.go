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
func (r *Renderer) RenderToHTMLPage(title string, htmlContent string, nav []NavItem, currentPath string) string {
	navHTML := r.renderNavigation(nav)
	breadcrumbs := r.renderBreadcrumbs(currentPath)
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
		.content { flex: 1; }
		.breadcrumbs { padding: 10px 0; border-bottom: 1px solid #eee; margin-bottom: 20px; color: #666; }
		.breadcrumbs a { color: #007acc; text-decoration: none; }
		.breadcrumbs a:hover { text-decoration: underline; }
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
	<div class="content">
		<div class="breadcrumbs">` + breadcrumbs + `</div>
		` + htmlContent + `
	</div>
</body>
</html>`
}

// renderNavigation renders the navigation menu HTML
func (r *Renderer) renderNavigation(nav []NavItem) string {
	var sb strings.Builder
	
	// Add main navigation
	sb.WriteString("<h3>Main Navigation</h3>")
	sb.WriteString("<ul>")
	sb.WriteString("<li><a href=\"/\">🏠 Home</a></li>")
	sb.WriteString("</ul>")
	
	// Add documentation navigation
	if len(nav) == 0 {
		sb.WriteString("<h3>Documentation</h3><p>No documents found.</p>")
		return sb.String()
	}

	sb.WriteString("<h3>Documentation</h3>")
	
	// Group nav items by folder
	folders := make(map[string][]NavItem)
	var rootItems []NavItem
	
	for _, item := range nav {
		parts := strings.Split(item.Path, "/")
		if len(parts) > 1 {
			// This is in a folder
			folderName := parts[0]
			folders[folderName] = append(folders[folderName], item)
		} else {
			// This is a root item
			rootItems = append(rootItems, item)
		}
	}
	
	// Render root items first
	if len(rootItems) > 0 {
		sb.WriteString("<ul>")
		for _, item := range rootItems {
			sb.WriteString("<li><a href=\"/docs/")
			sb.WriteString(html.EscapeString(item.Path))
			sb.WriteString("\">📄 ")
			sb.WriteString(html.EscapeString(item.Title))
			sb.WriteString("</a></li>")
		}
		sb.WriteString("</ul>")
	}
	
	// Render folders
	folderOrder := []string{"business", "technical", "examples"} // Preferred order
	renderedFolders := make(map[string]bool)
	
	// Render folders in preferred order
	for _, folderName := range folderOrder {
		if items, exists := folders[folderName]; exists {
			r.renderFolder(&sb, folderName, items)
			renderedFolders[folderName] = true
		}
	}
	
	// Render remaining folders
	for folderName, items := range folders {
		if !renderedFolders[folderName] {
			r.renderFolder(&sb, folderName, items)
		}
	}
	
	return sb.String()
}

// renderFolder renders a folder section in navigation
func (r *Renderer) renderFolder(sb *strings.Builder, folderName string, items []NavItem) {
	var folderIcon string
	var folderTitle string
	
	switch folderName {
	case "business":
		folderIcon = "👔"
		folderTitle = "Business Documentation"
	case "technical":
		folderIcon = "👨‍💻"
		folderTitle = "Technical Documentation"
	case "examples":
		folderIcon = "📁"
		folderTitle = "Examples"
	default:
		folderIcon = "📁"
		folderTitle = strings.ToUpper(folderName[:1]) + folderName[1:]
	}
	
	sb.WriteString("<h4>")
	sb.WriteString("<a href=\"/docs/")
	sb.WriteString(html.EscapeString(folderName))
	sb.WriteString("/\">")
	sb.WriteString(folderIcon)
	sb.WriteString(" ")
	sb.WriteString(html.EscapeString(folderTitle))
	sb.WriteString("</a>")
	sb.WriteString("</h4>")
	sb.WriteString("<ul style=\"margin-left: 15px;\">")
	
	for _, item := range items {
		// Don't show README.md files in folder listing (they're the folder index)
		fileName := strings.Split(item.Path, "/")[len(strings.Split(item.Path, "/"))-1]
		if fileName == "README.md" {
			continue
		}
		
		sb.WriteString("<li><a href=\"/docs/")
		sb.WriteString(html.EscapeString(item.Path))
		sb.WriteString("\">📄 ")
		sb.WriteString(html.EscapeString(item.Title))
		sb.WriteString("</a></li>")
	}
	
	sb.WriteString("</ul>")
}

// renderBreadcrumbs renders breadcrumb navigation
func (r *Renderer) renderBreadcrumbs(currentPath string) string {
	if currentPath == "" || currentPath == "README.md" {
		return "📚 Documentation"
	}
	
	var sb strings.Builder
	parts := strings.Split(currentPath, "/")
	
	sb.WriteString("<a href=\"/docs/\">📚 Documentation</a>")
	
	// Build path incrementally
	currentUrlPath := ""
	for i, part := range parts {
		currentUrlPath += part
		
		if i == len(parts)-1 {
			// Last part (current file) - don't make it a link, but show friendly name
			if part == "README.md" {
				// Don't show README.md in breadcrumbs for folders
				continue
			}
			sb.WriteString(" / ")
			sb.WriteString(html.EscapeString(r.fileNameToReadable(part)))
		} else {
			// Folder part - make it a link
			sb.WriteString(" / ")
			sb.WriteString("<a href=\"/docs/")
			sb.WriteString(html.EscapeString(currentUrlPath))
			sb.WriteString("/\">")
			
			// Use friendly folder names
			switch part {
			case "business":
				sb.WriteString("👔 Business")
			case "technical":
				sb.WriteString("👨‍💻 Technical") 
			case "examples":
				sb.WriteString("📁 Examples")
			default:
				sb.WriteString("📁 ")
				sb.WriteString(html.EscapeString(strings.ToUpper(part[:1]) + part[1:]))
			}
			
			sb.WriteString("</a>")
			currentUrlPath += "/"
		}
	}
	
	return sb.String()
}

// fileNameToReadable converts a filename to readable format
func (r *Renderer) fileNameToReadable(filename string) string {
	// Remove extension
	name := strings.TrimSuffix(filename, ".md")
	
	// Convert common patterns
	name = strings.ReplaceAll(name, "_", " ")
	name = strings.ReplaceAll(name, "-", " ")
	
	// Handle specific filenames
	switch name {
	case "AI_POWERED":
		return "🧠 AI-Powered Features"
	case "AI_MCP_INTEGRATION": 
		return "🌐 AI MCP Integration"
	case "MOBILE_AI_STRATEGY":
		return "📱 Mobile AI Strategy"
	case "BRANDING_GUIDE":
		return "🎨 Branding Guide"
	case "WORKFLOWS":
		return "🔄 Workflows"
	case "EVERYTHING_AS_GO":
		return "📦 Everything-as-Go-Import"
	case "BETA_TESTING":
		return "🧪 Beta Testing"
	case "SCALING":
		return "📈 Scaling"
	default:
		// Capitalize first letter of each word
		words := strings.Fields(name)
		for i, word := range words {
			if len(word) > 0 {
				words[i] = strings.ToUpper(word[:1]) + word[1:]
			}
		}
		return strings.Join(words, " ")
	}
}
