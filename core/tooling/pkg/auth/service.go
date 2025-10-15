package auth

import (
	"context"
	"io"
	"strings"
	"time"
	cf "github.com/cloudflare/cloudflare-go"
	flyapi "github.com/superfly/fly-go"

	sharedcfg "github.com/joeblew999/infra/core/pkg/shared/config"
	"github.com/joeblew999/infra/core/tooling/pkg/cloudflare"
	"github.com/joeblew999/infra/core/tooling/pkg/fly"
)

// Service handles authentication for Fly and Cloudflare.
type Service struct{}

// New creates a new authentication service.
func New() *Service {
	return &Service{}
}

// Options contains IO streams and behavior flags for auth operations.
type Options struct {
	Stdin     io.Reader
	Stdout    io.Writer
	Stderr    io.Writer
	NoBrowser bool
	Prompter  Prompter
}

// EnsureFly ensures a valid Fly token exists, authenticating if needed.
func (s *Service) EnsureFly(ctx context.Context, profile sharedcfg.ToolingProfile, opts Options) error {
	in := opts.Stdin
	if in == nil {
		in = strings.NewReader("")
	}

	out := opts.Stdout
	if out == nil {
		out = io.Discard
	}

	prompter := opts.Prompter
	if prompter == nil {
		prompter = NewIOPrompter(in, out, opts.NoBrowser)
	}

	return fly.EnsureFlyToken(ctx, profile, in, out, opts.NoBrowser, prompter)
}

// EnsureCloudflare ensures a valid Cloudflare token exists, authenticating if needed.
func (s *Service) EnsureCloudflare(ctx context.Context, profile sharedcfg.ToolingProfile, opts Options) error {
	in := opts.Stdin
	if in == nil {
		in = strings.NewReader("")
	}

	out := opts.Stdout
	if out == nil {
		out = io.Discard
	}

	prompter := opts.Prompter
	if prompter == nil {
		prompter = NewIOPrompter(in, out, opts.NoBrowser)
	}

	return cloudflare.EnsureCloudflareToken(ctx, profile, in, out, opts.NoBrowser, prompter)
}

// Wrapper functions that delegate to provider packages

// RunCloudflareAuth performs Cloudflare authentication and delegates to cloudflare package.
func RunCloudflareAuth(ctx context.Context, profile sharedcfg.ToolingProfile, tokenInput, tokenPath string, noBrowser bool, in io.Reader, out io.Writer, prompter Prompter) error {
	return cloudflare.RunCloudflareAuth(ctx, profile, tokenInput, tokenPath, noBrowser, in, out, prompter)
}

// VerifyCloudflareToken verifies a Cloudflare token and delegates to cloudflare package.
func VerifyCloudflareToken(ctx context.Context, token string) (cf.APITokenVerifyBody, *cf.API, error) {
	return cloudflare.VerifyCloudflareToken(ctx, token)
}

// BootstrapOptions contains options for Cloudflare bootstrap authentication.
type BootstrapOptions = cloudflare.BootstrapOptions

// RunCloudflareBootstrap performs Cloudflare bootstrap authentication and delegates to cloudflare package.
func RunCloudflareBootstrap(ctx context.Context, profile sharedcfg.ToolingProfile, opts BootstrapOptions, in io.Reader, out io.Writer, prompter Prompter) error {
	return cloudflare.RunCloudflareBootstrap(ctx, profile, opts, in, out, prompter)
}

// RunFlyAuth performs Fly authentication and delegates to fly package.
func RunFlyAuth(ctx context.Context, profile sharedcfg.ToolingProfile, tokenInput, tokenPath string, noBrowser bool, timeout time.Duration, in io.Reader, out io.Writer, prompter Prompter) error {
	return fly.RunFlyAuth(ctx, profile, tokenInput, tokenPath, noBrowser, timeout, in, out, prompter)
}

// VerifyFlyToken verifies a Fly token and delegates to fly package.
func VerifyFlyToken(ctx context.Context, profile sharedcfg.ToolingProfile, token string) (string, *flyapi.Client, error) {
	return fly.VerifyFlyToken(ctx, profile, token)
}
