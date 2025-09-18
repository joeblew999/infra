package web

import (
	"bytes"
	_ "embed"
	"fmt"
	"html/template"

	"github.com/joeblew999/infra/pkg/config"
	"github.com/joeblew999/infra/pkg/webapp/templates"
)

func init() {
	templates.RegisterNavItem(templates.NavItem{
		Href:  config.LogsHTTPPath,
		Text:  "Logs",
		Icon:  "📝",
		Color: "orange",
		Order: 50,
	})
}

//go:embed templates/logs-page.html
var logsPageHTML string

var (
	logsPageTpl = template.Must(template.New("logs-page").Parse(logsPageHTML))
)

// RenderLogsPage returns the full logs page HTML with shared chrome.
func RenderLogsPage() (string, error) {
	var content bytes.Buffer
	data := struct {
		BasePath string
	}{BasePath: config.LogsHTTPPath}

	if err := logsPageTpl.Execute(&content, data); err != nil {
		return "", fmt.Errorf("render logs page content: %w", err)
	}

	return templates.RenderBasePage("Logs", content.String(), config.LogsHTTPPath)
}
