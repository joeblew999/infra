package runtime

import (
	"sync"
	"testing"

	"github.com/joeblew999/infra/pkg/caddy"
)

func TestNotifyCaddyRoutesChangedInvokesReload(t *testing.T) {
	original := caddyReload
	defer func() { caddyReload = original }()

	var mu sync.Mutex
	var captured []caddy.ProxyRoute
	caddyReload = func(cfg *caddy.CaddyConfig) error {
		mu.Lock()
		defer mu.Unlock()
		captured = append([]caddy.ProxyRoute(nil), cfg.Routes...)
		return nil
	}

	activeServiceSpecs = []ServiceSpec{
		{
			ID:          ServiceCaddy,
			DisplayName: "Caddy",
			Required:    true,
		},
		{
			ID:          ServiceWeb,
			DisplayName: "Web",
			Port:        "1337",
		},
		{
			ID:          ServiceBento,
			DisplayName: "Bento",
			Routes:      []RouteSpec{{Path: "/bento/*", Target: "localhost:4195"}},
		},
	}

	NotifyCaddyRoutesChanged()

	mu.Lock()
	defer mu.Unlock()
	if len(captured) != 1 {
		t.Fatalf("expected 1 route, got %d", len(captured))
	}
	if captured[0].Path != "/bento/*" {
		t.Fatalf("unexpected route path %s", captured[0].Path)
	}
}
