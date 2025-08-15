package cmd

import (
	"github.com/joeblew999/infra/pkg/deck/cmd"
)

// RunDeck registers deck commands
func RunDeck() {
	rootCmd.AddCommand(cmd.GetDeckCmd())
}