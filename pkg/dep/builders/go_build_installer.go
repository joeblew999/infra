package builders

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/joeblew999/infra/pkg/config"
	"github.com/joeblew999/infra/pkg/dep/storage"
	"github.com/joeblew999/infra/pkg/log"
)

// Platform represents a target build platform
type Platform struct {
	OS   string `json:"os"`
	Arch string `json:"arch"`
}

// GoBuildInstaller uses pkg/deck's build pattern for Go source compilation
type GoBuildInstaller struct{}

// fileExists checks if a file exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func (i *GoBuildInstaller) Install(name, repo, pkg, version string, debug bool) error {
	return i.InstallWithPlatforms(name, repo, pkg, version, debug, nil)
}

func (i *GoBuildInstaller) InstallWithPlatforms(name, repo, pkg, version string, debug bool, platforms []Platform) error {
	if len(platforms) > 1 {
		platformList := make([]string, len(platforms))
		for i, p := range platforms {
			platformList[i] = fmt.Sprintf("%s/%s", p.OS, p.Arch)
		}
		log.Info("Installing via 2-phase pipeline (cross-platform)", "binary", name, "repo", repo, "package", pkg, "platforms", platformList)
	} else {
		log.Info("Installing via 2-phase pipeline", "binary", name, "repo", repo, "package", pkg)
	}

	// Parse owner/repo from repo string
	parts := strings.Split(repo, "/")
	if len(parts) != 2 {
		return fmt.Errorf("invalid repo format: %s (expected owner/repo)", repo)
	}
	owner, repoName := parts[0], parts[1]

	// If no platforms specified, default to current platform
	if len(platforms) == 0 {
		platforms = []Platform{{OS: runtime.GOOS, Arch: runtime.GOARCH}}
	}

	// Skip GitHub Packages optimization for go-build sources
	// These are meant to be built from source and don't require external dependencies
	log.Info("Building from source (go-build)", "binary", name, "platforms", len(platforms))

	// Phase 2: Build from source using git clone + go build
	// Create isolated build directory in .dep/build
	buildDir := filepath.Join(config.GetDepPath(), "build", name)
	if err := os.MkdirAll(buildDir, 0755); err != nil {
		return fmt.Errorf("failed to create build directory: %w", err)
	}
	defer os.RemoveAll(buildDir) // Clean up after build

	// Clone the repository
	repoURL := fmt.Sprintf("https://github.com/%s.git", repo)
	log.Info("Cloning repository", "url", repoURL, "version", version)
	
	cloneCmd := exec.Command("git", "clone", "--depth", "1", repoURL, buildDir)
	if version != "latest" {
		// For specific versions, clone with branch/tag
		cloneCmd = exec.Command("git", "clone", "--depth", "1", "--branch", version, repoURL, buildDir)
	}
	cloneCmd.Stdout = os.Stdout
	if debug {
		cloneCmd.Stderr = os.Stderr
	}

	if err := cloneCmd.Run(); err != nil {
		return fmt.Errorf("failed to clone repository %s: %w", repo, err)
	}

	// Create isolated go.mod in build directory to avoid go.work interference
	goModPath := filepath.Join(buildDir, "go.mod")
	if _, err := os.Stat(goModPath); os.IsNotExist(err) {
		// If no go.mod exists, create a minimal one
		goModContent := "module build\n\ngo 1.21\n"
		if err := os.WriteFile(goModPath, []byte(goModContent), 0644); err != nil {
			return fmt.Errorf("failed to create go.mod: %w", err)
		}
	}

	// Determine the build path - handle both full package paths and relative paths
	// e.g., github.com/benbjohnson/litestream/cmd/litestream -> ./cmd/litestream
	// e.g., tools/goctl -> ./tools/goctl
	repoPrefix := fmt.Sprintf("github.com/%s/", repo)
	buildPath := "./"
	if suffix, found := strings.CutPrefix(pkg, repoPrefix); found {
		// Full package path - extract relative part
		buildPath = "./" + suffix
	} else if pkg != "" && pkg != "." {
		// Already relative path - use as is
		buildPath = "./" + pkg
	}

	var builtPaths []string

	// Build for each platform
	for _, platform := range platforms {
		// Determine output path
		var outputPath string
		if len(platforms) == 1 {
			// Single platform: use standard .dep/binary_name path
			outputPath = filepath.Join(config.GetDepPath(), name)
			if platform.OS == "windows" {
				outputPath += ".exe"
			}
		} else {
			// Multi-platform: use .dep/binary_name-os-arch path
			binaryName := fmt.Sprintf("%s-%s-%s", name, platform.OS, platform.Arch)
			if platform.OS == "windows" {
				binaryName += ".exe"
			}
			outputPath = filepath.Join(config.GetDepPath(), binaryName)
		}

		absOutputPath, err := filepath.Abs(outputPath)
		if err != nil {
			return fmt.Errorf("failed to get absolute path for %s: %w", outputPath, err)
		}

		// Determine working directory and build command
		workDir := buildDir
		buildTarget := "."
		
		// If the package is a subdirectory with its own go.mod, change to that directory
		if pkg != "" && pkg != "." {
			packageDir := filepath.Join(buildDir, pkg)
			if goModPath := filepath.Join(packageDir, "go.mod"); fileExists(goModPath) {
				workDir = packageDir
				buildTarget = "."
				log.Info("Found separate go.mod in package directory", "package", pkg, "workdir", workDir)
			} else {
				buildTarget = buildPath
			}
		}

		// Build the binary for this platform
		log.Info("Building binary", "platform", fmt.Sprintf("%s/%s", platform.OS, platform.Arch), "build_target", buildTarget, "workdir", workDir, "output", absOutputPath)
		
		buildCmd := exec.Command("go", "build", "-o", absOutputPath, buildTarget)
		buildCmd.Dir = workDir
		// Prepare environment for cross-compilation
		env := append(os.Environ(), 
			"GO111MODULE=on",
			"GOWORK=off", // Disable go.work to avoid workspace interference
			"GOOS="+platform.OS,
			"GOARCH="+platform.Arch,
		)
		
		// For cross-platform builds, disable CGO by default unless building for current platform
		if platform.OS != runtime.GOOS || platform.Arch != runtime.GOARCH {
			env = append(env, "CGO_ENABLED=0")
		}
		
		buildCmd.Env = env
		
		// Only show output for current platform or if debug is enabled
		if platform.OS == runtime.GOOS && platform.Arch == runtime.GOARCH || debug {
			buildCmd.Stdout = os.Stdout
			buildCmd.Stderr = os.Stderr
		}

		if err := buildCmd.Run(); err != nil {
			// Check if this is a CGO cross-compilation error
			if platform.OS != runtime.GOOS || platform.Arch != runtime.GOARCH {
				return fmt.Errorf("failed to build %s for %s/%s (cross-compilation): %w\n\nThis may be due to CGO dependencies that require platform-specific C toolchains.\nConsider using github-release source instead of go-build for CGO-enabled binaries.", name, platform.OS, platform.Arch, err)
			}
			return fmt.Errorf("failed to build %s for %s/%s: %w", name, platform.OS, platform.Arch, err)
		}

		// Ensure executable permissions (Unix platforms)
		if platform.OS != "windows" {
			if err := os.Chmod(absOutputPath, 0755); err != nil {
				return fmt.Errorf("failed to set executable permissions for %s: %w", absOutputPath, err)
			}
		}

		builtPaths = append(builtPaths, absOutputPath)
		log.Info("Successfully built", "platform", fmt.Sprintf("%s/%s", platform.OS, platform.Arch), "path", absOutputPath)
	}

	// Phase 3: Upload to GitHub Packages for future use (single platform only)
	if len(platforms) == 1 && platforms[0].OS == runtime.GOOS && platforms[0].Arch == runtime.GOARCH {
		uploadStorage := storage.NewGitHub()
		if err := uploadStorage.UploadToPackages(owner, repoName, name, version, builtPaths[0]); err != nil {
			log.Warn("Failed to upload to GitHub Packages, but binary is available locally", "error", err)
		} else {
			log.Info("Uploaded to GitHub Packages", "binary", name, "version", version)
		}
	}

	if len(platforms) > 1 {
		log.Info("Successfully installed cross-platform", "binary", name, "platforms", len(platforms), "binaries_created", len(builtPaths))
	} else {
		log.Info("Successfully installed", "binary", name, "path", builtPaths[0])
	}
	return nil
}
