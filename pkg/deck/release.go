package deck

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/joeblew999/infra/pkg/log"
)

// Release handles cross-platform packaging for GitHub releases
type Release struct {
	Name        string
	Version     string
	BuildDir    string
	WASMDir     string
	TargetOS    string
	TargetArch  string
	OutputDir   string
}

// GetGitVersion returns the current git tag or commit hash
func GetGitVersion() string {
	// Try to get git tag first
	cmd := exec.Command("git", "describe", "--tags", "--exact-match")
	output, err := cmd.Output()
	if err == nil {
		return strings.TrimSpace(string(output))
	}
	
	// Fallback to commit hash
	cmd = exec.Command("git", "rev-parse", "--short", "HEAD")
	output, err = cmd.Output()
	if err == nil {
		return "v0.0.0-" + strings.TrimSpace(string(output))
	}
	
	// Final fallback
	return "v0.0.0-dev"
}

// NewRelease creates a new release packager
func NewRelease(version string) *Release {
	if version == "" {
		version = GetGitVersion()
	}
	return &Release{
		Name:       "deck-tools",
		Version:    version,
		BuildDir:   filepath.Join("pkg", "deck", "build", "bin"),
		WASMDir:    filepath.Join("pkg", "deck", "build", "wasm"),
		OutputDir:  "pkg/deck/release",
		TargetOS:   runtime.GOOS,
		TargetArch: runtime.GOARCH,
	}
}

// NewTargetRelease creates release for specific platform/arch
func NewTargetRelease(version, os, arch string) *Release {
	if version == "" {
		version = GetGitVersion()
	}
	return &Release{
		Name:       "deck-tools",
		Version:    version,
		BuildDir:   filepath.Join("pkg", "deck", "build", "bin"),
		WASMDir:    filepath.Join("pkg", "deck", "build", "wasm"),
		OutputDir:  "pkg/deck/release",
		TargetOS:   os,
		TargetArch: arch,
	}
}

// Build builds for the current platform
func (r *Release) Build() error {
	return r.buildForPlatform(r.TargetOS, r.TargetArch)
}

// BuildAllTargets builds for all supported platforms
func (r *Release) BuildAllTargets() error {
	targets := []struct {
		os   string
		arch string
	}{
		{"darwin", "amd64"},
		{"darwin", "arm64"},
		{"linux", "amd64"},
		{"linux", "arm64"},
		{"windows", "amd64"},
	}

	for _, target := range targets {
		release := NewTargetRelease(r.Version, target.os, target.arch)
		if err := release.Build(); err != nil {
			log.Warn("Failed to build for target", "os", target.os, "arch", target.arch, "error", err)
			continue
		}
	}

	return nil
}

// buildForPlatform creates release package for target platform
func (r *Release) buildForPlatform(targetOS, targetArch string) error {
	packageName := fmt.Sprintf("%s-%s-%s-%s", r.Name, r.Version, targetOS, targetArch)
	packageDir := filepath.Join(r.OutputDir, packageName)
	
	// Create directories
	if err := os.MkdirAll(filepath.Join(packageDir, "bin"), 0755); err != nil {
		return fmt.Errorf("failed to create bin directory: %w", err)
	}
	if err := os.MkdirAll(filepath.Join(packageDir, "wasm"), 0755); err != nil {
		return fmt.Errorf("failed to create wasm directory: %w", err)
	}
	
	// Build native binaries
	if err := r.buildBinaries(packageDir, targetOS, targetArch); err != nil {
		return fmt.Errorf("failed to build binaries: %w", err)
	}
	
	// Copy WASM modules
	if err := r.copyWASM(packageDir, targetOS, targetArch); err != nil {
		return fmt.Errorf("failed to copy WASM: %w", err)
	}
	
	// Create tar.gz package
	return r.createPackage(packageDir, targetOS, targetArch)
}

// buildBinaries copies existing binaries for packaging
func (r *Release) buildBinaries(packageDir, targetOS, _ string) error {
	tools := []string{"decksh", "dshfmt", "dshlint", "svgdeck", "pngdeck", "pdfdeck"}
	
	for _, tool := range tools {
		sourcePath := filepath.Join(r.BuildDir, tool)
		destPath := filepath.Join(packageDir, "bin", tool)
		
		if targetOS == "windows" {
			destPath += ".exe"
		}
		
		// Check if binary exists
		if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
			log.Warn("Binary not found, skipping", "tool", tool, "path", sourcePath)
			continue
		}
		
		// Copy binary
		if err := r.copyFile(sourcePath, destPath); err != nil {
			return fmt.Errorf("failed to copy binary %s: %w", tool, err)
		}
		
		log.Info("Copied binary", "tool", tool, "dest", destPath)
	}
	
	return nil
}

// copyWASM copies WASM modules to package
func (r *Release) copyWASM(packageDir, _, _ string) error {
	wasmTools := []string{"decksh", "svgdeck"} // Key WASM modules
	
	for _, tool := range wasmTools {
		source := filepath.Join(r.WASMDir, tool+".wasm")
		dest := filepath.Join(packageDir, "wasm", tool+".wasm")
		
		if _, err := os.Stat(source); os.IsNotExist(err) {
			// Build WASM if not exists
			if err := r.buildWASM(tool); err != nil {
				return fmt.Errorf("failed to build WASM %s: %w", tool, err)
			}
		}
		
		if err := r.copyFile(source, dest); err != nil {
			return fmt.Errorf("failed to copy WASM %s: %w", tool, err)
		}
	}
	
	return nil
}

// buildWASM builds specific WASM module
func (r *Release) buildWASM(tool string) error {
	var sourceDir, packagePath string
	
	switch tool {
	case "decksh":
		sourceDir = filepath.Join("pkg", "deck", "source", "decksh")
		packagePath = "github.com/ajstarks/decksh/cmd/decksh"
	case "svgdeck":
		sourceDir = filepath.Join("pkg", "deck", "source", "deck")
		packagePath = "github.com/ajstarks/deck/cmd/svgdeck"
	}
	
	output := filepath.Join(r.WASMDir, tool+".wasm")
	cmd := fmt.Sprintf("cd %s && GOOS=js GOARCH=wasm go build -o %s %s", 
		sourceDir, output, packagePath)
	
	return r.executeCommand(cmd)
}

// createPackage creates tar.gz package
func (r *Release) createPackage(packageDir, targetOS, targetArch string) error {
	packageName := fmt.Sprintf("%s-%s-%s-%s.tar.gz", 
		r.Name, r.Version, targetOS, targetArch)
	outputPath := filepath.Join(r.OutputDir, packageName)
	
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create package: %w", err)
	}
	defer file.Close()
	
	gz := gzip.NewWriter(file)
	defer gz.Close()
	
	tarWriter := tar.NewWriter(gz)
	defer tarWriter.Close()
	
	// Add files to tar
	return filepath.Walk(packageDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		if info.IsDir() {
			return nil
		}
		
		relPath, err := filepath.Rel(r.OutputDir, path)
		if err != nil {
			return err
		}
		
		header := &tar.Header{
			Name:    relPath,
			Size:    info.Size(),
			Mode:    int64(info.Mode()),
			ModTime: info.ModTime(),
		}
		
		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}
		
		srcFile, err := os.Open(path)
		if err != nil {
			return err
		}
		defer srcFile.Close()
		
		_, err = io.Copy(tarWriter, srcFile)
		return err
	})
}

// executeCommand executes shell command
func (r *Release) executeCommand(cmd string) error {
	log.Info("Executing", "command", cmd)
	
	// Use shell to execute the command
	actualCmd := exec.Command("sh", "-c", cmd)
	actualCmd.Stdout = os.Stdout
	actualCmd.Stderr = os.Stderr
	
	return actualCmd.Run()
}

// copyFile copies file from source to destination
func (r *Release) copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

// GetPackageName returns the package name for given platform/arch
func GetPackageName(version, os, arch string) string {
	return fmt.Sprintf("deck-tools-%s-%s-%s.tar.gz", version, os, arch)
}