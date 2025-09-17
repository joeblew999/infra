package web

import (
	"bytes"
	_ "embed"
	"fmt"
	"html/template"
	"runtime"
	"sync"
	"time"

	"github.com/joeblew999/infra/pkg/config"
)

// ConfigTemplateData contains all values required to render the live config dashboard.
type ConfigTemplateData struct {
	LastUpdatedDisplay  string
	LastUpdatedISO      string
	EnvironmentIcon     string
	EnvironmentTitle    string
	EnvironmentDesc     string
	EnvironmentBorder   string
	EnvironmentGradient string
	Config              ConfigInfo
	EnvVars             []EnvVarInfo
	Binaries            []BinaryInfo
	AllEnvVarsSet       bool
	EnvStatusText       string
}

// ConfigInfo captures core configuration details
type ConfigInfo struct {
	IsProduction bool
	Platform     string
	HTTPS        bool
	Version      string
	GitHash      string
	BuildTime    string
	GoVersion    string
	Paths        []PathInfo
}

// PathInfo represents a system path
type PathInfo struct {
	Name string
	Path string
}

// EnvVarInfo represents an environment variable status
type EnvVarInfo struct {
	Name        string
	Status      string
	Description string
	Icon        string
	Border      string
	Pill        string
}

// BinaryInfo represents a binary configuration
type BinaryInfo struct {
	Name string
	Path string
}

var (
	configCardsTplOnce sync.Once
	configCardsTpl     *template.Template
	configCardsTplErr  error
)

//go:embed templates/config-cards.html
var configCardsTemplate string

// RenderConfigCards renders the live config dashboard partial for DataStar SSE updates.
func RenderConfigCards(data ConfigTemplateData) (string, error) {
	configCardsTplOnce.Do(func() {
		configCardsTpl, configCardsTplErr = template.New("config-cards").Parse(configCardsTemplate)
	})

	if configCardsTplErr != nil {
		return "", fmt.Errorf("parse config cards template: %w", configCardsTplErr)
	}

	var buf bytes.Buffer
	if err := configCardsTpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("execute config cards template: %w", err)
	}

	return buf.String(), nil
}

// mapConfigToTemplate converts config data to template format
func mapConfigToTemplate() ConfigTemplateData {
	cfg := config.GetConfig()
	envStatus := config.GetEnvStatus()
	missingVars := config.GetMissingEnvVars()

	// Environment status
	var envIcon, envTitle, envDesc, envBorder, envGradient string
	if len(missingVars) > 0 {
		envIcon = "⚠️"
		envTitle = "Configuration Issues"
		envDesc = fmt.Sprintf("%d environment variables missing", len(missingVars))
		envBorder = "border-amber-200 dark:border-amber-600"
		envGradient = "from-amber-100 to-amber-50 dark:from-amber-900/20 dark:to-amber-800/20"
	} else {
		envIcon = "✅"
		envTitle = "Configuration Ready"
		envDesc = "All environment variables are properly configured"
		envBorder = "border-emerald-200 dark:border-emerald-600"
		envGradient = "from-emerald-100 to-emerald-50 dark:from-emerald-900/20 dark:to-emerald-800/20"
	}

	// Config info
	configInfo := ConfigInfo{
		IsProduction: config.IsProduction(),
		Platform:     fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
		HTTPS:        cfg.Environment.ShouldUseHTTPS,
		Version:      config.GetVersion(),
		GitHash:      config.GetShortHash(),
		BuildTime:    config.BuildTime,
		GoVersion:    runtime.Version(),
		Paths: []PathInfo{
			{Name: "Data", Path: cfg.Paths.Data},
			{Name: "Dependencies", Path: cfg.Paths.Dep},
			{Name: "Binaries", Path: cfg.Paths.Bin},
			{Name: "Docs", Path: cfg.Paths.Docs},
		},
	}

	// Environment variables
	envVars := make([]EnvVarInfo, 0, len(envStatus))
	for envVar, status := range envStatus {
		var icon, border, pill string
		var statusText string

		if status == "set" {
			icon = "✅"
			statusText = "SET"
			border = "border border-emerald-200 dark:border-emerald-600"
			pill = "bg-emerald-500/80 text-white"
		} else {
			icon = "❌"
			statusText = "MISSING"
			border = "border border-red-200 dark:border-red-700"
			pill = "bg-red-500/90 text-white"
		}

		envVars = append(envVars, EnvVarInfo{
			Name:   envVar,
			Status: statusText,
			Icon:   icon,
			Border: border,
			Pill:   pill,
		})
	}

	// Binaries
	binaries := []BinaryInfo{
		{Name: "flyctl", Path: cfg.Binaries.Flyctl},
		{Name: "ko", Path: cfg.Binaries.Ko},
		{Name: "caddy", Path: cfg.Binaries.Caddy},
		{Name: "task", Path: cfg.Binaries.Task},
		{Name: "tofu", Path: cfg.Binaries.Tofu},
	}

	// Status text
	envStatusText := fmt.Sprintf("%d of %d set", len(envStatus)-len(missingVars), len(envStatus))

	return ConfigTemplateData{
		LastUpdatedDisplay:  time.Now().Format("15:04:05"),
		LastUpdatedISO:      time.Now().Format(time.RFC3339),
		EnvironmentIcon:     envIcon,
		EnvironmentTitle:    envTitle,
		EnvironmentDesc:     envDesc,
		EnvironmentBorder:   envBorder,
		EnvironmentGradient: envGradient,
		Config:              configInfo,
		EnvVars:             envVars,
		Binaries:            binaries,
		AllEnvVarsSet:       len(missingVars) == 0,
		EnvStatusText:       envStatusText,
	}
}