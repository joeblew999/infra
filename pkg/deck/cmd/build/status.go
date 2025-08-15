package build

import (
	"fmt"

	"github.com/joeblew999/infra/pkg/deck"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show build status of deck tools",
	Long:  "Show which deck tools are built and their locations",
	Run: func(cmd *cobra.Command, args []string) {
		manager := deck.NewManager()
		status := manager.Status()
		
		if len(status) == 0 {
			fmt.Println("No deck tools found. Run 'go run . deck install' to build them.")
			return
		}
		
		fmt.Println("Deck Tools Status:")
		fmt.Println("==================")
		
		for tool, paths := range status {
			fmt.Printf("\n%s:\n", tool)
			fmt.Printf("  Binary: %s\n", paths["binary"])
			fmt.Printf("  WASM:   %s\n", paths["wasm"])
		}
	},
}

func init() {
	BuildCmd.AddCommand(statusCmd)
}