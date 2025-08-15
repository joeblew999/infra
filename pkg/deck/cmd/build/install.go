package build

import (
	"fmt"
	"os"

	"github.com/joeblew999/infra/pkg/deck"
	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install [tool]",
	Short: "Install deck tools from source",
	Long: `Install deck tools by downloading source code and compiling to native binaries and WASM.

Without arguments, installs all tools. With a tool name, installs only that tool.

Available tools: decksh, svgdeck, dshfmt, dshlint, pngdeck, pdfdeck`,
	Run: func(cmd *cobra.Command, args []string) {
		manager := deck.NewManager()
		
		var err error
		if len(args) == 0 {
			err = manager.Install()
		} else {
			err = manager.InstallTool(args[0])
		}
		
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		
		fmt.Println("Deck tools installed successfully!")
	},
}

func init() {
	BuildCmd.AddCommand(installCmd)
}