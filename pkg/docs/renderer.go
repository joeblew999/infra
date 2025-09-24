package docs

import (
	"bytes"
	"embed"
	"html"
	"html/template"
	"path"
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
	ancestors := r.navAncestors(currentPath)
	return r.renderNavigationLevel(node.Children, currentPath, ancestors, 0)
}

func (r *Renderer) renderNavigationLevel(children []*NavNode, currentPath string, ancestors map[string]struct{}, depth int) string {
	if len(children) == 0 {
		return ""
	}

	var sb strings.Builder
	ulClass := "docs-nav__list"
	if depth > 0 {
		ulClass += " docs-nav__list--nested"
	}
	sb.WriteString(`<ul class="` + ulClass + `">`)

	for _, child := range children {
		if child == nil {
			continue
		}

		pathKey := strings.TrimPrefix(child.Path, "/")
		url := "/docs/"
		switch {
		case pathKey == "":
			// already set to /docs/
		case len(child.Children) > 0 && !strings.HasSuffix(pathKey, ".md"):
			url += pathKey + "/"
		default:
			url += pathKey
		}

		normalizedChildKey := strings.TrimSuffix(pathKey, "/")
		isCurrent := child.Path == currentPath

		isAncestor := false
		if !isCurrent && len(child.Children) > 0 {
			candidate := normalizedChildKey
			if candidate == "" && pathKey != "" {
				candidate = pathKey
			}
			if currentPath == candidate+"/README.md" {
				isAncestor = true
			}
			if !isAncestor {
				if _, ok := ancestors[candidate]; ok {
					isAncestor = true
				}
			}
		}

		liClass := "docs-nav__item"
		if isCurrent {
			liClass += " docs-nav__item--current"
		} else if isAncestor {
			liClass += " docs-nav__item--ancestor"
		}

		linkClass := "docs-nav__link"
		if isCurrent {
			linkClass += " docs-nav__link--current"
		}

		icon := "📄"
		if len(child.Children) > 0 {
			icon = "📁"
		}

		sb.WriteString(`<li class="` + liClass + `">`)
		sb.WriteString(`<a href="` + url + `" class="` + linkClass + `">`)
		sb.WriteString(`<span class="docs-nav__icon">` + icon + `</span>`)
		sb.WriteString(`<span>` + html.EscapeString(child.Title) + `</span>`)
		sb.WriteString(`</a>`)
		sb.WriteString(r.renderNavigationLevel(child.Children, currentPath, ancestors, depth+1))
		sb.WriteString(`</li>`)
	}

	sb.WriteString(`</ul>`)
	return sb.String()
}

// renderBreadcrumbs renders breadcrumb navigation
func (r *Renderer) renderBreadcrumbs(currentPath string) string {
	const separator = `<span class="breadcrumbs-sep">/</span>`

	if currentPath == "" || currentPath == "README.md" {
		return `<span class="breadcrumbs-current">Documentation</span>`
	}

	var sb strings.Builder
	sb.WriteString(`<a href="/docs/">Documentation</a>`)

	segments := strings.Split(strings.TrimPrefix(currentPath, "/"), "/")
	var currentURL strings.Builder

	for i, part := range segments {
		isLast := i == len(segments)-1
		name := part
		currentURL.WriteString(part)

		if part == "README.md" {
			continue
		}

		sb.WriteString(separator)

		if isLast {
			sb.WriteString(`<span class="breadcrumbs-current">` + html.EscapeString(r.fileNameToReadable(name)) + `</span>`)
			continue
		}

		friendly := r.folderNameToReadable(part)
		sb.WriteString(`<a href="/docs/` + currentURL.String() + `/">` + html.EscapeString(friendly) + `</a>`)
		currentURL.WriteString("/")
	}

	return sb.String()
}

func (r *Renderer) navAncestors(currentPath string) map[string]struct{} {
	ancestors := make(map[string]struct{})
	clean := strings.TrimPrefix(currentPath, "/")

	if clean == "" {
		return ancestors
	}

	if strings.HasSuffix(clean, "README.md") {
		clean = strings.TrimSuffix(clean, "README.md")
		clean = strings.TrimSuffix(clean, "/")
	} else if strings.HasSuffix(clean, ".md") {
		dir := path.Dir(clean)
		if dir != "." {
			clean = dir
		} else {
			clean = ""
		}
	}

	if clean == "" {
		return ancestors
	}

	parts := strings.Split(clean, "/")
	var current strings.Builder
	for i, part := range parts {
		if part == "" || part == "." {
			continue
		}
		if i > 0 && current.Len() > 0 {
			current.WriteString("/")
		}
		current.WriteString(part)
		ancestors[current.String()] = struct{}{}
	}

	return ancestors
}

func (r *Renderer) folderNameToReadable(name string) string {
	switch name {
	case "business":
		return "Business"
	case "technical":
		return "Technical"
	case "examples":
		return "Examples"
	default:
		if len(name) == 0 {
			return name
		}
		return strings.ToUpper(name[:1]) + name[1:]
	}
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
