package main

import (
	"strings"
	"testing"

	"github.com/joeblew999/infra/pkg/mjml"
)

func TestTemplateRendering(t *testing.T) {
	renderer := mjml.NewRenderer(
		mjml.WithCache(true),
		mjml.WithDebug(false),
		mjml.WithTemplateDir("../templates"),
	)

	err := renderer.LoadTemplatesFromDir("../templates")
	if err != nil {
		t.Fatalf("Failed to load templates: %v", err)
	}

	templates := renderer.ListTemplates()
	if len(templates) == 0 {
		t.Fatal("No templates loaded")
	}

	// Test simple template
	simpleData := mjml.EmailData{
		Name:    "Test User",
		Subject: "Test Email",
		Title:   "Test",
		Message: "Test message",
		ButtonText: "Test Button",
		ButtonURL:  "https://example.com",
	}

	html, err := renderer.RenderTemplate("simple", simpleData)
	if err != nil {
		t.Fatalf("Failed to render simple template: %v", err)
	}

	if len(html) == 0 {
		t.Error("Generated HTML is empty")
	}

	// Test that HTML contains expected elements
	if !strings.Contains(html, "Test User") {
		t.Error("HTML doesn't contain user name")
	}

	if !strings.Contains(html, "Test message") {
		t.Error("HTML doesn't contain message")
	}
}