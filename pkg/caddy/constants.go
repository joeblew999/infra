package caddy

import (
	"path/filepath"
	"strconv"

	"github.com/joeblew999/infra/pkg/config"
)

// Default ports and targets sourced from configuration package.
func defaultMainTarget() string {
	cfg := config.GetConfig()
	return "localhost:" + cfg.Ports.WebServer
}

func defaultBentoTarget() string {
	return "localhost:" + config.GetBentoPort()
}

func defaultMCPProcTarget() string {
	return "localhost:" + config.GetMCPPort()
}

func defaultDocsTarget() string {
	return "localhost:" + config.GetHugoPort()
}

func defaultXTemplateTarget() string {
	return "localhost:" + config.GetXTemplatePort()
}

func defaultCaddyPort() int {
	return mustAtoi(config.GetCaddyPort())
}

func docRoutes() []ProxyRoute {
	target := defaultDocsTarget()
	return []ProxyRoute{
		{Path: "/docs-hugo/*", Target: target},
		{Path: "/docs/*", Target: target},
	}
}

func developmentSupportRoutes() []ProxyRoute {
	routes := []ProxyRoute{
		{Path: "/bento-playground/*", Target: defaultBentoTarget()},
	}
	routes = append(routes, docRoutes()...)
	routes = append(routes, ProxyRoute{Path: "/xtemplate/*", Target: defaultXTemplateTarget()})
	return routes
}

func fullSupportRoutes() []ProxyRoute {
	return []ProxyRoute{
		{Path: "/bento-playground/*", Target: defaultBentoTarget()},
		{Path: "/mcp/*", Target: defaultMCPProcTarget()},
		{Path: "/docs/*", Target: defaultDocsTarget()},
	}
}

func infrastructureRoutes() []ProxyRoute {
	routes := []ProxyRoute{
		{Path: "/bento-playground/*", Target: defaultBentoTarget()},
		{Path: "/mcp/*", Target: defaultMCPProcTarget()},
	}
	routes = append(routes, docRoutes()...)
	return routes
}

func mustAtoi(value string) int {
	if value == "" {
		return 0
	}
	if v, err := strconv.Atoi(value); err == nil {
		return v
	}
	return 0
}

func defaultCaddyfilePath() string {
	return filepath.Join(config.GetCaddyPath(), "Caddyfile")
}
