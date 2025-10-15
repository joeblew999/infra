package web

import (
	"strings"
	"testing"
)

func TestRender(t *testing.T) {
	html := Render(Page{Title: "Demo", Content: "Hello"})
	if html == "" {
		t.Fatal("expected html output")
	}
	if !strings.HasPrefix(html, "<section") {
		t.Fatalf("unexpected output: %s", html)
	}
}
