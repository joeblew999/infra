package nats

import (
	"context"
	"fmt"
	"log"
	"path/filepath"

	"github.com/delaneyj/toolbelt/embeddednats"
	"github.com/nats-io/nats-server/v2/server"

	"github.com/joeblew999/infra/pkg/store"
)

// StartEmbeddedNATS starts an embedded NATS server and returns its client URL.
func StartEmbeddedNATS(ctx context.Context) (string, error) {
	// Configure NATS server options for logging
	natsOpts := &server.Options{
		Debug: true, // Enable debug logging
		Trace: true, // Enable trace logging
	}

	// Initialize embedded NATS server
	log.Println("Starting embedded NATS server...")

	// Use pkg/store for the data path
	natsDataPath := filepath.Join(store.GetDataPath(), "nats")

	natsServer, err := embeddednats.New(ctx,
		embeddednats.WithDirectory(natsDataPath),
		embeddednats.WithNATSServerOptions(natsOpts),
	)
	if err != nil {
		log.Printf("Failed to create embedded NATS server: %v", err)
		return "", fmt.Errorf("Failed to create embedded NATS server: %w", err)
	}

	// Wait for the server to be ready
	natsServer.WaitForServer()
	log.Printf("Embedded NATS server started")

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
