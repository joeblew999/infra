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

	"github.com/joeblew999/infra/pkg/log"
	"github.com/joeblew999/infra/pkg/config"
	"github.com/joeblew999/infra/pkg/goreman"
)

// Watcher monitors filesystem for .dsh file changes
type Watcher struct {
	Builder      *Builder
	WatchPaths   []string
	OutputDir    string
	CacheDir     string
	Processing   map[string]bool
	Formats      []string // Output formats to generate (svg, png, pdf)
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
		OutputDir:  filepath.Join(config.GetDataPath(), CacheDirPath),
		CacheDir:   filepath.Join(config.GetDataPath(), CacheDirPath),
		Processing: make(map[string]bool),
		Formats:    []string{"svg"}, // Default to SVG only
		ctx:        ctx,
		cancel:     cancel,
	}
}

// SetFormats configures which output formats to generate
func (w *Watcher) SetFormats(formats []string) {
	w.Formats = formats
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
	
	// Step 2: XML -> Multiple formats
	var outputPaths []string
	baseName := filepath.Base(dshPath)
	
	for _, format := range w.Formats {
		outputPath := filepath.Join(w.OutputDir, baseName+"."+format)
		
		var err error
		switch format {
		case "svg":
			err = w.runSvgdeck(xmlPath, outputPath)
		case "png":
			err = w.runPngdeck(xmlPath, outputPath)
		case "pdf":
			err = w.runPdfdeck(xmlPath, outputPath)
		default:
			log.Warn("Unsupported format", "format", format, "file", dshPath)
			continue
		}
		
		if err != nil {
			log.Error("Failed to convert XML to format", "format", format, "error", err)
			continue
		}
		
		outputPaths = append(outputPaths, outputPath)
	}
	
	log.Info("Pipeline completed", "dsh", dshPath, "xml", xmlPath, "outputs", outputPaths)
}

// runDecksh runs decksh to compile .dsh to XML
func (w *Watcher) runDecksh(inputPath, outputPath string) error {
	deckshPath := filepath.Join(BuildRoot, "bin", DeckshBinary)
	
	// Check if tool exists
	if _, err := os.Stat(deckshPath); os.IsNotExist(err) {
		return fmt.Errorf("decksh not built: %s", deckshPath)
	}
	
	cmd := exec.Command(deckshPath, inputPath)
	cmd.Env = append(os.Environ(), "DECKFONTS="+config.GetFontPath())
	
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
	svgdeckPath := filepath.Join(BuildRoot, "bin", DecksvgBinary)
	
	// Check if tool exists
	if _, err := os.Stat(svgdeckPath); os.IsNotExist(err) {
		return fmt.Errorf("decksvg not built: %s", svgdeckPath)
	}
	
	cmd := exec.Command(svgdeckPath, inputPath)
	cmd.Env = append(os.Environ(), "DECKFONTS="+config.GetFontPath())
	cmd.Dir = filepath.Dir(outputPath) // Output to same directory
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("decksvg failed: %w, output: %s", err, string(output))
	}
	
	// Rename the output file to match our expected name
	expectedSVG := strings.TrimSuffix(inputPath, ".xml") + ".svg"
	if _, err := os.Stat(expectedSVG); err == nil {
		return os.Rename(expectedSVG, outputPath)
	}
	
	return nil
}

// runPngdeck runs pngdeck to convert XML to PNG
func (w *Watcher) runPngdeck(inputPath, outputPath string) error {
	pngdeckPath := filepath.Join(BuildRoot, "bin", DeckpngBinary)
	
	// Check if tool exists
	if _, err := os.Stat(pngdeckPath); os.IsNotExist(err) {
		return fmt.Errorf("deckpng not built: %s", pngdeckPath)
	}
	
	cmd := exec.Command(pngdeckPath, inputPath)
	cmd.Env = append(os.Environ(), "DECKFONTS="+config.GetFontPath())
	cmd.Dir = filepath.Dir(outputPath) // Output to same directory
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("deckpng failed: %w, output: %s", err, string(output))
	}
	
	// Rename the output file to match our expected name
	expectedPNG := strings.TrimSuffix(inputPath, ".xml") + ".png"
	if _, err := os.Stat(expectedPNG); err == nil {
		return os.Rename(expectedPNG, outputPath)
	}
	
	return nil
}

// runPdfdeck runs pdfdeck to convert XML to PDF
func (w *Watcher) runPdfdeck(inputPath, outputPath string) error {
	pdfdeckPath := filepath.Join(BuildRoot, "bin", DeckpdfBinary)
	
	// Check if tool exists
	if _, err := os.Stat(pdfdeckPath); os.IsNotExist(err) {
		return fmt.Errorf("deckpdf not built: %s", pdfdeckPath)
	}
	
	cmd := exec.Command(pdfdeckPath, inputPath)
	cmd.Env = append(os.Environ(), "DECKFONTS="+config.GetFontPath())
	cmd.Dir = filepath.Dir(outputPath) // Output to same directory
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("deckpdf failed: %w, output: %s", err, string(output))
	}
	
	// Rename the output file to match our expected name
	expectedPDF := strings.TrimSuffix(inputPath, ".xml") + ".pdf"
	if _, err := os.Stat(expectedPDF); err == nil {
		return os.Rename(expectedPDF, outputPath)
	}
	
	return nil
}

// StartAPISupervised starts the deck API service under goreman supervision (idempotent)
// This starts the go-zero based deck API server on the configured port
func StartAPISupervised(port int) error {
	if port == 0 {
		port = 8888 // Default deck API port
	}
	
	// Ensure config directory exists for deck API
	configDir := filepath.Join(config.GetDataPath(), "deck-api")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create deck API config directory: %w", err)
	}
	
	// Create deck API config file
	configPath := filepath.Join(configDir, "deck-api.yaml")
	configContent := fmt.Sprintf(`Name: deck-api
Host: 0.0.0.0
Port: %d
`, port)
	
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		return fmt.Errorf("failed to create deck API config: %w", err)
	}
	
	// Register and start with goreman supervision
	return goreman.RegisterAndStart("deck-api", &goreman.ProcessConfig{
		Command:    "go",
		Args:       []string{"run", "api/deck/deck.go", "-f", configPath},
		WorkingDir: ".",
		Env:        os.Environ(),
	})
}

// StartWatcherSupervised starts the deck file watcher service under goreman supervision (idempotent)  
// This starts the background .dsh file processing service
func StartWatcherSupervised(watchPaths []string, formats []string) error {
	if len(watchPaths) == 0 {
		watchPaths = []string{"test/deck", "docs/deck"} // Default watch paths
	}
	if len(formats) == 0 {
		formats = []string{"svg", "png", "pdf"} // Default formats
	}
	
	// Create a simple config file for the watcher
	configDir := filepath.Join(config.GetDataPath(), "deck-watcher")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create deck watcher config directory: %w", err)
	}
	
	// We'll use environment variables to configure the watcher
	env := append(os.Environ(),
		fmt.Sprintf("DECK_WATCH_PATHS=%s", filepath.Join(watchPaths...)),
		fmt.Sprintf("DECK_FORMATS=%s", filepath.Join(formats...)),
	)
	
	// Register and start with goreman supervision  
	args := append([]string{"run", ".", "cli", "deck", "watch"}, watchPaths...)
	args = append(args, "--formats", strings.Join(formats, ","))
	
	return goreman.RegisterAndStart("deck-watcher", &goreman.ProcessConfig{
		Command:    "go", 
		Args:       args,
		WorkingDir: ".",
		Env:        env,
	})
}

