package api

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

// CheckCompatibility runs apidiff between two commits and reports breaking changes.
func CheckCompatibility(oldCommit, newCommit string, packages ...string) error {
	if oldCommit == "" {
		oldCommit = "HEAD~1"
	}
	if newCommit == "" {
		newCommit = "HEAD"
	}

	if err := exec.Command("git", "rev-parse", "--verify", oldCommit).Run(); err != nil {
		fmt.Printf("No previous commit found (%s), skipping API compatibility check\n", oldCommit)
		return nil
	}

	packageList, err := preparePackageList(packages)
	if err != nil {
		return err
	}
	if len(packageList) == 0 {
		fmt.Println("No Go packages to check; skipping API compatibility check")
		return nil
	}

	fmt.Printf("Checking API compatibility between %s and %s\n", oldCommit, newCommit)

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

	if err := exec.Command("git", "worktree", "add", oldDir, oldCommit).Run(); err != nil {
		return fmt.Errorf("failed to checkout %s: %w", oldCommit, err)
	}
	defer exec.Command("git", "worktree", "remove", oldDir, "--force").Run()

	if err := exec.Command("git", "worktree", "add", newDir, newCommit).Run(); err != nil {
		exec.Command("git", "worktree", "remove", oldDir, "--force").Run()
		return fmt.Errorf("failed to checkout %s: %w", newCommit, err)
	}
	defer exec.Command("git", "worktree", "remove", newDir, "--force").Run()

	breakingChanges := false
	for _, pkg := range packageList {
		rel := cleanPackagePath(pkg)
		oldPkgPath := filepath.Join(oldDir, rel)
		newPkgPath := filepath.Join(newDir, rel)

		if _, err := os.Stat(oldPkgPath); os.IsNotExist(err) {
			continue
		}
		if _, err := os.Stat(newPkgPath); os.IsNotExist(err) {
			continue
		}

		fmt.Printf("Checking package %s...\n", rel)
		cmd := exec.Command("apidiff", oldPkgPath, newPkgPath)
		if err := cmd.Run(); err != nil {
			fmt.Printf("⚠️  Breaking changes detected in %s\n", rel)
			breakingChanges = true
		} else {
			fmt.Printf("✅ %s is API compatible\n", rel)
		}
	}

	if breakingChanges {
		return fmt.Errorf("breaking changes detected")
	}

	fmt.Println("✅ No breaking changes detected")
	return nil
}

func findGoPackages() ([]string, error) {
	seen := make(map[string]struct{})

	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			name := info.Name()
			if name == "vendor" || name == ".git" {
				return filepath.SkipDir
			}
			return nil
		}

		if filepath.Ext(path) != ".go" {
			return nil
		}

		dir := filepath.Dir(path)
		seen[dir] = struct{}{}
		return nil
	})
	if err != nil {
		return nil, err
	}

	var packages []string
	for dir := range seen {
		packages = append(packages, dir)
	}

	return normalizePackages(packages), nil
}

func preparePackageList(packages []string) ([]string, error) {
	if len(packages) == 0 {
		return findGoPackages()
	}
	return normalizePackages(packages), nil
}

func normalizePackages(packages []string) []string {
	seen := make(map[string]struct{})
	var result []string

	for _, pkg := range packages {
		clean := cleanPackagePath(pkg)
		if clean == "" {
			continue
		}
		if _, ok := seen[clean]; ok {
			continue
		}
		seen[clean] = struct{}{}
		result = append(result, clean)
	}

	sort.Strings(result)
	return result
}

func cleanPackagePath(pkg string) string {
	pkg = filepath.Clean(pkg)
	pkg = strings.TrimPrefix(pkg, "./")
	pkg = strings.TrimPrefix(pkg, string(os.PathSeparator))
	if pkg == "" {
		return "."
	}
	pkg = filepath.ToSlash(pkg)
	return pkg
}
