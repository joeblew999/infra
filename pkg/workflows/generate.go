package workflows

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/joeblew999/infra/pkg/gozero"
	"github.com/joeblew999/infra/pkg/log"
)

// GenerateGoZeroCode generates go-zero API code from an .api file.
// It uses the gozero.GoZeroRunner to execute the goctl command.
func GenerateGoZeroCode(ctx context.Context, apiFile, outputDir string) error {
	log.Info("Generating go-zero API code", "apiFile", apiFile, "outputDir", outputDir)

	// Ensure output directory exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory %s: %w", outputDir, err)
	}

	// Create a new gozero runner
	runner := gozero.NewGoZeroRunner(true) // true for debug output

	// Set the working directory to the service directory
	// goctl needs to run from the service directory to find the .api file correctly
	runner.SetWorkDir(outputDir)

	// Get the API filename (without path) for goctl
	apiFileName := filepath.Base(apiFile)

	// Generate the API code using just the filename and current directory
	if err := runner.ApiGenerate(apiFileName, "."); err != nil {
		return fmt.Errorf("failed to generate go-zero API code for %s: %w", apiFile, err)
	}

	log.Info("✅ Successfully generated go-zero API code.")
	return nil
}
