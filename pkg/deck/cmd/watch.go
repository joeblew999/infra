package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/joeblew999/infra/pkg/deck"
	"github.com/spf13/cobra"
)

// WatchCmd represents the watch command
var WatchCmd = &cobra.Command{
	Use:   "watch [path...]",
	Short: "Watch .dsh files and auto-process to XML/SVG",
	Long: `Watch directories for .dsh file changes and automatically
process them through the pipeline: .dsh → XML → SVG

Examples:
  deck watch ./slides/
  deck watch ./examples/ ./templates/
  deck watch .`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		watcher := deck.NewWatcher()
		
		for _, path := range args {
			absPath, err := filepath.Abs(path)
			if err != nil {
				fmt.Printf("Error resolving path %s: %v\n", path, err)
				os.Exit(1)
			}
			
			if _, err := os.Stat(absPath); os.IsNotExist(err) {
				fmt.Printf("Path does not exist: %s\n", absPath)
				os.Exit(1)
			}
			
			watcher.AddPath(absPath)
			fmt.Printf("Watching: %s\n", absPath)
		}
		
		fmt.Printf("\nStarting .dsh file watcher...\n")
		fmt.Printf("Pipeline: .dsh → XML → SVG\n")
		fmt.Printf("Output: %s\n", watcher.OutputDir)
		fmt.Printf("Press Ctrl+C to stop\n\n")
		
		if err := watcher.Start(); err != nil {
			fmt.Printf("Error starting watcher: %v\n", err)
			os.Exit(1)
		}
	},
}

// Removed init() - command is registered in root.go