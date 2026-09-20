package tools

import (
	"testing"

	"pentagi/pkg/config"
)

func TestRuntimeConfigCatalogMasksSecretsAndGroupsSettings(t *testing.T) {
	entries := RuntimeConfigCatalog(&config.Config{
		MCPAPIKey:     "secret-value",
		TavilyBaseURL: "https://gateway.example.test/api/tavily",
	})
	var mcp, tavily *RuntimeConfigEntry
	for i := range entries {
		switch entries[i].Key {
		case "MCP_API_KEY":
			mcp = &entries[i]
		case "TAVILY_BASE_URL":
			tavily = &entries[i]
		}
	}
	if mcp == nil || !mcp.Sensitive || mcp.Value == "secret-value" || !mcp.Configured {
		t.Fatalf("MCP secret was not safely represented: %+v", mcp)
	}
	if mcp.Category != "MCP 安全桥" {
		t.Fatalf("unexpected MCP category: %q", mcp.Category)
	}
	if tavily == nil || tavily.Value != "https://gateway.example.test/api/tavily" {
		t.Fatalf("Tavily URL was not exposed as non-secret configuration: %+v", tavily)
	}
	if tavily.Category != "搜索与安全情报" || !tavily.RestartRequired {
		t.Fatalf("unexpected Tavily metadata: %+v", tavily)
	}
}

func TestRuntimeConfigCatalogNilConfig(t *testing.T) {
	if got := RuntimeConfigCatalog(nil); got != nil {
		t.Fatalf("expected nil catalog for nil config, got %d entries", len(got))
	}
}
