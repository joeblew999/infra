package live

import (
	"context"
	"fmt"

	runtimeprocess "github.com/joeblew999/infra/core/pkg/runtime/process"
)

// StartProcess requests that Process Compose start the named process.
func (s *Store) StartProcess(ctx context.Context, port int, name string) error {
	if err := runtimeprocess.StartComposeProcess(ctx, port, name); err != nil {
		return err
	}
	s.AppendEvent(fmt.Sprintf("process %s started", name))
	return nil
}

// StopProcess requests that Process Compose stop the named process.
func (s *Store) StopProcess(ctx context.Context, port int, name string) error {
	if err := runtimeprocess.StopComposeProcess(ctx, port, name); err != nil {
		return err
	}
	s.AppendEvent(fmt.Sprintf("process %s stopping", name))
	return nil
}

// RestartProcess requests that Process Compose restart the named process.
func (s *Store) RestartProcess(ctx context.Context, port int, name string) error {
	if err := runtimeprocess.RestartComposeProcess(ctx, port, name); err != nil {
		return err
	}
	s.AppendEvent(fmt.Sprintf("process %s restart triggered", name))
	return nil
}

// ScaleProcess requests that Process Compose scale the named process to the desired count.
func (s *Store) ScaleProcess(ctx context.Context, port int, name string, count int) error {
	if err := runtimeprocess.ScaleComposeProcess(ctx, port, name, count); err != nil {
		return err
	}
	s.AppendEvent(fmt.Sprintf("process %s scaled to %d", name, count))
	return nil
}
