package types

import (
	"io"
	"time"
)

// DeployRequest represents the configurable inputs for a deployment.
type DeployRequest struct {
	AppName   string
	OrgSlug   string
	Region    string
	Repo      string
	Verbose   bool
	NoBrowser bool
	Stdin     io.Reader
	Stdout    io.Writer
	Stderr    io.Writer
}

// DeployResult captures the outcome of a release.
type DeployResult struct {
	ImageReference string
	ReleaseSummary string
	ReleaseID      string
	Elapsed        time.Duration
	AppName        string
	OrgSlug        string
}

// ProfileSummary captures key fields from a tooling profile.
type ProfileSummary struct {
	Name         string `json:"name"`
	Mode         string `json:"mode"`
	FlyApp       string `json:"fly_app"`
	FlyOrg       string `json:"fly_org"`
	FlyRegion    string `json:"fly_region"`
	KORepository string `json:"ko_repository"`
}

// FlySettingsSummary summarises cached Fly preferences.
type FlySettingsSummary struct {
	OrgSlug    string    `json:"org_slug"`
	RegionCode string    `json:"region_code"`
	RegionName string    `json:"region_name"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// CloudflareSummary summarises cached Cloudflare preferences.
type CloudflareSummary struct {
	ZoneName  string    `json:"zone_name"`
	ZoneID    string    `json:"zone_id"`
	AccountID string    `json:"account_id"`
	R2Bucket  string    `json:"r2_bucket"`
	R2Region  string    `json:"r2_region"`
	AppDomain string    `json:"app_domain"`
	UpdatedAt time.Time `json:"updated_at"`
}

// FlyLiveInfo describes live data returned from the Fly API.
type FlyLiveInfo struct {
	AppName         string    `json:"app_name"`
	Hostname        string    `json:"hostname"`
	URL             string    `json:"url"`
	OrgSlug         string    `json:"org_slug"`
	Status          string    `json:"status"`
	Deployed        bool      `json:"deployed"`
	Version         int       `json:"version"`
	ReleaseStatus   string    `json:"release_status,omitempty"`
	ReleaseVersion  int       `json:"release_version,omitempty"`
	PlatformVersion string    `json:"platform_version,omitempty"`
	PrimaryRegion   string    `json:"primary_region,omitempty"`
	CNAME           string    `json:"cname,omitempty"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// CloudflareLiveInfo describes the DNS state recorded at Cloudflare.
type CloudflareLiveInfo struct {
	ZoneName     string    `json:"zone_name"`
	ZoneID       string    `json:"zone_id"`
	AccountID    string    `json:"account_id"`
	Hostname     string    `json:"hostname"`
	Target       string    `json:"target"`
	Proxied      bool      `json:"proxied"`
	TTL          int       `json:"ttl"`
	Bucket       string    `json:"bucket,omitempty"`
	BucketRegion string    `json:"bucket_region,omitempty"`
	UpdatedAt    time.Time `json:"updated_at"`
}
