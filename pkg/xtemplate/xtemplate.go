package xtemplate

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/joeblew999/infra/pkg/config"
	"github.com/joeblew999/infra/pkg/dep"
	"github.com/joeblew999/infra/pkg/goreman"
	"github.com/joeblew999/infra/pkg/log"
	"github.com/joeblew999/infra/pkg/service"
)

func init() {
	// Register xtemplate service factory for decoupled access
	goreman.RegisterService("xtemplate", func() error {
		return StartSupervised()
	})
}

// Runner constructs the *exec.Cmd used to launch xtemplate.
type Runner func(context.Context, string, []string) *exec.Cmd

// Option configures a Service instance.
type Option func(*Service)

// Service manages the xtemplate web server for template-based web development.
type Service struct {
	templateDir    string
	port           string
	debug          bool
	minify         bool
	watchTemplates bool
	binaryPath     string
	runner         Runner
}

// NewService creates a new xtemplate service instance.
func NewService(opts ...Option) *Service {
	svc := &Service{
		templateDir:    config.GetXTemplatePath(),
		port:           config.GetXTemplatePort(),
		debug:          config.IsDevelopment(),
		minify:         true,
		watchTemplates: true,
		runner: func(ctx context.Context, bin string, args []string) *exec.Cmd {
			return exec.CommandContext(ctx, bin, args...)
		},
	}

	for _, opt := range opts {
		if opt != nil {
			opt(svc)
		}
	}

	return svc
}

// WithTemplateDir sets the template directory used by the service.
func WithTemplateDir(dir string) Option {
	return func(s *Service) {
		if dir != "" {
			s.templateDir = dir
		}
	}
}

// WithPort sets the listening port before startup.
func WithPort(port string) Option {
	return func(s *Service) {
		if port != "" {
			s.port = port
		}
	}
}

// WithBinaryPath overrides the xtemplate binary path (useful for tests).
func WithBinaryPath(path string) Option {
	return func(s *Service) {
		s.binaryPath = path
	}
}

// WithRunner overrides the command runner used to launch xtemplate.
func WithRunner(r Runner) Option {
	return func(s *Service) {
		if r != nil {
			s.runner = r
		}
	}
}

// WithDebug toggles debug logging.
func WithDebug(enabled bool) Option {
	return func(s *Service) {
		s.debug = enabled
	}
}

// WithMinify toggles HTML minification.
func WithMinify(enabled bool) Option {
	return func(s *Service) {
		s.minify = enabled
	}
}

// WithWatchTemplates toggles template watching.
func WithWatchTemplates(enabled bool) Option {
	return func(s *Service) {
		s.watchTemplates = enabled
	}
}

// SetPort overrides the listening port before Start is invoked (legacy helper).
func (s *Service) SetPort(port string) {
	if port != "" {
		s.port = port
	}
}

// SetTemplateDir overrides the template directory before Start is invoked (legacy helper).
func (s *Service) SetTemplateDir(dir string) {
	if dir != "" {
		s.templateDir = dir
	}
}

// Start starts the xtemplate server with the configured settings
func (s *Service) Start(ctx context.Context) error {
	log.Info("Starting xtemplate server", "template_dir", s.templateDir, "port", s.port)

	// Ensure template directory exists
	if err := os.MkdirAll(s.templateDir, 0755); err != nil {
		return fmt.Errorf("failed to create template directory: %w", err)
	}

	// Resolve xtemplate binary path
	binPath, err := s.resolveBinary()
	if err != nil {
		return err
	}

	// Create a basic index.html if templates directory is empty
	if err := s.ensureBasicTemplates(); err != nil {
		return fmt.Errorf("failed to setup basic templates: %w", err)
	}

	// Build xtemplate command arguments
	args := []string{
		"--template-dir", s.templateDir,
		"--listen", "0.0.0.0:" + s.port,
		"--minify=" + strconv.FormatBool(s.minify),
		"--watchtemplates=" + strconv.FormatBool(s.watchTemplates),
	}

	if s.debug {
		args = append(args, "--loglevel", "-2") // More verbose logging in development
	}

	// Start xtemplate server
	cmd := s.runner(ctx, binPath, args)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	log.Info("Executing xtemplate command", "binary", binPath, "args", args)
	return cmd.Run()
}

func (s *Service) ensureBasicTemplates() error {
	empty, err := isDirEffectivelyEmpty(s.templateDir)
	if err != nil {
		return err
	}
	if !empty {
		return nil
	}

	if err := s.copyEmbeddedDir("templates/upstream"); err == nil {
		if hasRenderableTemplates(s.templateDir) {
			log.Info("Seeded xtemplate templates from upstream bundle", "dir", s.templateDir)
			return nil
		}
		// No usable templates found; clear and fall back to local seeds.
		if err := removeDirContents(s.templateDir); err != nil {
			return err
		}
	}

	if err := s.copyEmbeddedDir("templates/seed"); err != nil {
		return err
	}
	log.Info("Seeded xtemplate templates from local bundle", "dir", s.templateDir)
	return nil
}

func (s *Service) copyEmbeddedDir(root string) error {
	if _, err := fs.Stat(assetsFS, root); err != nil {
		return err
	}

	return fs.WalkDir(assetsFS, root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if path == root {
			return nil
		}

		relPath := strings.TrimPrefix(path, root)
		relPath = strings.TrimPrefix(relPath, "/")
		if relPath == "" {
			return nil
		}

		targetPath := filepath.Join(s.templateDir, relPath)

		if d.IsDir() {
			return os.MkdirAll(targetPath, 0755)
		}

		if filepath.Base(path) == ".keep" {
			return nil
		}

		data, err := assetsFS.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(targetPath, data, 0644)
	})
}

func isDirEffectivelyEmpty(dir string) (bool, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false, err
	}
	for _, entry := range entries {
		if entry.Name() == ".keep" {
			continue
		}
		return false, nil
	}
	return true, nil
}

func hasRenderableTemplates(dir string) bool {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	for _, entry := range entries {
		if entry.IsDir() {
			if hasRenderableTemplates(filepath.Join(dir, entry.Name())) {
				return true
			}
			continue
		}
		ext := strings.ToLower(filepath.Ext(entry.Name()))
		if ext == ".html" || ext == ".templ" {
			return true
		}
	}
	return false
}

func removeDirContents(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		name := entry.Name()
		if name == ".keep" {
			continue
		}
		path := filepath.Join(dir, name)
		if err := os.RemoveAll(path); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) resolveBinary() (string, error) {
	if s.binaryPath != "" {
		return s.binaryPath, nil
	}

	if err := dep.InstallBinary(config.BinaryXtemplate, false); err != nil {
		return "", fmt.Errorf("failed to ensure xtemplate binary: %w", err)
	}

	return config.GetXTemplateBinPath(), nil
}

// Stop stops the xtemplate server (handled by context cancellation)
func (s *Service) Stop() error {
	log.Info("Stopping xtemplate server")
	return nil
}

// GetURL returns the local URL where xtemplate server is accessible
func (s *Service) GetURL() string {
	return config.FormatLocalHTTP(s.port)
}

// GetTemplateDir returns the templates directory path
func (s *Service) GetTemplateDir() string {
	return s.templateDir
}

// StartSupervised starts xtemplate under goreman supervision (idempotent)
func StartSupervised() error {
	// Ensure template directory exists
	templateDir := config.GetXTemplatePath()
	if err := os.MkdirAll(templateDir, 0755); err != nil {
		return fmt.Errorf("failed to create template directory: %w", err)
	}

	svc := NewService()
	svc.SetTemplateDir(templateDir)
	if err := svc.ensureBasicTemplates(); err != nil {
		return fmt.Errorf("failed to seed xtemplate templates: %w", err)
	}

	binPath, err := svc.resolveBinary()
	if err != nil {
		return err
	}

	// Build xtemplate command arguments using config
	args := []string{
		"--template-dir", templateDir,
		"--listen", "0.0.0.0:" + config.GetXTemplatePort(),
		"--minify=true",
		"--watchtemplates=true", // Enable live reload for development
	}

	if config.IsDevelopment() {
		args = append(args, "--loglevel", "-2") // More verbose logging in development
	}

	// Register and start with goreman supervision
	processCfg := service.NewConfig(binPath, args)
	return service.Start("xtemplate", processCfg)
}
