package config

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	// Load config
	cfg := Load()

	// Check basic fields populated
	if cfg.Environment == "" {
		t.Error("Expected Environment to be set")
	}

	// Check paths populated
	if cfg.Paths.AppRoot == "" {
		t.Error("Expected AppRoot to be set")
	}
	if cfg.Paths.Dep == "" {
		t.Error("Expected Dep to be set")
	}
	if cfg.Paths.Data == "" {
		t.Error("Expected Data to be set")
	}

	// Check service URLs populated
	if cfg.Services.PocketBase == "" {
		t.Error("Expected PocketBase URL to be set")
	}
	if cfg.Services.NATS == "" {
		t.Error("Expected NATS URL to be set")
	}
	if cfg.Services.NATSHTtp == "" {
		t.Error("Expected NATS HTTP URL to be set")
	}
	if cfg.Services.Caddy == "" {
		t.Error("Expected Caddy URL to be set")
	}
}

func TestLoadServiceURLs_Defaults(t *testing.T) {
	// Clear any env vars
	oldPB := os.Getenv("POCKETBASE_URL")
	oldNATS := os.Getenv("NATS_URL")
	os.Unsetenv("POCKETBASE_URL")
	os.Unsetenv("NATS_URL")
	defer func() {
		if oldPB != "" {
			os.Setenv("POCKETBASE_URL", oldPB)
		}
		if oldNATS != "" {
			os.Setenv("NATS_URL", oldNATS)
		}
	}()

	urls := loadServiceURLs()

	// Should have default or registry-derived URLs
	if urls.PocketBase == "" {
		t.Error("Expected PocketBase URL")
	}
	if urls.NATS == "" {
		t.Error("Expected NATS URL")
	}
	if urls.NATSHTtp == "" {
		t.Error("Expected NATS HTTP URL")
	}
	if urls.Caddy == "" {
		t.Error("Expected Caddy URL")
	}

	// Should match expected format
	if urls.PocketBase != "http://127.0.0.1:8090" {
		t.Logf("PocketBase URL: %s (expected http://127.0.0.1:8090)", urls.PocketBase)
	}
	if urls.NATS != "nats://127.0.0.1:4222" {
		t.Logf("NATS URL: %s (expected nats://127.0.0.1:4222)", urls.NATS)
	}
}

func TestLoadServiceURLs_EnvOverrides(t *testing.T) {
	// Set environment variables
	customPB := "http://custom-pb:9999"
	customNATS := "nats://custom-nats:5555"
	customNATSHTTP := "http://custom-nats:8888"
	customCaddy := "http://custom-caddy:3333"

	os.Setenv("POCKETBASE_URL", customPB)
	os.Setenv("NATS_URL", customNATS)
	os.Setenv("NATS_HTTP_URL", customNATSHTTP)
	os.Setenv("CADDY_URL", customCaddy)
	defer func() {
		os.Unsetenv("POCKETBASE_URL")
		os.Unsetenv("NATS_URL")
		os.Unsetenv("NATS_HTTP_URL")
		os.Unsetenv("CADDY_URL")
	}()

	urls := loadServiceURLs()

	// Should use environment variables
	if urls.PocketBase != customPB {
		t.Errorf("Expected PocketBase URL %q, got %q", customPB, urls.PocketBase)
	}
	if urls.NATS != customNATS {
		t.Errorf("Expected NATS URL %q, got %q", customNATS, urls.NATS)
	}
	if urls.NATSHTtp != customNATSHTTP {
		t.Errorf("Expected NATS HTTP URL %q, got %q", customNATSHTTP, urls.NATSHTtp)
	}
	if urls.Caddy != customCaddy {
		t.Errorf("Expected Caddy URL %q, got %q", customCaddy, urls.Caddy)
	}
}

func TestGetNATSPorts(t *testing.T) {
	clientPort, httpPort, ok := getNATSPorts()
	
	if !ok {
		t.Skip("NATS service.json not found or couldn't be read")
	}

	// Should return valid ports
	if clientPort == 0 {
		t.Error("Expected non-zero client port")
	}
	if httpPort == 0 {
		t.Error("Expected non-zero HTTP port")
	}

	// Should return expected default ports
	if clientPort != 4222 {
		t.Logf("Client port: %d (expected 4222)", clientPort)
	}
	if httpPort != 8222 {
		t.Logf("HTTP port: %d (expected 8222)", httpPort)
	}
}

func TestGetPocketBasePorts(t *testing.T) {
	port, ok := getPocketBasePorts()
	
	if !ok {
		t.Skip("PocketBase service.json not found or couldn't be read")
	}

	// Should return valid port
	if port == 0 {
		t.Error("Expected non-zero port")
	}

	// Should return expected default port
	if port != 8090 {
		t.Logf("Port: %d (expected 8090)", port)
	}
}

func TestGetCaddyPorts(t *testing.T) {
	port, ok := getCaddyPorts()
	
	if !ok {
		t.Skip("Caddy service.json not found or couldn't be read")
	}

	// Should return valid port
	if port == 0 {
		t.Error("Expected non-zero port")
	}

	// Should return expected default port
	if port != 2015 {
		t.Logf("Port: %d (expected 2015)", port)
	}
}

func TestDeriveURLsFromSpecs(t *testing.T) {
	defaults := ServiceURLs{
		PocketBase: "http://default:1111",
		NATS:       "nats://default:2222",
		NATSHTtp:   "http://default:3333",
		Caddy:      "http://default:4444",
	}

	urls := deriveURLsFromSpecs(defaults)

	// Should either use registry values or keep defaults
	if urls.PocketBase == "" {
		t.Error("Expected PocketBase URL")
	}
	if urls.NATS == "" {
		t.Error("Expected NATS URL")
	}
	if urls.NATSHTtp == "" {
		t.Error("Expected NATS HTTP URL")
	}
	if urls.Caddy == "" {
		t.Error("Expected Caddy URL")
	}

	t.Logf("Derived URLs: PB=%s, NATS=%s, NATS_HTTP=%s, Caddy=%s",
		urls.PocketBase, urls.NATS, urls.NATSHTtp, urls.Caddy)
}
