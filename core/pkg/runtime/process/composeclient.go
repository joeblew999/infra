package process

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	composeDefaultHost    = "127.0.0.1"
	composeRequestTimeout = 3 * time.Second
)

// ComposeProcessState mirrors the upstream process state payload we care about.
type ComposeProcessState struct {
	Name           string `json:"name"`
	Namespace      string `json:"namespace"`
	Status         string `json:"status"`
	Health         string `json:"is_ready"`
	HasHealthProbe bool   `json:"has_ready_probe"`
	Restarts       int    `json:"restarts"`
	ExitCode       int    `json:"exit_code"`
	IsRunning      bool   `json:"is_running"`
	Replicas       int    `json:"replicas"`
}

// ProjectState represents the raw project state returned by Process Compose.
type ProjectState map[string]any

// ErrComposeProcessNotFound is returned when a process name cannot be resolved.
var ErrComposeProcessNotFound = errors.New("process not found")

// IsComposeRunning checks if the Process Compose server is reachable.
// Returns true if the server responds to health requests, false otherwise.
func IsComposeRunning(ctx context.Context, port int) bool {
	if ctx == nil {
		ctx = context.Background()
	}
	// Use a short timeout context to avoid blocking
	checkCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	// Try to fetch processes - if this succeeds, compose is running
	_, err := FetchComposeProcesses(checkCtx, port)
	return err == nil
}

func composeBaseURL(port int) string {
	if port <= 0 {
		port = composeServerPort
	}
	return fmt.Sprintf("http://%s:%d", composeDefaultHost, port)
}

func newHTTPClient() *http.Client {
	return &http.Client{Timeout: composeRequestTimeout}
}

func isConnErr(err error) bool {
	if err == nil {
		return false
	}
	var netErr net.Error
	if errors.As(err, &netErr) {
		return true
	}
	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "connection refused"),
		strings.Contains(msg, "connect: operation timed out"),
		strings.Contains(msg, "no such file or directory"),
		strings.Contains(msg, "cannot assign requested address"),
		strings.Contains(msg, "connection reset"),
		strings.Contains(msg, "connectex"),
		strings.Contains(msg, "econnrefused"):
		return true
	default:
		return false
	}
}

func FetchComposeProcesses(ctx context.Context, port int) ([]ComposeProcessState, error) {
	url := composeBaseURL(port) + "/processes"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := newHTTPClient().Do(req)
	if err != nil {
		if isConnErr(err) {
			return nil, ErrComposeUnavailable
		}
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("processes request failed: %s", resp.Status)
	}
	var data struct {
		States []ComposeProcessState `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}
	return data.States, nil
}

func FetchComposeProcess(ctx context.Context, port int, name string) (*ComposeProcessState, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("process name is required")
	}
	states, err := FetchComposeProcesses(ctx, port)
	if err != nil {
		return nil, err
	}
	for _, st := range states {
		if composeProcessMatches(st, name) {
			match := st
			return &match, nil
		}
	}
	return nil, ErrComposeProcessNotFound
}

func FetchComposeProcessLogs(ctx context.Context, port int, name string, endOffset, limit int) ([]string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("process name is required")
	}
	if endOffset < 0 {
		endOffset = 0
	}
	if limit < 0 {
		limit = 0
	}
	base := composeBaseURL(port)
	path := fmt.Sprintf("/process/logs/%s/%d/%d", url.PathEscape(name), endOffset, limit)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, base+path, nil)
	if err != nil {
		return nil, err
	}
	resp, err := newHTTPClient().Do(req)
	if err != nil {
		if isConnErr(err) {
			return nil, ErrComposeUnavailable
		}
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, decodeComposeError(resp)
	}
	var payload struct {
		Logs []string `json:"logs"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}
	return payload.Logs, nil
}

func TruncateComposeProcessLogs(ctx context.Context, port int, name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("process name is required")
	}
	path := fmt.Sprintf("/process/logs/%s", url.PathEscape(name))
	resp, err := composeDo(ctx, http.MethodDelete, composeBaseURL(port)+path, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		_, _ = io.Copy(io.Discard, resp.Body)
		return nil
	}
	return decodeComposeError(resp)
}

func ShutdownCompose(ctx context.Context, port int) error {
	url := composeBaseURL(port) + "/project/stop/"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return err
	}
	resp, err := newHTTPClient().Do(req)
	if err != nil {
		if isConnErr(err) {
			return ErrComposeUnavailable
		}
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("project stop failed: %s", resp.Status)
	}
	return nil
}

func StartComposeProcess(ctx context.Context, port int, name string) error {
	url := composeBaseURL(port) + "/process/start/" + name
	resp, err := composeDo(ctx, http.MethodPost, url, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		_, _ = io.Copy(io.Discard, resp.Body)
		return nil
	}
	return decodeComposeError(resp)
}

func StopComposeProcess(ctx context.Context, port int, name string) error {
	url := composeBaseURL(port) + "/process/stop/" + name
	resp, err := composeDo(ctx, http.MethodPatch, url, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		_, _ = io.Copy(io.Discard, resp.Body)
		return nil
	}
	return decodeComposeError(resp)
}

func StopComposeProcesses(ctx context.Context, port int, names []string) (map[string]string, error) {
	url := composeBaseURL(port) + "/processes/stop"
	resp, err := composeDo(ctx, http.MethodPatch, url, names)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusMultiStatus {
		var result map[string]string
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return nil, err
		}
		return result, nil
	}
	return nil, decodeComposeError(resp)
}

func RestartComposeProcess(ctx context.Context, port int, name string) error {
	url := composeBaseURL(port) + "/process/restart/" + name
	resp, err := composeDo(ctx, http.MethodPost, url, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		_, _ = io.Copy(io.Discard, resp.Body)
		return nil
	}
	return decodeComposeError(resp)
}

func ScaleComposeProcess(ctx context.Context, port int, name string, scale int) error {
	url := composeBaseURL(port) + "/process/scale/" + name + fmt.Sprintf("/%d", scale)
	resp, err := composeDo(ctx, http.MethodPatch, url, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		_, _ = io.Copy(io.Discard, resp.Body)
		return nil
	}
	return decodeComposeError(resp)
}

func ReloadComposeProject(ctx context.Context, port int) (map[string]string, error) {
	url := composeBaseURL(port) + "/project/configuration"
	resp, err := composeDo(ctx, http.MethodPost, url, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusMultiStatus {
		var result map[string]string
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return nil, err
		}
		return result, nil
	}
	return nil, decodeComposeError(resp)
}

func UpdateComposeProject(ctx context.Context, port int, payload []byte) (map[string]string, error) {
	url := composeBaseURL(port) + "/project"
	resp, err := composeDo(ctx, http.MethodPost, url, json.RawMessage(payload))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusMultiStatus {
		var result map[string]string
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return nil, err
		}
		return result, nil
	}
	return nil, decodeComposeError(resp)
}

func GetComposeProjectState(ctx context.Context, port int, withMemory bool) (ProjectState, error) {
	url := composeBaseURL(port) + fmt.Sprintf("/project/state/?withMemory=%v", withMemory)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := newHTTPClient().Do(req)
	if err != nil {
		if isConnErr(err) {
			return nil, ErrComposeUnavailable
		}
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, decodeComposeError(resp)
	}
	var state ProjectState
	if err := json.NewDecoder(resp.Body).Decode(&state); err != nil {
		return nil, err
	}
	return state, nil
}

func composeProcessMatches(st ComposeProcessState, lookup string) bool {
	lookup = strings.TrimSpace(lookup)
	if lookup == "" {
		return false
	}
	candidates := []string{st.Name}
	if st.Namespace != "" {
		comboSlash := st.Namespace + "/" + st.Name
		comboDot := st.Namespace + "." + st.Name
		candidates = append(candidates, st.Namespace, comboSlash, comboDot)
	}
	for _, candidate := range candidates {
		if candidate == "" {
			continue
		}
		if strings.EqualFold(candidate, lookup) {
			return true
		}
	}
	return false
}

func composeDo(ctx context.Context, method, url string, payload any) (*http.Response, error) {
	var body io.Reader
	if payload != nil {
		data, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}
		body = bytes.NewReader(data)
	}
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := newHTTPClient().Do(req)
	if err != nil {
		if isConnErr(err) {
			return nil, ErrComposeUnavailable
		}
		return nil, err
	}
	return resp, nil
}

func decodeComposeError(resp *http.Response) error {
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("process-compose: %s", resp.Status)
	}
	var pe struct {
		Error string `json:"error"`
	}
	if len(data) > 0 {
		if json.Unmarshal(data, &pe) == nil && pe.Error != "" {
			return errors.New(pe.Error)
		}
		msg := strings.TrimSpace(string(data))
		if msg != "" {
			return errors.New(msg)
		}
	}
	return fmt.Errorf("process-compose: %s", resp.Status)
}
