package dep

import shareddep "github.com/joeblew999/infra/core/pkg/shared/dep"

// EnsureManifest installs all binaries defined in the manifest using the
// provided installer. This thin wrapper allows runtime packages to control the
// installer used (e.g., real download vs. stub) while keeping imports pointed
// at the shared module.
func EnsureManifest(manifest *shareddep.Manifest, installer shareddep.Installer) (map[string]string, error) {
	return manifest.EnsureAll(installer)
}

// DefaultInstaller exposes the shared default installer to runtime packages.
var DefaultInstaller = shareddep.DefaultInstaller{}

// Re-export shared types so runtime code can stay within the runtime namespace.
type (
	Manifest   = shareddep.Manifest
	BinarySpec = shareddep.BinarySpec
	Asset      = shareddep.Asset
	Source     = shareddep.Source
)

const (
	SourcePlaceholder   = shareddep.SourcePlaceholder
	SourceGithubRelease = shareddep.SourceGithubRelease
	SourceGoBuild       = shareddep.SourceGoBuild
)
