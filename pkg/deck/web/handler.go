package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gorilla/mux"
	"github.com/joeblew999/infra/pkg/deck"
	"github.com/joeblew999/infra/pkg/log"
)

type Server struct {
	testRunner *deck.GoldenTestRunner
}

type Example struct {
	Name     string `json:"name"`
	Filename string `json:"filename"`
}

type GenerationResult struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
	SVGUrl  string `json:"svgUrl,omitempty"`
	PNGUrl  string `json:"pngUrl,omitempty"`
	PDFUrl  string `json:"pdfUrl,omitempty"`
}

func NewServer() (*Server, error) {
	testRunner, err := deck.NewGoldenTestRunner(deck.BuildRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to create test runner: %w", err)
	}

	return &Server{
		testRunner: testRunner,
	}, nil
}

func (s *Server) ListExamples(w http.ResponseWriter, r *http.Request) {
	inputDir := filepath.Join(deck.PkgDir, "testdata", "input")
	
	files, err := filepath.Glob(filepath.Join(inputDir, "*.dsh"))
	if err != nil {
		log.Error("Failed to list examples", "error", err)
		http.Error(w, "Failed to list examples", http.StatusInternalServerError)
		return
	}

	var examples []Example
	for _, file := range files {
		basename := filepath.Base(file)
		name := strings.TrimSuffix(basename, ".dsh")
		examples = append(examples, Example{
			Name:     name,
			Filename: basename,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(examples)
}

func (s *Server) GenerateExample(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	exampleName := vars["example"]
	
	if exampleName == "" {
		http.Error(w, "Example name required", http.StatusBadRequest)
		return
	}

	// Run the pipeline for this specific example
	result, err := s.runPipelineForExample(exampleName)
	if err != nil {
		log.Error("Pipeline failed", "example", exampleName, "error", err)
		result = &GenerationResult{
			Success: false,
			Error:   err.Error(),
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (s *Server) runPipelineForExample(exampleName string) (*GenerationResult, error) {
	// Run the deck pipeline: .dsh → XML → SVG/PNG/PDF
	err := s.generateOutputsForExample(exampleName)
	if err != nil {
		return nil, fmt.Errorf("pipeline execution failed: %w", err)
	}
	
	basePath := fmt.Sprintf("/outputs/%s", exampleName)
	
	return &GenerationResult{
		Success: true,
		SVGUrl:  basePath + ".svg",
		PNGUrl:  basePath + ".png", 
		PDFUrl:  basePath + ".pdf",
	}, nil
}

func (s *Server) generateOutputsForExample(exampleName string) error {
	// This is a simplified version - we'll run the deck pipeline directly
	// TODO: This could be enhanced to use the golden test runner more directly
	
	inputDir := filepath.Join(deck.PkgDir, "testdata", "input")
	outputDir := filepath.Join(deck.PkgDir, "testdata", "output")
	
	dshFile := filepath.Join(inputDir, exampleName+".dsh")
	xmlFile := filepath.Join(outputDir, exampleName+".xml")
	
	// Step 1: Generate XML from DSH using decksh
	deckshPath := filepath.Join(deck.BuildRoot, "bin", deck.DeckshBinary)
	if err := s.runCommand(deckshPath, "-o", xmlFile, dshFile); err != nil {
		return fmt.Errorf("decksh failed: %w", err)
	}
	
	// Step 2: Generate SVG from XML using decksvg
	decksvgPath := filepath.Join(deck.BuildRoot, "bin", deck.DecksvgBinary)
	if err := s.runCommandInDir(outputDir, decksvgPath, xmlFile); err != nil {
		return fmt.Errorf("decksvg failed: %w", err)
	}
	
	// Step 3: Generate PNG from XML using deckpng  
	deckpngPath := filepath.Join(deck.BuildRoot, "bin", deck.DeckpngBinary)
	if err := s.runCommandInDir(outputDir, deckpngPath, xmlFile); err != nil {
		return fmt.Errorf("deckpng failed: %w", err)
	}
	
	// Step 4: Generate PDF from XML using deckpdf
	deckpdfPath := filepath.Join(deck.BuildRoot, "bin", deck.DeckpdfBinary)
	if err := s.runCommandInDir(outputDir, deckpdfPath, xmlFile); err != nil {
		return fmt.Errorf("deckpdf failed: %w", err)
	}
	
	return nil
}

func (s *Server) ServeOutputs(w http.ResponseWriter, r *http.Request) {
	// Serve files from testdata/output directory
	outputDir := filepath.Join(deck.PkgDir, "testdata", "output")
	fileServer := http.StripPrefix("/outputs/", http.FileServer(http.Dir(outputDir)))
	fileServer.ServeHTTP(w, r)
}

func (s *Server) SetupRoutes() *mux.Router {
	r := mux.NewRouter()
	
	// API endpoints
	r.HandleFunc("/api/examples", s.ListExamples).Methods("GET")
	r.HandleFunc("/api/generate/{example}", s.GenerateExample).Methods("POST")
	
	// Static file serving
	r.PathPrefix("/outputs/").HandlerFunc(s.ServeOutputs)
	
	// Serve the main HTML page
	r.HandleFunc("/", s.ServeIndex).Methods("GET")
	
	return r
}

func (s *Server) runCommand(command string, args ...string) error {
	cmd := exec.Command(command, args...)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("command failed: %s %v - %w", command, args, err)
	}
	return nil
}

func (s *Server) runCommandInDir(dir string, command string, args ...string) error {
	cmd := exec.Command(command, args...)
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("command failed: %s %v in %s - %w", command, args, dir, err)
	}
	return nil
}

func (s *Server) ServeIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(indexHTML))
}