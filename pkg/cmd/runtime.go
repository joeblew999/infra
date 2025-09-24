package cmd

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/joeblew999/infra/pkg/config"
	"github.com/joeblew999/infra/pkg/log"
	runtimeevents "github.com/joeblew999/infra/pkg/runtime/events"
	serviceruntime "github.com/joeblew999/infra/pkg/service/runtime"
	servicestate "github.com/joeblew999/infra/pkg/service/state"
	preflight "github.com/joeblew999/infra/pkg/workflows/preflight"
)

func init() {
	rootCmd.AddCommand(newRuntimeCmd())
}

func newRuntimeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "runtime",
		Short: "Control the local infrastructure supervisor",
		Long:  "Manage the supervised services that make up the local infrastructure runtime.",
	}

	cmd.AddCommand(newRuntimeUpCmd())
	cmd.AddCommand(newRuntimeDownCmd())
	cmd.AddCommand(newRuntimeListCmd())
	cmd.AddCommand(newRuntimeStatusCmd())
	cmd.AddCommand(newRuntimeWatchCmd())
	cmd.AddCommand(newRuntimeContainerCmd())

	return cmd
}

func newRuntimeUpCmd() *cobra.Command {
	var (
		env          string
		onlyServices []string
		skipServices []string
	)

	cmd := &cobra.Command{
		Use:   "up",
		Short: "Start the local runtime supervisor",
		RunE: func(cmd *cobra.Command, args []string) error {
			log.Info("Ensuring any existing runtime is shut down before startup...")
			serviceruntime.Shutdown()

			opts := defaultServiceOptions(env)
			opts.Preflight = preflight.RunIfNeeded

			onlyIDs, unknownOnly := resolveServiceFilters(onlyServices)
			if len(unknownOnly) > 0 {
				return fmt.Errorf("unknown services in --only: %s", strings.Join(unknownOnly, ", "))
			}
			skipIDs, unknownSkip := resolveServiceFilters(skipServices)
			if len(unknownSkip) > 0 {
				return fmt.Errorf("unknown services in --skip: %s", strings.Join(unknownSkip, ", "))
			}
			opts.OnlyServices = onlyIDs
			opts.SkipServices = skipIDs

			if env != "" {
				_ = os.Setenv("ENVIRONMENT", env)
			}

			return serviceruntime.Start(opts)
		},
	}

	cmd.Flags().StringVar(&env, "env", "development", "Environment profile to use")
	cmd.Flags().StringSliceVar(&onlyServices, "only", nil, "Limit runtime to a comma-separated list of services")
	cmd.Flags().StringSliceVar(&skipServices, "skip", nil, "Skip a comma-separated list of services")

	return cmd
}

func newRuntimeDownCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "down",
		Short: "Stop all supervised services",
		Run: func(cmd *cobra.Command, args []string) {
			serviceruntime.Shutdown()
		},
	}
}

func newRuntimeStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show the live runtime snapshot",
		RunE: func(cmd *cobra.Command, args []string) error {
			states := servicestate.Snapshot()
			if len(states) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "no services registered")
				return nil
			}

			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 2, 2, ' ', 0)
			fmt.Fprintln(w, "SERVICE\tSTATE\tPID\tPORT\tOWNERSHIP\tLAST ACTION")
			for _, state := range states {
				ownership := state.Ownership
				if ownership == "" {
					ownership = "free"
				}
				lastAction := state.LastAction
				if state.LastActionKind != "" {
					lastAction = fmt.Sprintf("%s (%s)", lastAction, state.LastActionKind)
				}
				fmt.Fprintf(
					w,
					"%s\t%s\t%d\t%d\t%s\t%s\n",
					state.Name,
					humanizeState(state.State, state.Running),
					state.PID,
					state.Port,
					ownership,
					lastAction,
				)
			}
			return w.Flush()
		},
	}
}

func newRuntimeListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List services available to the runtime supervisor",
		RunE: func(cmd *cobra.Command, args []string) error {
			specs := serviceruntime.AllServiceSpecs()
			if len(specs) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "no services registered")
				return nil
			}

			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 2, 2, ' ', 0)
			fmt.Fprintln(w, "SERVICE\tPORT\tREQUIRED\tDESCRIPTION")
			for _, spec := range specs {
				required := "optional"
				if spec.Required {
					required = "required"
				}
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", spec.ID, spec.Port, required, spec.Description)
			}
			return w.Flush()
		},
	}
}

func newRuntimeWatchCmd() *cobra.Command {
	var (
		baseURL      string
		serviceNames []string
		eventTypes   []string
		actionKinds  []string
		jsonOutput   bool
		skipInitial  bool
	)

	cmd := &cobra.Command{
		Use:   "watch",
		Short: "Stream live runtime events",
		RunE: func(cmd *cobra.Command, args []string) error {
			if baseURL == "" {
				baseURL = fmt.Sprintf("http://127.0.0.1:%s%s", config.GetWebServerPort(), config.StatusHTTPPath)
			}
			eventsURL := strings.TrimRight(baseURL, "/") + "/api/events"

			req, err := http.NewRequestWithContext(cmd.Context(), http.MethodGet, eventsURL, nil)
			if err != nil {
				return fmt.Errorf("build request: %w", err)
			}

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return fmt.Errorf("connect to %s: %w", eventsURL, err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				return fmt.Errorf("unexpected status: %s", resp.Status)
			}

			scanner := bufio.NewScanner(resp.Body)
			scanner.Buffer(make([]byte, 0, 4096), 1<<20)

			serviceFilter := makeStringSet(serviceNames)
			typeFilter := normalizeEventTypes(eventTypes)
			kindFilter := makeStringSet(actionKinds)

			var dataLines []string
			out := cmd.OutOrStdout()
			errOut := cmd.ErrOrStderr()

			fmt.Fprintf(out, "Watching runtime events from %s\n", eventsURL)

			flushEvent := func() {
				if len(dataLines) == 0 {
					return
				}
				payload := strings.Join(dataLines, "\n")
				dataLines = dataLines[:0]

				var evt runtimeEventEnvelope
				if err := json.Unmarshal([]byte(payload), &evt); err != nil {
					fmt.Fprintf(errOut, "decode event error: %v\n", err)
					return
				}

				if skipInitial && evt.Type == "initial-state" {
					return
				}
				if len(typeFilter) > 0 {
					if _, ok := typeFilter[evt.Type]; !ok {
						return
					}
				}
				if len(serviceFilter) > 0 {
					serviceID := strings.ToLower(evt.ServiceID)
					if serviceID == "" {
						return
					}
					if _, ok := serviceFilter[serviceID]; !ok {
						return
					}
				}
				if evt.Type == string(runtimeevents.EventServiceAction) && len(kindFilter) > 0 {
					kind := strings.ToLower(evt.Kind)
					if kind == "" {
						return
					}
					if _, ok := kindFilter[kind]; !ok {
						return
					}
				}

				if jsonOutput {
					fmt.Fprintln(out, payload)
					return
				}

				fmt.Fprintln(out, formatRuntimeEvent(evt))
			}

			for scanner.Scan() {
				line := scanner.Text()
				switch {
				case strings.HasPrefix(line, "data:"):
					dataLines = append(dataLines, strings.TrimSpace(line[5:]))
				case line == "":
					flushEvent()
				default:
					// ignore other SSE metadata lines
				}
			}

			flushEvent()

			if err := scanner.Err(); err != nil && !errors.Is(err, context.Canceled) {
				return fmt.Errorf("stream error: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&baseURL, "url", "", "Base URL for the status service (defaults to local web server)")
	cmd.Flags().StringSliceVar(&serviceNames, "service", nil, "Limit output to specific service IDs or names")
	cmd.Flags().StringSliceVar(&eventTypes, "types", nil, "Limit to event types (status, action, registered)")
	cmd.Flags().StringSliceVar(&actionKinds, "kinds", nil, "Limit service.action events to specific kinds (started, ensure_failed, ...)")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Emit raw JSON payloads")
	cmd.Flags().BoolVar(&skipInitial, "no-initial", false, "Skip the initial-state snapshot event")

	return cmd
}

func newRuntimeContainerCmd() *cobra.Command {
	var env string
	cmd := &cobra.Command{
		Use:   "container",
		Short: "Run the runtime inside a container using ko",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintln(cmd.OutOrStdout(), "Building and running containerized runtime...")
			return serviceruntime.RunContainer(env)
		},
	}

	cmd.Flags().StringVar(&env, "env", "production", "Environment profile for the containerized runtime")
	return cmd
}

type runtimeEventEnvelope struct {
	Type        string    `json:"type"`
	Timestamp   time.Time `json:"timestamp"`
	ServiceID   string    `json:"service_id"`
	Running     bool      `json:"running"`
	PID         int       `json:"pid"`
	Port        int       `json:"port"`
	Ownership   string    `json:"ownership"`
	State       string    `json:"state"`
	Message     string    `json:"message"`
	Kind        string    `json:"kind"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Required    bool      `json:"required"`
}

func formatRuntimeEvent(evt runtimeEventEnvelope) string {
	ts := evt.Timestamp.Local().Format(time.RFC3339)
	switch evt.Type {
	case "service.action":
		return fmt.Sprintf("%s [%s] action=%s message=%s", ts, evt.ServiceID, evt.Kind, evt.Message)
	case "service.status":
		state := evt.State
		if state == "" {
			if evt.Running {
				state = "running"
			} else {
				state = "pending"
			}
		}
		details := []string{fmt.Sprintf("state=%s", state)}
		if evt.PID != 0 {
			details = append(details, fmt.Sprintf("pid=%d", evt.PID))
		}
		if evt.Port != 0 {
			details = append(details, fmt.Sprintf("port=%d", evt.Port))
		}
		if evt.Ownership != "" {
			details = append(details, fmt.Sprintf("owner=%s", evt.Ownership))
		}
		if evt.Message != "" {
			details = append(details, fmt.Sprintf("msg=%s", evt.Message))
		}
		return fmt.Sprintf("%s [%s] %s", ts, evt.ServiceID, strings.Join(details, " "))
	case "service.registered":
		return fmt.Sprintf("%s [%s] registered name=%s required=%t", ts, evt.ServiceID, evt.Name, evt.Required)
	case "initial-state":
		return fmt.Sprintf("%s initial state snapshot", ts)
	default:
		return fmt.Sprintf("%s [%s] %s", ts, evt.ServiceID, evt.Type)
	}
}

func makeStringSet(values []string) map[string]struct{} {
	set := make(map[string]struct{}, len(values))
	for _, raw := range values {
		normalized := strings.ToLower(strings.TrimSpace(raw))
		if normalized == "" {
			continue
		}
		set[normalized] = struct{}{}
	}
	return set
}

func normalizeEventTypes(values []string) map[string]struct{} {
	set := make(map[string]struct{}, len(values))
	for _, raw := range values {
		n := strings.ToLower(strings.TrimSpace(raw))
		if n == "" {
			continue
		}
		switch n {
		case "status", "service.status":
			set[string(runtimeevents.EventServiceStatus)] = struct{}{}
		case "action", "service.action":
			set[string(runtimeevents.EventServiceAction)] = struct{}{}
		case "registered", "register", "service.registered":
			set[string(runtimeevents.EventServiceRegistered)] = struct{}{}
		case "initial", "initial-state":
			set["initial-state"] = struct{}{}
		default:
			set[n] = struct{}{}
		}
	}
	return set
}
func resolveServiceFilters(values []string) ([]serviceruntime.ServiceID, []string) {
	if len(values) == 0 {
		return nil, nil
	}
	var (
		ids      []serviceruntime.ServiceID
		unknowns []string
	)
	for _, raw := range values {
		candidate := strings.TrimSpace(strings.ToLower(raw))
		if candidate == "" {
			continue
		}
		if id, ok := serviceruntime.ResolveServiceID(candidate); ok {
			ids = append(ids, id)
		} else {
			unknowns = append(unknowns, raw)
		}
	}
	return ids, unknowns
}

func humanizeState(state string, running bool) string {
	switch state {
	case "running":
		return "running"
	case "pending":
		return "pending"
	case "ready":
		return "ready"
	case "reclaimed":
		return "reclaimed"
	case "blocked":
		return "blocked"
	case "error":
		return "error"
	case "stopped":
		return "stopped"
	default:
		if running {
			return "running"
		}
		return "stopped"
	}
}
