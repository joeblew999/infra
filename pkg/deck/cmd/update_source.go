package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/joeblew999/infra/pkg/log"
)

// UpdateSourceCmd represents the update-source command
var UpdateSourceCmd = &cobra.Command{
		Use:   "update-source",
		Short: "Update the .source directory with latest upstream deck repositories",
		Long: `Updates the .source directory by cloning or pulling the latest versions
of the upstream deck repositories that our binary pipeline is based on:

- github.com/ajstarks/deck (main deck library)
- github.com/ajstarks/decksh (decksh DSL processor)
- github.com/ajstarks/dubois-data-portraits (example collection)
- github.com/ajstarks/deckviz (data visualization examples)

This ensures our binary pipeline can be rebuilt with the latest upstream changes.`,
	RunE: runUpdateSource,
}

// Removed init() - command is registered in root.go  
// Flags are set up in root.go init()

func runUpdateSource(cmd *cobra.Command, args []string) error {
	force, _ := cmd.Flags().GetBool("force")
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	// Get the pkg/deck directory path
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	// Find pkg/deck directory (could be relative to current location)
	deckDir := findDeckDirectory(wd)
	if deckDir == "" {
		return fmt.Errorf("could not find pkg/deck directory from %s", wd)
	}

	sourceDir := filepath.Join(deckDir, ".source")
	
	log.Info("Updating deck source repositories", 
		"deckDir", deckDir, 
		"sourceDir", sourceDir,
		"force", force,
		"dryRun", dryRun)

	if dryRun {
		fmt.Printf("Would update sources in: %s\n", sourceDir)
		fmt.Printf("Force mode: %v\n", force)
		return nil
	}

	// Remove existing .source if force flag is set
	if force {
		if _, err := os.Stat(sourceDir); err == nil {
			log.Info("Removing existing .source directory")
			if err := os.RemoveAll(sourceDir); err != nil {
				return fmt.Errorf("failed to remove existing .source directory: %w", err)
			}
		}
	}

	// Create .source directory if it doesn't exist
	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		return fmt.Errorf("failed to create .source directory: %w", err)
	}

	// Update each repository
	repos := []struct {
		name        string
		url         string
		dir         string
		description string
	}{
		{
			name:        "deck",
			url:         "https://github.com/ajstarks/deck.git",
			dir:         "deck",
			description: "Main deck library and command-line tools",
		},
		{
			name:        "decksh", 
			url:         "https://github.com/ajstarks/decksh.git",
			dir:         "decksh",
			description: "Decksh DSL processor and language tools",
		},
		{
			name:        "dubois-data-portraits",
			url:         "https://github.com/ajstarks/dubois-data-portraits.git", 
			dir:         "dubois-data-portraits",
			description: "Large example collection: Du Bois data portraits",
		},
		{
			name:        "deckviz",
			url:         "https://github.com/ajstarks/deckviz.git",
			dir:         "deckviz", 
			description: "Large example collection: Data visualization examples",
		},
	}

	var repoNames []string
	for _, repo := range repos {
		log.Info("Processing repository", "name", repo.name, "description", repo.description)
		if err := updateRepository(sourceDir, repo.name, repo.url, repo.dir); err != nil {
			return fmt.Errorf("failed to update %s repository: %w", repo.name, err)
		}
		repoNames = append(repoNames, repo.name)
	}

	// Update go.work file to include all modules
	if err := updateGoWorkFile(sourceDir); err != nil {
		log.Warn("Failed to update go.work file", "error", err)
		// Don't fail the entire command if go.work update fails
	}

	log.Info("Successfully updated all source repositories")
	fmt.Printf("✅ Updated .source directory: %s\n", sourceDir)
	fmt.Printf("📁 Updated repositories: %s\n", strings.Join(repoNames, ", "))
	fmt.Printf("📝 Updated go.work file with all modules\n")
	
	return nil
}

func findDeckDirectory(startDir string) string {
	// Common paths to try
	candidates := []string{
		"pkg/deck",
		"./pkg/deck", 
		"../pkg/deck",
		"../../pkg/deck",
		filepath.Join(startDir, "pkg/deck"),
	}

	for _, candidate := range candidates {
		absPath, err := filepath.Abs(candidate)
		if err != nil {
			continue
		}
		if _, err := os.Stat(absPath); err == nil {
			return absPath
		}
	}

	return ""
}

func updateRepository(baseDir, name, url, dir string) error {
	repoPath := filepath.Join(baseDir, dir)
	
	log.Info("Updating repository", "name", name, "url", url, "path", repoPath)

	// Check if repository already exists
	if _, err := os.Stat(filepath.Join(repoPath, ".git")); err == nil {
		// Repository exists, pull latest changes
		log.Info("Repository exists, pulling latest changes", "name", name)
		return pullRepository(repoPath)
	} else {
		// Repository doesn't exist, clone it
		log.Info("Repository doesn't exist, cloning", "name", name)
		return cloneRepository(baseDir, url, dir)
	}
}

func cloneRepository(baseDir, url, dir string) error {
	// Use shallow clone to avoid downloading full history
	cmd := exec.Command("git", "clone", "--depth", "1", url, dir)
	cmd.Dir = baseDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git clone failed: %w", err)
	}

	return nil
}

func pullRepository(repoPath string) error {
	// First, fetch the latest changes
	fetchCmd := exec.Command("git", "fetch", "origin")
	fetchCmd.Dir = repoPath
	fetchCmd.Stdout = os.Stdout
	fetchCmd.Stderr = os.Stderr

	if err := fetchCmd.Run(); err != nil {
		return fmt.Errorf("git fetch failed: %w", err)
	}

	// Then, reset to origin/master (or main)
	// First, check which default branch exists
	var defaultBranch string
	branches := []string{"main", "master"}
	
	for _, branch := range branches {
		checkCmd := exec.Command("git", "show-ref", "--verify", "--quiet", "refs/remotes/origin/"+branch)
		checkCmd.Dir = repoPath
		if checkCmd.Run() == nil {
			defaultBranch = branch
			break
		}
	}

	if defaultBranch == "" {
		return fmt.Errorf("could not find default branch (main or master)")
	}

	// Reset to the default branch
	resetCmd := exec.Command("git", "reset", "--hard", "origin/"+defaultBranch)
	resetCmd.Dir = repoPath
	resetCmd.Stdout = os.Stdout
	resetCmd.Stderr = os.Stderr

	if err := resetCmd.Run(); err != nil {
		return fmt.Errorf("git reset failed: %w", err)
	}

	log.Info("Successfully updated repository", "branch", defaultBranch)
	return nil
}

// updateGoWorkFile creates or updates the go.work file in .source directory
func updateGoWorkFile(sourceDir string) error {
	goWorkPath := filepath.Join(sourceDir, "go.work")
	
	// Find all Go modules in the source directory
	var modules []string
	
	// Look for go.mod files in immediate subdirectories
	entries, err := os.ReadDir(sourceDir)
	if err != nil {
		return fmt.Errorf("failed to read source directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		
		// Check if this directory has a go.mod file
		modulePath := filepath.Join(sourceDir, entry.Name())
		goModPath := filepath.Join(modulePath, "go.mod")
		
		if _, err := os.Stat(goModPath); err == nil {
			// Add relative path to modules list
			modules = append(modules, "./"+entry.Name())
			log.Info("Found Go module", "path", entry.Name())
		}

		// Also check for go.mod files in subdirectories (for examples)
		subEntries, err := os.ReadDir(modulePath)
		if err != nil {
			continue
		}
		
		for _, subEntry := range subEntries {
			if !subEntry.IsDir() {
				continue
			}
			
			subModulePath := filepath.Join(modulePath, subEntry.Name())
			subGoModPath := filepath.Join(subModulePath, "go.mod")
			
			if _, err := os.Stat(subGoModPath); err == nil {
				// Add relative path to modules list
				relativePath := filepath.Join(".", entry.Name(), subEntry.Name())
				modules = append(modules, relativePath)
				log.Info("Found Go submodule", "path", relativePath)
			}
		}
	}

	if len(modules) == 0 {
		log.Info("No Go modules found, skipping go.work creation")
		return nil
	}

	// Create go.work content
	goWorkContent := "go 1.21\n\n"
	goWorkContent += "use (\n"
	for _, module := range modules {
		goWorkContent += fmt.Sprintf("    %s\n", module)
	}
	goWorkContent += ")\n"

	// Write go.work file
	if err := os.WriteFile(goWorkPath, []byte(goWorkContent), 0644); err != nil {
		return fmt.Errorf("failed to write go.work file: %w", err)
	}

	log.Info("Updated go.work file", "path", goWorkPath, "modules", len(modules))
	return nil
}