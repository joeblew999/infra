package webtemplates

import (
	"embed"
	"fmt"
	"html/template"
	"net/url"
	"strings"
)

//go:embed *.html components/*.html
var files embed.FS

var funcMap = template.FuncMap{
	"upper": strings.ToUpper,
	"join":  strings.Join,
	"urlquery": func(value string) string {
		return url.QueryEscape(value)
	},
}

func Parse() (*template.Template, error) {
	tmpl, err := template.New("page.html").Funcs(funcMap).ParseFS(files, "components/*.html", "page.html")
	if err != nil {
		return nil, fmt.Errorf("parse web template: %w", err)
	}
	return tmpl, nil
}
