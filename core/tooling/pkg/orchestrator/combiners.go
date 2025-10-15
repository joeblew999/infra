package orchestrator

import (
	"context"

	"github.com/joeblew999/infra/core/tooling/pkg/auth"
	types "github.com/joeblew999/infra/core/tooling/pkg/types"
)

// combineEmitters merges two progress emitters.
func combineEmitters(primary, secondary ProgressEmitter) ProgressEmitter {
	if primary == nil {
		return secondary
	}
	if secondary == nil {
		return primary
	}
	return ProgressEmitterFunc(func(evt ProgressEvent) {
		primary.Emit(evt)
		secondary.Emit(evt)
	})
}

// combinePrompters merges two prompters.
func combinePrompters(primary, secondary auth.Prompter) auth.Prompter {
	if primary == nil {
		return secondary
	}
	if secondary == nil {
		return primary
	}
	return &combinedPrompter{primary: primary, secondary: secondary}
}

type combinedPrompter struct {
	primary   auth.Prompter
	secondary auth.Prompter
}

func (p *combinedPrompter) Notify(ctx context.Context, req types.PromptMessage) error {
	_ = p.primary.Notify(ctx, req)
	return p.secondary.Notify(ctx, req)
}

func (p *combinedPrompter) PromptSecret(ctx context.Context, req types.PromptMessage) (string, error) {
	return p.primary.PromptSecret(ctx, req)
}
