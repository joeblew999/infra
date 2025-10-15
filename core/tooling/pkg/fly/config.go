package fly

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	flyapi "github.com/superfly/fly-go"

)

// ConfigureFlyPreferences configures Fly organization and region preferences.
func ConfigureFlyPreferences(ctx context.Context, client *flyapi.Client, out io.Writer, defaultOrg, defaultRegion string) error {
	settings, err := LoadSettings()
	if err != nil {
		return fmt.Errorf("load fly settings: %w", err)
	}

	updated := false

	orgSlug := strings.TrimSpace(settings.OrgSlug)
	if orgSlug != "" {
		if _, err := client.GetOrganizationBySlug(ctx, orgSlug); err != nil {
			fmt.Fprintf(out, "Stored Fly organization %s is not accessible: %v\n", orgSlug, err)
			orgSlug = ""
			updated = true
		}
	}
	if orgSlug == "" && strings.TrimSpace(defaultOrg) != "" {
		if _, err := client.GetOrganizationBySlug(ctx, strings.TrimSpace(defaultOrg)); err == nil {
			orgSlug = strings.TrimSpace(defaultOrg)
			fmt.Fprintf(out, "Using profile Fly organization %s.\n", orgSlug)
			updated = true
		}
	}
	if orgSlug == "" {
		orgs, err := client.GetOrganizations(ctx)
		if err != nil {
			return fmt.Errorf("list fly organizations: %w", err)
		}
		if len(orgs) == 0 {
			return errors.New("fly token has no accessible organizations")
		}
		orgSlug = strings.TrimSpace(orgs[0].Slug)
		fmt.Fprintf(out, "Auto-selected Fly organization %s (%s).\n", orgs[0].Name, orgSlug)
		updated = true
	}
	settings.OrgSlug = orgSlug

	regions, _, err := client.PlatformRegions(ctx)
	if err != nil {
		return fmt.Errorf("list fly regions: %w", err)
	}
	if len(regions) == 0 {
		return errors.New("fly API returned no regions")
	}

	lookupRegion := func(code string) (string, bool) {
		for _, region := range regions {
			if strings.EqualFold(region.Code, code) {
				return region.Name, true
			}
		}
		return "", false
	}

	regionCode := strings.TrimSpace(settings.RegionCode)
	if regionCode != "" {
		if name, ok := lookupRegion(regionCode); ok {
			settings.RegionName = name
		} else {
			fmt.Fprintf(out, "Stored Fly region %s is not available.\n", regionCode)
			regionCode = ""
			updated = true
		}
	}
	if regionCode == "" && strings.TrimSpace(defaultRegion) != "" {
		if name, ok := lookupRegion(strings.TrimSpace(defaultRegion)); ok {
			regionCode = strings.TrimSpace(defaultRegion)
			settings.RegionName = name
			fmt.Fprintf(out, "Using profile Fly region %s.\n", regionCode)
			updated = true
		}
	}
	if regionCode == "" {
		regionCode = regions[0].Code
		settings.RegionName = regions[0].Name
		fmt.Fprintf(out, "Auto-selected Fly region %s (%s).\n", regions[0].Name, regions[0].Code)
		updated = true
	}
	settings.RegionCode = regionCode

	if updated {
		if err := SaveSettings(settings); err != nil {
			return fmt.Errorf("save fly settings: %w", err)
		}
	}

	fmt.Fprintf(out, "Fly preferences: org=%s, region=%s (%s)\n", settings.OrgSlug, settings.RegionCode, settings.RegionName)
	return nil
}
