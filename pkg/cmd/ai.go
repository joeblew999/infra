package cmd

import (
	"github.com/joeblew999/infra/pkg/ai"
	"github.com/spf13/cobra"
)

// RunAI registers AI commands - now moved to CLI namespace
func RunAI() {
	// AI commands now added via AddAIToCLI in cli.go
}

// AddAIToCLI adds AI commands to the CLI namespace
func AddAIToCLI(cliParent *cobra.Command) {
	cliParent.AddCommand(ai.NewAICmd())
}