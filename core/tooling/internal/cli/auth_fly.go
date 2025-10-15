package cli

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/spf13/cobra"

	sharedcfg "github.com/joeblew999/infra/core/pkg/shared/config"
	"github.com/joeblew999/infra/core/tooling/pkg/auth"
	flyprefs "github.com/joeblew999/infra/core/tooling/pkg/fly"
	"github.com/joeblew999/infra/core/tooling/pkg/fly"
	profiles "github.com/joeblew999/infra/core/tooling/pkg/profiles"
)

func newAuthFlyCommand(profileFlag *string) *cobra.Command {
	var (
		tokenInput string
		tokenPath  string
		noBrowser  bool
		timeout    time.Duration
	)

	cmd := &cobra.Command{
		Use:   "fly",
		Short: "Authenticate with Fly.io (idempotent)",
		Long: `Authenticate with Fly.io for deployments.

This command is idempotent - run it anytime to verify or refresh credentials.

AUTHENTICATION METHOD:
  Interactive browser-based login (recommended and most secure)
  OR provide token directly with --token flag

WHAT THIS COMMAND DOES:
  ✓ Checks if you already have valid credentials (idempotent)
  ✓ Opens browser for Fly.io authentication if needed
  ✓ Verifies token with Fly.io API
  ✓ Tests connectivity to your Fly app (if configured)
  ✓ Saves credentials securely for future use

EXAMPLE USAGE:
  # Interactive authentication (recommended):
  $ ./core-tool auth fly

  # With explicit token:
  $ ./core-tool auth fly --token <your-fly-token>

  # Without opening browser:
  $ ./core-tool auth fly --no-browser`,
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, _ := profiles.ResolveProfile(strings.TrimSpace(*profileFlag))
			path := profiles.FirstNonEmpty(tokenPath, profile.TokenPath, flyprefs.DefaultTokenPath())
			
			// Check existing token
			token, err := flyprefs.LoadToken(path)
			if err == nil {
				fmt.Fprintln(cmd.OutOrStdout(), "Checking existing Fly.io token...")
				if identity, _, verr := auth.VerifyFlyToken(cmd.Context(), profile, token); verr == nil {
					fmt.Fprintf(cmd.OutOrStdout(), "✓ Fly.io token is valid for %s\n", strings.TrimSpace(identity))
					if err := printFlyLive(cmd.Context(), cmd.OutOrStdout(), profile, strings.TrimSpace(profile.FlyApp)); err == nil {
						fmt.Fprintln(cmd.OutOrStdout(), "✓ Fly.io authentication verified successfully")
						fmt.Fprintln(cmd.OutOrStdout(), "✓ All checks passed - you're ready to deploy!")
						return nil
					}
					appName := strings.TrimSpace(profile.FlyApp)
					if appName != "" {
						fmt.Fprintf(cmd.OutOrStdout(), "\n⚠  Token valid but app '%s' unreachable: %v\n", appName, err)
						fmt.Fprintln(cmd.OutOrStdout(), "   (This is OK if the app doesn't exist yet)")
					}
					return nil
				} else {
					fmt.Fprintf(cmd.OutOrStdout(), "⚠  Cached token invalid: %v\n", verr)
					fmt.Fprintln(cmd.OutOrStdout(), "   Requesting new token...")
				}
			} else {
				fmt.Fprintln(cmd.OutOrStdout(), "No cached Fly.io token found")
				fmt.Fprintln(cmd.OutOrStdout(), "Starting authentication flow...")
			}
			
			prompter := auth.NewIOPrompter(cmd.InOrStdin(), cmd.OutOrStdout(), noBrowser)
			if err := auth.RunFlyAuth(cmd.Context(), profile, tokenInput, path, noBrowser, timeout, cmd.InOrStdin(), cmd.OutOrStdout(), prompter); err != nil {
				return fmt.Errorf("fly auth failed: %w", err)
			}
			
			fmt.Fprintln(cmd.OutOrStdout(), "✓ Fly.io authentication complete!")
			return nil
		},
	}

	cmd.Flags().StringVar(&tokenInput, "token", "", "Fly API token (skips interactive browser login)")
	cmd.Flags().StringVar(&tokenPath, "path", "", "Override token file path")
	cmd.Flags().BoolVar(&noBrowser, "no-browser", false, "Print login URL instead of opening browser")
	cmd.Flags().DurationVar(&timeout, "timeout", 5*time.Minute, "Authentication timeout")

	cmd.AddCommand(newAuthFlyVerifyCommand(profileFlag))

	return cmd
}

func newAuthFlyVerifyCommand(profileFlag *string) *cobra.Command {
	var tokenPath string

	cmd := &cobra.Command{
		Use:   "verify",
		Short: "Verify cached Fly.io credentials",
		Long: `Verify that your cached Fly.io token is valid.

This command checks:
  ✓ Token exists and is properly formatted
  ✓ Token is valid with Fly.io API
  ✓ App is accessible (if configured)`,
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, _ := profiles.ResolveProfile(strings.TrimSpace(*profileFlag))
			path := profiles.FirstNonEmpty(tokenPath, profile.TokenPath, flyprefs.DefaultTokenPath())
			
			fmt.Fprintln(cmd.OutOrStdout(), "Verifying Fly.io credentials...")
			
			token, err := flyprefs.LoadToken(path)
			if err != nil {
				return fmt.Errorf("✗ load fly token: %w", err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "✓ Token loaded from %s\n", path)
			
			identity, _, err := auth.VerifyFlyToken(cmd.Context(), profile, token)
			if err != nil {
				return fmt.Errorf("✗ verify fly token: %w", err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "✓ Token valid for %s\n", strings.TrimSpace(identity))
			
			if err := printFlyLive(cmd.Context(), cmd.OutOrStdout(), profile, strings.TrimSpace(profile.FlyApp)); err != nil {
				appName := strings.TrimSpace(profile.FlyApp)
				if appName != "" {
					fmt.Fprintf(cmd.OutOrStdout(), "\n⚠  App '%s' check failed: %v\n", appName, err)
					fmt.Fprintln(cmd.OutOrStdout(), "   (This is OK if the app doesn't exist yet)")
				}
			} else {
				fmt.Fprintln(cmd.OutOrStdout(), "✓ All Fly.io checks passed!")
			}
			
			return nil
		},
	}

	cmd.Flags().StringVar(&tokenPath, "path", "", "Override token file path")
	return cmd
}

func printFlyLive(ctx context.Context, out io.Writer, profile sharedcfg.ToolingProfile, appName string) error {
	appName = strings.TrimSpace(appName)
	if appName == "" {
		fmt.Fprintln(out, "⚠  No Fly app configured; skipping live check")
		fmt.Fprintln(out, "   (Specify app with --app flag or in profile)")
		return nil
	}
	
	info, err := fly.DescribeFly(ctx, profile, appName)
	if err != nil {
		return err
	}
	
	fmt.Fprintf(out, "✓ Fly app '%s' is reachable\n", info.AppName)
	fmt.Fprintf(out, "  • Status: %s\n", info.Status)
	fmt.Fprintf(out, "  • Version: %d\n", info.Version)
	if info.Hostname != "" {
		fmt.Fprintf(out, "  • Hostname: %s\n", info.Hostname)
	}
	return nil
}
