package nats

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	gonats "github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"

	"github.com/delaneyj/toolbelt/embeddednats"

	"github.com/joeblew999/infra/pkg/config"
	"github.com/joeblew999/infra/pkg/goreman"
	"github.com/joeblew999/infra/pkg/log"
)

// StartEmbeddedNATS starts an embedded NATS server and returns its client URL.
func StartEmbeddedNATS(ctx context.Context) (string, error) {
	// Initialize embedded NATS server
	log.Info("Starting embedded NATS server...")

	// Use .data folder for NATS data
	natsDataPath := filepath.Join(config.GetDataPath(), "nats")

	natsServer, err := embeddednats.New(ctx,
		embeddednats.WithDirectory(natsDataPath),
		embeddednats.WithShouldClearData(false), // Don't clear data
	)
	if err != nil {
		log.Error("Failed to create embedded NATS server", "error", err)
		return "", fmt.Errorf("Failed to create embedded NATS server: %w", err)
	}

	// Wait for the server to be ready with longer timeout
	log.Info("Waiting for NATS server to be ready...")
	maxWait := 15 * time.Second
	done := make(chan struct{})
	go func() {
		natsServer.WaitForServer()
		close(done)
	}()
	
	select {
	case <-done:
		log.Info("Embedded NATS server started and ready")
	case <-time.After(maxWait):
		// Log more detailed error
		log.Error("NATS server timeout", "data_path", natsDataPath)
		return "", fmt.Errorf("timeout waiting for NATS server after %v", maxWait)
	}
	
	// Get server info for debugging
	log.Info("NATS server started successfully")

	// Get client connection from the embedded server
	nc, err := natsServer.Client()
	if err != nil {
		return "", fmt.Errorf("Failed to get NATS client: %w", err)
	}

	// Close the client connection when the context is done
	go func() {
		<-ctx.Done()
		nc.Close()
		natsServer.Close()
	}()

	return nc.ConnectedUrl(), nil
}

// EnsureLoggingStream creates the logging stream if it doesn't exist
func EnsureLoggingStream(ctx context.Context, nc *gonats.Conn) error {
	js, err := jetstream.New(nc)
	if err != nil {
		return fmt.Errorf("failed to create jetstream context: %w", err)
	}

	// Create logging stream if it doesn't exist
	streamConfig := jetstream.StreamConfig{
		Name:     config.NATSLogStreamName,
		Subjects: []string{config.NATSLogStreamSubject},
		Storage:  jetstream.FileStorage,
		Retention: jetstream.LimitsPolicy,
		MaxAge:   24 * 30 * time.Hour, // 30 days
	}

	_, err = js.CreateOrUpdateStream(ctx, streamConfig)
	if err != nil {
		// Log warning but don't fail - JetStream might not be enabled
		log.Warn("Failed to create logging stream, continuing without NATS logging", "error", err)
		return nil
	}

	return nil
}

// StartSupervised starts NATS server under goreman supervision (idempotent)
// This uses the standalone nats-server binary instead of embedded NATS
func StartSupervised(port int) error {
	if port == 0 {
		port = 4222 // Default NATS port
	}
	
	// Ensure data directory exists
	natsDataPath := filepath.Join(config.GetDataPath(), "nats")
	if err := os.MkdirAll(natsDataPath, 0755); err != nil {
		return fmt.Errorf("failed to create NATS data directory: %w", err)
	}
	
	// Create basic NATS config
	configPath := filepath.Join(natsDataPath, "nats.conf")
	natsConfig := fmt.Sprintf(`
# NATS Server Configuration
port: %d
jetstream: enabled
store_dir: %s
`, port, natsDataPath)
	
	if err := os.WriteFile(configPath, []byte(natsConfig), 0644); err != nil {
		return fmt.Errorf("failed to create NATS config: %w", err)
	}
	
	// Register and start with goreman supervision
	return goreman.RegisterAndStart("nats", &goreman.ProcessConfig{
		Command:    config.Get(config.BinaryNats),
		Args:       []string{"--config", configPath},
		WorkingDir: ".",
		Env:        os.Environ(),
	})
}
