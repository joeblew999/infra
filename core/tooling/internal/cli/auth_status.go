package cli

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/joeblew999/infra/core/tooling/pkg/app"
	profiles "github.com/joeblew999/infra/core/tooling/pkg/profiles"
)

func newAuthStatusCommand(profileFlag *string) *cobra.Command {
	var asJSON bool

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show cached Fly and Cloudflare settings",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			svc := app.New()
			status, err := svc.Status(ctx, profiles.ContextOptions{ProfileOverride: strings.TrimSpace(*profileFlag)})
			if err != nil {
				return err
			}

			if asJSON {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(status)
			}

			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "Profile: %s\n", status.ProfileName)
			fmt.Fprintf(out, "  Fly App:    %s\n", status.Profile.FlyApp)
			fmt.Fprintf(out, "  Fly Org:    %s\n", status.Fly.OrgSlug)
			fmt.Fprintf(out, "  Fly Region: %s (%s)\n", status.Fly.RegionCode, status.Fly.RegionName)
			fmt.Fprintf(out, "  Cached at:  %s\n", status.Fly.UpdatedAt.Format(time.RFC3339))
			fmt.Fprintln(out)
			fmt.Fprintf(out, "  Cloudflare Zone: %s\n", status.Cloudflare.ZoneName)
			fmt.Fprintf(out, "  Account ID:      %s\n", status.Cloudflare.AccountID)
			if status.Cloudflare.AppDomain != "" {
				fmt.Fprintf(out, "  Hostname:       %s\n", status.Cloudflare.AppDomain)
			}
			if status.Cloudflare.R2Bucket != "" {
				fmt.Fprintf(out, "  R2 Bucket:      %s (%s)\n", status.Cloudflare.R2Bucket, status.Cloudflare.R2Region)
			}
			if !status.Cloudflare.UpdatedAt.IsZero() {
				fmt.Fprintf(out, "  Cached at:      %s\n", status.Cloudflare.UpdatedAt.Format(time.RFC3339))
			}
			fmt.Fprintln(out)
			if status.FlyLive != nil {
				fmt.Fprintf(out, "  Fly Live: version=%d status=%s hostname=%s\n", status.FlyLive.Version, status.FlyLive.Status, status.FlyLive.Hostname)
			}
			if status.CloudflareLive != nil {
				fmt.Fprintf(out, "  Cloudflare DNS: host=%s -> %s (proxied=%t)\n", status.CloudflareLive.Hostname, status.CloudflareLive.Target, status.CloudflareLive.Proxied)
			}
			fmt.Fprintln(out)
			fmt.Fprintf(out, "Repo: %s\n", status.RepoRoot)
			fmt.Fprintf(out, "Core: %s\n", status.CoreDir)
			return nil
		},
	}

	cmd.Flags().BoolVar(&asJSON, "json", false, "Output status as JSON")
	return cmd
}
