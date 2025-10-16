package cloudflare

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	sharedcfg "github.com/joeblew999/infra/core/pkg/shared/config"
	"github.com/joeblew999/infra/core/pkg/shared/secrets"
)

const (
	cloudflareSubdir = "cloudflare"
	defaultFileName  = "api_token"
)

// TokenKind identifies the origin of a cached Cloudflare token.
type TokenKind string

const (
	TokenKindManual    TokenKind = "manual"
	TokenKindBootstrap TokenKind = "bootstrap"
)

// Settings stores persisted Cloudflare preferences for DNS and R2.
type Settings struct {
	ZoneName  string    `json:"zone_name"`
	ZoneID    string    `json:"zone_id"`
	AccountID string    `json:"account_id"`
	R2Bucket  string    `json:"r2_bucket"`
	R2Region  string    `json:"r2_region"`
	AppDomain string    `json:"app_domain"`
	UpdatedAt time.Time `json:"updated_at"`
}

const (
	tokenKey    = "cloudflare.token"
	settingsKey = "cloudflare.settings"
	userID      = "local" // CLI mode uses "local" userID
)

// DefaultTokenPath returns the path used to cache Cloudflare API tokens.
// Deprecated: Kept for compatibility, but storage now uses secrets backend.
func DefaultTokenPath() string {
	tooling := sharedcfg.Tooling()
	if path := strings.TrimSpace(tooling.Active.CloudflareTokenPath); path != "" {
		return path
	}
	return filepath.Join(sharedcfg.GetDataPath(), "core", "secrets", cloudflareSubdir, defaultFileName)
}

// SaveToken persists the provided token using the secrets backend.
func SaveToken(path, token string) error {
	ctx := context.Background()
	backend, err := secrets.NewBackend(ctx, "")
	if err != nil {
		return fmt.Errorf("create secrets backend: %w", err)
	}

	token = strings.TrimSpace(token)
	if token == "" {
		return errors.New("token cannot be empty")
	}

	// Save as active token
	if err := backend.Set(ctx, userID, tokenKey, []byte(token)); err != nil {
		return err
	}

	// Also save path pointer for backward compatibility (deprecated pattern)
	if path != "" {
		activeKey := "cloudflare.active_token_path"
		return backend.Set(ctx, userID, activeKey, []byte(path))
	}
	return nil
}

// LoadToken reads the cached token using the secrets backend.
func LoadToken(path string) (string, error) {
	ctx := context.Background()
	backend, err := secrets.NewBackend(ctx, "")
	if err != nil {
		return "", fmt.Errorf("create secrets backend: %w", err)
	}

	// If specific path provided, try to resolve it (backward compatibility)
	// For now, just use the active token
	data, err := backend.Get(ctx, userID, tokenKey)
	if err != nil {
		return "", err
	}

	token := strings.TrimSpace(string(data))
	if token == "" {
		return "", errors.New("cached cloudflare token is empty or not found")
	}
	return token, nil
}

// SaveTokenForKind stores the token under a kind-specific key.
// Returns the key used (for backward compatibility).
func SaveTokenForKind(kind TokenKind, token string) (string, error) {
	ctx := context.Background()
	backend, err := secrets.NewBackend(ctx, "")
	if err != nil {
		return "", fmt.Errorf("create secrets backend: %w", err)
	}

	token = strings.TrimSpace(token)
	if token == "" {
		return "", errors.New("token cannot be empty")
	}

	key := fmt.Sprintf("cloudflare.token.%s", string(kind))
	if err := backend.Set(ctx, userID, key, []byte(token)); err != nil {
		return "", err
	}
	return key, nil
}

// LoadSettings returns stored Cloudflare preferences using the secrets backend.
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

// SaveSettings writes Cloudflare preferences using the secrets backend.
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
	cfg.ZoneName = strings.TrimSpace(cfg.ZoneName)
	cfg.ZoneID = strings.TrimSpace(cfg.ZoneID)
	cfg.AccountID = strings.TrimSpace(cfg.AccountID)
	cfg.R2Bucket = strings.TrimSpace(cfg.R2Bucket)
	cfg.R2Region = strings.TrimSpace(cfg.R2Region)
	cfg.AppDomain = strings.TrimSpace(cfg.AppDomain)
	return cfg
}
