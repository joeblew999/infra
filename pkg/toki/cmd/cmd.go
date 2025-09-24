package cmd

import (
	"fmt"

	"github.com/joeblew999/infra/pkg/toki"
	"github.com/spf13/cobra"
)

// Register mounts the toki command under the provided parent.
func Register(parent *cobra.Command) {
	parent.AddCommand(GetTokiCmd())
}

// GetTokiCmd returns the toki command for CLI integration
func GetTokiCmd() *cobra.Command {
	return tokiCmd
}

var tokiCmd = &cobra.Command{
	Use:   "toki",
	Short: "Run toki translation commands",
	Long:  `Run toki translation commands directly via CLI.`,
}

var tokiGenerateCmd = &cobra.Command{
	Use:   "generate [module-path] [source-lang] [target-langs...]",
	Short: "Run toki generate command",
	Args:  cobra.MinimumNArgs(2),
	RunE:  runTokiGenerate,
}

var tokiLintCmd = &cobra.Command{
	Use:   "lint [module-path]",
	Short: "Run toki lint command",
	Args:  cobra.ExactArgs(1),
	RunE:  runTokiLint,
}

var tokiWebEditCmd = &cobra.Command{
	Use:   "webedit [module-path]",
	Short: "Run toki webedit command",
	Args:  cobra.ExactArgs(1),
	RunE:  runTokiWebEdit,
}

var tokiVersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show toki version",
	RunE:  showTokiVersion,
}

func init() {
	// Build toki command structure
	tokiCmd.AddCommand(tokiGenerateCmd)
	tokiCmd.AddCommand(tokiLintCmd)
	tokiCmd.AddCommand(tokiWebEditCmd)
	tokiCmd.AddCommand(tokiVersionCmd)
}

func runTokiGenerate(cmd *cobra.Command, args []string) error {
	modulePath := args[0]
	sourceLang := args[1]
	targetLangs := args[2:]
	
	if len(targetLangs) == 0 {
		targetLangs = []string{"es"} // Default to Spanish
	}
	
	runner, err := toki.New()
	if err != nil {
		return fmt.Errorf("failed to create toki runner: %w", err)
	}
	
	fmt.Printf("Running toki generate for module: %s\n", modulePath)
	fmt.Printf("Source language: %s\n", sourceLang)
	fmt.Printf("Target languages: %v\n", targetLangs)
	
	return runner.Generate(modulePath, sourceLang, targetLangs)
}

func runTokiLint(cmd *cobra.Command, args []string) error {
	modulePath := args[0]
	
	runner, err := toki.New()
	if err != nil {
		return fmt.Errorf("failed to create toki runner: %w", err)
	}
	
	fmt.Printf("Running toki lint for module: %s\n", modulePath)
	return runner.Lint(modulePath)
}

func runTokiWebEdit(cmd *cobra.Command, args []string) error {
	modulePath := args[0]
	
	runner, err := toki.New()
	if err != nil {
		return fmt.Errorf("failed to create toki runner: %w", err)
	}
	
	fmt.Printf("Starting toki webedit for module: %s\n", modulePath)
	return runner.WebEdit(modulePath)
}

func showTokiVersion(cmd *cobra.Command, args []string) error {
	runner, err := toki.New()
	if err != nil {
		return fmt.Errorf("failed to create toki runner: %w", err)
	}
	
	version, err := runner.Version()
	if err != nil {
		return fmt.Errorf("failed to get version: %w", err)
	}
	
	fmt.Printf("Toki version: %s\n", version)
	return nil
}
