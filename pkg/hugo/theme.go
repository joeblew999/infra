package hugo

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/joeblew999/infra/pkg/log"
)

// ThemeManager handles Hugo theme installation and management
type ThemeManager struct {
	sourceDir string
	themeName string
	themeRepo string
}

// NewThemeManager creates a new theme manager
func NewThemeManager(sourceDir string) *ThemeManager {
	return &ThemeManager{
		sourceDir: sourceDir,
		themeName: "hugoplate",
		themeRepo: "https://github.com/zeon-studio/hugoplate.git",
	}
}

// InstallTheme installs the HugoPlate theme
func (tm *ThemeManager) InstallTheme() error {
	themesDir := filepath.Join(tm.sourceDir, "themes")
	themeDir := filepath.Join(themesDir, tm.themeName)

	// Check if theme already exists
	if _, err := os.Stat(themeDir); err == nil {
		log.Info("HugoPlate theme already installed", "path", themeDir)
		return nil
	}

	// Create themes directory
	if err := os.MkdirAll(themesDir, 0755); err != nil {
		return fmt.Errorf("failed to create themes directory: %w", err)
	}

	log.Info("Installing HugoPlate theme", "repo", tm.themeRepo, "path", themeDir)

	// Clone the theme repository
	cmd := exec.Command("git", "clone", tm.themeRepo, themeDir)
	cmd.Dir = tm.sourceDir
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to clone HugoPlate theme: %w", err)
	}

	log.Info("HugoPlate theme installed successfully")

	// Copy example config if it exists
	if err := tm.copyExampleConfig(); err != nil {
		log.Warn("Failed to copy example config", "error", err)
	}

	return nil
}

// copyExampleConfig copies theme example configuration
func (tm *ThemeManager) copyExampleConfig() error {
	themeDir := filepath.Join(tm.sourceDir, "themes", tm.themeName)
	exampleConfig := filepath.Join(themeDir, "exampleSite", "hugo.yaml")

	if _, err := os.Stat(exampleConfig); os.IsNotExist(err) {
		// Try hugo.toml
		exampleConfig = filepath.Join(themeDir, "exampleSite", "hugo.toml")
		if _, err := os.Stat(exampleConfig); os.IsNotExist(err) {
			return fmt.Errorf("no example config found")
		}
	}

	targetConfig := filepath.Join(tm.sourceDir, "hugo.example.yaml")

	// Read example config
	data, err := os.ReadFile(exampleConfig)
	if err != nil {
		return err
	}

	// Write to example file
	return os.WriteFile(targetConfig, data, 0644)
}

// MigrateContent migrates existing markdown content to Hugo structure
func (tm *ThemeManager) MigrateContent(docsDir string) error {
	contentDir := filepath.Join(tm.sourceDir, "content")

	log.Info("Migrating content to Hugo structure", "from", docsDir, "to", contentDir)

	// Create content directory
	if err := os.MkdirAll(contentDir, 0755); err != nil {
		return fmt.Errorf("failed to create content directory: %w", err)
	}

	// Copy all markdown files maintaining directory structure
	return filepath.Walk(docsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and non-markdown files
		if info.IsDir() || filepath.Ext(path) != ".md" {
			return nil
		}

		// Calculate relative path
		relPath, err := filepath.Rel(docsDir, path)
		if err != nil {
			return err
		}

		// Create target path in content directory
		targetPath := filepath.Join(contentDir, relPath)
		targetDir := filepath.Dir(targetPath)

		// Create target directory
		if err := os.MkdirAll(targetDir, 0755); err != nil {
			return err
		}

		// Copy file
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		// Add Hugo front matter if not present
		content := string(data)
		if !tm.hasFrontMatter(content) {
			title := tm.extractTitle(content, filepath.Base(path))
			frontMatter := fmt.Sprintf(`---
title: "%s"
draft: false
---

`, title)
			content = frontMatter + content
		}

		return os.WriteFile(targetPath, []byte(content), 0644)
	})
}

// hasFrontMatter checks if content already has Hugo front matter
func (tm *ThemeManager) hasFrontMatter(content string) bool {
	return len(content) > 4 && content[:4] == "---\n"
}

// extractTitle extracts title from content or filename
func (tm *ThemeManager) extractTitle(content, filename string) string {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "# ") {
			return strings.TrimSpace(line[2:])
		}
	}

	// Fall back to filename
	name := strings.TrimSuffix(filename, filepath.Ext(filename))
	name = strings.ReplaceAll(name, "_", " ")
	name = strings.ReplaceAll(name, "-", " ")
	return strings.Title(name)
}