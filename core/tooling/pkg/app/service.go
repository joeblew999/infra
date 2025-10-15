package app

import (
	"context"

	"github.com/joeblew999/infra/core/tooling/pkg/auth"
	"github.com/joeblew999/infra/core/tooling/pkg/orchestrator"
	profiles "github.com/joeblew999/infra/core/tooling/pkg/profiles"
)

// Service exposes tooling command/query operations for CLI, TUI, or GUI adapters.
type Service struct {
	auth         *auth.Service
	orchestrator *orchestrator.Service
}

// Option customises Service construction.
type Option func(*Service)

// WithAuth allows injecting a custom auth service (useful for tests).
func WithAuth(a *auth.Service) Option {
	return func(s *Service) {
		if a != nil {
			s.auth = a
		}
	}
}

// WithOrchestrator allows injecting a custom orchestrator service.
func WithOrchestrator(o *orchestrator.Service) Option {
	return func(s *Service) {
		if o != nil {
			s.orchestrator = o
		}
	}
}

// New constructs a Service with sensible defaults.
func New(opts ...Option) *Service {
	svc := &Service{
		auth:         auth.New(),
		orchestrator: orchestrator.NewService(),
	}
	for _, opt := range opts {
		opt(svc)
	}
	return svc
}

// Orchestrator exposes the underlying orchestrator.Service for advanced workflows.
func (s *Service) Orchestrator() *orchestrator.Service {
	return s.orchestrator
}

// Auth exposes the underlying auth.Service.
func (s *Service) Auth() *auth.Service {
	return s.auth
}

// Deploy executes the deployment synchronously.
func (s *Service) Deploy(ctx context.Context, opts orchestrator.DeployOptions) (*orchestrator.DeployResult, error) {
	return s.orchestrator.Deploy(ctx, opts)
}

// Launch executes the deployment asynchronously and returns channels suitable for SSE streaming.
func (s *Service) Launch(ctx context.Context, opts orchestrator.DeployOptions) (*orchestrator.StreamAdapter, <-chan *orchestrator.DeployResult, <-chan error) {
	return s.orchestrator.Launch(ctx, opts)
}

// Status returns the current profile + provider settings snapshot.
func (s *Service) Status(ctx context.Context, opts profiles.ContextOptions) (orchestrator.Status, error) {
	return orchestrator.StatusSnapshot(ctx, opts)
}
