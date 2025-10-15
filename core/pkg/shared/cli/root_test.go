package cli

import "testing"

func TestNewRootCommandDefaults(t *testing.T) {
	cmd := NewRootCommand(BuilderOptions{})
	if cmd.Use != "core" {
		t.Fatalf("unexpected use: %s", cmd.Use)
	}
	if cmd.HelpTemplate() == "" {
		t.Fatal("expected help template")
	}
}
