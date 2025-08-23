package cmd

import (
	"fmt"
	"os"
)

// GenerateDemo creates a demo SVG file for testing
func GenerateDemo() error {
	inputFile := "../test/simple.dsh"
	outputFile := "../test/demo.svg"
	
	err := RenderCmd([]string{inputFile, outputFile})
	if err != nil {
		return fmt.Errorf("failed to generate demo: %w", err)
	}
	
	fmt.Printf("Demo SVG generated: %s\n", outputFile)
	
	// Also show the content
	content, err := os.ReadFile(outputFile)
	if err != nil {
		return fmt.Errorf("failed to read demo file: %w", err)
	}
	
	fmt.Printf("\nSVG Content:\n%s\n", string(content))
	return nil
}