package orchestrator

import (
	"context"
	"fmt"
	"io"
	"time"

	sharedcfg "github.com/joeblew999/infra/core/pkg/shared/config"
	"github.com/joeblew999/infra/core/tooling/pkg/auth"
	cloudflare "github.com/joeblew999/infra/core/tooling/pkg/cloudflare"
	"github.com/joeblew999/infra/core/tooling/pkg/deploy"
	flyprefs "github.com/joeblew999/infra/core/tooling/pkg/fly"
	profiles "github.com/joeblew999/infra/core/tooling/pkg/profiles"
	types "github.com/joeblew999/infra/core/tooling/pkg/types"
)

// Service orchestrates the end-to-end deployment workflow.
type Service struct {
	auth           AuthProvider
	makeDeployer   deployerFactory
	resolveProfile profileResolver
}

// Option customises Service construction.
type Option func(*Service)

// WithAuthProvider overrides the default authentication provider.
func WithAuthProvider(provider AuthProvider) Option {
	return func(s *Service) {
		if provider != nil {
			s.auth = provider
		}
	}
}

// WithDeployerFactory overrides the deployer constructor.
func WithDeployerFactory(factory deployerFactory) Option {
	return func(s *Service) {
		if factory != nil {
			s.makeDeployer = factory
		}
	}
}

// WithProfileResolver overrides the profile resolution helper.
func WithProfileResolver(resolver profileResolver) Option {
	return func(s *Service) {
		if resolver != nil {
			s.resolveProfile = resolver
		}
	}
}

// NewService constructs a deployment orchestrator with optional overrides.
func NewService(opts ...Option) *Service {
	svc := &Service{
		auth: auth.New(),
		makeDeployer: func(profile sharedcfg.ToolingProfile, profileName, repoRoot, coreDir string) Deployer {
			return deploy.New(profile, profileName, repoRoot, coreDir)
		},
		resolveProfile: profiles.ResolveProfile,
	}
	for _, opt := range opts {
		opt(svc)
	}
	return svc
}

// DeployOptions configures the deployment workflow.
type DeployOptions struct {
	ProfileOverride string
	RepoRoot        string
	CoreDir         string
	Timeout         time.Duration
	types.DeployRequest
	Emitter  ProgressEmitter
	Prompter auth.Prompter
}

// DeployResult captures the outcome of a deployment workflow.
type DeployResult struct {
	ProfileName string
	Profile     string
	types.DeployResult
}

// Deploy runs the full deployment workflow.
func (s *Service) Deploy(ctx context.Context, opts DeployOptions) (*DeployResult, error) {
	if s.auth == nil {
		s.auth = auth.New()
	}

	if opts.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, opts.Timeout)
		defer cancel()
	}

	req := opts.DeployRequest
	out := req.Stdout
	if out == nil {
		out = io.Discard
	}
	if req.Stdout == nil {
		req.Stdout = out
	}
	if req.Stderr == nil {
		req.Stderr = io.Discard
	}

	emitter := opts.Emitter
	if emitter == nil {
		emitter = NewTextEmitter(out)
	}

	emit := func(phase ProgressPhase, message string, details map[string]string) {
		if emitter != nil {
			emitter.Emit(ProgressEvent{Phase: phase, Message: message, Details: details, Time: time.Now().UTC()})
		}
	}

	emit(PhaseStarted, "🚀 Starting deployment workflow...", nil)

	ctxInfo, err := profiles.ResolveContext(profiles.ContextOptions{
		ProfileOverride: opts.ProfileOverride,
		RepoRoot:        opts.RepoRoot,
		CoreDir:         opts.CoreDir,
	})
	if err != nil {
		emit(PhaseFailed, "Failed to resolve tooling context.", map[string]string{"error": err.Error()})
		return nil, err
	}

	if s.resolveProfile != nil {
		profile, profileName := s.resolveProfile(opts.ProfileOverride)
		if profile.Name != "" {
			ctxInfo.Profile = profile
			ctxInfo.ProfileName = profileName
		}
	}

	profile := ctxInfo.Profile
	profileName := ctxInfo.ProfileName
	repoRoot := ctxInfo.RepoRoot
	coreDir := ctxInfo.CoreDir
	flySettings := ctxInfo.Fly
	cloudflareSettings := ctxInfo.Cloudflare

	prompter := opts.Prompter
	if prompter == nil {
		prompter = auth.NewIOPrompter(req.Stdin, out, req.NoBrowser)
	}

	authOpts := auth.Options{
		Stdin:     req.Stdin,
		Stdout:    out,
		Stderr:    req.Stderr,
		NoBrowser: req.NoBrowser,
		Prompter:  prompter,
	}

	emit(PhaseFlyAuth, "Authenticating with Fly.io...", nil)
	if err := s.auth.EnsureFly(ctx, profile, authOpts); err != nil {
		emit(PhaseFailed, "Fly authentication failed.", map[string]string{"error": err.Error()})
		return nil, fmt.Errorf("fly authentication failed: %w", err)
	}
	flySettings, _ = flyprefs.LoadSettings()
	emit(PhaseFlyAuthCompleted, "Fly authentication complete.", map[string]string{
		"org":    flySettings.OrgSlug,
		"region": flySettings.RegionCode,
	})

	emit(PhaseCloudflareAuth, "Authenticating with Cloudflare...", nil)
	if err := s.auth.EnsureCloudflare(ctx, profile, authOpts); err != nil {
		emit(PhaseFailed, "Cloudflare authentication failed.", map[string]string{"error": err.Error()})
		return nil, fmt.Errorf("cloudflare authentication failed: %w", err)
	}
	cloudflareSettings, _ = cloudflare.LoadSettings()
	emit(PhaseCloudflareComplete, "Cloudflare authentication complete.", map[string]string{
		"zone":   cloudflareSettings.ZoneName,
		"bucket": cloudflareSettings.R2Bucket,
	})

	factory := s.makeDeployer
	if factory == nil {
		factory = func(profile sharedcfg.ToolingProfile, profileName, repoRoot, coreDir string) Deployer {
			return deploy.New(profile, profileName, repoRoot, coreDir)
		}
	}

	deployer := factory(profile, profileName, repoRoot, coreDir)
	emit(PhaseDeploying, "🏗️  Building and deploying...", map[string]string{
		"profile":   profileName,
		"repo_root": repoRoot,
		"core_dir":  coreDir,
		"app":       req.AppName,
	})

	res, err := deployer.Deploy(ctx, req)
	if err != nil {
		emit(PhaseFailed, "Deployment failed.", map[string]string{"error": err.Error()})
		return nil, err
	}

	emit(PhaseCloudflareDNS, "Configuring Cloudflare DNS...", nil)
	hostname, dnsErr := cloudflare.EnsureAppHostname(ctx, profile, cloudflareSettings, res.AppName)
	if dnsErr != nil {
		emit(PhaseFailed, "Cloudflare DNS configuration failed.", map[string]string{"error": dnsErr.Error()})
		return nil, dnsErr
	}

	details := map[string]string{
		"image":      res.ImageReference,
		"deployment": res.ReleaseSummary,
		"release_id": res.ReleaseID,
		"elapsed":    res.Elapsed.String(),
		"app":        res.AppName,
		"org":        res.OrgSlug,
		"profile":    profileName,
		"fly_region": flySettings.RegionCode,
		"fly_org":    flySettings.OrgSlug,
		"cf_zone":    cloudflareSettings.ZoneName,
		"cf_account": cloudflareSettings.AccountID,
		"cf_bucket":  cloudflareSettings.R2Bucket,
	}
	if hostname != "" {
		details["cf_hostname"] = hostname
	}

	emit(PhaseSucceeded, "✅ Deployment successful!", details)

	return &DeployResult{
		ProfileName:  profileName,
		Profile:      profile.Name,
		DeployResult: *res,
	}, nil
}

// Launch starts the deployment workflow using a stream adapter and returns
// channels for progress, prompts, results, and errors. The caller must read
// from resultCh and errCh until they are closed.
func (s *Service) Launch(ctx context.Context, opts DeployOptions) (*StreamAdapter, <-chan *DeployResult, <-chan error) {
	adapter := NewStreamAdapter()
	resultCh := make(chan *DeployResult, 1)
	errCh := make(chan error, 1)

	go func() {
		defer adapter.Close()
		defer close(resultCh)
		defer close(errCh)

		// copy options so caller's struct remains untouched
		launchOpts := opts
		launchOpts.Emitter = combineEmitters(adapter.Emitter(), opts.Emitter)
		launchOpts.Prompter = combinePrompters(adapter.Prompter(), opts.Prompter)

		res, err := s.Deploy(ctx, launchOpts)
		if err != nil {
			errCh <- err
			return
		}
		resultCh <- res
	}()

	return adapter, resultCh, errCh
}

