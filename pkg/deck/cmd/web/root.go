package web

import (
	"github.com/spf13/cobra"
)

// webCmd represents the web command
var WebCmd = &cobra.Command{
	Use:   "web",
	Short: "Web interface for deck examples",
	Long: `Web interface commands for viewing and managing deck examples.
	
This command group provides web-based tools for:
- Viewing deck examples in a browser
- Generating SVG, PNG, and PDF outputs
- Interactive example exploration`,
}

func init() {
	WebCmd.AddCommand(serveCmd)
}