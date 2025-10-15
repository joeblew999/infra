package cloudflare

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	cf "github.com/cloudflare/cloudflare-go"

	sharedcfg "github.com/joeblew999/infra/core/pkg/shared/config"
	types "github.com/joeblew999/infra/core/tooling/pkg/types"
)

// DescribeCloudflare queries Cloudflare for live DNS information tied to the
// hostname managed by tooling.
func DescribeCloudflare(ctx context.Context, profile sharedcfg.ToolingProfile, settings Settings, appName string) (types.CloudflareLiveInfo, error) {
	var info types.CloudflareLiveInfo

	tokenPath := strings.TrimSpace(profile.CloudflareTokenPath)
	if tokenPath == "" {
		tokenPath = DefaultTokenPath()
	}

	token, err := LoadToken(tokenPath)
	if err != nil {
		return info, fmt.Errorf("load cloudflare token: %w", err)
	}

	api, err := cf.NewWithAPIToken(token)
	if err != nil {
		return info, fmt.Errorf("create cloudflare client: %w", err)
	}

	zoneName := strings.TrimSpace(settings.ZoneName)
	zoneID := strings.TrimSpace(settings.ZoneID)
	if zoneID == "" && zoneName != "" {
		zoneID, err = api.ZoneIDByName(zoneName)
		if err != nil {
			return info, fmt.Errorf("resolve zone id: %w", err)
		}
	}
	if zoneID == "" {
		return info, fmt.Errorf("cloudflare zone id not configured")
	}

	details, err := api.ZoneDetails(ctx, zoneID)
	if err != nil {
		return info, fmt.Errorf("zone details: %w", err)
	}

	info = types.CloudflareLiveInfo{
		ZoneName:  details.Name,
		ZoneID:    details.ID,
		AccountID: details.Account.ID,
		UpdatedAt: time.Now().UTC(),
	}
	if info.AccountID == "" {
		info.AccountID = strings.TrimSpace(settings.AccountID)
	}

	hostname := strings.TrimSpace(settings.AppDomain)
	if hostname == "" && zoneName != "" {
		hostname = fmt.Sprintf("%s.%s", strings.TrimSpace(appName), zoneName)
	}

	if hostname != "" {
		params := cf.ListDNSRecordsParams{Name: hostname, Type: "CNAME"}
		records, _, err := api.ListDNSRecords(ctx, cf.ZoneIdentifier(zoneID), params)
		if err != nil {
			return info, fmt.Errorf("list dns records: %w", err)
		}
		info.Hostname = hostname
		if len(records) > 0 {
			record := records[0]
			info.Target = strings.TrimSpace(record.Content)
			if record.Proxied != nil {
				info.Proxied = *record.Proxied
			}
			info.TTL = record.TTL
		}
	}

	if bucket := strings.TrimSpace(settings.R2Bucket); bucket != "" && info.AccountID != "" {
		b, err := api.GetR2Bucket(ctx, cf.AccountIdentifier(info.AccountID), bucket)
		if err != nil {
			var cfErr *cf.Error
			if errors.As(err, &cfErr) && cfErr.StatusCode == http.StatusNotFound {
				// bucket does not exist yet; keep verification successful but leave live info empty
			} else {
				return info, fmt.Errorf("get r2 bucket %s: %w", bucket, err)
			}
		} else {
			info.Bucket = b.Name
			info.BucketRegion = strings.TrimSpace(b.Location)
		}
	}

	return info, nil
}
