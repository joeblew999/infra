package cmd

import (
	"github.com/joeblew999/infra/pkg/deck/cmd/build"
	"github.com/spf13/cobra"
)

var deckCmd = &cobra.Command{
	Use:   "deck",
	Short: "Deck visualization tools",
	Long: `Deck provides tools for creating SVG graphics from declarative markup.

This command manages the deck tools by downloading source code and compiling
to both native binaries and WASM modules for use in the system.`,
}

// GetDeckCmd returns the deck command for registration
func GetDeckCmd() *cobra.Command {
	return deckCmd
}

func init() {
	deckCmd.AddCommand(build.BuildCmd)
}