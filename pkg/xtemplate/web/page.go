package web

import (
	"bytes"
	_ "embed"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"sort"

	"github.com/joeblew999/infra/pkg/config"
	"github.com/joeblew999/infra/pkg/webapp/templates"
)

//go:embed templates/overview.html
var overviewTemplate string

var overviewTpl = template.Must(template.New("xtemplate-overview").Parse(overviewTemplate))

func init() {
	templates.RegisterNavItem(templates.NavItem{
		Href:  "/xtemplate",
		Text:  "XTemplate",
		Icon:  "🧩",
		Color: "pink",
		Order: 45,
	})
}

// RenderOverviewPage renders the xtemplate overview page using shared chrome.
func RenderOverviewPage() (string, error) {
	templateDir := config.GetXTemplatePath()
	files := discoverTemplates(templateDir)

	data := struct {
		Header        template.HTML
		DataStar      template.HTML
		Navigation    template.HTML
		Footer        template.HTML
		ProxyURL      string
		TemplateDir   string
		ListenAddr    string
		BinaryPath    string
		TemplateFiles []string
	}{
		Header:        templates.GetHeaderHTML(),
		DataStar:      templates.GetDataStarScript(),
		ProxyURL:      "/xtemplate/",
		TemplateDir:   templateDir,
		ListenAddr:    fmt.Sprintf("0.0.0.0:%s", config.GetXTemplatePort()),
		BinaryPath:    config.GetXTemplateBinPath(),
		TemplateFiles: files,
	}

	navHTML, err := templates.RenderNav("/xtemplate")
	if err == nil {
		data.Navigation = template.HTML(navHTML)
	}

	footerHTML, err := templates.RenderFooter()
	if err == nil {
		data.Footer = template.HTML(footerHTML)
	}

	var content bytes.Buffer
	if err := overviewTpl.Execute(&content, data); err != nil {
		return "", fmt.Errorf("render xtemplate overview content: %w", err)
	}

	return templates.RenderBasePage("XTemplate", content.String(), "/xtemplate")
}

func discoverTemplates(root string) []string {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil
	}

	var files []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if filepath.Ext(name) == ".html" || filepath.Ext(name) == ".templ" {
			files = append(files, name)
		}
	}
	sort.Strings(files)
	if len(files) > 8 {
		files = append(files[:8], "…")
	}
	return files
}
