package fly

import (
	"context"
	"fmt"
	"strings"
	"time"

	flyapi "github.com/superfly/fly-go"

	sharedbuild "github.com/joeblew999/infra/core/pkg/shared/build"
	sharedcfg "github.com/joeblew999/infra/core/pkg/shared/config"
	types "github.com/joeblew999/infra/core/tooling/pkg/types"
)

// DescribeFly queries Fly for live information about the app using the
// credentials defined in the tooling profile.
func DescribeFly(ctx context.Context, profile sharedcfg.ToolingProfile, appName string) (types.FlyLiveInfo, error) {
	var info types.FlyLiveInfo

	tokenPath := strings.TrimSpace(profile.TokenPath)
	if tokenPath == "" {
		tokenPath = DefaultTokenPath()
	}

	token, err := LoadToken(tokenPath)
	if err != nil {
		return info, fmt.Errorf("load fly token: %w", err)
	}

	if base := strings.TrimSpace(profile.FlyAPIBase); base != "" {
		flyapi.SetBaseURL(base)
	} else {
		flyapi.SetBaseURL(sharedcfg.Tooling().Active.FlyAPIBase)
	}

	build := sharedbuild.Get()
	client := flyapi.NewClientFromOptions(flyapi.ClientOptions{
		AccessToken: token,
		Name:        "tooling-app",
		Version:     strings.TrimSpace(build.Version),
		Logger:      NewLogger("tooling-app", false),
	})

	app, err := client.GetApp(ctx, appName)
	if err != nil {
		return info, err
	}

	info = types.FlyLiveInfo{
		AppName:         app.Name,
		Hostname:        strings.TrimSpace(app.Hostname),
		URL:             strings.TrimSpace(app.AppURL),
		OrgSlug:         strings.TrimSpace(app.Organization.Slug),
		Status:          strings.TrimSpace(app.Status),
		Deployed:        app.Deployed,
		Version:         app.Version,
		CNAME:           strings.TrimSpace(app.CNAMETarget),
		PlatformVersion: strings.TrimSpace(app.PlatformVersion),
		PrimaryRegion:   strings.TrimSpace(profile.FlyRegion),
		UpdatedAt:       time.Now().UTC(),
	}

	if app.CurrentRelease != nil {
		info.ReleaseStatus = strings.TrimSpace(app.CurrentRelease.Status)
		info.ReleaseVersion = app.CurrentRelease.Version
	}

	return info, nil
}
