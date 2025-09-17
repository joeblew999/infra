package goreman

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/nats-io/nats.go"

	"github.com/joeblew999/infra/pkg/log"
)

const (
	// CommandSubject is the NATS subject that CLI clients publish control commands to.
	CommandSubject = "goreman.commands"
	commandQueue   = "goreman-control"
)

// ControlCommand represents a remote instruction for the goreman supervisor.
type ControlCommand struct {
	Action string `json:"action"`
	Name   string `json:"name,omitempty"`
}

// ControlResponse represents the outcome of executing a remote command.
type ControlResponse struct {
	Success  bool              `json:"success"`
	Message  string            `json:"message,omitempty"`
	Statuses map[string]string `json:"statuses,omitempty"`
}

// StartCommandListener subscribes to the control subject and executes incoming commands.
func StartCommandListener(ctx context.Context, nc *nats.Conn) error {
	if nc == nil {
		return errors.New("nil nats connection")
	}

	sub, err := nc.QueueSubscribe(CommandSubject, commandQueue, func(msg *nats.Msg) {
		handleCommandMessage(msg)
	})
	if err != nil {
		return fmt.Errorf("subscribe to goreman commands: %w", err)
	}

	go func() {
		<-ctx.Done()
		_ = sub.Drain()
	}()

	if err := nc.Flush(); err != nil {
		return fmt.Errorf("flush goreman command subscription: %w", err)
	}

	return nil
}

func handleCommandMessage(msg *nats.Msg) {
	var cmd ControlCommand
	if err := json.Unmarshal(msg.Data, &cmd); err != nil {
		respondCommand(msg, ControlResponse{Success: false, Message: fmt.Sprintf("invalid command payload: %v", err)})
		log.Warn("Received invalid goreman command", "error", err)
		_ = msg.Ack()
		return
	}

	action := strings.ToLower(strings.TrimSpace(cmd.Action))
	var resp ControlResponse

	switch action {
	case "start":
		if cmd.Name == "" {
			resp = ControlResponse{Success: false, Message: "missing process name"}
		} else if err := Start(cmd.Name); err != nil {
			resp = ControlResponse{Success: false, Message: err.Error()}
			log.Warn("Failed to start process via control channel", "name", cmd.Name, "error", err)
		} else {
			resp = ControlResponse{Success: true, Message: fmt.Sprintf("started %s", cmd.Name)}
			log.Info("Started process via control channel", "name", cmd.Name)
		}
	case "stop":
		if cmd.Name == "" {
			resp = ControlResponse{Success: false, Message: "missing process name"}
		} else if err := Stop(cmd.Name); err != nil {
			resp = ControlResponse{Success: false, Message: err.Error()}
			log.Warn("Failed to stop process via control channel", "name", cmd.Name, "error", err)
		} else {
			resp = ControlResponse{Success: true, Message: fmt.Sprintf("stopped %s", cmd.Name)}
			log.Info("Stopped process via control channel", "name", cmd.Name)
		}
	case "restart":
		if cmd.Name == "" {
			resp = ControlResponse{Success: false, Message: "missing process name"}
		} else if err := Restart(cmd.Name); err != nil {
			resp = ControlResponse{Success: false, Message: err.Error()}
			log.Warn("Failed to restart process via control channel", "name", cmd.Name, "error", err)
		} else {
			resp = ControlResponse{Success: true, Message: fmt.Sprintf("restarted %s", cmd.Name)}
			log.Info("Restarted process via control channel", "name", cmd.Name)
		}
	case "status":
		resp = ControlResponse{Success: true, Statuses: GetAllStatus()}
		log.Debug("Reported goreman status via control channel", "count", len(resp.Statuses))
	default:
		resp = ControlResponse{Success: false, Message: fmt.Sprintf("unknown action %q", cmd.Action)}
		log.Warn("Received unknown goreman control action", "action", cmd.Action)
	}

	respondCommand(msg, resp)
}

func respondCommand(msg *nats.Msg, resp ControlResponse) {
	if msg.Reply == "" {
		return
	}

	data, err := json.Marshal(resp)
	if err != nil {
		return
	}
	_ = msg.Respond(data)
}

// StartCommandSupervisor establishes a resilient NATS connection for goreman control and
// launches the command listener. The returned cleanup function terminates the connection.
func StartCommandSupervisor(ctx context.Context, addr string) (func(), error) {
	if addr == "" {
		return func() {}, errors.New("empty NATS address")
	}

	nc, err := nats.Connect(addr,
		nats.RetryOnFailedConnect(true),
		nats.MaxReconnects(-1),
		nats.ReconnectWait(2*time.Second),
		nats.Timeout(2*time.Second),
	)
	if err != nil {
		return func() {}, fmt.Errorf("connect to NATS: %w", err)
	}

	if err := StartCommandListener(ctx, nc); err != nil {
		nc.Close()
		return func() {}, err
	}

	cleanup := func() {
		nc.Close()
	}

	go func() {
		<-ctx.Done()
		cleanup()
	}()

	return cleanup, nil
}

// ExecuteCommand publishes a control command using the provided NATS connection and waits for a response.
func ExecuteCommand(ctx context.Context, nc *nats.Conn, cmd ControlCommand) (*ControlResponse, error) {
	if nc == nil {
		return nil, errors.New("nil nats connection")
	}

	payload, err := json.Marshal(cmd)
	if err != nil {
		return nil, fmt.Errorf("marshal command: %w", err)
	}

	reqCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	respMsg, err := nc.RequestWithContext(reqCtx, CommandSubject, payload)
	if err != nil {
		return nil, fmt.Errorf("await command response: %w", err)
	}

	var resp ControlResponse
	if err := json.Unmarshal(respMsg.Data, &resp); err != nil {
		return nil, fmt.Errorf("decode command response: %w", err)
	}

	return &resp, nil
}
