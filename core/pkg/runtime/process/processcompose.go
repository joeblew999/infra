package process

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"

	runtimecfg "github.com/joeblew999/infra/core/pkg/runtime/config"
	sharedcfg "github.com/joeblew999/infra/core/pkg/shared/config"
	caddyservice "github.com/joeblew999/infra/core/services/caddy"
	natssvc "github.com/joeblew999/infra/core/services/nats"
	pocketbasesvc "github.com/joeblew999/infra/core/services/pocketbase"
)

const (
	// StackStateDirName is where stack metadata (compose file, pid, etc.) lives.
	StackStateDirName = ".core-stack"
	// ComposeFileName is the generated Process Compose configuration filename.
	ComposeFileName = "process-compose.yaml"
)

const composeServerPort = 28081

// GenerateComposeConfig writes a process-compose configuration based on the
// embedded service manifests and returns the absolute path to the config.
func GenerateComposeConfig(appRoot string) (string, error) {
	restore := overrideAppRoot(appRoot)
	defer restore()

	cfg := runtimecfg.Load()
	root := cfg.Paths.AppRoot
	stateDir := filepath.Join(root, StackStateDirName)
	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		return "", fmt.Errorf("create compose state dir: %w", err)
	}
	composePath := filepath.Join(stateDir, ComposeFileName)

	composeDef, err := buildComposeDefinition(root)
	if err != nil {
		return "", err
	}
	data, err := yaml.Marshal(composeDef)
	if err != nil {
		return "", fmt.Errorf("marshal compose config: %w", err)
	}
	if err := os.WriteFile(composePath, data, 0o644); err != nil {
		return "", fmt.Errorf("write compose config: %w", err)
	}
	return composePath, nil
}

// ExecuteCompose invokes Process Compose with the generated configuration. The
// args should include the subcommand (e.g. "up", "down", "status").
func ExecuteCompose(ctx context.Context, appRoot string, args ...string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	composePath, err := GenerateComposeConfig(appRoot)
	if err != nil {
		return err
	}

	command := "up"
	if len(args) > 0 {
		command = args[0]
	}
	tail := []string{}
	if len(args) > 1 {
		tail = append(tail, args[1:]...)
	}
	port := resolveComposePort(tail)

	if command == "up" {
		if !hasFlag(tail, "--tui", "-t") {
			tail = append([]string{"--tui=false"}, tail...)
		}
		if !hasFlag(tail, "--keep-project") {
			tail = append([]string{"--keep-project"}, tail...)
		}
	}

	// Build path to process-compose binary
	composeBinPath := filepath.Join(appRoot, ".dep", "process-compose")

	// Build cmdArgs for process-compose
	cmdArgs := []string{}
	if composeCommandNeedsConfig(command) {
		cmdArgs = append(cmdArgs, "--config", composePath)
	}
	if !hasFlag(tail, "--port", "-p") {
		cmdArgs = append(cmdArgs, "--port", strconv.Itoa(composeServerPort))
		port = composeServerPort
	}
	cmdArgs = append(cmdArgs, command)
	cmdArgs = append(cmdArgs, tail...)

	cmd := exec.CommandContext(ctx, composeBinPath, cmdArgs...)
	if appRoot != "" {
		cmd.Dir = appRoot
	}
	stdout := io.Writer(os.Stdout)
	stderr := io.Writer(os.Stderr)
	var stdoutBuf bytes.Buffer
	var stderrBuf bytes.Buffer
	if command != "up" {
		stdout = io.MultiWriter(os.Stdout, &stdoutBuf)
		stderr = io.MultiWriter(os.Stderr, &stderrBuf)
	}
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	cmd.Stdin = os.Stdin
	cmd.Env = composeCommandEnv(appRoot, port)

	if err := cmd.Run(); err != nil {
		if command != "up" {
			stderrStr := strings.ToLower(stderrBuf.String())
			if strings.Contains(stderrStr, "connection refused") || strings.Contains(stderrStr, "no such file or directory") || strings.Contains(stderrStr, "cannot assign requested address") {
				return ErrComposeUnavailable
			}
		}
		return err
	}
	return nil
}

// ErrComposeUnavailable indicates the Process Compose supervisor is not reachable.
var ErrComposeUnavailable = errors.New("process compose unavailable")

func buildComposeDefinition(root string) (map[string]any, error) {
	if err := EnsureServiceBinaries(root); err != nil {
		return nil, err
	}

	natsSpec, err := natssvc.LoadSpec()
	if err != nil {
		return nil, fmt.Errorf("nats spec: %w", err)
	}
	natsPaths, err := natsSpec.EnsureBinaries()
	if err != nil {
		return nil, fmt.Errorf("nats ensure binaries: %w", err)
	}

	pbSpec, err := pocketbasesvc.LoadSpec()
	if err != nil {
		return nil, fmt.Errorf("pocketbase spec: %w", err)
	}
	pbPaths, err := pbSpec.EnsureBinaries()
	if err != nil {
		return nil, fmt.Errorf("pocketbase ensure binaries: %w", err)
	}

	caddyCfg, err := caddyservice.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("caddy config: %w", err)
	}
	caddyPaths, err := caddyCfg.EnsureBinaries()
	if err != nil {
		return nil, fmt.Errorf("caddy ensure binaries: %w", err)
	}

	processes := map[string]any{}

	natsEnv := natsSpec.ResolveEnv(natsPaths)
	natsArgs := relativeArgs(root, natsSpec.ResolveArgs(natsPaths))
	natsEntry := composeProcessEntry(root, relativeCommand(root, natsPaths["nats"]), natsArgs, natsEnv, natsSpec.ComposeOverrides())
	processes["nats"] = natsEntry

	pbEnv := pbSpec.ResolveEnv(pbPaths)
	pbArgs := relativeArgs(root, pbSpec.ResolveArgs(pbPaths))
	pbEntry := composeProcessEntry(root, relativeCommand(root, pbPaths["pocketbase"]), pbArgs, pbEnv, pbSpec.ComposeOverrides())
	ensureDependsOn(pbEntry, map[string]map[string]any{
		"nats": {"condition": "process_healthy"},
	})
	processes["pocketbase"] = pbEntry

	caddyEnv := caddyCfg.Process.Env
	caddyEntry := composeProcessEntry(root, relativeCommand(root, caddyPaths["caddy"]), nil, caddyEnv, caddyCfg.ComposeOverrides())
	ensureDependsOn(caddyEntry, map[string]map[string]any{
		"pocketbase": {"condition": "process_healthy"},
	})
	processes["caddy"] = caddyEntry

	return map[string]any{
		"version":   "0.5",
		"processes": processes,
	}, nil
}

func relativeCommand(root, binaryPath string) string {
	if binaryPath == "" {
		return binaryPath
	}
	if root == "" {
		return binaryPath
	}
	rel, err := filepath.Rel(root, binaryPath)
	if err != nil {
		return binaryPath
	}
	if strings.HasPrefix(rel, "..") {
		return binaryPath
	}
	if runtime.GOOS == "windows" {
		rel = strings.ReplaceAll(rel, "\\", "/")
	}
	if !strings.HasPrefix(rel, ".") && !strings.HasPrefix(rel, "/") {
		rel = "./" + rel
	}
	return rel
}

func envMapToSlice(m map[string]string) []string {
	if len(m) == 0 {
		return nil
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	result := make([]string, 0, len(keys))
	for _, k := range keys {
		result = append(result, fmt.Sprintf("%s=%s", k, m[k]))
	}
	return result
}

func composeProcessEntry(root, command string, args []string, env map[string]string, overrides map[string]any) map[string]any {
	entry := map[string]any{
		"command":     command,
		"working_dir": root,
	}
	if len(args) > 0 {
		entry["args"] = args
	}
	if len(env) > 0 {
		entry["environment"] = envMapToSlice(env)
	}
	entry["availability"] = map[string]any{"restart": "always"}
	if len(overrides) > 0 {
		mergeComposeOverrides(entry, overrides)
	}
	ensureRestartPolicy(entry)
	return entry
}

func ensureDependsOn(entry map[string]any, defaults map[string]map[string]any) {
	if len(defaults) == 0 {
		return
	}
	if _, ok := entry["depends_on"]; ok {
		return
	}
	depends := make(map[string]any, len(defaults))
	for name, cfg := range defaults {
		depends[name] = cloneValue(cfg)
	}
	entry["depends_on"] = depends
}

func ensureRestartPolicy(entry map[string]any) {
	val, ok := entry["availability"]
	if !ok {
		entry["availability"] = map[string]any{"restart": "always"}
		return
	}
	availability, ok := val.(map[string]any)
	if !ok {
		entry["availability"] = map[string]any{"restart": "always"}
		return
	}
	if _, ok := availability["restart"]; !ok {
		availability = cloneValue(availability).(map[string]any)
		availability["restart"] = "always"
		entry["availability"] = availability
	}
}

func mergeComposeOverrides(dst map[string]any, overrides map[string]any) {
	for k, v := range overrides {
		dst[k] = cloneValue(v)
	}
}

func cloneValue(v any) any {
	switch val := v.(type) {
	case map[string]any:
		dup := make(map[string]any, len(val))
		for k, sub := range val {
			dup[k] = cloneValue(sub)
		}
		return dup
	case []any:
		dup := make([]any, len(val))
		for i, sub := range val {
			dup[i] = cloneValue(sub)
		}
		return dup
	case []string:
		dup := make([]string, len(val))
		copy(dup, val)
		return dup
	default:
		return val
	}
}

func relativeArgs(root string, args []string) []string {
	if len(args) == 0 {
		return nil
	}
	result := make([]string, len(args))
	for i, arg := range args {
		result[i] = makeRelative(root, arg)
	}
	return result
}

func makeRelative(root, path string) string {
	if path == "" || root == "" {
		return path
	}
	rel, err := filepath.Rel(root, path)
	if err != nil || strings.HasPrefix(rel, "..") {
		return path
	}
	if runtime.GOOS == "windows" {
		rel = strings.ReplaceAll(rel, "\\", "/")
	}
	if !strings.HasPrefix(rel, ".") && !strings.HasPrefix(rel, "/") {
		rel = "./" + rel
	}
	return rel
}

func composeCommandEnv(appRoot string, port int) []string {
	env := os.Environ()
	root := resolveAppRoot(appRoot)
	env = append(env, fmt.Sprintf("%s=%s", sharedcfg.EnvVarAppRoot, root))
	if port <= 0 {
		port = composeServerPort
	}
	env = append(env, fmt.Sprintf("PC_PORT_NUM=%d", port))
	return env
}

func resolveAppRoot(appRoot string) string {
	root := appRoot
	if root == "" {
		root = runtimecfg.Load().Paths.AppRoot
	}
	if abs, err := filepath.Abs(root); err == nil {
		return abs
	}
	return root
}

func composeCommandNeedsConfig(command string) bool {
	switch command {
	case "", "up", "run":
		return true
	default:
		return false
	}
}

func hasFlag(args []string, names ...string) bool {
	if len(names) == 0 {
		return false
	}
	for _, arg := range args {
		for _, name := range names {
			if arg == name || strings.HasPrefix(arg, name+"=") {
				return true
			}
		}
	}
	return false
}

func resolveComposePort(args []string) int {
	if port, ok := parsePortArg(args); ok {
		return port
	}
	if env := os.Getenv("PC_PORT_NUM"); env != "" {
		if v, err := strconv.Atoi(env); err == nil && v > 0 {
			return v
		}
	}
	return composeServerPort
}

func parsePortArg(args []string) (int, bool) {
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--port" || arg == "-p":
			if i+1 < len(args) {
				if v, err := strconv.Atoi(args[i+1]); err == nil && v > 0 {
					return v, true
				}
			}
		case strings.HasPrefix(arg, "--port="):
			if v, err := strconv.Atoi(arg[len("--port="):]); err == nil && v > 0 {
				return v, true
			}
		case strings.HasPrefix(arg, "-p="):
			if v, err := strconv.Atoi(arg[len("-p="):]); err == nil && v > 0 {
				return v, true
			}
		case strings.HasPrefix(arg, "-p") && len(arg) > 2:
			if v, err := strconv.Atoi(arg[2:]); err == nil && v > 0 {
				return v, true
			}
		}
	}
	return 0, false
}

// ComposePort returns the port Process Compose should use based on the provided arguments
// or environment variables. Falls back to the default composeServerPort.
func ComposePort(args []string) int {
	return resolveComposePort(args)
}

func overrideAppRoot(appRoot string) func() {
	original := os.Getenv(sharedcfg.EnvVarAppRoot)
	root := appRoot
	if strings.TrimSpace(root) == "" {
		root = runtimecfg.Load().Paths.AppRoot
	}
	if strings.TrimSpace(root) == "" {
		return func() {}
	}
	_ = os.Setenv(sharedcfg.EnvVarAppRoot, root)
	return func() {
		if original == "" {
			_ = os.Unsetenv(sharedcfg.EnvVarAppRoot)
		} else {
			_ = os.Setenv(sharedcfg.EnvVarAppRoot, original)
		}
	}
}
