package tui

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
	"text/template"
	"time"

	uv "github.com/charmbracelet/ultraviolet"
	runtimeui "github.com/joeblew999/infra/core/pkg/runtime/ui"
	"github.com/joeblew999/infra/core/pkg/runtime/ui/live"
	"github.com/joeblew999/infra/core/pkg/runtime/ui/render"
	tuitemplates "github.com/joeblew999/infra/core/pkg/runtime/ui/templates/tui"
	xterm "golang.org/x/term"
)

var (
	tmplOnce sync.Once
	tmplErr  error
	tmpl     *template.Template
)

func loadTemplate() (*template.Template, error) {
	tmplOnce.Do(func() {
		tmpl, tmplErr = tuitemplates.Parse()
	})
	return tmpl, tmplErr
}

// Run renders the TUI snapshot snippet. When a live store is provided the output
// updates continuously by re-rendering the template whenever the store changes.
func Run(ctx context.Context, out io.Writer, page string, store *live.Store) error {
	template, err := loadTemplate()
	if err != nil {
		return err
	}

	if store == nil {
		snapshot := runtimeui.LoadTestSnapshot()
		vm := render.NewViewModel("Core Runtime", snapshot, page, false)
		return renderOnce(out, template, vm)
	}

	if isTTY(os.Stdin) && isTTY(os.Stdout) {
		return runInteractive(ctx, template, store, page)
	}

	return runStreaming(ctx, out, template, store, page)
}

func renderOnce(out io.Writer, tmpl *template.Template, vm render.ViewModel) error {
	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "tui", vm); err != nil {
		return fmt.Errorf("render tui template: %w", err)
	}

	if _, err := fmt.Fprintf(out, "\033[H\033[2J%s", buf.String()); err != nil {
		return fmt.Errorf("write tui: %w", err)
	}
	return nil
}

func runStreaming(ctx context.Context, out io.Writer, tmpl *template.Template, store *live.Store, page string) error {
	updates, cancel := store.Subscribe()
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return nil
		case snapshot, ok := <-updates:
			if !ok {
				return nil
			}
			vm := render.NewViewModel("Core Runtime", snapshot, page, true)
			if err := renderOnce(out, tmpl, vm); err != nil {
				return err
			}
			time.Sleep(50 * time.Millisecond)
		}
	}
}

func runInteractive(ctx context.Context, tmpl *template.Template, store *live.Store, page string) error {
	term := uv.DefaultTerminal()
	if err := term.Start(); err != nil {
		if errors.Is(err, uv.ErrNotTerminal) {
			return runStreaming(ctx, os.Stdout, tmpl, store, page)
		}
		return fmt.Errorf("start terminal: %w", err)
	}
	term.EnterAltScreen()
	term.HideCursor()

	session := newSession(ctx, term, tmpl, store, page)
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		defer cancel()
		_ = term.Shutdown(shutdownCtx)
	}()

	return session.run(ctx)
}

func isTTY(f *os.File) bool {
	if f == nil {
		return false
	}
	return xterm.IsTerminal(int(f.Fd()))
}
