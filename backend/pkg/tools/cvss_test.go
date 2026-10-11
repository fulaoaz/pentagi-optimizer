package tools

import (
	"math"
	"strings"
	"testing"
)

// TestScoreCVSS3_KnownVectors checks the v3.1 base equations against vectors
// whose official scores are published in the CVSS v3.1 specification examples
// and the NVD calculator. These cover both Scope values and the full range of
// severity ratings, so a regression in any weight table or in the Roundup
// function shows up here.
func TestScoreCVSS3_KnownVectors(t *testing.T) {
	tests := []struct {
		name   string
		vector string
		want   float64
	}{
		// CVE-2021-44228 (Log4Shell): the canonical 10.0, Scope:Changed.
		{
			name:   "log4shell critical scope changed",
			vector: "CVSS:3.1/AV:N/AC:L/PR:N/UI:N/S:C/C:H/I:H/A:H",
			want:   10.0,
		},
		// Same impact but Scope:Unchanged drops it to 9.8.
		{
			name:   "network full impact scope unchanged",
			vector: "CVSS:3.1/AV:N/AC:L/PR:N/UI:N/S:U/C:H/I:H/A:H",
			want:   9.8,
		},
		{
			name:   "local high privileges low impact",
			vector: "CVSS:3.1/AV:L/AC:H/PR:H/UI:R/S:U/C:L/I:L/A:L",
			want:   3.8,
		},
		{
			name:   "physical access none impact is zero",
			vector: "CVSS:3.1/AV:P/AC:H/PR:H/UI:R/S:U/C:N/I:N/A:N",
			want:   0.0,
		},
		{
			name:   "adjacent network confidentiality only",
			vector: "CVSS:3.1/AV:A/AC:L/PR:N/UI:N/S:U/C:H/I:N/A:N",
			want:   6.5,
		},
		{
			name:   "user interaction required medium",
			vector: "CVSS:3.1/AV:N/AC:L/PR:N/UI:R/S:U/C:H/I:N/A:N",
			want:   6.5,
		},
		{
			name:   "low privileges integrity only",
			vector: "CVSS:3.1/AV:N/AC:L/PR:L/UI:N/S:U/C:N/I:H/A:N",
			want:   6.5,
		},
		{
			name:   "availability only network",
			vector: "CVSS:3.1/AV:N/AC:L/PR:N/UI:N/S:U/C:N/I:N/A:H",
			want:   7.5,
		},
		// v3.0 uses the same base equations as v3.1.
		{
			name:   "v3.0 prefix accepted",
			vector: "CVSS:3.0/AV:N/AC:L/PR:N/UI:N/S:U/C:H/I:H/A:H",
			want:   9.8,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := scoreCVSSVector(tt.vector)
			if got.Err != nil {
				t.Fatalf("scoreCVSSVector(%q) returned error: %v", tt.vector, got.Err)
			}
			if math.Abs(got.Score-tt.want) > 0.001 {
				t.Errorf("scoreCVSSVector(%q) score = %.2f, want %.2f", tt.vector, got.Score, tt.want)
			}
		})
	}
}

// TestScoreCVSS3_ScopeAffectsPrivilegesRequired pins the one weight in the v3.1
// table that is not a constant: PR:L and PR:H weigh more when Scope is Changed.
// Without that dependency both pairs below would score identically.
func TestScoreCVSS3_ScopeAffectsPrivilegesRequired(t *testing.T) {
	unchanged := scoreCVSSVector("CVSS:3.1/AV:N/AC:L/PR:L/UI:N/S:U/C:H/I:H/A:H")
	changed := scoreCVSSVector("CVSS:3.1/AV:N/AC:L/PR:L/UI:N/S:C/C:H/I:H/A:H")

	if unchanged.Err != nil || changed.Err != nil {
		t.Fatalf("unexpected errors: %v / %v", unchanged.Err, changed.Err)
	}
	if changed.Score <= unchanged.Score {
		t.Errorf("Scope:Changed score %.1f should exceed Scope:Unchanged %.1f",
			changed.Score, unchanged.Score)
	}
}

// TestParseCVSSVector_Rejects covers the malformed shapes an LLM is most likely
// to emit, since each one must produce a corrective message rather than a
// silently wrong score.
func TestParseCVSSVector_Rejects(t *testing.T) {
	tests := []struct {
		name   string
		vector string
	}{
		{"empty", ""},
		{"no version prefix", "AV:N/AC:L/PR:N/UI:N/S:U/C:H/I:H/A:H"},
		{"prefix only", "CVSS:3.1"},
		{"malformed metric", "CVSS:3.1/AV/AC:L/PR:N/UI:N/S:U/C:H/I:H/A:H"},
		{"unsupported version", "CVSS:2.0/AV:N/AC:L/Au:N/C:P/I:P/A:P"},
		{"prose instead of vector", "high severity remote code execution"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := scoreCVSSVector(tt.vector); got.Err == nil {
				t.Errorf("scoreCVSSVector(%q) = %.1f, want an error", tt.vector, got.Score)
			}
		})
	}
}

// TestScoreCVSS3_MissingAndInvalidMetrics ensures an incomplete or out-of-range
// base vector is refused instead of being scored from partial data.
func TestScoreCVSS3_MissingAndInvalidMetrics(t *testing.T) {
	tests := []struct {
		name   string
		vector string
	}{
		{"missing availability", "CVSS:3.1/AV:N/AC:L/PR:N/UI:N/S:U/C:H/I:H"},
		{"missing scope", "CVSS:3.1/AV:N/AC:L/PR:N/UI:N/C:H/I:H/A:H"},
		{"invalid attack vector", "CVSS:3.1/AV:X/AC:L/PR:N/UI:N/S:U/C:H/I:H/A:H"},
		{"invalid scope", "CVSS:3.1/AV:N/AC:L/PR:N/UI:N/S:X/C:H/I:H/A:H"},
		{"invalid privileges", "CVSS:3.1/AV:N/AC:L/PR:X/UI:N/S:U/C:H/I:H/A:H"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := scoreCVSSVector(tt.vector); got.Err == nil {
				t.Errorf("scoreCVSSVector(%q) = %.1f, want an error", tt.vector, got.Score)
			}
		})
	}
}

// TestScoreCVSS4_Bounds checks the v4.0 estimate stays inside the rating scale
// and orders findings sensibly. Exact values are deliberately not asserted: the
// official v4.0 score comes from a published macrovector lookup table that
// scoreCVSS4 approximates (see its doc comment).
func TestScoreCVSS4_Bounds(t *testing.T) {
	worst := scoreCVSSVector("CVSS:4.0/AV:N/AC:L/AT:N/PR:N/UI:N/VC:H/VI:H/VA:H/SC:H/SI:H/SA:H")
	if worst.Err != nil {
		t.Fatalf("worst-case vector failed: %v", worst.Err)
	}
	if worst.Score < 9.0 || worst.Score > 10.0 {
		t.Errorf("worst-case v4.0 score = %.1f, want between 9.0 and 10.0", worst.Score)
	}

	none := scoreCVSSVector("CVSS:4.0/AV:P/AC:H/AT:P/PR:H/UI:A/VC:N/VI:N/VA:N/SC:N/SI:N/SA:N")
	if none.Err != nil {
		t.Fatalf("no-impact vector failed: %v", none.Err)
	}
	if none.Score != 0.0 {
		t.Errorf("no-impact v4.0 score = %.1f, want 0.0", none.Score)
	}

	// A less exploitable variant of the worst case must not outrank it.
	harder := scoreCVSSVector("CVSS:4.0/AV:L/AC:H/AT:P/PR:H/UI:A/VC:H/VI:H/VA:H/SC:H/SI:H/SA:H")
	if harder.Err != nil {
		t.Fatalf("hard-to-exploit vector failed: %v", harder.Err)
	}
	if harder.Score >= worst.Score {
		t.Errorf("hard-to-exploit score %.1f should be below worst-case %.1f",
			harder.Score, worst.Score)
	}
}

// TestScoreCVSS4_MissingMetrics confirms v4.0 requires its full base metric set
// rather than falling back to v3.1-shaped input.
func TestScoreCVSS4_MissingMetrics(t *testing.T) {
	// A v3.1 metric set carrying a 4.0 prefix lacks AT and the VC/SC families.
	if got := scoreCVSSVector("CVSS:4.0/AV:N/AC:L/PR:N/UI:N/S:U/C:H/I:H/A:H"); got.Err == nil {
		t.Errorf("v3.1 metrics under a 4.0 prefix scored %.1f, want an error", got.Score)
	}
	if got := scoreCVSSVector("CVSS:4.0/AV:N/AC:L/AT:X/PR:N/UI:N/VC:H/VI:H/VA:H/SC:N/SI:N/SA:N"); got.Err == nil {
		t.Errorf("invalid AT value scored %.1f, want an error", got.Score)
	}
}

// TestRoundUpCVSS covers the spec's Roundup function, including the values
// where a naive math.Ceil(x*10)/10 gives the wrong answer because of float
// representation.
func TestRoundUpCVSS(t *testing.T) {
	tests := []struct {
		in   float64
		want float64
	}{
		{0.0, 0.0},
		{4.02, 4.1},
		{4.0, 4.0},
		{5.55, 5.6},
		{9.99, 10.0},
		// 6.1 is representable slightly below its decimal value; Roundup must
		// return 6.1 rather than 6.2.
		{6.1, 6.1},
	}

	for _, tt := range tests {
		if got := roundUpCVSS(tt.in); math.Abs(got-tt.want) > 0.001 {
			t.Errorf("roundUpCVSS(%v) = %v, want %v", tt.in, got, tt.want)
		}
	}
}

// TestCVSSSeverityLabel pins the rating-scale boundaries, which are inclusive
// at the lower bound of each band.
func TestCVSSSeverityLabel(t *testing.T) {
	tests := []struct {
		score float64
		want  string
	}{
		{0.0, "None"},
		{0.1, "Low"},
		{3.9, "Low"},
		{4.0, "Medium"},
		{6.9, "Medium"},
		{7.0, "High"},
		{8.9, "High"},
		{9.0, "Critical"},
		{10.0, "Critical"},
	}

	for _, tt := range tests {
		if got := cvssSeverityLabel(tt.score); got != tt.want {
			t.Errorf("cvssSeverityLabel(%.1f) = %q, want %q", tt.score, got, tt.want)
		}
	}
}

// TestFormatCVSSResults_MixedInput verifies the report renders scored vectors in
// a table while still surfacing the ones that failed, so a single bad vector in
// a batch does not hide the good results.
func TestFormatCVSSResults_MixedInput(t *testing.T) {
	out := formatCVSSResults([]string{
		"CVSS:3.1/AV:N/AC:L/PR:N/UI:N/S:U/C:H/I:H/A:H",
		"not a vector",
	})

	if !strings.Contains(out, "9.8") {
		t.Errorf("expected the valid vector's score 9.8 in output, got:\n%s", out)
	}
	if !strings.Contains(out, "Critical") {
		t.Errorf("expected severity Critical in output, got:\n%s", out)
	}
	if !strings.Contains(out, "Could not be scored") {
		t.Errorf("expected a failure section for the invalid vector, got:\n%s", out)
	}
}

// TestFormatCVSSResults_AllInvalid checks that a fully invalid batch returns a
// corrective explanation with the expected format instead of an empty table.
func TestFormatCVSSResults_AllInvalid(t *testing.T) {
	out := formatCVSSResults([]string{"garbage", "CVSS:9.9/AV:N"})

	if strings.Contains(out, "| Vector | Version |") {
		t.Errorf("expected no results table when nothing scored, got:\n%s", out)
	}
	if !strings.Contains(out, "No vector could be scored") {
		t.Errorf("expected an explanation that nothing scored, got:\n%s", out)
	}
	if !strings.Contains(out, "CVSS:3.1/") {
		t.Errorf("expected the expected-format hint in output, got:\n%s", out)
	}
}
