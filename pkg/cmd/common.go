package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/joeblew999/infra/pkg/log"
	"github.com/joeblew999/infra/pkg/config"
)

func EnsureInfraDirectories() error {
	// Create .dep directory
	if err := os.MkdirAll(config.GetDepPath(), 0755); err != nil {
		return fmt.Errorf("failed to create .dep directory: %w", err)
	}
	log.Info("Ensured directory exists", "path", config.GetDepPath())

	// Create .bin directory
	if err := os.MkdirAll(config.GetBinPath(), 0755); err != nil {
		return fmt.Errorf("failed to create .bin directory: %w", err)
	}
	log.Info("Ensured directory exists", "path", config.GetBinPath())

	// Create .data directory
	if err := os.MkdirAll(config.GetDataPath(), 0755); err != nil {
		return fmt.Errorf("failed to create .data directory: %w", err)
	}
	log.Info("Ensured directory exists", "path", config.GetDataPath())

	// Create taskfiles directory
	if err := os.MkdirAll(config.GetTaskfilesPath(), 0755); err != nil {
		return fmt.Errorf("failed to create taskfiles directory: %w", err)
	}
	log.Info("Ensured directory exists", "path", config.GetTaskfilesPath())

	return nil
}

func ExecuteBinary(binary string, args ...string) error {
	// Save current working directory
	oldDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory: %w", err)
	}

	// Get absolute path of the binary before changing directory
	absoluteBinaryPath, err := filepath.Abs(binary)
	if err != nil {
		return fmt.Errorf("failed to get absolute path of binary: %w", err)
	}

	// Change to the terraform directory
	if err := os.Chdir(config.GetTerraformPath()); err != nil {
		return fmt.Errorf("failed to change directory to terraform: %w", err)
	}

	cmd := exec.Command(absoluteBinaryPath, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()

	// Change back to the original working directory
	if err := os.Chdir(oldDir); err != nil {
		return fmt.Errorf("failed to change back to original directory: %w", err)
	}

	return err
}

// checkStagedGoFiles checks if there are staged Go files
func checkStagedGoFiles() error {
	cmd := exec.Command("git", "diff", "--cached", "--name-only")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to check staged files: %w", err)
	}

	files := strings.TrimSpace(string(output))
	if files == "" {
		fmt.Println("No files staged")
		os.Exit(0)
	}

	// Check each line for .go files
	hasGoFiles := false
	for _, line := range strings.Split(files, "\n") {
		if filepath.Ext(strings.TrimSpace(line)) == ".go" {
			hasGoFiles = true
			break
		}
	}
	
	if !hasGoFiles {
		fmt.Println("No Go files staged, skipping checks")
		os.Exit(0)
	}
	
	return nil
}

// runAPICompatibilityCheck runs the API compatibility check logic
func runAPICompatibilityCheck(oldCommit, newCommit string) error {
	if oldCommit == "" {
		oldCommit = "HEAD~1"
	}
	if newCommit == "" {
		newCommit = "HEAD"
	}

	// Check if this is the first commit - if so, skip API check
	if err := exec.Command("git", "rev-parse", "--verify", oldCommit).Run(); err != nil {
		fmt.Printf("No previous commit found (%s), skipping API compatibility check\n", oldCommit)
		return nil
	}

	fmt.Printf("Checking API compatibility between %s and %s\n", oldCommit, newCommit)

	// Create temporary directories
	oldDir, err := os.MkdirTemp("", "api-check-old-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(oldDir)

	newDir, err := os.MkdirTemp("", "api-check-new-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(newDir)

	// Add worktrees
	if err := exec.Command("git", "worktree", "add", oldDir, oldCommit).Run(); err != nil {
		return fmt.Errorf("failed to checkout %s: %w", oldCommit, err)
	}
	defer exec.Command("git", "worktree", "remove", oldDir, "--force").Run()

	if err := exec.Command("git", "worktree", "add", newDir, newCommit).Run(); err != nil {
		exec.Command("git", "worktree", "remove", oldDir, "--force").Run()
		return fmt.Errorf("failed to checkout %s: %w", newCommit, err)
	}
	defer exec.Command("git", "worktree", "remove", newDir, "--force").Run()

	// Find all Go packages
	packages, err := findGoPackages()
	if err != nil {
		return fmt.Errorf("failed to find Go packages: %w", err)
	}

	breakingChanges := false
	for _, pkg := range packages {
		oldPkgPath := filepath.Join(oldDir, pkg)
		newPkgPath := filepath.Join(newDir, pkg)

		if _, err := os.Stat(oldPkgPath); os.IsNotExist(err) {
			continue
		}
		if _, err := os.Stat(newPkgPath); os.IsNotExist(err) {
			continue
		}

		fmt.Printf("Checking package %s...\n", pkg)
		cmd := exec.Command("apidiff", oldPkgPath, newPkgPath)
		if err := cmd.Run(); err != nil {
			fmt.Printf("⚠️  Breaking changes detected in %s\n", pkg)
			breakingChanges = true
		} else {
			fmt.Printf("✅ %s is API compatible\n", pkg)
		}
	}

	if breakingChanges {
		return fmt.Errorf("breaking changes detected")
	}

	fmt.Println("✅ No breaking changes detected")
	return nil
}

// checkDocumentationQuality checks documentation quality for staged files
func checkDocumentationQuality() error {
	fmt.Println("📚 Checking documentation quality...")

	cmd := exec.Command("git", "diff", "--cached", "--name-only")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get staged files: %w", err)
	}

	files := strings.TrimSpace(string(output))
	missingDocs := false

	for _, file := range strings.Split(files, "\n") {
		file = strings.TrimSpace(file)
		if filepath.Ext(file) != ".go" || strings.HasSuffix(file, "_test.go") {
			continue
		}

		pkgDir := filepath.Dir(file)
		cmd := exec.Command("go", "doc", pkgDir)
		output, err := cmd.Output()
		if err != nil {
			continue
		}

		if !strings.Contains(string(output), "Package") || !strings.Contains(string(output), "provides") {
			fmt.Printf("❌ Package %s lacks proper documentation\n", pkgDir)
			missingDocs = true
		}
	}

	if missingDocs {
		return fmt.Errorf("missing documentation detected")
	}

	fmt.Println("✅ Documentation quality check passed")
	return nil
}

// verifyAllDocumentationQuality checks documentation for all packages
func verifyAllDocumentationQuality() error {
	fmt.Println("📚 Verifying documentation quality for all packages...")

	packages, err := findGoPackages()
	if err != nil {
		return fmt.Errorf("failed to find Go packages: %w", err)
	}

	for _, pkg := range packages {
		cmd := exec.Command("go", "doc", pkg)
		output, err := cmd.Output()
		if err != nil {
			continue
		}

		if !strings.Contains(string(output), "Package") || !strings.Contains(string(output), "provides") {
			fmt.Printf("❌ Package %s lacks proper documentation\n", pkg)
			return fmt.Errorf("missing documentation in package %s", pkg)
		}
	}

	fmt.Println("✅ All packages have proper documentation")
	return nil
}

// findGoPackages finds all Go packages in the current directory
func findGoPackages() ([]string, error) {
	var packages []string

	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() && (info.Name() == "vendor" || info.Name() == ".git") {
			return filepath.SkipDir
		}

		if filepath.Ext(path) == ".go" {
			dir := filepath.Dir(path)
			// Avoid duplicates
			found := false
			for _, pkg := range packages {
				if pkg == dir {
					found = true
					break
				}
			}
			if !found {
				packages = append(packages, dir)
			}
		}

		return nil
	})

	return packages, err
}
