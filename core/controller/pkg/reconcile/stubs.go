package reconcile

import (
	"context"
	"log"

	controllerspec "github.com/joeblew999/infra/core/controller/pkg/spec"
)

// NullMachines is a no-op MachinesProvider used until real integration is wired.
type NullMachines struct{}

// EnsureMachines simply echoes desired counts for observability.
func (NullMachines) EnsureMachines(ctx context.Context, svc controllerspec.Service) (ServiceRuntimeState, error) {
	runtime := ServiceRuntimeState{Regions: make(map[string]int, len(svc.Scale.Regions))}
	for _, region := range svc.Scale.Regions {
		runtime.Regions[region.Name] = region.Desired
	}
	select {
	case <-ctx.Done():
		return runtime, ctx.Err()
	default:
	}
	log.Printf("    [machines] service=%s strategy=%s desired=%v", svc.ID, svc.Scale.Strategy, runtime.Regions)
	return runtime, nil
}

// NullRouting is a no-op RoutingProvider used until Cloudflare integration is ready.
type NullRouting struct{}

// EnsureRouting logs routing intent without performing changes.
func (NullRouting) EnsureRouting(ctx context.Context, svc controllerspec.Service, runtime ServiceRuntimeState) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	log.Printf("    [routing] service=%s provider=%s zone=%s regions=%v", svc.ID, svc.Routing.Provider, svc.Routing.Zone, runtime.Regions)
	return nil
}
