package cli

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

func newControllerCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "controller",
		Short: "Interact with the core controller service",
	}
	cmd.AddCommand(newControllerWatchCommand())
	return cmd
}

func newControllerWatchCommand() *cobra.Command {
	var controller string
	cmd := &cobra.Command{
		Use:   "watch",
		Short: "Stream desired state events from the controller",
		RunE: func(cmd *cobra.Command, args []string) error {
			if controller == "" {
				controller = os.Getenv("CONTROLLER_ADDR")
			}
			if controller == "" {
				return fmt.Errorf("controller address required (set --controller or CONTROLLER_ADDR)")
			}
			url := controller
			if !startsWithHTTP(url) {
				url = "http://" + url
			}
			base := strings.TrimRight(url, "/")
			req, err := http.NewRequestWithContext(cmd.Context(), http.MethodGet, base+"/v1/events", nil)
			if err != nil {
				return err
			}
			client := &http.Client{Timeout: 0}
			resp, err := client.Do(req)
			if err != nil {
				return err
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				data, _ := io.ReadAll(resp.Body)
				if len(data) == 0 {
					data = []byte(resp.Status)
				}
				return fmt.Errorf("controller returned %s", strings.TrimSpace(string(data)))
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Streaming controller events from %s\n", controller)
			scanner := bufio.NewScanner(resp.Body)
			scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
			lastEvent := time.Now()
			for scanner.Scan() {
				line := scanner.Text()
				if strings.HasPrefix(line, "data:") {
					payload := strings.TrimSpace(line[len("data:"):])
					fmt.Fprintln(cmd.OutOrStdout(), payload)
					lastEvent = time.Now()
				} else if strings.TrimSpace(line) == "" {
					continue
				}
			}
			if err := scanner.Err(); err != nil {
				return fmt.Errorf("controller stream error after %s: %w", time.Since(lastEvent).Truncate(time.Millisecond), err)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&controller, "controller", os.Getenv("CONTROLLER_ADDR"), "controller API address (e.g. http://127.0.0.1:4400)")
	return cmd
}
