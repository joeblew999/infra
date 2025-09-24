package caddy

import (
	"fmt"
	"strings"
	"testing"

	infraConfig "github.com/joeblew999/infra/pkg/config"
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
			name:           "Development preset includes core developer services",
			preset:         PresetDevelopment,
			port:           8080,
			expectedRoutes: 4,
			requiredPaths:  []string{"/bento-playground/*", "/xtemplate/*", "/docs-hugo/*", "/docs/*"},
		},
		{
			name:           "Full preset includes bento, MCP and docs",
			preset:         PresetFull,
			port:           8080,
			expectedRoutes: 3,
			requiredPaths:  []string{"/bento-playground/*", "/mcp/*", "/docs/*"},
		},
		{
			name:           "Microservices preset has base microservice routes plus infrastructure",
			preset:         PresetMicroservices,
			port:           8080,
			expectedRoutes: 8,
			requiredPaths:  []string{"/api/*", "/auth/*", "/static/*", "/ws/*", "/docs/*", "/docs-hugo/*", "/bento-playground/*", "/mcp/*"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := NewPresetConfig(tt.preset, tt.port)

			if cfg.Port != tt.port {
				t.Errorf("Expected port %d, got %d", tt.port, cfg.Port)
			}

			if len(cfg.Routes) != tt.expectedRoutes {
				t.Errorf("Expected %d routes, got %d", tt.expectedRoutes, len(cfg.Routes))
			}

			if len(tt.requiredPaths) > 0 {
				for _, required := range tt.requiredPaths {
					found := false
					for _, route := range cfg.Routes {
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
	cfg := NewPresetConfig(PresetSimple, 8080).
		WithMainTarget("localhost:3000").
		AddBentoPlayground().
		AddMCPServer().
		AddService("/custom/*", "localhost:9000")

	if cfg.Port != 8080 {
		t.Errorf("Expected port 8080, got %d", cfg.Port)
	}

	if cfg.Target != "localhost:3000" {
		t.Errorf("Expected target localhost:3000, got %s", cfg.Target)
	}

	if len(cfg.Routes) != 3 {
		t.Errorf("Expected 3 routes, got %d", len(cfg.Routes))
	}

	expectedRoutes := []string{"/bento-playground/*", "/mcp/*", "/custom/*"}
	for _, expectedPath := range expectedRoutes {
		found := false
		for _, route := range cfg.Routes {
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
	cfg := CaddyConfig{
		Port:   8080,
		Target: defaultMainTarget(),
		Routes: []ProxyRoute{
			{Path: "/api/*", Target: "localhost:4000"},
			{Path: "/auth/*", Target: "localhost:5000"},
		},
	}

	caddyfile := GenerateCaddyfile(cfg)

	if !strings.Contains(caddyfile, "8080") {
		t.Error("Caddyfile should contain port 8080")
	}

	if !strings.Contains(caddyfile, "reverse_proxy "+defaultMainTarget()) {
		t.Error("Caddyfile should contain main target")
	}

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
	cfg := NewPresetConfig(PresetFull, 8080)

	bentoFound := false
	mcpFound := false

	for _, route := range cfg.Routes {
		if route.Path == "/bento-playground/*" && route.Target == "localhost:"+infraConfig.GetBentoPort() {
			bentoFound = true
		}
		if route.Path == "/mcp/*" && route.Target == "localhost:"+infraConfig.GetMCPPort() {
			mcpFound = true
		}
	}

	if !bentoFound {
		t.Errorf("Bento playground should be at /bento-playground/* -> localhost:%s", infraConfig.GetBentoPort())
	}

	if !mcpFound {
		t.Errorf("MCP server should be at /mcp/* -> localhost:%s", infraConfig.GetMCPPort())
	}

	if cfg.Target != "localhost:"+infraConfig.GetWebServerPort() {
		t.Errorf("Default main target should be localhost:%s, got %s", infraConfig.GetWebServerPort(), cfg.Target)
	}
}

func TestBackwardCompatibility(t *testing.T) {
	// Test that old GenerateCaddyfileSimple still works
	caddyfile := GenerateCaddyfileSimple(8080, 1337)

	if !strings.Contains(caddyfile, "8080") {
		t.Error("Legacy function should work with port 8080")
	}

	if !strings.Contains(caddyfile, fmt.Sprintf("reverse_proxy localhost:%d", 1337)) {
		t.Error("Legacy function should work with target 1337")
	}

	if !strings.Contains(caddyfile, "/bento-playground/*") {
		t.Error("Legacy function should include bento playground")
	}
}
