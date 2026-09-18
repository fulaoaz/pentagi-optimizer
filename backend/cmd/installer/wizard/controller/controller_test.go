package controller

import (
	"os"
	"path/filepath"
	"testing"

	"pentagi/cmd/installer/checker"
	installerfiles "pentagi/cmd/installer/files"
	installerstate "pentagi/cmd/installer/state"
)

func newControllerForServerSettingsTest(t *testing.T, envContent string) Controller {
	t.Helper()

	envPath := filepath.Join(t.TempDir(), ".env")
	if err := os.WriteFile(envPath, []byte(envContent), 0600); err != nil {
		t.Fatalf("write env file: %v", err)
	}

	appState, err := installerstate.NewState(envPath)
	if err != nil {
		t.Fatalf("create state: %v", err)
	}

	return NewController(appState, installerfiles.NewFiles(), checker.CheckResult{})
}

func TestServerSettingsConfigReadsMCPValues(t *testing.T) {
	c := newControllerForServerSettingsTest(t, "MCP_ENABLED=false\nMCP_SERVER_NAME=PentAGI Test\nMCP_SERVER_VERSION=2.0.0\nMCP_API_KEY=read-key\nMCP_WRITE_API_KEY=write-key\nMCP_ALLOW_ANONYMOUS=true\nMCP_ALLOWED_ORIGINS=https://console.example.com,http://localhost:3000\nMCP_ALLOWED_TOOLS=get_flow_status,list_assistants\nMCP_MAX_REQUEST_BYTES=2097152\nMCP_ENABLE_WRITE_TOOLS=true\nMCP_READ_TOOL_RATE_LIMIT=120\nMCP_WRITE_TOOL_RATE_LIMIT=4\nMCP_APPROVAL_MODE=destructive\n")

	cfg := c.GetServerSettingsConfig()

	if cfg.MCPEnabled.Value != "false" || cfg.MCPEnabled.Default != "true" {
		t.Fatalf("MCP enabled = %#v, want value false and default true", cfg.MCPEnabled)
	}
	if cfg.MCPServerName.Value != "PentAGI Test" || cfg.MCPServerVersion.Value != "2.0.0" {
		t.Fatalf("MCP identity = %q/%q", cfg.MCPServerName.Value, cfg.MCPServerVersion.Value)
	}
	if cfg.MCPAPIKey.Value != "read-key" || cfg.MCPWriteAPIKey.Value != "write-key" {
		t.Fatalf("MCP credentials were not loaded")
	}
	if cfg.MCPAllowAnonymous.Value != "true" || cfg.MCPAllowedOrigins.Value != "https://console.example.com,http://localhost:3000" {
		t.Fatalf("MCP access policy = %q/%q", cfg.MCPAllowAnonymous.Value, cfg.MCPAllowedOrigins.Value)
	}
	if cfg.MCPAllowedTools.Value != "get_flow_status,list_assistants" || cfg.MCPMaxRequestBytes.Value != "2097152" || cfg.MCPEnableWriteTools.Value != "true" {
		t.Fatalf("MCP tool policy = %q/%q/%q", cfg.MCPAllowedTools.Value, cfg.MCPMaxRequestBytes.Value, cfg.MCPEnableWriteTools.Value)
	}
	if cfg.MCPReadToolRateLimit.Value != "120" || cfg.MCPWriteToolRateLimit.Value != "4" || cfg.MCPApprovalMode.Value != "destructive" {
		t.Fatalf("MCP governance = %q/%q/%q", cfg.MCPReadToolRateLimit.Value, cfg.MCPWriteToolRateLimit.Value, cfg.MCPApprovalMode.Value)
	}
}

func TestServerSettingsConfigUpdatesAndResetsMCPValues(t *testing.T) {
	c := newControllerForServerSettingsTest(t, "MCP_ENABLED=true\nMCP_SERVER_NAME=PentAGI\nMCP_SERVER_VERSION=1.0.0\nMCP_API_KEY=old-read\nMCP_WRITE_API_KEY=old-write\nMCP_ALLOW_ANONYMOUS=false\nMCP_ALLOWED_ORIGINS=https://old.example.com\nMCP_ALLOWED_TOOLS=get_flow_status\nMCP_MAX_REQUEST_BYTES=1048576\nMCP_ENABLE_WRITE_TOOLS=false\nMCP_READ_TOOL_RATE_LIMIT=60\nMCP_WRITE_TOOL_RATE_LIMIT=10\nMCP_APPROVAL_MODE=scope\n")

	cfg := c.GetServerSettingsConfig()
	cfg.MCPEnabled.Value = "false"
	cfg.MCPServerName.Value = "Updated PentAGI"
	cfg.MCPServerVersion.Value = "3.0.0"
	cfg.MCPAPIKey.Value = "new-read"
	cfg.MCPWriteAPIKey.Value = "new-write"
	cfg.MCPAllowAnonymous.Value = "true"
	cfg.MCPAllowedOrigins.Value = "https://new.example.com"
	cfg.MCPAllowedTools.Value = "list_assistants"
	cfg.MCPMaxRequestBytes.Value = "4194304"
	cfg.MCPEnableWriteTools.Value = "true"
	cfg.MCPReadToolRateLimit.Value = "240"
	cfg.MCPWriteToolRateLimit.Value = "8"
	cfg.MCPApprovalMode.Value = "write"

	if err := c.UpdateServerSettingsConfig(cfg); err != nil {
		t.Fatalf("update MCP settings: %v", err)
	}

	state := c.GetState()
	updated, _ := state.GetVars([]string{
		"MCP_ENABLED", "MCP_SERVER_NAME", "MCP_SERVER_VERSION", "MCP_API_KEY", "MCP_WRITE_API_KEY",
		"MCP_ALLOW_ANONYMOUS", "MCP_ALLOWED_ORIGINS", "MCP_ALLOWED_TOOLS", "MCP_MAX_REQUEST_BYTES", "MCP_ENABLE_WRITE_TOOLS",
		"MCP_READ_TOOL_RATE_LIMIT", "MCP_WRITE_TOOL_RATE_LIMIT", "MCP_APPROVAL_MODE",
	})
	want := map[string]string{
		"MCP_ENABLED": "false", "MCP_SERVER_NAME": "Updated PentAGI", "MCP_SERVER_VERSION": "3.0.0",
		"MCP_API_KEY": "new-read", "MCP_WRITE_API_KEY": "new-write", "MCP_ALLOW_ANONYMOUS": "true",
		"MCP_ALLOWED_ORIGINS": "https://new.example.com", "MCP_ALLOWED_TOOLS": "list_assistants",
		"MCP_MAX_REQUEST_BYTES": "4194304", "MCP_ENABLE_WRITE_TOOLS": "true",
		"MCP_READ_TOOL_RATE_LIMIT": "240", "MCP_WRITE_TOOL_RATE_LIMIT": "8", "MCP_APPROVAL_MODE": "write",
	}
	for name, expected := range want {
		if updated[name].Value != expected {
			t.Errorf("%s = %q, want %q", name, updated[name].Value, expected)
		}
	}

	if reset := c.ResetServerSettingsConfig(); reset == nil {
		t.Fatal("reset MCP settings returned nil")
	}
	resetVars, _ := state.GetVars([]string{"MCP_ENABLED", "MCP_API_KEY", "MCP_ALLOWED_TOOLS", "MCP_MAX_REQUEST_BYTES", "MCP_READ_TOOL_RATE_LIMIT", "MCP_WRITE_TOOL_RATE_LIMIT", "MCP_APPROVAL_MODE"})
	if resetVars["MCP_ENABLED"].Value != "true" || resetVars["MCP_API_KEY"].Value != "old-read" || resetVars["MCP_ALLOWED_TOOLS"].Value != "get_flow_status" || resetVars["MCP_MAX_REQUEST_BYTES"].Value != "1048576" || resetVars["MCP_READ_TOOL_RATE_LIMIT"].Value != "60" || resetVars["MCP_WRITE_TOOL_RATE_LIMIT"].Value != "10" || resetVars["MCP_APPROVAL_MODE"].Value != "scope" {
		t.Fatalf("reset did not restore MCP env values: %#v", resetVars)
	}
}
