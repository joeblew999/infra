package workflows

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/joeblew999/infra/pkg/config"
	"github.com/joeblew999/infra/pkg/log"
)

// BinaryBuildOptions configures binary build behavior
type BinaryBuildOptions struct {
	OutputDir    string
	BinaryName   string
	Platforms    []string
	Environments []string
	BuildAll     bool
	Verbose      bool
	LocalOnly    bool
}

// BinaryBuildWorkflow handles cross-platform binary builds
type BinaryBuildWorkflow struct {
	opts BinaryBuildOptions
}

// NewBinaryBuildWorkflow creates a new binary build workflow
func NewBinaryBuildWorkflow(opts BinaryBuildOptions) *BinaryBuildWorkflow {
	// Set defaults
	if opts.OutputDir == "" {
		opts.OutputDir = config.GetBinPath()
	}
	if opts.BinaryName == "" {
		opts.BinaryName = "infra"
	}
	if len(opts.Platforms) == 0 && !opts.BuildAll {
		// Default to local platform
		opts.Platforms = []string{runtime.GOOS}
	}
	if len(opts.Environments) == 0 && !opts.BuildAll {
		// Default to local architecture
		opts.Environments = []string{runtime.GOARCH}
	}
	if opts.BuildAll {
		// Build for all supported platforms and architectures
		opts.Platforms = []string{"linux", "darwin", "windows"}
		opts.Environments = []string{"amd64", "arm64"}
	}

	return &BinaryBuildWorkflow{opts: opts}
}

// Execute runs the binary build workflow
func (b *BinaryBuildWorkflow) Execute() error {
	log.Info("Starting binary build workflow",
		"output_dir", b.opts.OutputDir,
		"binary_name", b.opts.BinaryName,
		"platforms", b.opts.Platforms,
		"architectures", b.opts.Environments,
		"build_all", b.opts.BuildAll)

	// Ensure output directory exists
	if err := os.MkdirAll(b.opts.OutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Collect metadata for all built binaries
	var binaries []BinaryMetadata
	
	// Build for each platform/arch combination
	for _, platform := range b.opts.Platforms {
		for _, arch := range b.opts.Environments {
			if b.opts.LocalOnly && (platform != runtime.GOOS || arch != runtime.GOARCH) {
				continue // Skip non-local builds when local-only is set
			}

			metadata, err := b.buildForPlatform(platform, arch)
			if err != nil {
				return fmt.Errorf("failed to build for %s/%s: %w", platform, arch, err)
			}
			binaries = append(binaries, metadata)
		}
	}

	// Generate meta.json
	if err := generateMetaJSON(b.opts, binaries); err != nil {
		return fmt.Errorf("failed to generate meta.json: %w", err)
	}

	log.Info("Binary build completed successfully", "binaries", len(binaries))
	return nil
}

// buildForPlatform builds the binary for a specific platform and architecture
func (b *BinaryBuildWorkflow) buildForPlatform(platform, arch string) (BinaryMetadata, error) {
	outputName := b.getBinaryName(platform, arch)
	outputPath := filepath.Join(b.opts.OutputDir, outputName)

	log.Info("Building binary",
		"platform", platform,
		"arch", arch,
		"output", outputPath)

	// Set build environment variables
	buildFlags := b.getLDFlags()
	env := append(os.Environ(),
		fmt.Sprintf("GOOS=%s", platform),
		fmt.Sprintf("GOARCH=%s", arch),
		"CGO_ENABLED=0", // Static linking for portability
	)

	// Build arguments
	args := []string{
		"build",
		"-o", outputPath,
		"-ldflags", buildFlags,
		"-trimpath", // Remove filesystem paths from binary
		".",
	}

	if b.opts.Verbose {
		log.Info("Running go build", "args", args)
	}

	buildStart := time.Now()
	
	// Execute go build
	cmd := exec.Command("go", args...)
	cmd.Env = env
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return BinaryMetadata{}, fmt.Errorf("go build failed: %w", err)
	}

	// Make binary executable on Unix-like systems
	if platform != "windows" {
		if err := os.Chmod(outputPath, 0755); err != nil {
			return BinaryMetadata{}, fmt.Errorf("failed to set executable permissions: %w", err)
		}
	}

	// Get file info for metadata
	info, err := os.Stat(outputPath)
	if err != nil {
		return BinaryMetadata{}, fmt.Errorf("failed to get file info: %w", err)
	}

	// Calculate SHA256 hash
	hash, err := calculateSHA256(outputPath)
	if err != nil {
		return BinaryMetadata{}, fmt.Errorf("failed to calculate hash: %w", err)
	}

	metadata := BinaryMetadata{
		Name:       b.opts.BinaryName,
		Platform:   platform,
		Arch:       arch,
		Filename:   outputName,
		Path:       outputPath,
		Size:       info.Size(),
		SHA256:     hash,
		BuildTime:  buildStart,
		GoVersion:  runtime.Version(),
		BuildFlags: buildFlags,
		CGOEnabled: "0",
	}

	// Try to get git info using centralized approach
	commit := config.GetRuntimeGitHash()
	branch := config.GetRuntimeGitBranch()
	metadata.GitCommit = commit
	metadata.GitBranch = branch

	log.Info("Binary built successfully", "path", outputPath, "size", info.Size(), "sha256", hash)
	return metadata, nil
}

// getBinaryName returns the appropriate binary name for the platform/arch
func (b *BinaryBuildWorkflow) getBinaryName(platform, arch string) string {
	name := fmt.Sprintf("%s_%s_%s", b.opts.BinaryName, platform, arch)
	if platform == "windows" {
		name += ".exe"
	}
	return name
}

// getLDFlags returns the linker flags for the build
func (b *BinaryBuildWorkflow) getLDFlags() string {
	var flags []string
	
	// Strip debug info and symbol table
	flags = append(flags, "-s", "-w")
	
	// Inject git info and build time using centralized config package
	commit := config.GetRuntimeGitHash()
	if commit != "" {
		flags = append(flags, fmt.Sprintf("-X github.com/joeblew999/infra/pkg/cmd.GitHash=%s", commit))
	}
	flags = append(flags, fmt.Sprintf("-X github.com/joeblew999/infra/pkg/cmd.BuildTime=%s", time.Now().UTC().Format(time.RFC3339)))
	
	return strings.Join(flags, " ")
}

// GetSupportedPlatforms returns all supported platform/arch combinations
func GetSupportedPlatforms() []string {
	return []string{
		"linux/amd64",
		"linux/arm64",
		"darwin/amd64",
		"darwin/arm64",
		"windows/amd64",
		"windows/arm64",
	}
}

// BinaryMetadata represents metadata for a built binary
type BinaryMetadata struct {
	Name        string    `json:"name"`
	Platform    string    `json:"platform"`
	Arch        string    `json:"arch"`
	Filename    string    `json:"filename"`
	Path        string    `json:"path"`
	Size        int64     `json:"size"`
	SHA256      string    `json:"sha256"`
	BuildTime   time.Time `json:"build_time"`
	GoVersion   string    `json:"go_version"`
	GitCommit   string    `json:"git_commit,omitempty"`
	GitBranch   string    `json:"git_branch,omitempty"`
	BuildFlags  string    `json:"build_flags"`
	CGOEnabled  string    `json:"cgo_enabled"`
}

// BuildMeta represents the complete metadata for all built binaries
type BuildMeta struct {
	BinaryName   string           `json:"binary_name"`
	BuildTime    time.Time        `json:"build_time"`
	GoVersion    string           `json:"go_version"`
	GitCommit    string           `json:"git_commit,omitempty"`
	GitBranch    string           `json:"git_branch,omitempty"`
	BuildHost    string           `json:"build_host"`
	BuildUser    string           `json:"build_user"`
	Binaries     []BinaryMetadata `json:"binaries"`
	TotalSize    int64            `json:"total_size"`
	TotalCount   int              `json:"total_count"`
}

// GetLocalBinaryName returns the binary name for the current platform
func GetLocalBinaryName(name string) string {
	binaryName := fmt.Sprintf("%s_%s_%s", name, runtime.GOOS, runtime.GOARCH)
	if runtime.GOOS == "windows" {
		binaryName += ".exe"
	}
	return binaryName
}

// Removed: git functions now centralized in pkg/build.GetRuntimeGitHash/Branch()

// calculateSHA256 calculates SHA256 hash of a file
func calculateSHA256(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	
	hash := sha256.Sum256(data)
	return fmt.Sprintf("%x", hash), nil
}

// generateMetaJSON generates metadata for all built binaries
func generateMetaJSON(opts BinaryBuildOptions, binaries []BinaryMetadata) error {
	commit := config.GetRuntimeGitHash()
	branch := config.GetRuntimeGitBranch()
	
	buildHost, _ := os.Hostname()
	buildUser := os.Getenv("USER")
	if buildUser == "" {
		buildUser = "unknown"
	}
	
	var totalSize int64
	for _, b := range binaries {
		totalSize += b.Size
	}
	
	meta := BuildMeta{
		BinaryName:  opts.BinaryName,
		BuildTime:   time.Now().UTC(),
		GoVersion:   runtime.Version(),
		GitCommit:   commit,
		GitBranch:   branch,
		BuildHost:   buildHost,
		BuildUser:   buildUser,
		Binaries:    binaries,
		TotalSize:   totalSize,
		TotalCount:  len(binaries),
	}
	
	metaPath := filepath.Join(opts.OutputDir, "meta.json")
	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}
	
	if err := os.WriteFile(metaPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write meta.json: %w", err)
	}
	
	log.Info("Generated meta.json", "path", metaPath, "binaries", len(binaries))
	return nil
}