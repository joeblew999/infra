package build

import (
	"fmt"
	"os"

	"github.com/joeblew999/infra/pkg/deck"
	"github.com/spf13/cobra"
)

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Clean all built deck tools",
	Long:  "Remove all compiled deck binaries and WASM modules",
	Run: func(cmd *cobra.Command, args []string) {
		manager := deck.NewManager()
		
		if err := manager.Clean(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		
		fmt.Println("Deck tools cleaned successfully!")
	},
}

func init() {
	BuildCmd.AddCommand(cleanCmd)
}