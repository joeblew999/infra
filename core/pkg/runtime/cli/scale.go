package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	controllerspec "github.com/joeblew999/infra/core/controller/pkg/spec"
)

func newScaleCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "scale",
		Short: "Inspect scaling desired state",
	}

	var (
		specFile   string
		controller string
	)
	show := &cobra.Command{
		Use:   "show",
		Short: "Show the controller desired state specification",
		RunE: func(cmd *cobra.Command, args []string) error {
			var (
				state  controllerspec.DesiredState
				source string
			)

			if controller != "" {
				fetched, err := fetchControllerState(controller)
				if err == nil {
					state = fetched
					source = fmt.Sprintf("controller %s", controller)
				} else {
					fmt.Fprintf(cmd.ErrOrStderr(), "(warn) controller unreachable: %v\n", err)
				}
			}

			if len(state.Services) == 0 {
				if specFile == "" {
					specFile = "controller/spec.yaml"
				}
				fallback, err := controllerspec.LoadFile(specFile)
				if err != nil {
					if os.IsNotExist(err) {
						return fmt.Errorf("spec file %s not found (and controller unavailable)", specFile)
					}
					return err
				}
				state = fallback
				source = specFile
			}

			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "Spec source: %s\n", source)
			for _, svc := range state.Services {
				fmt.Fprintf(out, "\nService: %s\n", svc.ID)
				if svc.DisplayName != "" {
					fmt.Fprintf(out, "  Name: %s\n", svc.DisplayName)
				}
				fmt.Fprintf(out, "  Strategy: %s (autoscale: %s)\n", svc.Scale.Strategy, valueOrDefault(svc.Scale.Autoscale, "manual"))
				for _, region := range svc.Scale.Regions {
					fmt.Fprintf(out, "    - %s min=%d desired=%d max=%d\n", region.Name, region.Min, region.Desired, region.Max)
				}
				if svc.Storage.Provider != "" {
					fmt.Fprintf(out, "  Storage: %s\n", svc.Storage.Provider)
					if svc.Storage.Buckets.Litestream != "" {
						fmt.Fprintf(out, "    Litestream bucket: %s\n", svc.Storage.Buckets.Litestream)
					}
					if svc.Storage.Buckets.Assets != "" {
						fmt.Fprintf(out, "    Assets bucket: %s\n", svc.Storage.Buckets.Assets)
					}
				}
				if svc.Routing.Provider != "" {
					fmt.Fprintf(out, "  Routing: %s zone=%s\n", svc.Routing.Provider, svc.Routing.Zone)
				}
			}
			return nil
		},
	}
	show.Flags().StringVar(&specFile, "file", "controller/spec.yaml", "path to desired state specification (fallback)")
	show.Flags().StringVar(&controller, "controller", os.Getenv("CONTROLLER_ADDR"), "controller API address (e.g. http://127.0.0.1:4400)")

	cmd.AddCommand(show)

	set := &cobra.Command{
		Use:   "set",
		Short: "Apply a service scale specification via the controller",
		RunE: func(cmd *cobra.Command, args []string) error {
			if controller == "" {
				controller = os.Getenv("CONTROLLER_ADDR")
			}
			if controller == "" {
				return fmt.Errorf("controller address required (set --controller or CONTROLLER_ADDR)")
			}
			if !startsWithHTTP(controller) {
				controller = "http://" + controller
			}
			if specFile == "" {
				return fmt.Errorf("service spec file required")
			}
			svc, err := controllerspec.LoadServiceFile(specFile)
			if err != nil {
				return err
			}
			if err := postServiceUpdate(controller, svc); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "updated service %s via %s\n", svc.ID, controller)
			return nil
		},
	}
	set.Flags().StringVar(&specFile, "file", "", "service spec file to apply")
	set.Flags().StringVar(&controller, "controller", os.Getenv("CONTROLLER_ADDR"), "controller API address (e.g. http://127.0.0.1:4400)")

	cmd.AddCommand(set)
	return cmd
}

func valueOrDefault(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func fetchControllerState(addr string) (controllerspec.DesiredState, error) {
	if addr == "" {
		return controllerspec.DesiredState{}, errors.New("controller address empty")
	}
	client := &http.Client{Timeout: 3 * time.Second}
	url := addr
	if !startsWithHTTP(url) {
		url = "http://" + url
	}
	resp, err := client.Get(url + "/v1/services")
	if err != nil {
		return controllerspec.DesiredState{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return controllerspec.DesiredState{}, fmt.Errorf("controller returned %s", resp.Status)
	}
	var payload struct {
		Services []controllerspec.Service `json:"services"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return controllerspec.DesiredState{}, err
	}
	return controllerspec.DesiredState{Services: payload.Services}, nil
}

func startsWithHTTP(addr string) bool {
	return len(addr) >= 7 && (addr[:7] == "http://" || (len(addr) >= 8 && addr[:8] == "https://"))
}

func postServiceUpdate(controller string, svc controllerspec.Service) error {
	payload := struct {
		Service controllerspec.Service `json:"service"`
	}{Service: svc}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPatch, controller+"/v1/services/update", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		data, _ := io.ReadAll(resp.Body)
		if len(data) == 0 {
			data = []byte(resp.Status)
		}
		return fmt.Errorf("controller error: %s", strings.TrimSpace(string(data)))
	}
	return nil
}
