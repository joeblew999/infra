package templates

import (
	_ "embed"
	"html/template"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/joeblew999/infra/pkg/config"
)

//go:embed nav.html
var navTemplateHTML string

var (
	navOnce     sync.Once
	navTemplate *template.Template
	navTplErr   error
)

// NavItem represents a navigation menu item.
type NavItem struct {
	Href  string
	Text  string
	Icon  string
	Color string
	Order int
}

// Navigation holds the complete navigation structure.
type Navigation struct {
	Title string
	Items []navItemState
}

type navItemState struct {
	NavItem
	Active bool
}

var (
	navMu           sync.RWMutex
	baseNavItems    = []NavItem{{Href: "/", Text: "Home", Icon: "🏠", Color: "blue", Order: 10}, {Href: config.DocsHTTPPath, Text: "Docs", Icon: "📚", Color: "green", Order: 20}}
	registeredNav   []NavItem
	navigationTitle = "🏗️ Infrastructure Management"
)

// RegisterNavItem registers a navigation entry supplied by feature packages.
func RegisterNavItem(item NavItem) {
	navMu.Lock()
	registeredNav = append(registeredNav, item)
	navMu.Unlock()
}

// GetNavigation returns the navigation model with active state resolved.
func GetNavigation(currentPath string) Navigation {
	items := collectNavItems()

	for i := range items {
		items[i].Active = isActive(items[i].Href, currentPath)
	}

	return Navigation{Title: navigationTitle, Items: items}
}

func collectNavItems() []navItemState {
	navMu.RLock()
	combined := make([]NavItem, 0, len(baseNavItems)+len(registeredNav))
	combined = append(combined, baseNavItems...)
	combined = append(combined, registeredNav...)
	navMu.RUnlock()

	sort.SliceStable(combined, func(i, j int) bool {
		if combined[i].Order == combined[j].Order {
			return combined[i].Text < combined[j].Text
		}
		return combined[i].Order < combined[j].Order
	})

	states := make([]navItemState, len(combined))
	for i, item := range combined {
		states[i] = navItemState{NavItem: item}
	}
	return states
}

func isActive(href, current string) bool {
	if current == href {
		return true
	}
	if href == "/" {
		return current == "/"
	}
	if !strings.HasSuffix(href, "/") {
		href = href + "/"
	}
	if !strings.HasSuffix(current, "/") {
		current = current + "/"
	}
	return strings.HasPrefix(current, href)
}

// RenderNav renders the navigation HTML for a given path.
func RenderNav(currentPath string) (string, error) {
	nav := GetNavigation(currentPath)

	navOnce.Do(func() {
		navTemplate, navTplErr = template.New("nav").Parse(navTemplateHTML)
	})
	if navTplErr != nil {
		return "", navTplErr
	}

	var builder strings.Builder
	if err := navTemplate.Execute(&builder, nav); err != nil {
		return "", err
	}
	return builder.String(), nil
}

// Footer and header helpers remain below.

// GetDataStarScript returns the standard DataStar script tag for consistent loading.
func GetDataStarScript() template.HTML {
	return template.HTML(`<script type="module" defer src="https://cdn.jsdelivr.net/gh/starfederation/datastar@main/bundles/datastar.js"></script>`)
}

// GetHeaderHTML returns the standard HTML header with meta tags and CSS.
func GetHeaderHTML() template.HTML {
	return template.HTML(`<meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <script src="https://unpkg.com/@tailwindcss/browser@4"></script>`)
}

// RenderHeader renders the standard header HTML.
func RenderHeader() (string, error) {
	return string(GetHeaderHTML()), nil
}

// Footer holds the footer information.
type Footer struct {
	Version     string
	GitHash     string
	ShortHash   string
	BuildTime   string
	CurrentTime string
	Environment string
}

// FooterHTML returns the footer HTML template.
const FooterHTML = `
<footer class="mt-8 pt-6 border-t border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-800/50 rounded-lg">
    <div class="px-4 py-3">
        <div class="flex flex-col sm:flex-row justify-between items-center text-xs text-gray-600 dark:text-gray-400 space-y-2 sm:space-y-0">
            <div class="flex flex-col sm:flex-row items-center space-y-1 sm:space-y-0 sm:space-x-4">
                <div class="flex items-center space-x-2">
                    <span class="font-mono bg-gray-200 dark:bg-gray-700 px-2 py-1 rounded">{{.Version}}</span>
                    <span class="text-gray-500">•</span>
                    <span class="font-mono">{{.ShortHash}}</span>
                </div>
                <div class="flex items-center space-x-2">
                    <span class="text-gray-500">Built:</span>
                    <span>{{.BuildTime}}</span>
                </div>
                <div class="flex items-center space-x-2">
                    <span class="text-gray-500">Env:</span>
                    <span class="{{if eq .Environment "production"}}text-red-600 dark:text-red-400{{else}}text-blue-600 dark:text-blue-400{{end}} font-medium">{{.Environment}}</span>
                </div>
            </div>

            <div class="flex items-center space-x-4">
                <div class="flex items-center space-x-2">
                    <span class="w-2 h-2 bg-green-500 rounded-full animate-pulse"></span>
                    <span>Server Online</span>
                </div>
                <div class="text-gray-500">
                    {{.CurrentTime}}
                </div>
            </div>
        </div>

        <div class="mt-2 pt-2 border-t border-gray-200 dark:border-gray-600 flex flex-col sm:flex-row justify-between items-center text-xs text-gray-500 dark:text-gray-500">
            <div>
                🏗️ Infrastructure Management System
            </div>
            <div class="flex items-center space-x-4 mt-1 sm:mt-0">
                <a href="/config/api/build" class="hover:text-gray-700 dark:hover:text-gray-300 transition-colors" target="_blank">
                    📊 Build Info API
                </a>
                <a href="/status" class="hover:text-gray-700 dark:hover:text-gray-300 transition-colors">
                    ⚡ System Status
                </a>
            </div>
        </div>
    </div>
</footer>
`

// GetFooter returns the footer information.
func GetFooter() Footer {
	env := "development"
	if config.IsProduction() {
		env = "production"
	}

	return Footer{
		Version:     config.GetVersion(),
		GitHash:     config.GitHash,
		ShortHash:   config.GetShortHash(),
		BuildTime:   config.BuildTime,
		CurrentTime: time.Now().Format("15:04:05 UTC"),
		Environment: env,
	}
}

// RenderFooter renders the footer HTML.
func RenderFooter() (string, error) {
	footer := GetFooter()

	tmpl, err := template.New("footer").Parse(FooterHTML)
	if err != nil {
		return "", err
	}

	var result strings.Builder
	if err := tmpl.Execute(&result, footer); err != nil {
		return "", err
	}

	return result.String(), nil
}
