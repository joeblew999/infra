package dep

import (
	"fmt"

	"github.com/joeblew999/infra/pkg/dep/builders"
)

type nscInstaller struct{}

func (i *nscInstaller) Install(binary DepBinary, debug bool) error {
	builder := builders.GitHubReleaseInstaller{}
	assets := make([]builders.AssetSelector, 0, len(binary.Assets))
	for _, asset := range binary.Assets {
		assets = append(assets, builders.AssetSelector{
			OS:    asset.OS,
			Arch:  asset.Arch,
			Match: asset.Match,
		})
	}
	if err := builder.Install(binary.Name, binary.Repo, binary.Version, assets, debug); err != nil {
		return fmt.Errorf("nsc install failed: %w", err)
	}
	return nil
}
