package nats

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/nats-io/nats-server/v2/server"
	gonats "github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"

	"github.com/joeblew999/infra/pkg/config"
	"github.com/joeblew999/infra/pkg/dep"
	"github.com/joeblew999/infra/pkg/log"
	"github.com/joeblew999/infra/pkg/service"
)

// StartEmbeddedNATS starts an embedded NATS server acting as a leaf node and returns its client URL, connection, and a cleanup function.
func StartEmbeddedNATS(ctx context.Context, remotes []string, credentialsPath string) (string, *gonats.Conn, func(), error) {
	// Initialize embedded NATS server
	log.Info("Starting embedded NATS server...")

	// Use .data folder for NATS data
	natsDataPath := filepath.Join(config.GetDataPath(), "nats")
	if err := os.MkdirAll(natsDataPath, 0o755); err != nil {
		return "", nil, nil, fmt.Errorf("failed to create NATS data directory: %w", err)
	}

	port := 0
	httpPort := 0

	if !config.IsTestEnvironment() {
		portStr := config.GetNATSPort()
		var err error
		port, err = strconv.Atoi(portStr)
		if err != nil {
			return "", nil, nil, fmt.Errorf("invalid NATS port %s: %w", portStr, err)
		}

		if port != 0 {
			ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
			if err != nil {
				log.Warn("Requested NATS port in use, falling back to random", "port", port)
				port = 0
			} else {
				_ = ln.Close()
			}
		}
	}

	serverOpts := &server.Options{
		JetStream: true,
		StoreDir:  natsDataPath,
		Port:      port,
		HTTPPort:  httpPort,
	}

	if len(remotes) > 0 {
		remoteOpts := make([]*server.RemoteLeafOpts, 0, len(remotes))
		for _, remote := range remotes {
			u, err := url.Parse(remote)
			if err != nil {
				return "", nil, nil, fmt.Errorf("invalid leaf remote %s: %w", remote, err)
			}
			opts := &server.RemoteLeafOpts{URLs: []*url.URL{u}}
			if credentialsPath != "" {
				opts.Credentials = credentialsPath
			}
			remoteOpts = append(remoteOpts, opts)
		}
		serverOpts.LeafNode = server.LeafNodeOpts{Remotes: remoteOpts}
		log.Info("Configured leaf remotes", "count", len(remoteOpts))
	}

	natsServer, err := newEmbeddedServer(ctx,
		withEmbeddedDirectory(natsDataPath),
		withEmbeddedServerOptions(serverOpts),
	)
	if err != nil {
		log.Error("Failed to create embedded NATS server", "error", err)
		return "", nil, nil, fmt.Errorf("Failed to create embedded NATS server: %w", err)
	}

	// Wait for the server to accept client connections
	log.Info("Waiting for NATS server to be ready...")
	maxWait := 45 * time.Second
	deadline := time.Now().Add(maxWait)
	var nc *gonats.Conn
	for {
		nc, err = gonats.Connect(natsServer.server.ClientURL())
		if err == nil {
			log.Info("Embedded NATS server started and ready")
			break
		}

		if time.Now().After(deadline) {
			log.Error("NATS server timeout", "data_path", natsDataPath, "error", err)
			natsServer.Close()
			return "", nil, nil, fmt.Errorf("timeout waiting for NATS server after %v: %w", maxWait, err)
		}

		log.Warn("Embedded NATS server not ready, retrying", "error", err)
		time.Sleep(750 * time.Millisecond)
	}

	cleanup := func() {
		nc.Close()
		natsServer.Close()
	}

	return nc.ConnectedUrl(), nc, cleanup, nil
}

// Embedded NATS server helpers (inlined from upstream toolbelt library).

type embeddedOptions struct {
	DataDirectory     string
	ShouldClearData   bool
	NATSServerOptions *server.Options
}

type embeddedOption func(*embeddedOptions)

func withEmbeddedDirectory(dir string) embeddedOption {
	return func(o *embeddedOptions) {
		o.DataDirectory = dir
	}
}

func withEmbeddedServerOptions(opts *server.Options) embeddedOption {
	return func(o *embeddedOptions) {
		o.NATSServerOptions = opts
	}
}

type embeddedServer struct {
	server *server.Server
}

func newEmbeddedServer(ctx context.Context, opts ...embeddedOption) (*embeddedServer, error) {
	options := &embeddedOptions{
		DataDirectory: "./data/nats",
	}
	for _, opt := range opts {
		opt(options)
	}

	if options.ShouldClearData {
		if err := os.RemoveAll(options.DataDirectory); err != nil {
			return nil, err
		}
	}

	if options.NATSServerOptions == nil {
		options.NATSServerOptions = &server.Options{
			JetStream: true,
			StoreDir:  options.DataDirectory,
		}
	}

	ns, err := server.NewServer(options.NATSServerOptions)
	if err != nil {
		return nil, err
	}

	go func() {
		<-ctx.Done()
		ns.Shutdown()
	}()

	ns.ConfigureLogger()
	if opts := options.NATSServerOptions; opts != nil {
		ns.SetLogger(newStructuredNATSLogger(), opts.Debug, opts.Trace)
	} else {
		ns.SetLogger(newStructuredNATSLogger(), false, false)
	}
	ns.Start()

	if !ns.ReadyForConnections(10 * time.Second) {
		ns.Shutdown()
		return nil, fmt.Errorf("nats server not ready")
	}

	return &embeddedServer{server: ns}, nil
}

func (s *embeddedServer) Close() error {
	if s.server != nil && s.server.Running() {
		s.server.Shutdown()
	}
	return nil
}

// EnsureLoggingStream creates the logging stream if it doesn't exist
func EnsureLoggingStream(ctx context.Context, nc *gonats.Conn) error {
	js, err := jetstream.New(nc)
	if err != nil {
		return fmt.Errorf("failed to create jetstream context: %w", err)
	}

	// Create logging stream if it doesn't exist
	streamConfig := jetstream.StreamConfig{
		Name:      config.NATSLogStreamName,
		Subjects:  []string{config.NATSLogStreamSubject},
		Storage:   jetstream.FileStorage,
		Retention: jetstream.LimitsPolicy,
		MaxAge:    24 * 30 * time.Hour, // 30 days
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

	if err := dep.InstallBinary(config.BinaryNatsServer, false); err != nil {
		return fmt.Errorf("failed to ensure nats binary: %w", err)
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
	processCfg := service.NewConfig(config.Get(config.BinaryNatsServer), []string{"--config", configPath})
	return service.Start("nats", processCfg)
}
