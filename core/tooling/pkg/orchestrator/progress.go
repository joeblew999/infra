package orchestrator

import (
	"encoding/json"
	"fmt"
	"io"
	"time"
)

// ProgressPhase identifies a high-level deployment stage.
type ProgressPhase string

const (
	PhaseStarted            ProgressPhase = "started"
	PhaseFlyAuth            ProgressPhase = "fly_auth"
	PhaseFlyAuthCompleted   ProgressPhase = "fly_auth_completed"
	PhaseCloudflareAuth     ProgressPhase = "cloudflare_auth"
	PhaseCloudflareComplete ProgressPhase = "cloudflare_auth_completed"
	PhaseCloudflareDNS      ProgressPhase = "cloudflare_dns"
	PhaseDeploying          ProgressPhase = "deploying"
	PhaseSucceeded          ProgressPhase = "succeeded"
	PhaseFailed             ProgressPhase = "failed"
)

// ProgressEvent describes a deploy-stage update.
type ProgressEvent struct {
	Phase   ProgressPhase     `json:"phase"`
	Message string            `json:"message"`
	Details map[string]string `json:"details,omitempty"`
	Time    time.Time         `json:"time"`
}

// ProgressEmitter consumes progress events.
type ProgressEmitter interface {
	Emit(ProgressEvent)
}

// ProgressEmitterFunc is an adapter for functions.
type ProgressEmitterFunc func(ProgressEvent)

func (f ProgressEmitterFunc) Emit(evt ProgressEvent) { f(evt) }

// TextEmitter renders events as human-readable text.
type TextEmitter struct {
	out io.Writer
}

func NewTextEmitter(w io.Writer) ProgressEmitter {
	return &TextEmitter{out: w}
}

func (t *TextEmitter) Emit(evt ProgressEvent) {
	if t == nil || t.out == nil {
		return
	}

	switch evt.Phase {
	case PhaseStarted:
		fmt.Fprintln(t.out, evt.Message)
		fmt.Fprintln(t.out)
	case PhaseFlyAuthCompleted, PhaseCloudflareComplete:
		if evt.Message != "" {
			fmt.Fprintln(t.out, evt.Message)
		}
		if evt.Details != nil {
			for key, value := range evt.Details {
				if value != "" {
					fmt.Fprintf(t.out, "  %s: %s\n", key, value)
				}
			}
		}
		fmt.Fprintln(t.out)
	case PhaseDeploying:
		if evt.Message != "" {
			fmt.Fprintln(t.out, evt.Message)
		}
	case PhaseSucceeded:
		fmt.Fprintln(t.out, evt.Message)
		fmt.Fprintln(t.out)
		if evt.Details != nil {
			if v := evt.Details["image"]; v != "" {
				fmt.Fprintf(t.out, "  Image:      %s\n", v)
			}
			if v := evt.Details["deployment"]; v != "" {
				fmt.Fprintf(t.out, "  Deployment: %s\n", v)
			}
			if v := evt.Details["release_id"]; v != "" {
				fmt.Fprintf(t.out, "  Release ID: %s\n", v)
			}
		}
	case PhaseFailed:
		fmt.Fprintln(t.out, evt.Message)
	}
}

// JSONEmitter renders events as JSON lines.
type JSONEmitter struct {
	out io.Writer
	enc *json.Encoder
}

func NewJSONEmitter(w io.Writer) ProgressEmitter {
	return &JSONEmitter{out: w, enc: json.NewEncoder(w)}
}

func (j *JSONEmitter) Emit(evt ProgressEvent) {
	if j == nil || j.enc == nil {
		return
	}
	_ = j.enc.Encode(evt)
}
