package webgui_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/joeblew999/infra/core/pkg/testing/webgui"
)

// TestStackHealth verifies the full stack is running and healthy.
//
// This test requires the stack to be running (go run ./cmd/core stack up).
// Set SKIP_INTEGRATION_TESTS=1 to skip this test.
func TestStackHealth(t *testing.T) {
	if os.Getenv("SKIP_INTEGRATION_TESTS") != "" {
		t.Skip("Skipping integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	tests := []struct {
		name    string
		baseURL string
		service string
	}{
		{
			name:    "PocketBase direct",
			baseURL: "http://127.0.0.1:8090",
			service: "pocketbase",
		},
		{
			name:    "Caddy proxy",
			baseURL: "http://127.0.0.1:2015",
			service: "caddy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := webgui.NewClient(tt.baseURL)

			// Wait for service to be ready
			if err := client.WaitForReady(ctx); err != nil {
				t.Fatalf("Service %s not ready: %v", tt.service, err)
			}

			// Verify health endpoint
			if err := client.CheckHealth(ctx); err != nil {
				t.Errorf("Health check failed for %s: %v", tt.service, err)
			}

			t.Logf("✓ Service %s is healthy at %s", tt.service, tt.baseURL)
		})
	}
}

// TestNATSHealth verifies NATS is running and healthy.
func TestNATSHealth(t *testing.T) {
	if os.Getenv("SKIP_INTEGRATION_TESTS") != "" {
		t.Skip("Skipping integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client := webgui.NewClient("http://127.0.0.1:8222")

	// NATS health endpoint is /healthz
	resp, err := client.Get(ctx, "/healthz")
	if err != nil {
		t.Fatalf("NATS health check failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("NATS health check returned %d, expected 200", resp.StatusCode)
	}

	t.Logf("✓ NATS is healthy at http://127.0.0.1:8222")
}

// TestPocketBaseAdmin verifies PocketBase admin UI is accessible.
func TestPocketBaseAdmin(t *testing.T) {
	if os.Getenv("SKIP_INTEGRATION_TESTS") != "" {
		t.Skip("Skipping integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client := webgui.NewClient("http://127.0.0.1:8090")

	// Admin UI should be accessible
	resp, err := client.Get(ctx, "/_/")
	if err != nil {
		t.Fatalf("Admin UI not accessible: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("Admin UI returned %d, expected 200", resp.StatusCode)
	}

	t.Logf("✓ PocketBase admin UI is accessible at http://127.0.0.1:8090/_/")
}

// TestCaddyProxy verifies Caddy correctly proxies to PocketBase.
func TestCaddyProxy(t *testing.T) {
	if os.Getenv("SKIP_INTEGRATION_TESTS") != "" {
		t.Skip("Skipping integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	caddyClient := webgui.NewClient("http://127.0.0.1:2015")
	pocketbaseClient := webgui.NewClient("http://127.0.0.1:8090")

	// Both should return the same health status
	caddyErr := caddyClient.CheckHealth(ctx)
	pocketbaseErr := pocketbaseClient.CheckHealth(ctx)

	if caddyErr != nil && pocketbaseErr == nil {
		t.Errorf("Caddy proxy failing but PocketBase healthy: %v", caddyErr)
	}

	if caddyErr == nil && pocketbaseErr == nil {
		t.Log("✓ Caddy successfully proxies to PocketBase")
	}
}
