package cli

import shared "github.com/joeblew999/infra/core/pkg/shared/cli"

// BuilderOptions re-exports the shared CLI builder options.
type BuilderOptions = shared.BuilderOptions

var (
	NewRootCommand = shared.NewRootCommand
	AddCommand     = shared.AddCommand
	MarkRequired   = shared.MarkRequired
)
