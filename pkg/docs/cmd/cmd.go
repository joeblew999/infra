package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/joeblew999/infra/pkg/deck"
)

// Register mounts the docs command under the provided parent.
func Register(parent *cobra.Command) {
	parent.AddCommand(GetDocsCmd())
}

var docsCmd = &cobra.Command{
	Use:   "docs",
	Short: "Documentation utilities",
	Long:  "Tools for processing and converting documentation files",
}

var docsToPdfCmd = &cobra.Command{
	Use:   "to-pdf [markdown-file]",
	Short: "Convert Markdown files to multiple formats using deck",
	Long:  "Convert business documentation from Markdown to SVG/PNG/PDF using Deck's native markdown support",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		markdownFile := args[0]
		
		// Check if file exists
		if _, err := os.Stat(markdownFile); os.IsNotExist(err) {
			return fmt.Errorf("markdown file not found: %s", markdownFile)
		}

		// Use deck's native markdown support for conversion
		baseName := strings.TrimSuffix(markdownFile, filepath.Ext(markdownFile))
		
		// Create a simple DSH file that uses deck's native markdown:// support
		dshContent := fmt.Sprintf(`deck
slide
	text "Documentation" 50 95 4 "sans" "black"
	content "markdown://%s" 50 50 1.2
eslide
edeck`, markdownFile)
		
		dshFile := baseName + ".dsh"
		if err := os.WriteFile(dshFile, []byte(dshContent), 0644); err != nil {
			return fmt.Errorf("failed to create DSH file: %w", err)
		}
		defer os.Remove(dshFile) // Clean up temporary DSH file
		
		// Create a watcher to process the DSH file
		watcher := deck.NewWatcher()
		watcher.SetFormats([]string{"svg", "png", "pdf"})
		
		// Process the single DSH file directly
		if err := os.MkdirAll(watcher.OutputDir, 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}
		
		// Use the watcher's processing logic for our single file
		watcher.ProcessDSHFile(dshFile)
		
		fmt.Printf("Successfully converted %s using deck's native markdown support\n", markdownFile)
		return nil
	},
}

var docsAllToPdfCmd = &cobra.Command{
	Use:   "all-to-pdf",
	Short: "Convert all business docs to multiple formats using deck",
	Long:  "Convert all Markdown files in docs/business/ to SVG/PNG/PDF using Deck's native markdown support",
	RunE: func(cmd *cobra.Command, args []string) error {
		docsDir := "docs/business"
		
		// Check if docs directory exists
		if _, err := os.Stat(docsDir); os.IsNotExist(err) {
			return fmt.Errorf("business docs directory not found: %s", docsDir)
		}

		// Use deck watcher for batch processing
		watcher := deck.NewWatcher()
		watcher.SetFormats([]string{"svg", "png", "pdf"})

		// Find all .md files
		err := filepath.Walk(docsDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if strings.HasSuffix(strings.ToLower(path), ".md") {
				fmt.Printf("Converting %s...\n", path)
				
				baseName := strings.TrimSuffix(path, filepath.Ext(path))
				
				// Create DSH file with native markdown support
				dshContent := fmt.Sprintf(`deck
slide
	text "Documentation" 50 95 4 "sans" "black"
	content "markdown://%s" 50 50 1.2
eslide
edeck`, path)
				
				dshFile := baseName + ".dsh"
				if err := os.WriteFile(dshFile, []byte(dshContent), 0644); err != nil {
					return fmt.Errorf("failed to create DSH file for %s: %w", path, err)
				}
				defer os.Remove(dshFile)
				
				// Process using watcher's DSH processing
				if err := os.MkdirAll(watcher.OutputDir, 0755); err != nil {
					return fmt.Errorf("failed to create output directory: %w", err)
				}
				
				watcher.ProcessDSHFile(dshFile)
				fmt.Printf("✓ Converted %s to multiple formats\n", path)
			}
			return nil
		})

		if err != nil {
			return fmt.Errorf("failed to process docs directory: %w", err)
		}

		fmt.Println("All business documentation converted to multiple formats using deck's native markdown support!")
		return nil
	},
}

// GetDocsCmd returns the docs command for CLI integration
func GetDocsCmd() *cobra.Command {
	return docsCmd
}

func init() {
	docsCmd.AddCommand(docsToPdfCmd)
	docsCmd.AddCommand(docsAllToPdfCmd)
}
