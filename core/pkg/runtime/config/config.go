package config

import (
	"encoding/json"
	"fmt"
	"os"

	shared "github.com/joeblew999/infra/core/pkg/shared/config"
	"github.com/joeblew999/infra/core/pkg/shared/controller"
)

// Paths bundles the resolved filesystem locations required by the orchestrator
// runtime. Runtime packages should use these helpers instead of recomputing
// joins against shared config to keep the directory layout consistent.
type Paths struct {
	AppRoot  string
	Dep      string
	Bin      string
	Data     string
	Logs     string
	TestData string
}

// ServiceURLs holds the runtime URLs for core services.
// URLs are derived from the service registry (service.json files).
type ServiceURLs struct {
	PocketBase string // http://127.0.0.1:8090
	NATS       string // nats://127.0.0.1:4222
	NATSHTtp   string // http://127.0.0.1:8222 (monitoring)
	Caddy      string // http://127.0.0.1:2015
}

// Settings represents the high-level runtime configuration derived from shared
// helpers. Additional runtime-specific values (service specs, feature flags,
// etc.) should be added here as the orchestrator matures.
type Settings struct {
	Environment       string
	Paths             Paths
	Services          ServiceURLs
	EnsureBusCluster  bool
	IsTestEnvironment bool
	IsProduction      bool
}

// Load constructs a Settings value using the shared configuration helpers. It
// should be the single entry point for runtime packages needing configuration
// values so changes in shared helpers propagate automatically.
func Load() Settings {
	return Settings{
		Environment: shared.Environment(),
		Paths: Paths{
			AppRoot:  shared.GetAppRoot(),
			Dep:      shared.GetDepPath(),
			Bin:      shared.GetBinPath(),
			Data:     shared.GetDataPath(),
			Logs:     shared.GetLogsPath(),
			TestData: shared.GetTestDataPath(),
		},
		Services:          loadServiceURLs(),
		EnsureBusCluster:  shared.ShouldEnsureBusCluster(),
		IsTestEnvironment: shared.IsTestEnvironment(),
		IsProduction:      shared.IsProduction(),
	}
}

// loadServiceURLs derives service URLs from the service registry.
// Environment variables can override registry-derived URLs.
func loadServiceURLs() ServiceURLs {
	// Default URLs as fallback
	urls := ServiceURLs{
		PocketBase: "http://127.0.0.1:8090",
		NATS:       "nats://127.0.0.1:4222",
		NATSHTtp:   "http://127.0.0.1:8222",
		Caddy:      "http://127.0.0.1:2015",
	}

	// Try to derive from service specs
	// Note: We load specs directly here rather than using a registry
	// because the registry isn't initialized at config load time
	urls = deriveURLsFromSpecs(urls)

	// Environment variable overrides have highest priority
	if pbURL := os.Getenv("POCKETBASE_URL"); pbURL != "" {
		urls.PocketBase = pbURL
	}
	if natsURL := os.Getenv("NATS_URL"); natsURL != "" {
		urls.NATS = natsURL
	}
	if natsHTTPURL := os.Getenv("NATS_HTTP_URL"); natsHTTPURL != "" {
		urls.NATSHTtp = natsHTTPURL
	}
	if caddyURL := os.Getenv("CADDY_URL"); caddyURL != "" {
		urls.Caddy = caddyURL
	}

	return urls
}

// deriveURLsFromSpecs attempts to load service specs and derive URLs.
// Falls back to provided defaults if specs can't be loaded.
func deriveURLsFromSpecs(defaults ServiceURLs) ServiceURLs {
	urls := defaults

	// NATS service
	if natsPort, natsHTTPPort, ok := getNATSPorts(); ok {
		urls.NATS = fmt.Sprintf("nats://127.0.0.1:%d", natsPort)
		urls.NATSHTtp = fmt.Sprintf("http://127.0.0.1:%d", natsHTTPPort)
	}

	// PocketBase service
	if pbPort, ok := getPocketBasePorts(); ok {
		urls.PocketBase = fmt.Sprintf("http://127.0.0.1:%d", pbPort)
	}

	// Caddy service
	if caddyPort, ok := getCaddyPorts(); ok {
		urls.Caddy = fmt.Sprintf("http://127.0.0.1:%d", caddyPort)
	}

	return urls
}

// GetServiceURL returns the URL for a service by ID.
// This is a helper for services not yet in ServiceURLs struct.
func GetServiceURL(serviceID string, portName string) (string, error) {
	registry := controller.NewRegistry()
	// Note: This creates empty registry. In future, should use global registry.
	// For now, services should use Settings.Services directly.

	spec, ok := registry.Get(serviceID)
	if !ok {
		return "", fmt.Errorf("service %s not found in registry", serviceID)
	}

	for _, port := range spec.Ports {
		if port.Name == portName || portName == "" {
			return fmt.Sprintf("%s://127.0.0.1:%d", port.Protocol, port.Port), nil
		}
	}

	return "", fmt.Errorf("port %s not found for service %s", portName, serviceID)
}

// getNATSPorts loads NATS service spec and returns client and HTTP monitoring ports.
func getNATSPorts() (clientPort, httpPort int, ok bool) {
	// Import dynamically to avoid circular dependencies
	// We'll use a simpler approach - just parse the JSON directly
	data, err := os.ReadFile("services/nats/service.json")
	if err != nil {
		return 0, 0, false
	}
	
	var spec struct {
		Ports struct {
			Client struct {
				Port int `json:"port"`
			} `json:"client"`
			HTTP struct {
				Port int `json:"port"`
			} `json:"http"`
		} `json:"ports"`
	}
	
	if err := json.Unmarshal(data, &spec); err != nil {
		return 0, 0, false
	}
	
	return spec.Ports.Client.Port, spec.Ports.HTTP.Port, true
}

// getPocketBasePorts loads PocketBase service spec and returns the HTTP port.
func getPocketBasePorts() (port int, ok bool) {
	data, err := os.ReadFile("services/pocketbase/service.json")
	if err != nil {
		return 0, false
	}
	
	var spec struct {
		Ports struct {
			Primary struct {
				Port int `json:"port"`
			} `json:"primary"`
		} `json:"ports"`
	}
	
	if err := json.Unmarshal(data, &spec); err != nil {
		return 0, false
	}
	
	return spec.Ports.Primary.Port, true
}

// getCaddyPorts loads Caddy service spec and returns the HTTP port.
func getCaddyPorts() (port int, ok bool) {
	data, err := os.ReadFile("services/caddy/service.json")
	if err != nil {
		return 0, false
	}
	
	var spec struct {
		Ports struct {
			HTTP struct {
				Port int `json:"port"`
			} `json:"http"`
		} `json:"ports"`
	}
	
	if err := json.Unmarshal(data, &spec); err != nil {
		return 0, false
	}
	
	return spec.Ports.HTTP.Port, true
}
