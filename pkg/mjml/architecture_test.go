package mjml

import (
	"testing"
)

// TestCacheKeyDeterministic verifies cache keys are deterministic
func TestCacheKeyDeterministic(t *testing.T) {
	renderer := NewRenderer(WithCache(true))
	
	factory := NewTestDataFactory()
	data := factory.SimpleEmailData()
	
	// Create cache key multiple times
	key1, err := renderer.createCacheKey("simple", data)
	if err != nil {
		t.Fatalf("Failed to create cache key: %v", err)
	}
	
	key2, err := renderer.createCacheKey("simple", data)
	if err != nil {
		t.Fatalf("Failed to create second cache key: %v", err)
	}
	
	if key1 != key2 {
		t.Errorf("Cache keys are not deterministic: %s != %s", key1, key2)
	}
	
	t.Logf("Cache key: %s", key1)
}

// TestDataFactoryReducesDuplication tests the test data factory
func TestDataFactoryReducesDuplication(t *testing.T) {
	factory := NewTestDataFactory()
	
	// Test all factory methods
	simple := factory.SimpleEmailData()
	welcome := factory.WelcomeEmailData()
	reset := factory.ResetPasswordData()
	notification := factory.NotificationData()
	
	// Verify data is populated correctly
	if simple.Name == "" || simple.Subject == "" {
		t.Error("Simple email data not properly populated")
	}
	
	if welcome.ActivationURL == "" || welcome.EmailData.CompanyName == "" {
		t.Error("Welcome email data not properly populated")
	}
	
	if reset.ResetURL == "" || reset.ExpiresIn == 0 {
		t.Error("Reset password data not properly populated")
	}
	
	if notification.NotificationType == "" || notification.Priority == "" {
		t.Error("Notification data not properly populated")
	}
	
	t.Logf("✓ All factory methods produce valid data")
}

// TestDefaultTemplateNamesConsistency verifies default template names match
func TestDefaultTemplateNamesConsistency(t *testing.T) {
	names := DefaultTemplateNames()
	templates := DefaultTemplates()
	
	// Verify names list matches template map keys
	if len(names) != len(templates) {
		t.Errorf("Name count (%d) doesn't match template count (%d)", len(names), len(templates))
	}
	
	nameSet := make(map[string]bool)
	for _, name := range names {
		nameSet[name] = true
	}
	
	for templateName := range templates {
		if !nameSet[templateName] {
			t.Errorf("Template %s exists but not in names list", templateName)
		}
	}
	
	t.Logf("✓ Default template names consistent: %v", names)
}

// TestLoadDefaultTemplatesIdempotent verifies loading default templates is idempotent
func TestLoadDefaultTemplatesIdempotent(t *testing.T) {
	renderer := NewRenderer()
	
	// Load default templates multiple times
	err1 := renderer.LoadDefaultTemplates()
	if err1 != nil {
		t.Fatalf("First load failed: %v", err1)
	}
	
	templates1 := renderer.ListTemplates()
	
	err2 := renderer.LoadDefaultTemplates()
	if err2 != nil {
		t.Fatalf("Second load failed: %v", err2)
	}
	
	templates2 := renderer.ListTemplates()
	
	// Should have same templates after multiple loads
	if len(templates1) != len(templates2) {
		t.Errorf("Template count changed: %d != %d", len(templates1), len(templates2))
	}
	
	// Verify all templates still work
	factory := NewTestDataFactory()
	data := factory.SimpleEmailData()
	
	_, err := renderer.RenderTemplate("simple", data)
	if err != nil {
		t.Errorf("Template rendering failed after multiple loads: %v", err)
	}
	
	t.Logf("✓ LoadDefaultTemplates is idempotent")
}

// TestCacheKeyChangesWithData verifies cache keys change when data changes
func TestCacheKeyChangesWithData(t *testing.T) {
	renderer := NewRenderer(WithCache(true))
	
	factory := NewTestDataFactory()
	data1 := factory.SimpleEmailData()
	data2 := factory.SimpleEmailData()
	data2.Name = "Different Name" // Change the data
	
	key1, err := renderer.createCacheKey("simple", data1)
	if err != nil {
		t.Fatalf("Failed to create first cache key: %v", err)
	}
	
	key2, err := renderer.createCacheKey("simple", data2)
	if err != nil {
		t.Fatalf("Failed to create second cache key: %v", err)
	}
	
	if key1 == key2 {
		t.Error("Cache keys should be different when data changes")
	}
	
	t.Logf("✓ Cache keys change appropriately with data: %s vs %s", key1, key2)
}