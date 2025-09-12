package deck

import (
	"path/filepath"
	"runtime"
	
	"github.com/joeblew999/infra/pkg/config"
)

// Build and release constants
const (
	// Repository information
	RepoOwner = "joeblew999"
	RepoName  = "infra"

	// Release tag prefix
	ReleaseTagPrefix = "deck-"
	
	// GitHub release configuration
	GitHubOwner = "joeblew999"
	GitHubRepo  = "infra"

	// Package name for releases
	PackageName = "deck-tools"

	// Source repository URLs
	DeckshRepo  = "https://github.com/ajstarks/decksh.git"  // decksh, dshfmt, dshlint tools
	SvgdeckRepo = "https://github.com/ajstarks/deck.git"    // svgdeck, pngdeck, pdfdeck tools

	// Package directory structure
	PkgDir     = "pkg/deck"
	
	// Build directory structure (legacy constants - use config functions instead)
	RepoTestsDir = "pkg/deck/repo-tests"  // Upstream repo examples (525+ DSH files)
	UnitTestsDir = "pkg/deck/unit-tests"  // Our focused unit tests
	ReleaseDir   = "pkg/deck/.release"

	// Binary names with deck prefix
	DeckshBinary   = "decksh"
	DeckfmtBinary  = "deckfmt"
	DecklintBinary = "decklint"
	DecksvgBinary  = "decksvg"
	DeckpngBinary  = "deckpng"
	DeckpdfBinary  = "deckpdf"

	// WASM module names - consistent with deck prefix
	DeckshWASM     = "decksh.wasm"
	DeckshfmtWASM  = "decksh-fmt.wasm"
	DeckshlintWASM = "decksh-lint.wasm"
	DecksvgWASM    = "decksvg.wasm"
	DeckpngWASM    = "deckpng.wasm"
	DeckpdfWASM    = "deckpdf.wasm"

	// File watcher timing constants
	WatcherPollInterval     = 2  // seconds between filesystem scans
	FileModificationTimeout = 10 // seconds to wait before processing modified files

	// Health check constants
	HealthCheckTimeout     = 30 // seconds timeout for health operations
	TempDirPrefix         = "deck-health-"
	
	// System dependencies
	GitCommand = "git"
	GoCommand  = "go"
)

// Tool collections for packaging (use constants to prevent obfuscation)
var (
	AllBinaries = []string{DeckshBinary, DeckfmtBinary, DecklintBinary, DecksvgBinary, DeckpngBinary, DeckpdfBinary}
	WASMBinaries = []string{DeckshBinary, DecksvgBinary}
)

// GetBuildTarget returns the build target for the current platform
func GetBuildTarget() string {
	return runtime.GOOS + "_" + runtime.GOARCH
}

// GetReleaseFilename returns the release filename for a given platform and arch
func GetReleaseFilename(version, platform, arch string) string {
	if platform == "windows" {
		return PackageName + "-" + version + "-" + platform + "-" + arch + ".zip"
	}
	return PackageName + "-" + version + "-" + platform + "-" + arch + ".tar.gz"
}

// GetBinaryPath returns the path to a binary in the build directory
func GetBinaryPath(name string) string {
	return filepath.Join(config.GetDeckBinPath(), name)
}

// GetWASMPath returns the path to a WASM module in the build directory
func GetWASMPath(name string) string {
	return filepath.Join(config.GetDeckWASMPath(), name)
}

// GetBuildRoot returns the environment-aware build root directory
func GetBuildRoot() string {
	return config.GetDeckPath()
}

// GetCachePath returns the environment-aware cache directory
func GetCachePath() string {
	return config.GetDeckCachePath()
}

// GetGoWorkFile returns the environment-aware go.work file path
func GetGoWorkFile() string {
	return filepath.Join(GetBuildRoot(), "go.work")
}
