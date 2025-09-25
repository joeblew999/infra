package config

import shared "github.com/joeblew999/infra/core/pkg/shared/config"

// Paths bundles the resolved filesystem locations required by the orchestrator
// runtime. Runtime packages should use these helpers instead of recomputing
// joins against shared config to keep the directory layout consistent.
type Paths struct {
	AppRoot  string
	Dep      string
	Bin      string
	Data     string
	Logs     string
	TestData string
}

// Settings represents the high-level runtime configuration derived from shared
// helpers. Additional runtime-specific values (service specs, feature flags,
// etc.) should be added here as the orchestrator matures.
type Settings struct {
	Environment       string
	Paths             Paths
	EnsureBusCluster  bool
	IsTestEnvironment bool
	IsProduction      bool
}

// Load constructs a Settings value using the shared configuration helpers. It
// should be the single entry point for runtime packages needing configuration
// values so changes in shared helpers propagate automatically.
func Load() Settings {
	return Settings{
		Environment: shared.Environment(),
		Paths: Paths{
			AppRoot:  shared.GetAppRoot(),
			Dep:      shared.GetDepPath(),
			Bin:      shared.GetBinPath(),
			Data:     shared.GetDataPath(),
			Logs:     shared.GetLogsPath(),
			TestData: shared.GetTestDataPath(),
		},
		EnsureBusCluster:  shared.ShouldEnsureBusCluster(),
		IsTestEnvironment: shared.IsTestEnvironment(),
		IsProduction:      shared.IsProduction(),
	}
}
