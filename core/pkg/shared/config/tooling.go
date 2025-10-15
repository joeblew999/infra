package config

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	// EnvVarToolingProfile selects the active tooling profile.
	EnvVarToolingProfile = "CORE_TOOLING_PROFILE"
	// EnvVarToolingKORepository overrides the KO repository used for image publishes.
	EnvVarToolingKORepository = "CORE_KO_REPOSITORY"
	// EnvVarToolingFlyApp overrides the Fly.io application name for deploy operations.
	EnvVarToolingFlyApp = "CORE_FLY_APP"
	// EnvVarToolingFlyOrg overrides the Fly.io organisation slug.
	EnvVarToolingFlyOrg = "CORE_FLY_ORG"
	// EnvVarToolingTagTemplate overrides the tag template applied to KO publishes.
	EnvVarToolingTagTemplate = "CORE_TAG_TEMPLATE"
	// EnvVarToolingSupportsDocker toggles docker support for the active profile.
	EnvVarToolingSupportsDocker = "CORE_TOOLING_SUPPORTS_DOCKER"
	// EnvVarFlyTokenPath overrides the cached Fly token path.
	EnvVarFlyTokenPath = "CORE_FLY_TOKEN_PATH"
	// EnvVarCloudflareTokenPath overrides the cached Cloudflare token path.
	EnvVarCloudflareTokenPath = "CORE_CLOUDFLARE_TOKEN_PATH"
	// EnvVarFlyAPIBase overrides the Fly API base URL.
	EnvVarFlyAPIBase = "CORE_FLY_API_BASE"
)

const (
	defaultToolingProfileName      = "local"
	defaultTokenFileName           = "access_token"
	defaultCloudflareTokenFileName = "api_token"
	defaultFlyAPIBase              = "https://api.fly.io"
)

// ToolingMode indicates which release workflow a profile should execute.
type ToolingMode string

const (
	ToolingModeLocal ToolingMode = "local"
	ToolingModeFly   ToolingMode = "fly"
)

// ToolingProfile describes a single tooling configuration profile.
type ToolingProfile struct {
	Name                string
	Mode                ToolingMode
	KORepository        string
	FlyApp              string
	FlyOrg              string
	TagTemplate         string
	DockerArgs          []string
	SupportsDocker      bool
	TokenPath           string
	CloudflareTokenPath string
	KoConfig            string
	FlyConfig           string
	ImportPath          string
	KoTemplate          string
	FlyTemplate         string
	FlyRegion           string
	FlyAPIBase          string
}

// ToolingSettings captures the resolved tooling configuration including all
// profiles and the active selection.
type ToolingSettings struct {
	Active   ToolingProfile
	Profiles map[string]ToolingProfile
}

// Tooling returns the resolved tooling configuration using the shared
// environment-aware defaults.
func Tooling() ToolingSettings {
	profiles := defaultToolingProfiles()
	activeName := valueOrDefault(strings.TrimSpace(os.Getenv(EnvVarToolingProfile)), defaultToolingProfileName)
	active, ok := profiles[activeName]
	if !ok {
		authoritative := profiles[defaultToolingProfileName]
		active = authoritative
		activeName = defaultToolingProfileName
	}

	applyProfileOverrides(&active)
	profiles[activeName] = active

	return ToolingSettings{Active: active, Profiles: profiles}
}

// ProfileNames returns the known profile identifiers in stable order.
func (t ToolingSettings) ProfileNames() []string {
	names := make([]string, 0, len(t.Profiles))
	for name := range t.Profiles {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func defaultToolingProfiles() map[string]ToolingProfile {
	secretsRoot := filepath.Join(GetDataPath(), "core", "secrets")
	baseToken := filepath.Join(secretsRoot, "fly", defaultTokenFileName)
	baseCloudflare := filepath.Join(secretsRoot, "cloudflare", defaultCloudflareTokenFileName)

	local := baseProfile("local", ToolingModeLocal, baseToken, baseCloudflare)
	local.KORepository = "ko.local/core"
	local.TagTemplate = "git-{short}"
	local.DockerArgs = []string{"-p8080:8080"}
	local.SupportsDocker = true

	flyProfile := baseProfile("fly", ToolingModeFly, baseToken, baseCloudflare)
	flyProfile.KORepository = "registry.fly.io/infra-mgmt"
	flyProfile.FlyApp = "infra-mgmt"
	flyProfile.TagTemplate = "git-{short}"
	flyProfile.SupportsDocker = false

	return map[string]ToolingProfile{
		"local": local,
		"fly":   flyProfile,
	}
}

func baseProfile(name string, mode ToolingMode, tokenPath, cloudflarePath string) ToolingProfile {
	return ToolingProfile{
		Name:                name,
		Mode:                mode,
		KORepository:        "ko.local/core",
		TagTemplate:         "git-{short}",
		DockerArgs:          nil,
		SupportsDocker:      false,
		TokenPath:           tokenPath,
		CloudflareTokenPath: cloudflarePath,
		KoConfig:            ".ko.yaml",
		KoTemplate:          "config/templates/ko.yaml.tmpl",
		FlyConfig:           "fly.toml",
		FlyTemplate:         "config/templates/fly.toml.tmpl",
		FlyRegion:           "syd",
		FlyAPIBase:          defaultFlyAPIBase,
		ImportPath:          "./cmd/core",
	}
}

func applyProfileOverrides(profile *ToolingProfile) {
	overrideString(&profile.KORepository, EnvVarToolingKORepository)
	overrideString(&profile.FlyApp, EnvVarToolingFlyApp)
	overrideString(&profile.FlyOrg, EnvVarToolingFlyOrg)
	overrideString(&profile.TagTemplate, EnvVarToolingTagTemplate)
	overrideBool(&profile.SupportsDocker, EnvVarToolingSupportsDocker)
	overrideString(&profile.TokenPath, EnvVarFlyTokenPath)
	overrideString(&profile.CloudflareTokenPath, EnvVarCloudflareTokenPath)
	overrideString(&profile.FlyAPIBase, EnvVarFlyAPIBase)
}

func overrideString(target *string, env string) {
	if value := strings.TrimSpace(os.Getenv(env)); value != "" {
		*target = value
	}
}

func overrideBool(target *bool, env string) {
	if value := strings.TrimSpace(os.Getenv(env)); value != "" {
		switch strings.ToLower(value) {
		case "1", "true", "yes", "on":
			*target = true
		case "0", "false", "no", "off":
			*target = false
		}
	}
}

func valueOrDefault(value, fallback string) string {
	if value != "" {
		return value
	}
	return fallback
}
