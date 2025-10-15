package process

import "time"

// RestartPolicy represents the strategy used for restarting a process.
type RestartPolicy string

const (
	RestartPolicyNever     RestartPolicy = "never"
	RestartPolicyOnFailure RestartPolicy = "on-failure"
	RestartPolicyAlways    RestartPolicy = "always"
)

// Backoff describes retry backoff configuration for restarts.
type Backoff struct {
	Initial    time.Duration
	Max        time.Duration
	Multiplier float64
}

// Sequence computes the backoff duration for a given attempt number (starting
// at zero). When Max is set, the value is capped accordingly.
func (b Backoff) Sequence(attempt int) time.Duration {
	if attempt <= 0 {
		return b.Initial
	}
	delay := float64(b.Initial)
	for i := 0; i < attempt; i++ {
		delay *= b.Multiplier
		if b.Max > 0 && time.Duration(delay) >= b.Max {
			return b.Max
		}
	}
	return time.Duration(delay)
}

// Spec describes how a runtime process should be executed.
type Spec struct {
	Command       string
	Args          []string
	Env           map[string]string
	WorkingDir    string
	RestartPolicy RestartPolicy
	Backoff       Backoff
	Ensure        []EnsureStep
}

// EnsureStep represents a deterministic setup action executed before process
// startup.
type EnsureStep struct {
	Type     string
	Source   string
	Target   string
	Contents string
}

// ShouldRestart returns true when the restart policy demands another attempt.
func (s Spec) ShouldRestart(exitErr error) bool {
	switch s.RestartPolicy {
	case RestartPolicyAlways:
		return true
	case RestartPolicyOnFailure:
		return exitErr != nil
	default:
		return false
	}
}
