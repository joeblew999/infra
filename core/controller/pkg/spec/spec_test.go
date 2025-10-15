package spec

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeTempFile(t *testing.T, dir, name, contents string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	return path
}

func TestLoadServiceFileVariants(t *testing.T) {
	dir := t.TempDir()

	cases := []struct {
		name     string
		contents string
	}{
		{
			name: "root",
			contents: "id: pocketbase\n" +
				"scale:\n" +
				"  strategy: infra\n" +
				"  regions:\n" +
				"    - name: iad\n" +
				"      min: 1\n" +
				"      desired: 2\n" +
				"      max: 3\n",
		},
		{
			name: "service_wrapper",
			contents: "service:\n" +
				"  id: pocketbase\n" +
				"  scale:\n" +
				"    strategy: infra\n" +
				"    regions:\n" +
				"      - name: iad\n" +
				"        min: 1\n" +
				"        desired: 2\n" +
				"        max: 3\n",
		},
		{
			name: "services_wrapper",
			contents: "services:\n" +
				"  - id: pocketbase\n" +
				"    scale:\n" +
				"      strategy: infra\n" +
				"      regions:\n" +
				"        - name: iad\n" +
				"          min: 1\n" +
				"          desired: 2\n" +
				"          max: 3\n",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			path := writeTempFile(t, dir, tc.name+".yaml", tc.contents)
			svc, err := LoadServiceFile(path)
			if err != nil {
				t.Fatalf("LoadServiceFile returned error: %v", err)
			}
			if svc.ID != "pocketbase" {
				t.Fatalf("expected service id pocketbase, got %q", svc.ID)
			}
			if len(svc.Scale.Regions) != 1 {
				t.Fatalf("expected one region, got %d", len(svc.Scale.Regions))
			}
		})
	}
}

func TestLoadServiceFileErrors(t *testing.T) {
	dir := t.TempDir()

	multiple := "services:\n" +
		"  - id: pocketbase\n" +
		"    scale:\n" +
		"      strategy: infra\n" +
		"      regions:\n" +
		"        - name: iad\n" +
		"          min: 1\n" +
		"          desired: 2\n" +
		"          max: 3\n" +
		"  - id: worker\n" +
		"    scale:\n" +
		"      strategy: local\n" +
		"      regions:\n" +
		"        - name: iad\n" +
		"          min: 1\n" +
		"          desired: 1\n" +
		"          max: 1\n"
	path := writeTempFile(t, dir, "multiple.yaml", multiple)
	if _, err := LoadServiceFile(path); err == nil {
		t.Fatal("expected error for multiple services, got nil")
	}
}

func TestDesiredStateValidate(t *testing.T) {
	cases := []struct {
		name    string
		state   DesiredState
		wantErr string
	}{
		{
			name: "valid",
			state: DesiredState{Services: []Service{{
				ID: "pocketbase",
				Scale: ScaleSpec{Regions: []RegionScaleSpec{{
					Name: "iad", Min: 1, Desired: 2, Max: 3,
				}}},
			}}},
			wantErr: "",
		},
		{
			name: "missing_id",
			state: DesiredState{Services: []Service{{
				ID:    "",
				Scale: ScaleSpec{Regions: []RegionScaleSpec{{Name: "iad", Min: 1, Desired: 1, Max: 1}}},
			}}},
			wantErr: "service id is required",
		},
		{
			name: "duplicate_id",
			state: DesiredState{Services: []Service{{
				ID: "svc", Scale: ScaleSpec{Regions: []RegionScaleSpec{{Name: "iad", Min: 1, Desired: 1, Max: 1}}},
			}, {
				ID: "svc", Scale: ScaleSpec{Regions: []RegionScaleSpec{{Name: "iad", Min: 1, Desired: 1, Max: 1}}},
			}}},
			wantErr: "service id svc defined multiple times",
		},
		{
			name: "missing_region",
			state: DesiredState{Services: []Service{{
				ID: "svc",
			}}},
			wantErr: "service svc must define at least one region",
		},
		{
			name: "bad_counts",
			state: DesiredState{Services: []Service{{
				ID:    "svc",
				Scale: ScaleSpec{Regions: []RegionScaleSpec{{Name: "iad", Min: 2, Desired: 1, Max: 3}}},
			}}},
			wantErr: "min > desired",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.state.Validate()
			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tc.wantErr)
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("expected error containing %q, got %q", tc.wantErr, err.Error())
			}
		})
	}
}
