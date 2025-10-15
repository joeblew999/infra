package web

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/starfederation/datastar-go/datastar"

	runtimeui "github.com/joeblew999/infra/core/pkg/runtime/ui"
	"github.com/joeblew999/infra/core/pkg/runtime/ui/live"
	"github.com/joeblew999/infra/core/pkg/runtime/ui/render"
	webtemplates "github.com/joeblew999/infra/core/pkg/runtime/ui/templates/web"
)

// Options controls how the web UI is served.
type Options struct {
	Page  string
	Store *live.Store
}

var (
	templateOnce sync.Once
	templateErr  error
	pageTemplate *template.Template
)

func loadTemplate() (*template.Template, error) {
	templateOnce.Do(func() {
		pageTemplate, templateErr = webtemplates.Parse()
	})
	return pageTemplate, templateErr
}

// Run starts the Datastar-enabled web UI driven by the snapshot data. The
// server exits when the provided context is cancelled.
func Run(ctx context.Context, listener net.Listener, out io.Writer, opts Options) error {
	if ctx == nil {
		ctx = context.Background()
	}

	tmpl, err := loadTemplate()
	if err != nil {
		return err
	}

	snapshot := runtimeui.LoadTestSnapshot()
	if opts.Store != nil {
		snapshot = opts.Store.Snapshot()
	}
	fallback := runtimeui.NormalizePage(snapshot, opts.Page)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		currentSnapshot := snapshot
		liveMode := opts.Store != nil
		if liveMode {
			currentSnapshot = opts.Store.Snapshot()
		}

		page := runtimeui.NormalizePage(currentSnapshot, r.URL.Query().Get("page"))
		if page == "" {
			page = fallback
		}

		data := render.NewViewModel("Core Runtime", currentSnapshot, page, liveMode)

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := tmpl.Execute(w, data); err != nil {
			http.Error(w, fmt.Sprintf("render template: %v", err), http.StatusInternalServerError)
		}
	})

	mux := http.NewServeMux()
	mux.Handle("/", handler)
	if opts.Store != nil {
		mux.Handle("/live", makeSSEHandler(opts.Store, tmpl, opts.Page))
		mux.Handle("/actions/events", makeEventMutationHandler(opts.Store))
		mux.Handle("/actions/process/start", makeProcessActionHandler(opts.Store, "start", func(ctx context.Context, port int, name string) error {
			return opts.Store.StartProcess(ctx, port, name)
		}))
		mux.Handle("/actions/process/stop", makeProcessActionHandler(opts.Store, "stop", func(ctx context.Context, port int, name string) error {
			return opts.Store.StopProcess(ctx, port, name)
		}))
		mux.Handle("/actions/process/restart", makeProcessActionHandler(opts.Store, "restart", func(ctx context.Context, port int, name string) error {
			return opts.Store.RestartProcess(ctx, port, name)
		}))
		mux.Handle("/actions/process/scale", makeProcessScaleHandler(opts.Store))
	}

	server := &http.Server{
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		if err := server.Serve(listener); err != nil {
			errCh <- err
		}
	}()

	if out != nil {
		fmt.Fprintf(out, "core web UI available at http://%s\n", listener.Addr().String())
	}

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
		if err := ctx.Err(); err != nil && !errors.Is(err, context.Canceled) {
			return err
		}
		return nil
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		if ne, ok := err.(*net.OpError); ok && ne.Op == "accept" {
			return nil
		}
		return err
	}
}

func makeEventMutationHandler(store *live.Store) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
		if err != nil {
			http.Error(w, fmt.Sprintf("read body: %v", err), http.StatusBadRequest)
			return
		}
		_ = r.Body.Close()

		message := ""
		if len(body) > 0 {
			var payload map[string]any
			if err := json.Unmarshal(body, &payload); err == nil {
				if raw, ok := payload["message"].(string); ok {
					message = strings.TrimSpace(raw)
				}
			}
			if message == "" {
				if values, err := url.ParseQuery(string(body)); err == nil {
					message = strings.TrimSpace(values.Get("message"))
				}
			}
		}

		if message == "" {
			r.Body = io.NopCloser(bytes.NewReader(body))
			if err := r.ParseForm(); err == nil {
				message = strings.TrimSpace(r.FormValue("message"))
			}
		}

		if message == "" {
			message = "manual event triggered"
		}

		store.AppendEvent(message)

		sse := datastar.NewSSE(w, r)
		_ = sse.MarshalAndPatchSignals(map[string]string{
			"eventAck": message,
		})
	})
}

func makeProcessActionHandler(store *live.Store, action string, fn func(context.Context, int, string) error) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		body, err := readProcessActionPayload(r)
		if err != nil {
			http.Error(w, fmt.Sprintf("read body: %v", err), http.StatusBadRequest)
			return
		}

		name := parseProcessName(body, r)
		if name == "" {
			http.Error(w, "missing process name", http.StatusBadRequest)
			return
		}

		port := store.ComposePort()
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		if err := fn(ctx, port, name); err != nil {
			http.Error(w, fmt.Sprintf("process %s failed: %v", action, err), http.StatusBadRequest)
			return
		}

		sse := datastar.NewSSE(w, r)
		_ = sse.MarshalAndPatchSignals(map[string]string{
			"processActionStatus": fmt.Sprintf("%s %s requested", action, name),
		})
	})
}

func makeProcessScaleHandler(store *live.Store) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		body, err := readProcessActionPayload(r)
		if err != nil {
			http.Error(w, fmt.Sprintf("read body: %v", err), http.StatusBadRequest)
			return
		}

		name := parseProcessName(body, r)
		if name == "" {
			http.Error(w, "missing process name", http.StatusBadRequest)
			return
		}
		count, ok := parseProcessCount(body, r)
		if !ok {
			http.Error(w, "missing process count", http.StatusBadRequest)
			return
		}

		port := store.ComposePort()
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		if err := store.ScaleProcess(ctx, port, name, count); err != nil {
			http.Error(w, fmt.Sprintf("process scale failed: %v", err), http.StatusBadRequest)
			return
		}

		sse := datastar.NewSSE(w, r)
		_ = sse.MarshalAndPatchSignals(map[string]string{
			"processActionStatus": fmt.Sprintf("scaled %s to %d", name, count),
		})
	})
}

func readProcessActionPayload(r *http.Request) ([]byte, error) {
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		return nil, err
	}
	_ = r.Body.Close()
	return body, nil
}

func parseProcessName(body []byte, r *http.Request) string {
	name := ""
	if len(body) > 0 {
		var payload map[string]any
		if err := json.Unmarshal(body, &payload); err == nil {
			if raw, ok := payload["name"].(string); ok {
				name = strings.TrimSpace(raw)
			}
		}
		if name == "" {
			if values, err := url.ParseQuery(string(body)); err == nil {
				name = strings.TrimSpace(values.Get("name"))
			}
		}
	}
	if name == "" {
		name = strings.TrimSpace(r.URL.Query().Get("name"))
	}
	return name
}

func parseProcessCount(body []byte, r *http.Request) (int, bool) {
	if len(body) > 0 {
		var payload map[string]any
		if err := json.Unmarshal(body, &payload); err == nil {
			switch value := payload["count"].(type) {
			case float64:
				return int(value), true
			case string:
				if v, err := strconv.Atoi(strings.TrimSpace(value)); err == nil {
					return v, true
				}
			}
		}
		if values, err := url.ParseQuery(string(body)); err == nil {
			if v := strings.TrimSpace(values.Get("count")); v != "" {
				if parsed, err := strconv.Atoi(v); err == nil {
					return parsed, true
				}
			}
		}
	}
	if v := strings.TrimSpace(r.URL.Query().Get("count")); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil {
			return parsed, true
		}
	}
	return 0, false
}

func makeSSEHandler(store *live.Store, tmpl *template.Template, defaultPage string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sse := datastar.NewSSE(w, r)
		ctx := r.Context()

		streamPage := r.URL.Query().Get("page")
		if streamPage == "" {
			streamPage = defaultPage
		}

		updates, cancel := store.Subscribe()
		defer cancel()

		for {
			select {
			case <-ctx.Done():
				return
			case snapshot, ok := <-updates:
				if !ok {
					return
				}
				vm := render.NewViewModel("Core Runtime", snapshot, streamPage, true)
				var buf bytes.Buffer
				if err := tmpl.ExecuteTemplate(&buf, "core_page", vm); err != nil {
					fmt.Fprintf(w, "render partial error: %v\n", err)
					continue
				}
				if err := sse.PatchElements(buf.String(), datastar.WithSelector("#core-page")); err != nil {
					fmt.Fprintf(w, "patch elements error: %v\n", err)
					return
				}

				var eventsBuf bytes.Buffer
				if err := tmpl.ExecuteTemplate(&eventsBuf, "recent_events_section", vm); err == nil {
					if err := sse.PatchElements(eventsBuf.String(), datastar.WithSelector("#recent-events")); err != nil {
						fmt.Fprintf(w, "patch events error: %v\n", err)
						return
					}
				}

				var eventsBody bytes.Buffer
				if err := tmpl.ExecuteTemplate(&eventsBody, "recent_events_body", vm); err == nil {
					if err := sse.MarshalAndPatchSignals(map[string]string{
						"recentEventsHTML": eventsBody.String(),
					}); err != nil {
						fmt.Fprintf(w, "patch events signals error: %v\n", err)
						return
					}
				}
			}
		}
	})
}
