package runtime

import (
	"testing"
)

func TestBuildCaddyTemplate(t *testing.T) {
	specs := []ServiceSpec{
		{
			ID:          ServiceWeb,
			DisplayName: "Web Server",
			Port:        "1337",
		},
		{
			ID:          ServiceBento,
			DisplayName: "Bento",
			Port:        "4195",
			Routes: []RouteSpec{
				{Path: "/bento/*", Target: "localhost:4195"},
			},
		},
		{
			ID:          ServiceHugo,
			DisplayName: "Hugo",
			Port:        "1313",
			Routes: []RouteSpec{
				{Path: "/docs/*", Target: "localhost:1313"},
			},
		},
	}

	tpl, err := buildCaddyTemplate(specs)
	if err != nil {
		t.Fatalf("buildCaddyTemplate returned error: %v", err)
	}

	if tpl.ListenPort == 0 {
		t.Fatalf("expected listen port to be set")
	}

	if tpl.RootTarget != "localhost:1337" {
		t.Fatalf("expected root target localhost:1337, got %s", tpl.RootTarget)
	}

	if len(tpl.Routes) != 2 {
		t.Fatalf("expected two routes, got %d", len(tpl.Routes))
	}
}
