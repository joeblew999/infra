package orchestrator

import (
	"context"
	"io"
	"testing"
	"time"

	sharedcfg "github.com/joeblew999/infra/core/pkg/shared/config"
	"github.com/joeblew999/infra/core/tooling/pkg/auth"
	types "github.com/joeblew999/infra/core/tooling/pkg/types"
)

type fakeAuthProvider struct {
	flyCalls        int
	cloudflareCalls int
}

func (f *fakeAuthProvider) EnsureFly(ctx context.Context, profile sharedcfg.ToolingProfile, opts auth.Options) error {
	f.flyCalls++
	return nil
}

func (f *fakeAuthProvider) EnsureCloudflare(ctx context.Context, profile sharedcfg.ToolingProfile, opts auth.Options) error {
	f.cloudflareCalls++
	return nil
}

type fakeDeployer struct {
	calls int
	opts  types.DeployRequest
	res   *types.DeployResult
}

func (f *fakeDeployer) Deploy(ctx context.Context, opts types.DeployRequest) (*types.DeployResult, error) {
	f.calls++
	f.opts = opts
	if f.res != nil {
		return f.res, nil
	}
	return &types.DeployResult{}, nil
}

func TestDeployEmitsProgressAndInvokesDependencies(t *testing.T) {
	ctx := context.Background()
	profile := sharedcfg.ToolingProfile{Name: "test", FlyApp: "my-app"}

	authProvider := &fakeAuthProvider{}
	deployed := &fakeDeployer{
		res: &types.DeployResult{
			ImageReference: "registry.fly.io/my-app:latest",
			ReleaseSummary: "release completed",
			ReleaseID:      "rel-123",
			Elapsed:        2 * time.Second,
			AppName:        "my-app",
			OrgSlug:        "my-org",
		},
	}

	var capturedProfile sharedcfg.ToolingProfile
	var capturedProfileName string
	svc := NewService(
		WithAuthProvider(authProvider),
		WithDeployerFactory(func(p sharedcfg.ToolingProfile, name, repoRoot, coreDir string) Deployer {
			capturedProfile = p
			capturedProfileName = name
			return deployed
		}),
		WithProfileResolver(func(string) (sharedcfg.ToolingProfile, string) {
			return profile, profile.Name
		}),
	)

	var events []ProgressEvent
	emitter := ProgressEmitterFunc(func(evt ProgressEvent) {
		events = append(events, evt)
	})

	baseOpts := DeployOptions{
		ProfileOverride: "custom",
		RepoRoot:        "/repo",
		CoreDir:         "/repo/core",
		DeployRequest: types.DeployRequest{
			AppName: "app-override",
			OrgSlug: "my-org",
			Region:  "syd",
			Repo:    "registry.fly.io/my-app",
			Stdout:  io.Discard,
			Stderr:  io.Discard,
		},
		Emitter: emitter,
	}

	result, err := svc.Deploy(ctx, baseOpts)
	if err != nil {
		t.Fatalf("Deploy() error = %v", err)
	}

	if authProvider.flyCalls != 1 {
		t.Fatalf("expected fly auth 1 call, got %d", authProvider.flyCalls)
	}
	if authProvider.cloudflareCalls != 1 {
		t.Fatalf("expected cloudflare auth 1 call, got %d", authProvider.cloudflareCalls)
	}
	if deployed.calls != 1 {
		t.Fatalf("expected deploy to be invoked once, got %d", deployed.calls)
	}
	if capturedProfile.Name != profile.Name {
		t.Fatalf("expected profile %q, got %q", profile.Name, capturedProfile.Name)
	}
	if capturedProfileName != profile.Name {
		t.Fatalf("expected profile name %q, got %q", profile.Name, capturedProfileName)
	}

	expectedPhases := []ProgressPhase{
		PhaseStarted,
		PhaseFlyAuth,
		PhaseFlyAuthCompleted,
		PhaseCloudflareAuth,
		PhaseCloudflareComplete,
		PhaseDeploying,
		PhaseCloudflareDNS,
		PhaseSucceeded,
	}
	if len(events) != len(expectedPhases) {
		t.Fatalf("expected %d events, got %d", len(expectedPhases), len(events))
	}
	for i, phase := range expectedPhases {
		if events[i].Phase != phase {
			t.Fatalf("event %d phase = %q, want %q", i, events[i].Phase, phase)
		}
	}

	success := events[len(events)-1]
	if success.Details["image"] != deployed.res.ImageReference {
		t.Fatalf("expected success image %q, got %q", deployed.res.ImageReference, success.Details["image"])
	}
	if success.Details["release_id"] != deployed.res.ReleaseID {
		t.Fatalf("expected release id %q, got %q", deployed.res.ReleaseID, success.Details["release_id"])
	}
	if result == nil || result.ReleaseID != deployed.res.ReleaseID {
		t.Fatalf("expected result release id %q, got %#v", deployed.res.ReleaseID, result)
	}
}

func TestLaunchUsesAdapter(t *testing.T) {
	ctx := context.Background()
	profile := sharedcfg.ToolingProfile{Name: "test", FlyApp: "my-app"}

	authProvider := &fakeAuthProvider{}
	deployed := &fakeDeployer{
		res: &types.DeployResult{AppName: "my-app", OrgSlug: "my-org"},
	}

	svc := NewService(
		WithAuthProvider(authProvider),
		WithDeployerFactory(func(p sharedcfg.ToolingProfile, name, repoRoot, coreDir string) Deployer {
			return deployed
		}),
		WithProfileResolver(func(string) (sharedcfg.ToolingProfile, string) {
			return profile, profile.Name
		}),
	)

	adapter, resultCh, errCh := svc.Launch(ctx, DeployOptions{
		ProfileOverride: "test",
		DeployRequest: types.DeployRequest{
			AppName: "my-app",
			OrgSlug: "my-org",
			Region:  "syd",
			Repo:    "registry.fly.io/my-app",
		},
	})
	defer adapter.Close()

	var events []types.ProgressMessage
	done := make(chan struct{})
	go func() {
		for msg := range adapter.Progress {
			events = append(events, msg)
		}
		close(done)
	}()

	select {
	case res := <-resultCh:
		if res == nil {
			t.Fatal("expected result")
		}
	case err := <-errCh:
		t.Fatalf("unexpected error: %v", err)
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for result")
	}

	if err, ok := <-errCh; ok && err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	<-done

	if len(events) == 0 {
		t.Fatal("expected progress messages")
	}
}
