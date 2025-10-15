package orchestrator

import (
	"context"
	"io"

	sharedcfg "github.com/joeblew999/infra/core/pkg/shared/config"
	"github.com/joeblew999/infra/core/tooling/pkg/auth"
)

// AuthProvider handles authentication for both Fly and Cloudflare.
type AuthProvider interface {
	EnsureFly(ctx context.Context, profile sharedcfg.ToolingProfile, opts auth.Options) error
	EnsureCloudflare(ctx context.Context, profile sharedcfg.ToolingProfile, opts auth.Options) error
}

// Deployer performs the actual deployment to Fly.
type Deployer interface {
	Deploy(ctx context.Context) (string, error)
}

type deployerFactory func(sharedcfg.ToolingProfile, string, string, string) Deployer
type profileResolver func(string) (sharedcfg.ToolingProfile, string)
