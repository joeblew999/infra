package orchestrator

import (
	"context"

	"github.com/joeblew999/infra/core/tooling/pkg/cloudflare"
	"github.com/joeblew999/infra/core/tooling/pkg/fly"
	profiles "github.com/joeblew999/infra/core/tooling/pkg/profiles"
	types "github.com/joeblew999/infra/core/tooling/pkg/types"
)

// Status describes the current tooling context including cached provider settings.
type Status struct {
	ProfileName    string                    `json:"profile_name"`
	Profile        types.ProfileSummary      `json:"profile"`
	RepoRoot       string                    `json:"repo_root"`
	CoreDir        string                    `json:"core_dir"`
	Fly            types.FlySettingsSummary  `json:"fly"`
	Cloudflare     types.CloudflareSummary   `json:"cloudflare"`
	FlyLive        *types.FlyLiveInfo        `json:"fly_live,omitempty"`
	CloudflareLive *types.CloudflareLiveInfo `json:"cloudflare_live,omitempty"`
}

// StatusSnapshot resolves context and cached provider preferences for UI consumption.
func StatusSnapshot(ctx context.Context, opts profiles.ContextOptions) (Status, error) {
	ctxInfo, err := profiles.ResolveContext(opts)
	if err != nil {
		return Status{}, err
	}

	flySummary := types.FlySettingsSummary{
		OrgSlug:    ctxInfo.Fly.OrgSlug,
		RegionCode: ctxInfo.Fly.RegionCode,
		RegionName: ctxInfo.Fly.RegionName,
		UpdatedAt:  ctxInfo.Fly.UpdatedAt,
	}

	cfSummary := types.CloudflareSummary{
		ZoneName:  ctxInfo.Cloudflare.ZoneName,
		ZoneID:    ctxInfo.Cloudflare.ZoneID,
		AccountID: ctxInfo.Cloudflare.AccountID,
		R2Bucket:  ctxInfo.Cloudflare.R2Bucket,
		R2Region:  ctxInfo.Cloudflare.R2Region,
		AppDomain: ctxInfo.Cloudflare.AppDomain,
		UpdatedAt: ctxInfo.Cloudflare.UpdatedAt,
	}

	profileSummary := types.ProfileSummary{
		Name:         ctxInfo.Profile.Name,
		Mode:         string(ctxInfo.Profile.Mode),
		FlyApp:       ctxInfo.Profile.FlyApp,
		FlyOrg:       ctxInfo.Profile.FlyOrg,
		FlyRegion:    ctxInfo.Profile.FlyRegion,
		KORepository: ctxInfo.Profile.KORepository,
	}

	status := Status{
		ProfileName: ctxInfo.ProfileName,
		Profile:     profileSummary,
		RepoRoot:    ctxInfo.RepoRoot,
		CoreDir:     ctxInfo.CoreDir,
		Fly:         flySummary,
		Cloudflare:  cfSummary,
	}

	if profileSummary.FlyApp != "" {
		if live, err := fly.DescribeFly(ctx, ctxInfo.Profile, profileSummary.FlyApp); err == nil {
			status.FlyLive = &live
		}
	}

	if cfSummary.ZoneID != "" || cfSummary.ZoneName != "" {
		if live, err := cloudflare.DescribeCloudflare(ctx, ctxInfo.Profile, ctxInfo.Cloudflare, profileSummary.FlyApp); err == nil {
			status.CloudflareLive = &live
		}
	}

	return status, nil
}
