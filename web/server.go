package web

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/delaneyj/toolbelt/embeddednats"
	"github.com/go-chi/chi/v5"
	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
	"github.com/starfederation/datastar-go/datastar"
	"github.com/yuin/goldmark"
	goldmark_parser "github.com/yuin/goldmark/parser"
	goldmark_renderer_html "github.com/yuin/goldmark/renderer/html"

	"github.com/joeblew999/infra/pkg/embeds"
	"github.com/joeblew999/infra/pkg/store"
)

//go:embed index.html
var helloWorldHTML []byte

const port = 1337

type App struct {
	natsConn *nats.Conn
	router   *chi.Mux
}

type Store struct {
	Delay time.Duration `json:"delay"`
}

type Message struct {
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

func StartServer(devDocs bool) error {
	ctx := context.Background()

	// Configure NATS server options for logging
	natsOpts := &server.Options{
		Debug: true, // Enable debug logging
		Trace: true, // Enable trace logging
	}

	// Initialize embedded NATS server
	log.Println("Starting embedded NATS server...")
	natsServer, err := embeddednats.New(ctx,
		embeddednats.WithDirectory("./.data/nats"), // Store directory
		embeddednats.WithNATSServerOptions(natsOpts),
	)
	if err != nil {
		log.Printf("Failed to create embedded NATS server: %v", err)
		return fmt.Errorf("Failed to create embedded NATS server: %w", err)
	}
	defer natsServer.Close()

	// Wait for the server to be ready
	natsServer.WaitForServer()
	log.Printf("Embedded NATS server started")

	// Get client connection from the embedded server
	nc, err := natsServer.Client()
	if err != nil {
		return fmt.Errorf("Failed to get NATS client: %w", err)
	}
	defer nc.Close()

	app := &App{
		natsConn: nc,
		router:   chi.NewRouter(),
	}

	app.setupRoutes(devDocs)
	if err := app.setupNATSStreams(ctx); err != nil {
		return fmt.Errorf("Failed to setup NATS streams: %w", err)
	}

	log.Printf("Starting web server on http://localhost:%d", port)
	log.Printf("Embedded NATS server running with JetStream enabled")

	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), app.router); err != nil {
		return fmt.Errorf("Failed to start web server: %w", err)
	}
	return nil
}

func (app *App) setupRoutes(devDocs bool) {
	app.router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write(helloWorldHTML)
	})

	// Original hello-world endpoint
	app.router.Get("/hello-world", func(w http.ResponseWriter, r *http.Request) {
		store := &Store{}
		if err := datastar.ReadSignals(r, store); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		sse := datastar.NewSSE(w, r)
		const message = "Hello, world!"

		for i := 0; i < len(message); i++ {
			if err := sse.PatchElements(`<div id="message">` + message[:i+1] + `</div>`); err != nil {
				return
			}
			time.Sleep(store.Delay * time.Millisecond)
		}
	})

	// New NATS-powered endpoint
	app.router.Get("/nats-stream", func(w http.ResponseWriter, r *http.Request) {
		app.handleNATSStream(w, r)
	})

	// Endpoint to publish messages to NATS
	app.router.Post("/publish", func(w http.ResponseWriter, r *http.Request) {
		app.handlePublish(w, r)
	})

	// Docs handler
	app.router.Get(store.DocsHTTPPath+"*", app.handleDocs(devDocs))
}

func (app *App) handleDocs(devDocs bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		filePath := r.URL.Path[len(store.DocsHTTPPath):]
		log.Printf("Requested filePath: %s", filePath)

		// Debugging: List files in embeds.RootFS
		fs.WalkDir(embeds.RootFS, ".", func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			log.Printf("RootFS contains: %s", path)
			return nil
		})

		var content []byte
		var err error

		if devDocs {
			fullPath := filepath.Join(store.DocsDir, filePath)
			log.Printf("Serving from disk. Full path: %s", fullPath)
			content, err = os.ReadFile(fullPath)
		} else {
			log.Printf("Serving from embedded. FilePath: %s", filePath)
			docsFS, subErr := fs.Sub(embeds.RootFS, store.DocsDir)
			if subErr != nil {
				log.Printf("Error getting sub-filesystem: %v", subErr)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
			// Debugging: List files in docsFS
			fs.WalkDir(docsFS, ".", func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					return err
				}
				log.Printf("docsFS contains: %s", path)
				return nil
			})
			content, err = fs.ReadFile(docsFS, filePath)
		}

		if err != nil {
			log.Printf("Error reading document %s: %v", filePath, err)
			http.Error(w, "Document not found", http.StatusNotFound)
			return
		}

		// Render Markdown to HTML
		md := goldmark.New(
			goldmark.WithParserOptions(
				goldmark_parser.WithAutoHeadingID(),
			),
			goldmark.WithRendererOptions(
				goldmark_renderer_html.WithUnsafe(),
			),
		)
		var buf strings.Builder
		if err := md.Convert(content, &buf); err != nil {
			http.Error(w, "Failed to render document", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte("<!DOCTYPE html><html><head><title>Docs</title></head><body>"))
		w.Write([]byte(buf.String()))
		w.Write([]byte("</body></html>"))
	}
}

func (app *App) setupNATSStreams(ctx context.Context) error {
	// Create JetStream context
	js, err := app.natsConn.JetStream()
	if err != nil {
		log.Printf("Failed to create JetStream context: %v", err)
		return fmt.Errorf("Failed to create JetStream context: %w", err)
	}

	// Create a stream for messages
	streamName := "MESSAGES"
	_, err = js.AddStream(&nats.StreamConfig{
		Name:      streamName,
		Subjects:  []string{"messages.*"},
		Retention: nats.LimitsPolicy,
		MaxAge:    time.Hour * 24, // Keep messages for 24 hours
	})
	if err != nil {
		log.Printf("Stream may already exist: %v", err)
	}

	log.Printf("NATS JetStream setup complete - Stream: %s", streamName)
	return nil
}

func (app *App) handleNATSStream(w http.ResponseWriter, r *http.Request) {
	sse := datastar.NewSSE(w, r)

	// Subscribe to NATS messages
	sub, err := app.natsConn.Subscribe("messages.demo", func(msg *nats.Msg) {
		var message Message
		if err := json.Unmarshal(msg.Data, &message); err != nil {
			log.Printf("Error unmarshaling message: %v", err)
			return
		}

		// Send the message to the browser via DataStar
		html := fmt.Sprintf(`<div id="messages" data-store='{"timestamp":"%s"}'>%s</div>`,
			message.Timestamp.Format(time.RFC3339), message.Content)

		if err := sse.PatchElements(html); err != nil {
			log.Printf("Error sending SSE: %v", err)
		}
	})

	if err != nil {
		http.Error(w, "Failed to subscribe to NATS", http.StatusInternalServerError)
		return
	}
	defer sub.Unsubscribe()

	// Keep the connection alive
	<-r.Context().Done()
}

func (app *App) handlePublish(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Content string `json:"content"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	message := Message{
		Content:   req.Content,
		Timestamp: time.Now(),
	}

	data, err := json.Marshal(message)
	if err != nil {
		http.Error(w, "Failed to marshal message", http.StatusInternalServerError)
		return
	}

	// Publish to NATS
	if err := app.natsConn.Publish("messages.demo", data); err != nil {
		http.Error(w, "Failed to publish message", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "published"})
}
