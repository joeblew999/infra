package auth

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/skratchdot/open-golang/open"

	types "github.com/joeblew999/infra/core/tooling/pkg/types"
)

// Prompter handles interactive authentication prompts.
type Prompter interface {
	Notify(context.Context, types.PromptMessage) error
	PromptSecret(context.Context, types.PromptMessage) (string, error)
}

// IOPrompter implements Prompter using stdin/stdout.
type IOPrompter struct {
	in        io.Reader
	out       io.Writer
	noBrowser bool
}

// NewIOPrompter returns a CLI-based prompter.
func NewIOPrompter(in io.Reader, out io.Writer, noBrowser bool) Prompter {
	return &IOPrompter{in: in, out: out, noBrowser: noBrowser}
}

// Notify prints informational messages and optionally opens URLs.
func (p *IOPrompter) Notify(_ context.Context, req types.PromptMessage) error {
	if p == nil {
		return nil
	}
	if req.Message != "" {
		fmt.Fprintln(p.out, req.Message)
	}
	if req.URL != "" && !p.noBrowser {
		_ = open.Run(req.URL)
	} else if req.URL != "" {
		fmt.Fprintf(p.out, "Visit: %s\n", req.URL)
	}
	if len(req.Scopes) > 0 {
		fmt.Fprintf(p.out, "Scopes: %v\n", req.Scopes)
	}
	return nil
}

// PromptSecret asks the user to paste a secret/token.
func (p *IOPrompter) PromptSecret(_ context.Context, req types.PromptMessage) (string, error) {
	if p == nil {
		return "", fmt.Errorf("prompter unavailable")
	}
	if req.Message != "" {
		fmt.Fprint(p.out, req.Message)
	}
	reader := bufio.NewReader(p.in)
	value, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return "", err
	}
	return strings.TrimSpace(value), nil
}
