package tuitemplates

import (
	"bytes"
	"strings"
	"testing"

	runtimeui "github.com/joeblew999/infra/core/pkg/runtime/ui"
	"github.com/joeblew999/infra/core/pkg/runtime/ui/live"
	"github.com/joeblew999/infra/core/pkg/runtime/ui/render"
)

func TestTuiRecentEventsReflectStoreMutation(t *testing.T) {
	tmpl, err := Parse()
	if err != nil {
		t.Fatalf("parse template: %v", err)
	}

	store := live.NewStore(runtimeui.LoadTestSnapshot())

	vmBefore := render.NewViewModel("Core Runtime", store.Snapshot(), "overview", true)
	var before bytes.Buffer
	if err := tmpl.ExecuteTemplate(&before, "overview", vmBefore); err != nil {
		t.Fatalf("render before: %v", err)
	}

	message := "tui-test-event"
	if strings.Contains(before.String(), message) {
		t.Fatalf("expected %q absent before mutation", message)
	}

	store.AppendEvent(message)

	vmAfter := render.NewViewModel("Core Runtime", store.Snapshot(), "overview", true)
	var after bytes.Buffer
	if err := tmpl.ExecuteTemplate(&after, "overview", vmAfter); err != nil {
		t.Fatalf("render after: %v", err)
	}

	if !strings.Contains(after.String(), message) {
		t.Fatalf("expected rendered overview to include %q after mutation\noutput: %s", message, after.String())
	}
}

func TestSummaryIncludesBuildVersion(t *testing.T) {
	tmpl, err := Parse()
	if err != nil {
		t.Fatalf("parse template: %v", err)
	}

	snapshot := runtimeui.LoadTestSnapshot()
	snapshot.Build.Version = "v9.9.9"

	vm := render.NewViewModel("Core Runtime", snapshot, "overview", false)
	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "summary", vm); err != nil {
		t.Fatalf("render summary: %v", err)
	}

	if !strings.Contains(buf.String(), "v9.9.9") {
		t.Fatalf("expected summary to include build version, output: %s", buf.String())
	}
}
