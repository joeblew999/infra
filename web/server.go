package web

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/nats-io/nats.go"
	"github.com/starfederation/datastar-go/datastar"

	"github.com/joeblew999/infra/pkg/docs"
	"github.com/joeblew999/infra/pkg/log"
	"github.com/joeblew999/infra/pkg/config"
	"github.com/joeblew999/infra/pkg/goreman/web"
)

//go:embed index.html
var helloWorldHTML []byte

const port = 1337

type App struct {
	natsConn     *nats.Conn
	router       *chi.Mux
	docsService  *docs.Service
	docsRenderer *docs.Renderer
}

type Store struct {
	Delay time.Duration `json:"delay"`
}

type Message struct {
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

func StartServer(natsAddr string, devDocs bool) error {
	ctx := context.Background()

	// Connect to NATS server (optional for debugging)
	nc, err := nats.Connect(natsAddr)
	if err != nil {
		log.Warn("Failed to connect to NATS, continuing without NATS features", "error", err)
		nc = nil // Allow web server to start without NATS
	} else {
		defer nc.Close()
	}

	app := &App{
		natsConn:     nc,
		router:       chi.NewRouter(),
		docsService:  docs.New(devDocs, config.DocsDir),
		docsRenderer: docs.NewRenderer(),
	}

	app.setupRoutes()
	if nc != nil {
		if err := app.setupNATSStreams(ctx); err != nil {
			log.Warn("Failed to setup NATS streams", "error", err)
		}
	}

	log.Info("Starting web server", "address", fmt.Sprintf("http://localhost:%d", port))
	if nc != nil {
		log.Info("Connected to NATS", "address", natsAddr)
	} else {
		log.Info("Running without NATS (debug mode)")
	}

	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), app.router); err != nil {
		return fmt.Errorf("Failed to start web server: %w", err)
	}
	return nil
}

func (app *App) setupRoutes() {
	app.router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write(helloWorldHTML)
	})

	// Original hello-world endpoint
	app.router.Get("/hello-world", func(w http.ResponseWriter, r *http.Request) {
		store := &Store{}
		if err := datastar.ReadSignals(r, store); err != nil {
			log.Error("Error reading signals", "error", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		sse := datastar.NewSSE(w, r)
		const message = "Hello, world!"

		for i := range len(message) {
			if err := sse.PatchElements(`<div id="message">` + message[:i+1] + `</div>`); err != nil {
				log.Error("Error patching elements", "error", err)
				return
			}
			time.Sleep(100 * time.Millisecond) // 100ms delay for typing effect
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

	// Navigation routes
	app.router.Get(config.MetricsHTTPPath, app.handleMetrics)
	app.router.Get(config.LogsHTTPPath, app.handleLogs)
	app.router.Get(config.StatusHTTPPath, app.handleStatus)

	// Docs handler
	app.router.Get(config.DocsHTTPPath+"*", app.handleDocs)

	// Process monitoring routes (goreman web GUI) - using sub-router pattern
	webHandler := web.NewWebHandler("pkg/goreman/web")
	app.router.Route("/processes", webHandler.SetupRoutes)
}

func (app *App) handleDocs(w http.ResponseWriter, r *http.Request) {
	filePath := r.URL.Path[len(config.DocsHTTPPath):]
	log.Info("Requested filePath", "path", filePath)

	// Read document content
	content, err := app.docsService.ReadFile(filePath)
	if err != nil {
		log.Error("Error reading document", "path", filePath, "error", err)
		http.Error(w, "Document not found", http.StatusNotFound)
		return
	}

	// Convert markdown to HTML
	htmlContent, err := app.docsRenderer.RenderToHTML(content)
	if err != nil {
		log.Error("Error rendering document", "path", filePath, "error", err)
		http.Error(w, "Failed to render document", http.StatusInternalServerError)
		return
	}

	// Wrap in HTML page structure with navigation
	nav := app.docsService.GetNavigation()
	fullHTML := app.docsRenderer.RenderToHTMLPage("Docs", htmlContent, nav)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(fullHTML))
}

func (app *App) setupNATSStreams(_ context.Context) error {
	// Create JetStream context
	js, err := app.natsConn.JetStream()
	if err != nil {
		log.Error("Failed to create JetStream context", "error", err)
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
		log.Warn("NATS Stream may already exist", "error", err)
	}

	log.Info("NATS JetStream setup complete", "stream", streamName)
	return nil
}

func (app *App) handleNATSStream(w http.ResponseWriter, r *http.Request) {
	sse := datastar.NewSSE(w, r)

	// Subscribe to NATS messages
	sub, err := app.natsConn.Subscribe("messages.demo", func(msg *nats.Msg) {
		var message Message
		if err := json.Unmarshal(msg.Data, &message); err != nil {
			log.Error("Error unmarshaling message", "error", err)
			return
		}

		// Send the message to the browser via DataStar - append to existing messages
		messageHTML := fmt.Sprintf(`<div class="p-2 bg-blue-50 dark:bg-blue-900/20 rounded border-l-4 border-blue-400">` +
			`<div class="text-sm text-gray-600 dark:text-gray-400">%s</div>` +
			`<div class="text-gray-900 dark:text-white">%s</div>` +
			`</div>`,
			message.Timestamp.Format("15:04:05"), message.Content)

		// Append to the messages div using DataStar append mechanism
		if err := sse.PatchElements(fmt.Sprintf(`<div id="messages" _="append '%s'"></div>`, messageHTML)); err != nil {
			log.Error("Error sending SSE", "error", err)
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

func (app *App) handleMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <title>Metrics - Infrastructure Management</title>
    <script src="https://unpkg.com/@tailwindcss/browser@4"></script>
</head>
<body class="bg-white dark:bg-gray-900 text-lg max-w-4xl mx-auto my-8">
    <nav class="mb-8 p-4 bg-gray-100 dark:bg-gray-800 rounded-lg shadow-md">
        <div class="flex justify-between items-center">
            <h1 class="text-xl font-bold text-gray-900 dark:text-white">🏗️ Infrastructure Management</h1>
            <div class="flex space-x-4">
                <a href="/" class="px-3 py-1 text-sm font-medium text-blue-600 dark:text-blue-400 hover:text-blue-800 dark:hover:text-blue-200 rounded hover:bg-blue-50 dark:hover:bg-blue-900/20">🏠 Home</a>
                <a href="/docs/" class="px-3 py-1 text-sm font-medium text-green-600 dark:text-green-400 hover:text-green-800 dark:hover:text-green-200 rounded hover:bg-green-50 dark:hover:bg-green-900/20">📚 Docs</a>
                <a href="/bento-playground" class="px-3 py-1 text-sm font-medium text-red-600 dark:text-red-400 hover:text-red-800 dark:hover:text-red-200 rounded hover:bg-red-50 dark:hover:bg-red-900/20">🎮 Bento Playground</a>
                <a href="/metrics" class="px-3 py-1 text-sm font-medium text-purple-600 dark:text-purple-400 hover:text-purple-800 dark:hover:text-purple-200 rounded hover:bg-purple-50 dark:hover:bg-purple-900/20 bg-purple-100 dark:bg-purple-900/30">📊 Metrics</a>
                <a href="/logs" class="px-3 py-1 text-sm font-medium text-orange-600 dark:text-orange-400 hover:text-orange-800 dark:hover:text-orange-200 rounded hover:bg-orange-50 dark:hover:bg-orange-900/20">📝 Logs</a>
                <a href="/processes" class="px-3 py-1 text-sm font-medium text-indigo-600 dark:text-indigo-400 hover:text-indigo-800 dark:hover:text-indigo-200 rounded hover:bg-indigo-50 dark:hover:bg-indigo-900/20">🔍 Processes</a>
                <a href="/status" class="px-3 py-1 text-sm font-medium text-gray-600 dark:text-gray-400 hover:text-gray-800 dark:hover:text-gray-200 rounded hover:bg-gray-50 dark:hover:bg-gray-900/20">⚡ Status</a>
            </div>
        </div>
    </nav>
    <div class="bg-white dark:bg-gray-800 text-gray-900 dark:text-white rounded-lg px-6 py-8 ring shadow-xl ring-gray-900/5">
        <h2 class="text-2xl font-semibold mb-4">📊 System Metrics</h2>
        <p class="text-gray-600 dark:text-gray-400">Metrics monitoring will be implemented here.</p>
    </div>
</body>
</html>`
	w.Write([]byte(html))
}

func (app *App) handleLogs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <title>Logs - Infrastructure Management</title>
    <script src="https://unpkg.com/@tailwindcss/browser@4"></script>
</head>
<body class="bg-white dark:bg-gray-900 text-lg max-w-4xl mx-auto my-8">
    <nav class="mb-8 p-4 bg-gray-100 dark:bg-gray-800 rounded-lg shadow-md">
        <div class="flex justify-between items-center">
            <h1 class="text-xl font-bold text-gray-900 dark:text-white">🏗️ Infrastructure Management</h1>
            <div class="flex space-x-4">
                <a href="/" class="px-3 py-1 text-sm font-medium text-blue-600 dark:text-blue-400 hover:text-blue-800 dark:hover:text-blue-200 rounded hover:bg-blue-50 dark:hover:bg-blue-900/20">🏠 Home</a>
                <a href="/docs/" class="px-3 py-1 text-sm font-medium text-green-600 dark:text-green-400 hover:text-green-800 dark:hover:text-green-200 rounded hover:bg-green-50 dark:hover:bg-green-900/20">📚 Docs</a>
                <a href="/bento-playground" class="px-3 py-1 text-sm font-medium text-red-600 dark:text-red-400 hover:text-red-800 dark:hover:text-red-200 rounded hover:bg-red-50 dark:hover:bg-red-900/20">🎮 Bento Playground</a>
                <a href="/metrics" class="px-3 py-1 text-sm font-medium text-purple-600 dark:text-purple-400 hover:text-purple-800 dark:hover:text-purple-200 rounded hover:bg-purple-50 dark:hover:bg-purple-900/20">📊 Metrics</a>
                <a href="/logs" class="px-3 py-1 text-sm font-medium text-orange-600 dark:text-orange-400 hover:text-orange-800 dark:hover:text-orange-200 rounded hover:bg-orange-50 dark:hover:bg-orange-900/20 bg-orange-100 dark:bg-orange-900/30">📝 Logs</a>
                <a href="/processes" class="px-3 py-1 text-sm font-medium text-indigo-600 dark:text-indigo-400 hover:text-indigo-800 dark:hover:text-indigo-200 rounded hover:bg-indigo-50 dark:hover:bg-indigo-900/20">🔍 Processes</a>
                <a href="/status" class="px-3 py-1 text-sm font-medium text-gray-600 dark:text-gray-400 hover:text-gray-800 dark:hover:text-gray-200 rounded hover:bg-gray-50 dark:hover:bg-gray-900/20">⚡ Status</a>
            </div>
        </div>
    </nav>
    <div class="bg-white dark:bg-gray-800 text-gray-900 dark:text-white rounded-lg px-6 py-8 ring shadow-xl ring-gray-900/5">
        <h2 class="text-2xl font-semibold mb-4">📝 System Logs</h2>
        <p class="text-gray-600 dark:text-gray-400">Log monitoring will be implemented here.</p>
    </div>
</body>
</html>`
	w.Write([]byte(html))
}

func (app *App) handleStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <title>Status - Infrastructure Management</title>
    <script src="https://unpkg.com/@tailwindcss/browser@4"></script>
</head>
<body class="bg-white dark:bg-gray-900 text-lg max-w-4xl mx-auto my-8">
    <nav class="mb-8 p-4 bg-gray-100 dark:bg-gray-800 rounded-lg shadow-md">
        <div class="flex justify-between items-center">
            <h1 class="text-xl font-bold text-gray-900 dark:text-white">🏗️ Infrastructure Management</h1>
            <div class="flex space-x-4">
                <a href="/" class="px-3 py-1 text-sm font-medium text-blue-600 dark:text-blue-400 hover:text-blue-800 dark:hover:text-blue-200 rounded hover:bg-blue-50 dark:hover:bg-blue-900/20">🏠 Home</a>
                <a href="/docs/" class="px-3 py-1 text-sm font-medium text-green-600 dark:text-green-400 hover:text-green-800 dark:hover:text-green-200 rounded hover:bg-green-50 dark:hover:bg-green-900/20">📚 Docs</a>
                <a href="/bento-playground" class="px-3 py-1 text-sm font-medium text-red-600 dark:text-red-400 hover:text-red-800 dark:hover:text-red-200 rounded hover:bg-red-50 dark:hover:bg-red-900/20">🎮 Bento Playground</a>
                <a href="/metrics" class="px-3 py-1 text-sm font-medium text-purple-600 dark:text-purple-400 hover:text-purple-800 dark:hover:text-purple-200 rounded hover:bg-purple-50 dark:hover:bg-purple-900/20">📊 Metrics</a>
                <a href="/logs" class="px-3 py-1 text-sm font-medium text-orange-600 dark:text-orange-400 hover:text-orange-800 dark:hover:text-orange-200 rounded hover:bg-orange-50 dark:hover:bg-orange-900/20">📝 Logs</a>
                <a href="/processes" class="px-3 py-1 text-sm font-medium text-indigo-600 dark:text-indigo-400 hover:text-indigo-800 dark:hover:text-indigo-200 rounded hover:bg-indigo-50 dark:hover:bg-indigo-900/20">🔍 Processes</a>
                <a href="/status" class="px-3 py-1 text-sm font-medium text-gray-600 dark:text-gray-400 hover:text-gray-800 dark:hover:text-gray-200 rounded hover:bg-gray-50 dark:hover:bg-gray-900/20 bg-gray-100 dark:bg-gray-900/30">⚡ Status</a>
            </div>
        </div>
    </nav>
    <div class="bg-white dark:bg-gray-800 text-gray-900 dark:text-white rounded-lg px-6 py-8 ring shadow-xl ring-gray-900/5">
        <h2 class="text-2xl font-semibold mb-4">⚡ System Status</h2>
        <div class="space-y-4">
            <div class="flex items-center justify-between p-4 bg-green-50 dark:bg-green-900/20 rounded-lg">
                <span class="text-green-800 dark:text-green-200">🟢 Web Server</span>
                <span class="text-green-600 dark:text-green-400 font-semibold">Running</span>
            </div>
            <div class="flex items-center justify-between p-4 bg-green-50 dark:bg-green-900/20 rounded-lg">
                <span class="text-green-800 dark:text-green-200">🟢 NATS Server</span>
                <span class="text-green-600 dark:text-green-400 font-semibold">Running</span>
            </div>
        </div>
    </div>
</body>
</html>`
	w.Write([]byte(html))
}
