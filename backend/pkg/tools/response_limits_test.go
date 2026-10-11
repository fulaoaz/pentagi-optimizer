package tools

import (
	"strings"
	"testing"
)

func TestDecodeBoundedJSONEnforcesLimit(t *testing.T) {
	var value map[string]string
	if err := decodeBoundedJSON(strings.NewReader(strings.Repeat("x", 128)), &value, 64); err == nil {
		t.Fatal("decodeBoundedJSON() should reject an oversized response")
	} else if !strings.Contains(err.Error(), "exceeds") {
		t.Fatalf("decodeBoundedJSON() error = %v, want a limit error", err)
	}
}

func TestDecodeBoundedJSONRejectsTrailingData(t *testing.T) {
	var value struct {
		Status string `json:"status"`
	}
	if err := decodeBoundedJSON(strings.NewReader(`{"status":"ok"} trailing`), &value, 128); err == nil {
		t.Fatal("decodeBoundedJSON() should reject trailing data")
	}
}
