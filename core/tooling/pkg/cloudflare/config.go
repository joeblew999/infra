package cloudflare

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strings"

	cf "github.com/cloudflare/cloudflare-go"

	sharedcfg "github.com/joeblew999/infra/core/pkg/shared/config"
	sharedprompt "github.com/joeblew999/infra/core/pkg/shared/prompt"
)

// ConfigureCloudflarePreferences handles zone and account selection.
func ConfigureCloudflarePreferences(ctx context.Context, api *cf.API, in io.Reader, out io.Writer, profile sharedcfg.ToolingProfile) error {
	reader := bufio.NewReader(in)
	settings, err := LoadSettings()
	if err != nil {
		return fmt.Errorf("load cloudflare settings: %w", err)
	}

	updated := false
	zoneName := strings.TrimSpace(settings.ZoneName)
	accountID := strings.TrimSpace(settings.AccountID)

	// Try to use existing zone configuration
	if zoneName != "" {
		zoneID, err := api.ZoneIDByName(zoneName)
		if err != nil {
			fmt.Fprintf(out, "Stored Cloudflare zone %s is not accessible: %v\n", zoneName, err)
			zoneName = ""
			updated = true
		} else {
			zoneDetails, err := api.ZoneDetails(ctx, zoneID)
			if err == nil {
				accountID = strings.TrimSpace(zoneDetails.Account.ID)
				settings.ZoneID = zoneDetails.ID
				fmt.Fprintf(out, "✓ Using Cloudflare zone %s (account %s)\n", zoneDetails.Name, zoneDetails.Account.ID)
			} else {
				fmt.Fprintf(out, "Unable to load zone details for %s: %v\n", zoneName, err)
				zoneName = ""
				updated = true
			}
		}
	}

	// No valid zone - prompt user to select one
	if zoneName == "" {
		fmt.Fprintln(out)
		fmt.Fprintln(out, "Listing accessible Cloudflare zones...")
		zones, err := api.ListZones(ctx)
		if err != nil {
			return fmt.Errorf("list zones: %w", err)
		}
		if len(zones) == 0 {
			fmt.Fprintln(out, "⚠  No zones found; ensure token has Zone:Zone:Read permission")
			return nil
		}
		
		fmt.Fprintln(out)
		fmt.Fprintln(out, "Available zones:")
		for i, z := range zones {
			fmt.Fprintf(out, "  %d. %s (ID: %s)\n", i+1, z.Name, z.ID)
		}
		fmt.Fprintln(out)
		
		zoneName, err = sharedprompt.String(reader, out, "Enter zone name:")
		if err != nil {
			return err
		}
		zoneName = strings.TrimSpace(zoneName)
		
		zoneID, err := api.ZoneIDByName(zoneName)
		if err != nil {
			return fmt.Errorf("resolve zone %s: %w", zoneName, err)
		}
		
		zoneDetails, err := api.ZoneDetails(ctx, zoneID)
		if err != nil {
			return fmt.Errorf("fetch zone details: %w", err)
		}
		
		settings.ZoneName = zoneDetails.Name
		settings.ZoneID = zoneDetails.ID
		accountID = strings.TrimSpace(zoneDetails.Account.ID)
		updated = true
		
		fmt.Fprintf(out, "✓ Selected zone: %s\n", zoneDetails.Name)
	}

	if strings.TrimSpace(accountID) != "" && accountID != settings.AccountID {
		settings.AccountID = accountID
		updated = true
	}

	if updated {
		if err := SaveSettings(settings); err != nil {
			return fmt.Errorf("save cloudflare settings: %w", err)
		}
		fmt.Fprintln(out, "✓ Cloudflare settings saved")
	}

	return nil
}
