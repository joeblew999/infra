package cloudflare

import (
	"context"
	"fmt"
	"io"
	"strings"

	cf "github.com/cloudflare/cloudflare-go"
)

// verifyCloudflareToken checks if a token is valid with the Cloudflare API.
func verifyCloudflareToken(ctx context.Context, token string) (cf.APITokenVerifyBody, *cf.API, error) {
	api, err := cf.NewWithAPIToken(strings.TrimSpace(token))
	if err != nil {
		return cf.APITokenVerifyBody{}, nil, err
	}
	body, err := api.VerifyAPIToken(ctx)
	if err != nil {
		return cf.APITokenVerifyBody{}, nil, err
	}
	return body, api, nil
}

// VerifyCloudflareToken is the exported version for external use.
func VerifyCloudflareToken(ctx context.Context, token string) (cf.APITokenVerifyBody, *cf.API, error) {
	return verifyCloudflareToken(ctx, token)
}

// verifyTokenPermissions tests if a token has the required permissions.
func verifyTokenPermissions(ctx context.Context, api *cf.API, out io.Writer) error {
	fmt.Fprintln(out, "Testing token permissions...")
	fmt.Fprintln(out)
	
	// Test 1: Zone listing (REQUIRED)
	fmt.Fprint(out, "  • Checking Zone:Zone:Read permission... ")
	zones, err := api.ListZones(ctx)
	if err != nil {
		fmt.Fprintln(out, "✗ FAILED")
		fmt.Fprintf(out, "    Error: %v\n", err)
		return fmt.Errorf("token lacks Zone:Zone:Read permission: %w", err)
	}
	fmt.Fprintf(out, "✓ OK (found %d zones)\n", len(zones))
	
	// Test 2: DNS record access (Optional)
	if len(zones) > 0 {
		zoneID := zones[0].ID
		zoneName := zones[0].Name
		
		fmt.Fprintf(out, "  • Checking Zone:DNS access for %s... ", zoneName)
		records, _, err := api.ListDNSRecords(ctx, cf.ZoneIdentifier(zoneID), cf.ListDNSRecordsParams{})
		if err != nil {
			fmt.Fprintln(out, "⚠  LIMITED")
			fmt.Fprintln(out, "    Note: Cannot list DNS records. Zone:DNS:Edit may be missing.")
		} else {
			fmt.Fprintf(out, "✓ OK (found %d DNS records)\n", len(records))
		}
	}
	
	// Test 3: Account/R2 access (Optional)
	fmt.Fprint(out, "  • Checking Account access... ")
	accounts, _, err := api.Accounts(ctx, cf.AccountsListParams{})
	if err != nil {
		fmt.Fprintln(out, "⚠  LIMITED")
		fmt.Fprintln(out, "    Note: Cannot list accounts. R2 access may be unavailable.")
	} else {
		fmt.Fprintf(out, "✓ OK (found %d accounts)\n", len(accounts))
	}
	
	fmt.Fprintln(out)
	fmt.Fprintln(out, "✓ Token permissions verified!")
	return nil
}
