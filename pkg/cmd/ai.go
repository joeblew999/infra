package cmd

import (
	"github.com/joeblew999/infra/pkg/ai"
)

// RunAI registers AI commands
func RunAI() {
	rootCmd.AddCommand(ai.NewAICmd())
}