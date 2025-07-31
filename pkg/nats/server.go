package nats

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	gonats "github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"

	"github.com/delaneyj/toolbelt/embeddednats"
	"github.com/nats-io/nats-server/v2/server"

	"github.com/joeblew999/infra/pkg/log"
	"github.com/joeblew999/infra/pkg/config"
)

// StartEmbeddedNATS starts an embedded NATS server and returns its client URL.
func StartEmbeddedNATS(ctx context.Context) (string, error) {
	// Configure NATS server options for logging
	natsOpts := &server.Options{
		Debug: true, // Enable debug logging
		Trace: true, // Enable trace logging
		JetStream: true, // Enable JetStream for logging
		Port: 4222, // Explicitly set NATS port
		Host: "127.0.0.1", // Bind to localhost only
	}

	// Initialize embedded NATS server
	log.Info("Starting embedded NATS server...")

	// Use pkg/store for the data path
	natsDataPath := filepath.Join(config.GetDataPath(), "nats")

	natsServer, err := embeddednats.New(ctx,
		embeddednats.WithDirectory(natsDataPath),
		embeddednats.WithNATSServerOptions(natsOpts),
	)
	if err != nil {
		log.Error("Failed to create embedded NATS server", "error", err)
		return "", fmt.Errorf("Failed to create embedded NATS server: %w", err)
	}

	// Wait for the server to be ready
	log.Info("Waiting for NATS server to be ready...")
	natsServer.WaitForServer()
	log.Info("Embedded NATS server started and ready")
	
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
