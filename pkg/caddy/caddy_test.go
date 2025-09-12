package caddy

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/joeblew999/infra/pkg/config"
)

func TestCaddyConfigGeneration(t *testing.T) {
	// Test config generation and artifact visibility
	cfg := DevelopmentConfig(8080)
	
	// Generate Caddyfile
	caddyfile := GenerateCaddyfile(cfg)
	if caddyfile == "" {
		t.Fatal("Generated Caddyfile is empty")
	}
	
	// Save to test-isolated path 
	testDir := config.GetCaddyPath()
	err := os.MkdirAll(testDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create caddy test directory: %v", err)
	}
	
	testFile := filepath.Join(testDir, "test-config.caddyfile")
	err = os.WriteFile(testFile, []byte(caddyfile), 0644)
	if err != nil {
		t.Fatalf("Failed to write test Caddyfile: %v", err)
	}
	
	t.Logf("✅ Caddyfile generated: %s", testFile)
	t.Logf("📁 Test artifacts in: %s", testDir)
	
	// Verify file exists and is readable
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Fatal("Generated Caddyfile not found")
	}
	
	// Read back and verify content
	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read generated Caddyfile: %v", err)
	}
	
	if len(content) == 0 {
		t.Fatal("Generated Caddyfile is empty")
	}
	
	t.Logf("✅ Caddyfile verified: %d bytes", len(content))
}

func TestCaddyPresets(t *testing.T) {
	testCases := []struct {
		name   string
		preset Preset
		port   int
	}{
		{"Simple", PresetSimple, 8080},
		{"Development", PresetDevelopment, 8081},
		{"Full", PresetFull, 8082},
		{"Microservices", PresetMicroservices, 8083},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := NewPresetConfig(tc.preset, tc.port)
			
			// Generate and save each preset
			testDir := config.GetCaddyPath()
			err := os.MkdirAll(testDir, 0755)
			if err != nil {
				t.Fatalf("Failed to create caddy test directory: %v", err)
			}
			
			filename := filepath.Join(testDir, tc.name+".caddyfile")
			caddyfile := GenerateCaddyfile(cfg)
			
			err = os.WriteFile(filename, []byte(caddyfile), 0644)
			if err != nil {
				t.Fatalf("Failed to write %s preset: %v", tc.name, err)
			}
			
			t.Logf("✅ %s preset saved: %s", tc.name, filename)
			
			// Verify basic content
			if len(caddyfile) < 50 {
				t.Errorf("%s preset too short: %d chars", tc.name, len(caddyfile))
			}
		})
	}
}