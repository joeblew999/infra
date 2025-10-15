package cloudflare

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	cf "github.com/cloudflare/cloudflare-go"
	"github.com/skratchdot/open-golang/open"

	sharedcfg "github.com/joeblew999/infra/core/pkg/shared/config"
	sharedprompt "github.com/joeblew999/infra/core/pkg/shared/prompt"
	types "github.com/joeblew999/infra/core/tooling/pkg/types"
)

const cloudflareTokenURL = "https://dash.cloudflare.com/profile/api-tokens"

// EnsureCloudflareToken checks for existing valid credentials or triggers authentication.
func EnsureCloudflareToken(ctx context.Context, profile sharedcfg.ToolingProfile, in io.Reader, out io.Writer, noBrowser bool, prompter types.Prompter) error {
	path := firstNonEmpty(profile.CloudflareTokenPath, DefaultTokenPath())
	token, err := LoadToken(path)
	if err == nil {
		body, api, err := verifyCloudflareToken(ctx, token)
		if err == nil {
			fmt.Fprintf(out, "Cloudflare token already valid (status: %s)\n", strings.TrimSpace(body.Status))
			return ConfigureCloudflarePreferences(ctx, api, in, out, profile)
		}
		fmt.Fprintln(out, "Cached Cloudflare token is invalid, requesting new token...")
	}
	return RunCloudflareAuth(ctx, profile, "", path, noBrowser, in, out, prompter)
}

// RunCloudflareAuth performs manual API token authentication.
func RunCloudflareAuth(ctx context.Context, profile sharedcfg.ToolingProfile, tokenInput, tokenPath string, noBrowser bool, in io.Reader, out io.Writer, prompter types.Prompter) error {
	token := strings.TrimSpace(tokenInput)
	reader := bufio.NewReader(in)
	
	if token == "" {
		fmt.Fprintln(out, "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Fprintln(out, "  CLOUDFLARE API TOKEN REQUIRED")
		fmt.Fprintln(out, "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Fprintln(out)
		fmt.Fprintln(out, "Required permissions:")
		fmt.Fprintln(out, "  • Zone:DNS:Edit")
		fmt.Fprintln(out, "  • Zone:Zone:Read")
		fmt.Fprintln(out, "  • Account:R2:Edit (if using R2)")
		fmt.Fprintln(out)
		
		if !noBrowser {
			fmt.Fprint(out, "Press ENTER to open browser (or Ctrl+C to cancel): ")
			if _, err := reader.ReadString('\n'); err != nil {
				return fmt.Errorf("cancelled: %w", err)
			}
			
			if err := open.Run(cloudflareTokenURL); err != nil {
				fmt.Fprintf(out, "⚠  Could not open browser: %v\n", err)
				fmt.Fprintf(out, "   Visit: %s\n", cloudflareTokenURL)
			} else {
				fmt.Fprintln(out, "✓ Browser opened")
			}
		} else {
			fmt.Fprintf(out, "Visit: %s\n", cloudflareTokenURL)
		}
		
		fmt.Fprintln(out)
		fmt.Fprint(out, "Paste your Cloudflare API token: ")
		line, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		token = strings.TrimSpace(line)
	}
	
	if token == "" {
		return errors.New("token cannot be empty")
	}

	if tokenPath == "" {
		tokenPath = firstNonEmpty(profile.CloudflareTokenPath, DefaultTokenPath())
	}

	fmt.Fprintln(out)
	fmt.Fprintln(out, "Verifying token...")
	
	body, api, err := verifyCloudflareToken(ctx, token)
	if err != nil {
		return fmt.Errorf("verify token: %w", err)
	}
	
	fmt.Fprintf(out, "✓ Token verified (status: %s)\n", strings.TrimSpace(body.Status))
	
	fmt.Fprintln(out)
	if err := verifyTokenPermissions(ctx, api, out); err != nil {
		fmt.Fprintf(out, "\n⚠  Warning: %v\n", err)
		fmt.Fprint(out, "Continue anyway? (y/N): ")
		response, _ := reader.ReadString('\n')
		if !strings.HasPrefix(strings.ToLower(strings.TrimSpace(response)), "y") {
			return fmt.Errorf("cancelled due to insufficient permissions")
		}
	}
	
	fmt.Fprintln(out)
	if err := ConfigureCloudflarePreferences(ctx, api, in, out, profile); err != nil {
		return err
	}

	if err := SaveToken(tokenPath, token); err != nil {
		return fmt.Errorf("save token: %w", err)
	}
	
	_, _ = SaveTokenForKind(TokenKindManual, token)

	fmt.Fprintln(out)
	fmt.Fprintf(out, "✓ Token saved to %s\n", tokenPath)
	return nil
}

// RunCloudflareBootstrap performs bootstrap authentication using Global API Key.
func RunCloudflareBootstrap(ctx context.Context, profile sharedcfg.ToolingProfile, opts BootstrapOptions, in io.Reader, out io.Writer, prompter types.Prompter) error {
	reader := bufio.NewReader(in)

	email := strings.TrimSpace(opts.Email)
	if email == "" {
		value, err := sharedprompt.String(reader, out, "Cloudflare email:")
		if err != nil {
			return err
		}
		email = strings.TrimSpace(value)
	}
	if email == "" {
		return errors.New("email is required")
	}

	apiKey := strings.TrimSpace(opts.GlobalAPIKey)
	if apiKey == "" {
		if prompter != nil {
			secret, err := prompter.PromptSecret(ctx, types.PromptMessage{
				Provider: "cloudflare",
				Kind:     types.PromptKindToken,
				Message:  "Paste Cloudflare global API key: ",
			})
			if err != nil {
				return err
			}
			apiKey = strings.TrimSpace(secret)
		} else {
			fmt.Fprint(out, "Cloudflare global API key: ")
			value, err := reader.ReadString('\n')
			if err != nil {
				return err
			}
			apiKey = strings.TrimSpace(value)
		}
	}
	if apiKey == "" {
		return errors.New("global API key is required")
	}

	client, err := cf.New(apiKey, email)
	if err != nil {
		return fmt.Errorf("create API client: %w", err)
	}

	if err := ConfigureCloudflarePreferences(ctx, client, in, out, profile); err != nil {
		return err
	}

	settings, err := LoadSettings()
	if err != nil {
		return fmt.Errorf("load settings: %w", err)
	}
	if strings.TrimSpace(settings.ZoneID) == "" || strings.TrimSpace(settings.AccountID) == "" {
		return errors.New("zone configuration incomplete")
	}

	permissionGroups, err := client.ListAPITokensPermissionGroups(ctx)
	if err != nil {
		return fmt.Errorf("list permission groups: %w", err)
	}

	selected, err := SelectPermissionGroups(permissionGroups, opts.IncludeR2)
	if err != nil {
		return err
	}

	tokenName := strings.TrimSpace(opts.TokenName)
	if tokenName == "" {
		tokenName = fmt.Sprintf("core-tooling-%s", time.Now().UTC().Format("20060102-150405"))
	}

	created, err := CreateScopedToken(ctx, client, CreateScopedTokenParams{
		Name:          tokenName,
		AccountID:     settings.AccountID,
		ZoneID:        settings.ZoneID,
		ZoneName:      settings.ZoneName,
		IncludeR2:     opts.IncludeR2,
		PermissionSet: selected,
	})
	if err != nil {
		return err
	}

	verifyBody, _, err := VerifyCloudflareToken(ctx, created.Value)
	if err != nil {
		return fmt.Errorf("verify generated token: %w", err)
	}

	tokenPath := firstNonEmpty(opts.TokenPath, profile.CloudflareTokenPath, DefaultTokenPath())
	if err := SaveToken(tokenPath, created.Value); err != nil {
		return fmt.Errorf("save token: %w", err)
	}
	
	if _, err := SaveTokenForKind(TokenKindBootstrap, created.Value); err != nil {
		fmt.Fprintf(out, "warning: unable to persist bootstrap token copy: %v\n", err)
	}

	fmt.Fprintf(out, "✓ Token %q created (status=%s) and saved to %s\n", created.Name, verifyBody.Status, tokenPath)
	return nil
}

// BootstrapOptions configures bootstrap authentication.
type BootstrapOptions struct {
	Email        string
	GlobalAPIKey string
	TokenName    string
	IncludeR2    bool
	TokenPath    string
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if trimmed := strings.TrimSpace(v); trimmed != "" {
			return trimmed
		}
	}
	return ""
}
