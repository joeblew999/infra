package orchestrator

import (
	"context"

	sharedcfg "github.com/joeblew999/infra/core/pkg/shared/config"
	"github.com/joeblew999/infra/core/tooling/pkg/auth"
	types "github.com/joeblew999/infra/core/tooling/pkg/types"
)

// AuthProvider abstracts Fly and Cloudflare authentication.
type AuthProvider interface {
	EnsureFly(context.Context, sharedcfg.ToolingProfile, auth.Options) error
	EnsureCloudflare(context.Context, sharedcfg.ToolingProfile, auth.Options) error
}

// Deployer executes the final deployment steps.
type Deployer interface {
	Deploy(context.Context, types.DeployRequest) (*types.DeployResult, error)
}

type deployerFactory func(sharedcfg.ToolingProfile, string, string, string) Deployer
type profileResolver func(string) (sharedcfg.ToolingProfile, string)
