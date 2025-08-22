package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/joeblew999/infra/pkg/deck/cmd"
	"github.com/spf13/cobra"
)

// RunDeck registers deck commands
func RunDeck() {
	rootCmd.AddCommand(NewDeckCmd())
}

// AddDeckToCLI adds deck commands to the CLI namespace
func AddDeckToCLI(cliParent *cobra.Command) {
	cliParent.AddCommand(NewDeckCmd())
}

func NewDeckCmd() *cobra.Command {
	deckCmd := &cobra.Command{
		Use:   "deck",
		Short: "Deck presentation generator",
		Long:  `Commands for generating presentations from decksh markup to various formats (SVG, PNG, PDF)`,
	}

	// Add deck subcommands
	deckCmd.AddCommand(
		newRenderCmd(),
		newServeCmd(),
		newDemoCmd(),
		newExamplesCmd(),
	)

	return deckCmd
}

func newRenderCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "render <input.dsh> [output.svg]",
		Short: "Convert decksh files to SVG",
		Long:  "Convert decksh markup files to SVG format",
		RunE: func(c *cobra.Command, args []string) error {
			return cmd.RenderCmd(args)
		},
	}
}

func newServeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "serve",
		Short: "Start deck development server",
		Long:  "Start a development server for deck presentations",
		RunE: func(c *cobra.Command, args []string) error {
			return cmd.ServeCmd(args)
		},
	}
}

func newDemoCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "demo",
		Short: "Generate demo presentations",
		Long:  "Generate example presentations for testing and demonstration",
		RunE: func(c *cobra.Command, args []string) error {
			return cmd.GenerateDemo()
		},
	}
}

func newExamplesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "examples",
		Short: "List and show deck examples",
		Long:  "Display available deck presentation examples",
		RunE: func(c *cobra.Command, args []string) error {
			categories, err := cmd.GetExampleCategories()
			if err != nil {
				return fmt.Errorf("failed to get examples: %w", err)
			}
			
			// Pretty print the examples
			encoder := json.NewEncoder(os.Stdout)
			encoder.SetIndent("", "  ")
			return encoder.Encode(categories)
		},
	}
}