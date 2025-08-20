package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joeblew999/infra/pkg/caddy"
)

func main() {
	fmt.Println("Caddy Package Example")
	fmt.Println("=====================")

	// Example 1: Default configuration (bento playground + main app)
	fmt.Println("\n1. Default Configuration:")
	defaultConfig := caddy.DefaultConfig()
	defaultCaddyfile := caddy.GenerateCaddyfile(defaultConfig)
	fmt.Printf("Port: %d\nTarget: %s\nRoutes: %d\n", 
		defaultConfig.Port, defaultConfig.Target, len(defaultConfig.Routes))
	fmt.Printf("Generated Caddyfile:\n%s\n", defaultCaddyfile)

	// Example 2: Using Presets for Common Scenarios
	fmt.Println("\n2. Preset Configurations:")
	
	// Development preset (main app + bento playground)
	devConfig := caddy.NewPresetConfig(caddy.PresetDevelopment, 8080)
	fmt.Printf("Development Preset: %d routes\n", len(devConfig.Routes))
	fmt.Printf("Generated Caddyfile:\n%s\n", caddy.GenerateCaddyfile(devConfig))
	
	// Full preset (main app + bento + MCP)
	fullConfig := caddy.NewPresetConfig(caddy.PresetFull, 8080)
	fmt.Printf("Full Preset: %d routes\n", len(fullConfig.Routes))
	
	// Custom microservices preset with modifications
	customConfig := caddy.NewPresetConfig(caddy.PresetMicroservices, 8080).
		WithMainTarget("localhost:3000").
		AddBentoPlayground()
	fmt.Printf("Custom Microservices: %d routes\n", len(customConfig.Routes))

	// Example 3: Simple preset configurations
	fmt.Println("\n3. Simple Preset Examples:")
	
	// Simple single app
	simpleConfig := caddy.NewPresetConfig(caddy.PresetSimple, 9000)
	fmt.Printf("Simple Preset (single app):\n%s\n", caddy.GenerateCaddyfile(simpleConfig))
	
	// Quick convenience functions
	fmt.Println("4. Convenience Functions:")
	fmt.Println("   caddy.StartDevelopmentServer(8080) // Starts with bento")
	fmt.Println("   caddy.StartFullServer(8080)        // Starts with bento + MCP")

	// Example 5: Using the runner to start mock services
	fmt.Println("\n5. Starting Mock Services and Caddy:")
	
	// Start mock HTTP services
	go startMockService(3000, "Main App")
	go startMockService(4000, "API Service")
	go startMockService(5000, "Auth Service")
	go startMockService(6000, "Static Files")
	go startMockService(7000, "WebSocket Service")

	// Create caddy runner (downloads caddy binary if needed)
	runner := caddy.New()
	
	// Generate and save Caddyfile using production pattern (.data/caddy/)
	fmt.Println("Generating Caddyfile...")
	if err := customConfig.GenerateAndSave("Caddyfile"); err != nil {
		fmt.Printf("Error writing Caddyfile: %v\n", err)
	} else {
		fmt.Printf("✅ Caddyfile saved to .data/caddy/ (production pattern)\n")
	}
	
	fmt.Println("\n🚀 To start Caddy with HTTPS, run:")
	fmt.Println("   .dep/caddy run --config .data/caddy/Caddyfile")
	fmt.Println("\n🌐 Then visit:")
	fmt.Println("   https://localhost:8080 (HTTPS with automatic certs)")
	fmt.Println("   https://localhost:8080/api/ (API service)")
	fmt.Println("   https://localhost:8080/auth/ (Auth service)") 
	fmt.Println("   https://localhost:8080/static/ (Static files)")
	fmt.Println("   https://localhost:8080/ws/ (WebSocket service)")
	
	// Actually start Caddy with HTTPS
	fmt.Println("\n🔒 Starting Caddy with HTTPS...")
	fmt.Println("💡 Press Ctrl+C to stop all services")
	fmt.Println("")
	
	runner.StartInBackground(".data/caddy/Caddyfile")

	// Give Caddy a moment to start
	time.Sleep(2 * time.Second)
	
	fmt.Println("🎉 All services are now running!")
	fmt.Println("🌐 Open your browser to:")
	fmt.Println("   https://localhost:8080")
	fmt.Println("")
	fmt.Println("📝 This example demonstrates:")
	fmt.Println("   ✅ Automatic caddy binary installation")
	fmt.Println("   ✅ Configurable reverse proxy setup")
	fmt.Println("   ✅ HTTPS with automatic certificates")
	fmt.Println("   ✅ Multi-service routing")
	fmt.Println("   ✅ Production pattern using .data/caddy/ directory")
	fmt.Println("   ✅ Docker-ready persistent configuration")

	// Wait for interrupt signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	fmt.Println("\nShutting down...")
}

// startMockService starts a simple HTTP server for demonstration
func startMockService(port int, name string) {
	mux := http.NewServeMux()
	
	// Serve static files for main app
	if name == "Main App" {
		fs := http.FileServer(http.Dir("./web"))
		mux.Handle("/", fs)
	} else {
		// Service-specific routes
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			// Serve the appropriate HTML page based on service
			var filename string
			switch name {
			case "API Service":
				filename = "./web/api.html"
			case "Auth Service":
				filename = "./web/auth.html"
			case "Static Files":
				filename = "./web/static.html"
			case "WebSocket Service":
				filename = "./web/ws.html"
			default:
				fmt.Fprintf(w, "Hello from %s on port %d!\n", name, port)
				return
			}
			
			http.ServeFile(w, r, filename)
		})
	}
	
	// Health check endpoint for all services
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"ok","service":"%s","port":%d}`, name, port)
	})
	
	fmt.Printf("Starting %s on :%d\n", name, port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), mux); err != nil {
		log.Printf("Mock service %s failed: %v", name, err)
	}
}