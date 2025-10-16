package events

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog/log"
)

// Consumer subscribes to process events from NATS.
type Consumer struct {
	natsURL string
	nc      *nats.Conn
	js      nats.JetStreamContext
	subs    []*nats.Subscription
	ctx     context.Context
	cancel  context.CancelFunc
}

// NewConsumer creates a new event consumer.
func NewConsumer(natsURL string) (*Consumer, error) {
	if natsURL == "" {
		natsURL = "nats://127.0.0.1:4222"
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Consumer{
		natsURL: natsURL,
		ctx:     ctx,
		cancel:  cancel,
	}, nil
}

// Connect establishes connection to NATS.
func (c *Consumer) Connect() error {
	nc, err := nats.Connect(c.natsURL,
		nats.Name("core-event-consumer"),
		nats.MaxReconnects(-1),
		nats.ReconnectWait(1*time.Second),
	)
	if err != nil {
		return fmt.Errorf("connect to nats: %w", err)
	}
	c.nc = nc

	js, err := nc.JetStream()
	if err != nil {
		nc.Close()
		return fmt.Errorf("setup jetstream: %w", err)
	}
	c.js = js

	return nil
}

// Close stops all subscriptions and closes the connection.
func (c *Consumer) Close() error {
	c.cancel()

	// Unsubscribe all
	for _, sub := range c.subs {
		sub.Unsubscribe()
	}

	if c.nc != nil {
		c.nc.Close()
	}

	return nil
}

// Subscribe subscribes to events matching the pattern and calls handler for each event.
func (c *Consumer) Subscribe(pattern string, handler func(Event) error) error {
	sub, err := c.js.Subscribe(pattern, func(msg *nats.Msg) {
		var evt Event
		if err := json.Unmarshal(msg.Data, &evt); err != nil {
			log.Error().Err(err).Msg("Failed to unmarshal event")
			msg.Nak() // Negative acknowledge
			return
		}

		if err := handler(evt); err != nil {
			log.Error().
				Err(err).
				Str("process", evt.Process).
				Str("type", string(evt.Type)).
				Msg("Event handler failed")
			msg.Nak()
			return
		}

		msg.Ack() // Acknowledge successful processing
	}, nats.Durable("core-event-consumer"), nats.DeliverNew())

	if err != nil {
		return fmt.Errorf("subscribe to %s: %w", pattern, err)
	}

	c.subs = append(c.subs, sub)
	log.Info().Str("pattern", pattern).Msg("Subscribed to events")
	return nil
}

// SubscribeAll subscribes to all process events.
func (c *Consumer) SubscribeAll(handler func(Event) error) error {
	return c.Subscribe(SubjectPattern(AllEvents()), handler)
}

// SubscribeProcess subscribes to all events for a specific process.
func (c *Consumer) SubscribeProcess(processName string, handler func(Event) error) error {
	return c.Subscribe(SubjectPattern(ForProcess(processName)), handler)
}

// SubscribeEventType subscribes to a specific event type across all processes.
func (c *Consumer) SubscribeEventType(eventType EventType, handler func(Event) error) error {
	return c.Subscribe(SubjectPattern(ForEventType(eventType)), handler)
}

// SubscribeProcessEvent subscribes to a specific event type for a specific process.
func (c *Consumer) SubscribeProcessEvent(processName string, eventType EventType, handler func(Event) error) error {
	return c.Subscribe(SubjectPattern(ForProcessAndType(processName, eventType)), handler)
}

// Wait blocks until the consumer is closed or context is cancelled.
func (c *Consumer) Wait() {
	<-c.ctx.Done()
}
