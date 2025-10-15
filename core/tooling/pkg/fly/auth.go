package fly

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/skratchdot/open-golang/open"
	flyapi "github.com/superfly/fly-go"

	sharedcfg "github.com/joeblew999/infra/core/pkg/shared/config"
	types "github.com/joeblew999/infra/core/tooling/pkg/types"
)

// EnsureFlyToken ensures a valid Fly token exists, authenticating if needed.
func EnsureFlyToken(ctx context.Context, profile sharedcfg.ToolingProfile, in io.Reader, out io.Writer, noBrowser bool, prompter types.Prompter) error {
	tokenPath := firstNonEmpty(profile.TokenPath, DefaultTokenPath())
	token, err := LoadToken(tokenPath)
	if err == nil {
		if identity, client, err := VerifyFlyToken(ctx, profile, token); err == nil {
			if err := ConfigureFlyPreferences(ctx, client, out, profile.FlyOrg, profile.FlyRegion); err != nil {
				return err
			}
			fmt.Fprintf(out, "Fly token already valid for %s\n", identity)
			return nil
		}
	}
	return RunFlyAuth(ctx, profile, "", tokenPath, noBrowser, 5*time.Minute, in, out, prompter)
}

// RunFlyAuth performs Fly authentication and saves the token.
func RunFlyAuth(ctx context.Context, profile sharedcfg.ToolingProfile, tokenInput, tokenPath string, noBrowser bool, timeout time.Duration, in io.Reader, out io.Writer, prompter types.Prompter) error {
	token := strings.TrimSpace(tokenInput)
	if token == "" {
		var err error
		token, err = InteractiveFlyAuth(ctx, profile, noBrowser, timeout, in, out, prompter)
		if err != nil {
			return err
		}
	}
	if token == "" {
		return errors.New("token cannot be empty")
	}

	identity, client, err := VerifyFlyToken(ctx, profile, token)
	if err != nil {
		return fmt.Errorf("verify fly token: %w", err)
	}

	if err := ConfigureFlyPreferences(ctx, client, out, profile.FlyOrg, profile.FlyRegion); err != nil {
		return err
	}

	if tokenPath == "" {
		tokenPath = DefaultTokenPath()
	}
	if err := SaveToken(tokenPath, token); err != nil {
		return fmt.Errorf("save fly token: %w", err)
	}

	msg := fmt.Sprintf("Fly token saved to %s (%s)", tokenPath, identity)
	if prompter != nil {
		_ = prompter.Notify(ctx, types.PromptMessage{Provider: "fly", Kind: types.PromptKindInfo, Message: msg})
	} else {
		fmt.Fprintln(out, msg)
	}
	return nil
}

// InteractiveFlyAuth performs interactive browser-based Fly authentication.
func InteractiveFlyAuth(ctx context.Context, profile sharedcfg.ToolingProfile, noBrowser bool, timeout time.Duration, in io.Reader, out io.Writer, prompter types.Prompter) (string, error) {
	host, err := os.Hostname()
	if err != nil || host == "" {
		host = "core-tool"
	}

	flyapi.SetBaseURL(sharedcfg.Tooling().Active.FlyAPIBase)

	session, err := flyapi.StartCLISessionWebAuth(host, false)
	if err != nil {
		return "", fmt.Errorf("start fly web auth: %w", err)
	}

	message := fmt.Sprintf("To authenticate with Fly.io, visit:\n  %s", session.URL)
	if prompter != nil {
		_ = prompter.Notify(ctx, types.PromptMessage{
			Provider: "fly",
			Kind:     types.PromptKindLink,
			Message:  message,
			URL:      session.URL,
		})
	} else {
		fmt.Fprintln(out, message)
		if !noBrowser {
			_ = open.Run(session.URL)
		}
	}

	authCtx := ctx
	if timeout > 0 {
		var cancel context.CancelFunc
		authCtx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-authCtx.Done():
			return "", fmt.Errorf("login timeout: %w", authCtx.Err())
		case <-ticker.C:
			token, err := flyapi.GetAccessTokenForCLISession(ctx, session.ID)
			if err != nil {
				continue
			}
			token = strings.TrimSpace(token)
			if token != "" {
				return token, nil
			}
		}
	}
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if trimmed := strings.TrimSpace(v); trimmed != "" {
			return trimmed
		}
	}
	return ""
}
