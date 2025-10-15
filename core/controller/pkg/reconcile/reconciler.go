package reconcile

import (
	"context"
	"log"
	"time"

	"github.com/joeblew999/infra/core/controller/pkg/apiserver"
	controllerspec "github.com/joeblew999/infra/core/controller/pkg/spec"
)

// MachinesProvider reconciles machine-level capacity (Fly.io).
type MachinesProvider interface {
	EnsureMachines(ctx context.Context, svc controllerspec.Service) (ServiceRuntimeState, error)
}

// RoutingProvider reconciles routing/DNS (Cloudflare).
type RoutingProvider interface {
	EnsureRouting(ctx context.Context, svc controllerspec.Service, runtime ServiceRuntimeState) error
}

// Options configure the reconciler behaviour.
type Options struct {
	Tick     time.Duration
	Machines MachinesProvider
	Routing  RoutingProvider
}

// Reconciler drives desired state towards infrastructure reality.
type Reconciler struct {
	server   *apiserver.Server
	tick     time.Duration
	machines MachinesProvider
	routing  RoutingProvider
}

// New constructs a reconciler with the provided options.
func New(server *apiserver.Server, opts Options) *Reconciler {
	tick := opts.Tick
	if tick <= 0 {
		tick = 30 * time.Second
	}
	machines := opts.Machines
	if machines == nil {
		machines = NullMachines{}
	}
	routing := opts.Routing
	if routing == nil {
		routing = NullRouting{}
	}
	return &Reconciler{
		server:   server,
		tick:     tick,
		machines: machines,
		routing:  routing,
	}
}

// Run starts the reconciliation loop until the context is cancelled.
func (r *Reconciler) Run(ctx context.Context) {
	ticker := time.NewTicker(r.tick)
	defer ticker.Stop()

	updates, cancel := r.server.Subscribe()
	defer cancel()
	r.reconcileOnce(ctx, "startup")

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			r.reconcileOnce(ctx, "periodic")
		case <-updates:
			r.reconcileOnce(ctx, "update")
		}
	}
}

func (r *Reconciler) reconcileOnce(ctx context.Context, reason string) {
	desired := r.server.State()
	log.Printf("reconcile (%s): services=%d", reason, len(desired.Services))

	for _, svc := range desired.Services {
		if err := r.reconcileService(ctx, svc); err != nil {
			log.Printf("  service %s error: %v", svc.ID, err)
		}
	}
}

func (r *Reconciler) reconcileService(ctx context.Context, svc controllerspec.Service) error {
	runtime, err := r.machines.EnsureMachines(ctx, svc)
	if err != nil {
		return err
	}
	if err := r.routing.EnsureRouting(ctx, svc, runtime); err != nil {
		return err
	}
	log.Printf("  service %s reconciled regions=%v", svc.ID, runtime.Regions)
	return nil
}

// ServiceRuntimeState reflects current infrastructure for a service.
type ServiceRuntimeState struct {
	Regions map[string]int
}
