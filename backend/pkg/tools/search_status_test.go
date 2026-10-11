package tools

import (
	"testing"

	"pentagi/pkg/config"
)

func TestSearchEnginesStatusEmptyConfig(t *testing.T) {
	status := SearchEnginesStatus(&config.Config{})
	if len(status) == 0 {
		t.Fatal("expected entries for all engines")
	}
	for _, s := range status {
		if s.Available {
			t.Fatalf("engine %s must be unavailable on empty config", s.Name)
		}
		if s.Missing == "" {
			t.Fatalf("engine %s must explain the missing env var", s.Name)
		}
	}
}

func TestSearchEnginesStatusTavilyConfigured(t *testing.T) {
	cfg := &config.Config{TavilyAPIKey: "tvly-test"}
	status := SearchEnginesStatus(cfg)
	for _, s := range status {
		if s.EngineType == "tavily" {
			if !s.Available {
				t.Fatal("tavily must be available with an API key")
			}
			if s.Missing != "" {
				t.Fatalf("available tavily must not report missing: %s", s.Missing)
			}
		}
	}
}

func TestSearchEnginesStatusGooglePartial(t *testing.T) {
	cfg := &config.Config{GoogleAPIKey: "k"}
	status := SearchEnginesStatus(cfg)
	for _, s := range status {
		if s.EngineType == "google" {
			if s.Available {
				t.Fatal("google must be unavailable without the CX key")
			}
			if s.Missing == "" || !contains(s.Missing, "GOOGLE_CX_KEY") {
				t.Fatalf("google missing hint must name GOOGLE_CX_KEY: %q", s.Missing)
			}
		}
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (func() bool {
		for i := 0; i+len(sub) <= len(s); i++ {
			if s[i:i+len(sub)] == sub {
				return true
			}
		}
		return false
	})()
}
