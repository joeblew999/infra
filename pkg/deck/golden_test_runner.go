package deck

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// GoldenTest represents a single golden test case
type GoldenTest struct {
	Name     string `json:"name"`
	Category string `json:"category"`
	Input    struct {
		Dsh string `json:"dsh"`
	} `json:"input"`
	Outputs map[string]string `json:"outputs"`
}

// GoldenTestCatalog represents the JSON catalog structure
type GoldenTestCatalog struct {
	Version     string        `json:"version"`
	Description string        `json:"description"`
	Generated   string        `json:"generated"`
	SourceBase  string        `json:"source_base"`
	TotalTests  int           `json:"total_tests"`
	TestCases   []GoldenTest  `json:"test_cases"`
}

// GoldenTestRunner runs automated golden tests
type GoldenTestRunner struct {
	sourceDir   string
	buildDir    string
	outputDir   string
	goldenTests []GoldenTest
}

// NewGoldenTestRunner creates a new golden test runner
func NewGoldenTestRunner(sourceDir, buildDir string) (*GoldenTestRunner, error) {
	// Convert to absolute paths
	absSourceDir, err := filepath.Abs(sourceDir)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve source dir: %w", err)
	}
	
	absBuildDir, err := filepath.Abs(buildDir)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve build dir: %w", err)
	}

	// Create output directory path
	outputDir := filepath.Join(filepath.Dir(absSourceDir), ".golden-test-output")
	
	runner := &GoldenTestRunner{
		sourceDir: absSourceDir,
		buildDir:  absBuildDir,
		outputDir: outputDir,
	}

	// Load golden tests from JSON
	testsFile := filepath.Join(absSourceDir, "..", "golden_tests.json")
	data, err := os.ReadFile(testsFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read golden tests: %w", err)
	}

	var catalog GoldenTestCatalog
	if err := json.Unmarshal(data, &catalog); err != nil {
		return nil, fmt.Errorf("failed to parse golden tests: %w", err)
	}
	runner.goldenTests = catalog.TestCases

	return runner, nil
}

// TestResult represents the result of a single test
type TestResult struct {
	Name       string
	Category   string
	Passed     bool
	XMLPassed  bool
	SVGPassed  bool
	Errors     []string
}

// RunTest runs a single golden test case with proper comparison
func (r *GoldenTestRunner) RunTest(test GoldenTest) (*TestResult, error) {
	result := &TestResult{
		Name:     test.Name,
		Category: test.Category,
		Passed:   true,
		Errors:   []string{},
	}

	fmt.Printf("Running test: %s (%s)\n", test.Name, test.Category)

	// Build full path to DSH file
	dshPath := filepath.Join(r.sourceDir, test.Input.Dsh)
	if _, err := os.Stat(dshPath); os.IsNotExist(err) {
		result.Passed = false
		result.Errors = append(result.Errors, fmt.Sprintf("DSH file not found: %s", dshPath))
		return result, nil
	}

	// Create mirrored output directory structure
	testRelativeDir := filepath.Dir(test.Input.Dsh)
	outputTestDir := filepath.Join(r.outputDir, testRelativeDir)
	if err := os.MkdirAll(outputTestDir, 0755); err != nil {
		result.Passed = false
		result.Errors = append(result.Errors, fmt.Sprintf("failed to create output directory: %v", err))
		return result, nil
	}

	// Get DSH file name without extension for generating output names
	dshFile := filepath.Base(dshPath)
	baseName := strings.TrimSuffix(dshFile, ".dsh")

	// Step 1: DSH → XML comparison
	if err := r.compareXMLGeneration(test, dshPath, outputTestDir, baseName, result); err != nil {
		return result, err
	}

	// Step 2: XML → SVG comparison (only if XML stage passed)
	if result.XMLPassed {
		if err := r.compareSVGGeneration(test, outputTestDir, baseName, result); err != nil {
			return result, err
		}
	}

	// Overall result
	result.Passed = result.XMLPassed && result.SVGPassed

	if result.Passed {
		fmt.Printf("  ✓ Test passed (XML: ✓, SVG: ✓)\n")
	} else {
		fmt.Printf("  ✗ Test failed (XML: %s, SVG: %s)\n", 
			boolToStatus(result.XMLPassed), boolToStatus(result.SVGPassed))
		for _, err := range result.Errors {
			fmt.Printf("    - %s\n", err)
		}
	}

	return result, nil
}

// boolToStatus converts boolean to status symbol
func boolToStatus(b bool) string {
	if b {
		return "✓"
	}
	return "✗"
}

// compareXMLGeneration runs DSH → XML pipeline stage and compares result
func (r *GoldenTestRunner) compareXMLGeneration(test GoldenTest, dshPath, outputTestDir, baseName string, result *TestResult) error {
	// Check if golden XML exists
	goldenXMLPath := filepath.Join(r.sourceDir, filepath.Dir(test.Input.Dsh), baseName+".xml")
	if _, err := os.Stat(goldenXMLPath); os.IsNotExist(err) {
		result.XMLPassed = false
		result.Errors = append(result.Errors, fmt.Sprintf("Golden XML not found: %s", goldenXMLPath))
		return nil
	}

	// Generate XML from DSH
	outputXMLPath := filepath.Join(outputTestDir, baseName+".xml")
	deckshPath := filepath.Join(r.buildDir, "bin", "decksh")
	
	cmd := exec.Command(deckshPath, "-o", outputXMLPath, dshPath)
	if err := cmd.Run(); err != nil {
		result.XMLPassed = false
		result.Errors = append(result.Errors, fmt.Sprintf("Failed to generate XML: %v", err))
		return nil
	}

	// Compare generated XML with golden XML
	if equal, err := r.compareFiles(outputXMLPath, goldenXMLPath); err != nil {
		result.XMLPassed = false
		result.Errors = append(result.Errors, fmt.Sprintf("Failed to compare XML files: %v", err))
		return nil
	} else if !equal {
		result.XMLPassed = false
		result.Errors = append(result.Errors, "Generated XML differs from golden XML")
		return nil
	}

	result.XMLPassed = true
	return nil
}

// compareSVGGeneration runs XML → SVG pipeline stage and compares result
func (r *GoldenTestRunner) compareSVGGeneration(test GoldenTest, outputTestDir, baseName string, result *TestResult) error {
	// Check if golden SVG exists
	goldenSVGPath := filepath.Join(r.sourceDir, filepath.Dir(test.Input.Dsh), baseName+"-00001.svg")
	if _, err := os.Stat(goldenSVGPath); os.IsNotExist(err) {
		result.SVGPassed = false
		result.Errors = append(result.Errors, fmt.Sprintf("Golden SVG not found: %s", goldenSVGPath))
		return nil
	}

	// Generate SVG from XML
	xmlPath := filepath.Join(outputTestDir, baseName+".xml")
	decksvgPath := filepath.Join(r.buildDir, "bin", "decksvg")
	
	cmd := exec.Command(decksvgPath, xmlPath)
	cmd.Dir = outputTestDir
	if err := cmd.Run(); err != nil {
		result.SVGPassed = false
		result.Errors = append(result.Errors, fmt.Sprintf("Failed to generate SVG: %v", err))
		return nil
	}

	// Compare generated SVG with golden SVG
	outputSVGPath := filepath.Join(outputTestDir, baseName+"-00001.svg")
	if equal, err := r.compareFiles(outputSVGPath, goldenSVGPath); err != nil {
		result.SVGPassed = false
		result.Errors = append(result.Errors, fmt.Sprintf("Failed to compare SVG files: %v", err))
		return nil
	} else if !equal {
		result.SVGPassed = false
		result.Errors = append(result.Errors, "Generated SVG differs from golden SVG")
		return nil
	}

	result.SVGPassed = true
	return nil
}

// compareFiles does byte-for-byte comparison of two files
func (r *GoldenTestRunner) compareFiles(file1, file2 string) (bool, error) {
	data1, err := os.ReadFile(file1)
	if err != nil {
		return false, fmt.Errorf("failed to read %s: %w", file1, err)
	}

	data2, err := os.ReadFile(file2)
	if err != nil {
		return false, fmt.Errorf("failed to read %s: %w", file2, err)
	}

	return string(data1) == string(data2), nil
}

// RunAllTests runs all golden tests
func (r *GoldenTestRunner) RunAllTests() error {
	fmt.Printf("Running %d golden tests...\n\n", len(r.goldenTests))

	passed := 0
	failed := 0
	xmlPassed := 0
	svgPassed := 0

	for _, test := range r.goldenTests {
		result, err := r.RunTest(test)
		if err != nil {
			fmt.Printf("  ✗ ERROR: %v\n\n", err)
			failed++
			continue
		}

		if result.Passed {
			passed++
		} else {
			failed++
		}
		
		if result.XMLPassed {
			xmlPassed++
		}
		if result.SVGPassed {
			svgPassed++
		}
	}

	fmt.Printf("\nResults Summary:\n")
	fmt.Printf("Overall: %d passed, %d failed\n", passed, failed)
	fmt.Printf("XML Pipeline: %d passed, %d failed\n", xmlPassed, len(r.goldenTests)-xmlPassed)
	fmt.Printf("SVG Pipeline: %d passed, %d failed\n", svgPassed, len(r.goldenTests)-svgPassed)
	
	if failed > 0 {
		return fmt.Errorf("%d tests failed", failed)
	}

	return nil
}

// RunTestsInCategory runs tests for a specific category
func (r *GoldenTestRunner) RunTestsInCategory(category string) error {
	var categoryTests []GoldenTest
	for _, test := range r.goldenTests {
		if test.Category == category {
			categoryTests = append(categoryTests, test)
		}
	}

	if len(categoryTests) == 0 {
		return fmt.Errorf("no tests found for category: %s", category)
	}

	fmt.Printf("Running %d tests in category '%s'...\n\n", len(categoryTests), category)

	passed := 0
	failed := 0
	xmlPassed := 0
	svgPassed := 0

	for _, test := range categoryTests {
		result, err := r.RunTest(test)
		if err != nil {
			fmt.Printf("  ✗ ERROR: %v\n\n", err)
			failed++
			continue
		}

		if result.Passed {
			passed++
		} else {
			failed++
		}
		
		if result.XMLPassed {
			xmlPassed++
		}
		if result.SVGPassed {
			svgPassed++
		}
	}

	fmt.Printf("\nResults for '%s':\n", category)
	fmt.Printf("Overall: %d passed, %d failed\n", passed, failed)
	fmt.Printf("XML Pipeline: %d passed, %d failed\n", xmlPassed, len(categoryTests)-xmlPassed)
	fmt.Printf("SVG Pipeline: %d passed, %d failed\n", svgPassed, len(categoryTests)-svgPassed)
	
	if failed > 0 {
		return fmt.Errorf("%d tests failed in category %s", failed, category)
	}

	return nil
}

// CleanupTestOutputs removes all test output files
func (r *GoldenTestRunner) CleanupTestOutputs() error {
	if _, err := os.Stat(r.outputDir); os.IsNotExist(err) {
		fmt.Printf("No test output directory to clean: %s\n", r.outputDir)
		return nil
	}

	if err := os.RemoveAll(r.outputDir); err != nil {
		return fmt.Errorf("failed to remove test output directory: %w", err)
	}

	fmt.Printf("Cleaned up test output directory: %s\n", r.outputDir)
	return nil
}