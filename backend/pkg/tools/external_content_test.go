package tools

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestWrapUntrustedToolOutputMarksAndBoundsContent(t *testing.T) {
	value := "标题：\u4f60好\n" + strings.Repeat("x", 512)
	wrapped := wrapUntrustedToolOutput(value, 512)

	if !strings.HasPrefix(wrapped, untrustedContentPrefix) {
		t.Fatalf("wrapped output is missing the trust notice: %q", wrapped)
	}
	if !strings.Contains(wrapped, untrustedContentSuffix) {
		t.Fatalf("wrapped output is missing the closing boundary: %q", wrapped)
	}
	if !strings.Contains(wrapped, untrustedContentTruncatedNotice) {
		t.Fatalf("wrapped output is missing the truncation notice: %q", wrapped)
	}
	if len(wrapped) > 512 {
		t.Fatalf("wrapped output is %d bytes, want at most 512", len(wrapped))
	}
	if !utf8.ValidString(wrapped) {
		t.Fatal("wrapped output is not valid UTF-8")
	}
}

func TestWrapUntrustedToolOutputHandlesInvalidUTF8AndEmptyValues(t *testing.T) {
	wrapped := wrapUntrustedToolOutput("a\x00\xffb", 256)
	if strings.ContainsRune(wrapped, '\x00') {
		t.Fatal("wrapped output retained a NUL byte")
	}
	if !utf8.ValidString(wrapped) {
		t.Fatal("wrapped output is not valid UTF-8")
	}
	if !strings.Contains(wrapped, "a\uFFFDb") {
		t.Fatalf("invalid byte was not normalized: %q", wrapped)
	}
	if got := wrapUntrustedToolOutput("", 256); got != "" {
		t.Fatalf("empty output = %q, want empty string", got)
	}
}

func TestWrapUntrustedToolOutputIsIdempotent(t *testing.T) {
	first := wrapUntrustedToolOutput("content", 512)
	second := wrapUntrustedToolOutput(first, 512)
	if second != first {
		t.Fatalf("second wrapping changed the output: first=%q second=%q", first, second)
	}
}

func TestIsUntrustedToolOutput(t *testing.T) {
	for _, name := range []string{
		BrowserToolName,
		WebSearchToolName,
		SearchInMemoryToolName,
		GraphitiSearchToolName,
		SploitusToolName,
	} {
		if !isUntrustedToolOutput(name) {
			t.Errorf("isUntrustedToolOutput(%q) = false, want true", name)
		}
	}
	for _, name := range []string{TerminalToolName, FileToolName, FinalyToolName} {
		if isUntrustedToolOutput(name) {
			t.Errorf("isUntrustedToolOutput(%q) = true, want false", name)
		}
	}
}

func TestTruncateUTF8DoesNotSplitRune(t *testing.T) {
	value := "前缀-安全数据"
	for limit := 1; limit < len(value); limit++ {
		got, truncated := truncateUTF8(value, limit)
		if !truncated || len(got) > limit || !utf8.ValidString(got) {
			t.Fatalf("truncateUTF8(limit=%d) = %q, truncated=%t", limit, got, truncated)
		}
	}
}
