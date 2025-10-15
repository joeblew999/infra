package release

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	client "github.com/superfly/fly-go"

	sharedcfg "github.com/joeblew999/infra/core/pkg/shared/config"
	sharedlog "github.com/joeblew999/infra/core/pkg/shared/log"

	flyutil "github.com/joeblew999/infra/core/tooling/pkg/fly"

	"github.com/joeblew999/infra/core/tooling/pkg/ko"
)

// Options controls the end-to-end Fly deployment pipeline.
type Options struct {
	AppName         string
	ConfigPath      string
	KoConfigPath    string
	ImportPath      string
	Token           string
	TokenFile       string
	Tag             string
	Tags            []string
	PlatformVersion string
	Strategy        string
	Verbose         bool
	CoreDir         string
	OrgSlug         string
	Profile         string
	Repository      string
}

// Result captures metadata about a pipeline run.
type Result struct {
	ImageReference string
	ReleaseSummary string
	Skipped        bool
	ReleaseID      string
	Elapsed        time.Duration
}

// Run executes the full pipeline: ensure token, build/push image, and deploy.
func Run(ctx context.Context, opts Options) (Result, error) {
	var result Result
	start := time.Now()

	toolingCfg := sharedcfg.Tooling()
	profile := toolingCfg.Active
	if strings.TrimSpace(opts.Profile) != "" {
		if selected, ok := toolingCfg.Profiles[strings.TrimSpace(opts.Profile)]; ok {
			profile = selected
		}
	}

	if opts.AppName == "" {
		return result, fmt.Errorf("fly pipeline: app name is required")
	}
	if opts.ConfigPath == "" {
		return result, fmt.Errorf("fly pipeline: config path is required")
	}
	if opts.ImportPath == "" {
		return result, fmt.Errorf("fly pipeline: import path is required")
	}
	if opts.Tag == "" {
		opts.Tag = "latest"
	}

	token, err := resolveToken(opts.Token, opts.TokenFile)
	if err != nil {
		return result, err
	}

	client.SetBaseURL(toolingCfg.Active.FlyAPIBase)
	log := flyutil.NewLogger("tooling-pipeline", opts.Verbose)

	cl := client.NewClientFromOptions(client.ClientOptions{
		AccessToken: token,
		Name:        "core-pipeline",
		Version:     "dev",
		Logger:      log,
	})

	var org *client.Organization
	app, err := cl.GetApp(ctx, opts.AppName)
	if isNotFound(err) {
		org, err = resolveOrganization(ctx, cl, opts.OrgSlug)
		if err != nil {
			return result, err
		}
		if err := ensureApp(ctx, cl, org, opts.AppName); err != nil {
			return result, err
		}
		app, err = cl.GetApp(ctx, opts.AppName)
	}
	if err != nil {
		return result, fmt.Errorf("lookup app %q: %w", opts.AppName, err)
	}

	dockerConfigDir, err := flyutil.CreateDockerConfig(token)
	if err != nil {
		return result, fmt.Errorf("prepare docker auth: %w", err)
	}
	defer os.RemoveAll(dockerConfigDir)

	koConfigPath := strings.TrimSpace(opts.KoConfigPath)
	if koConfigPath == "" {
		candidate := strings.TrimSpace(profile.KoConfig)
		if candidate == "" {
			candidate = ".ko.yaml"
		}
		if filepath.IsAbs(candidate) {
			koConfigPath = candidate
		} else {
			koConfigPath = filepath.Join(opts.CoreDir, candidate)
		}
	}
	koOpts := ko.PublishOptions{
		ConfigPath: koConfigPath,
		Args:       []string{opts.ImportPath},
		Env: map[string]string{
			"DOCKER_CONFIG": dockerConfigDir,
		},
		WorkingDir: opts.CoreDir,
		Bare:       true,
	}
	for _, tag := range opts.Tags {
		trimmed := strings.TrimSpace(tag)
		if trimmed != "" {
			koOpts.Tags = append(koOpts.Tags, trimmed)
		}
	}
	if len(koOpts.Tags) == 0 && strings.TrimSpace(opts.Tag) != "" {
		koOpts.Tags = []string{strings.TrimSpace(opts.Tag)}
	}
	koOpts = ko.ApplyProfileDefaults(profile, koOpts)
	if strings.TrimSpace(opts.Repository) != "" {
		koOpts.Repo = strings.TrimSpace(opts.Repository)
	}
	if strings.TrimSpace(koOpts.Repo) == "" && strings.TrimSpace(profile.KORepository) != "" {
		koOpts.Repo = strings.TrimSpace(profile.KORepository)
	}
	if strings.TrimSpace(koOpts.Repo) == "" && strings.TrimSpace(opts.AppName) != "" {
		koOpts.Repo = fmt.Sprintf("registry.fly.io/%s", strings.TrimSpace(opts.AppName))
	}
	if len(koOpts.Tags) == 0 {
		if strings.TrimSpace(profile.TagTemplate) != "" {
			koOpts.Tags = []string{strings.TrimSpace(profile.TagTemplate)}
		} else if strings.TrimSpace(opts.Tag) != "" {
			koOpts.Tags = []string{strings.TrimSpace(opts.Tag)}
		}
	}
	if len(koOpts.Tags) == 0 {
		koOpts.Tags = []string{"latest"}
	}
	if strings.TrimSpace(koOpts.Repo) == "" {
		return result, fmt.Errorf("fly pipeline: unable to resolve KO repository (provide --repository or set %s)", sharedcfg.EnvVarToolingKORepository)
	}

	ctx, cancel := context.WithTimeout(ctx, 30*time.Minute)
	defer cancel()

	refs, err := ko.Publish(ctx, koOpts)
	if err != nil {
		return result, fmt.Errorf("ko publish: %w", err)
	}
	imageRef := refs[len(refs)-1]
	result.ImageReference = imageRef

	if opts.PlatformVersion == "" {
		opts.PlatformVersion = app.PlatformVersion
	}
	if opts.PlatformVersion == "" {
		opts.PlatformVersion = "machines"
	}

	if alreadyCurrent(app, imageRef, koOpts.Tags[len(koOpts.Tags)-1]) {
		result.Skipped = true
		result.ReleaseSummary = "No changes detected; existing release already uses this image"
		result.Elapsed = time.Since(start)
		return result, nil
	}

	releaseSummary, err := flyutil.Deploy(ctx, flyutil.DeployOptions{
		AccessToken:     token,
		AppName:         opts.AppName,
		ImageRef:        imageRef,
		ConfigPath:      opts.ConfigPath,
		PlatformVersion: opts.PlatformVersion,
		Strategy:        opts.Strategy,
		Verbose:         opts.Verbose,
	})
	if err != nil {
		return result, err
	}
	result.ReleaseSummary = releaseSummary
	result.ReleaseID = extractReleaseID(releaseSummary)
	result.Elapsed = time.Since(start)
	logSummary(ctx, result)
	return result, nil
}

func resolveToken(raw, tokenFile string) (string, error) {
	if strings.TrimSpace(raw) != "" {
		return strings.TrimSpace(raw), nil
	}
	token, err := flyutil.LoadToken(tokenFile)
	if err != nil {
		return "", fmt.Errorf("fly pipeline: unable to resolve token (use fly-auth or --token): %w", err)
	}
	return token, nil
}

func alreadyCurrent(app *client.App, imageRef, tag string) bool {
	if app == nil || app.ImageDetails.Repository == "" {
		return false
	}

	current := fmt.Sprintf("%s/%s:%s", strings.TrimSuffix(app.ImageDetails.Registry, "/"), strings.TrimPrefix(app.ImageDetails.Repository, "/"), app.ImageDetails.Tag)

	if imageRef == current {
		return true
	}
	return strings.HasSuffix(imageRef, ":"+tag) && current == imageRef
}

func resolveOrganization(ctx context.Context, cl *client.Client, slug string) (*client.Organization, error) {
	if strings.TrimSpace(slug) != "" {
		org, err := cl.GetOrganizationBySlug(ctx, slug)
		if err != nil {
			return nil, fmt.Errorf("resolve organization %q: %w", slug, err)
		}
		return org, nil
	}

	orgs, err := cl.GetOrganizations(ctx)
	if err != nil {
		msg := strings.ToLower(err.Error())
		if strings.Contains(msg, "not authorized") {
			return nil, fmt.Errorf("unable to list organizations (%w). provide --org <slug> to select an organization explicitly", err)
		}
		return nil, fmt.Errorf("list organizations: %w", err)
	}
	if len(orgs) == 0 {
		return nil, errors.New("fly pipeline: no organizations available for this token")
	}
	return &orgs[0], nil
}

func ensureApp(ctx context.Context, cl *client.Client, org *client.Organization, appName string) error {
	if org == nil {
		return errors.New("fly pipeline: organization context missing for app creation")
	}
	input := client.CreateAppInput{
		OrganizationID: org.ID,
		Name:           appName,
		Machines:       true,
	}
	if _, err := cl.CreateApp(ctx, input); err != nil {
		msg := strings.ToLower(err.Error())
		if strings.Contains(msg, "already exists") {
			return nil
		}
		return fmt.Errorf("create app %q: %w", appName, err)
	}
	return waitForAppReady(ctx, cl, appName)
}

var releaseIDRegex = regexp.MustCompile(`(?i)release\s+([A-Za-z0-9_-]+)`)

func extractReleaseID(summary string) string {
	matches := releaseIDRegex.FindStringSubmatch(summary)
	if len(matches) != 2 {
		return ""
	}
	return strings.Trim(matches[1], " ,.")
}

func logSummary(ctx context.Context, result Result) {
	logger := sharedlog.Default()
	fields := map[string]any{
		"image":   result.ImageReference,
		"skipped": result.Skipped,
		"elapsed": result.Elapsed.String(),
	}
	if result.ReleaseID != "" {
		fields["release_id"] = result.ReleaseID
	}
	body, err := json.Marshal(fields)
	if err == nil {
		sharedlog.Info(ctx, logger, string(body), nil)
	} else {
		sharedlog.Info(ctx, logger, result.ReleaseSummary, sharedlog.Fields(fields))
	}
}

func isNotFound(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, client.ErrNotFound) {
		return true
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "could not find app") || strings.Contains(msg, "not found")
}

func waitForAppReady(ctx context.Context, cl *client.Client, appName string) error {
	start := time.Now()
	deadline := start.Add(30 * time.Second)
	for {
		app, err := cl.GetApp(ctx, appName)
		if err == nil && app != nil {
			if time.Since(start) < 2*time.Second {
				time.Sleep(5 * time.Second)
			}
			return nil
		}
		if err != nil && !isNotFound(err) {
			return fmt.Errorf("verify app %q: %w", appName, err)
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("timed out waiting for app %q to become available", appName)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(2 * time.Second):
		}
	}
}
