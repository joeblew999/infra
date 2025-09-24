package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/joeblew999/infra/pkg/config"
	"github.com/joeblew999/infra/pkg/hugo"
)

// Register mounts the hugo command under the provided parent.
func Register(parent *cobra.Command) {
	parent.AddCommand(GetHugoCmd())
}

var hugoCmd = &cobra.Command{
	Use:   "hugo",
	Short: "Hugo static site generator commands",
	Long:  "Manage Hugo documentation site generation and development server",
}

var hugoInitCmd = &cobra.Command{
	Use:   "init [directory]",
	Short: "Initialize Hugo site with HugoPlate theme",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		docsDir := "docs-hugo"
		if len(args) > 0 {
			docsDir = args[0]
		}

		// Ensure directory exists
		if err := os.MkdirAll(docsDir, 0755); err != nil {
			return fmt.Errorf("failed to create docs directory: %w", err)
		}

		// Initialize Hugo service and theme
		service := hugo.NewService(true, docsDir)
		themeManager := hugo.NewThemeManager(docsDir)

		// Initialize site structure
		if err := service.Start(context.Background()); err != nil {
			return fmt.Errorf("failed to initialize Hugo site: %w", err)
		}

		// Install HugoPlate theme
		if err := themeManager.InstallTheme(); err != nil {
			return fmt.Errorf("failed to install HugoPlate theme: %w", err)
		}

		// Migrate existing docs if they exist
		existingDocs := "docs"
		if _, err := os.Stat(existingDocs); err == nil {
			fmt.Printf("Migrating existing docs from %s...\n", existingDocs)
			if err := themeManager.MigrateContent(existingDocs); err != nil {
				fmt.Printf("Warning: failed to migrate some content: %v\n", err)
			}
		}

		fmt.Printf("✅ Hugo site initialized in %s\n", docsDir)
		fmt.Printf("📁 Theme: HugoPlate (modern, responsive)\n")
		fmt.Printf("🚀 Start development server: go run . hugo serve\n")
		fmt.Printf("🔧 Configure: edit %s/hugo.yaml\n", docsDir)

		return nil
	},
}

var hugoServeCmd = &cobra.Command{
	Use:   "serve [directory]",
	Short: "Start Hugo development server",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		docsDir := "docs-hugo"
		if len(args) > 0 {
			docsDir = args[0]
		}

		// Check if Hugo site exists
		configFile := filepath.Join(docsDir, "hugo.yaml")
		if _, err := os.Stat(configFile); os.IsNotExist(err) {
			return fmt.Errorf("Hugo site not found in %s. Run 'hugo init' first", docsDir)
		}

		service := hugo.NewService(true, docsDir)
		cfg := config.GetConfig()

		fmt.Printf("🚀 Starting Hugo development server...\n")
		fmt.Printf("📂 Source: %s\n", docsDir)
		fmt.Printf("🌐 URL: %s\n", config.FormatLocalHTTP(cfg.Ports.Hugo))
		fmt.Printf("📝 Live reload enabled\n")

		return service.Start(context.Background())
	},
}

var hugoBuildCmd = &cobra.Command{
	Use:   "build [directory]",
	Short: "Build static site for production",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		docsDir := "docs-hugo"
		if len(args) > 0 {
			docsDir = args[0]
		}

		service := hugo.NewService(false, docsDir)

		fmt.Printf("🔨 Building static site...\n")
		fmt.Printf("📂 Source: %s\n", docsDir)
		fmt.Printf("📦 Output: %s\n", service.GetOutputDir())

		if err := service.Start(context.Background()); err != nil {
			return fmt.Errorf("build failed: %w", err)
		}

		fmt.Printf("✅ Site built successfully\n")
		return nil
	},
}

var hugoThemeCmd = &cobra.Command{
	Use:   "theme",
	Short: "Theme management commands",
}

var hugoThemeInstallCmd = &cobra.Command{
	Use:   "install [directory]",
	Short: "Install or reinstall HugoPlate theme",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		docsDir := "docs-hugo"
		if len(args) > 0 {
			docsDir = args[0]
		}

		themeManager := hugo.NewThemeManager(docsDir)

		fmt.Printf("📥 Installing HugoPlate theme...\n")
		if err := themeManager.InstallTheme(); err != nil {
			return fmt.Errorf("theme installation failed: %w", err)
		}

		fmt.Printf("✅ HugoPlate theme installed\n")
		return nil
	},
}

// GetHugoCmd returns the hugo command for CLI integration
func GetHugoCmd() *cobra.Command {
	return hugoCmd
}

func init() {
	hugoCmd.AddCommand(hugoInitCmd)
	hugoCmd.AddCommand(hugoServeCmd)
	hugoCmd.AddCommand(hugoBuildCmd)
	hugoCmd.AddCommand(hugoThemeCmd)

	hugoThemeCmd.AddCommand(hugoThemeInstallCmd)
}
