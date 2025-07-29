package caddy

import (
	"os"
	"strings"
	"testing"
)

func TestGenerateCaddyfile(t *testing.T) {
	// Save original environment
	originalEnv := os.Getenv("ENVIRONMENT")
	originalFlyAppName := os.Getenv("FLY_APP_NAME")
	
	// Ensure we're in development mode
	os.Unsetenv("ENVIRONMENT")
	os.Unsetenv("FLY_APP_NAME")
	defer func() {
		os.Setenv("ENVIRONMENT", originalEnv)
		os.Setenv("FLY_APP_NAME", originalFlyAppName)
	}()

	// Test development mode
	devCaddyfile := GenerateCaddyfile(443, 1337)
	if !strings.Contains(devCaddyfile, "localhost:443") {
		t.Error("Development Caddyfile should contain localhost:443")
	}
	if !strings.Contains(devCaddyfile, "tls internal") {
		t.Error("Development Caddyfile should contain tls internal")
	}

	// Test production mode
	os.Setenv("ENVIRONMENT", "production")
	os.Unsetenv("FLY_APP_NAME") // Ensure not in production mode via FLY_APP_NAME
	
	prodCaddyfile := GenerateCaddyfile(80, 1337)
	if !strings.Contains(prodCaddyfile, ":80") {
		t.Error("Production Caddyfile should contain :80")
	}
	if strings.Contains(prodCaddyfile, "tls internal") {
		t.Error("Production Caddyfile should NOT contain tls internal")
	}
	if strings.Contains(prodCaddyfile, "localhost:80") {
		t.Error("Production Caddyfile should not specify localhost")
	}
}