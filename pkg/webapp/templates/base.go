package templates

import (
	"html/template"
	"strings"
)

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

// RenderBasePage renders the shared chrome around page content.
func RenderBasePage(title, content, currentPath string) (string, error) {
	navHTML, err := RenderNav(currentPath)
	if err != nil {
		navHTML = ""
	}

	footerHTML, err := RenderFooter()
	if err != nil {
		footerHTML = ""
	}

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
		Navigation: template.HTML(navHTML),
		Content:    template.HTML(content),
		Footer:     template.HTML(footerHTML),
	}

	tmpl, err := template.New("basePage").Parse(baseTemplate)
	if err != nil {
		return "", err
	}

	var result strings.Builder
	if err := tmpl.Execute(&result, data); err != nil {
		return "", err
	}

	return result.String(), nil
}
