package dep

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/joeblew999/infra/pkg/config"
	"github.com/joeblew999/infra/pkg/log"
)

func CleanSystem() error {
	depPath := config.GetDepPath()

	fmt.Printf("🗑️  Removing .dep directory: %s\n", depPath)
	if err := os.RemoveAll(depPath); err != nil {
		log.Error("Failed to remove .dep directory", "path", depPath, "error", err)
		return fmt.Errorf("failed to remove .dep directory: %w", err)
	}
	fmt.Println("✅ .dep directory removed")

	collectionPath := filepath.Join(depPath, ".collection")
	fmt.Printf("🗑️  Collection directory: %s (already removed with .dep)\n", collectionPath)

	generatedFile := "pkg/config/binaries_gen.go"
	fmt.Printf("🗑️  Removing generated code: %s\n", generatedFile)
	if err := os.Remove(generatedFile); err != nil {
		if !os.IsNotExist(err) {
			log.Error("Failed to remove generated file", "file", generatedFile, "error", err)
			return fmt.Errorf("failed to remove generated file %s: %w", generatedFile, err)
		}
		fmt.Printf("📝 Generated file %s did not exist\n", generatedFile)
	} else {
		fmt.Println("✅ Generated code removed")
	}

	fmt.Println("🗑️  Cleaning Go build cache for dep system")
	fmt.Println("✅ Build cache cleanup deferred to Go")

	return nil
}
