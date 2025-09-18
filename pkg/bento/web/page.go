package web

import (
	"bytes"
	_ "embed"
	"fmt"
	"html/template"

	"github.com/joeblew999/infra/pkg/webapp/templates"
)

func init() {
	templates.RegisterNavItem(templates.NavItem{
		Href:  "/bento/playground",
		Text:  "Bento",
		Icon:  "🎮",
		Color: "red",
		Order: 40,
	})
}

//go:embed templates/playground-page.html
var playgroundPageHTML string

var (
	playgroundTpl = template.Must(template.New("bento-playground").Parse(playgroundPageHTML))
)

// RenderPlaygroundPage returns the Bento playground HTML wrapped in shared chrome.
func RenderPlaygroundPage() (string, error) {
	var content bytes.Buffer
	if err := playgroundTpl.Execute(&content, nil); err != nil {
		return "", fmt.Errorf("render bento playground content: %w", err)
	}

	return templates.RenderBasePage("Bento Pipeline Builder", content.String(), "/bento/playground")
}
