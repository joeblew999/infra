package web

import (
	"html/template"
	"strings"
	"time"

	"github.com/joeblew999/infra/pkg/config"
)

// GetDataStarScript returns the standard DataStar script tag for consistent loading
func GetDataStarScript() template.HTML {
	return template.HTML(`<script type="module" defer src="https://cdn.jsdelivr.net/gh/starfederation/datastar@main/bundles/datastar.js"></script>`)
}

// GetHeaderHTML returns the standard HTML header with meta tags and CSS
func GetHeaderHTML() template.HTML {
	return template.HTML(`<meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <script src="https://unpkg.com/@tailwindcss/browser@4"></script>`)
}

// RenderHeader renders the standard header HTML
func RenderHeader() (string, error) {
	return string(GetHeaderHTML()), nil
}

// NavItem represents a navigation menu item
type NavItem struct {
	Href   string
	Text   string
	Icon   string
	Color  string
	Active bool
}

// Navigation holds the complete navigation structure
type Navigation struct {
	Title string
	Items []NavItem
}

// GetNavigation returns the standard navigation menu
func GetNavigation(currentPath string) Navigation {
	items := []NavItem{
		{Href: "/", Text: "Home", Icon: "🏠", Color: "blue"},
		{Href: "/docs/", Text: "Docs", Icon: "📚", Color: "green"},
		{Href: "/bento-playground", Text: "Bento Playground", Icon: "🎮", Color: "red"},
		{Href: "/metrics", Text: "Metrics", Icon: "📊", Color: "purple"},
		{Href: "/logs", Text: "Logs", Icon: "📝", Color: "orange"},
		{Href: "/processes", Text: "Processes", Icon: "🔍", Color: "indigo"},
		{Href: "/config", Text: "Config", Icon: "🛠️", Color: "cyan"},
		{Href: "/auth", Text: "Auth", Icon: "🔐", Color: "emerald"},
		{Href: "/status", Text: "Status", Icon: "⚡", Color: "gray"},
	}

	// Mark current item as active
	for i := range items {
		if items[i].Href == currentPath {
			items[i].Active = true
		} else if currentPath != "/" && items[i].Href != "/" && len(currentPath) > 1 && len(items[i].Href) > 1 {
			// Check if current path starts with the navigation item's href (for nested paths)
			// But only if the nav href is shorter or equal to current path length
			if len(items[i].Href) <= len(currentPath) && currentPath[:len(items[i].Href)] == items[i].Href {
				items[i].Active = true
			}
		}
	}

	return Navigation{
		Title: "🏗️ Infrastructure Management",
		Items: items,
	}
}

// NavHTML returns the responsive navigation HTML template with mobile hamburger menu
const NavHTML = `
<nav class="mb-8 bg-gray-100 dark:bg-gray-800 rounded-lg shadow-md">
    <div class="px-4 py-3">
        <!-- Desktop and Mobile Header -->
        <div class="flex justify-between items-center">
            <h1 class="text-xl font-bold text-gray-900 dark:text-white">
                {{.Title}}
            </h1>
            
            <!-- Hamburger Menu Button (Mobile Only) -->
            <button id="mobile-menu-button" type="button" class="md:hidden inline-flex items-center justify-center p-2 rounded-md text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white hover:bg-gray-200 dark:hover:bg-gray-700 focus:outline-none focus:ring-2 focus:ring-inset focus:ring-blue-500 transition-colors duration-200" aria-controls="mobile-menu" aria-expanded="false">
                <span class="sr-only">Open main menu</span>
                <!-- Hamburger Icon -->
                <svg id="hamburger-icon" class="block h-6 w-6" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M3.75 6.75h16.5M3.75 12h16.5m-16.5 5.25h16.5" />
                </svg>
                <!-- Close Icon (hidden by default) -->
                <svg id="close-icon" class="hidden h-6 w-6" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
                </svg>
            </button>
            
            <!-- Desktop Navigation (Hidden on Mobile) -->
            <div class="hidden md:flex space-x-4">
                {{range .Items}}
                <a href="{{.Href}}" class="px-3 py-1 text-sm font-medium text-{{.Color}}-600 dark:text-{{.Color}}-400 hover:text-{{.Color}}-800 dark:hover:text-{{.Color}}-200 rounded hover:bg-{{.Color}}-50 dark:hover:bg-{{.Color}}-900/20{{if .Active}} bg-{{.Color}}-100 dark:bg-{{.Color}}-900/30{{end}} transition-colors duration-200">
                    {{.Icon}} {{.Text}}
                </a>
                {{end}}
            </div>
        </div>
        
        <!-- Mobile Navigation Menu (Hidden by default) -->
        <div id="mobile-menu" class="hidden md:hidden mt-4 pb-2">
            <div class="space-y-2">
                {{range .Items}}
                <a href="{{.Href}}" class="block px-3 py-2 rounded-md text-base font-medium text-{{.Color}}-600 dark:text-{{.Color}}-400 hover:text-{{.Color}}-800 dark:hover:text-{{.Color}}-200 hover:bg-{{.Color}}-50 dark:hover:bg-{{.Color}}-900/20{{if .Active}} bg-{{.Color}}-100 dark:bg-{{.Color}}-900/30 text-{{.Color}}-800 dark:text-{{.Color}}-200{{end}} transition-colors duration-200">
                    {{.Icon}} {{.Text}}
                </a>
                {{end}}
            </div>
        </div>
    </div>
</nav>

<script>
// Mobile menu toggle functionality
(function() {
    const button = document.getElementById('mobile-menu-button');
    const menu = document.getElementById('mobile-menu');
    const hamburgerIcon = document.getElementById('hamburger-icon');
    const closeIcon = document.getElementById('close-icon');
    
    if (button && menu && hamburgerIcon && closeIcon) {
        button.addEventListener('click', function() {
            const isOpen = menu.classList.contains('hidden');
            
            if (isOpen) {
                // Open menu
                menu.classList.remove('hidden');
                hamburgerIcon.classList.add('hidden');
                closeIcon.classList.remove('hidden');
                button.setAttribute('aria-expanded', 'true');
            } else {
                // Close menu
                menu.classList.add('hidden');
                hamburgerIcon.classList.remove('hidden');
                closeIcon.classList.add('hidden');
                button.setAttribute('aria-expanded', 'false');
            }
        });
        
        // Close mobile menu when clicking on a link
        const mobileLinks = menu.querySelectorAll('a');
        mobileLinks.forEach(function(link) {
            link.addEventListener('click', function() {
                menu.classList.add('hidden');
                hamburgerIcon.classList.remove('hidden');
                closeIcon.classList.add('hidden');
                button.setAttribute('aria-expanded', 'false');
            });
        });
        
        // Close mobile menu when clicking outside
        document.addEventListener('click', function(event) {
            if (!button.contains(event.target) && !menu.contains(event.target)) {
                menu.classList.add('hidden');
                hamburgerIcon.classList.remove('hidden');
                closeIcon.classList.add('hidden');
                button.setAttribute('aria-expanded', 'false');
            }
        });
        
        // Handle escape key
        document.addEventListener('keydown', function(event) {
            if (event.key === 'Escape' && !menu.classList.contains('hidden')) {
                menu.classList.add('hidden');
                hamburgerIcon.classList.remove('hidden');
                closeIcon.classList.add('hidden');
                button.setAttribute('aria-expanded', 'false');
                button.focus(); // Return focus to button
            }
        });
    }
})();
</script>
`

// RenderNav renders the navigation HTML for a given path
func RenderNav(currentPath string) (string, error) {
	nav := GetNavigation(currentPath)

	tmpl, err := template.New("nav").Parse(NavHTML)
	if err != nil {
		return "", err
	}

	var result strings.Builder
	err = tmpl.Execute(&result, nav)
	if err != nil {
		return "", err
	}

	return result.String(), nil
}

// BaseTemplate renders a complete HTML page with consistent structure
func RenderBasePage(title, content string, currentPath string) (string, error) {
	nav, err := RenderNav(currentPath)
	if err != nil {
		nav = ""
	}

	footer, err := RenderFooter()
	if err != nil {
		footer = ""
	}

	const baseTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <title>{{.Title}} - Infrastructure Management</title>
    {{.Header}}
    {{.DataStar}}
</head>
<body class="bg-white dark:bg-gray-900 text-base sm:text-lg max-w-4xl mx-auto px-4 sm:px-6 py-4 sm:py-8">
    {{.Navigation}}
    <div class="bg-white dark:bg-gray-800 text-gray-900 dark:text-white rounded-lg px-4 sm:px-6 py-4 sm:py-8 ring shadow-xl ring-gray-900/5">
        {{.Content}}
    </div>
    {{.Footer}}
</body>
</html>`

	data := struct {
		Title      string
		Header     template.HTML
		DataStar   template.HTML
		Navigation template.HTML
		Content    template.HTML
		Footer     template.HTML
	}{
		Title:      title,
		Header:     GetHeaderHTML(),
		DataStar:   GetDataStarScript(),
		Navigation: template.HTML(nav),
		Content:    template.HTML(content),
		Footer:     template.HTML(footer),
	}

	tmpl, err := template.New("basePage").Parse(baseTemplate)
	if err != nil {
		return "", err
	}

	var result strings.Builder
	err = tmpl.Execute(&result, data)
	if err != nil {
		return "", err
	}

	return result.String(), nil
}

// Footer holds the footer information
type Footer struct {
	Version     string
	GitHash     string
	ShortHash   string
	BuildTime   string
	CurrentTime string
	Environment string
}

// FooterHTML returns the footer HTML template
const FooterHTML = `
<footer class="mt-8 pt-6 border-t border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-800/50 rounded-lg">
    <div class="px-4 py-3">
        <div class="flex flex-col sm:flex-row justify-between items-center text-xs text-gray-600 dark:text-gray-400 space-y-2 sm:space-y-0">
            <!-- Build Info -->
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
            
            <!-- Status Info -->
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
        
        <!-- Additional Info Row -->
        <div class="mt-2 pt-2 border-t border-gray-200 dark:border-gray-600 flex flex-col sm:flex-row justify-between items-center text-xs text-gray-500 dark:text-gray-500">
            <div>
                🏗️ Infrastructure Management System
            </div>
            <div class="flex items-center space-x-4 mt-1 sm:mt-0">
                <a href="/api/build" class="hover:text-gray-700 dark:hover:text-gray-300 transition-colors" target="_blank">
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

// GetFooter returns the footer information
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

// RenderFooter renders the footer HTML
func RenderFooter() (string, error) {
	footer := GetFooter()

	tmpl, err := template.New("footer").Parse(FooterHTML)
	if err != nil {
		return "", err
	}

	var result strings.Builder
	err = tmpl.Execute(&result, footer)
	if err != nil {
		return "", err
	}

	return result.String(), nil
}
