package tuitemplates

import (
	"embed"
	"fmt"
	"strings"
	"text/template"
)

//go:embed *.tmpl components/*.tmpl
var files embed.FS

var funcMap = template.FuncMap{
	"upper": strings.ToUpper,
	"join":  strings.Join,
	"repeat": func(s string, n int) string {
		return strings.Repeat(s, n)
	},
}

// Parse returns the parsed TUI template.
func Parse() (*template.Template, error) {
	tmpl, err := template.New("layout.tmpl").Funcs(funcMap).ParseFS(files, "components/*.tmpl", "layout.tmpl")
	if err != nil {
		return nil, fmt.Errorf("parse tui template: %w", err)
	}
	return tmpl, nil
}
