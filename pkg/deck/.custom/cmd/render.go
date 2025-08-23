package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/joeblew999/infra/pkg/deck"
)

// RenderCmd converts decksh files to SVG
func RenderCmd(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: render <file.dsh> [output.svg]")
	}

	inputFile := args[0]
	
	// Determine output file
	var outputFile string
	if len(args) > 1 {
		outputFile = args[1]
	} else {
		// Default: replace .dsh with .svg
		ext := filepath.Ext(inputFile)
		if ext == ".dsh" {
			outputFile = strings.TrimSuffix(inputFile, ext) + ".svg"
		} else {
			outputFile = inputFile + ".svg"
		}
	}

	// Read input file
	var input io.Reader
	if inputFile == "-" {
		input = os.Stdin
	} else {
		file, err := os.Open(inputFile)
		if err != nil {
			return fmt.Errorf("failed to open input file: %w", err)
		}
		defer file.Close()
		input = file
	}

	// Read the decksh content
	content, err := io.ReadAll(input)
	if err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}

	// Create renderer
	renderer := deck.NewDefaultRenderer()
	opts := deck.DefaultRenderOptions()
	opts.Title = filepath.Base(inputFile)

	// Convert to SVG
	svg, err := renderer.DeckshToSVG(string(content), opts)
	if err != nil {
		return fmt.Errorf("failed to render SVG: %w", err)
	}

	// Write output
	var output io.Writer
	if outputFile == "-" {
		output = os.Stdout
	} else {
		file, err := os.Create(outputFile)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer file.Close()
		output = file
	}

	_, err = output.Write([]byte(svg))
	if err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	if outputFile != "-" {
		fmt.Printf("Rendered %s -> %s\n", inputFile, outputFile)
	}

	return nil
}