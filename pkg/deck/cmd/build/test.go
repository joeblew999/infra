package build

import (
	"github.com/joeblew999/infra/pkg/deck"
	"github.com/spf13/cobra"
)

// testCmd represents the test command
var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Run golden tests for deck functionality",
	Long: `Run automated golden tests using the binary pipeline.

This command runs tests from the golden_tests.json catalog, verifying that
the deck binary pipeline (decksh → XML → [decksvg|deckpng|deckpdf]) produces
expected outputs for known good input files.`,
}

var testAllCmd = &cobra.Command{
	Use:   "all",
	Short: "Run all golden tests",
	RunE: func(cmd *cobra.Command, args []string) error {
		buildDir := deck.BuildRoot

		runner, err := deck.NewGoldenTestRunner(buildDir)
		if err != nil {
			return err
		}

		return runner.RunAllTests()
	},
}

var testCategoryCmd = &cobra.Command{
	Use:   "category [category-name]",
	Short: "Run golden tests for a specific category",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		category := args[0]
		buildDir := deck.BuildRoot

		runner, err := deck.NewGoldenTestRunner(buildDir)
		if err != nil {
			return err
		}

		return runner.RunTestsInCategory(category)
	},
}

var testCleanupCmd = &cobra.Command{
	Use:   "cleanup",
	Short: "Clean up golden test output files",
	RunE: func(cmd *cobra.Command, args []string) error {
		buildDir := deck.BuildRoot

		runner, err := deck.NewGoldenTestRunner(buildDir)
		if err != nil {
			return err
		}

		return runner.CleanupTestOutputs()
	},
}

func init() {
	testCmd.AddCommand(testAllCmd)
	testCmd.AddCommand(testCategoryCmd)
	testCmd.AddCommand(testCleanupCmd)
}