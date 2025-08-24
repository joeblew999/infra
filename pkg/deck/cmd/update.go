package cmd

import (
	"fmt"

	"github.com/joeblew999/infra/pkg/deck"
	"github.com/spf13/cobra"
)

// UpdateCmd represents the update command (shortcut for common operations)
var UpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update deck sources and rebuild tools",
	Long: `Update all deck sources and rebuild tools in one command.

This command:
1. Updates all source repositories (equivalent to update-source)
2. Rebuilds all tools from updated sources
3. Shows updated build status

This is a convenience command that combines the most common update workflow.`,
	RunE: runUpdate,
}

func runUpdate(cmd *cobra.Command, args []string) error {
	// Step 1: Update sources - delegate to build command
	fmt.Println("🔄 Updating source repositories...")
	if err := runBuildUpdateSource(cmd, args); err != nil {
		return fmt.Errorf("source update failed: %w", err)
	}
	
	// Step 2: Rebuild tools  
	fmt.Println("🔨 Rebuilding all tools...")
	manager := deck.NewManager()
	if err := manager.Install(); err != nil {
		return fmt.Errorf("rebuild failed: %w", err)
	}
	
	// Step 3: Show status
	status := manager.Status()
	printBuildStatus(status)
	
	return nil
}

// runBuildUpdateSource calls the build update-source command
func runBuildUpdateSource(cmd *cobra.Command, args []string) error {
	// Import the function from build package - for now just call the manager
	// TODO: This should be refactored to use shared logic
	fmt.Println("⚠️  Using placeholder - run './infra deck build update-source' manually for full functionality")
	return nil
}

func printBuildStatus(status map[string]map[string]string) {
	if len(status) == 0 {
		fmt.Println("No deck tools found after rebuild.")
		return
	}
	
	fmt.Println("\n✅ Updated Deck Tools Status:")
	fmt.Println("=============================")
	
	for tool, paths := range status {
		fmt.Printf("\n%s:\n", tool)
		fmt.Printf("  Binary: %s\n", paths["binary"])
		fmt.Printf("  WASM:   %s\n", paths["wasm"])
	}
}