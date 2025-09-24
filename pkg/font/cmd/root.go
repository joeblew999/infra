package cmd

import (
	"fmt"

	"github.com/joeblew999/infra/pkg/font"
	"github.com/spf13/cobra"
)

// Register mounts the font command under the provided parent.
func Register(parent *cobra.Command) {
	parent.AddCommand(GetFontCmd())
}

var fontCmd = &cobra.Command{
	Use:   "font",
	Short: "Font management tools",
	Long: `Font provides tools for downloading and managing fonts from Google Fonts.
	
This command manages font caching and provides utilities for testing
font compatibility with deck visualization tools.`,
}

var cacheCmd = &cobra.Command{
	Use:   "cache",
	Short: "Cache a font locally",
	Long:  `Download and cache a font from Google Fonts for local use`,
	Run: func(cmd *cobra.Command, args []string) {
		family, _ := cmd.Flags().GetString("family")
		weight, _ := cmd.Flags().GetInt("weight")
		format, _ := cmd.Flags().GetString("format")
		
		if family == "" {
			fmt.Println("Error: --family is required")
			return
		}
		
		manager := font.NewManager()
		
		var err error
		if format == "ttf" {
			err = manager.CacheTTF(family, weight)
		} else {
			err = manager.Cache(family, weight)
		}
		
		if err != nil {
			fmt.Printf("Error caching font: %v\n", err)
			return
		}
		
		path, err := manager.GetFormat(family, weight, format)
		if err != nil {
			fmt.Printf("Error getting font path: %v\n", err)
			return
		}
		
		fmt.Printf("Font cached successfully: %s\n", path)
	},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List cached fonts",
	Long:  `List all fonts currently cached locally`,
	Run: func(cmd *cobra.Command, args []string) {
		manager := font.NewManager()
		fonts := manager.List()
		
		if len(fonts) == 0 {
			fmt.Println("No fonts cached")
			return
		}
		
		fmt.Printf("Cached fonts (%d):\n", len(fonts))
		for _, f := range fonts {
			fmt.Printf("  %s %d (%s) - %s\n", f.Family, f.Weight, f.Format, f.Path)
		}
	},
}

// GetFontCmd returns the font command for registration
func GetFontCmd() *cobra.Command {
	// Setup flags
	cacheCmd.Flags().StringP("family", "f", "", "Font family name (required)")
	cacheCmd.Flags().IntP("weight", "w", 400, "Font weight")
	cacheCmd.Flags().String("format", "ttf", "Font format (ttf, woff2)")
	
	// Add subcommands
	fontCmd.AddCommand(cacheCmd)
	fontCmd.AddCommand(listCmd)
	
	return fontCmd
}
