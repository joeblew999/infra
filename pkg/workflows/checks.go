package workflows

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/joeblew999/infra/pkg/workflows/api"
)

var vetIgnorePrefixes = []string{"pkg/deck/repo-tests/"}

// RunPreCommitChecks executes lightweight quality gates for staged Go files.
func RunPreCommitChecks() error {
	fmt.Println("🔍 Running pre-commit checks...")

	staged, err := stagedGoFiles()
	if err != nil {
		return err
	}
	if len(staged) == 0 {
		fmt.Println("No Go files staged; skipping pre-commit checks")
		return nil
	}

	if err := ensureGoFmt(staged); err != nil {
		return err
	}

	packages := packagesFromFiles(staged)

	if err := runGoVetPackages(packages); err != nil {
		return err
	}

	if err := runGoTestPackages(packages); err != nil {
		return err
	}

	if err := api.CheckCompatibility("", "", packages...); err != nil {
		return err
	}

	if err := checkDocumentationQuality(staged); err != nil {
		return err
	}

	fmt.Println("✅ Pre-commit checks passed")
	return nil
}

// RunCIChecks executes the full suite of repository quality checks.
func RunCIChecks() error {
	fmt.Println("🔍 Running CI checks...")

	if err := runGoVet(); err != nil {
		return err
	}

	if err := runGoTest(); err != nil {
		return err
	}

	if err := api.CheckCompatibility("", ""); err != nil {
		return err
	}

	if err := verifyAllDocumentationQuality(); err != nil {
		return err
	}

	fmt.Println("✅ CI checks passed")
	return nil
}

func stagedGoFiles() ([]string, error) {
	cmd := exec.Command("git", "diff", "--cached", "--name-only")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to check staged files: %w", err)
	}

	var files []string
	for _, line := range strings.Split(strings.TrimSpace(string(output)), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if filepath.Ext(line) != ".go" {
			continue
		}
		files = append(files, line)
	}

	return files, nil
}

func ensureGoFmt(files []string) error {
	if len(files) == 0 {
		return nil
	}

	args := append([]string{"-l"}, files...)
	cmd := exec.Command("gofmt", args...)
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("gofmt failed: %w", err)
	}

	unformatted := strings.TrimSpace(string(output))
	if unformatted != "" {
		fmt.Println("❌ gofmt detected unformatted files:")
		for _, file := range strings.Split(unformatted, "\n") {
			if file != "" {
				fmt.Printf("  %s\n", file)
			}
		}
		return fmt.Errorf("gofmt required")
	}

	fmt.Println("✨ gofmt formatting looks good")
	return nil
}

func runGoVet() error {
	fmt.Println("🔧 Running go vet ./...")
	cmd := exec.Command("go", "vet", "./...")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("go vet failed: %w", err)
	}
	return nil
}

func runGoTest() error {
	fmt.Println("🧪 Running go test ./...")
	cmd := exec.Command("go", "test", "./...")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("go test failed: %w", err)
	}
	return nil
}

func runGoVetPackages(packages []string) error {
	filtered := filterIgnoredPackages(packages, vetIgnorePrefixes)
	if len(filtered) == 0 {
		fmt.Println("🔧 go vet skipped (no staged packages)")
		return nil
	}

	args := append([]string{"vet"}, filtered...)
	cmd := exec.Command("go", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("go vet failed: %w", err)
	}
	return nil
}

func runGoTestPackages(packages []string) error {
	filtered := filterIgnoredPackages(packages, vetIgnorePrefixes)
	if len(filtered) == 0 {
		fmt.Println("🧪 go test skipped (no staged packages)")
		return nil
	}

	args := append([]string{"test"}, filtered...)
	cmd := exec.Command("go", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("go test failed: %w", err)
	}
	return nil
}

func packagesFromFiles(files []string) []string {
	set := make(map[string]struct{})
	for _, file := range files {
		dir := filepath.Dir(file)
		set[normalizePackagePath(dir)] = struct{}{}
	}
	var packages []string
	for pkg := range set {
		packages = append(packages, pkg)
	}
	sort.Strings(packages)
	return packages
}

func normalizePackagePath(dir string) string {
	if dir == "." {
		return "."
	}
	if strings.HasPrefix(dir, "./") {
		return dir
	}
	return "./" + dir
}

func filterIgnoredPackages(packages []string, ignores []string) []string {
	var filtered []string
	for _, pkg := range packages {
		ignored := false
		for _, prefix := range ignores {
			if strings.HasPrefix(strings.TrimPrefix(pkg, "./"), prefix) {
				ignored = true
				break
			}
		}
		if !ignored {
			filtered = append(filtered, pkg)
		}
	}
	return filtered
}

func checkDocumentationQuality(files []string) error {
	fmt.Println("📚 Checking documentation quality for staged packages...")

	seen := make(map[string]struct{})
	missingDocs := false

	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}

		pkgDir := filepath.Dir(file)
		if _, dup := seen[pkgDir]; dup {
			continue
		}
		seen[pkgDir] = struct{}{}

		cmd := exec.Command("go", "doc", pkgDir)
		output, err := cmd.Output()
		if err != nil {
			continue
		}

		text := string(output)
		if !strings.Contains(text, "Package") || !strings.Contains(text, "provides") {
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

func verifyAllDocumentationQuality() error {
	fmt.Println("📚 Verifying documentation quality for all packages...")

	packages, err := listAllGoPackages()
	if err != nil {
		return fmt.Errorf("failed to find Go packages: %w", err)
	}

	for _, pkg := range packages {
		cmd := exec.Command("go", "doc", pkg)
		output, err := cmd.Output()
		if err != nil {
			continue
		}

		text := string(output)
		if !strings.Contains(text, "Package") || !strings.Contains(text, "provides") {
			return fmt.Errorf("missing documentation in package %s", pkg)
		}
	}

	fmt.Println("✅ All packages have proper documentation")
	return nil
}

func listAllGoPackages() ([]string, error) {
	packages := []string{}
	seen := make(map[string]struct{})

	err := filepath.WalkDir(".", func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			name := d.Name()
			if name == "vendor" || name == ".git" {
				return filepath.SkipDir
			}
			return nil
		}

		if filepath.Ext(path) != ".go" {
			return nil
		}

		dir := filepath.Dir(path)
		if _, ok := seen[dir]; ok {
			return nil
		}
		seen[dir] = struct{}{}
		packages = append(packages, dir)
		return nil
	})

	return packages, err
}
