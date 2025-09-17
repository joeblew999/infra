package caddy

import (
	"strings"
	"testing"
)

func TestPresetConfigurations(t *testing.T) {
	tests := []struct {
		name           string
		preset         Preset
		port           int
		expectedRoutes int
		requiredPaths  []string
	}{
		{
			name:           "Simple preset has no additional routes",
			preset:         PresetSimple,
			port:           8080,
			expectedRoutes: 0,
			requiredPaths:  nil,
		},
		{
			name:           "Development preset includes bento playground and xtemplate",
			preset:         PresetDevelopment,
			port:           8080,
			expectedRoutes: 2,
			requiredPaths:  []string{"/bento-playground/*", "/xtemplate/*"},
		},
		{
			name:           "Full preset includes bento and MCP",
			preset:         PresetFull,
			port:           8080,
			expectedRoutes: 2,
			requiredPaths:  []string{"/bento-playground/*", "/mcp/*"},
		},
		{
			name:           "Microservices preset has 4 routes",
			preset:         PresetMicroservices,
			port:           8080,
			expectedRoutes: 4,
			requiredPaths:  []string{"/api/*", "/auth/*", "/static/*", "/ws/*"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := NewPresetConfig(tt.preset, tt.port)

			if config.Port != tt.port {
				t.Errorf("Expected port %d, got %d", tt.port, config.Port)
			}

			if len(config.Routes) != tt.expectedRoutes {
				t.Errorf("Expected %d routes, got %d", tt.expectedRoutes, len(config.Routes))
			}

			if len(tt.requiredPaths) > 0 {
				for _, required := range tt.requiredPaths {
					found := false
					for _, route := range config.Routes {
						if route.Path == required {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("Expected route %s not found", required)
					}
				}
			}
		})
	}
}

func TestFluentAPI(t *testing.T) {
	config := NewPresetConfig(PresetSimple, 8080).
		WithMainTarget("localhost:3000").
		AddBentoPlayground().
		AddMCPServer().
		AddService("/custom/*", "localhost:9000")

	// Check port
	if config.Port != 8080 {
		t.Errorf("Expected port 8080, got %d", config.Port)
	}

	// Check main target
	if config.Target != "localhost:3000" {
		t.Errorf("Expected target localhost:3000, got %s", config.Target)
	}

	// Check routes were added (simple + bento + MCP + custom = 3)
	if len(config.Routes) != 3 {
		t.Errorf("Expected 3 routes, got %d", len(config.Routes))
	}

	// Check specific routes exist
	expectedRoutes := []string{"/bento-playground/*", "/mcp/*", "/custom/*"}
	for _, expectedPath := range expectedRoutes {
		found := false
		for _, route := range config.Routes {
			if route.Path == expectedPath {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected route %s not found", expectedPath)
		}
	}
}

func TestCaddyfileGeneration(t *testing.T) {
	config := CaddyConfig{
		Port:   8080,
		Target: "localhost:1337",
		Routes: []ProxyRoute{
			{Path: "/api/*", Target: "localhost:4000"},
			{Path: "/auth/*", Target: "localhost:5000"},
		},
	}

	caddyfile := GenerateCaddyfile(config)

	// Check basic structure
	if !strings.Contains(caddyfile, "8080") {
		t.Error("Caddyfile should contain port 8080")
	}

	if !strings.Contains(caddyfile, "reverse_proxy localhost:1337") {
		t.Error("Caddyfile should contain main target")
	}

	// Check routes
	if !strings.Contains(caddyfile, "handle /api/*") {
		t.Error("Caddyfile should contain API route")
	}

	if !strings.Contains(caddyfile, "reverse_proxy localhost:4000") {
		t.Error("Caddyfile should contain API target")
	}

	if !strings.Contains(caddyfile, "handle /auth/*") {
		t.Error("Caddyfile should contain auth route")
	}
}

func TestInfrastructureConstants(t *testing.T) {
	// Test that infrastructure services use consistent ports/paths
	config := NewPresetConfig(PresetFull, 8080)

	// Check bento playground
	bentoFound := false
	mcpFound := false

	for _, route := range config.Routes {
		if route.Path == "/bento-playground/*" && route.Target == "localhost:4195" {
			bentoFound = true
		}
		if route.Path == "/mcp/*" && route.Target == "localhost:8080" {
			mcpFound = true
		}
	}

	if !bentoFound {
		t.Error("Bento playground should be at /bento-playground/* -> localhost:4195")
	}

	if !mcpFound {
		t.Error("MCP server should be at /mcp/* -> localhost:8080")
	}

	// Check main target default
	if config.Target != "localhost:1337" {
		t.Errorf("Default main target should be localhost:1337, got %s", config.Target)
	}
}

func TestBackwardCompatibility(t *testing.T) {
	// Test that old GenerateCaddyfileSimple still works
	caddyfile := GenerateCaddyfileSimple(8080, 1337)

	if !strings.Contains(caddyfile, "8080") {
		t.Error("Legacy function should work with port 8080")
	}

	if !strings.Contains(caddyfile, "reverse_proxy localhost:1337") {
		t.Error("Legacy function should work with target 1337")
	}

	if !strings.Contains(caddyfile, "/bento-playground/*") {
		t.Error("Legacy function should include bento playground")
	}
}
