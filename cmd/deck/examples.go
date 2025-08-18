package cmd

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// Example represents a decksh example file
type Example struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	Category    string `json:"category"`
	Description string `json:"description"`
	Content     string `json:"content,omitempty"`
	Size        int64  `json:"size"`
}

// ExampleCategory groups examples by theme
type ExampleCategory struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Examples    []Example `json:"examples"`
	Count       int       `json:"count"`
}

// GetExampleCategories returns all examples organized by category
func GetExampleCategories() ([]ExampleCategory, error) {
	examplesDir := "../examples"
	
	// Check if examples directory exists
	if _, err := os.Stat(examplesDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("examples directory not found - run 'go run fetch.go' first")
	}

	categories := map[string]*ExampleCategory{
		"dubois": {
			Name:        "DuBois Data Portraits",
			Description: "Historical data visualizations in the style of W.E.B. Du Bois",
			Examples:    []Example{},
		},
		"deckviz": {
			Name:        "Modern Data Viz",
			Description: "Contemporary data visualization examples",
			Examples:    []Example{},
		},
		"test": {
			Name:        "Simple Examples",
			Description: "Basic examples for learning",
			Examples:    []Example{},
		},
	}

	// Walk through all .dsh files
	err := filepath.Walk(examplesDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !strings.HasSuffix(path, ".dsh") {
			return nil
		}

		// Determine category from path
		var category string
		relativePath := strings.TrimPrefix(path, examplesDir+"/")
		
		switch {
		case strings.HasPrefix(relativePath, "dubois-data-portraits"):
			category = "dubois"
		case strings.HasPrefix(relativePath, "deckviz"):
			category = "deckviz"  
		case strings.HasPrefix(relativePath, "test"):
			category = "test"
		default:
			category = "test" // default
		}

		// Create example entry
		example := Example{
			Name:        filepath.Base(path),
			Path:        relativePath,
			Category:    category,
			Description: generateDescription(relativePath),
			Size:        info.Size(),
		}

		// Add to appropriate category
		if cat, exists := categories[category]; exists {
			cat.Examples = append(cat.Examples, example)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to scan examples: %w", err)
	}

	// Convert map to slice and update counts
	result := []ExampleCategory{}
	for _, cat := range categories {
		if len(cat.Examples) > 0 {
			cat.Count = len(cat.Examples)
			result = append(result, *cat)
		}
	}

	return result, nil
}

// GetExampleContent loads the content of a specific example
func GetExampleContent(examplePath string) (string, error) {
	fullPath := filepath.Join("../examples", examplePath)
	
	// Security check - ensure path is within examples directory
	if !strings.HasPrefix(fullPath, "../examples/") {
		return "", fmt.Errorf("invalid example path")
	}

	content, err := os.ReadFile(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to read example: %w", err)
	}

	return string(content), nil
}

// GetFeaturedExamples returns a curated list of good starter examples
func GetFeaturedExamples() []Example {
	return []Example{
		{
			Name:        "nothing.dsh",
			Path:        "deckviz/ur/nothing.dsh",
			Category:    "deckviz",
			Description: "Simple text overlay - great for beginners",
		},
		{
			Name:        "simple.dsh", 
			Path:        "test/simple.dsh",
			Category:    "test",
			Description: "Basic shapes and text example",
		},
		{
			Name:        "conjugal.dsh",
			Path:        "dubois-data-portraits/plate53/conjugal.dsh", 
			Category:    "dubois",
			Description: "Complex data visualization with multiple elements",
		},
	}
}

// generateDescription creates a description based on the file path
func generateDescription(path string) string {
	parts := strings.Split(path, "/")
	
	switch {
	case strings.Contains(path, "dubois-data-portraits"):
		if len(parts) >= 2 {
			return fmt.Sprintf("DuBois %s data visualization", parts[1])
		}
		return "Historical data portrait"
		
	case strings.Contains(path, "deckviz"):
		if len(parts) >= 2 {
			return fmt.Sprintf("%s visualization example", strings.Title(parts[1]))
		}
		return "Data visualization example"
		
	case strings.Contains(path, "test"):
		return "Simple test example"
		
	default:
		return "Deck example"
	}
}