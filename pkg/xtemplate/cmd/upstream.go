package cmd

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

func newUpstreamCommand() *cobra.Command {
	upstreamCmd := &cobra.Command{
		Use:   "upstream",
		Short: "Manage upstream xtemplate fixtures",
	}

	upstreamCmd.AddCommand(newUpstreamSyncCommand())
	return upstreamCmd
}

func newUpstreamSyncCommand() *cobra.Command {
	var (
		repo   string
		ref    string
		source string
		output string
	)

	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Pull upstream xtemplate test templates into the repo",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUpstreamSync(repo, ref, source, output)
		},
	}

	cmd.Flags().StringVar(&repo, "repo", "https://github.com/infogulch/xtemplate.git", "Git repository to sync from")
	cmd.Flags().StringVar(&ref, "ref", "master", "Git ref (branch, tag, commit) to checkout")
	cmd.Flags().StringVar(&source, "source", "test", "Relative path inside the repo to copy templates from")
	cmd.Flags().StringVar(&output, "output", filepath.FromSlash("pkg/xtemplate/templates/upstream"), "Destination directory inside this repository")

	return cmd
}

func runUpstreamSync(repo, ref, source, output string) error {
	tmpDir, err := os.MkdirTemp("", "xtemplate-upstream-*")
	if err != nil {
		return fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	cloneArgs := []string{"clone", "--depth", "1"}
	if ref != "" {
		cloneArgs = append(cloneArgs, "--branch", ref)
	}
	cloneArgs = append(cloneArgs, repo, tmpDir)

	clone := exec.Command("git", cloneArgs...)
	clone.Stdout = os.Stdout
	clone.Stderr = os.Stderr
	if err := clone.Run(); err != nil {
		return fmt.Errorf("git clone failed: %w", err)
	}

	sourceDir := filepath.Join(tmpDir, filepath.FromSlash(source))
	if _, err := os.Stat(sourceDir); err != nil {
		return fmt.Errorf("source directory %s not found in repo: %w", source, err)
	}

	if output == "" {
		return fmt.Errorf("output directory must be provided")
	}

	if err := os.RemoveAll(output); err != nil {
		return fmt.Errorf("remove existing output: %w", err)
	}
	if err := copyDir(sourceDir, output); err != nil {
		return err
	}

	revCmd := exec.Command("git", "-C", tmpDir, "rev-parse", "HEAD")
	revOut, err := revCmd.Output()
	if err == nil {
		versionPath := filepath.Join(output, ".upstream-version")
		if writeErr := os.WriteFile(versionPath, revOut, 0644); writeErr != nil {
			return fmt.Errorf("write upstream version: %w", writeErr)
		}
	}

	fmt.Printf("✅ Synced xtemplate upstream templates from %s@%s into %s\n", repo, ref, output)
	return nil
}

func copyDir(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return os.MkdirAll(dst, 0755)
		}

		target := filepath.Join(dst, rel)
		if d.IsDir() {
			if strings.HasPrefix(d.Name(), ".git") {
				return filepath.SkipDir
			}
			return os.MkdirAll(target, 0755)
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, 0644)
	})
}
