package ui

import (
	"fmt"
	"sort"
	"strings"
	"time"

	runtimecfg "github.com/joeblew999/infra/core/pkg/runtime/config"
	runtimecontroller "github.com/joeblew999/infra/core/pkg/runtime/controller"
	sharedbuild "github.com/joeblew999/infra/core/pkg/shared/build"
)

var (
	fallbackServices = []ServiceCard{
		{
			ID:            "pocketbase",
			Status:        "running",
			Command:       "core pocketbase run",
			Ports:         []string{"primary → 8090/http"},
			Health:        "healthy",
			LastEvent:     "pocketbase target ready",
			Description:   "Embedded PocketBase instance for identity demos",
			Scalable:      false,
			ScaleStrategy: "infra",
		},
		{
			ID:            "nats",
			Status:        "running",
			Command:       "core nats run",
			Ports:         []string{"client → 4222/nats", "cluster → 6222/nats", "http → 8222/http", "leaf → 7422/nats"},
			Health:        "healthy",
			LastEvent:     "auto-scaling ready (8 nodes)",
			Description:   "NATS with Pillow auto-scaling: 1 local, 8 production (3 hub + 5 leaf)",
			Scalable:      false,
			ScaleStrategy: "infra",
		},
		{
			ID:            "core.caddy",
			Status:        "ready",
			Command:       "core services caddy --managed",
			Ports:         []string{"https → 443/tcp", "http → 80/tcp"},
			Health:        "initialising",
			LastEvent:     "requesting certificates",
			Description:   "Managed Caddy proxy bundling custom modules",
			Scalable:      false,
			ScaleStrategy: "infra",
		},
	}

	baseMetrics = []MetricCard{
		{Label: "JetStream Streams", Value: "4", Hint: "core.demo, core.runtime, core.audit, core.metrics"},
		{Label: "Process Restarts", Value: "0", Hint: "deterministic boot pipeline"},
		{Label: "Active Services", Value: "3/3", Hint: "all profiles enabled"},
	}

	baseEvents = []EventLog{
		{Timestamp: "10:15:02", Level: "info", Message: "nats: pillow auto-scaling initialized (1 local node)"},
		{Timestamp: "10:15:00", Level: "info", Message: "nats: hub-spoke topology ready (3 hub + 5 leaf = 8 total)"},
		{Timestamp: "10:14:58", Level: "info", Message: "nats: jetstream quorum established in iad hub region"},
		{Timestamp: "10:14:56", Level: "info", Message: "nats: leaf regions [lhr,nrt,syd,fra,sjc] connected"},
		{Timestamp: "10:14:55", Level: "info", Message: "demo.alpha: queued initial sync"},
		{Timestamp: "10:14:47", Level: "warn", Message: "core.caddy: awaiting certificate provisioning"},
	}

	baseTips = []string{
		"Press Ctrl+C to stop the web UI server",
		"Run 'core nats spec' to inspect backend and topology",
		"Try 'core nats command --env' to view resolved launch parameters",
		"NATS Pillow integration provides simplified Fly.io deployment",
	}
)

// LoadTestSnapshot assembles a snapshot seeded with representative data so the
// UI shells can iterate before the live process runner is wired in.
func LoadTestSnapshot() Snapshot {
	cfg := runtimecfg.Load()
	build := sharedbuild.Get()
	snapshot := Snapshot{
		Environment: cfg.Environment,
		DataDir:     cfg.Paths.Data,
		GeneratedAt: time.Now().Round(time.Second),
		Build: BuildInfo{
			Version:   build.Version,
			Revision:  build.Revision,
			BuildTime: build.BuildTime,
			Dirty:     build.Modified,
		},
		Metrics:     cloneMetrics(baseMetrics),
		Events:      cloneEvents(baseEvents),
		Tips:        append([]string(nil), baseTips...),
		TextIslands: cloneTextIslands(loadTextIslands()),
	}

	services := mergeServicesWithRegistry(fallbackServices)
	snapshot.Services = services
	snapshot.Navigation = buildNavigation(services)
	snapshot.ServiceDetails = buildServiceDetails(services)
	snapshot.Processes = buildProcessDetailsFromServices(services, snapshot.GeneratedAt)
	return snapshot
}

func mergeServicesWithRegistry(fallback []ServiceCard) []ServiceCard {
	registered := loadRegisteredServices()
	if len(registered) == 0 {
		return cloneServices(fallback)
	}

	base := make(map[string]ServiceCard, len(fallback))
	for _, svc := range fallback {
		base[svc.ID] = svc
	}

	result := make([]ServiceCard, 0, len(registered)+len(fallback))
	seen := make(map[string]struct{})
	for idx, svc := range registered {
		card := renderServiceCard(idx, svc)
		if template, ok := base[card.ID]; ok {
			template.Command = card.Command
			template.Ports = card.Ports
			template.Status = card.Status
			template.Health = card.Health
			template.LastEvent = card.LastEvent
			template.Scalable = card.Scalable
			template.ScaleStrategy = card.ScaleStrategy
			result = append(result, template)
		} else {
			result = append(result, card)
		}
		seen[card.ID] = struct{}{}
	}

	for _, svc := range fallback {
		if _, ok := seen[svc.ID]; ok {
			continue
		}
		result = append(result, svc)
	}

	return result
}

func buildNavigation(services []ServiceCard) []NavigationItem {
	items := []NavigationItem{{
		Title:       "Overview",
		Route:       "overview",
		Description: "Stack summary and metrics",
	}}

	sorted := make([]ServiceCard, len(services))
	copy(sorted, services)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].ID < sorted[j].ID
	})

	for _, svc := range sorted {
		items = append(items, NavigationItem{
			Title:       prettyTitle(svc.ID),
			Route:       fmt.Sprintf("service/%s", svc.ID),
			Description: fmt.Sprintf("Detail view for %s", svc.ID),
		})
	}
	return items
}

func buildServiceDetails(services []ServiceCard) map[string]ServiceDetail {
	details := make(map[string]ServiceDetail, len(services))
	for _, svc := range services {
		notes := []string{}
		if svc.Description != "" {
			notes = append(notes, svc.Description)
		}
		notes = append(notes, fmt.Sprintf("Command: %s", svc.Command))
		if len(svc.Ports) > 0 {
			notes = append(notes, "Ports: "+strings.Join(svc.Ports, ", "))
		}
		notes = append(notes, "Status: "+svc.Status+" | Health: "+svc.Health)
		notes = append(notes, fmt.Sprintf("Scaling: scalable=%t (strategy=%s)", svc.Scalable, svc.ScaleStrategy))

		key := fmt.Sprintf("service/%s", svc.ID)
		details[key] = ServiceDetail{
			Card:      svc,
			Notes:     notes,
			Checklist: []string{"Validate process runner wiring", "Expose metrics feed", "Attach event stream"},
		}
	}
	return details
}

func buildProcessDetailsFromServices(services []ServiceCard, ts time.Time) map[string]ProcessDetail {
	processes := make(map[string]ProcessDetail, len(services))
	for _, svc := range services {
		processes[svc.ID] = ProcessDetail{
			Runtime: ProcessRuntime{
				ID:        svc.ID,
				Status:    svc.Status,
				Health:    svc.Health,
				Command:   svc.Command,
				Ports:     append([]string(nil), svc.Ports...),
				Replicas:  1,
				UpdatedAt: ts,
			},
			Scalable:      svc.Scalable,
			ScaleStrategy: svc.ScaleStrategy,
		}
	}
	return processes
}

func loadRegisteredServices() []runtimecontroller.ServiceSpec {
	registry, err := runtimecontroller.LoadBuiltIn()
	if err != nil {
		return nil
	}
	return registry.List()
}

func renderServiceCard(index int, spec runtimecontroller.ServiceSpec) ServiceCard {
	ports := make([]string, 0, len(spec.Ports))
	for _, port := range spec.Ports {
		label := port.Name
		if label == "" {
			label = fmt.Sprintf("port%d", index)
		}
		ports = append(ports, fmt.Sprintf("%s → %d/%s", label, port.Port, port.Protocol))
	}

	command := spec.Process.Command
	if len(spec.Process.Args) > 0 {
		command = command + " " + strings.Join(spec.Process.Args, " ")
	}
	description := spec.Summary
	if description == "" {
		description = fmt.Sprintf("Service %s managed by the deterministic controller", spec.ID)
	}
	scalable := spec.Metadata["scale.local"] == "true"
	strategy := spec.Metadata["scale.strategy"]
	if strategy == "" {
		if scalable {
			strategy = "local"
		} else {
			strategy = "infra"
		}
	}

	return ServiceCard{
		ID:            spec.ID,
		Status:        "running",
		Command:       command,
		Ports:         ports,
		Health:        "healthy",
		LastEvent:     "registered",
		Description:   description,
		Scalable:      scalable,
		ScaleStrategy: strategy,
	}
}

func loadTextIslands() []TextIsland {
	return []TextIsland{
		{Key: "welcome", Locale: "en", Title: "Welcome", Body: "Explore the deterministic core runtime snapshot."},
		{Key: "welcome", Locale: "de", Title: "Willkommen", Body: "Erkunde den deterministischen Core-Laufzeit-Snapshot."},
	}
}

func prettyTitle(id string) string {
	if idx := strings.LastIndex(id, "/"); idx >= 0 {
		id = id[idx+1:]
	}
	parts := strings.Split(id, ".")
	for i, part := range parts {
		if len(part) == 0 {
			continue
		}
		parts[i] = strings.ToUpper(part[:1]) + part[1:]
	}
	return strings.Join(parts, " ")
}
