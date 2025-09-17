package cmd

import (
	docscmd "github.com/joeblew999/infra/pkg/docs/cmd"
)

// RunDocs adds docs commands to the root command
func RunDocs() {
	rootCmd.AddCommand(docscmd.GetDocsCmd())
}

