package mjml

import (
	"os"
	"strings"
	"testing"
	"time"
)

func TestNewRenderer(t *testing.T) {
	renderer := NewRenderer()
	if renderer == nil {
		t.Fatal("NewRenderer returned nil")
	}
	
	if renderer.templates == nil {
		t.Error("templates map not initialized")
	}
	
	if renderer.cache == nil {
		t.Error("cache map not initialized")
	}
}

func TestRendererWithOptions(t *testing.T) {
	renderer := NewRenderer(
		WithCache(true),
		WithDebug(true),
		WithValidation(false),
		WithTemplateDir("custom"),
	)
	
	if !renderer.options.EnableCache {
		t.Error("EnableCache option not set")
	}
	
	if !renderer.options.EnableDebug {
		t.Error("EnableDebug option not set")
	}
	
	if renderer.options.EnableValidation {
		t.Error("EnableValidation should be false")
	}
	
	if renderer.options.TemplateDir != "custom" {
		t.Error("TemplateDir not set correctly")
	}
}

func TestLoadTemplate(t *testing.T) {
	renderer := NewRenderer()
	
	templateContent := `<mjml><mj-body><mj-section><mj-column><mj-text>Hello {{.name}}!</mj-text></mj-column></mj-section></mj-body></mjml>`
	
	err := renderer.LoadTemplate("test", templateContent)
	if err != nil {
		t.Fatalf("LoadTemplate failed: %v", err)
	}
	
	if !renderer.HasTemplate("test") {
		t.Error("Template not found after loading")
	}
	
	templates := renderer.ListTemplates()
	if len(templates) != 1 || templates[0] != "test" {
		t.Error("ListTemplates returned unexpected result")
	}
}

func TestLoadDefaultTemplates(t *testing.T) {
	renderer := NewRenderer()
	
	err := renderer.LoadDefaultTemplates()
	if err != nil {
		t.Fatalf("LoadDefaultTemplates failed: %v", err)
	}
	
	expectedTemplates := []string{"welcome", "reset_password", "notification", "simple"}
	
	for _, tmpl := range expectedTemplates {
		if !renderer.HasTemplate(tmpl) {
			t.Errorf("Default template %s not found", tmpl)
		}
	}
}

func TestRenderTemplate(t *testing.T) {
	renderer := NewRenderer()
	
	// Load a simple template
	templateContent := `<mjml><mj-body><mj-section><mj-column><mj-text>Hello {{.name}}!</mj-text></mj-column></mj-section></mj-body></mjml>`
	
	err := renderer.LoadTemplate("simple_test", templateContent)
	if err != nil {
		t.Fatalf("LoadTemplate failed: %v", err)
	}
	
	data := map[string]interface{}{
		"name": "John Doe",
	}
	
	html, err := renderer.RenderTemplate("simple_test", data)
	if err != nil {
		t.Fatalf("RenderTemplate failed: %v", err)
	}
	
	if !strings.Contains(html, "Hello John Doe!") {
		t.Error("Template variable not substituted correctly")
	}
	
	if !strings.Contains(html, "<!doctype html>") {
		t.Error("Generated HTML appears malformed")
	}
}

func TestRenderTemplateNotFound(t *testing.T) {
	renderer := NewRenderer()
	
	_, err := renderer.RenderTemplate("nonexistent", nil)
	if err == nil {
		t.Error("Expected error for nonexistent template")
	}
	
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("Unexpected error message: %v", err)
	}
}

func TestRenderString(t *testing.T) {
	renderer := NewRenderer()
	
	mjmlContent := `<mjml><mj-body><mj-section><mj-column><mj-text>Direct render test</mj-text></mj-column></mj-section></mj-body></mjml>`
	
	html, err := renderer.RenderString(mjmlContent)
	if err != nil {
		t.Fatalf("RenderString failed: %v", err)
	}
	
	if !strings.Contains(html, "Direct render test") {
		t.Error("Direct render failed")
	}
}

func TestCacheOperations(t *testing.T) {
	renderer := NewRenderer(WithCache(true))
	
	templateContent := `<mjml><mj-body><mj-section><mj-column><mj-text>Cached {{.value}}</mj-text></mj-column></mj-section></mj-body></mjml>`
	
	err := renderer.LoadTemplate("cache_test", templateContent)
	if err != nil {
		t.Fatalf("LoadTemplate failed: %v", err)
	}
	
	data := map[string]interface{}{
		"value": "test",
	}
	
	// First render
	_, err = renderer.RenderTemplate("cache_test", data)
	if err != nil {
		t.Fatalf("First render failed: %v", err)
	}
	
	if renderer.GetCacheSize() != 1 {
		t.Error("Cache not populated after first render")
	}
	
	// Second render should use cache
	_, err = renderer.RenderTemplate("cache_test", data)
	if err != nil {
		t.Fatalf("Second render failed: %v", err)
	}
	
	// Clear cache
	renderer.ClearCache()
	if renderer.GetCacheSize() != 0 {
		t.Error("Cache not cleared")
	}
}

func TestRemoveTemplate(t *testing.T) {
	renderer := NewRenderer()
	
	templateContent := `<mjml><mj-body><mj-section><mj-column><mj-text>Remove test</mj-text></mj-column></mj-section></mj-body></mjml>`
	
	err := renderer.LoadTemplate("remove_test", templateContent)
	if err != nil {
		t.Fatalf("LoadTemplate failed: %v", err)
	}
	
	if !renderer.HasTemplate("remove_test") {
		t.Error("Template not found after loading")
	}
	
	renderer.RemoveTemplate("remove_test")
	
	if renderer.HasTemplate("remove_test") {
		t.Error("Template still found after removal")
	}
}

func TestEmailDataStructs(t *testing.T) {
	now := time.Now()
	
	emailData := EmailData{
		Name:        "Test User",
		Email:       "test@example.com",
		Subject:     "Test Email",
		Timestamp:   now,
		CompanyName: "Test Company",
		Title:       "Test Title",
		Message:     "Test message",
		Data:        map[string]interface{}{"key": "value"},
	}
	
	if emailData.Name != "Test User" {
		t.Error("EmailData fields not set correctly")
	}
	
	welcomeData := WelcomeEmailData{
		EmailData:     emailData,
		ActivationURL: "https://example.com/activate",
		LoginURL:      "https://example.com/login",
	}
	
	if welcomeData.ActivationURL != "https://example.com/activate" {
		t.Error("WelcomeEmailData fields not set correctly")
	}
}

func TestDefaultTemplatesContent(t *testing.T) {
	templates := DefaultTemplates()
	
	expectedTemplates := []string{"welcome", "reset_password", "notification", "simple"}
	
	for _, name := range expectedTemplates {
		content, exists := templates[name]
		if !exists {
			t.Errorf("Default template %s not found", name)
		}
		
		if !strings.Contains(content, "<mjml>") {
			t.Errorf("Template %s does not contain valid MJML", name)
		}
		
		if !strings.Contains(content, "{{.") {
			t.Errorf("Template %s does not contain template variables", name)
		}
	}
}

// TestLoadTemplateFromFile tests loading from an actual file
func TestLoadTemplateFromFile(t *testing.T) {
	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "test_template_*.mjml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	
	templateContent := `<mjml><mj-body><mj-section><mj-column><mj-text>File test {{.name}}</mj-text></mj-column></mj-section></mj-body></mjml>`
	
	if _, err := tmpFile.WriteString(templateContent); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpFile.Close()
	
	renderer := NewRenderer()
	
	err = renderer.LoadTemplateFromFile("file_test", tmpFile.Name())
	if err != nil {
		t.Fatalf("LoadTemplateFromFile failed: %v", err)
	}
	
	if !renderer.HasTemplate("file_test") {
		t.Error("Template not loaded from file")
	}
}