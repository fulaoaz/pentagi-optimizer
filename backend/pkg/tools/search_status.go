package tools

import (
	"pentagi/pkg/config"
)

// SearchEngineStatus describes one search engine of the web_search tool:
// whether it is usable right now and which env vars an operator must set to
// enable it. The GraphQL layer exposes this so the settings UI can render a
// visual configuration overview with actionable explanations.
type SearchEngineStatus struct {
	Name        string `json:"name"`
	EngineType  string `json:"engineType"`
	Available   bool   `json:"available"`
	Description string `json:"description"`
	Missing     string `json:"missing,omitempty"`
}

// SearchEnginesStatus derives per-engine availability purely from cfg so the
// UI shows what is active and what is missing without touching the network.
func SearchEnginesStatus(cfg *config.Config) []SearchEngineStatus {
	if cfg == nil {
		return nil
	}
	return []SearchEngineStatus{
		{Name: "Tavily", EngineType: "tavily", Available: cfg.TavilyAPIKey != "",
			Description: "Analytic search with content extraction; first choice for answers and research. A Tavily-compatible gateway can be used via TAVILY_BASE_URL (PentAGI appends /search) plus TAVILY_USE_BEARER_AUTH=true for Bearer-token auth.",
			Missing:     missingCfg(cfg.TavilyAPIKey == "", "TAVILY_API_KEY")},
		{Name: "DuckDuckGo", EngineType: "duckduckgo", Available: cfg.DuckDuckGoEnabled,
			Description: "Free link discovery, no API key required; can be disabled via DUCKDUCKGO_ENABLED=false.",
			Missing:     missingBool(cfg.DuckDuckGoEnabled, "DUCKDUCKGO_ENABLED")},
		{Name: "Google CSE", EngineType: "google", Available: cfg.GoogleAPIKey != "" && cfg.GoogleCXKey != "",
			Description: "Programmable Search Engine; stable link discovery when both credentials are set.",
			Missing:     joinMissing([]bool{cfg.GoogleAPIKey == "", cfg.GoogleCXKey == ""}, []string{"GOOGLE_API_KEY", "GOOGLE_CX_KEY"})},
		{Name: "Perplexity", EngineType: "perplexity", Available: cfg.PerplexityAPIKey != "",
			Description: "Reasoning-strong research engine with citations; leads the research fallback chain.",
			Missing:     missingCfg(cfg.PerplexityAPIKey == "", "PERPLEXITY_API_KEY")},
		{Name: "SearXNG", EngineType: "searxng", Available: cfg.SearxngURL != "",
			Description: "Self-hosted meta search; point SEARXNG_URL at your instance.",
			Missing:     missingCfg(cfg.SearxngURL == "", "SEARXNG_URL")},
		{Name: "Firecrawl", EngineType: "firecrawl", Available: cfg.FirecrawlAPIKey != "",
			Description: "Deep crawl and extraction; strong for reading specific pages.",
			Missing:     missingCfg(cfg.FirecrawlAPIKey == "", "FIRECRAWL_API_KEY")},
		{Name: "Traversaal", EngineType: "traversaal", Available: cfg.TraversaalAPIKey != "",
			Description: "Alternative answer engine used as a mid-chain fallback.",
			Missing:     missingCfg(cfg.TraversaalAPIKey == "", "TRAVERSAAL_API_KEY")},
		{Name: "Sploitus", EngineType: "sploitus", Available: cfg.SploitusEnabled,
			Description: "Exploit/PoC index; powers the web_search mode=exploit chain. Off by default.",
			Missing:     missingBool(cfg.SploitusEnabled, "SPLOITUS_ENABLED")},
	}
}

func missingCfg(missing bool, envVar string) string {
	if !missing {
		return ""
	}
	return "set " + envVar + " in the deployment environment (.env)"
}

func missingBool(enabled bool, envVar string) string {
	if enabled {
		return ""
	}
	return "set " + envVar + "=true in the deployment environment (.env)"
}

func joinMissing(missing []bool, envVars []string) string {
	var out []string
	for i, m := range missing {
		if m {
			out = append(out, envVars[i])
		}
	}
	if len(out) == 0 {
		return ""
	}
	result := "set "
	for i, v := range out {
		if i > 0 {
			result += " and "
		}
		result += v
	}
	return result + " in the deployment environment (.env)"
}
