package cmd

import (
	docscmd "github.com/joeblew999/infra/pkg/docs/cmd"
	hugocmd "github.com/joeblew999/infra/pkg/hugo/cmd"
)

// RunDocs adds docs commands to the root command.
func RunDocs() {
	docscmd.Register(rootCmd)
	hugocmd.Register(rootCmd)
}
