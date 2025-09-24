package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/joeblew999/infra/pkg/config"
	"github.com/joeblew999/infra/pkg/xtemplate"
	"github.com/spf13/cobra"
)

// Register attaches the XTemplate command tree to the provided parent.
func Register(parent *cobra.Command) {
	parent.AddCommand(NewCommand())
}

// NewCommand constructs the root command for XTemplate operations.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "xtemplate",
		Short: "HTML/template-based rapid web development server",
		Long: `XTemplate provides rapid web development with HTML/template-based preprocessing.

Features:
- Live reload development server
- File-based routing (index.html → /)
- Custom route definitions with Go 1.22 ServeMux patterns
- Template-based responses with dynamic context
- Built-in asset serving with caching and compression
- Server-sent events for real-time updates`,
	}

	cmd.AddCommand(
		newServeCommand(),
		newInitCommand(),
		newVersionCommand(),
		newUpstreamCommand(),
	)

	return cmd
}

func newServeCommand() *cobra.Command {
	serveCmd := &cobra.Command{
		Use:   "serve [template-dir]",
		Short: "Start xtemplate development server",
		Long: `Start the xtemplate development server with live reload enabled.

Templates are served from the specified directory (or default data/xtemplate).
The server will automatically reload when template files change.`,
		Args: cobra.MaximumNArgs(1),
		RunE: runXTemplateServe,
	}

	serveCmd.Flags().StringP("port", "p", config.GetXTemplatePort(), "Port to serve on")
	serveCmd.Flags().Bool("debug", config.IsDevelopment(), "Enable debug logging")
	serveCmd.Flags().Bool("no-reload", false, "Disable live reload")
	serveCmd.Flags().Bool("minify", true, "Minify HTML output")

	return serveCmd
}

func newInitCommand() *cobra.Command {
	initCmd := &cobra.Command{
		Use:   "init [template-dir]",
		Short: "Initialize xtemplate project with sample templates",
		Long: `Initialize a new xtemplate project with sample template files.

This creates a basic project structure with:
- index.html (homepage)
- about.html (sample page)
- shared/ directory for common templates
- assets/ directory for static files`,
		Args: cobra.MaximumNArgs(1),
		RunE: runXTemplateInit,
	}

	return initCmd
}

func newVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show xtemplate version",
		RunE:  showXTemplateVersion,
	}
}

func runXTemplateServe(cmd *cobra.Command, args []string) error {
	// Resolve template directory from argument or config.
	templateDir := config.GetXTemplatePath()
	if len(args) > 0 {
		templateDir = args[0]
	}

	// Hydrate flags.
	port, _ := cmd.Flags().GetString("port")
	debug, _ := cmd.Flags().GetBool("debug")
	noReload, _ := cmd.Flags().GetBool("no-reload")
	minify, _ := cmd.Flags().GetBool("minify")

	if err := os.MkdirAll(templateDir, 0755); err != nil {
		return fmt.Errorf("failed to create template directory: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	service := xtemplate.NewService(
		xtemplate.WithTemplateDir(templateDir),
		xtemplate.WithPort(port),
		xtemplate.WithDebug(debug),
		xtemplate.WithWatchTemplates(!noReload),
		xtemplate.WithMinify(minify),
	)

	deps, err := xtemplate.EnsureRuntime(ctx)
	if err != nil {
		return err
	}
	defer deps.Cleanup()

	if port != config.GetXTemplatePort() {
		fmt.Printf("Custom port %s will be used\n", port)
	}

	fmt.Printf("🚀 Starting xtemplate server...\n")
	fmt.Printf("📁 Template directory: %s\n", templateDir)
	fmt.Printf("🌐 Server URL: %s\n", config.FormatLocalHTTP(port))
	fmt.Printf("🔗 Browser: http://127.0.0.1:%s\n", port)
	fmt.Printf("🔄 Live reload: %t\n", !noReload)
	fmt.Printf("📦 Minify: %t\n", minify)
	fmt.Printf("🐛 Debug: %t\n", debug)
	fmt.Printf("⛔ Stop with CTRL+C\n")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigChan)

	go func() {
		<-sigChan
		fmt.Println("\n🛑 Shutting down xtemplate server...")
		cancel()
	}()

	if err := service.Start(ctx); err != nil {
		if ctx.Err() != nil {
			fmt.Println("✅ XTemplate server stopped gracefully")
			return nil
		}
		return fmt.Errorf("xtemplate server error: %w", err)
	}

	return nil
}

func runXTemplateInit(cmd *cobra.Command, args []string) error {
	templateDir := config.GetXTemplatePath()
	if len(args) > 0 {
		templateDir = args[0]
	}

	if err := os.MkdirAll(templateDir, 0755); err != nil {
		return fmt.Errorf("failed to create template directory: %w", err)
	}

	if err := createSampleProject(templateDir); err != nil {
		return fmt.Errorf("failed to initialize project: %w", err)
	}

	fmt.Printf("✅ XTemplate project initialized at: %s\n", templateDir)
	fmt.Printf("🚀 Run 'go run . tools xtemplate serve' to start development server\n")
	fmt.Printf("🌐 Visit %s to view your site\n", config.FormatLocalHTTP(config.GetXTemplatePort()))

	return nil
}

func showXTemplateVersion(cmd *cobra.Command, args []string) error {
	binPath := config.GetXTemplateBinPath()

	execCmd := exec.Command(binPath, "--version")
	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr
	return execCmd.Run()
}

func createSampleProject(templateDir string) error {
	dirs := []string{
		filepath.Join(templateDir, "shared"),
		filepath.Join(templateDir, "assets", "css"),
		filepath.Join(templateDir, "assets", "js"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	templates := map[string]string{
		"index.html": `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.X.Title | default "XTemplate Project"}}</title>
    <link rel="stylesheet" href="/assets/css/style.css">
</head>
<body>
    {{template "/shared/header.html" .}}
    
    <main>
        <h1>Welcome to XTemplate! 🚀</h1>
        <p>This is your homepage. Edit <code>index.html</code> to customize it.</p>
        
        <div class="features">
            <div class="feature">
                <h3>✨ Live Reload</h3>
                <p>Changes to templates are automatically reflected in the browser.</p>
            </div>
            <div class="feature">
                <h3>🎯 File-based Routing</h3>
                <p>Create <code>about.html</code> and visit <code>/about</code> to see it.</p>
            </div>
            <div class="feature">
                <h3>🎨 Template Context</h3>
                <p>Access request data: <strong>{{.Req.URL.Path}}</strong></p>
            </div>
        </div>
    </main>
    
    {{template "/shared/footer.html" .}}
    <script src="/assets/js/app.js"></script>
    
    {{- if eq .X.DevMode true}}
    <script>new EventSource('/reload').onmessage = () => location.reload()</script>
    {{- end}}
</body>
</html>`,

		"about.html": `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>About - XTemplate Project</title>
    <link rel="stylesheet" href="/assets/css/style.css">
</head>
<body>
    {{template "/shared/header.html" .}}
    
    <main>
        <h1>About This Project</h1>
        <p>This is a sample about page created with XTemplate.</p>
        
        <h2>Template Features</h2>
        <ul>
            <li>HTML/template syntax with Go functions</li>
            <li>Built-in context providers (database, filesystem, etc.)</li>
            <li>Custom route definitions using ServeMux patterns</li>
            <li>Server-sent events for real-time updates</li>
        </ul>
    </main>
    
    {{template "/shared/footer.html" .}}
    <script src="/assets/js/app.js"></script>
</body>
</html>`,

		"shared/header.html": `<header>
    <nav>
        <div class="logo">XTemplate Project</div>
        <ul>
            <li><a href="/">Home</a></li>
            <li><a href="/about">About</a></li>
        </ul>
    </nav>
</header>`,

		"shared/footer.html": `<footer>
    <p>&copy; 2024 XTemplate Project. Built with ❤️ using XTemplate.</p>
</footer>

{{- define "SSE /reload"}}
{{.WaitForServerStop}}data: reload{{printf "\n\n"}}
{{- end}}`,

		"assets/css/style.css": `/* XTemplate Project Styles */
* {
    margin: 0;
    padding: 0;
    box-sizing: border-box;
}

body {
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
    line-height: 1.6;
    color: #333;
    background: #f5f5f5;
}

header {
    background: #007bff;
    color: white;
    padding: 1rem 0;
    position: sticky;
    top: 0;
    z-index: 100;
}

nav {
    max-width: 1200px;
    margin: 0 auto;
    padding: 0 1rem;
    display: flex;
    justify-content: space-between;
    align-items: center;
}

.logo {
    font-size: 1.5rem;
    font-weight: bold;
}

nav ul {
    list-style: none;
    display: flex;
    gap: 2rem;
}

nav a {
    color: white;
    text-decoration: none;
    transition: opacity 0.3s;
}

nav a:hover {
    opacity: 0.8;
}

main {
    max-width: 1200px;
    margin: 2rem auto;
    padding: 0 1rem;
    background: white;
    border-radius: 8px;
    box-shadow: 0 2px 10px rgba(0,0,0,0.1);
    padding: 2rem;
}

h1 {
    color: #007bff;
    margin-bottom: 1rem;
}

.features {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
    gap: 2rem;
    margin: 2rem 0;
}

.feature {
    padding: 1.5rem;
    background: #f8f9fa;
    border-radius: 8px;
    border-left: 4px solid #007bff;
}

.feature h3 {
    color: #007bff;
    margin-bottom: 0.5rem;
}

code {
    background: #e9ecef;
    padding: 0.2rem 0.4rem;
    border-radius: 3px;
    font-family: 'Monaco', 'Consolas', monospace;
}

footer {
    text-align: center;
    padding: 2rem;
    color: #666;
    background: white;
    margin: 2rem auto;
    max-width: 1200px;
    border-radius: 8px;
}

@media (max-width: 768px) {
    nav {
        flex-direction: column;
        gap: 1rem;
    }
    
    nav ul {
        gap: 1rem;
    }
}`,

		"assets/js/app.js": `// XTemplate Project JavaScript
console.log('🚀 XTemplate project loaded!');

// Add any interactive functionality here
document.addEventListener('DOMContentLoaded', function() {
    // Example: highlight current page in navigation
    const currentPath = window.location.pathname;
    const navLinks = document.querySelectorAll('nav a');
    
    navLinks.forEach(link => {
        if (link.getAttribute('href') === currentPath || 
            (currentPath === '/' && link.getAttribute('href') === '/')) {
            link.style.textDecoration = 'underline';
        }
    });
});`,
	}

	for filename, content := range templates {
		filePath := filepath.Join(templateDir, filename)
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to create %s: %w", filename, err)
		}
	}

	return nil
}
