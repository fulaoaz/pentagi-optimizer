package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestUpdateRuntimeConfigPersistsAndReloadsTypedValues(t *testing.T) {
	dataDir := t.TempDir()
	cfg := &Config{DataDir: dataDir, ServerPort: 8080, MCPEnabled: true}

	if err := UpdateRuntimeConfig(cfg, map[string]string{
		"SERVER_PORT":  "9090",
		"MCP_ENABLED":  "false",
		"CORS_ORIGINS": "https://one.example, https://two.example",
	}); err != nil {
		t.Fatalf("UpdateRuntimeConfig() error = %v", err)
	}
	if cfg.ServerPort != 9090 || cfg.MCPEnabled || len(cfg.CorsOrigins) != 2 {
		t.Fatalf("live configuration was not updated: %+v", cfg)
	}

	contents, err := os.ReadFile(filepath.Join(dataDir, runtimeConfigFileName))
	if err != nil {
		t.Fatalf("read persisted overrides: %v", err)
	}
	var saved map[string]string
	if err := json.Unmarshal(contents, &saved); err != nil {
		t.Fatalf("parse persisted overrides: %v", err)
	}
	if saved["SERVER_PORT"] != "9090" || saved["MCP_ENABLED"] != "false" {
		t.Fatalf("unexpected persisted values: %#v", saved)
	}

	reloaded := &Config{DataDir: dataDir}
	if err := LoadRuntimeConfigOverrides(reloaded); err != nil {
		t.Fatalf("LoadRuntimeConfigOverrides() error = %v", err)
	}
	if reloaded.ServerPort != 9090 || reloaded.MCPEnabled || len(reloaded.CorsOrigins) != 2 {
		t.Fatalf("reloaded configuration was not typed correctly: %+v", reloaded)
	}
}

func TestUpdateRuntimeConfigRejectsInvalidValueWithoutMutation(t *testing.T) {
	cfg := &Config{DataDir: t.TempDir(), ServerPort: 8080}

	if err := UpdateRuntimeConfig(cfg, map[string]string{"SERVER_PORT": "not-a-number"}); err == nil {
		t.Fatal("UpdateRuntimeConfig() unexpectedly accepted an invalid integer")
	}
	if cfg.ServerPort != 8080 {
		t.Fatalf("invalid update mutated live configuration: %d", cfg.ServerPort)
	}
}
