package deploy

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	sharedcfg "github.com/joeblew999/infra/core/pkg/shared/config"
	configinit "github.com/joeblew999/infra/core/tooling/pkg/configinit"
	flyprefs "github.com/joeblew999/infra/core/tooling/pkg/fly"
	profiles "github.com/joeblew999/infra/core/tooling/pkg/profiles"
	releasepkg "github.com/joeblew999/infra/core/tooling/pkg/release"
	types "github.com/joeblew999/infra/core/tooling/pkg/types"
)

// Service orchestrates the deployment workflow.
type Service struct {
	profile     sharedcfg.ToolingProfile
	profileName string
	repoRoot    string
	coreDir     string
}

// New creates a new deployment service.
func New(profile sharedcfg.ToolingProfile, profileName, repoRoot, coreDir string) *Service {
	return &Service{
		profile:     profile,
		profileName: profileName,
		repoRoot:    repoRoot,
		coreDir:     coreDir,
	}
}

// Options contains deployment configuration.
type Options = types.DeployRequest

// Result contains deployment results.
type Result = types.DeployResult

// Deploy executes the full deployment workflow.
func (s *Service) Deploy(ctx context.Context, opts Options) (*Result, error) {
	out := opts.Stdout
	if out == nil {
		out = io.Discard
	}

	// Resolve settings
	appName := strings.TrimSpace(profiles.FirstNonEmpty(opts.AppName, s.profile.FlyApp))
	if appName == "" {
		return nil, fmt.Errorf("missing Fly app name")
	}

	flySettings, _ := flyprefs.LoadSettings()
	orgSlug := strings.TrimSpace(profiles.FirstNonEmpty(opts.OrgSlug, flySettings.OrgSlug, s.profile.FlyOrg))
	region := strings.TrimSpace(profiles.FirstNonEmpty(opts.Region, flySettings.RegionCode, s.profile.FlyRegion))

	repo := strings.TrimSpace(opts.Repo)
	if repo == "" {
		repo = strings.TrimSpace(s.profile.KORepository)
	}
	if repo == "" {
		repo = fmt.Sprintf("registry.fly.io/%s", appName)
	}

	importPath := profiles.FirstNonEmpty(s.profile.ImportPath, "./cmd/core")
	koOutput := filepath.Join(s.coreDir, profiles.FirstNonEmpty(s.profile.KoConfig, ".ko.yaml"))
	flyOutput := filepath.Join(s.repoRoot, profiles.FirstNonEmpty(s.profile.FlyConfig, "fly.toml"))

	// Generate config
	fmt.Fprintln(out, "⚙️  Generating configuration files...")
	_, err := configinit.Run(ctx, configinit.Options{
		Profile:     s.profile,
		ProfileName: s.profileName,
		RepoRoot:    s.repoRoot,
		CoreDir:     s.coreDir,
		AppName:     appName,
		OrgSlug:     orgSlug,
		Region:      region,
		Repository:  repo,
		Force:       true,
		SkipPrompt:  true,
		KoOutput:    koOutput,
		FlyOutput:   flyOutput,
		Stdout:      out,
		Stderr:      opts.Stderr,
		Stdin:       opts.Stdin,
	})
	if err != nil {
		return nil, fmt.Errorf("config init: %w", err)
	}
	fmt.Fprintln(out)

	tokenPath := profiles.FirstNonEmpty(s.profile.TokenPath, flyprefs.DefaultTokenPath())

	// Build and deploy
	fmt.Fprintln(out, "🏗️  Building and deploying...")
	fmt.Fprintln(out)
	result, err := releasepkg.Run(ctx, releasepkg.Options{
		AppName:      appName,
		ConfigPath:   flyOutput,
		KoConfigPath: koOutput,
		ImportPath:   importPath,
		TokenFile:    tokenPath,
		Tags:         []string{"latest"},
		Verbose:      opts.Verbose,
		CoreDir:      s.coreDir,
		OrgSlug:      orgSlug,
		Profile:      s.profileName,
		Repository:   repo,
	})
	if err != nil {
		return nil, err
	}

	return &Result{
		ImageReference: result.ImageReference,
		ReleaseSummary: result.ReleaseSummary,
		ReleaseID:      result.ReleaseID,
		Elapsed:        result.Elapsed,
		AppName:        appName,
		OrgSlug:        orgSlug,
	}, nil
}
