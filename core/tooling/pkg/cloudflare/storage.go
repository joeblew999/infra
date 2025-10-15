package cloudflare

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	sharedcfg "github.com/joeblew999/infra/core/pkg/shared/config"
)

const (
	cloudflareSubdir = "cloudflare"
	defaultFileName  = "api_token"
	tokenDirName     = "tokens"
	activeFileName   = "active_token"
	settingsFileName = "settings.json"
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

// DefaultTokenPath returns the path used to cache Cloudflare API tokens.
func DefaultTokenPath() string {
	tooling := sharedcfg.Tooling()
	if path := strings.TrimSpace(tooling.Active.CloudflareTokenPath); path != "" {
		return path
	}
	return filepath.Join(sharedcfg.GetDataPath(), "core", "secrets", cloudflareSubdir, defaultFileName)
}

func tokenStoreDir() string {
	return filepath.Join(sharedcfg.GetDataPath(), "core", cloudflareSubdir)
}

func tokenPathForKind(kind TokenKind) string {
	return filepath.Join(tokenStoreDir(), tokenDirName, string(kind))
}

func activeTokenPointer() string {
	return filepath.Join(tokenStoreDir(), activeFileName)
}

func writeTokenFile(path, token string) error {
	if path == "" {
		return fmt.Errorf("cloudflare token path not provided")
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(strings.TrimSpace(token)), 0o600)
}

// SaveToken persists the provided token to disk with 0600 permissions.
func SaveToken(path, token string) error {
	if path == "" {
		path = DefaultTokenPath()
	}
	if err := writeTokenFile(path, token); err != nil {
		return err
	}
	return setActiveTokenPath(path)
}

// LoadToken reads the cached token.
func LoadToken(path string) (string, error) {
	path = resolveTokenPath(path)
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	token := strings.TrimSpace(string(data))
	if token == "" {
		return "", errors.New("cached cloudflare token is empty")
	}
	return token, nil
}

// SaveTokenForKind stores the token under a kind-specific file without
// mutating the active token pointer. It returns the path used.
func SaveTokenForKind(kind TokenKind, token string) (string, error) {
	path := tokenPathForKind(kind)
	if err := writeTokenFile(path, token); err != nil {
		return "", err
	}
	return path, nil
}

func setActiveTokenPath(path string) error {
	if path == "" {
		return fmt.Errorf("cloudflare active token path is empty")
	}
	if err := os.MkdirAll(tokenStoreDir(), 0o700); err != nil {
		return err
	}
	return os.WriteFile(activeTokenPointer(), []byte(strings.TrimSpace(path)), 0o600)
}

func resolveTokenPath(path string) string {
	if strings.TrimSpace(path) != "" {
		return strings.TrimSpace(path)
	}
	if data, err := os.ReadFile(activeTokenPointer()); err == nil {
		if resolved := strings.TrimSpace(string(data)); resolved != "" {
			return resolved
		}
	}
	return DefaultTokenPath()
}

func settingsPath() string {
	root := filepath.Join(sharedcfg.GetDataPath(), "core", "cloudflare")
	return filepath.Join(root, settingsFileName)
}

// LoadSettings returns stored Cloudflare preferences. Missing files return zero values.
func LoadSettings() (Settings, error) {
	path := settingsPath()
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Settings{}, nil
		}
		return Settings{}, err
	}
	var cfg Settings
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Settings{}, err
	}
	return normaliseSettings(cfg), nil
}

// SaveSettings writes Cloudflare preferences to disk.
func SaveSettings(cfg Settings) error {
	path := settingsPath()
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	cfg = normaliseSettings(cfg)
	if cfg.UpdatedAt.IsZero() {
		cfg.UpdatedAt = time.Now().UTC()
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
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
