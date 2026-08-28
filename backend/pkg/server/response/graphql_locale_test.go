package response

import "testing"

func TestLocalizedGraphQLErrorMessage(t *testing.T) {
	tests := []struct {
		name           string
		acceptLanguage string
		message        string
		want           string
	}{
		{name: "English is preserved", acceptLanguage: "en-US", message: "model provider is required", want: "model provider is required"},
		{name: "exact message", acceptLanguage: "zh-CN", message: "model provider is required", want: "请选择模型提供商"},
		{name: "identifier is preserved", acceptLanguage: "zh-CN", message: "requested permission 'flows.edit' not found", want: "缺少执行此操作所需的权限（flows.edit）"},
		{name: "resource ID is preserved", acceptLanguage: "zh-CN", message: "resource 42 not accessible", want: "没有访问资源 42 的权限"},
		{name: "provider name is preserved", acceptLanguage: "zh-CN", message: "failed to get provider 'openai-main': database offline", want: "获取模型提供商“openai-main”失败"},
		{name: "environment key is preserved", acceptLanguage: "zh-CN", message: "missing DEEPSEEK_API_KEY environment variable", want: "未配置 DEEPSEEK_API_KEY 环境变量"},
		{name: "known prefix hides internals", acceptLanguage: "zh-CN", message: "failed to create token: crypto failure", want: "创建 API 令牌失败"},
		{name: "unknown internal error uses fallback", acceptLanguage: "zh-CN", message: "sql: connection reset", want: zhCNGraphQLFallback},
		{name: "existing Chinese is preserved", acceptLanguage: "zh-CN", message: "模型配置无效", want: "模型配置无效"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := LocalizedGraphQLErrorMessage(test.acceptLanguage, test.message); got != test.want {
				t.Fatalf("LocalizedGraphQLErrorMessage(%q, %q) = %q, want %q", test.acceptLanguage, test.message, got, test.want)
			}
		})
	}
}

func TestPreferredResponseLanguageExport(t *testing.T) {
	if got := PreferredResponseLanguage("zh-Hans, en;q=0.8"); got != responseLanguageChinese {
		t.Fatalf("PreferredResponseLanguage returned %q", got)
	}
}
