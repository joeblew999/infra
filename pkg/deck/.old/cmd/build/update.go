package build

import (
	"fmt"
	"os"

	"github.com/joeblew999/infra/pkg/deck"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update deck tools to latest source",
	Long: `Update deck tools by pulling latest source code and rebuilding.

This cleans all existing builds and rebuilds from fresh source code.`,
	Run: func(cmd *cobra.Command, args []string) {
		manager := deck.NewManager()
		
		if err := manager.Update(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		
		fmt.Println("Deck tools updated successfully!")
	},
}

func init() {
	BuildCmd.AddCommand(updateCmd)
}