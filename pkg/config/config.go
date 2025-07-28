package config

import (
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/joho/godotenv"
)

func init() {
	// Load .env from current dir and parent directories
	if err := godotenv.Load(); err != nil {
		// Try loading from parent directories
		cwd, _ := os.Getwd()
		for dir := cwd; dir != "/"; dir = filepath.Dir(dir) {
			envPath := filepath.Join(dir, ".env")
			if _, err := os.Stat(envPath); err == nil {
				if err := godotenv.Load(envPath); err == nil {
					log.Printf("Loaded .env from: %s", envPath)
					break
				}
			}
		}
	}

	// Load global .env from home directory
	home, _ := os.UserHomeDir()
	globalEnv := filepath.Join(home, ".infra.env")
	if _, err := os.Stat(globalEnv); err == nil {
		godotenv.Load(globalEnv)
		log.Printf("Loaded global .env from: %s", globalEnv)
	}
}

// Config represents the complete configuration structure
// This is used for JSON serialization and CLI display
type Config struct {
	Environment EnvironmentConfig `json:"environment"`
	Platform    PlatformConfig    `json:"platform"`
	Paths       PathsConfig       `json:"paths"`
	Binaries    BinariesConfig    `json:"binaries"`
	Registry    RegistryConfig    `json:"registry"`
	EnvironmentVars []string       `json:"environment_vars"`
	EnvironmentStatus map[string]string `json:"environment_status"`
}

type EnvironmentConfig struct {
	IsProduction   bool `json:"is_production"`
	IsDevelopment  bool `json:"is_development"`
	ShouldUseHTTPS bool `json:"should_use_https"`
}

type PlatformConfig struct {
	GOOS     string `json:"goos"`
	GOARCH   string `json:"goarch"`
	Platform string `json:"platform"`
}

type PathsConfig struct {
	Dep       string `json:"dep"`
	Bin       string `json:"bin"`
	Data      string `json:"data"`
	Docs      string `json:"docs"`
	Taskfiles string `json:"taskfiles"`
}

type BinariesConfig struct {
	Flyctl string `json:"flyctl"`
	Ko     string `json:"ko"`
	Caddy  string `json:"caddy"`
	Task   string `json:"task"`
	Tofu   string `json:"tofu"`
}

type RegistryConfig struct {
	KoDockerRepo string   `json:"ko_docker_repo"`
	KoBaseImage  string   `json:"ko_base_image"`
	KoPlatforms  []string `json:"ko_platforms"`
}

// GetConfig returns the complete configuration structure
func GetConfig() Config {
	return Config{
		Environment: EnvironmentConfig{
			IsProduction:   IsProduction(),
			IsDevelopment:  IsDevelopment(),
			ShouldUseHTTPS: ShouldUseHTTPS(),
		},
		Platform: PlatformConfig{
			GOOS:     runtime.GOOS,
			GOARCH:   runtime.GOARCH,
			Platform: GetPlatform(),
		},
		Paths: PathsConfig{
			Dep:       GetDepPath(),
			Bin:       GetBinPath(),
			Data:      GetDataPath(),
			Docs:      DocsDir,
			Taskfiles: GetTaskfilesPath(),
		},
		Binaries: BinariesConfig{
			Flyctl: GetFlyctlBinPath(),
			Ko:     GetKoBinPath(),
			Caddy:  GetCaddyBinPath(),
			Task:   GetTaskBinPath(),
			Tofu:   GetTofuBinPath(),
		},
		Registry: RegistryConfig{
			KoDockerRepo: GetKoDockerRepo(),
			KoBaseImage:  GetKoDefaultBaseImage(),
			KoPlatforms:  GetKoDefaultPlatforms(),
		},
		EnvironmentVars: []string{
			EnvVarEnvironment,
			EnvVarFlyAppName,
			EnvVarKoDockerRepo,
			"FLY_API_TOKEN",
			"FLY_REGION",
		},
		EnvironmentStatus: GetEnvStatus(),
	}
}
