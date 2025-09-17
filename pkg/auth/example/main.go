package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joeblew999/infra/pkg/auth"
	"github.com/joeblew999/infra/pkg/caddy"
	"github.com/joeblew999/infra/pkg/status"
	"github.com/nats-io/nats.go"
)

// Removed embedded fs - using pkg-level web directory now

const (
	HTTPPort     = 8082      // Internal HTTP port for auth service
	CaddyPort    = 8443      // External HTTPS port via Caddy
	TestUsername = "testuser"
)

var (
	authService *auth.AuthService
	nc          *nats.Conn
	forceFlag   = flag.Bool("force", false, "Kill processes using required ports before starting")
)

func init() {
	// Initialize user and session stores
	userStore := auth.NewInMemoryUserStore()
	var sessionStore auth.SessionStore

	// Initialize NATS (optional - fallback to in-memory if not available)
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = "nats://localhost:4222"
	}

	var err error
	nc, err = nats.Connect(natsURL)
	if err != nil {
		log.Printf("NATS not available, using in-memory sessions: %v", err)
		sessionStore = auth.NewInMemorySessionStore()
	} else {
		natsSessionStore, err := auth.NewNATSSessionStore(nc)
		if err != nil {
			log.Printf("NATS KV not available, using in-memory sessions: %v", err)
			sessionStore = auth.NewInMemorySessionStore()
		} else {
			sessionStore = natsSessionStore
		}
	}

	// Initialize complete auth service
	config := auth.WebAuthnConfig{
		RPDisplayName: "Auth Example",
		RPID:          "localhost",
		RPOrigins:     []string{fmt.Sprintf("https://localhost:%d", CaddyPort)},
	}

	webDir := "../web" // Relative to the example directory
	authService, err = auth.NewAuthService(config, userStore, sessionStore, webDir)
	if err != nil {
		log.Fatal(err)
	}
}


func checkPorts() error {
	ports := []int{HTTPPort, CaddyPort}
	
	for _, port := range ports {
		if !status.IsPortAvailable(port) {
			if *forceFlag {
				fmt.Printf("🔧 Killing process on port %d...\n", port)
				if err := status.KillProcessByPort(port); err != nil {
					return fmt.Errorf("failed to kill process on port %d: %v", port, err)
				}
				// Wait a moment for the port to be released
				if !status.WaitForPortAvailable(port, 5*time.Second) {
					return fmt.Errorf("port %d still not available after cleanup", port)
				}
			} else {
				pid := status.GetProcessByPort(port)
				return fmt.Errorf("port %d is already in use (PID: %s). Use --force to kill existing processes", port, pid)
			}
		}
	}
	return nil
}

func main() {
	flag.Parse()
	
	fmt.Println("🔐 WebAuthn Auth Example with HTTPS")
	fmt.Println("===================================")
	
	// Check and handle port conflicts
	if err := checkPorts(); err != nil {
		log.Fatal("Port conflict: ", err)
	}

	// Create Chi router
	r := chi.NewRouter()

	// Add middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Mount all auth routes on root using the auth service
	authService.RegisterRoutes(r)

	// Start the HTTP server in a goroutine
	go func() {
		fmt.Printf("🌐 Starting auth service on http://localhost:%d\n", HTTPPort)
		if err := http.ListenAndServe(fmt.Sprintf(":%d", HTTPPort), r); err != nil {
			log.Fatal("Auth server failed:", err)
		}
	}()

	// Configure and start Caddy HTTPS proxy
	config := caddy.NewPresetConfig(caddy.PresetSimple, CaddyPort).
		WithMainTarget(fmt.Sprintf("localhost:%d", HTTPPort))
	
	fmt.Printf("🔒 Starting Caddy HTTPS proxy on port %d\n", CaddyPort)
	fmt.Printf("🌍 WebAuthn will be available at: https://localhost:%d\n", CaddyPort)
	fmt.Println("💡 NATS URL:", os.Getenv("NATS_URL"))
	fmt.Println("")
	fmt.Println("✅ HTTPS is required for WebAuthn/passkeys to work!")
	fmt.Println("🎯 Open your browser to: https://localhost:" + fmt.Sprintf("%d", CaddyPort))
	fmt.Println("💡 Press Ctrl+C to stop all services")
	fmt.Println("")

	// Start Caddy with config generation and background startup
	caddy.StartWithConfig(&config)

	// Wait for interrupt signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	fmt.Println("\n🛑 Shutting down...")
}
