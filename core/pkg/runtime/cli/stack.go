package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/joeblew999/infra/core/pkg/observability/events"
	runtimecfg "github.com/joeblew999/infra/core/pkg/runtime/config"
	"github.com/joeblew999/infra/core/pkg/runtime/process"
	caddyservice "github.com/joeblew999/infra/core/services/caddy"
	natssvc "github.com/joeblew999/infra/core/services/nats"
	pocketbasesvc "github.com/joeblew999/infra/core/services/pocketbase"
)

func newStackCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stack",
		Short: "Manage the local service stack",
	}

	cmd.AddCommand(buildStackUpCommand("up", "Run core services (NATS, PocketBase, Caddy) locally", true))
	cmd.AddCommand(buildStackDownCommand("down", "Stop core services started with stack up"))
	cmd.AddCommand(buildStackStatusCommand("status", "Show whether the local stack is running"))
	cmd.AddCommand(newStackCleanCommand())
	cmd.AddCommand(newStackDoctorCommand())
	cmd.AddCommand(newStackProcessesCommand())
	cmd.AddCommand(newStackProcessCommand())
	cmd.AddCommand(newStackProjectCommand())
	cmd.AddCommand(newStackReloadCommand())
	cmd.AddCommand(newStackObserveCommand())
	return cmd
}

func newUpCommand() *cobra.Command {
	return buildStackUpCommand("up", "Start the deterministic core stack (blocks until interrupted)", true)
}

func newDownCommand() *cobra.Command {
	return buildStackDownCommand("down", "Stop the deterministic core stack")
}

func newStatusCommand() *cobra.Command {
	return buildStackStatusCommand("status", "Show the state of the deterministic core stack")
}

func newStackProcessesCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "processes",
		Short: "List processes reported by Process Compose",
		RunE:  stackProcessesRun,
	}
	cmd.Flags().Int("compose-port", 0, "Process Compose port (defaults to PC_PORT_NUM or 28081)")
	cmd.Flags().Bool("json", false, "Output processes as JSON")
	return cmd
}

func newStackProjectCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "project",
		Short: "Inspect and update the Process Compose project",
	}
	cmd.PersistentFlags().Int("compose-port", 0, "Process Compose port (defaults to PC_PORT_NUM or 28081)")

	state := &cobra.Command{
		Use:   "state",
		Short: "Show the Process Compose project state",
		Args:  cobra.NoArgs,
		RunE:  stackProjectState,
	}
	state.Flags().Bool("with-memory", false, "Include memory statistics when supported")
	state.Flags().Bool("json", true, "Output project state as JSON")

	update := &cobra.Command{
		Use:   "update",
		Short: "Update the Process Compose project using a JSON payload",
		RunE:  stackProjectUpdate,
	}
	update.Flags().StringP("file", "f", "", "Path to JSON file describing project overrides")
	update.Flags().Bool("json", false, "Output update results as JSON")

	reload := newStackReloadCommand()

	cmd.AddCommand(state, update, reload)
	return cmd
}

func buildStackUpCommand(use, short string, includeRunAlias bool) *cobra.Command {
	cmd := &cobra.Command{
		Use:   use,
		Short: short,
		RunE:  stackUpRun,
	}
	if includeRunAlias {
		cmd.Aliases = append(cmd.Aliases, "run")
	}
	cmd.Flags().BoolP("detach", "d", false, "Run process-compose in detached mode (background)")
	return cmd
}

func buildStackDownCommand(use, short string) *cobra.Command {
	cmd := &cobra.Command{
		Use:     use,
		Aliases: []string{"stop"},
		Short:   short,
		RunE:    stackDownRun,
	}
	return cmd
}

func buildStackStatusCommand(use, short string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   use,
		Short: short,
		RunE:  stackStatusRun,
	}
	cmd.Flags().Bool("json", false, "Output status as JSON")
	return cmd
}

func newStackCleanCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clean",
		Short: "Stop services, kill zombie processes, and clean generated files",
		Long: `Clean the core stack by:
1. Stopping process-compose if running
2. Killing any zombie processes on stack ports (4222, 8090, 2015, 8222, 28081)
3. Removing generated files (.core-stack/ directory)

By default, performs a full clean (all steps).
Use flags to clean only specific parts:
  --processes    Kill zombie processes only (skip file removal)
  --files        Remove generated files only (skip process management)`,
		RunE: stackCleanRun,
	}
	cmd.Flags().Bool("processes", false, "Kill zombie processes only (skip file removal)")
	cmd.Flags().Bool("files", false, "Remove generated files only (skip process management)")
	return cmd
}

func newStackDoctorCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Run diagnostics on the stack and detect issues",
		Long: `Diagnose common stack issues:
- Port conflicts
- Health endpoint availability
- Config file validity
- .data directory permissions and tokens
- Zombie processes
- Process-compose connectivity

Provides actionable suggestions for fixing detected issues.`,
		RunE: stackDoctorRun,
	}
	cmd.Flags().Bool("verbose", false, "Show detailed diagnostic information")
	return cmd
}

func stackUpRun(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}
	return runComposeStack(cmd, args)
}

func stackDownRun(cmd *cobra.Command, args []string) error {
	return stopComposeStack(cmd, args)
}

func stackStatusRun(cmd *cobra.Command, args []string) error {
	asJSON, _ := cmd.Flags().GetBool("json")
	return statusComposeStack(cmd, args, asJSON)
}

func stackCleanRun(cmd *cobra.Command, args []string) error {
	processesOnly, _ := cmd.Flags().GetBool("processes")
	filesOnly, _ := cmd.Flags().GetBool("files")

	// Determine what to clean
	// If neither flag is set, do both (full clean)
	cleanProcesses := !filesOnly || processesOnly
	cleanFiles := !processesOnly || filesOnly

	out := cmd.OutOrStdout()

	// Step 1: Stop process-compose if running (unless files-only mode)
	if cleanProcesses {
		fmt.Fprintln(out, "→ Stopping process-compose...")
		port := process.ComposePort(args)
		if err := process.ShutdownCompose(cmd.Context(), port); err != nil {
			if !errors.Is(err, process.ErrComposeUnavailable) {
				fmt.Fprintf(out, "  ⚠ Failed to stop process-compose: %v\n", err)
			} else {
				fmt.Fprintln(out, "  ✓ Process-compose not running")
			}
		} else {
			fmt.Fprintln(out, "  ✓ Process-compose stopped")
		}

		// Step 2: Kill zombie processes on stack ports
		fmt.Fprintln(out, "\n→ Checking for zombie processes...")
		ports, err := getStackPorts()
		if err != nil {
			return fmt.Errorf("get stack ports: %w", err)
		}

		killedAny := false
		for _, port := range ports {
			if killed, err := killProcessOnPort(port); err != nil {
				fmt.Fprintf(out, "  ⚠ Port %d: %v\n", port, err)
			} else if killed {
				fmt.Fprintf(out, "  ✓ Port %d: killed process\n", port)
				killedAny = true
			}
		}
		if !killedAny {
			fmt.Fprintln(out, "  ✓ No zombie processes found")
		}
	}

	// Step 3: Remove generated files
	if cleanFiles {
		fmt.Fprintln(out, "\n→ Cleaning generated files...")
		coreStackDir := ".core-stack"
		if _, err := os.Stat(coreStackDir); err == nil {
			if err := os.RemoveAll(coreStackDir); err != nil {
				fmt.Fprintf(out, "  ⚠ Failed to remove %s: %v\n", coreStackDir, err)
			} else {
				fmt.Fprintf(out, "  ✓ Removed %s/\n", coreStackDir)
			}
		} else {
			fmt.Fprintf(out, "  ✓ %s/ does not exist\n", coreStackDir)
		}
	}

	fmt.Fprintln(out, "\n✅ Clean complete!")
	return nil
}

func stackDoctorRun(cmd *cobra.Command, args []string) error {
	verbose, _ := cmd.Flags().GetBool("verbose")
	out := cmd.OutOrStdout()

	fmt.Fprintln(out, "🔍 Running stack diagnostics...\n")

	issues := 0
	warnings := 0

	// 1. Check port availability
	fmt.Fprintln(out, "→ Checking port availability...")
	ports, err := getStackPorts()
	if err != nil {
		fmt.Fprintf(out, "  ⚠ Could not determine stack ports: %v\n", err)
		warnings++
	} else {
		for _, port := range ports {
			if isPortBusy(port) {
				if verbose {
					fmt.Fprintf(out, "  ✓ Port %d: in use (OK if stack is running)\n", port)
				}
			} else {
				if verbose {
					fmt.Fprintf(out, "  • Port %d: available\n", port)
				}
			}
		}
		if !verbose {
			fmt.Fprintln(out, "  ✓ Port check complete")
		}
	}

	// 2. Check process-compose connectivity
	fmt.Fprintln(out, "\n→ Checking process-compose...")
	port := process.ComposePort(args)
	states, pcErr := process.FetchComposeProcesses(cmd.Context(), port)
	if pcErr != nil {
		if errors.Is(pcErr, process.ErrComposeUnavailable) {
			fmt.Fprintf(out, "  ℹ Process-compose not running (port %d)\n", port)
			fmt.Fprintln(out, "    Run: go run ./cmd/core stack up")
		} else {
			fmt.Fprintf(out, "  ❌ Process-compose error: %v\n", pcErr)
			issues++
		}
	} else {
		fmt.Fprintf(out, "  ✓ Process-compose running (%d processes)\n", len(states))

		// Check individual process health
		if verbose {
			for _, state := range states {
				status := "✓"
				if !state.IsRunning {
					status = "❌"
					issues++
				}
				fmt.Fprintf(out, "    %s %s: %s (restarts: %d)\n",
					status, state.Name, state.Status, state.Restarts)
			}
		}
	}

	// 3. Check health endpoints
	fmt.Fprintln(out, "\n→ Checking health endpoints...")
	healthChecks := []struct {
		name string
		url  string
		port int
	}{
		{"NATS", "http://127.0.0.1:8222/healthz", 8222},
		{"PocketBase", "http://127.0.0.1:8090/api/health", 8090},
		{"Caddy", "http://127.0.0.1:2015/api/health", 2015},
	}

	for _, hc := range healthChecks {
		if !isPortBusy(hc.port) {
			if verbose {
				fmt.Fprintf(out, "  • %s: not running (port %d not in use)\n", hc.name, hc.port)
			}
			continue
		}

		resp, err := http.Get(hc.url)
		if err != nil {
			fmt.Fprintf(out, "  ❌ %s: health check failed (%v)\n", hc.name, err)
			issues++
		} else {
			resp.Body.Close()
			if resp.StatusCode == 200 {
				fmt.Fprintf(out, "  ✓ %s: healthy\n", hc.name)
			} else {
				fmt.Fprintf(out, "  ⚠ %s: returned status %d\n", hc.name, resp.StatusCode)
				warnings++
			}
		}
	}

	// 4. Check .data directory
	fmt.Fprintln(out, "\n→ Checking .data directory...")
	dataDir := ".data/core"
	if stat, err := os.Stat(dataDir); err == nil {
		if !stat.IsDir() {
			fmt.Fprintf(out, "  ❌ %s exists but is not a directory\n", dataDir)
			issues++
		} else {
			fmt.Fprintf(out, "  ✓ %s exists\n", dataDir)

			// Check for tokens
			flyToken := ".data/core/fly/settings.json"
			if _, err := os.Stat(flyToken); err == nil {
				fmt.Fprintln(out, "  ✓ Fly.io token found")
			} else if verbose {
				fmt.Fprintln(out, "  • Fly.io token not found (optional)")
			}

			cfToken := ".data/core/cloudflare/settings.json"
			if _, err := os.Stat(cfToken); err == nil {
				fmt.Fprintln(out, "  ✓ Cloudflare token found")
			} else if verbose {
				fmt.Fprintln(out, "  • Cloudflare token not found (optional)")
			}
		}
	} else {
		fmt.Fprintf(out, "  ⚠ %s not found (deployment tokens unavailable)\n", dataDir)
		warnings++
	}

	// 5. Check for zombie processes
	fmt.Fprintln(out, "\n→ Checking for zombie processes...")
	if ports != nil {
		foundZombies := false
		if pcErr != nil && errors.Is(pcErr, process.ErrComposeUnavailable) {
			// process-compose not running, check if any ports are in use
			for _, port := range ports {
				if isPortBusy(port) {
					fmt.Fprintf(out, "  ⚠ Port %d in use but process-compose not running\n", port)
					warnings++
					foundZombies = true
					if !verbose {
						fmt.Fprintln(out, "    Run: go run ./cmd/core stack clean --processes")
					}
				}
			}
		}
		if !foundZombies {
			fmt.Fprintln(out, "  ✓ No zombie processes detected")
		}
	}

	// Summary
	fmt.Fprintln(out, "\n"+strings.Repeat("━", 60))
	if issues == 0 && warnings == 0 {
		fmt.Fprintln(out, "✅ All checks passed! Stack is healthy.")
	} else {
		if issues > 0 {
			fmt.Fprintf(out, "❌ Found %d issue(s)\n", issues)
		}
		if warnings > 0 {
			fmt.Fprintf(out, "⚠  Found %d warning(s)\n", warnings)
		}
		fmt.Fprintln(out, "\nSuggested actions:")
		if pcErr != nil && errors.Is(pcErr, process.ErrComposeUnavailable) {
			fmt.Fprintln(out, "  • Start stack: go run ./cmd/core stack up")
		}
		if issues > 0 {
			fmt.Fprintln(out, "  • Check logs: go run ./cmd/core stack processes")
			fmt.Fprintln(out, "  • Clean and restart: go run ./cmd/core stack clean && go run ./cmd/core stack up")
		}
	}

	return nil
}

func stackProcessesRun(cmd *cobra.Command, args []string) error {
	port := composePortFromCmd(cmd)
	jsonOut, _ := cmd.Flags().GetBool("json")
	states, err := process.FetchComposeProcesses(cmd.Context(), port)
	if err != nil {
		if errors.Is(err, process.ErrComposeUnavailable) {
			return fmt.Errorf("process-compose supervisor unavailable on port %d", port)
		}
		return err
	}
	if jsonOut {
		payload := composeProcessesPayload{
			Port:    port,
			Process: states,
		}
		return writeJSON(cmd.OutOrStdout(), payload)
	}
	printComposeStates(cmd.OutOrStdout(), states)
	return nil
}

func stackProjectState(cmd *cobra.Command, args []string) error {
	port := composePortFromCmd(cmd)
	withMemory, _ := cmd.Flags().GetBool("with-memory")
	jsonOut, _ := cmd.Flags().GetBool("json")
	state, err := process.GetComposeProjectState(cmd.Context(), port, withMemory)
	if err != nil {
		return err
	}
	payload := map[string]any{"port": port, "state": state}
	if jsonOut {
		return writeJSON(cmd.OutOrStdout(), payload)
	}
	fmt.Fprintf(cmd.OutOrStdout(), "Process Compose project state on port %d\n", port)
	if err := printProjectState(cmd.OutOrStdout(), state); err != nil {
		return err
	}
	return nil
}

func stackProjectUpdate(cmd *cobra.Command, args []string) error {
	port := composePortFromCmd(cmd)
	file, _ := cmd.Flags().GetString("file")
	jsonOut, _ := cmd.Flags().GetBool("json")
	if file == "" {
		return errors.New("no update payload provided; use --file <path> or --file - for stdin")
	}
	var data []byte
	var err error
	if file == "-" {
		data, err = io.ReadAll(cmd.InOrStdin())
		if err != nil {
			return fmt.Errorf("read update payload: %w", err)
		}
	} else {
		data, err = os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("read update payload: %w", err)
		}
	}
	data = bytes.TrimSpace(data)
	if len(data) == 0 {
		return errors.New("empty update payload")
	}
	var payload json.RawMessage
	if err := json.Unmarshal(data, &payload); err != nil {
		return fmt.Errorf("invalid JSON payload: %w", err)
	}
	result, err := process.UpdateComposeProject(cmd.Context(), port, data)
	if err != nil {
		return err
	}
	if jsonOut {
		return writeJSON(cmd.OutOrStdout(), map[string]any{"port": port, "status": result})
	}
	fmt.Fprintf(cmd.OutOrStdout(), "Updated Process Compose project on port %d\n", port)
	if len(result) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No services reported changes")
		return nil
	}
	keys := make([]string, 0, len(result))
	for name := range result {
		keys = append(keys, name)
	}
	sort.Strings(keys)
	for _, name := range keys {
		fmt.Fprintf(cmd.OutOrStdout(), "• %-16s -> %s\n", name, result[name])
	}
	return nil
}

func printServiceExpectations(out io.Writer, stackRunning bool) {
	services, err := collectServiceStatuses()
	if err != nil {
		fmt.Fprintf(out, "(unable to inspect service ports: %v)\n", err)
		return
	}
	printServiceExpectationsWith(out, services, stackRunning)
}

func printServiceExpectationsWith(out io.Writer, services []stackServiceStatus, stackRunning bool) {
	if len(services) == 0 {
		return
	}
	fmt.Fprintln(out, "Services:")
	for _, svc := range services {
		state := describeServiceState(stackRunning, svc.Running)
		fmt.Fprintf(out, "• %-12s port %-5d status: %-11s %s\n", svc.Name, svc.Port, state, svc.About)
	}
}

func printProjectState(out io.Writer, state process.ProjectState) error {
	if len(state) == 0 {
		_, err := fmt.Fprintln(out, "(empty project state)")
		return err
	}
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(out, string(data))
	return err
}

func printComposeProcess(out io.Writer, st process.ComposeProcessState) {
	fmt.Fprintf(out, "Process: %s\n", st.Name)
	if st.Namespace != "" {
		fmt.Fprintf(out, "Namespace: %s\n", st.Namespace)
	}
	fmt.Fprintf(out, "Status: %s\n", st.Status)
	if st.HasHealthProbe {
		fmt.Fprintf(out, "Health: %s\n", st.Health)
	} else {
		fmt.Fprintln(out, "Health: -")
	}
	fmt.Fprintf(out, "Running: %t\n", st.IsRunning)
	fmt.Fprintf(out, "Restarts: %d\n", st.Restarts)
	if !st.IsRunning {
		fmt.Fprintf(out, "Exit Code: %d\n", st.ExitCode)
	}
}

func printProcessLogs(out io.Writer, name string, logs []string) {
	fmt.Fprintf(out, "Logs for %s:\n", name)
	if len(logs) == 0 {
		fmt.Fprintln(out, "(no logs)")
		return
	}
	for _, line := range logs {
		if strings.HasSuffix(line, "\n") {
			fmt.Fprint(out, line)
			continue
		}
		fmt.Fprintln(out, line)
	}
}

func printComposeStates(out io.Writer, states []process.ComposeProcessState) {
	if len(states) == 0 {
		return
	}
	sort.Slice(states, func(i, j int) bool {
		if states[i].Namespace == states[j].Namespace {
			return states[i].Name < states[j].Name
		}
		return states[i].Namespace < states[j].Namespace
	})
	fmt.Fprintln(out, "Process Compose:")
	for _, st := range states {
		health := "-"
		if st.HasHealthProbe {
			health = st.Health
		}
		extra := ""
		if !st.IsRunning && st.ExitCode != 0 {
			extra = fmt.Sprintf(" exit=%d", st.ExitCode)
		}
		if st.Namespace != "" {
			fmt.Fprintf(out, "• %s/%-10s status: %-12s health: %-8s restarts: %d%s\n",
				st.Namespace, st.Name, st.Status, health, st.Restarts, extra)
			continue
		}
		fmt.Fprintf(out, "• %-12s status: %-12s health: %-8s restarts: %d%s\n",
			st.Name, st.Status, health, st.Restarts, extra)
	}
}

func describeServiceState(stackRunning, portBusy bool) string {
	switch {
	case stackRunning && portBusy:
		return "running"
	case stackRunning && !portBusy:
		return "starting"
	case !stackRunning && portBusy:
		return "orphaned"
	default:
		return "stopped"
	}
}

type stackServiceStatus struct {
	Name    string
	Port    int
	About   string
	Running bool
}

type stackServiceSummary struct {
	Name   string `json:"name"`
	Port   int    `json:"port"`
	Status string `json:"status"`
	About  string `json:"about"`
}

type composeStatusPayload struct {
	Mode     string                        `json:"mode"`
	Port     int                           `json:"port"`
	Ports    []int                         `json:"ports"`
	Running  bool                          `json:"running"`
	Compose  []process.ComposeProcessState `json:"processes"`
	Services []stackServiceSummary         `json:"services"`
	Warning  string                        `json:"warning,omitempty"`
}

type composeProcessesPayload struct {
	Port    int                           `json:"port"`
	Process []process.ComposeProcessState `json:"processes"`
}

func collectServiceStatuses() ([]stackServiceStatus, error) {
	var statuses []stackServiceStatus

	natsSpec, err := natssvc.LoadSpec()
	if err != nil {
		return nil, fmt.Errorf("nats: %w", err)
	}
	statuses = append(statuses, stackServiceStatus{
		Name:    "nats",
		Port:    natsSpec.Ports.Client.Port,
		About:   "client → 4222",
		Running: isPortBusy(natsSpec.Ports.Client.Port),
	})

	pbSpec, err := pocketbasesvc.LoadSpec()
	if err != nil {
		return nil, fmt.Errorf("pocketbase: %w", err)
	}
	statuses = append(statuses, stackServiceStatus{
		Name:    "pocketbase",
		Port:    pbSpec.Ports.Primary.Port,
		About:   "primary → 8090",
		Running: isPortBusy(pbSpec.Ports.Primary.Port),
	})

	caddyCfg, err := caddyservice.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("caddy: %w", err)
	}
	statuses = append(statuses, stackServiceStatus{
		Name:    "caddy",
		Port:    caddyCfg.Ports.HTTP.Port,
		About:   "http → 2015",
		Running: isPortBusy(caddyCfg.Ports.HTTP.Port),
	})

	return statuses, nil
}

func servicesFromStatuses(statuses []stackServiceStatus, stackRunning bool) []stackServiceSummary {
	if len(statuses) == 0 {
		return nil
	}
	result := make([]stackServiceSummary, len(statuses))
	for i, svc := range statuses {
		result[i] = stackServiceSummary{
			Name:   svc.Name,
			Port:   svc.Port,
			Status: describeServiceState(stackRunning, svc.Running),
			About:  svc.About,
		}
	}
	return result
}

func writeJSON(out io.Writer, payload any) error {
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(out, string(data))
	return err
}

func newStackProcessCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "process",
		Short: "Control individual processes managed by Process Compose",
	}
	cmd.PersistentFlags().Int("compose-port", 0, "Process Compose port (defaults to PC_PORT_NUM or 28081)")

	start := &cobra.Command{
		Use:   "start NAME",
		Short: "Start a stopped process",
		Args:  cobra.ExactArgs(1),
		RunE:  stackProcessStart,
	}

	stop := &cobra.Command{
		Use:   "stop NAME [NAME...]",
		Short: "Stop one or more processes",
		Args:  cobra.MinimumNArgs(1),
		RunE:  stackProcessStop,
	}
	stop.Flags().Bool("json", false, "Output stop results as JSON")

	restart := &cobra.Command{
		Use:   "restart NAME",
		Short: "Restart a process",
		Args:  cobra.ExactArgs(1),
		RunE:  stackProcessRestart,
	}

	scale := &cobra.Command{
		Use:   "scale NAME COUNT",
		Short: "Scale a process to the desired count",
		Args:  cobra.ExactArgs(2),
		RunE:  stackProcessScale,
	}

	logs := &cobra.Command{
		Use:   "logs NAME",
		Short: "Show recent logs for a process",
		Args:  cobra.ExactArgs(1),
		RunE:  stackProcessLogs,
	}
	logs.Flags().Int("lines", 100, "Number of log lines to fetch (0 for all available)")
	logs.Flags().Int("end-offset", 0, "Offset from the end of the log before reading (0 for latest)")
	logs.Flags().Bool("json", false, "Output logs as JSON")

	truncate := &cobra.Command{
		Use:   "truncate NAME",
		Short: "Truncate stored logs for a process",
		Args:  cobra.ExactArgs(1),
		RunE:  stackProcessTruncate,
	}
	truncate.Flags().Bool("json", false, "Output truncate result as JSON")

	info := &cobra.Command{
		Use:   "info NAME",
		Short: "Show detailed state for a process",
		Args:  cobra.ExactArgs(1),
		RunE:  stackProcessInfo,
	}
	info.Flags().Bool("json", false, "Output process info as JSON")

	list := &cobra.Command{
		Use:   "list",
		Short: "List processes reported by Process Compose",
		RunE:  stackProcessesRun,
	}
	list.Flags().Bool("json", false, "Output processes as JSON")

	cmd.AddCommand(start, stop, restart, scale, logs, truncate, info, list)
	return cmd
}

func newStackReloadCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reload",
		Short: "Reload the Process Compose project from the generated configuration",
		RunE:  stackReloadRun,
	}
	cmd.Flags().Int("compose-port", 0, "Process Compose port (defaults to PC_PORT_NUM or 28081)")
	cmd.Flags().Bool("json", false, "Output reload results as JSON")
	return cmd
}

func stackProcessStart(cmd *cobra.Command, args []string) error {
	port := composePortFromCmd(cmd)
	name := args[0]
	if err := process.StartComposeProcess(cmd.Context(), port, name); err != nil {
		return err
	}
	fmt.Fprintf(cmd.OutOrStdout(), "Started process %s\n", name)
	return nil
}

func stackProcessStop(cmd *cobra.Command, args []string) error {
	port := composePortFromCmd(cmd)
	jsonOut, _ := cmd.Flags().GetBool("json")
	if len(args) == 1 && !jsonOut {
		if err := process.StopComposeProcess(cmd.Context(), port, args[0]); err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Stopped process %s\n", args[0])
		return nil
	}
	result, err := process.StopComposeProcesses(cmd.Context(), port, args)
	if err != nil {
		return err
	}
	if jsonOut {
		return writeJSON(cmd.OutOrStdout(), map[string]any{"port": port, "stopped": result})
	}
	if len(result) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No processes reported stopped")
		return nil
	}
	keys := make([]string, 0, len(result))
	for name := range result {
		keys = append(keys, name)
	}
	sort.Strings(keys)
	for _, name := range keys {
		fmt.Fprintf(cmd.OutOrStdout(), "Stopped %-16s -> %s\n", name, result[name])
	}
	return nil
}

func stackProcessInfo(cmd *cobra.Command, args []string) error {
	port := composePortFromCmd(cmd)
	jsonOut, _ := cmd.Flags().GetBool("json")
	name := args[0]
	state, err := process.FetchComposeProcess(cmd.Context(), port, name)
	if err != nil {
		if errors.Is(err, process.ErrComposeProcessNotFound) {
			return fmt.Errorf("process %q not found", name)
		}
		return err
	}
	if jsonOut {
		return writeJSON(cmd.OutOrStdout(), map[string]any{"port": port, "process": state})
	}
	printComposeProcess(cmd.OutOrStdout(), *state)
	return nil
}

func stackProcessRestart(cmd *cobra.Command, args []string) error {
	port := composePortFromCmd(cmd)
	name := args[0]
	if err := process.RestartComposeProcess(cmd.Context(), port, name); err != nil {
		return err
	}
	fmt.Fprintf(cmd.OutOrStdout(), "Restarted process %s\n", name)
	return nil
}

func stackProcessScale(cmd *cobra.Command, args []string) error {
	port := composePortFromCmd(cmd)
	count, err := strconv.Atoi(args[1])
	if err != nil {
		return fmt.Errorf("invalid scale count %q", args[1])
	}
	if err := process.ScaleComposeProcess(cmd.Context(), port, args[0], count); err != nil {
		return err
	}
	fmt.Fprintf(cmd.OutOrStdout(), "Scaled process %s to %d\n", args[0], count)
	return nil
}

func stackProcessLogs(cmd *cobra.Command, args []string) error {
	port := composePortFromCmd(cmd)
	lines, _ := cmd.Flags().GetInt("lines")
	endOffset, _ := cmd.Flags().GetInt("end-offset")
	jsonOut, _ := cmd.Flags().GetBool("json")
	if lines < 0 {
		lines = 0
	}
	if endOffset < 0 {
		endOffset = 0
	}
	name := args[0]
	logs, err := process.FetchComposeProcessLogs(cmd.Context(), port, name, endOffset, lines)
	if err != nil {
		return err
	}
	if jsonOut {
		return writeJSON(cmd.OutOrStdout(), map[string]any{
			"port":   port,
			"name":   name,
			"logs":   logs,
			"offset": endOffset,
			"lines":  lines,
		})
	}
	printProcessLogs(cmd.OutOrStdout(), name, logs)
	return nil
}

func stackProcessTruncate(cmd *cobra.Command, args []string) error {
	port := composePortFromCmd(cmd)
	jsonOut, _ := cmd.Flags().GetBool("json")
	name := args[0]
	if err := process.TruncateComposeProcessLogs(cmd.Context(), port, name); err != nil {
		return err
	}
	if jsonOut {
		return writeJSON(cmd.OutOrStdout(), map[string]any{"port": port, "name": name, "truncated": true})
	}
	fmt.Fprintf(cmd.OutOrStdout(), "Truncated logs for %s\n", name)
	return nil
}

func stackReloadRun(cmd *cobra.Command, args []string) error {
	port := composePortFromCmd(cmd)
	jsonOut, _ := cmd.Flags().GetBool("json")
	result, err := process.ReloadComposeProject(cmd.Context(), port)
	if err != nil {
		return err
	}
	if jsonOut {
		return writeJSON(cmd.OutOrStdout(), map[string]any{"port": port, "status": result})
	}
	fmt.Fprintf(cmd.OutOrStdout(), "Reloaded Process Compose project on port %d\n", port)
	if len(result) > 0 {
		keys := make([]string, 0, len(result))
		for name := range result {
			keys = append(keys, name)
		}
		sort.Strings(keys)
		for _, name := range keys {
			fmt.Fprintf(cmd.OutOrStdout(), "• %-16s -> %s\n", name, result[name])
		}
	}
	return nil
}

func composePortFromCmd(cmd *cobra.Command) int {
	if flag := cmd.Flags().Lookup("compose-port"); flag != nil {
		if v, err := strconv.Atoi(flag.Value.String()); err == nil && v > 0 {
			return v
		}
	}
	if flag := cmd.InheritedFlags().Lookup("compose-port"); flag != nil {
		if v, err := strconv.Atoi(flag.Value.String()); err == nil && v > 0 {
			return v
		}
	}
	return process.ComposePort(nil)
}

func runComposeStack(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}
	cfg := runtimecfg.Load()
	if err := process.EnsureServiceBinaries(cfg.Paths.AppRoot); err != nil {
		return fmt.Errorf("ensure service binaries: %w", err)
	}

	// Check if --detach flag is set
	detach, _ := cmd.Flags().GetBool("detach")
	composeArgs := []string{"up"}
	if detach {
		composeArgs = append(composeArgs, "--detached")
	}
	composeArgs = append(composeArgs, args...)

	return process.ExecuteCompose(ctx, cfg.Paths.AppRoot, composeArgs...)
}

func stopComposeStack(cmd *cobra.Command, args []string) error {
	port := process.ComposePort(args)
	if err := process.ShutdownCompose(cmd.Context(), port); err != nil {
		if errors.Is(err, process.ErrComposeUnavailable) {
			fmt.Fprintln(cmd.OutOrStdout(), "Stack already stopped")
			return nil
		}
		return fmt.Errorf("process-compose down: %w", err)
	}
	fmt.Fprintln(cmd.OutOrStdout(), "Stopped core services")
	return nil
}

func statusComposeStack(cmd *cobra.Command, args []string, asJSON bool) error {
	port := process.ComposePort(args)
	states, err := process.FetchComposeProcesses(cmd.Context(), port)
	composeRunning := false
	if err != nil {
		if !errors.Is(err, process.ErrComposeUnavailable) {
			fmt.Fprintf(cmd.ErrOrStderr(), "%v\n", err)
		}
	} else {
		for _, st := range states {
			if st.IsRunning {
				composeRunning = true
				break
			}
		}
	}
	ports, err := getStackPorts()
	if err != nil {
		return err
	}
	running := composeRunning
	if !running {
		for _, port := range ports {
			if isPortBusy(port) {
				running = true
				break
			}
		}
	}
	services, serr := collectServiceStatuses()
	if asJSON {
		payload := composeStatusPayload{
			Mode:     composeStatusMode,
			Port:     port,
			Ports:    ports,
			Running:  running,
			Compose:  states,
			Services: servicesFromStatuses(services, running),
		}
		if err != nil && !errors.Is(err, process.ErrComposeUnavailable) {
			payload.Warning = err.Error()
		} else if serr != nil {
			payload.Warning = serr.Error()
		}
		return writeJSON(cmd.OutOrStdout(), payload)
	}
	if serr != nil {
		fmt.Fprintf(cmd.OutOrStdout(), "(unable to inspect service ports: %v)\n", serr)
	}
	if running {
		fmt.Fprintln(cmd.OutOrStdout(), "Stack status: running (process-compose)")
		fmt.Fprintf(cmd.OutOrStdout(), "Ports in use: %v\n", ports)
	} else {
		fmt.Fprintln(cmd.OutOrStdout(), "Stack status: stopped (process-compose)")
	}
	printComposeStates(cmd.OutOrStdout(), states)
	printServiceExpectationsWith(cmd.OutOrStdout(), services, running)
	return nil
}

func ensurePortsFree(ports []int) error {
	for _, port := range ports {
		if isPortBusy(port) {
			return fmt.Errorf("port %d is already in use; ensure no other stack is running", port)
		}
	}
	return nil
}

func waitPortsFree(ports []int, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for {
		busy := false
		for _, port := range ports {
			if isPortBusy(port) {
				busy = true
				break
			}
		}
		if !busy {
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("ports still busy after %s", timeout)
		}
		time.Sleep(250 * time.Millisecond)
	}
}

func isPortBusy(port int) bool {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", port), 150*time.Millisecond)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}

// killProcessOnPort kills any process listening on the given port.
// Returns (true, nil) if a process was killed, (false, nil) if no process found.
func killProcessOnPort(port int) (bool, error) {
	if !isPortBusy(port) {
		return false, nil
	}

	// Use lsof to find the PID listening on this port
	cmd := fmt.Sprintf("lsof -ti :%d", port)
	output, err := exec.Command("sh", "-c", cmd).Output()
	if err != nil {
		// No process found or lsof failed
		return false, nil
	}

	pidStr := strings.TrimSpace(string(output))
	if pidStr == "" {
		return false, nil
	}

	// Kill the process
	killCmd := fmt.Sprintf("kill -9 %s", pidStr)
	if err := exec.Command("sh", "-c", killCmd).Run(); err != nil {
		return false, fmt.Errorf("kill failed: %w", err)
	}

	// Give it a moment to die
	time.Sleep(100 * time.Millisecond)
	return true, nil
}

func getStackPorts() ([]int, error) {
	ports := []int{}

	type portExtractor func() (int, error)

	extractors := []portExtractor{
		func() (int, error) {
			spec, err := natssvc.LoadSpec()
			if err != nil {
				return 0, fmt.Errorf("load nats spec: %w", err)
			}
			return spec.Ports.Client.Port, nil
		},
		func() (int, error) {
			spec, err := pocketbasesvc.LoadSpec()
			if err != nil {
				return 0, fmt.Errorf("load pocketbase spec: %w", err)
			}
			return spec.Ports.Primary.Port, nil
		},
		func() (int, error) {
			cfg, err := caddyservice.LoadConfig()
			if err != nil {
				return 0, fmt.Errorf("load caddy config: %w", err)
			}
			return cfg.Ports.HTTP.Port, nil
		},
	}

	for _, extractor := range extractors {
		port, err := extractor()
		if err != nil {
			return nil, err
		}
		if port > 0 {
			ports = append(ports, port)
		}
	}

	return ports, nil
}

const composeStatusMode = "process-compose"

func newStackObserveCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "observe",
		Short: "Process observability and event streaming",
		Long: `Monitor process-compose events and stream them to NATS JetStream.

The observe command provides real-time visibility into process lifecycle events
including starts, stops, crashes, restarts, and health changes.`,
	}

	cmd.AddCommand(newStackObserveAdapterCommand())
	cmd.AddCommand(newStackObserveWatchCommand())

	return cmd
}

func newStackObserveAdapterCommand() *cobra.Command {
	var (
		composePort  int
		natsURL      string
		pollInterval time.Duration
	)

	cmd := &cobra.Command{
		Use:   "adapter",
		Short: "Run the event adapter (polls process-compose and publishes to NATS)",
		Long: `The event adapter continuously polls process-compose for state changes
and publishes lifecycle events to NATS JetStream.

Events are published to subjects like:
  core.process.{name}.started
  core.process.{name}.crashed
  core.process.{name}.healthy

Run this adapter in the background to enable real-time process monitoring.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			adapter, err := events.NewAdapter(events.Config{
				ComposePort:  composePort,
				NATSURL:      natsURL,
				PollInterval: pollInterval,
			})
			if err != nil {
				return fmt.Errorf("create adapter: %w", err)
			}

			if err := adapter.Start(); err != nil {
				return fmt.Errorf("start adapter: %w", err)
			}

			fmt.Fprintln(cmd.OutOrStdout(), "Event adapter running...")
			fmt.Fprintf(cmd.OutOrStdout(), "  Process Compose: http://127.0.0.1:%d\n", composePort)
			fmt.Fprintf(cmd.OutOrStdout(), "  NATS: %s\n", natsURL)
			fmt.Fprintf(cmd.OutOrStdout(), "  Poll interval: %s\n", pollInterval)
			fmt.Fprintln(cmd.OutOrStdout(), "\nPress Ctrl+C to stop")

			// Wait for interrupt
			sigCh := make(chan os.Signal, 1)
			signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
			<-sigCh

			fmt.Fprintln(cmd.OutOrStdout(), "\nStopping adapter...")
			return adapter.Stop()
		},
	}

	cmd.Flags().IntVar(&composePort, "compose-port", 28081, "Process Compose API port")
	cmd.Flags().StringVar(&natsURL, "nats-url", "nats://127.0.0.1:4222", "NATS server URL")
	cmd.Flags().DurationVar(&pollInterval, "poll-interval", 2*time.Second, "How often to poll for state changes")

	return cmd
}

func newStackObserveWatchCommand() *cobra.Command {
	var (
		natsURL    string
		process    string
		eventType  string
		jsonOutput bool
	)

	cmd := &cobra.Command{
		Use:   "watch",
		Short: "Watch process events in real-time",
		Long: `Subscribe to process events from NATS and display them in real-time.

Examples:
  # Watch all events
  core stack observe watch

  # Watch events for a specific process
  core stack observe watch --process nats

  # Watch only crash events
  core stack observe watch --type crashed

  # Watch crashes for specific process
  core stack observe watch --process pocketbase --type crashed`,
		RunE: func(cmd *cobra.Command, args []string) error {
			consumer, err := events.NewConsumer(natsURL)
			if err != nil {
				return fmt.Errorf("create consumer: %w", err)
			}
			defer consumer.Close()

			if err := consumer.Connect(); err != nil {
				return fmt.Errorf("connect: %w", err)
			}

			// Build subscription pattern
			var pattern string
			if process != "" && eventType != "" {
				pattern = events.SubjectPattern(
					events.ForProcessAndType(process, events.EventType(eventType)),
				)
			} else if process != "" {
				pattern = events.SubjectPattern(events.ForProcess(process))
			} else if eventType != "" {
				pattern = events.SubjectPattern(events.ForEventType(events.EventType(eventType)))
			} else {
				pattern = events.SubjectPattern(events.AllEvents())
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Watching events: %s\n", pattern)
			fmt.Fprintln(cmd.OutOrStdout(), "Press Ctrl+C to stop\n")

			// Subscribe with handler
			handler := func(evt events.Event) error {
				if jsonOutput {
					// JSON output
					data, _ := evt.MarshalJSON()
					fmt.Fprintln(cmd.OutOrStdout(), string(data))
				} else {
					// Human-readable output with timestamp and severity
					timestamp := evt.Timestamp.Format("15:04:05.000")
					severity := evt.Severity()
					icon := severityIcon(severity)
					fmt.Fprintf(cmd.OutOrStdout(), "%s %s %s\n", timestamp, icon, evt.String())
				}
				return nil
			}

			if err := consumer.Subscribe(pattern, handler); err != nil {
				return fmt.Errorf("subscribe: %w", err)
			}

			// Wait for interrupt
			ctx, cancel := context.WithCancel(cmd.Context())
			defer cancel()

			sigCh := make(chan os.Signal, 1)
			signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

			select {
			case <-sigCh:
				fmt.Fprintln(cmd.OutOrStdout(), "\nStopping...")
			case <-ctx.Done():
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&natsURL, "nats-url", "nats://127.0.0.1:4222", "NATS server URL")
	cmd.Flags().StringVarP(&process, "process", "p", "", "Filter by process name")
	cmd.Flags().StringVarP(&eventType, "type", "t", "", "Filter by event type (started, stopped, crashed, healthy, unhealthy)")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output events as JSON")

	return cmd
}

func severityIcon(severity events.Severity) string {
	switch severity {
	case events.SeverityError:
		return "❌"
	case events.SeverityWarning:
		return "⚠️ "
	case events.SeverityInfo:
		return "ℹ️ "
	case events.SeverityDebug:
		return "🐛"
	default:
		return "  "
	}
}
