package response

import (
	"strings"

	"golang.org/x/text/language"
)

const (
	responseLanguageEnglish = "en"
	responseLanguageChinese = "zh-CN"
)

var responseLanguageMatcher = language.NewMatcher([]language.Tag{
	language.English,
	language.SimplifiedChinese,
})

var zhCNErrorMessages = map[string]string{
	"Internal":                                        "服务器内部错误",
	"Internal.DBNotFound":                             "数据库不可用",
	"Internal.ServiceNotFound":                        "服务不可用",
	"Internal.DBEncryptorNotFound":                    "数据库加密服务不可用",
	"NotPermitted":                                    "没有执行此操作的权限",
	"AuthRequired":                                    "请先登录",
	"LocalUserRequired":                               "此操作仅限本地账户",
	"PrivilegesRequired":                              "缺少执行此操作所需的权限",
	"AdminRequired":                                   "此操作需要管理员权限",
	"SuperRequired":                                   "此操作需要超级管理员权限",
	"Auth.InvalidLoginRequest":                        "登录信息无效",
	"Auth.InvalidAuthorizeQuery":                      "授权请求参数无效",
	"Auth.InvalidLoginCallbackRequest":                "登录回调信息无效",
	"Auth.InvalidAuthorizationState":                  "授权状态无效",
	"Auth.InvalidSwitchServiceHash":                   "服务切换参数无效",
	"Auth.InvalidAuthorizationNonce":                  "授权随机数无效",
	"Auth.InvalidCredentials":                         "登录名或密码错误",
	"Auth.InvalidUserData":                            "用户数据无效",
	"Auth.InactiveUser":                               "用户账户未启用",
	"Auth.ExchangeTokenFail":                          "交换令牌失败",
	"Auth.TokenExpired":                               "令牌已过期",
	"Auth.VerificationTokenFail":                      "令牌验证失败",
	"Auth.InvalidServiceData":                         "服务数据无效",
	"Auth.InvalidTenantData":                          "租户数据无效",
	"Info.UserNotFound":                               "未找到用户",
	"Info.InvalidUserData":                            "用户数据无效",
	"Info.InvalidServiceData":                         "服务数据无效",
	"Users.NotFound":                                  "未找到用户",
	"Users.InvalidData":                               "用户数据无效",
	"Users.InvalidRequest":                            "用户请求无效",
	"Users.ChangePasswordCurrentUser.InvalidPassword": "密码验证失败",
	"Users.ChangePasswordCurrentUser.InvalidCurrentPassword": "当前密码不正确",
	"Users.ChangePasswordCurrentUser.InvalidNewPassword":     "新密码信息无效",
	"Users.GetUser.ModelsNotFound":                           "未找到用户关联的模型",
	"Users.CreateUser.InvalidUser":                           "用户信息验证失败",
	"Users.PatchUser.ModelsNotFound":                         "未找到用户关联的模型",
	"Users.DeleteUser.ModelsNotFound":                        "未找到用户关联的模型",
	"Roles.InvalidRequest":                                   "角色请求无效",
	"Roles.InvalidData":                                      "角色数据无效",
	"Roles.NotFound":                                         "未找到角色",
	"Prompts.InvalidRequest":                                 "提示词请求无效",
	"Prompts.InvalidData":                                    "提示词数据无效",
	"Prompts.NotFound":                                       "未找到提示词",
	"Screenshots.InvalidRequest":                             "截图请求无效",
	"Screenshots.NotFound":                                   "未找到截图",
	"Screenshots.InvalidData":                                "截图数据无效",
	"Containers.InvalidRequest":                              "容器请求无效",
	"Containers.NotFound":                                    "未找到容器",
	"Containers.InvalidData":                                 "容器数据无效",
	"Agentlogs.InvalidRequest":                               "智能体日志请求无效",
	"Agentlogs.InvalidData":                                  "智能体日志数据无效",
	"Assistantlogs.InvalidRequest":                           "智能体会话日志请求无效",
	"Assistantlogs.InvalidData":                              "智能体会话日志数据无效",
	"Msglogs.InvalidRequest":                                 "消息日志请求无效",
	"Msglogs.InvalidData":                                    "消息日志数据无效",
	"Searchlogs.InvalidRequest":                              "搜索日志请求无效",
	"Searchlogs.InvalidData":                                 "搜索日志数据无效",
	"Termlogs.InvalidRequest":                                "终端日志请求无效",
	"Termlogs.InvalidData":                                   "终端日志数据无效",
	"Vecstorelogs.InvalidRequest":                            "向量库日志请求无效",
	"Vecstorelogs.InvalidData":                               "向量库日志数据无效",
	"Flows.InvalidRequest":                                   "任务流请求无效",
	"Flows.NotFound":                                         "未找到任务流",
	"Flows.InvalidData":                                      "任务流数据无效",
	"FlowFiles.InvalidRequest":                               "任务流文件请求无效",
	"FlowFiles.NotFound":                                     "未找到任务流文件",
	"FlowFiles.InvalidData":                                  "任务流文件数据无效",
	"FlowFiles.AlreadyExists":                                "任务流中已存在同名文件",
	"FlowFiles.ContainerNotRunning":                          "容器未运行",
	"Tasks.InvalidRequest":                                   "任务请求无效",
	"Tasks.NotFound":                                         "未找到任务",
	"Tasks.InvalidData":                                      "任务数据无效",
	"Subtasks.InvalidRequest":                                "子任务请求无效",
	"Subtasks.NotFound":                                      "未找到子任务",
	"Subtasks.InvalidData":                                   "子任务数据无效",
	"Assistants.InvalidRequest":                              "智能体请求无效",
	"Assistants.NotFound":                                    "未找到智能体",
	"Assistants.InvalidData":                                 "智能体数据无效",
	"Resources.InvalidRequest":                               "资源请求无效",
	"Resources.NotFound":                                     "未找到资源",
	"Resources.AlreadyExists":                                "资源已存在",
	"Resources.InvalidData":                                  "资源数据无效",
	"Resources.Conflict":                                     "资源存在冲突，请覆盖或合并后重试",
	"Knowledge.InvalidRequest":                               "知识条目请求无效",
	"Knowledge.NotFound":                                     "未找到知识条目",
	"Knowledge.Unauthorized":                                 "没有管理此知识条目的权限",
	"Knowledge.StoreUnavailable":                             "知识库尚未配置嵌入模型提供商",
	"Knowledge.InvalidData":                                  "知识条目数据无效",
	"Toolcalls.InvalidRequest":                               "工具调用请求无效",
	"Toolcalls.NotFound":                                     "未找到工具调用",
	"Toolcalls.InvalidData":                                  "工具调用数据无效",
	"Anonymize.InvalidRequest":                               "内容脱敏请求无效",
	"Anonymize.Unavailable":                                  "内容脱敏服务尚未配置",
	"Token.CreationDisabled":                                 "默认配置下不能创建 API 令牌",
	"Token.NotFound":                                         "未找到 API 令牌",
	"Token.Unauthorized":                                     "没有管理此 API 令牌的权限",
	"Token.InvalidRequest":                                   "API 令牌请求无效",
	"Token.InvalidData":                                      "API 令牌数据无效",
}

func localizedErrorMessage(acceptLanguage string, err *HttpError) (string, string) {
	responseLanguage := preferredResponseLanguage(acceptLanguage)
	if responseLanguage == responseLanguageChinese {
		if message, ok := zhCNErrorMessages[err.Code()]; ok {
			return message, responseLanguage
		}
	}

	return err.Msg(), responseLanguageEnglish
}

func preferredResponseLanguage(acceptLanguage string) string {
	if strings.TrimSpace(acceptLanguage) == "" {
		return responseLanguageEnglish
	}

	tags, _, err := language.ParseAcceptLanguage(acceptLanguage)
	if err != nil || len(tags) == 0 {
		return responseLanguageEnglish
	}

	_, index, _ := responseLanguageMatcher.Match(tags...)
	if index == 1 {
		return responseLanguageChinese
	}

	return responseLanguageEnglish
}

// PreferredResponseLanguage returns the language tag used for user-facing API
// messages. It intentionally supports only the locales shipped by the UI.
func PreferredResponseLanguage(acceptLanguage string) string {
	return preferredResponseLanguage(acceptLanguage)
}
