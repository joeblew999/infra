package docs

import (
	"bytes"
	"embed"
	"html/template"
	"sort"
	"strings"

	"github.com/joeblew999/infra/pkg/config"
	"github.com/yuin/goldmark"
	goldmark_parser "github.com/yuin/goldmark/parser"
	goldmark_renderer_html "github.com/yuin/goldmark/renderer/html"
)

//go:embed templates/layout.html
var templates embed.FS

// Renderer handles markdown to HTML conversion
type Renderer struct {
	md       goldmark.Markdown
	template *template.Template
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

	tpl, err := template.ParseFS(templates, "templates/layout.html")
	if err != nil {
		panic(err) // Should not happen in production
	}

	return &Renderer{md: md, template: tpl}
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
	navTree := r.buildNavTree(nav)
	navHTML := r.renderNavigation(navTree, currentPath)
	breadcrumbs := r.renderBreadcrumbs(currentPath)

	data := struct {
		Title       string
		NavHTML     template.HTML
		Breadcrumbs template.HTML
		Content     template.HTML
		GitHash     string
		ShortHash   string
	}{
		Title:       title,
		NavHTML:     template.HTML(navHTML),
		Breadcrumbs: template.HTML(breadcrumbs),
		Content:     template.HTML(htmlContent),
		GitHash:     config.GitHash,
		ShortHash:   config.GetShortHash(),
	}

	var buf bytes.Buffer
	if err := r.template.Execute(&buf, data); err != nil {
		return "Error rendering template: " + err.Error()
	}

	return buf.String()
}

type NavNode struct {
	Title    string
	Path     string
	Children []*NavNode
}

func (r *Renderer) buildNavTree(nav []NavItem) *NavNode {
	root := &NavNode{Title: "Documentation"}
	nodes := map[string]*NavNode{"": root}

	for _, item := range nav {
		parts := strings.Split(item.Path, "/")
		for i := 1; i <= len(parts); i++ {
			path := strings.Join(parts[:i], "/")
			if _, ok := nodes[path]; !ok {
				parentPath := ""
				if i > 1 {
					parentPath = strings.Join(parts[:i-1], "/")
				}
				parentNode := nodes[parentPath]
				var node *NavNode
				if i == len(parts) {
					node = &NavNode{Title: item.Title, Path: item.Path}
				} else {
					node = &NavNode{Title: r.fileNameToReadable(parts[i-1]), Path: path}
				}
				nodes[path] = node
				parentNode.Children = append(parentNode.Children, node)
			}
		}
	}

	for _, node := range nodes {
		sort.Slice(node.Children, func(i, j int) bool {
			return node.Children[i].Title < node.Children[j].Title
		})
	}

	return root
}

func (r *Renderer) renderNavigation(node *NavNode, currentPath string) string {
	var sb strings.Builder
	sb.WriteString("<ul>")
	for _, child := range node.Children {
		activeClass := ""
		if child.Path == currentPath {
			activeClass = " class=\"active\""
		}

		icon := "📄"
		if len(child.Children) > 0 {
			icon = "📁"
		}

		sb.WriteString("<li><a href=\"/docs/")
		sb.WriteString(child.Path)
		sb.WriteString("\"" + activeClass + ">" + icon + " ")
		sb.WriteString(child.Title)
		sb.WriteString("</a>")

		if len(child.Children) > 0 {
			sb.WriteString(r.renderNavigation(child, currentPath))
		}
		sb.WriteString("</li>")
	}
	sb.WriteString("</ul>")
	return sb.String()
}

// renderBreadcrumbs renders breadcrumb navigation
func (r *Renderer) renderBreadcrumbs(currentPath string) string {
	if currentPath == "" || currentPath == "README.md" {
		return "Documentation"
	}
	
	var sb strings.Builder
	parts := strings.Split(currentPath, "/")
	
	sb.WriteString("<a href=\"/docs/\">Documentation</a>")
	
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
			sb.WriteString(r.fileNameToReadable(part))
		} else {
			// Folder part - make it a link
			sb.WriteString(" / ")
			sb.WriteString("<a href=\"/docs/")
			sb.WriteString(currentUrlPath)
			sb.WriteString("/")
					
			// Use friendly folder names
			switch part {
			case "business":
				sb.WriteString("Business")
			case "technical":
				sb.WriteString("Technical") 
			case "examples":
				sb.WriteString("Examples")
			default:
				sb.WriteString(strings.ToUpper(part[:1]) + part[1:])
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
		return "AI-Powered Features"
	case "AI_MCP_INTEGRATION": 
		return "AI MCP Integration"
	case "MOBILE_AI_STRATEGY":
		return "Mobile AI Strategy"
	case "BRANDING_GUIDE":
		return "Branding Guide"
	case "WORKFLOWS":
		return "Workflows"
	case "EVERYTHING_AS_GO":
		return "Everything-as-Go-Import"
	case "BETA_TESTING":
		return "Beta Testing"
	case "SCALING":
		return "Scaling"
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
