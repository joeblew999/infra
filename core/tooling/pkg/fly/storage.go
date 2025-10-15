package fly

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	sharedcfg "github.com/joeblew999/infra/core/pkg/shared/config"
	"github.com/superfly/fly-go/tokens"
)

// Settings captures persisted Fly preferences like organisation and region.
type Settings struct {
	OrgSlug    string    `json:"org_slug"`
	RegionCode string    `json:"region_code"`
	RegionName string    `json:"region_name"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// DefaultTokenPath returns the path used to cache Fly API tokens.
func DefaultTokenPath() string {
	tooling := sharedcfg.Tooling()
	if path := strings.TrimSpace(tooling.Active.TokenPath); path != "" {
		return path
	}
	return filepath.Join(sharedcfg.GetDataPath(), "core", "secrets", "fly", "access_token")
}

// SaveToken persists the provided token to disk with 0600 permissions.
func SaveToken(path, token string) error {
	if path == "" {
		path = DefaultTokenPath()
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(token), 0o600)
}

// LoadToken reads the token from disk. It returns an error if the file is
// missing or empty.
func LoadToken(path string) (string, error) {
	if path == "" {
		path = DefaultTokenPath()
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	token := strings.TrimSpace(string(data))
	if token == "" {
		return "", errors.New("cached fly token is empty")
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
	json := fmt.Sprintf(`{"auths":{"registry.fly.io":{"auth":"%s"}}}`, auth)
	if err := os.WriteFile(configPath, []byte(json), 0o600); err != nil {
		_ = os.RemoveAll(dir)
		return "", err
	}
	return dir, nil
}

func settingsPath() string {
	root := filepath.Join(sharedcfg.GetDataPath(), "core", "fly")
	return filepath.Join(root, "settings.json")
}

// LoadSettings returns stored Fly preferences. Missing files return zero-value settings.
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

// SaveSettings writes the provided Fly preferences to disk.
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
	cfg.OrgSlug = strings.TrimSpace(cfg.OrgSlug)
	cfg.RegionCode = strings.TrimSpace(cfg.RegionCode)
	cfg.RegionName = strings.TrimSpace(cfg.RegionName)
	return cfg
}
