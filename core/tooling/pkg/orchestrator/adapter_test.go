package orchestrator

import (
	"context"
	"testing"
	"time"

	types "github.com/joeblew999/infra/core/tooling/pkg/types"
)

func TestStreamAdapterEmitterAndPrompter(t *testing.T) {
	adapter := NewStreamAdapter()
	defer adapter.Close()

	emitter := adapter.Emitter()
	prompter := adapter.Prompter()

	// progress event
	evt := ProgressEvent{Phase: PhaseStarted, Message: "start", Time: time.Now().UTC()}
	emitter.Emit(evt)

	select {
	case msg := <-adapter.Progress:
		if msg.Message != "start" || msg.Phase != string(PhaseStarted) {
			t.Fatalf("unexpected progress message: %#v", msg)
		}
	case <-time.After(time.Second):
		t.Fatal("expected progress message")
	}

	// prompt interaction
	go func() {
		select {
		case prompt := <-adapter.Prompts:
			adapter.Respond(prompt.ID, types.PromptResponse{ID: prompt.ID, Secret: "secret"})
		case <-time.After(time.Second):
			t.Errorf("prompt not received")
		}
	}()

	secret, err := prompter.PromptSecret(context.Background(), types.PromptMessage{Provider: "test", Kind: types.PromptKindToken, Message: "token:"})
	if err != nil {
		t.Fatalf("PromptSecret error: %v", err)
	}
	if secret != "secret" {
		t.Fatalf("expected secret 'secret', got %q", secret)
	}
}
