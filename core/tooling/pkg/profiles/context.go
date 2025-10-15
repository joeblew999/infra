package profiles

import (
	"fmt"
	"strings"

	sharedcfg "github.com/joeblew999/infra/core/pkg/shared/config"
	cloudflareprefs "github.com/joeblew999/infra/core/tooling/pkg/cloudflare"
	flyprefs "github.com/joeblew999/infra/core/tooling/pkg/fly"
)

// ContextOptions defines the inputs used to resolve a workflow context.
type ContextOptions struct {
	ProfileOverride string
	RepoRoot        string
	CoreDir         string
}

// Context summarises the active tooling configuration and cached provider settings.
type Context struct {
	Profile     sharedcfg.ToolingProfile
	ProfileName string
	RepoRoot    string
	CoreDir     string
	Fly         flyprefs.Settings
	Cloudflare  cloudflareprefs.Settings
}

// ResolveContext builds a workflow context using shared helpers and cached settings.
func ResolveContext(opts ContextOptions) (Context, error) {
	profile, profileName := ResolveProfile(strings.TrimSpace(opts.ProfileOverride))

	repoRoot := strings.TrimSpace(opts.RepoRoot)
	if repoRoot == "" {
		var err error
		repoRoot, err = FindRepoRoot("")
		if err != nil {
			return Context{}, fmt.Errorf("profiles: resolve repo root: %w", err)
		}
	}

	coreDir := strings.TrimSpace(opts.CoreDir)
	if coreDir == "" {
		coreDir = ResolveCoreDir(repoRoot)
	}

	flySettings, err := flyprefs.LoadSettings()
	if err != nil {
		return Context{}, fmt.Errorf("profiles: load fly settings: %w", err)
	}

	cloudflareSettings, err := cloudflareprefs.LoadSettings()
	if err != nil {
		return Context{}, fmt.Errorf("profiles: load cloudflare settings: %w", err)
	}

	return Context{
		Profile:     profile,
		ProfileName: profileName,
		RepoRoot:    repoRoot,
		CoreDir:     coreDir,
		Fly:         flySettings,
		Cloudflare:  cloudflareSettings,
	}, nil
}
