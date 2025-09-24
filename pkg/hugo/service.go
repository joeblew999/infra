package hugo

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/joeblew999/infra/pkg/config"
	"github.com/joeblew999/infra/pkg/dep"
	"github.com/joeblew999/infra/pkg/goreman"
	"github.com/joeblew999/infra/pkg/log"
	"github.com/joeblew999/infra/pkg/service"
)

func init() {
	// Register Hugo service factory for decoupled access
	goreman.RegisterService("hugo", func() error {
		return StartSupervised()
	})
}

// Service manages Hugo site generation and development server
type Service struct {
	devMode   bool
	hugoPath  string
	sourceDir string
	outputDir string
	baseURL   string
	port      string
}

// NewService creates a new Hugo service instance
func NewService(devMode bool, sourceDir string) *Service {
	// Get Hugo binary path and port from config
	hugoPath := config.Get(config.BinaryHugo)

	// Ensure absolute path for Hugo binary
	if !filepath.IsAbs(hugoPath) {
		abs, err := filepath.Abs(hugoPath)
		if err == nil {
			hugoPath = abs
		}
	}

	outputDir := filepath.Join(sourceDir, "public")

	// Get configuration
	cfg := config.GetConfig()
	port := cfg.Ports.Hugo
	baseURL := config.FormatLocalHTTP(port)

	return &Service{
		devMode:   devMode,
		hugoPath:  hugoPath,
		sourceDir: sourceDir,
		outputDir: outputDir,
		baseURL:   baseURL,
		port:      port,
	}
}

// Start starts the Hugo service (dev server or builds static site)
func (s *Service) Start(ctx context.Context) error {
	// Ensure Hugo is available
	if err := s.ensureHugo(); err != nil {
		return fmt.Errorf("hugo not available: %w", err)
	}

	// Initialize Hugo site if needed
	if err := s.initializeSite(); err != nil {
		return fmt.Errorf("failed to initialize Hugo site: %w", err)
	}

	if s.devMode {
		log.Info("Starting Hugo development server", "port", s.port, "source", s.sourceDir)
		return s.startDevServer(ctx)
	}

	log.Info("Building Hugo static site", "source", s.sourceDir, "output", s.outputDir)
	return s.buildSite(ctx)
}

// startDevServer starts Hugo in development mode with live reload
func (s *Service) startDevServer(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, s.hugoPath, "server",
		"--source", s.sourceDir,
		"--port", s.port,
		"--baseURL", config.FormatLocalHTTP(s.port)+"/hugo",
		"--appendPort=false",
		"--watch",
		"--disableFastRender",
		"--buildDrafts",
		"--buildFuture",
		"--forceSyncStatic",
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = s.sourceDir

	return cmd.Run()
}

// buildSite builds the static site for production
func (s *Service) buildSite(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, s.hugoPath,
		"--source", s.sourceDir,
		"--destination", s.outputDir,
		"--minify",
		"--gc",
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = s.sourceDir

	return cmd.Run()
}

// initializeSite creates a Hugo site structure if it doesn't exist
func (s *Service) initializeSite() error {
	configFile := filepath.Join(s.sourceDir, "hugo.yaml")
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		log.Info("Initializing new Hugo site", "path", s.sourceDir)

		// Create basic Hugo structure
		if err := os.MkdirAll(s.sourceDir, 0755); err != nil {
			return err
		}

		// Create basic hugo.yaml config
		config := fmt.Sprintf(`baseURL: '%s'
title: 'Infrastructure Documentation'
theme: 'hugoplate'

params:
  title: 'Infrastructure Documentation'
  description: 'Comprehensive documentation for infrastructure management system'
  logo: '/images/logo.png'

menu:
  main:
    - name: 'Home'
      url: '/'
      weight: 1
    - name: 'Business'
      url: '/business/'
      weight: 2
    - name: 'Technical'
      url: '/technical/'
      weight: 3
    - name: 'Examples'
      url: '/examples/'
      weight: 4

markup:
  goldmark:
    renderer:
      unsafe: true
  highlight:
    style: github
    lineNos: true
`, s.baseURL)

		return os.WriteFile(configFile, []byte(config), 0644)
	}
	return nil
}

// ensureHugo checks if Hugo binary is available
func (s *Service) ensureHugo() error {
	if _, err := os.Stat(s.hugoPath); os.IsNotExist(err) {
		return fmt.Errorf("hugo binary not found at %s", s.hugoPath)
	}

	// Test Hugo version
	cmd := exec.Command(s.hugoPath, "version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("hugo binary not executable: %w", err)
	}

	return nil
}

// Stop gracefully stops the Hugo service
func (s *Service) Stop(ctx context.Context) error {
	// Hugo dev server will be stopped by context cancellation
	return nil
}

// GetOutputDir returns the directory where Hugo builds static files
func (s *Service) GetOutputDir() string {
	return s.outputDir
}

// IsReady checks if the Hugo service is ready to serve content
func (s *Service) IsReady() bool {
	if s.devMode {
		// Check if dev server is responding
		return s.checkDevServer()
	}
	// Check if static files exist
	return s.checkStaticFiles()
}

// checkDevServer verifies the Hugo dev server is responding
func (s *Service) checkDevServer() bool {
	// Simple check - we could add HTTP health check here
	return true
}

// checkStaticFiles verifies static site has been built
func (s *Service) checkStaticFiles() bool {
	indexFile := filepath.Join(s.outputDir, "index.html")
	_, err := os.Stat(indexFile)
	return err == nil
}

// StartSupervised starts Hugo under goreman supervision (package-level function)
func StartSupervised() error {
	cfg := config.GetConfig()
	docsDir := "docs-hugo"

	// Ensure Hugo site exists before starting
	ctx := context.Background()
	if err := EnsureHugoSite(ctx, docsDir); err != nil {
		return fmt.Errorf("failed to prepare Hugo site: %w", err)
	}

	if err := dep.InstallBinary(config.BinaryHugo, false); err != nil {
		return fmt.Errorf("failed to ensure hugo binary: %w", err)
	}

	// Get Hugo binary path using config
	binPath := config.Get(config.BinaryHugo)

	// Build Hugo command arguments using config
	args := []string{
		"server",
		"--source", docsDir,
		"--port", cfg.Ports.Hugo,
		"--baseURL", config.FormatLocalHTTP(cfg.Ports.Hugo) + "/docs-hugo",
		"--appendPort=false",
		"--watch",
		"--disableFastRender",
		"--buildDrafts",
		"--buildFuture",
		"--forceSyncStatic",
	}

	processCfg := service.NewConfig(binPath, args)
	return service.Start("hugo", processCfg)
}

// EnsureHugoSite ensures Hugo site exists before starting service
func EnsureHugoSite(ctx context.Context, docsDir string) error {
	// Check if Hugo site directory exists
	if _, err := os.Stat(docsDir); os.IsNotExist(err) {
		// Initialize site if it doesn't exist
		service := NewService(true, docsDir)
		if err := service.initializeSite(); err != nil {
			return fmt.Errorf("failed to initialize Hugo site: %w", err)
		}

		// Install theme if needed
		themeManager := NewThemeManager(docsDir)
		if err := themeManager.InstallTheme(); err != nil {
			return fmt.Errorf("failed to install theme: %w", err)
		}
	}

	return nil
}
