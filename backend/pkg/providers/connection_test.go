package providers

import (
	"encoding/json"
	"strings"
	"testing"

	"pentagi/pkg/config"
	"pentagi/pkg/providers/pconfig"
)

func TestProviderConnectionEncryptionRoundTrip(t *testing.T) {
	cfg := &config.Config{CookieSigningSalt: "test-provider-connection-salt"}
	providerConfig := &pconfig.ProviderConfig{
		Connection: &pconfig.ConnectionConfig{
			BaseURL: "https://models.example.test/v1/",
			APIKey:  "super-secret",
		},
	}

	if err := sealProviderConnection(cfg, providerConfig); err != nil {
		t.Fatalf("sealProviderConnection: %v", err)
	}
	if providerConfig.Connection.APIKey != "" || providerConfig.Connection.EncryptedAPIKey == "" {
		t.Fatalf("plain key was not cleared safely: %+v", providerConfig.Connection)
	}
	raw, err := json.Marshal(providerConfig)
	if err != nil {
		t.Fatalf("marshal config: %v", err)
	}
	if strings.Contains(string(raw), "super-secret") {
		t.Fatal("plain API key was persisted in provider config JSON")
	}
	if err := hydrateProviderConnection(cfg, providerConfig); err != nil {
		t.Fatalf("hydrateProviderConnection: %v", err)
	}
	if providerConfig.Connection.APIKey != "super-secret" {
		t.Fatalf("decrypted key = %q, want original", providerConfig.Connection.APIKey)
	}
}

func TestRuntimeConfigForCustomProviderUsesConnection(t *testing.T) {
	cfg := &config.Config{CookieSigningSalt: "test-salt", LLMServerURL: "https://global.example.test/v1"}
	providerConfig := &pconfig.ProviderConfig{
		Simple: &pconfig.AgentConfig{Model: "my-model"},
		Connection: &pconfig.ConnectionConfig{
			BaseURL:     "https://user.example.test/v1",
			APIKey:      "user-key",
			ModelPrefix: "tenant",
		},
	}
	runtimeCfg, err := runtimeConfigForProvider(cfg, "custom", providerConfig)
	if err != nil {
		t.Fatalf("runtimeConfigForProvider: %v", err)
	}
	if runtimeCfg.LLMServerURL != "https://user.example.test/v1" || runtimeCfg.LLMServerKey != "user-key" || runtimeCfg.LLMServerProvider != "tenant" || runtimeCfg.LLMServerModel != "my-model" {
		t.Fatalf("custom runtime config was not overridden: %+v", runtimeCfg)
	}
}

func TestResolveStoredProviderConnectionUsesSavedKeyForMatchingURL(t *testing.T) {
	cfg := &config.Config{CookieSigningSalt: "test-provider-connection-salt"}
	stored := &pconfig.ProviderConfig{
		Connection: &pconfig.ConnectionConfig{
			BaseURL: "https://models.example.test/v1",
			APIKey:  "saved-key",
		},
	}
	if err := sealProviderConnection(cfg, stored); err != nil {
		t.Fatalf("sealProviderConnection: %v", err)
	}

	resolved, err := ResolveStoredProviderConnection(cfg, stored, &pconfig.ConnectionConfig{
		BaseURL: " https://models.example.test/v1/ ",
	})
	if err != nil {
		t.Fatalf("ResolveStoredProviderConnection: %v", err)
	}
	if resolved.APIKey != "saved-key" {
		t.Fatalf("resolved API key = %q, want saved key", resolved.APIKey)
	}
	if resolved.BaseURL != "https://models.example.test/v1" {
		t.Fatalf("resolved base URL = %q, want normalized URL", resolved.BaseURL)
	}
	if stored.Connection.APIKey != "" {
		t.Fatal("stored connection was unexpectedly hydrated or mutated")
	}
}

func TestResolveStoredProviderConnectionRejectsMismatchedURL(t *testing.T) {
	cfg := &config.Config{CookieSigningSalt: "test-provider-connection-salt"}
	stored := &pconfig.ProviderConfig{
		Connection: &pconfig.ConnectionConfig{
			BaseURL: "https://models.example.test/v1",
			APIKey:  "saved-key",
		},
	}
	if err := sealProviderConnection(cfg, stored); err != nil {
		t.Fatalf("sealProviderConnection: %v", err)
	}

	_, err := ResolveStoredProviderConnection(cfg, stored, &pconfig.ConnectionConfig{
		BaseURL: "https://another.example.test/v1",
	})
	if err == nil || !strings.Contains(err.Error(), "does not match") {
		t.Fatalf("mismatched URL error = %v, want saved URL mismatch", err)
	}
}

func TestResolveStoredProviderConnectionPrefersExplicitKey(t *testing.T) {
	resolved, err := ResolveStoredProviderConnection(&config.Config{}, &pconfig.ProviderConfig{
		Connection: &pconfig.ConnectionConfig{
			BaseURL:         "https://saved.example.test/v1",
			EncryptedAPIKey: "not-used",
		},
	}, &pconfig.ConnectionConfig{
		BaseURL: "https://new.example.test/v1",
		APIKey:  "new-key",
	})
	if err != nil {
		t.Fatalf("ResolveStoredProviderConnection: %v", err)
	}
	if resolved.APIKey != "new-key" {
		t.Fatalf("resolved API key = %q, want explicit key", resolved.APIKey)
	}
}
