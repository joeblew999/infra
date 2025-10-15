package tui

import "strings"

// View models a simplified TUI rendering shared across runtime/services.
type View struct {
	Title string
	Lines []string
}

// Render flattens the view into a single string for snapshot style tests.
func Render(v View) string {
	var b strings.Builder
	if v.Title != "" {
		b.WriteString(v.Title)
		b.WriteString("\n")
	}
	for i, line := range v.Lines {
		b.WriteString(line)
		if i != len(v.Lines)-1 {
			b.WriteString("\n")
		}
	}
	return b.String()
}
