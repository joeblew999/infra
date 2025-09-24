package runtime

import (
	"context"
	"strings"
	"time"

	"github.com/joeblew999/infra/pkg/config"
	runtimeevents "github.com/joeblew999/infra/pkg/runtime/events"
)

type ServiceID string

const (
	ServiceWeb        ServiceID = "web"
	ServiceNATS       ServiceID = "nats"
	ServicePocketBase ServiceID = "pocketbase"
	ServiceCaddy      ServiceID = "caddy"
	ServiceBento      ServiceID = "bento"
	ServiceDeckAPI    ServiceID = "deck-api"
	ServiceDeckWatch  ServiceID = "deck-watcher"
	ServiceXTemplate  ServiceID = "xtemplate"
	ServiceMox        ServiceID = "mox"
	ServiceHugo       ServiceID = "hugo"
	ServiceNatsS3     ServiceID = "nats-s3"
)

type ServiceSpec struct {
	ID               ServiceID
	DisplayName      string
	Description      string
	Icon             string
	Required         bool
	Port             string
	Start            func(ctx context.Context, opts Options, record func(error)) (func(), error)
	Ensure           func(ctx context.Context, opts Options) error
	AdditionalPorts  []string
	GoremanProcesses []string
	Enabled          func(opts Options) bool
	Routes           []RouteSpec
}

func (s ServiceSpec) String() string {
	return s.DisplayName
}

// ServicePort describes a service's externally exposed port for shutdown/logging.
type ServicePort struct {
	Service string
	Port    string
}

var baseServiceSpecs = []ServiceSpec{
	{
		ID:          ServiceWeb,
		DisplayName: "Web Server",
		Description: "HTTP server that hosts the control panel.",
		Icon:        "🌐",
		Required:    true,
		Port:        config.GetWebServerPort(),
		Start: func(ctx context.Context, opts Options, record func(error)) (func(), error) {
			return startWebServer(ctx, opts, record)
		},
		Ensure: func(ctx context.Context, opts Options) error {
			return ensureWebDirectories()
		},
		Enabled: func(opts Options) bool { return true },
	},
}

func init() {
	AllServiceSpecs()
}

func buildServiceSpecs(opts Options) []ServiceSpec {
	include := make(map[ServiceID]struct{})
	if len(opts.OnlyServices) > 0 {
		for _, id := range opts.OnlyServices {
			include[id] = struct{}{}
		}
	}

	skip := make(map[ServiceID]struct{})
	for _, id := range opts.SkipServices {
		skip[id] = struct{}{}
	}

	shouldInclude := func(id ServiceID) bool {
		if len(include) > 0 {
			if _, ok := include[id]; !ok {
				return false
			}
		}
		if _, ok := skip[id]; ok {
			return false
		}
		return true
	}

	services := make([]ServiceSpec, 0, len(baseServiceSpecs))
	for _, spec := range baseServiceSpecs {
		if shouldInclude(spec.ID) {
			services = append(services, spec)
			publishServiceSpec(spec)
		}
	}

	appendService := func(spec ServiceSpec) {
		if shouldInclude(spec.ID) {
			services = append(services, spec)
			publishServiceSpec(spec)
		}
	}

	appendService(ServiceSpec{
		ID:          ServiceNATS,
		DisplayName: "Embedded NATS",
		Description: "Embedded messaging backbone (JetStream).",
		Icon:        "📡",
		Required:    true,
		Port:        config.GetNATSPort(),
		Start: func(ctx context.Context, opts Options, record func(error)) (func(), error) {
			return startNATS(ctx, record)
		},
		Ensure: func(ctx context.Context, opts Options) error {
			return ensureNATSDirectories()
		},
		AdditionalPorts:  []string{config.GetNatsS3Port()},
		GoremanProcesses: []string{"nats-s3"},
		Enabled:          func(opts Options) bool { return true },
	})

	appendService(ServiceSpec{
		ID:          ServicePocketBase,
		DisplayName: "PocketBase",
		Description: "Embedded database/UI backend.",
		Icon:        "🗄️",
		Required:    false,
		Port:        config.GetPocketBasePort(),
		Start: func(ctx context.Context, opts Options, record func(error)) (func(), error) {
			return startPocketBase(ctx, opts.Mode, record)
		},
		Ensure: func(ctx context.Context, opts Options) error {
			return ensurePocketBaseDirectories()
		},
		Enabled: func(opts Options) bool { return true },
		Routes: []RouteSpec{
			{Path: "/pocketbase/*", Target: config.FormatLocalHostPort(config.GetPocketBasePort())},
		},
	})

	appendService(ServiceSpec{
		ID:          ServiceCaddy,
		DisplayName: "Caddy Reverse Proxy",
		Description: "HTTPS/TLS reverse proxy fronting services.",
		Icon:        "🛡️",
		Required:    true,
		Port:        config.GetCaddyPort(),
		Start: func(ctx context.Context, opts Options, record func(error)) (func(), error) {
			return startCaddy()
		},
		Ensure: func(ctx context.Context, opts Options) error {
			return ensureCaddyDirectories()
		},
		Enabled:          func(opts Options) bool { return true },
		GoremanProcesses: []string{"caddy"},
	})

	appendService(ServiceSpec{
		ID:          ServiceBento,
		DisplayName: "Bento Stream Processor",
		Description: "Stream processing pipeline (Bento).",
		Icon:        "🍱",
		Required:    false,
		Port:        config.GetBentoPort(),
		Start: func(ctx context.Context, opts Options, record func(error)) (func(), error) {
			return startBento()
		},
		Ensure: func(ctx context.Context, opts Options) error {
			return ensureBentoDirectories()
		},
		Enabled:          func(opts Options) bool { return true },
		GoremanProcesses: []string{"bento"},
		Routes: []RouteSpec{
			{Path: "/bento-playground/*", Target: config.FormatLocalHostPort(config.GetBentoPort())},
		},
	})

	appendService(ServiceSpec{
		ID:          ServiceDeckAPI,
		DisplayName: "Deck API",
		Description: "On-demand presentation builder.",
		Icon:        "🃏",
		Required:    false,
		Port:        config.GetDeckAPIPort(),
		Start: func(ctx context.Context, opts Options, record func(error)) (func(), error) {
			return startDeckAPI()
		},
		Ensure: func(ctx context.Context, opts Options) error {
			return ensureDeckDirectories()
		},
		Enabled:          func(opts Options) bool { return true },
		GoremanProcesses: []string{"deck-api"},
		Routes: []RouteSpec{
			{Path: "/deck-api/*", Target: config.FormatLocalHostPort(config.GetDeckAPIPort())},
		},
	})

	appendService(ServiceSpec{
		ID:          ServiceDeckWatch,
		DisplayName: "Deck Watcher",
		Description: "Deck asset watcher",
		Icon:        "👀",
		Required:    false,
		Port:        "",
		Start: func(ctx context.Context, opts Options, record func(error)) (func(), error) {
			return nil, nil
		},
		// TODO: Re-enable deck watcher once the xtemplate watcher integration lands.
		Enabled: func(opts Options) bool { return false },
	})

	appendService(ServiceSpec{
		ID:          ServiceXTemplate,
		DisplayName: "XTemplate",
		Description: "Template dev server",
		Icon:        "🧩",
		Required:    false,
		Port:        config.GetXTemplatePort(),
		Start: func(ctx context.Context, opts Options, record func(error)) (func(), error) {
			return startXTemplate()
		},
		Ensure: func(ctx context.Context, opts Options) error {
			return ensureXTemplateDirectories()
		},
		Enabled:          func(opts Options) bool { return true },
		GoremanProcesses: []string{"xtemplate"},
		Routes: []RouteSpec{
			{Path: "/xtemplate/*", Target: config.FormatLocalHostPort(config.GetXTemplatePort())},
		},
	})

	appendService(ServiceSpec{
		ID:          ServiceHugo,
		DisplayName: "Hugo Docs",
		Description: "Documentation site",
		Icon:        "📚",
		Required:    false,
		Port:        config.GetHugoPort(),
		Start: func(ctx context.Context, opts Options, record func(error)) (func(), error) {
			return startHugo(record)
		},
		Ensure: func(ctx context.Context, opts Options) error {
			return ensureHugoDirectories()
		},
		Enabled:          func(opts Options) bool { return true },
		GoremanProcesses: []string{"hugo"},
		Routes: []RouteSpec{
			{Path: "/docs/*", Target: config.FormatLocalHostPort(config.GetHugoPort())},
		},
	})

	return services
}

func collectServicePorts(opts Options) []ServicePort {
	return collectServicePortsForSpecs(buildServiceSpecs(opts))
}

func collectServicePortsForSpecs(specs []ServiceSpec) []ServicePort {
	seen := make(map[string]struct{})
	var ports []ServicePort

	addPort := func(port, service string) {
		if port == "" {
			return
		}
		if _, ok := seen[port]; ok {
			return
		}
		seen[port] = struct{}{}
		ports = append(ports, ServicePort{Service: service, Port: port})
	}

	for _, svc := range specs {
		addPort(svc.Port, svc.DisplayName)
		for _, extra := range svc.AdditionalPorts {
			addPort(extra, svc.DisplayName)
		}
	}

	return ports
}

func collectGoremanProcesses(opts Options) []string {
	return collectGoremanProcessesForSpecs(buildServiceSpecs(opts))
}

func collectGoremanProcessesForSpecs(specs []ServiceSpec) []string {
	seen := make(map[string]struct{})
	var procs []string
	for _, svc := range specs {
		for _, name := range svc.GoremanProcesses {
			if name == "" {
				continue
			}
			if _, ok := seen[name]; !ok {
				seen[name] = struct{}{}
				procs = append(procs, name)
			}
		}
	}
	return procs
}

// AllServiceSpecs returns the service specifications without filtering.
func AllServiceSpecs() []ServiceSpec {
	return buildServiceSpecs(Options{})
}

func publishServiceSpec(spec ServiceSpec) {
	runtimeevents.Publish(runtimeevents.ServiceRegistered{
		TS:          time.Now(),
		ServiceID:   string(spec.ID),
		Name:        spec.DisplayName,
		Description: spec.Description,
		Icon:        spec.Icon,
		Required:    spec.Required,
		Port:        spec.Port,
	})
}

// ResolveServiceID converts user-provided identifiers into ServiceID values.
func ResolveServiceID(name string) (ServiceID, bool) {
	normalized := strings.ToLower(name)
	for _, spec := range AllServiceSpecs() {
		if string(spec.ID) == normalized || strings.ToLower(spec.DisplayName) == normalized {
			return spec.ID, true
		}
	}
	return "", false
}
