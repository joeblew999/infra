package release

import "testing"

func TestExtractReleaseID(t *testing.T) {
	tests := []struct {
		summary string
		want    string
	}{
		{"deploy release 123 (success)", "123"},
		{"release abc123 completed", "abc123"},
	}
	for _, tt := range tests {
		got := extractReleaseID(tt.summary)
		if got != tt.want {
			t.Fatalf("extractReleaseID(%q) = %q, want %q", tt.summary, got, tt.want)
		}
	}
}
