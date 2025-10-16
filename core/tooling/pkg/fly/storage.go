package fly

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	sharedcfg "github.com/joeblew999/infra/core/pkg/shared/config"
	"github.com/joeblew999/infra/core/pkg/shared/secrets"
	"github.com/superfly/fly-go/tokens"
)

// Settings captures persisted Fly preferences like organisation and region.
type Settings struct {
	OrgSlug    string    `json:"org_slug"`
	RegionCode string    `json:"region_code"`
	RegionName string    `json:"region_name"`
	UpdatedAt  time.Time `json:"updated_at"`
}

const (
	tokenKey    = "fly.token"
	settingsKey = "fly.settings"
	userID      = "local" // CLI mode uses "local" userID
)

// DefaultTokenPath returns the path used to cache Fly API tokens.
// Deprecated: Kept for compatibility, but storage now uses secrets backend.
func DefaultTokenPath() string {
	tooling := sharedcfg.Tooling()
	if path := strings.TrimSpace(tooling.Active.TokenPath); path != "" {
		return path
	}
	return filepath.Join(sharedcfg.GetDataPath(), "core", "secrets", "fly", "access_token")
}

// SaveToken persists the provided token using the secrets backend.
func SaveToken(path, token string) error {
	ctx := context.Background()
	backend, err := secrets.NewBackend(ctx, "") // Empty string uses env var or defaults to filesystem
	if err != nil {
		return fmt.Errorf("create secrets backend: %w", err)
	}

	token = strings.TrimSpace(token)
	if token == "" {
		return errors.New("token cannot be empty")
	}

	return backend.Set(ctx, userID, tokenKey, []byte(token))
}

// LoadToken reads the token using the secrets backend.
func LoadToken(path string) (string, error) {
	ctx := context.Background()
	backend, err := secrets.NewBackend(ctx, "") // Empty string uses env var or defaults to filesystem
	if err != nil {
		return "", fmt.Errorf("create secrets backend: %w", err)
	}

	data, err := backend.Get(ctx, userID, tokenKey)
	if err != nil {
		return "", err
	}

	token := strings.TrimSpace(string(data))
	if token == "" {
		return "", errors.New("cached fly token is empty or not found")
	}
	return token, nil
}

// CreateDockerConfig writes a minimal Docker config.json containing credentials
// for Fly's registry and returns the directory that should be used as
// DOCKER_CONFIG.
func CreateDockerConfig(token string) (string, error) {
	parsed := tokens.Parse(token)
	dockerToken := parsed.Docker()
	auth := base64.StdEncoding.EncodeToString([]byte("x:" + dockerToken))

	dir, err := os.MkdirTemp("", "core-fly-docker-")
	if err != nil {
		return "", err
	}
	configPath := filepath.Join(dir, "config.json")
	jsonStr := fmt.Sprintf(`{"auths":{"registry.fly.io":{"auth":"%s"}}}`, auth)
	if err := os.WriteFile(configPath, []byte(jsonStr), 0o600); err != nil {
		_ = os.RemoveAll(dir)
		return "", err
	}
	return dir, nil
}

// LoadSettings returns stored Fly preferences using the secrets backend.
func LoadSettings() (Settings, error) {
	ctx := context.Background()
	backend, err := secrets.NewBackend(ctx, "")
	if err != nil {
		return Settings{}, fmt.Errorf("create secrets backend: %w", err)
	}

	data, err := backend.Get(ctx, userID, settingsKey)
	if err != nil {
		return Settings{}, err
	}

	if len(data) == 0 {
		return Settings{}, nil // No settings saved yet
	}

	var cfg Settings
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Settings{}, fmt.Errorf("parse settings: %w", err)
	}
	return normaliseSettings(cfg), nil
}

// SaveSettings writes the provided Fly preferences using the secrets backend.
func SaveSettings(cfg Settings) error {
	ctx := context.Background()
	backend, err := secrets.NewBackend(ctx, "")
	if err != nil {
		return fmt.Errorf("create secrets backend: %w", err)
	}

	cfg = normaliseSettings(cfg)
	if cfg.UpdatedAt.IsZero() {
		cfg.UpdatedAt = time.Now().UTC()
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal settings: %w", err)
	}

	return backend.Set(ctx, userID, settingsKey, data)
}

func normaliseSettings(cfg Settings) Settings {
	cfg.OrgSlug = strings.TrimSpace(cfg.OrgSlug)
	cfg.RegionCode = strings.TrimSpace(cfg.RegionCode)
	cfg.RegionName = strings.TrimSpace(cfg.RegionName)
	return cfg
}
