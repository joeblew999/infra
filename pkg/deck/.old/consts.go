package deck

import (
	"path/filepath"
	"runtime"
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
	DeckshRepo  = "https://github.com/ajstarks/deck.git"
	SvgdeckRepo = "https://github.com/ajstarks/svgo.git"
	DshfmtRepo  = "https://github.com/ajstarks/dshfmt.git"
	DshlintRepo = "https://github.com/ajstarks/dshlint.git"

	// Build directory structure
	BuildRoot  = "pkg/deck/.build"
	SourceDir  = "pkg/deck/.source"
	ReleaseDir = "pkg/deck/.release"
	GoWorkFile = "pkg/deck/.build/go.work"

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
	return filepath.Join(BuildRoot, "bin", name)
}

// GetWASMPath returns the path to a WASM module in the build directory
func GetWASMPath(name string) string {
	return filepath.Join(BuildRoot, "wasm", name)
}
