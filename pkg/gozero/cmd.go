package gozero

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/joeblew999/infra/pkg/config"
	"github.com/joeblew999/infra/pkg/dep"
	"github.com/joeblew999/infra/pkg/log"
)

// GoZeroRunner manages goctl command execution with infra patterns
type GoZeroRunner struct {
	debug   bool
	workDir string
}

// NewGoZeroRunner creates a new go-zero command runner
func NewGoZeroRunner(debug bool) *GoZeroRunner {
	return &GoZeroRunner{
		debug:   debug,
		workDir: ".",
	}
}

// SetWorkDir sets the working directory for goctl commands
func (r *GoZeroRunner) SetWorkDir(dir string) {
	r.workDir = dir
}

// ApiGenerate generates go-zero API service from .api file
func (r *GoZeroRunner) ApiGenerate(apiFile, outputDir string) error {
	args := []string{"api", "go", "-api", apiFile, "-dir", outputDir}
	return r.runGoctl("generate API service", args...)
}

// ApiSwagger generates Swagger documentation from .api file  
func (r *GoZeroRunner) ApiSwagger(apiFile, outputDir string) error {
	args := []string{"api", "swagger", "-api", apiFile, "-dir", outputDir}
	return r.runGoctl("generate Swagger docs", args...)
}

// ApiNew creates a new API service project
func (r *GoZeroRunner) ApiNew(serviceName, outputDir string) error {
	args := []string{"api", "new", serviceName, "-dir", outputDir}
	return r.runGoctl("create new API service", args...)
}

// ApiFormat formats .api files
func (r *GoZeroRunner) ApiFormat(apiFile string) error {
	args := []string{"api", "format", "-api", apiFile}
	return r.runGoctl("format API file", args...)
}

// ApiValidate validates .api file syntax
func (r *GoZeroRunner) ApiValidate(apiFile string) error {
	args := []string{"api", "validate", "-api", apiFile}
	return r.runGoctl("validate API file", args...)
}

// RpcNew creates a new RPC service
func (r *GoZeroRunner) RpcNew(serviceName, outputDir string) error {
	args := []string{"rpc", "new", serviceName, "-dir", outputDir}
	return r.runGoctl("create new RPC service", args...)
}

// ModelGenerate generates model code from database
func (r *GoZeroRunner) ModelGenerate(dsn, table, outputDir string) error {
	args := []string{"model", "mysql", "datasource", "-url", dsn, "-table", table, "-dir", outputDir}
	return r.runGoctl("generate model code", args...)
}

// DockerGenerate generates Dockerfile
func (r *GoZeroRunner) DockerGenerate(goFile string, options DockerOptions) error {
	args := []string{"docker", "--go", goFile}
	
	if options.Port > 0 {
		args = append(args, "--port", fmt.Sprintf("%d", options.Port))
	}
	if options.Base != "" {
		args = append(args, "--base", options.Base)
	}
	if options.Exe != "" {
		args = append(args, "--exe", options.Exe)
	}
	
	return r.runGoctl("generate Dockerfile", args...)
}

// KubeGenerate generates Kubernetes deployment files
func (r *GoZeroRunner) KubeGenerate(serviceName, namespace, image string, port int, outputDir string) error {
	args := []string{"kube", "deploy", "-name", serviceName, "-namespace", namespace, "-image", image}
	
	if port > 0 {
		args = append(args, "-port", fmt.Sprintf("%d", port))
	}
	if outputDir != "" {
		args = append(args, "-o", outputDir)
	}
	
	return r.runGoctl("generate Kubernetes files", args...)
}

// QuickStart creates a quickstart project
func (r *GoZeroRunner) QuickStart(serviceType, outputDir string) error {
	args := []string{"quickstart", "-t", serviceType, "-dir", outputDir}
	return r.runGoctl("create quickstart project", args...)
}

// DockerOptions configures Docker generation
type DockerOptions struct {
	Port int
	Base string
	Exe  string
}

// runGoctl executes goctl with the given arguments
func (r *GoZeroRunner) runGoctl(operation string, args ...string) error {
	// Ensure goctl binary is installed before running
	if err := dep.InstallBinary(config.BinaryGoctl, false); err != nil {
		return fmt.Errorf("failed to ensure goctl binary: %w", err)
	}

	goctlPath := config.Get(config.BinaryGoctl)

	// Handle relative paths by finding the repo root
	if !filepath.IsAbs(goctlPath) {
		wd, _ := os.Getwd()
		for dir := wd; dir != "/" && dir != "."; {
			if _, err := os.Stat(filepath.Join(dir, "go.work")); err == nil {
				goctlPath = filepath.Join(dir, goctlPath)
				break
			}
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
	}
	
	log.Info("Running goctl", "operation", operation, "args", args, "workdir", r.workDir)
	
	cmd := exec.Command(goctlPath, args...)
	cmd.Dir = r.workDir
	
	if r.debug {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("goctl %s failed: %w", operation, err)
	}
	
	log.Info("goctl operation completed", "operation", operation)
	return nil
}

// GenerateInfraAPI generates a go-zero API service following infra patterns
func (r *GoZeroRunner) GenerateInfraAPI(ctx context.Context, packageName string, apiContent string, outputDir string) error {
	// 1. Create the API file
	apiFile := filepath.Join(outputDir, packageName+".api")
	if err := os.WriteFile(apiFile, []byte(apiContent), 0644); err != nil {
		return fmt.Errorf("failed to write API file: %w", err)
	}
	
	// 2. Generate the service
	if err := r.ApiGenerate(apiFile, outputDir); err != nil {
		return fmt.Errorf("failed to generate API service: %w", err)
	}
	
	// 3. Generate Swagger docs  
	if err := r.ApiSwagger(apiFile, outputDir); err != nil {
		log.Warn("Failed to generate Swagger docs", "error", err)
	}
	
	// 4. Initialize go.mod if needed
	goModPath := filepath.Join(outputDir, "go.mod")
	if _, err := os.Stat(goModPath); os.IsNotExist(err) {
		log.Info("Initializing go.mod", "package", packageName)
		initCmd := exec.Command("go", "mod", "init", fmt.Sprintf("github.com/joeblew999/infra/api/%s", packageName))
		initCmd.Dir = outputDir
		if err := initCmd.Run(); err != nil {
			log.Warn("Failed to initialize go.mod", "error", err)
		}
	}
	
	// 5. Run go mod tidy
	tidyCmd := exec.Command("go", "mod", "tidy")
	tidyCmd.Dir = outputDir
	if err := tidyCmd.Run(); err != nil {
		log.Warn("Failed to run go mod tidy", "error", err)
	}
	
	log.Info("Infra API generation completed", "package", packageName, "output", outputDir)
	return nil
}

// Plugin management with individual install functions
// This follows the pattern where plugins are treated as separate binaries

// InstallOpenAPIPlugin installs the goctl-openapi plugin
func (r *GoZeroRunner) InstallOpenAPIPlugin() error {
	return r.installPlugin("goctl-openapi", "jayvynl/goctl-openapi")
}

// InstallGoCompactPlugin installs the goctl-go-compact plugin
func (r *GoZeroRunner) InstallGoCompactPlugin() error {
	return r.installPlugin("goctl-go-compact", "zeromicro/goctl-go-compact")
}

// InstallPHPPlugin installs the goctl-php plugin
func (r *GoZeroRunner) InstallPHPPlugin() error {
	return r.installPlugin("goctl-php", "zeromicro/goctl-php")
}

// InstallSwagPlugin installs the goctl-swag plugin
func (r *GoZeroRunner) InstallSwagPlugin() error {
	return r.installPlugin("goctl-swag", "CloverOS/goctl-swag")
}

// InstallGenginPlugin installs the gengin plugin
func (r *GoZeroRunner) InstallGenginPlugin() error {
	return r.installPlugin("gengin", "MasterJoyHunan/gengin")
}

// installPlugin is a helper that installs a plugin using go install
func (r *GoZeroRunner) installPlugin(pluginName, repo string) error {
	log.Info("Installing goctl plugin", "plugin", pluginName, "repo", repo)

	// Install plugin using go install
	installPath := fmt.Sprintf("github.com/%s@latest", repo)
	cmd := exec.Command("go", "install", installPath)

	if r.debug {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install plugin %s: %w", pluginName, err)
	}

	log.Info("Plugin installed successfully", "plugin", pluginName)
	return nil
}

// RunPlugin executes a goctl plugin
func (r *GoZeroRunner) RunPlugin(pluginName, apiFile, outputDir string, extraArgs ...string) error {
	// Build plugin command arguments
	args := []string{"api", "plugin", "-p", pluginName, "-api", apiFile, "-dir", outputDir}
	args = append(args, extraArgs...)

	return r.runGoctl(fmt.Sprintf("run plugin %s", pluginName), args...)
}