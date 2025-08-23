package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/joeblew999/infra/pkg/deck"
)

// RenderPDFCmd converts decksh files to PDF
func RenderPDFCmd(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: render-pdf <file.dsh> [output.pdf]")
	}

	inputFile := args[0]
	
	// Determine output file
	var outputFile string
	if len(args) > 1 {
		outputFile = args[1]
	} else {
		// Default: replace .dsh with .pdf
		ext := filepath.Ext(inputFile)
		if ext == ".dsh" {
			outputFile = strings.TrimSuffix(inputFile, ext) + ".pdf"
		} else {
			outputFile = inputFile + ".pdf"
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

	// Convert to PDF
	pdfBytes, err := renderer.DeckshToPDF(string(content), opts)
	if err != nil {
		return fmt.Errorf("failed to render PDF: %w", err)
	}

	// Write output
	if outputFile == "-" {
		_, err = os.Stdout.Write(pdfBytes)
		if err != nil {
			return fmt.Errorf("failed to write output: %w", err)
		}
	} else {
		err = os.WriteFile(outputFile, pdfBytes, 0644)
		if err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}
		fmt.Printf("Rendered %s -> %s\n", inputFile, outputFile)
	}

	return nil
}