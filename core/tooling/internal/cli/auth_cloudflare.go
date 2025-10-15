package cli

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"

	sharedcfg "github.com/joeblew999/infra/core/pkg/shared/config"
	"github.com/joeblew999/infra/core/tooling/pkg/auth"
	cloudflareprefs "github.com/joeblew999/infra/core/tooling/pkg/cloudflare"
	profiles "github.com/joeblew999/infra/core/tooling/pkg/profiles"
	"github.com/joeblew999/infra/core/tooling/pkg/cloudflare"
)

func newAuthCloudflareCommand(profileFlag *string) *cobra.Command {
	var (
		tokenInput string
		tokenPath  string
		noBrowser  bool
	)

	cmd := &cobra.Command{
		Use:   "cloudflare",
		Short: "Authenticate with Cloudflare (idempotent)",
		Long: `Authenticate with Cloudflare for DNS and R2 storage.

This command is idempotent - run it anytime to verify or refresh credentials.

TWO AUTHENTICATION METHODS:

1. API Token (Recommended - Most Secure):
   $ ./core-tool auth cloudflare

   Create a scoped token at: https://dash.cloudflare.com/profile/api-tokens
   
   Required permissions:
     • Zone:DNS:Edit
     • Zone:Zone:Read  
     • Account:R2:Edit (if using R2 storage)

2. Bootstrap (One-time Setup):
   $ ./core-tool auth cloudflare bootstrap

   Uses your Global API Key to automatically create a scoped token.
   Requires Cloudflare email + Global API Key once.
   After bootstrap, the Global API Key is not stored.

WHAT THIS COMMAND DOES:
  ✓ Checks if you already have valid credentials (idempotent)
  ✓ Verifies token permissions with Cloudflare API
  ✓ Configures zone and account settings
  ✓ Tests connectivity to your Cloudflare zone
  ✓ Saves credentials securely for future use`,
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, _ := profiles.ResolveProfile(strings.TrimSpace(*profileFlag))
			path := profiles.FirstNonEmpty(tokenPath, profile.CloudflareTokenPath, cloudflareprefs.DefaultTokenPath())
			
			// Check existing token
			token, err := cloudflareprefs.LoadToken(path)
			if err == nil {
				fmt.Fprintln(cmd.OutOrStdout(), "Checking existing Cloudflare token...")
				if body, _, verr := auth.VerifyCloudflareToken(cmd.Context(), token); verr == nil {
					fmt.Fprintf(cmd.OutOrStdout(), "✓ Cloudflare token is valid (status: %s)\n", strings.TrimSpace(body.Status))
					settings, serr := cloudflareprefs.LoadSettings()
					if serr == nil {
						if err := printCloudflareLive(cmd.Context(), cmd.OutOrStdout(), profile, settings, strings.TrimSpace(profile.FlyApp)); err == nil {
							fmt.Fprintln(cmd.OutOrStdout(), "✓ Cloudflare authentication verified successfully")
							fmt.Fprintln(cmd.OutOrStdout(), "✓ All checks passed - you're ready to deploy!")
							return nil
						}
						fmt.Fprintf(cmd.OutOrStdout(), "\n⚠  Token valid but zone unreachable: %v\n", err)
						fmt.Fprintln(cmd.OutOrStdout(), "   Continuing with zone configuration...")
					} else {
						fmt.Fprintf(cmd.OutOrStdout(), "\n⚠  Token valid but settings incomplete: %v\n", serr)
						fmt.Fprintln(cmd.OutOrStdout(), "   Continuing with configuration...")
					}
				} else {
					fmt.Fprintf(cmd.OutOrStdout(), "⚠  Cached token invalid: %v\n", verr)
					fmt.Fprintln(cmd.OutOrStdout(), "   Requesting new token...")
				}
			} else {
				fmt.Fprintln(cmd.OutOrStdout(), "No cached Cloudflare token found")
				fmt.Fprintln(cmd.OutOrStdout(), "Starting authentication flow...")
			}
			
			prompter := auth.NewIOPrompter(cmd.InOrStdin(), cmd.OutOrStdout(), noBrowser)
			if err := auth.RunCloudflareAuth(cmd.Context(), profile, tokenInput, path, noBrowser, cmd.InOrStdin(), cmd.OutOrStdout(), prompter); err != nil {
				return fmt.Errorf("cloudflare auth failed: %w", err)
			}
			
			fmt.Fprintln(cmd.OutOrStdout(), "✓ Cloudflare authentication complete!")
			return nil
		},
	}

	cmd.Flags().StringVar(&tokenInput, "token", "", "Cloudflare API token (skips interactive prompt)")
	cmd.Flags().StringVar(&tokenPath, "path", "", "Override token file path")
	cmd.Flags().BoolVar(&noBrowser, "no-browser", false, "Print URLs instead of opening browser")

	cmd.AddCommand(newAuthCloudflareVerifyCommand(profileFlag))
	cmd.AddCommand(newAuthCloudflareBootstrapCommand(profileFlag))

	return cmd
}

func newAuthCloudflareVerifyCommand(profileFlag *string) *cobra.Command {
	var tokenPath string

	cmd := &cobra.Command{
		Use:   "verify",
		Short: "Verify cached Cloudflare credentials",
		Long: `Verify that your cached Cloudflare token is valid and has correct permissions.

This command checks:
  ✓ Token exists and is properly formatted
  ✓ Token is valid with Cloudflare API
  ✓ Zone configuration is correct
  ✓ DNS records are accessible
  ✓ R2 buckets are accessible (if configured)`,
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, _ := profiles.ResolveProfile(strings.TrimSpace(*profileFlag))
			path := profiles.FirstNonEmpty(tokenPath, profile.CloudflareTokenPath, cloudflareprefs.DefaultTokenPath())
			
			fmt.Fprintln(cmd.OutOrStdout(), "Verifying Cloudflare credentials...")
			
			token, err := cloudflareprefs.LoadToken(path)
			if err != nil {
				return fmt.Errorf("✗ load cloudflare token: %w", err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "✓ Token loaded from %s\n", path)
			
			body, _, err := auth.VerifyCloudflareToken(cmd.Context(), token)
			if err != nil {
				return fmt.Errorf("✗ verify cloudflare token: %w", err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "✓ Token status: %s\n", strings.TrimSpace(body.Status))
			
			settings, err := cloudflareprefs.LoadSettings()
			if err != nil {
				return fmt.Errorf("✗ load cloudflare settings: %w", err)
			}
			
			if err := printCloudflareLive(cmd.Context(), cmd.OutOrStdout(), profile, settings, strings.TrimSpace(profile.FlyApp)); err != nil {
				return fmt.Errorf("✗ zone connectivity check: %w", err)
			}
			
			fmt.Fprintln(cmd.OutOrStdout(), "✓ All Cloudflare checks passed!")
			return nil
		},
	}

	cmd.Flags().StringVar(&tokenPath, "path", "", "Override token file path")
	return cmd
}

func newAuthCloudflareBootstrapCommand(profileFlag *string) *cobra.Command {
	var (
		email     string
		apiKey    string
		tokenName string
		tokenPath string
		includeR2 bool
	)

	cmd := &cobra.Command{
		Use:   "bootstrap",
		Short: "Create scoped token using Global API Key (one-time setup)",
		Long: `Bootstrap Cloudflare authentication using your Global API Key.

This is a ONE-TIME setup command that:
  1. Uses your Global API Key to create a scoped API token
  2. Saves the scoped token for future use
  3. Does NOT store your Global API Key

After running this once, use './core-tool auth cloudflare' to verify.

REQUIRED INPUTS:
  • Cloudflare account email
  • Global API Key (get it from: https://dash.cloudflare.com/profile/api-tokens)

The generated scoped token will have:
  • Zone:DNS:Edit permissions
  • Zone:Zone:Read permissions
  • Account:R2:Edit permissions (if --include-r2 flag is used)

This is more secure than using your Global API Key directly.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintln(cmd.OutOrStdout(), "Starting Cloudflare bootstrap...")
			fmt.Fprintln(cmd.OutOrStdout(), "This will create a scoped API token using your Global API Key")
			
			profile, _ := profiles.ResolveProfile(strings.TrimSpace(*profileFlag))
			options := auth.BootstrapOptions{
				Email:        strings.TrimSpace(email),
				GlobalAPIKey: strings.TrimSpace(apiKey),
				TokenName:    strings.TrimSpace(tokenName),
				IncludeR2:    includeR2,
				TokenPath:    strings.TrimSpace(tokenPath),
			}
			prompter := auth.NewIOPrompter(cmd.InOrStdin(), cmd.OutOrStdout(), false)
			
			if err := auth.RunCloudflareBootstrap(cmd.Context(), profile, options, cmd.InOrStdin(), cmd.OutOrStdout(), prompter); err != nil {
				return fmt.Errorf("bootstrap failed: %w", err)
			}
			
			fmt.Fprintln(cmd.OutOrStdout(), "✓ Cloudflare bootstrap complete!")
			fmt.Fprintln(cmd.OutOrStdout(), "✓ Scoped token created and saved")
			fmt.Fprintln(cmd.OutOrStdout(), "Next: Run './core-tool auth cloudflare' to verify your setup")
			return nil
		},
	}

	cmd.Flags().StringVar(&email, "email", "", "Cloudflare account email")
	cmd.Flags().StringVar(&apiKey, "global-key", "", "Cloudflare Global API Key")
	cmd.Flags().StringVar(&tokenName, "token-name", "", "Custom name for generated token")
	cmd.Flags().StringVar(&tokenPath, "path", "", "Override token file path")
	cmd.Flags().BoolVar(&includeR2, "include-r2", false, "Include R2 storage permissions")
	
	return cmd
}

func printCloudflareLive(ctx context.Context, out io.Writer, profile sharedcfg.ToolingProfile, settings cloudflareprefs.Settings, appName string) error {
	zoneName := strings.TrimSpace(settings.ZoneName)
	if zoneName == "" {
		fmt.Fprintln(out, "⚠  No Cloudflare zone configured; skipping live check")
		fmt.Fprintln(out, "   (Zone will be configured during first deployment)")
		return nil
	}
	
	info, err := cloudflare.DescribeCloudflare(ctx, profile, settings, strings.TrimSpace(appName))
	if err != nil {
		return err
	}
	
	fmt.Fprintf(out, "✓ Cloudflare zone '%s' is reachable\n", info.ZoneName)
	if info.Hostname != "" {
		proxyStatus := "not proxied"
		if info.Proxied {
			proxyStatus = "proxied"
		}
		fmt.Fprintf(out, "  • DNS: %s → %s (%s, TTL=%d)\n", info.Hostname, info.Target, proxyStatus, info.TTL)
	}
	if info.Bucket != "" {
		fmt.Fprintf(out, "  • R2: bucket '%s' (region=%s)\n", info.Bucket, info.BucketRegion)
	}
	return nil
}
