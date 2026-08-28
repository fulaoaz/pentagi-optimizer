package response

import (
	"regexp"
	"strings"
	"unicode"
)

const zhCNGraphQLFallback = "请求处理失败，请稍后重试"

var zhCNGraphQLErrorMessages = map[string]string{
	"user ID not found":                                     "登录状态中缺少用户 ID",
	"user type not found":                                   "登录状态中缺少用户类型",
	"user permissions not found":                            "登录状态中缺少权限信息",
	"not permitted":                                         "没有执行此操作的权限",
	"model provider is required":                            "请选择模型提供商",
	"user input is required":                                "请输入任务内容",
	"token creation is disabled with default salt":          "当前使用默认签名盐，API 令牌创建功能已停用",
	"invalid TTL: must be between 60 and 94608000 seconds":  "API 令牌有效期必须在 60 至 94608000 秒之间",
	"anonymizer is not available":                           "内容脱敏服务尚未配置",
	"flow not found":                                        "未找到任务流",
	"assistant not found":                                   "未找到智能体",
	"template not found":                                    "未找到模板",
	"template content is empty":                             "模板内容为空",
	"knowledge: embedding provider is not configured":       "知识库尚未配置嵌入模型提供商",
	"knowledge: embedding provider is not available":        "知识库的嵌入模型提供商当前不可用",
	"knowledge: embedder returned no vectors for query":     "嵌入模型未返回查询向量",
	"knowledge: embedder returned no vectors":               "嵌入模型未返回向量",
	"no valid authentication method configured for Bedrock": "尚未为 Bedrock 配置有效的身份验证方式",
	"flow worker is not set":                                "任务流尚未就绪",
	"nothing to load":                                       "没有可加载的内容",
	"task stop timeout":                                     "停止任务超时",
	"assistant stop timeout":                                "停止智能体超时",
	"flow is in 'running' state; patching is not allowed while a task is executing": "任务执行期间不可修改子任务",
}

type graphQLErrorRule struct {
	pattern     *regexp.Regexp
	replacement string
}

var zhCNGraphQLErrorRules = []graphQLErrorRule{
	{regexp.MustCompile(`^requested permission '([^']+)' not found$`), "缺少执行此操作所需的权限（$1）"},
	{regexp.MustCompile(`^resource ([0-9]+) not found$`), "未找到资源 $1"},
	{regexp.MustCompile(`^resource ([0-9]+) not accessible$`), "没有访问资源 $1 的权限"},
	{regexp.MustCompile(`^failed to get provider '([^']+)':.*$`), "获取模型提供商“$1”失败"},
	{regexp.MustCompile(`^token not found:.*$`), "未找到 API 令牌"},
	{regexp.MustCompile(`^invalid token status: (.+)$`), "API 令牌状态无效：$1"},
	{regexp.MustCompile(`^flow not found:.*$`), "未找到任务流"},
	{regexp.MustCompile(`^template not found:.*$`), "未找到模板"},
	{regexp.MustCompile(`^(?:invalid|unsupported) period: (.+)$`), "不支持的统计周期：$1"},
	{regexp.MustCompile(`^missing ([A-Z][A-Z0-9_]+) environment variable$`), "未配置 $1 环境变量"},
	{regexp.MustCompile(`^failed to parse (.+) server URL:.*$`), "$1 服务地址格式无效"},
	{regexp.MustCompile(`^unsupported embedding provider: (.+)$`), "不支持的嵌入模型提供商：$1"},
	{regexp.MustCompile(`^task ID ([0-9]+) not found.*$`), "未找到任务 $1"},
	{regexp.MustCompile(`^assistant ([0-9]+) not found$`), "未找到智能体 $1"},
}

var zhCNGraphQLErrorPrefixes = []struct {
	prefix  string
	message string
}{
	{"unauthorized:", "当前登录状态无权执行此操作"},
	{"failed to generate token ID:", "生成 API 令牌 ID 失败"},
	{"failed to create token in database:", "保存 API 令牌失败"},
	{"failed to create token:", "创建 API 令牌失败"},
	{"failed to update token:", "更新 API 令牌失败"},
	{"failed to delete token:", "删除 API 令牌失败"},
	{"failed to add favorite flow:", "收藏任务流失败"},
	{"failed to delete favorite flow:", "取消收藏任务流失败"},
	{"failed to create template:", "创建模板失败"},
	{"failed to update template:", "更新模板失败"},
	{"failed to delete template:", "删除模板失败"},
	{"failed to get template:", "获取模板失败"},
	{"failed to get templates:", "获取模板列表失败"},
	{"failed to get user providers:", "获取模型提供商列表失败"},
	{"failed to get user preferences:", "获取用户偏好设置失败"},
	{"failed to get tokens:", "获取 API 令牌列表失败"},
	{"failed to list resources:", "获取资源列表失败"},
	{"failed to fetch resources:", "获取资源失败"},
	{"knowledge:", "知识库操作失败，请稍后重试"},
}

// LocalizedGraphQLErrorMessage localizes GraphQL errors that can reach the UI.
// English and already localized messages are preserved; unknown internal
// errors use a stable Chinese fallback instead of exposing implementation text.
func LocalizedGraphQLErrorMessage(acceptLanguage, message string) string {
	if preferredResponseLanguage(acceptLanguage) != responseLanguageChinese || strings.TrimSpace(message) == "" {
		return message
	}

	if localized, ok := zhCNGraphQLErrorMessages[message]; ok {
		return localized
	}

	for _, rule := range zhCNGraphQLErrorRules {
		if rule.pattern.MatchString(message) {
			return rule.pattern.ReplaceAllString(message, rule.replacement)
		}
	}

	for _, entry := range zhCNGraphQLErrorPrefixes {
		if strings.HasPrefix(message, entry.prefix) {
			return entry.message
		}
	}

	if containsHan(message) {
		return message
	}

	return zhCNGraphQLFallback
}

func containsHan(value string) bool {
	for _, char := range value {
		if unicode.Is(unicode.Han, char) {
			return true
		}
	}
	return false
}
