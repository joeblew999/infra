package demo

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/nats-io/nats.go"
	"github.com/starfederation/datastar-go/datastar"

	"github.com/joeblew999/infra/pkg/log"
)

// Store represents demo store structure
type Store struct {
	Delay time.Duration `json:"delay"`
}

// Message represents a demo message
type Message struct {
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

// DemoWebService provides web interface for demo functionality
type DemoWebService struct {
	natsConn *nats.Conn
}

// NewDemoWebService creates a new demo web service
func NewDemoWebService(natsConn *nats.Conn) *DemoWebService {
	return &DemoWebService{
		natsConn: natsConn,
	}
}

// RegisterRoutes mounts all demo routes on the provided router
func (s *DemoWebService) RegisterRoutes(r chi.Router) {
	// Demo routes
	r.Get("/hello-world", s.HandleHelloWorld)
	r.Get("/nats-stream", s.HandleNATSStream)
	r.Post("/publish", s.HandlePublish)
}

// HandleHelloWorld provides a DataStar typing effect demo
func (s *DemoWebService) HandleHelloWorld(w http.ResponseWriter, r *http.Request) {
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
}

// HandleNATSStream provides real-time NATS message streaming demo
func (s *DemoWebService) HandleNATSStream(w http.ResponseWriter, r *http.Request) {
	if s.natsConn == nil {
		http.Error(w, "NATS not connected", http.StatusServiceUnavailable)
		return
	}

	sse := datastar.NewSSE(w, r)

	// Subscribe to NATS messages
	sub, err := s.natsConn.Subscribe("messages.demo", func(msg *nats.Msg) {
		var message Message
		if err := json.Unmarshal(msg.Data, &message); err != nil {
			log.Error("Error unmarshaling message", "error", err)
			return
		}

		// Send the message to the browser via DataStar - append to existing messages
		messageHTML := fmt.Sprintf(`<div class="p-2 bg-blue-50 dark:bg-blue-900/20 rounded border-l-4 border-blue-400">`+
			`<div class="text-sm text-gray-600 dark:text-gray-400">%s</div>`+
			`<div class="text-gray-900 dark:text-white">%s</div>`+
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

// HandlePublish handles publishing messages to NATS for demo purposes
func (s *DemoWebService) HandlePublish(w http.ResponseWriter, r *http.Request) {
	if s.natsConn == nil {
		http.Error(w, "NATS not connected", http.StatusServiceUnavailable)
		return
	}

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
	if err := s.natsConn.Publish("messages.demo", data); err != nil {
		http.Error(w, "Failed to publish message", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "published"})
}

// SetupNATSStreams sets up JetStream for demo functionality
func (s *DemoWebService) SetupNATSStreams(ctx context.Context) error {
	if s.natsConn == nil {
		return nil // No NATS connection, skip setup
	}

	// Create JetStream context
	js, err := s.natsConn.JetStream()
	if err != nil {
		log.Error("Failed to create JetStream context", "error", err)
		return fmt.Errorf("Failed to create JetStream context: %w", err)
	}

	// Create a stream for demo messages
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