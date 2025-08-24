package deck

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/joeblew999/infra/pkg/dep"
	"github.com/joeblew999/infra/pkg/log"
	"github.com/joeblew999/infra/pkg/config"
)

// Watcher monitors filesystem for .dsh file changes
type Watcher struct {
	Builder      *Builder
	WatchPaths   []string
	OutputDir    string
	CacheDir     string
	Processing   map[string]bool
	mu           sync.RWMutex
	ctx          context.Context
	cancel       context.CancelFunc
	wg           sync.WaitGroup
}

// NewWatcher creates a new file watcher
func NewWatcher() *Watcher {
	ctx, cancel := context.WithCancel(context.Background())
	return &Watcher{
		Builder:    NewBuilder(),
		OutputDir:  filepath.Join(config.GetDataPath(), "deck", "cache"),
		CacheDir:   filepath.Join(config.GetDataPath(), "deck", "cache"),
		Processing: make(map[string]bool),
		ctx:        ctx,
		cancel:     cancel,
	}
}

// AddPath adds a path to watch
func (w *Watcher) AddPath(path string) {
	w.WatchPaths = append(w.WatchPaths, path)
}

// Start begins watching for .dsh file changes
func (w *Watcher) Start() error {
	if err := os.MkdirAll(w.OutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	log.Info("Starting .dsh file watcher", "paths", w.WatchPaths)
	
	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	
	// Initial scan
	w.scanFiles()
	
	// Start watch loop in goroutine
	go w.watchLoop()
	
	// Wait for shutdown signal
	<-sigChan
	log.Info("Shutting down file watcher...")
	return w.Stop()
}

// watchLoop runs the main file scanning loop
func (w *Watcher) watchLoop() {
	ticker := time.NewTicker(time.Duration(WatcherPollInterval) * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-w.ctx.Done():
			log.Info("Watch loop stopping due to context cancellation")
			return
		case <-ticker.C:
			w.scanFiles()
		}
	}
}

// Stop gracefully shuts down the watcher
func (w *Watcher) Stop() error {
	log.Info("Stopping file watcher and waiting for active tasks...")
	
	// Cancel context to stop new work
	w.cancel()
	
	// Wait for all goroutines to finish with timeout
	done := make(chan struct{})
	go func() {
		w.wg.Wait()
		close(done)
	}()
	
	select {
	case <-done:
		log.Info("All file processing tasks completed")
	case <-time.After(30 * time.Second):
		log.Warn("Timeout waiting for tasks to complete, forcing shutdown")
	}
	
	return nil
}

// scanFiles checks for .dsh files and processes new/changed ones
func (w *Watcher) scanFiles() {
	for _, watchPath := range w.WatchPaths {
		if err := filepath.Walk(watchPath, w.processFile); err != nil {
			log.Warn("Error scanning files", "path", watchPath, "error", err)
		}
	}
}

// processFile handles individual .dsh files
func (w *Watcher) processFile(path string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}
	
	if info.IsDir() {
		return nil
	}
	
	if !strings.HasSuffix(path, ".dsh") {
		return nil
	}
	
	// Skip if already processing (thread-safe check)
	w.mu.RLock()
	if w.Processing[path] {
		w.mu.RUnlock()
		return nil
	}
	w.mu.RUnlock()
	
	// Check if file has been modified recently
	if time.Since(info.ModTime()) > time.Duration(FileModificationTimeout)*time.Second {
		return nil
	}
	
	// Mark as processing (thread-safe)
	w.mu.Lock()
	w.Processing[path] = true
	w.mu.Unlock()
	
	// Start processing in goroutine with proper cleanup
	w.wg.Add(1)
	go w.processDSHFile(path)
	
	return nil
}

// processDSHFile processes a single .dsh file through the pipeline
func (w *Watcher) processDSHFile(dshPath string) {
	defer func() {
		// Clean up processing state (thread-safe)
		w.mu.Lock()
		delete(w.Processing, dshPath)
		w.mu.Unlock()
		w.wg.Done()
	}()
	
	// Check if context is cancelled before starting
	select {
	case <-w.ctx.Done():
		log.Info("Skipping file processing due to shutdown", "path", dshPath)
		return
	default:
	}
	
	log.Info("Processing .dsh file", "path", dshPath)
	
	// Step 1: .dsh -> XML
	xmlPath := filepath.Join(w.OutputDir, filepath.Base(dshPath)+".xml")
	if err := w.runDecksh(dshPath, xmlPath); err != nil {
		log.Error("Failed to compile .dsh to XML", "error", err)
		return
	}
	
	// Step 2: XML -> SVG
	svgPath := filepath.Join(w.OutputDir, filepath.Base(dshPath)+".svg")
	if err := w.runSvgdeck(xmlPath, svgPath); err != nil {
		log.Error("Failed to convert XML to SVG", "error", err)
		return
	}
	
	log.Info("Pipeline completed", "dsh", dshPath, "xml", xmlPath, "svg", svgPath)
}

// runDecksh runs decksh to compile .dsh to XML
func (w *Watcher) runDecksh(inputPath, outputPath string) error {
	deckshPath, err := dep.Get("decksh")
	if err != nil {
		return fmt.Errorf("decksh not found in .dep: %w", err)
	}
	
	// Validate tool exists and is executable
	if err := validateTool(deckshPath, "decksh"); err != nil {
		return fmt.Errorf("decksh validation failed: %w", err)
	}
	
	cmd := exec.Command(deckshPath, inputPath)
	cmd.Env = append(os.Environ(), "DECKFONTS="+filepath.Join(config.GetDataPath(), "deck", "fonts"))
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("decksh failed: %w, output: %s", err, string(output))
	}
	
	// Fix XML attributes - add quotes around color values
	fixedXML := strings.ReplaceAll(string(output), `color=red`, `color="red"`)
	fixedXML = strings.ReplaceAll(fixedXML, `color=gray`, `color="gray"`)
	
	// Write XML output
	return os.WriteFile(outputPath, []byte(fixedXML), 0644)
}

// runSvgdeck runs svgdeck to convert XML to SVG
func (w *Watcher) runSvgdeck(inputPath, outputPath string) error {
	svgdeckPath, err := dep.Get("svgdeck")
	if err != nil {
		return fmt.Errorf("svgdeck not found in .dep: %w", err)
	}
	
	// Validate tool exists and is executable
	if err := validateTool(svgdeckPath, "svgdeck"); err != nil {
		return fmt.Errorf("svgdeck validation failed: %w", err)
	}
	
	cmd := exec.Command(svgdeckPath, inputPath)
	cmd.Env = append(os.Environ(), "DECKFONTS="+filepath.Join(config.GetDataPath(), "deck", "fonts"))
	cmd.Dir = filepath.Dir(outputPath) // Output to same directory
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("svgdeck failed: %w, output: %s", err, string(output))
	}
	
	// Rename the output file to match our expected name
	expectedSVG := strings.TrimSuffix(inputPath, ".xml") + ".svg"
	if _, err := os.Stat(expectedSVG); err == nil {
		return os.Rename(expectedSVG, outputPath)
	}
	
	return nil
}

// validateTool checks if a tool exists and is executable
func validateTool(toolPath, toolName string) error {
	// Check if file exists
	if _, err := os.Stat(toolPath); os.IsNotExist(err) {
		return fmt.Errorf("tool %s not found at path: %s", toolName, toolPath)
	}
	
	// Check if executable
	if err := exec.Command(toolPath, "--version").Run(); err != nil {
		// Fallback: try --help if --version fails
		if err := exec.Command(toolPath, "--help").Run(); err != nil {
			return fmt.Errorf("tool %s at %s is not executable or functioning properly", toolName, toolPath)
		}
	}
	
	return nil
}