package process

import "testing"

func TestSequenceCapped(t *testing.T) {
	backoff := Backoff{Initial: 100, Max: 500, Multiplier: 2}
	if d := backoff.Sequence(3); d != 500 {
		t.Fatalf("expected capped backoff got %d", d)
	}
}

func TestShouldRestart(t *testing.T) {
	spec := Spec{RestartPolicy: RestartPolicyOnFailure}
	if !spec.ShouldRestart(assertErr{}) {
		t.Fatal("expected restart on failure")
	}
	spec.RestartPolicy = RestartPolicyNever
	if spec.ShouldRestart(nil) {
		t.Fatal("did not expect restart")
	}
}

type assertErr struct{}

func (assertErr) Error() string { return "fail" }
