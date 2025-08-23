package deck

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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
}

// NewWatcher creates a new file watcher
func NewWatcher() *Watcher {
	return &Watcher{
		Builder:    NewBuilder(),
		OutputDir:  filepath.Join(config.GetDataPath(), "deck", "cache"),
		CacheDir:   filepath.Join(config.GetDataPath(), "deck", "cache"),
		Processing: make(map[string]bool),
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
	
	// Initial scan
	w.scanFiles()
	
	// Watch loop
	for {
		w.scanFiles()
		time.Sleep(2 * time.Second)
	}
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
	
	// Skip if already processing
	if w.Processing[path] {
		return nil
	}
	
	// Check if file has been modified recently
	if time.Since(info.ModTime()) > 10*time.Second {
		return nil
	}
	
	w.Processing[path] = true
	go w.processDSHFile(path)
	
	return nil
}

// processDSHFile processes a single .dsh file through the pipeline
func (w *Watcher) processDSHFile(dshPath string) {
	defer func() { delete(w.Processing, dshPath) }()
	
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