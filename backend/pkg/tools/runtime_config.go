package tools

import (
	"net/url"
	"os"
	"reflect"
	"strconv"
	"strings"

	"pentagi/pkg/config"
)

type RuntimeConfigEntry struct {
	Key             string
	Category        string
	Value           string
	DefaultValue    string
	Configured      bool
	Sensitive       bool
	RestartRequired bool
	Description     string
}

func RuntimeConfigCatalog(cfg *config.Config) []RuntimeConfigEntry {
	if cfg == nil {
		return nil
	}
	typ := reflect.TypeOf(*cfg)
	value := reflect.ValueOf(*cfg)
	entries := make([]RuntimeConfigEntry, 0, typ.NumField())
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		key := strings.Split(field.Tag.Get("env"), ",")[0]
		if key == "" || key == "-" {
			continue
		}
		raw := runtimeConfigValue(value.Field(i))
		sensitive := sensitiveConfigKey(key)
		_, explicitlySet := os.LookupEnv(key)
		display := raw
		if sensitive && raw != "" {
			display = "••••••••"
		}
		entries = append(entries, RuntimeConfigEntry{
			Key: key, Category: runtimeConfigCategory(key), Value: display,
			DefaultValue: field.Tag.Get("envDefault"), Configured: explicitlySet || raw != "",
			Sensitive: sensitive, RestartRequired: true,
			Description: runtimeConfigDescription(key, field.Type.Kind()),
		})
	}
	return entries
}

func runtimeConfigValue(value reflect.Value) string {
	if !value.IsValid() {
		return ""
	}
	if value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return ""
		}
		if value.Type() == reflect.TypeOf((*url.URL)(nil)) {
			return value.Interface().(*url.URL).String()
		}
		return runtimeConfigValue(value.Elem())
	}
	switch value.Kind() {
	case reflect.Bool:
		return strconv.FormatBool(value.Bool())
	case reflect.String:
		return value.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(value.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(value.Uint(), 10)
	case reflect.Slice, reflect.Array:
		parts := make([]string, value.Len())
		for i := 0; i < value.Len(); i++ {
			parts[i] = runtimeConfigValue(value.Index(i))
		}
		return strings.Join(parts, ", ")
	default:
		return ""
	}
}

func sensitiveConfigKey(key string) bool {
	upper := strings.ToUpper(key)
	for _, marker := range []string{"KEY", "PASSWORD", "SECRET", "TOKEN", "SALT", "CREDENTIAL", "ACCESS_KEY"} {
		if strings.Contains(upper, marker) {
			return true
		}
	}
	return strings.Contains(upper, "DATABASE_URL") || strings.Contains(upper, "COOKIE_SIGNING")
}

func runtimeConfigCategory(key string) string {
	upper := strings.ToUpper(key)
	switch {
	case strings.HasPrefix(upper, "MCP_"):
		return "MCP 安全桥"
	case strings.HasPrefix(upper, "DOCKER_") || strings.HasPrefix(upper, "TERMINAL_"):
		return "容器与执行环境"
	case strings.HasPrefix(upper, "SERVER_") || strings.HasPrefix(upper, "STATIC_") || upper == "PUBLIC_URL" || upper == "CORS_ORIGINS":
		return "服务与网络"
	case strings.Contains(upper, "TAVILY") || strings.Contains(upper, "FIRECRAWL") || strings.Contains(upper, "SEARXNG") || strings.Contains(upper, "SEARCH") || strings.Contains(upper, "PERPLEXITY") || strings.Contains(upper, "DUCKDUCKGO") || strings.Contains(upper, "TRAVERSAAL") || strings.Contains(upper, "SPLOITUS") || strings.HasPrefix(upper, "GOOGLE_") || strings.HasPrefix(upper, "EPSS") || strings.HasPrefix(upper, "CVSS") || strings.HasPrefix(upper, "CVE") || strings.HasPrefix(upper, "KEV"):
		return "搜索与安全情报"
	case strings.HasPrefix(upper, "GRAPHITI") || strings.HasPrefix(upper, "LANGFUSE") || strings.HasPrefix(upper, "OTEL") || strings.HasPrefix(upper, "PPROF"):
		return "观测与知识"
	case strings.Contains(upper, "AGENT") || strings.Contains(upper, "ASSISTANT") || strings.Contains(upper, "SUMMARIZER") || strings.Contains(upper, "EXECUTION_MONITOR"):
		return "代理行为"
	case strings.HasPrefix(upper, "DATABASE_") || strings.HasPrefix(upper, "DB_"):
		return "数据库"
	default:
		return "核心运行时"
	}
}

func runtimeConfigDescription(key string, kind reflect.Kind) string {
	upper := strings.ToUpper(key)
	if key == "TAVILY_BASE_URL" {
		return "Tavily 兼容网关地址；应用会自动追加 /search。修改后重启服务。"
	}
	if key == "TAVILY_USE_BEARER_AUTH" {
		return "使用 Authorization: Bearer 认证，而不是把令牌写入请求体。"
	}
	if strings.HasPrefix(upper, "MCP_") {
		return "MCP 桥接治理设置，控制访问、审批、限流和工具暴露范围。"
	}
	if sensitiveConfigKey(key) {
		return "敏感凭据。页面只显示是否已配置，不显示原文。"
	}
	if kind == reflect.Bool {
		return "功能开关；修改 .env 后重启容器生效。"
	}
	if strings.Contains(upper, "URL") || strings.Contains(upper, "PATH") {
		return "服务地址或文件路径；修改 .env 后重启容器生效。"
	}
	return "运行时环境变量；修改 .env 后重启容器生效。"
}
