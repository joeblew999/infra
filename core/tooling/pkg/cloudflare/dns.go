package cloudflare

import (
	"context"
	"fmt"
	"strings"

	cf "github.com/cloudflare/cloudflare-go"
	sharedcfg "github.com/joeblew999/infra/core/pkg/shared/config"
)

// EnsureAppHostname ensures a CNAME record routes the Cloudflare hostname to the Fly app.
// Returns the hostname when managed or an empty string if no action was required.
func EnsureAppHostname(ctx context.Context, profile sharedcfg.ToolingProfile, settings Settings, appName string) (string, error) {
	zoneName := strings.TrimSpace(settings.ZoneName)
	zoneID := strings.TrimSpace(settings.ZoneID)
	if zoneName == "" && zoneID == "" {
		return "", nil
	}

	hostname := strings.TrimSpace(settings.AppDomain)
	if hostname == "" {
		if zoneName == "" {
			return "", nil
		}
		hostname = fmt.Sprintf("%s.%s", strings.TrimSpace(appName), zoneName)
	} else if zoneName != "" && !strings.Contains(hostname, ".") {
		hostname = fmt.Sprintf("%s.%s", hostname, zoneName)
	}

	tokenPath := strings.TrimSpace(profile.CloudflareTokenPath)
	if tokenPath == "" {
		tokenPath = DefaultTokenPath()
	}

	token, err := LoadToken(tokenPath)
	if err != nil {
		return "", fmt.Errorf("load cloudflare token: %w", err)
	}

	api, err := cf.NewWithAPIToken(token)
	if err != nil {
		return "", fmt.Errorf("create cloudflare client: %w", err)
	}

	if zoneID == "" && zoneName != "" {
		zoneID, err = api.ZoneIDByName(zoneName)
		if err != nil {
			return "", fmt.Errorf("resolve zone %s: %w", zoneName, err)
		}
	}
	if zoneID == "" {
		return "", fmt.Errorf("cloudflare zone id missing")
	}

	target := fmt.Sprintf("%s.fly.dev", strings.TrimSpace(appName))
	proxied := true

	records, _, err := api.ListDNSRecords(ctx, cf.ZoneIdentifier(zoneID), cf.ListDNSRecordsParams{Name: hostname, Type: "CNAME"})
	if err != nil {
		return "", fmt.Errorf("list dns records: %w", err)
	}

	if len(records) > 0 {
		record := records[0]
		if strings.EqualFold(strings.TrimSpace(record.Content), target) && record.Proxied != nil && *record.Proxied == proxied {
			return hostname, nil
		}
		update := cf.UpdateDNSRecordParams{
			ID:      record.ID,
			Type:    "CNAME",
			Name:    hostname,
			Content: target,
			TTL:     1,
			Proxied: &proxied,
		}
		if _, err := api.UpdateDNSRecord(ctx, cf.ZoneIdentifier(zoneID), update); err != nil {
			return "", fmt.Errorf("update dns record: %w", err)
		}
		return hostname, nil
	}

	create := cf.CreateDNSRecordParams{
		Type:    "CNAME",
		Name:    hostname,
		Content: target,
		TTL:     1,
		Proxied: &proxied,
	}
	if _, err := api.CreateDNSRecord(ctx, cf.ZoneIdentifier(zoneID), create); err != nil {
		return "", fmt.Errorf("create dns record: %w", err)
	}
	return hostname, nil
}
