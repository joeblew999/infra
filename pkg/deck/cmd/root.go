package cmd

import (
	"github.com/joeblew999/infra/pkg/deck/cmd/build"
	"github.com/joeblew999/infra/pkg/deck/cmd/web"
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
	deckCmd.AddCommand(web.WebCmd)
	deckCmd.AddCommand(WatchCmd)
	deckCmd.AddCommand(UpdateCmd)
	deckCmd.AddCommand(HealthCmd)
	
	// Set up flags for watch command
	WatchCmd.Flags().String("formats", "svg", "Output formats to generate (comma-separated): svg,png,pdf")
	
	// Set up flags for update command (simplified)
	UpdateCmd.Flags().BoolP("force", "f", false, "Force update by removing existing .source directory")
	
	// Set up flags for health command
	HealthCmd.Flags().BoolP("verbose", "v", false, "Verbose output during health checks")
	HealthCmd.Flags().BoolP("json", "j", false, "Output health report in JSON format")
	HealthCmd.Flags().String("tool", "", "Check specific tool only (e.g., decksh, svgdeck)")
}