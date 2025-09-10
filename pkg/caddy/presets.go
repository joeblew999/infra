package caddy

import (
	"os"
	"path/filepath"
	
	"github.com/joeblew999/infra/pkg/config"
)

// Preset represents a common Caddy configuration scenario
type Preset int

const (
	// PresetSimple serves a single application
	PresetSimple Preset = iota
	
	// PresetDevelopment includes main app + bento playground
	PresetDevelopment
	
	// PresetFull includes main app + bento + MCP server
	PresetFull
	
	// PresetMicroservices provides a base for custom microservice routing
	PresetMicroservices
)

// NewPresetConfig creates a CaddyConfig based on a preset scenario
func NewPresetConfig(preset Preset, port int) CaddyConfig {
	switch preset {
	case PresetSimple:
		return SimpleConfig(port)
		
	case PresetDevelopment:
		return DevelopmentConfig(port)
		
	case PresetFull:
		return FullConfig(port)
		
	case PresetMicroservices:
		return MicroservicesConfig(port)
		
	default:
		return DefaultConfig()
	}
}

// SimpleConfig returns configuration for a single application
func SimpleConfig(port int) CaddyConfig {
	return CaddyConfig{
		Port:   port,
		Target: "localhost:1337", // Standard web server port
		Routes: []ProxyRoute{},   // No additional routes
	}
}

// DevelopmentConfig returns configuration for development with bento playground
func DevelopmentConfig(port int) CaddyConfig {
	return CaddyConfig{
		Port:   port,
		Target: "localhost:1337", // Main web server
		Routes: []ProxyRoute{
			{Path: "/bento-playground/*", Target: "localhost:4195"}, // Bento playground
			{Path: "/xtemplate/*", Target: "localhost:" + config.GetXTemplatePort()}, // XTemplate development server
		},
	}
}

// FullConfig returns configuration with all standard infrastructure services
func FullConfig(port int) CaddyConfig {
	return CaddyConfig{
		Port:   port,
		Target: "localhost:1337", // Main web server
		Routes: []ProxyRoute{
			{Path: "/bento-playground/*", Target: "localhost:4195"}, // Bento playground
			{Path: "/mcp/*", Target: "localhost:8080"},              // MCP server
		},
	}
}

// MicroservicesConfig returns a base configuration for microservices
func MicroservicesConfig(port int) CaddyConfig {
	return CaddyConfig{
		Port:   port,
		Target: "localhost:1337", // Main web server (fallback)
		Routes: []ProxyRoute{
			{Path: "/api/*", Target: "localhost:4000"},     // API service
			{Path: "/auth/*", Target: "localhost:5000"},    // Auth service
			{Path: "/static/*", Target: "localhost:6000"},  // Static files
			{Path: "/ws/*", Target: "localhost:7000"},      // WebSocket service
		},
	}
}

// AddCommonInfrastructure adds standard infrastructure routes to an existing config
func (cfg CaddyConfig) AddCommonInfrastructure() CaddyConfig {
	cfg.Routes = append(cfg.Routes,
		ProxyRoute{Path: "/bento-playground/*", Target: "localhost:4195"},
		ProxyRoute{Path: "/mcp/*", Target: "localhost:8080"},
	)
	return cfg
}

// AddBentoPlayground adds bento playground to an existing config
func (cfg CaddyConfig) AddBentoPlayground() CaddyConfig {
	cfg.Routes = append(cfg.Routes, ProxyRoute{
		Path:   "/bento-playground/*", 
		Target: "localhost:4195",
	})
	return cfg
}

// AddMCPServer adds MCP server to an existing config  
func (cfg CaddyConfig) AddMCPServer() CaddyConfig {
	cfg.Routes = append(cfg.Routes, ProxyRoute{
		Path:   "/mcp/*",
		Target: "localhost:8080", 
	})
	return cfg
}

// AddService adds a custom service route to an existing config
func (cfg CaddyConfig) AddService(path, target string) CaddyConfig {
	cfg.Routes = append(cfg.Routes, ProxyRoute{
		Path:   path,
		Target: target,
	})
	return cfg
}

// WithPort creates a copy of the config with a different port
func (cfg CaddyConfig) WithPort(port int) CaddyConfig {
	cfg.Port = port
	return cfg
}

// WithMainTarget creates a copy of the config with a different main target
func (cfg CaddyConfig) WithMainTarget(target string) CaddyConfig {
	cfg.Target = target
	return cfg
}

// GenerateAndSave generates a Caddyfile and saves it to the specified path
// If path is just "Caddyfile", saves to .data/caddy/ (standard pattern)
// Use explicit paths like "./Caddyfile" only for custom locations
func (cfg CaddyConfig) GenerateAndSave(filePath string) error {
	caddyfile := GenerateCaddyfile(cfg)
	
	fullPath := filePath
	if filePath == "Caddyfile" {
		// Standard pattern: use .data/caddy/ directory (Docker-ready)
		caddyDir := config.GetCaddyPath()
		if err := os.MkdirAll(caddyDir, 0755); err != nil {
			return err
		}
		fullPath = filepath.Join(caddyDir, "Caddyfile")
	}
	
	return os.WriteFile(fullPath, []byte(caddyfile), 0644)
}


// Quick convenience functions for common patterns

// StartDevelopmentServer starts a development server with bento playground
func StartDevelopmentServer(port int) error {
	cfg := DevelopmentConfig(port)
	runner := New()
	
	if err := cfg.GenerateAndSave("Caddyfile"); err != nil {
		return err
	}
	
	return runner.Run("run", "--config", ".data/caddy/Caddyfile")
}

// StartFullServer starts a server with all infrastructure services
func StartFullServer(port int) error {
	cfg := FullConfig(port)
	runner := New()
	
	if err := cfg.GenerateAndSave("Caddyfile"); err != nil {
		return err
	}
	
	return runner.Run("run", "--config", ".data/caddy/Caddyfile")
}