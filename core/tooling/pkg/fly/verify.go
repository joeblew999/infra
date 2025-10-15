package fly

import (
	"context"
	"fmt"
	"strings"

	flyapi "github.com/superfly/fly-go"

	sharedbuild "github.com/joeblew999/infra/core/pkg/shared/build"
	sharedcfg "github.com/joeblew999/infra/core/pkg/shared/config"
)

// VerifyFlyToken verifies a Fly token and returns user identity and client.
func VerifyFlyToken(ctx context.Context, profile sharedcfg.ToolingProfile, token string) (string, *flyapi.Client, error) {
	flyapi.SetBaseURL(sharedcfg.Tooling().Active.FlyAPIBase)

	info := sharedbuild.Get()
	version := firstNonEmpty(strings.TrimSpace(info.Version), strings.TrimSpace(info.Revision), "dev")

	client := flyapi.NewClientFromOptions(flyapi.ClientOptions{
		AccessToken: token,
		Name:        "core-tool",
		Version:     version,
		Logger:      NewLogger("core-tool", false),
	})

	user, err := client.GetCurrentUser(ctx)
	if err != nil {
		return "", nil, err
	}

	identity := strings.TrimSpace(user.Email)
	if identity == "" {
		identity = strings.TrimSpace(user.Name)
	}
	if identity == "" {
		identity = "Fly user"
	} else {
		identity = fmt.Sprintf("Fly user %s", identity)
	}

	return identity, client, nil
}
