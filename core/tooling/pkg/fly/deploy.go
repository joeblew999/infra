package fly

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/pelletier/go-toml/v2"
	client "github.com/superfly/fly-go"
)

const (
	defaultAPIBaseURL = "https://api.fly.io"
)

// DeployOptions captures the inputs needed to roll out a new release on Fly.
type DeployOptions struct {
	AccessToken     string
	AppName         string
	ImageRef        string
	ConfigPath      string
	PlatformVersion string
	Strategy        string
	Verbose         bool
}

// Deploy publishes a new release for the given app using the supplied image
// reference and configuration definition.
func Deploy(ctx context.Context, opts DeployOptions) (string, error) {
	if opts.AppName == "" {
		return "", errors.New("fly deploy: app name is required")
	}
	if opts.ImageRef == "" {
		return "", errors.New("fly deploy: image reference is required")
	}

	rawToken := strings.TrimSpace(opts.AccessToken)
	if rawToken == "" {
		return "", errors.New("fly deploy: access token is required")
	}

	definition, err := loadDefinition(opts.ConfigPath)
	if err != nil {
		return "", fmt.Errorf("load fly config: %w", err)
	}

	strategy := parseStrategy(opts.Strategy)
	platform := opts.PlatformVersion

	client.SetBaseURL(defaultAPIBaseURL)
	log := NewLogger("tooling-deploy", opts.Verbose)

	cl := client.NewClientFromOptions(client.ClientOptions{
		AccessToken: rawToken,
		Name:        "core-tooling",
		Version:     "dev",
		Logger:      log,
	})

	app, err := cl.GetApp(ctx, opts.AppName)
	if err != nil {
		return "", fmt.Errorf("lookup app %q: %w", opts.AppName, err)
	}

	if platform == "" {
		platform = app.PlatformVersion
	}
	if platform == "" {
		platform = "machines"
	}

	mutationID := fmt.Sprintf("core-tooling-%d", time.Now().UnixNano())
	input := client.CreateReleaseInput{
		AppId:            app.ID,
		Image:            opts.ImageRef,
		Definition:       definition,
		PlatformVersion:  platform,
		Strategy:         strategy,
		ClientMutationId: mutationID,
	}

	resp, err := cl.CreateRelease(ctx, input)
	if err != nil {
		return "", fmt.Errorf("create release: %w", err)
	}

	release := resp.CreateRelease.Release
	return fmt.Sprintf("release %s (version %d)", release.Id, release.Version), nil
}

func loadDefinition(path string) (map[string]any, error) {
	resolved := path
	if resolved == "" {
		resolved = "fly.toml"
	}
	data, err := os.ReadFile(resolved)
	if err != nil {
		return nil, err
	}
	var def map[string]any
	if err := toml.Unmarshal(data, &def); err != nil {
		return nil, err
	}
	return def, nil
}

func parseStrategy(value string) client.DeploymentStrategy {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "bluegreen":
		return client.DeploymentStrategyBluegreen
	case "immediate":
		return client.DeploymentStrategyImmediate
	case "rolling":
		return client.DeploymentStrategyRolling
	case "canary", "":
		fallthrough
	default:
		return client.DeploymentStrategyCanary
	}
}
