package tui

import "testing"

func TestRender(t *testing.T) {
	out := Render(View{Title: "Demo", Lines: []string{"one", "two"}})
	if out == "" {
		t.Fatal("expected output")
	}
	if out[:4] != "Demo" {
		t.Fatalf("unexpected output: %s", out)
	}
}
