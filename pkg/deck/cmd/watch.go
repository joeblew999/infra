package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/joeblew999/infra/pkg/deck"
	"github.com/spf13/cobra"
)

// WatchCmd represents the watch command
var WatchCmd = &cobra.Command{
	Use:   "watch [path...]",
	Short: "Watch .dsh files and auto-process to multiple formats",
	Long: `Watch directories for .dsh file changes and automatically
process them through the pipeline: .dsh → XML → SVG/PNG/PDF

Examples:
  deck watch ./slides/
  deck watch ./examples/ ./templates/ --formats=svg,png
  deck watch . --formats=svg,pdf`,
	Args: cobra.MinimumNArgs(1),
	RunE: runWatch,
}

func runWatch(cmd *cobra.Command, args []string) error {
	watcher := deck.NewWatcher()
	
	// Parse formats flag
	formatsFlag, _ := cmd.Flags().GetString("formats")
	if formatsFlag != "" {
		formats := strings.Split(formatsFlag, ",")
		for i, format := range formats {
			formats[i] = strings.TrimSpace(format)
		}
		watcher.SetFormats(formats)
	}
	
	for _, path := range args {
		absPath, err := filepath.Abs(path)
		if err != nil {
			return fmt.Errorf("error resolving path %s: %w", path, err)
		}
		
		if _, err := os.Stat(absPath); os.IsNotExist(err) {
			return fmt.Errorf("path does not exist: %s", absPath)
		}
		
		watcher.AddPath(absPath)
		fmt.Printf("Watching: %s\n", absPath)
	}
	
	fmt.Printf("\nStarting .dsh file watcher...\n")
	fmt.Printf("Pipeline: .dsh → XML → %s\n", strings.Join(watcher.Formats, "/"))
	fmt.Printf("Output: %s\n", watcher.OutputDir)
	fmt.Printf("Press Ctrl+C to stop\n\n")
	
	return watcher.Start()
}

// Removed init() - command is registered in root.go