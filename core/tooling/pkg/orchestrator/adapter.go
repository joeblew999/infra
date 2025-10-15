package orchestrator

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/joeblew999/infra/core/tooling/pkg/auth"
	types "github.com/joeblew999/infra/core/tooling/pkg/types"
)

// StreamAdapter bridges progress events and prompt interactions to channels.
type StreamAdapter struct {
	Progress chan types.ProgressMessage
	Prompts  chan types.PromptMessage

	responses sync.Map
	counter   uint64
	closeOnce sync.Once
}

// NewStreamAdapter creates a stream adapter with default buffering.
func NewStreamAdapter() *StreamAdapter {
	return &StreamAdapter{
		Progress: make(chan types.ProgressMessage, 64),
		Prompts:  make(chan types.PromptMessage, 16),
	}
}

// Close closes the adapter channels.
func (a *StreamAdapter) Close() {
	if a == nil {
		return
	}
	a.closeOnce.Do(func() {
		close(a.Progress)
		close(a.Prompts)
	})
}

// Emitter returns a ProgressEmitter that forwards events to the adapter.
func (a *StreamAdapter) Emitter() ProgressEmitter {
	return ProgressEmitterFunc(func(evt ProgressEvent) {
		if a == nil {
			return
		}
		msg := types.ProgressMessage{
			Phase:   string(evt.Phase),
			Message: evt.Message,
			Details: evt.Details,
			Time:    evt.Time,
		}
		select {
		case a.Progress <- msg:
		default:
			// drop if consumer is slow; UI should keep up
		}
	})
}

// Prompter returns an auth.Prompter that surfaces prompts via the adapter.
func (a *StreamAdapter) Prompter() auth.Prompter {
	return &adapterPrompter{adapter: a}
}

// Respond resolves a pending prompt by ID.
func (a *StreamAdapter) Respond(id string, resp types.PromptResponse) {
	if v, ok := a.responses.Load(id); ok {
		if ch, ok := v.(chan types.PromptResponse); ok {
			ch <- resp
			close(ch)
		}
		a.responses.Delete(id)
	}
}

type adapterPrompter struct {
	adapter *StreamAdapter
}

func (p *adapterPrompter) Notify(ctx context.Context, req types.PromptMessage) error {
	if p == nil || p.adapter == nil {
		return nil
	}
	req.RequiresResponse = false
	select {
	case p.adapter.Prompts <- req:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (p *adapterPrompter) PromptSecret(ctx context.Context, req types.PromptMessage) (string, error) {
	if p == nil || p.adapter == nil {
		return "", context.Canceled
	}
	id := req.ID
	if id == "" {
		id = p.adapter.nextID()
	}
	req.ID = id
	req.RequiresResponse = true

	ch := make(chan types.PromptResponse, 1)
	p.adapter.responses.Store(id, ch)

	select {
	case p.adapter.Prompts <- req:
	case <-ctx.Done():
		p.adapter.responses.Delete(id)
		return "", ctx.Err()
	}

	select {
	case resp := <-ch:
		if resp.Error != "" {
			return "", errors.New(resp.Error)
		}
		return resp.Secret, nil
	case <-ctx.Done():
		return "", ctx.Err()
	}
}

func (a *StreamAdapter) nextID() string {
	val := atomic.AddUint64(&a.counter, 1)
	return fmt.Sprintf("prompt-%d", val)
}
