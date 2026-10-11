package searchers

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestReadSearchResponseBodyEnforcesHardLimit(t *testing.T) {
	if _, err := readSearchResponseBody(strings.NewReader(strings.Repeat("x", maxSearchResponseBytes+1))); err == nil {
		t.Fatal("readSearchResponseBody() should reject an oversized response")
	} else if !strings.Contains(err.Error(), "exceeds") {
		t.Fatalf("readSearchResponseBody() error = %v, want a limit error", err)
	}

	data, err := readSearchResponseBody(strings.NewReader("{}"))
	if err != nil {
		t.Fatalf("readSearchResponseBody() returned unexpected error: %v", err)
	}
	if string(data) != "{}" {
		t.Fatalf("readSearchResponseBody() = %q, want %q", data, "{}")
	}
}

func TestDecodeSearchResponseBodyRejectsTrailingData(t *testing.T) {
	var value struct {
		Name string `json:"name"`
	}
	if err := decodeSearchResponseBody(strings.NewReader(`{"name":"ok"} trailing`), &value); err == nil {
		t.Fatal("decodeSearchResponseBody() should reject trailing data")
	}
}

func TestTruncateSearchTextPreservesUTF8(t *testing.T) {
	value := "来源：安全数据"
	for limit := 1; limit < len(value); limit++ {
		got := truncateSearchText(value, limit)
		if len(got) > limit || !utf8.ValidString(got) {
			t.Fatalf("truncateSearchText(limit=%d) = %q, want valid UTF-8 within limit", limit, got)
		}
	}
}

func TestBoundSearchSummarizationInput(t *testing.T) {
	value := strings.Repeat("内容", maxSearchSummarizationInputBytes)
	got := boundSearchSummarizationInput(value)
	if len(got) > maxSearchSummarizationInputBytes {
		t.Fatalf("boundSearchSummarizationInput() returned %d bytes, want at most %d", len(got), maxSearchSummarizationInputBytes)
	}
	if !utf8.ValidString(got) {
		t.Fatal("boundSearchSummarizationInput() returned invalid UTF-8")
	}
}
