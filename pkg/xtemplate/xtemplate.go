package xtemplate

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/joeblew999/infra/pkg/config"
	"github.com/joeblew999/infra/pkg/goreman"
	"github.com/joeblew999/infra/pkg/log"
)

func init() {
	// Register xtemplate service factory for decoupled access
	goreman.RegisterService("xtemplate", func() error {
		return StartSupervised()
	})
}

// Service manages the xtemplate web server for template-based web development
type Service struct {
	templateDir string
	port        string
	debug       bool
}

// NewService creates a new xtemplate service instance
func NewService() *Service {
	return &Service{
		templateDir: config.GetXTemplatePath(),
		port:        config.GetXTemplatePort(),
		debug:       config.IsDevelopment(),
	}
}

// Start starts the xtemplate server with the configured settings
func (s *Service) Start(ctx context.Context) error {
	log.Info("Starting xtemplate server", "template_dir", s.templateDir, "port", s.port)

	// Ensure template directory exists
	if err := os.MkdirAll(s.templateDir, 0755); err != nil {
		return fmt.Errorf("failed to create template directory: %w", err)
	}

	// Create a basic index.html if templates directory is empty
	if err := s.ensureBasicTemplates(); err != nil {
		return fmt.Errorf("failed to setup basic templates: %w", err)
	}

	// Get xtemplate binary path
	binPath := config.GetXTemplateBinPath()

	// Build xtemplate command arguments
	args := []string{
		"--template-dir", s.templateDir,
		"--listen", "0.0.0.0:" + s.port,
		"--minify=true",
		"--watchtemplates=true", // Enable live reload for development
	}

	if s.debug {
		args = append(args, "--loglevel", "-2") // More verbose logging in development
	}

	// Start xtemplate server
	cmd := exec.CommandContext(ctx, binPath, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	log.Info("Executing xtemplate command", "binary", binPath, "args", args)
	return cmd.Run()
}

// ensureBasicTemplates creates basic template files if the templates directory is empty
func (s *Service) ensureBasicTemplates() error {
	// Check if index.html exists
	indexPath := filepath.Join(s.templateDir, "index.html")
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		// Create a basic index.html template
		indexContent := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>XTemplate Development Server</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif; 
               margin: 2rem; background: #f5f5f5; }
        .container { max-width: 800px; margin: 0 auto; background: white; 
                     padding: 2rem; border-radius: 8px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        h1 { color: #333; margin-top: 0; }
        .feature { margin: 1rem 0; padding: 1rem; background: #f8f9fa; 
                   border-left: 4px solid #007bff; border-radius: 4px; }
        code { background: #e9ecef; padding: 0.2rem 0.4rem; border-radius: 3px; 
               font-family: 'Monaco', 'Consolas', monospace; }
        .reload-script { margin-top: 2rem; padding: 1rem; background: #fff3cd; 
                         border: 1px solid #ffeaa7; border-radius: 4px; }
    </style>
</head>
<body>
    <div class="container">
        <h1>🚀 XTemplate Development Server</h1>
        
        <p>Welcome to your xtemplate development environment! This server provides 
        rapid web development with HTML/template-based preprocessing.</p>
        
        <div class="feature">
            <h3>✨ Live Reload Enabled</h3>
            <p>Templates automatically reload when you modify files. No need to restart the server!</p>
        </div>
        
        <div class="feature">
            <h3>📁 Template Directory</h3>
            <p>Templates are served from: <code>{{.X.TemplateDir}}</code></p>
        </div>
        
        <div class="feature">
            <h3>🎯 Getting Started</h3>
            <ul>
                <li>Edit this file: <code>{{.X.TemplateDir}}/index.html</code></li>
                <li>Create new templates like <code>about.html</code> → accessible at <code>/about</code></li>
                <li>Use Go template syntax with context data: <code>{{.Req.URL.Path}}</code></li>
            </ul>
        </div>
        
        <div class="feature">
            <h3>📚 Documentation</h3>
            <p>Learn more at <a href="https://github.com/infogulch/xtemplate">github.com/infogulch/xtemplate</a></p>
        </div>
        
        <div class="reload-script">
            <h4>🔄 Auto-Reload Script (Development Only)</h4>
            <p>This script automatically reloads the page when templates change:</p>
        </div>
    </div>

    <!-- Auto-reload script for development -->
    {{- if .X.DevMode}}
    <script>
        new EventSource('/reload').onmessage = () => location.reload();
        console.log('🔄 Auto-reload enabled - page will refresh when templates change');
    </script>
    {{- end}}
    
    <!-- Server-sent events endpoint for auto-reload -->
    {{- define "SSE /reload"}}
    {{.WaitForServerStop}}data: reload{{printf "\n\n"}}
    {{- end}}
</body>
</html>`

		if err := os.WriteFile(indexPath, []byte(indexContent), 0644); err != nil {
			return fmt.Errorf("failed to create index.html: %w", err)
		}

		log.Info("Created basic index.html template", "path", indexPath)
	}

	return nil
}

// Stop stops the xtemplate server (handled by context cancellation)
func (s *Service) Stop() error {
	log.Info("Stopping xtemplate server")
	return nil
}

// GetURL returns the local URL where xtemplate server is accessible
func (s *Service) GetURL() string {
	return fmt.Sprintf("http://localhost:%s", s.port)
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

	// Get xtemplate binary path using config
	binPath := config.GetXTemplateBinPath()

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
	return goreman.RegisterAndStart("xtemplate", &goreman.ProcessConfig{
		Command:    binPath,
		Args:       args,
		WorkingDir: ".",
		Env:        os.Environ(),
	})
}