package cmd

import (
	"os"
	"testing"
)

func TestRenderCmd(t *testing.T) {
	// Test rendering the simple.dsh file
	inputFile := "../test/simple.dsh"
	outputFile := "../test/simple.svg"
	
	// Clean up any existing output
	os.Remove(outputFile)
	
	// Run render command
	err := RenderCmd([]string{inputFile, outputFile})
	if err != nil {
		t.Fatalf("RenderCmd failed: %v", err)
	}
	
	// Check output file was created
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Error("Output SVG file was not created")
	}
	
	// Check file has content
	info, err := os.Stat(outputFile)
	if err != nil {
		t.Fatalf("Failed to stat output file: %v", err)
	}
	
	if info.Size() < 100 {
		t.Error("Output file seems too small")
	}
	
	// Clean up
	os.Remove(outputFile)
}

func TestRenderCmdDefaultOutput(t *testing.T) {
	// Test with default output filename
	inputFile := "../test/simple.dsh"
	expectedOutput := "../test/simple.svg"
	
	// Clean up
	os.Remove(expectedOutput)
	
	err := RenderCmd([]string{inputFile})
	if err != nil {
		t.Fatalf("RenderCmd with default output failed: %v", err)
	}
	
	// Check output file was created
	if _, err := os.Stat(expectedOutput); os.IsNotExist(err) {
		t.Error("Default output SVG file was not created")
	}
	
	// Clean up
	os.Remove(expectedOutput)
}