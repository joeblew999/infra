package deck

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/joeblew999/infra/pkg/log"
)

// Tool represents a deck tool to build
type Tool struct {
	Name    string
	Repo    string
	Package string // Go package path
	Binary  string // Output binary name
}

// Tools to build from source
var Tools = []Tool{
	{
		Name:    "decksh",
		Repo:    DeckshRepo,
		Package: "github.com/ajstarks/decksh/cmd/decksh",
		Binary:  DeckshBinary,
	},
	{
		Name:    "deckfmt",
		Repo:    DeckshRepo,
		Package: "github.com/ajstarks/decksh/cmd/dshfmt",
		Binary:  DeckfmtBinary,
	},
	{
		Name:    "decklint",
		Repo:    DeckshRepo,
		Package: "github.com/ajstarks/decksh/cmd/dshlint",
		Binary:  DecklintBinary,
	},
	{
		Name:    "decksvg",
		Repo:    SvgdeckRepo,
		Package: "github.com/ajstarks/deck/cmd/svgdeck",
		Binary:  DecksvgBinary,
	},
	{
		Name:    "deckpng",
		Repo:    SvgdeckRepo,
		Package: "github.com/ajstarks/deck/cmd/pngdeck",
		Binary:  DeckpngBinary,
	},
	{
		Name:    "deckpdf",
		Repo:    SvgdeckRepo,
		Package: "github.com/ajstarks/deck/cmd/pdfdeck",
		Binary:  DeckpdfBinary,
	},
}

// Builder handles source compilation
type Builder struct {
	SourceDir string // Where to clone source repos
	BuildDir  string // Where to build binaries
	WASMDir   string // Where to store WASM files
}

// NewBuilder creates a new deck builder
func NewBuilder() *Builder {
	return &Builder{
		SourceDir: SourceDir,
		BuildDir:  filepath.Join(BuildRoot, "bin"),
		WASMDir:   filepath.Join(BuildRoot, "wasm"),
	}
}

// EnsureDirectories creates build directories
func (b *Builder) EnsureDirectories() error {
	dirs := []string{b.SourceDir, b.BuildDir, b.WASMDir}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}
	return nil
}

// CloneRepo clones a git repository
func (b *Builder) CloneRepo(repo, dest string) error {
	if _, err := os.Stat(dest); err == nil {
		log.Info("Repository already cloned", "repo", repo, "dest", dest)
		return nil
	}

	log.Info("Cloning repository", "repo", repo, "dest", dest)
	cmd := exec.Command("git", "clone", "--depth", "1", repo, dest)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// BuildNative builds native binary for current platform
func (b *Builder) BuildNative(tool Tool) (string, error) {
	outputPath := filepath.Join(b.BuildDir, tool.Binary)
	if runtime.GOOS == "windows" {
		outputPath += ".exe"
	}

	// Check if already built
	if _, err := os.Stat(outputPath); err == nil {
		log.Info("Native binary already exists", "tool", tool.Name, "path", outputPath)
		return outputPath, nil
	}

	// Get repo name from URL for shared directory
	repoName := filepath.Base(tool.Repo)
	if filepath.Ext(repoName) == ".git" {
		repoName = repoName[:len(repoName)-4]
	}
	
	repoDir := filepath.Join(b.SourceDir, repoName)
	if err := b.CloneRepo(tool.Repo, repoDir); err != nil {
		return "", fmt.Errorf("failed to clone %s: %w", tool.Name, err)
	}

	// Always generate go.work for isolated module builds
	workFile := filepath.Join(b.SourceDir, "go.work")
	absWorkFile, _ := filepath.Abs(workFile)
	
	// Create new go.work file
	os.Remove(absWorkFile) // Remove existing go.work to start fresh
	
	workCmd := exec.Command("go", "work", "init")
	workCmd.Dir = b.SourceDir
	workCmd.Env = append(os.Environ(), "GOWORK="+absWorkFile)
	workCmd.Stdout = os.Stdout
	workCmd.Stderr = os.Stderr
	
	if err := workCmd.Run(); err != nil {
		log.Warn("Failed to initialize go.work, continuing with build", "error", err)
	}

	// Add the shared repo directory to go.work
	absToolDir, _ := filepath.Abs(filepath.Join(b.SourceDir, repoName))
	if _, err := os.Stat(absToolDir); err == nil {
		workCmd := exec.Command("go", "work", "use", absToolDir)
		workCmd.Dir = b.SourceDir
		workCmd.Env = append(os.Environ(), "GOWORK="+absWorkFile)
		workCmd.Stdout = os.Stdout
		workCmd.Stderr = os.Stderr
		if err := workCmd.Run(); err != nil {
			log.Warn("Failed to add tool to go.work", "tool", tool.Name, "error", err)
		}
	}

	log.Info("Building native binary", "tool", tool.Name, "package", tool.Package)
	
	// Use absolute path for output
	absOutputPath, _ := filepath.Abs(outputPath)
	cmd := exec.Command("go", "build", "-o", absOutputPath, tool.Package)
	cmd.Dir = repoDir // Build from specific repo directory
	
	// Use absolute path for go.work
	absWorkFile2, _ := filepath.Abs(filepath.Join(b.SourceDir, "go.work"))
	cmd.Env = append(os.Environ(), "GOWORK="+absWorkFile2)
	
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to build %s: %w", tool.Name, err)
	}

	log.Info("Built native binary", "tool", tool.Name, "path", outputPath)
	return outputPath, nil
}

// BuildWASM builds WASM module
func (b *Builder) BuildWASM(tool Tool) (string, error) {
	outputPath := filepath.Join(b.WASMDir, tool.Binary+".wasm")

	// Check if already built
	if _, err := os.Stat(outputPath); err == nil {
		log.Info("WASM module already exists", "tool", tool.Name, "path", outputPath)
		return outputPath, nil
	}

	// Get repo name from URL for shared directory
	repoName := filepath.Base(tool.Repo)
	if filepath.Ext(repoName) == ".git" {
		repoName = repoName[:len(repoName)-4]
	}
	
	repoDir := filepath.Join(b.SourceDir, repoName)
	if err := b.CloneRepo(tool.Repo, repoDir); err != nil {
		return "", fmt.Errorf("failed to clone %s: %w", tool.Name, err)
	}

	log.Info("Building WASM module", "tool", tool.Name, "package", tool.Package)
	
	// Use absolute path for output
	absOutputPath, _ := filepath.Abs(outputPath)
	cmd := exec.Command("go", "build", "-o", absOutputPath, tool.Package)
	cmd.Dir = repoDir // Build from specific repo directory
	
	// Use absolute path for go.work
	absWorkFile3, _ := filepath.Abs(filepath.Join(b.SourceDir, "go.work"))
	cmd.Env = append(os.Environ(), "GOOS=js", "GOARCH=wasm", "GOWORK="+absWorkFile3)
	
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to build WASM %s: %w", tool.Name, err)
	}

	log.Info("Built WASM module", "tool", tool.Name, "path", outputPath)
	return outputPath, nil
}

// BuildAll builds both native and WASM for all tools
func (b *Builder) BuildAll() error {
	if err := b.EnsureDirectories(); err != nil {
		return fmt.Errorf("failed to setup directories: %w", err)
	}

	log.Info("Building all deck tools...")
	
	for _, tool := range Tools {
		// Build native binary
		if _, err := b.BuildNative(tool); err != nil {
			log.Warn("Failed to build native binary", "tool", tool.Name, "error", err)
		}

		// Build WASM module
		if _, err := b.BuildWASM(tool); err != nil {
			log.Warn("Failed to build WASM module", "tool", tool.Name, "error", err)
		}
	}

	log.Info("All deck tools built successfully")
	return nil
}

// BuildTool builds specific tool
func (b *Builder) BuildTool(name string) error {
	for _, tool := range Tools {
		if tool.Name == name {
			if _, err := b.BuildNative(tool); err != nil {
				return fmt.Errorf("failed to build native %s: %w", name, err)
			}
			if _, err := b.BuildWASM(tool); err != nil {
				return fmt.Errorf("failed to build WASM %s: %w", name, err)
			}
			return nil
		}
	}
	return fmt.Errorf("tool %s not found", name)
}

// GetPaths returns paths to built binaries and WASM
func (b *Builder) GetPaths() (binaries, wasm map[string]string) {
	binaries = make(map[string]string)
	wasm = make(map[string]string)

	for _, tool := range Tools {
		binaryPath := filepath.Join(b.BuildDir, tool.Binary)
		if runtime.GOOS == "windows" {
			binaryPath += ".exe"
		}
		if _, err := os.Stat(binaryPath); err == nil {
			binaries[tool.Name] = binaryPath
		}

		wasmPath := filepath.Join(b.WASMDir, tool.Binary+".wasm")
		if _, err := os.Stat(wasmPath); err == nil {
			wasm[tool.Name] = wasmPath
		}
	}
	return
}