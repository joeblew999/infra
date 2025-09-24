package runtime

import (
	"fmt"
	"sort"
	"strconv"

	"github.com/joeblew999/infra/pkg/config"
)

// RouteSpec describes a path prefix that Caddy should proxy to a service.
type RouteSpec struct {
	Path   string
	Target string
}

// caddyTemplate captures the aggregated routing information needed to generate a Caddyfile.
type caddyTemplate struct {
	ListenPort int
	RootTarget string
	Routes     []RouteSpec
}

func buildCaddyTemplate(specs []ServiceSpec) (caddyTemplate, error) {
	var tpl caddyTemplate

	listenPort, err := strconv.Atoi(config.GetCaddyPort())
	if err != nil {
		return tpl, fmt.Errorf("invalid caddy port: %w", err)
	}
	tpl.ListenPort = listenPort

	routeSet := make(map[string]RouteSpec)

	for _, spec := range specs {
		if spec.ID == ServiceWeb && spec.Port != "" {
			tpl.RootTarget = config.FormatLocalHostPort(spec.Port)
		}
		for _, route := range spec.Routes {
			if route.Path == "" || route.Target == "" {
				continue
			}
			if _, ok := routeSet[route.Path]; ok {
				continue
			}
			routeSet[route.Path] = route
		}
	}

	if tpl.RootTarget == "" {
		tpl.RootTarget = config.FormatLocalHostPort(config.GetWebServerPort())
	}

	tpl.Routes = make([]RouteSpec, 0, len(routeSet))
	for _, route := range routeSet {
		tpl.Routes = append(tpl.Routes, route)
	}

	sort.Slice(tpl.Routes, func(i, j int) bool {
		return tpl.Routes[i].Path < tpl.Routes[j].Path
	})

	return tpl, nil
}
