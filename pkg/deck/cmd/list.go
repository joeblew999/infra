package cmd

import (
	"fmt"

	"github.com/joeblew999/infra/pkg/deck"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List available deck tools",
	Long:  "List all available deck tools that can be built from source",
	Run: func(cmd *cobra.Command, args []string) {
		manager := deck.NewManager()
		tools := manager.ListTools()
		
		fmt.Println("Available Deck Tools:")
		fmt.Println("====================")
		
		for _, tool := range tools {
			fmt.Printf("  - %s\n", tool)
		}
		
		fmt.Println("")
		fmt.Println("Tools:")
		fmt.Println("  decksh  - dsh to XML compiler")
		fmt.Println("  dshfmt  - dsh code formatter") 
		fmt.Println("  dshlint - dsh syntax validator")
		fmt.Println("  svgdeck - XML to SVG converter")
		fmt.Println("  pngdeck - XML to PNG converter")
		fmt.Println("  pdfdeck - XML to PDF converter")
	},
}

func init() {
	deckCmd.AddCommand(listCmd)
}

// GetToolDescription returns description for a specific tool
func GetToolDescription(name string) string {
	descriptions := map[string]string{
		"decksh":  "Compiler that transforms .dsh files into XML",
		"dshfmt":  "Code formatter for .dsh files",
		"dshlint": "Syntax validator for .dsh files",
		"svgdeck": "Converts XML deck markup to SVG graphics",
		"pngdeck": "Converts XML deck markup to PNG images",
		"pdfdeck": "Converts XML deck markup to PDF documents",
	}
	
	if desc, ok := descriptions[name]; ok {
		return desc
	}
	return fmt.Sprintf("Deck tool: %s", name)
}