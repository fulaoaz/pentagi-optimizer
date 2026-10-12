package artex

import "testing"

func TestNormalizeTaskStatusPreservesTimeoutAndPauseSemantics(t *testing.T) {
	for _, tc := range []struct{ raw, want string }{
		{"queued", "queued"}, {"running", "running"}, {"paused", "paused"},
		{"completed", "completed"}, {"timeout", "timeout"}, {"failed", "failed"}, {"future_state", "unknown"},
	} {
		if got := NormalizeTaskStatus(tc.raw); got != tc.want {
			t.Errorf("%q => %q, want %q", tc.raw, got, tc.want)
		}
	}
}
