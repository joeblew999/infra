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
		fmt.Println("  decksh     - dsh to XML compiler")
		fmt.Println("  deckshfmt  - dsh code formatter") 
		fmt.Println("  deckshlint - dsh syntax validator")
		fmt.Println("  decksvg    - XML to SVG converter")
		fmt.Println("  deckpng    - XML to PNG converter")
		fmt.Println("  deckpdf    - XML to PDF converter")
	},
}

func init() {
	deckCmd.AddCommand(listCmd)
}

// GetToolDescription returns description for a specific tool
func GetToolDescription(name string) string {
	descriptions := map[string]string{
		deck.DeckshBinary:     "Compiler that transforms .dsh files into XML",
		deck.DeckfmtBinary:  "Code formatter for .dsh files",
		deck.DecklintBinary: "Syntax validator for .dsh files",
		deck.DecksvgBinary:    "Converts XML deck markup to SVG graphics",
		deck.DeckpngBinary:    "Converts XML deck markup to PNG images",
		deck.DeckpdfBinary:    "Converts XML deck markup to PDF documents",
	}
	
	if desc, ok := descriptions[name]; ok {
		return desc
	}
	return fmt.Sprintf("Deck tool: %s", name)
}