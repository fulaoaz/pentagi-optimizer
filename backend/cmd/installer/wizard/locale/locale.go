package locale

// Common status and UI strings
const (
	// Common status and UI strings
	UIStatistics       = "统计信息"
	UIStatus           = "状态："
	UIMode             = "模式："
	UINoConfigSelected = "未选择任何配置"
	UILoading          = "加载中..."
	UINotImplemented   = "功能尚未实现"
	UIUnsavedChanges   = "有未保存的更改"
	UIConfigSaved      = "配置已保存"

	// Status labels
	StatusEnabled         = "已启用"
	StatusDisabled        = "已禁用"
	StatusConfigured      = "已配置"
	StatusNotConfigured   = "未配置"
	StatusEmbedded        = "内置"
	StatusExternal        = "外部"
	StatusEnabledInsecure = "已启用（⚠ 不安全）"
	StatusNotSet          = "未设置"

	// Success/Warning messages
	MessageSearchEnginesNone       = "⚠ 未配置搜索引擎"
	MessageSearchEnginesConfigured = "✓ 已配置 %d 个搜索引擎"
	MessageDockerConfigured        = "✓ Docker 环境已配置"
	MessageDockerNotConfigured     = "⚠ Docker 环境未配置"
)

// Legend constants
const (
	LegendConfigured    = "✓ 已配置"
	LegendNotConfigured = "✗ 未配置"
)

// Common Navigation Actions (always available)
const (
	NavBack       = "Esc: 返回"
	NavExit       = "Ctrl+Q: 退出"
	NavUpDown     = "↑/↓: 滚动/选择"
	NavLeftRight  = "←/→: 移动"
	NavPgUpPgDown = "PgUp/PgDn: 翻页"
	NavHomeEnd    = "Home/End: 首/末"
	NavEnter      = "Enter: 继续"
	NavYn         = "Y/N: 接受/拒绝"
	NavCtrlC      = "Ctrl+C: 取消"
	NavCtrlS      = "Ctrl+S: 保存"
	NavCtrlR      = "Ctrl+R: 重置"
	NavCtrlH      = "Ctrl+H: 显示/隐藏"
	NavTab        = "Tab: 补全"
	NavSeparator  = " • "
)

// Welcome Screen constants
const (
	// Form interface implementation
	WelcomeFormTitle       = "欢迎使用 PentAGI"
	WelcomeFormDescription = "PentAGI 是一个自主渗透测试平台，利用 AI 技术执行全面的安全评估。"
	WelcomeFormName        = "欢迎"
	WelcomeFormOverview    = `系统检查项：
• 环境配置文件是否存在
• Docker API 可访问性与版本兼容性
• Worker 环境就绪状态
• 系统资源（CPU、内存、磁盘空间）
• 外部依赖所需的网络连通性

所有检查通过后，即可进入配置向导，设置 LLM 提供商、监控与安全工具。

安装程序会逐个组件引导你完成配置，并针对不同部署场景给出建议。`

	// Configuration status messages
	WelcomeConfigurationFailed = "⚠ 未通过的检查：%s"
	WelcomeConfigurationPassed = "✓ 所有系统检查均已通过"

	// Workflow steps
	WelcomeWorkflowTitle = "安装流程："
	WelcomeWorkflowStep1 = "1. 接受最终用户许可协议"
	WelcomeWorkflowStep2 = "2. 配置 LLM 提供商（OpenAI、Anthropic 等）"
	WelcomeWorkflowStep3 = "3. 设置集成组件（Langfuse、可观测性）"
	WelcomeWorkflowStep4 = "4. 配置安全选项"
	WelcomeWorkflowStep5 = "5. 部署并启动 PentAGI 服务"
	WelcomeSystemReady   = "✓ 系统就绪 —— 按 Enter 继续"
)

// Troubleshooting on welcome screen constants
const (
	TroubleshootTitle = "系统要求未满足"

	// Environment file issues
	TroubleshootEnvFileTitle = "缺少环境配置文件"
	TroubleshootEnvFileDesc  = "PentAGI 配置需要 .env 文件，但未找到该文件或文件不可读。"
	TroubleshootEnvFileFix   = `解决方法：
1. 在安装目录中将 .env.example 复制为 .env
2. 编辑 .env，至少配置一个 LLM 提供商的 API 密钥
3. 确保文件具有读取权限（chmod 644 .env）

快速修复：
cp .env.example .env && chmod 644 .env`

	// Write permissions
	TroubleshootWritePermTitle = "需要写入权限"
	TroubleshootWritePermDesc  = "安装程序需要对配置目录的写入权限，以保存设置并部署服务。"
	TroubleshootWritePermFix   = `解决方法：
1. 检查目录权限：ls -la
2. 授予写入权限：chmod 755 .
3. 或在具有写入权限的位置运行安装程序
4. 确保有足够的磁盘空间`

	// Docker not installed
	TroubleshootDockerNotInstalledTitle = "未安装 Docker"
	TroubleshootDockerNotInstalledDesc  = "本系统未安装 Docker。PentAGI 需要 Docker 来运行容器。"
	TroubleshootDockerNotInstalledFix   = `解决方法：
1. 安装 Docker Desktop：https://docs.docker.com/get-docker/
2. Linux 用户：按发行版对应的说明安装
3. 验证安装：docker --version
4. 确保 docker 命令在 PATH 中`

	// Docker not running
	TroubleshootDockerNotRunningTitle = "Docker 守护进程未运行"
	TroubleshootDockerNotRunningDesc  = "已安装 Docker，但守护进程未运行。Docker 服务必须处于活动状态。"
	TroubleshootDockerNotRunningFix   = `解决方法：
1. 启动 Docker Desktop（Windows/Mac）
2. Linux：sudo systemctl start docker
3. 检查状态：docker ps
4. 若使用 DOCKER_HOST，请确认远程守护进程可访问`

	// Docker permission issues
	TroubleshootDockerPermissionTitle = "Docker 权限被拒绝"
	TroubleshootDockerPermissionDesc  = "当前用户账号没有访问 Docker 的权限。这在 Linux 系统上很常见。"
	TroubleshootDockerPermissionFix   = `解决方法：
1. 将用户加入 docker 组：sudo usermod -aG docker $USER
2. 注销并重新登录使变更生效
3. 或使用 sudo 运行（不建议用于生产环境）
4. 验证：docker ps（应无需 sudo 即可执行）`

	// Generic Docker API issues
	TroubleshootDockerAPITitle = "Docker API 连接失败"
	TroubleshootDockerAPIDesc  = "无法建立与 Docker API 的连接。可能是配置或网络问题导致。"
	TroubleshootDockerAPIFix   = `解决方法：
1. 检查 DOCKER_HOST 环境变量
2. 确认 Docker 正在运行：docker version
3. 远程 Docker：确保网络连通
4. 若使用 TCP 连接，请检查防火墙设置
5. 可尝试：export DOCKER_HOST=unix:///var/run/docker.sock`

	// Docker version issues
	TroubleshootDockerVersionTitle = "Docker 版本过低"
	TroubleshootDockerVersionDesc  = "当前 Docker 版本不兼容。PentAGI 需要 Docker 20.0.0 或更高版本。"
	TroubleshootDockerVersionFix   = `解决方法：
1. 将 Docker 升级到 20.0.0 或更高版本
2. 访问 https://docs.docker.com/engine/install/

当前版本：%s
要求版本：20.0.0+`

	// Docker Compose issues
	TroubleshootComposeTitle = "未找到 Docker Compose"
	TroubleshootComposeDesc  = "需要 Docker Compose v2 插件，但 `docker compose` 不可用。"
	TroubleshootComposeFix   = `解决方法：
1. 安装或升级 Docker Desktop，或为 Docker Engine 安装 Docker Compose v2 插件
2. 确认插件可用：docker compose version
3. 若仅安装了旧版 docker-compose，请同时安装 Docker Compose v2 插件

PentAGI 执行的是 "docker compose"，因此仅有旧版 "docker-compose" 并不足够。
文档：https://docs.docker.com/compose/install/`

	// Docker Compose version issues
	TroubleshootComposeVersionTitle = "Docker Compose 版本过低"
	TroubleshootComposeVersionDesc  = "当前 `docker compose` 版本不兼容。PentAGI 需要 Docker Compose 1.25.0 或更高版本。"
	TroubleshootComposeVersionFix   = `当前版本：%s
要求版本：1.25.0+

解决方法：
1. 将 Docker Desktop 或 Docker Compose v2 插件升级到更高版本
2. 用以下命令验证结果：docker compose version

文档：https://docs.docker.com/compose/install/`

	// Worker environment issues
	TroubleshootWorkerTitle = "无法访问 Worker 的 Docker 环境"
	TroubleshootWorkerDesc  = "无法连接到用于 worker 容器的 Docker 环境。可能是远程或本地 Docker 配置问题。"
	TroubleshootWorkerFix   = `解决方法：
1. 使用远程 Docker 时，在运行安装程序前设置环境变量：
   export DOCKER_HOST=tcp://remote:2376
   export DOCKER_CERT_PATH=/path/to/certs
   export DOCKER_TLS_VERIFY=1
2. 验证连接：docker -H $DOCKER_HOST ps
3. 使用本地 Docker 时，请不要设置这些变量
4. 检查防火墙是否放通 Docker 端口（2375/2376）
5. 若使用 TLS，请确保证书有效`

	// CPU issues
	TroubleshootCPUTitle = "CPU 核心数不足"
	TroubleshootCPUDesc  = "PentAGI 至少需要 2 个 CPU 核心才能正常运行。"
	TroubleshootCPUFix   = `当前系统有 %d 个 CPU 核心，但至少需要 2 个。

虚拟机用户：
1. 在虚拟机设置中增加 CPU 分配
2. 确保宿主机有足够资源

Docker Desktop 用户：
设置 → 资源 → CPUs：设置为 2 或更多`

	// Memory issues
	TroubleshootMemoryTitle = "内存不足"
	TroubleshootMemoryDesc  = "可用内存不足以运行所选组件。"
	TroubleshootMemoryFix   = `内存需求：
• 基础系统：0.5 GB
• PentAGI 核心：+0.5 GB
• Langfuse（如启用）：+1.5 GB
• Observability（如启用）：+1.5 GB

共需：%.1f GB
可用：%.1f GB

解决方法：
1. 关闭不必要的应用程序
2. 提高 Docker 内存上限
3. 禁用可选组件（Langfuse/Observability）`

	// Disk space issues
	TroubleshootDiskTitle = "磁盘空间不足"
	TroubleshootDiskDesc  = "可用磁盘空间不足以完成安装和运行。"
	TroubleshootDiskFix   = `磁盘需求：
• 基础安装：至少 5 GB
• 启用组件后：10 GB + 每个组件 2 GB
• Worker 镜像：25 GB（含 6GB+ 的 Kali 镜像）

需要：%.1f GB
可用：%.1f GB

解决方法：
1. 释放磁盘空间
2. 为 Docker 使用外部存储
3. 清理未使用的 Docker 资源：
   docker system prune -a`

	// Network issues
	TroubleshootNetworkTitle = "网络连接失败"
	TroubleshootNetworkDesc  = "无法访问必需的外部服务，这会导致无法下载 Docker 镜像和更新。"
	TroubleshootNetworkFix   = `失败的检查项：
%s

解决方法：
1. 验证网络连接：ping docker.io
2. 检查 DNS 解析：nslookup docker.io
3. 如使用代理，请在运行安装程序前设置：
   export HTTP_PROXY=http://proxy:port
   export HTTPS_PROXY=http://proxy:port
4. 如需长期使用代理，请写入 .env：
   PROXY_URL=http://proxy:port
5. 检查防火墙是否放通出站 HTTPS（443 端口）
6. 若 DNS 解析失败，可尝试更换 DNS 服务器`

	// Generic hint at the bottom
	TroubleshootFixHint = "\n请解决上述问题后重新运行安装程序。"

	// Network failure messages (used in checker/helpers.go)
	NetworkFailureDNS        = "• docker.io 的 DNS 解析失败"
	NetworkFailureHTTPS      = "• 无法通过 HTTPS 访问外部服务"
	NetworkFailureDockerPull = "• 无法从镜像仓库拉取 Docker 镜像"
)

// System Checks constants
const (
	ChecksTitle               = "系统检查"
	ChecksWarningFailed       = "⚠ 部分检查未通过"
	CheckEnvironmentFile      = "环境配置文件"
	CheckWritePermissions     = "写入权限"
	CheckDockerAPI            = "Docker API"
	CheckDockerVersion        = "Docker 版本"
	CheckDockerCompose        = "Docker Compose"
	CheckDockerComposeVersion = "Docker Compose 版本"
	CheckWorkerEnvironment    = "工作容器环境"
	CheckSystemResources      = "系统资源"
	CheckNetworkConnectivity  = "网络连通性"
)

// EULA Screen constants
const (
	// Form interface implementation
	EULAFormDescription = "PentAGI 使用的法律条款与条件"
	EULAFormName        = "EULA"
	EULAFormOverview    = `请阅读并接受最终用户许可协议（EULA），以继续安装 PentAGI。

EULA 包含以下内容：
• 软件许可条款与使用权利
• 责任限制与担保声明
• 数据收集与隐私政策
• 合规要求与使用限制
• 支持与维护条款

您必须完整浏览整份文档并接受条款，才能继续安装流程。

可使用方向键、PgUp/PgDn 或 Home/End 键浏览文档。`

	// Error and status messages
	EULAErrorLoadingTitle     = "# 加载 EULA 出错\n\n加载 EULA 失败：%v"
	EULAContentFallback       = "# EULA 内容\n\n%s\n\n---\n\n*注意：Markdown 渲染失败：%v*"
	EULAConfigurationRead     = "✓ EULA 已阅读"
	EULAConfigurationAccepted = "✓ EULA 已接受"
	EULAConfigurationPending  = "⚠ EULA 尚未阅读"
	EULALoading               = "正在加载 EULA..."
	EULAProgress              = "进度：%d%%"
	EULAProgressComplete      = " • 已完成"
)

// Main Menu Screen constants
const (
	MainMenuTitle       = "PentAGI 配置"
	MainMenuDescription = "配置 PentAGI 的全部组件与设置"
	MainMenuName        = "主菜单"
	MainMenuOverview    = `欢迎使用 PentAGI 配置中心。

需要配置的核心组件：
• LLM Providers（模型提供商）—— 用于自主测试的 AI 语言模型
• Monitoring（监控）—— 可观测性与分析平台
• Tools（工具）—— 增强测试能力的附加功能
• System Settings（系统设置）—— 环境与部署选项

依次完成各个部分，即可完成 PentAGI 的安装配置。`

	MenuTitle        = "配置菜单"
	MenuSystemStatus = "系统状态"
)

// Main Menu Status Labels (not used)
const (
	MainMenuStatusPentagiRunning     = "PentAGI 已在运行"
	MainMenuStatusPentagiNotRunning  = "可以启动 PentAGI 服务"
	MainMenuStatusUpToDate           = "PentAGI 已是最新版本"
	MainMenuStatusUpdatesAvailable   = "有可用更新"
	MainMenuStatusReadyToStart       = "准备启动"
	MainMenuStatusAllServicesRunning = "所有服务均在运行"
	MainMenuStatusNoUpdatesAvailable = "暂无可用更新"
)

// LLM Providers Screen constants
const (
	LLMProvidersTitle       = "LLM 提供商配置"
	LLMProvidersDescription = "为 AI 智能体配置大语言模型提供商"
	LLMProvidersName        = "LLM 提供商"
	LLMProvidersOverview    = `PentAGI 使用多个专职 AI 智能体（研究员、开发者、执行器、渗透测试员），它们需要不同的 LLM 能力才能取得最佳渗透测试效果。

为什么要配置多个提供商：
• 智能体专职化：不同智能体分别受益于擅长推理、编码或分析的模型
• 成本优化：复杂任务使用高价推理模型（o3、grok-4、claude-sonnet-4、gemini-2.5-pro），简单操作使用低价模型
• 性能优化：各提供商各有所长 — OpenAI 适合中等任务，Anthropic 适合复杂任务，Gemini 更省成本

提供商选择建议：
• 云端生产环境：OpenAI + Anthropic + Gemini，性能与可靠性业界领先
• 企业/合规场景：AWS Bedrock，满足 SOC2、HIPAA 要求，并可访问多个模型家族
• 隐私/本地部署：Ollama 或 vLLM 搭配 Llama 3.1、Qwen3 等开源模型，数据完全自主可控

OpenRouter、DeepInfra、vLLM、Ollama 等提供商的开箱即用配置位于容器内的 /opt/pentagi/conf/ 目录`
)

// LLM Provider titles and descriptions
const (
	LLMProviderOpenAI        = "OpenAI"
	LLMProviderAnthropic     = "Anthropic"
	LLMProviderGemini        = "Google Gemini"
	LLMProviderBedrock       = "AWS Bedrock"
	LLMProviderOllama        = "Ollama"
	LLMProviderDeepSeek      = "DeepSeek"
	LLMProviderGLM           = "GLM Zhipu AI"
	LLMProviderKimi          = "Kimi Moonshot AI"
	LLMProviderQwen          = "Qwen Alibaba Cloud"
	LLMProviderCustom        = "自定义"
	LLMProviderOpenAIDesc    = "业界领先的 GPT 模型，综合表现出色"
	LLMProviderAnthropicDesc = "Claude 模型，推理能力与安全特性更强"
	LLMProviderGeminiDesc    = "Google 的先进多模态模型，知识覆盖面广"
	LLMProviderBedrockDesc   = "通过 AWS 企业级服务访问多家基础模型提供商"
	LLMProviderOllamaDesc    = "本地与云端开源模型，兼顾隐私与灵活性"
	LLMProviderDeepSeekDesc  = "国产先进模型，推理与多语言能力强"
	LLMProviderGLMDesc       = "智谱 AI 的 GLM 模型，擅长中英文任务"
	LLMProviderKimiDesc      = "月之暗面的长上下文模型，适合文档分析"
	LLMProviderQwenDesc      = "阿里云通义千问模型，适合多语言任务"
	LLMProviderCustomDesc    = "自定义 OpenAI 兼容端点，灵活性最高"
)

// Provider-specific help text
const (
	LLMFormOpenAIHelp = `OpenAI 提供业界领先的模型，其前沿推理能力非常适合复杂的渗透测试场景。

PentAGI 默认模型：
• o1、o4-mini：高级推理模型，用于复杂漏洞分析与策略规划
• GPT-4.1、GPT-4.1-mini：旗舰模型，针对漏洞利用开发与代码生成优化
• 根据智能体类型和任务复杂度自动选择模型

主要优势：
• 最先进的推理能力，支持逐步分析（o 系列模型）
• 出色的编码能力，适合自定义漏洞利用开发与载荷生成
• 性能可靠，服务稳定，API 文档完善
• 在安全研究与渗透测试场景中久经验证

适用场景：需要前沿 AI 能力的生产环境，以及性能优先于成本的团队
成本：定价偏高，但经过优化的配置能在成本与质量之间取得平衡

配置方式：在 https://platform.openai.com/api-keys 获取 API 密钥`

	LLMFormAnthropicHelp = `Anthropic Claude 模型在注重安全合规的渗透测试中表现优异，具备出色的推理与分析能力。

PentAGI 默认模型：
• Claude Sonnet-4：高端推理模型，用于复杂安全分析与策略性漏洞评估
• Claude 3.5 Haiku：高速模型，针对快速信息收集与简单解析任务优化
• 在各类安全测试场景中兼顾成本与性能

主要优势：
• 高度重视安全与伦理，在保持安全测试效果的同时减少有害输出
• 推理能力出色，适合系统化的漏洞分析与渗透测试方法
• 上下文窗口大，非常适合分析大型代码库与配置文件
• 善于理解复杂的安全场景与合规要求

适用场景：重视负责任测试实践的安全团队、注重合规的环境、需要深入分析的场景
成本：中等定价，在推理密集型安全工作流中性价比出色

配置方式：在 https://console.anthropic.com/ 获取 API 密钥`

	LLMFormGeminiHelp = `Google Gemini 兼具多模态能力与高级推理能力，非常适合全面的安全评估。

PentAGI 默认模型：
• Gemini 2.5 Pro：高级推理模型，用于深度漏洞分析和复杂漏洞利用开发
• Gemini 2.5 Flash：高性能模型，在速度与智能之间取得平衡，适用于大多数安全测试任务
• Gemini 2.0 Flash Lite：高性价比模型，用于快速扫描和信息收集
• 具备逐步分析的推理能力，可支撑深入的渗透测试

主要优势：
• 支持多模态，可分析截图、网络拓扑图和安全文档
• 价格有竞争力，开发和测试环境的速率限制较为宽松
• 上下文窗口大（最高 200 万 tokens），可分析超大代码库和系统配置
• 在多种编程语言的代码分析与漏洞识别方面表现出色

适用场景：预算敏感的团队、开发环境，以及需要分析图片/文档的场景
成本：主流云厂商中性价比最高，性能与价格比出色

配置方式：在 https://aistudio.google.com/app/apikey 获取 API 密钥`

	LLMFormBedrockHelp = `AWS Bedrock 提供企业级的 20 多种基础模型访问能力，支持多种认证方式并具备增强的安全特性。

PentAGI 默认模型：
• Claude Sonnet-4.5（通过 Bedrock）：高端推理模型，具备 AWS 企业级安全和扩展思考能力
• OpenAI GPT OSS 120B：推理能力强，适合科学分析和复杂安全任务
• Claude Haiku-4.5、DeepSeek V3.2、Qwen3-32B：高效模型，用于特定智能体角色和成本优化
• 通过统一接口访问 Amazon Nova（多模态）、Mistral、Moonshot 等更多模型

认证方式（按优先级）：
1. AWS 默认认证（BEDROCK_DEFAULT_AUTH=true）：使用 AWS SDK 凭证链，推荐用于 EC2/ECS/Lambda
2. Bearer 令牌（BEDROCK_BEARER_TOKEN）：基于令牌的认证，适用于自定义认证场景
3. 静态凭证（ACCESS_KEY + SECRET_KEY）：传统 IAM 凭证，适用于开发和测试

主要优势：
• 企业合规：具备 SOC2、HIPAA、FedRAMP 认证，支持数据驻留和治理控制
• 多提供商接入：来自 Anthropic、Amazon、OpenAI、Qwen、DeepSeek、Cohere、Mistral、Moonshot 的 20 多种模型
• 灵活认证：三种方式适配不同部署场景和安全要求
• 增强安全：VPC 集成、CloudTrail 日志、IAM 控制、私有端点实现完全隔离
• 区域化部署：可部署在指定 AWS 区域，优化延迟并满足数据主权要求

适用场景：企业环境、受监管行业，以及需要合规控制和灵活认证的团队
成本：价格有竞争力并提供预置吞吐量选项，但新账号的速率限制较严（每分钟 2-20 次请求）
重要提示：用于生产渗透测试流程时，请通过 AWS Service Quotas 控制台申请提升配额

配置方式：选择认证方式并配置凭证。速率限制详见 https://docs.aws.amazon.com/bedrock/`

	LLMFormOllamaHelp = `Ollama 支持两种部署方式，灵活性极高。

方式一：本地 Ollama 服务器（自托管）
• 在自有硬件上运行 Ollama（建议 8GB 以上内存，GPU 可选但有帮助）
• 数据完全私有 —— 所有处理都在本地完成
• 无持续费用 —— 仅需承担基础设施成本
• 无需 API 密钥 —— 通过网络访问控制来鉴权
• 配置方式：从 https://ollama.ai/ 安装，并设置 OLLAMA_SERVER_URL=http://ollama-server:11434

方式二：Ollama Cloud（托管服务）
• 云端托管模型，无需本地基础设施
• 无需硬件 —— 模型运行在 Ollama 的基础设施上
• 按量计费，并提供免费额度
• 需要 API 密钥 —— 在 https://ollama.com/settings/keys 生成
• 配置方式：在 https://ollama.com 注册，设置 OLLAMA_SERVER_URL=https://ollama.com 和 OLLAMA_SERVER_API_KEY=你的密钥

PentAGI 默认模型：
• Llama 3.1:8b、Qwen3:32b 等开源模型
• 可自由定制 —— 可在 100 多个可用模型间切换
• 支持模型自动下载与加载，使用便捷

主要优势：
• 双部署方式：可在隐私性（本地）与便捷性（云端）之间自由选择
• 成本灵活：本地部署无持续费用，云端按量计费
• 模型库丰富：可使用最新的开源模型（Llama、Qwen、Mistral、Gemma 等）
• 支持离网环境：本地部署可在隔离网络中运行

适用场景：注重隐私的团队（本地）、预算有限的部署（云端）、有数据主权要求的组织
配置选项：从 https://10.10.10.10:11434 本地安装，或在 https://ollama.com 注册云服务`

	LLMFormDeepSeekHelp = `DeepSeek 提供推理能力出色、多语言支持良好的先进 AI 模型。

PentAGI 默认模型：
• deepseek-v4-flash：性价比高的通用模型，适用于对话、代码生成和工具调用
• deepseek-v4-pro：更高阶的推理模型，适用于复杂逻辑、数学推理和安全分析
• 定价经济，性能可与主流模型竞争

核心优势：
• 编码与推理能力强，适合安全分析和漏洞利用开发
• 支持中英双语，适用于跨国渗透测试场景
• 定价有竞争力，性价比出色
• 兼容 OpenAI 接口，可无缝集成

LiteLLM 集成（多provider 统一网关）：
• 使用 LiteLLM 代理时，将 Provider Name 设为 'deepseek'
• 无需修改 config.yml 即可启用模型前缀（如 deepseek/deepseek-v4-flash）
• 直连 DeepSeek API 时此项为可选

适用场景：需要多语言支持的团队、注重成本的部署、中文安全测试
成本：定价极具竞争力，性能表现优秀

配置方式：在 https://platform.deepseek.com/ 获取 API 密钥`

	LLMFormGLMHelp = `GLM 由智谱 AI（源自清华大学）研发，具备出色的自然语言处理与推理能力。

PentAGI 默认模型：
• GLM-4-Air：高性能通用对话模型，针对常规任务和工具调用优化
• GLM-4-Plus：旗舰模型，推理与代码生成能力强
• GLM-Z1-Plus：高阶推理模型，具备面向安全研究的深度分析能力

核心优势：
• 中英文自然语言处理能力突出
• 在多语言安全测试与分析场景中表现优异
• GLM-4 与 GLM-Z1 系列在推理和编码方面均有增强
• 兼容 OpenAI 接口，易于集成

可选 API 端点：
• 国际站：https://api.z.ai/api/paas/v4（默认）
• 中国站：https://open.bigmodel.cn/api/paas/v4
• 编码专用：https://api.z.ai/api/coding/paas/v4

LiteLLM 集成（多 provider 统一网关）：
• 使用 LiteLLM 代理时，将 Provider Name 设为 'zai'
• 无需修改 config.yml 即可启用模型前缀（如 zai/glm-4）
• 直连 GLM API 时此项为可选

适用场景：中英双语渗透测试、在亚洲市场运营的团队
成本：定价有竞争力，多语言任务性能良好

配置方式：在 https://open.bigmodel.cn/ 获取 API 密钥`

	LLMFormKimiHelp = `Kimi（Moonshot AI 出品）提供超长上下文模型，非常适合分析大型代码库和文档。

PentAGI 默认模型：
• Moonshot-v1-8k：长上下文模型，支持最多 8K tokens，用于常规对话
• Kimi-k2.5：进阶模型，推理与文档理解能力强
• 针对处理大量文本和代码做过优化

主要优势：
• 超长上下文窗口（最多 1M tokens），可完整分析整个代码库
• 中英文支持出色，适合多语言渗透测试
• 文档密集型安全评估和威胁情报分析的性价比高
• 擅长理解复杂系统架构和长篇技术文档

可选 API 端点：
• 国际站：https://api.moonshot.ai/v1（默认）
• 中国站：https://api.moonshot.cn/v1

LiteLLM 集成：
• 使用 LiteLLM 代理时，将 Provider Name 设为 'moonshot'
• 无需修改 config.yml 即可启用模型前缀（如 moonshot/kimi-k2.5）
• 直接调用 Kimi API 时可不填

适用场景：大型代码库分析、文档密集型评估、安全研究中需要超长上下文的团队
成本：定价有竞争力，长上下文场景性价比突出

配置方式：在 https://platform.moonshot.ai/ 获取 API 密钥`

	LLMFormQwenHelp = `Qwen（阿里云百炼 / DashScope 出品）提供强大的多语言模型，并具备多模态能力。

PentAGI 默认模型：
• Qwen-Turbo：最快的轻量模型，适合高频任务和实时响应场景
• Qwen-Plus：性能均衡，适合常规对话、代码生成和工具调用
• Qwen-Max：旗舰推理模型，指令遵循能力强，可处理复杂任务
• QwQ-Plus：深度推理模型，支持长链思考，适合复杂逻辑分析

主要优势：
• 多语言支持出色（中文、英文及多种其他语言）
• 借助 Qwen-VL 具备多模态能力，可做视觉化安全分析
• 可与阿里云集成，便于企业级部署
• DashScope 生态提供额外的 AI 服务和工具
• Qwen2.5、Qwen3、QwQ 系列涵盖多种规模和专长

可选 API 端点：
• 美国：https://dashscope-us.aliyuncs.com/compatible-mode/v1（默认）
• 新加坡：https://dashscope-intl.aliyuncs.com/compatible-mode/v1
• 中国：https://dashscope.aliyuncs.com/compatible-mode/v1

LiteLLM 集成：
• 使用 LiteLLM 代理时，将 Provider Name 设为 'dashscope'
• 无需修改 config.yml 即可启用模型前缀（如 dashscope/qwen-plus）
• 直接调用 Qwen API 时可不填

适用场景：亚洲市场的团队、多语言安全测试、用 Qwen-VL 做视觉分析、已接入阿里云生态
成本：定价有竞争力，分层灵活，可匹配不同使用场景

配置方式：在 https://dashscope.console.aliyun.com/ 获取 API 密钥`

	LLMFormCustomHelp = `可配置任意兼容 OpenAI 协议的 API 端点，灵活性最高，便于对接既有基础设施。

开箱可用的配置：
• vLLM 部署：本地高吞吐推理，GPU 利用率最优
• OpenRouter：通过单一 API 访问 200 多个模型，价格有竞争力
• DeepInfra：面向主流开源模型的 Serverless 推理，按量计费
• Together AI、Groq、Fireworks：可选云服务商，各有专门的性能优化
• LiteLLM Proxy：通用网关，接入 100 多家服务商，支持负载均衡和统一接口（用 LLM_SERVER_PROVIDER 指定模型前缀）
• 部分推理模型和服务商在使用工具调用时需要保留推理内容（LLM_SERVER_PRESERVE_REASONING=true）

常见本地部署方案：
• vLLM：生产级服务，支持 Qwen、Llama、Mistral，具备批处理和 GPU 优化
• LocalAI：兼容 OpenAI 协议的封装层，支持多种本地模型和向量化服务
• Text Generation WebUI：社区热门界面，模型支持广泛，可做微调
• Hugging Face TGI：企业级文本生成推理，支持自动扩缩容和监控

主要优势：
• 灵活度无上限：可对接任意兼容 OpenAI 协议的端点或服务
• 成本优化：可选择价格更优的服务商，或在自有设施上部署模型
• 摆脱厂商锁定：可在服务商和模型之间无缝切换
• 自定义微调：可部署针对自身安全测试场景训练的专用模型

适用场景：对模型有特定要求、需要控制成本，或已有 LLM 基础设施的团队
LiteLLM 集成：将 LLM_SERVER_PROVIDER 设为对应的服务商名称（如 "openrouter"、"moonshot"），即可让同一套配置文件同时用于直连 API 和 LiteLLM 代理
配置示例：容器内 /opt/pentagi/conf/ 目录下提供了主流服务商的预置配置`
)

// LLM Provider Form field labels and descriptions
const (
	LLMFormFieldBaseURL           = "基础 URL"
	LLMFormFieldAPIKey            = "API 密钥"
	LLMFormFieldDefaultAuth       = "使用 AWS 默认认证"
	LLMFormFieldBearerToken       = "Bearer 令牌"
	LLMFormFieldAccessKey         = "访问密钥 ID"
	LLMFormFieldSecretKey         = "秘密访问密钥"
	LLMFormFieldSessionToken      = "会话令牌"
	LLMFormFieldRegion            = "区域"
	LLMFormFieldModel             = "模型"
	LLMFormFieldConfigPath        = "配置文件路径"
	LLMFormFieldLegacyReasoning   = "旧版推理模式"
	LLMFormFieldPreserveReasoning = "保留推理内容"
	LLMFormFieldProviderName      = "提供商名称"
	LLMFormFieldPullTimeout       = "模型下载超时"
	LLMFormFieldPullEnabled       = "自动拉取模型"
	LLMFormFieldLoadModelsEnabled = "从服务端加载模型列表"
	LLMFormBaseURLDesc            = "该提供商的 API 端点地址"
	LLMFormAPIKeyDesc             = "用于身份验证的 API 密钥"
	LLMFormDefaultAuthDesc        = "使用 AWS SDK 默认凭证链（环境变量、EC2 角色、~/.aws/credentials），优先级最高"
	LLMFormBearerTokenDesc        = "用于身份验证的 Bearer Token，优先级高于静态凭证"
	LLMFormAccessKeyDesc          = "静态凭证认证所用的 AWS Access Key ID"
	LLMFormSecretKeyDesc          = "静态凭证认证所用的 AWS Secret Access Key"
	LLMFormSessionTokenDesc       = "临时凭证所用的 AWS Session Token（可选，与静态凭证配合使用）"
	LLMFormRegionDesc             = "Bedrock 服务所在的 AWS 区域"
	LLMFormModelDesc              = "该提供商默认使用的模型"
	LLMFormConfigPathDesc         = "配置文件路径（可选）"
	LLMFormLegacyReasoningDesc    = "启用旧版推理模式（true/false）"
	LLMFormPreserveReasoningDesc  = "在多轮对话中保留推理内容（部分提供商必需）"
	LLMFormProviderNameDesc       = "模型名称的提供商前缀（用于 LiteLLM 代理时有用）"
	LLMFormPullTimeoutDesc        = "下载模型的超时秒数（默认：600）"
	LLMFormPullEnabledDesc        = "启动时自动下载所需模型"
	LLMFormLoadModelsEnabledDesc  = "从 Ollama 服务端加载可用模型列表"
	LLMFormOllamaAPIKeyDesc       = "Ollama Cloud API 密钥（可选，使用本地 Ollama 服务端时留空）"
)

// LLM Provider Form status messages
const (
	LLMProviderFormTitle       = "LLM 提供商 %s 配置"
	LLMProviderFormDescription = "配置大语言模型提供商的相关设置"
	LLMProviderFormName        = "LLM 提供商 %s"
	LLMProviderFormOverview    = `智能体角色分配：
• Primary Agent 与 Pentester：使用推理型模型（o3、grok-4、claude-sonnet-4、gemini-2.5-pro）进行复杂漏洞分析
• Assistant 与 Adviser：使用高级模型（o4-mini、claude-sonnet-4）进行策略规划与建议
• Coder 与 Installer：使用精确型模型（gpt-4.1、claude-sonnet-4）进行漏洞利用开发与系统配置
• Searcher 与 Enricher：使用快速模型（gpt-4.1-mini、claude-3.5-haiku、gemini-2.0-flash-lite）进行信息收集
• 简单任务：使用轻量模型处理 JSON 解析和基础操作

性能考量：
• 推理型模型提供逐步分析，但速度更慢、成本更高
• 标准模型响应更快，适合高频次的智能体交互
• 每类智能体都使用针对安全测试流程优化的、提供商专属的模型配置

你的配置将决定各个智能体在不同渗透测试场景中所使用的模型。`
)

// Monitoring Screen
const (
	MonitoringTitle       = "监控配置"
	MonitoringDescription = "配置监控与可观测性平台，全面掌握系统运行状况"
	MonitoringName        = "监控"
	MonitoringOverview    = `为生产级部署提供全面的监控与可观测性能力。

监控的意义：
• 定位性能瓶颈：找出缓慢的 LLM 调用、数据库查询和系统资源占用
• 更快排查问题：详细的调用链有助于诊断分布式组件间的故障
• 优化成本：监控 token 使用模式，优化高开销的 LLM 交互
• 生产就绪：在关键环境中稳定运行的必备能力

平台选项：
Langfuse：专注 LLM 可观测性，提供会话追踪、提示词工程洞察和成本分析
Observability：全栈监控，为基础设施与应用健康提供指标、追踪、日志和告警

快速配置建议：
• 开发环境：仅启用 Langfuse 以获取 LLM 洞察
• 生产环境：同时启用两个平台以实现全面监控
• 成本敏感：使用内置模式，避免外部服务费用`
)

// Langfuse Integration constants
const (
	MonitoringLangfuseFormTitle       = "Langfuse 配置"
	MonitoringLangfuseFormDescription = "配置 Langfuse 集成，用于 LLM 监控"
	MonitoringLangfuseFormName        = "Langfuse"
	MonitoringLangfuseFormOverview    = `Langfuse（LLM 可观测性平台）提供：
• 完整的对话追踪
• 模型性能指标
• 成本监控与优化
• 用户行为分析
• AI 交互的调试链路

可选择使用内置实例或连接外部服务。`

	// Deployment types
	MonitoringLangfuseEmbedded = "内置服务"
	MonitoringLangfuseExternal = "外部服务"
	MonitoringLangfuseDisabled = "已禁用"

	// Form fields
	MonitoringLangfuseDeploymentType     = "部署方式"
	MonitoringLangfuseDeploymentTypeDesc = "选择 Langfuse 的部署方式"
	MonitoringLangfuseBaseURL            = "服务地址"
	MonitoringLangfuseBaseURLDesc        = "Langfuse 服务器地址（例如 https://cloud.langfuse.com）"
	MonitoringLangfuseProjectID          = "项目 ID"
	MonitoringLangfuseProjectIDDesc      = "Langfuse 中的项目标识符"
	MonitoringLangfusePublicKey          = "公钥"
	MonitoringLangfusePublicKeyDesc      = "用于访问项目的公开 API 密钥"
	MonitoringLangfuseSecretKey          = "私钥"
	MonitoringLangfuseSecretKeyDesc      = "用于访问项目的私密 API 密钥"
	MonitoringLangfuseListenIP           = "监听 IP"
	MonitoringLangfuseListenIPDesc       = "Docker 端口映射使用的绑定地址（例如 0.0.0.0 表示监听所有网卡）"
	MonitoringLangfuseListenPort         = "监听端口"
	MonitoringLangfuseListenPortDesc     = "Docker 为 Langfuse Web 界面暴露的外部 TCP 端口"

	// Admin settings for embedded
	MonitoringLangfuseAdminEmail        = "管理员邮箱"
	MonitoringLangfuseAdminEmailDesc    = "用于登录 Langfuse 管理后台的邮箱"
	MonitoringLangfuseAdminPassword     = "管理员密码"
	MonitoringLangfuseAdminPasswordDesc = "用于登录 Langfuse 管理后台的密码"
	MonitoringLangfuseAdminName         = "管理员用户名"
	MonitoringLangfuseAdminNameDesc     = "Langfuse 中的管理员用户名"
	MonitoringLangfuseLicenseKey        = "企业版许可证密钥"
	MonitoringLangfuseLicenseKeyDesc    = "Langfuse 企业版许可证密钥（可选）"

	// Help text
	MonitoringLangfuseModeGuide    = "选择部署方式：内置（本地自主管理）、外部（云端/已有实例）、禁用（不启用分析）"
	MonitoringLangfuseEmbeddedHelp = `内置模式会部署完整的 Langfuse 组件栈：
• PostgreSQL + ClickHouse 数据库
• MinIO S3 存储 + Redis 缓存
• 完整的 LLM 会话追踪
• 成本分析与性能指标
• 数据全部保留在你自己的服务器上

资源需求：
• 至少约 2GB 内存、5GB 磁盘空间
• 会话日志需要额外存储空间
• 自动完成部署与维护

适用场景：注重数据隐私、需要自定义配置或不希望依赖外部服务的团队。所有分析数据存储在本地，并拥有完整的管理权限。

默认管理员访问方式：
• Web 界面：http://localhost:4000
• 登录账号：admin@pentagi.com
• 密码：password（首次登录需修改）`
	MonitoringLangfuseExternalHelp = `外部模式连接 cloud.langfuse.com 或你已有的 Langfuse 服务器：

• 无需本地基础设施
• 由服务方负责升级与维护
• 可在团队间共享分析数据
• 可使用企业版功能
• 数据存储在外部服务方

配置前提：
• 拥有 Langfuse 账号及 API 密钥
• 需要可用的互联网连接
• 需要项目 ID 与认证密钥

适用场景：使用云服务、希望免运维，或需要与组织内已有 Langfuse 部署集成的团队。`
	MonitoringLangfuseDisabledHelp = `Langfuse 已禁用。缺少 LLM 可观测性后，你将无法获得：

• 会话历史追踪
• Token 用量与成本分析
• 模型性能指标
• AI 交互的调试链路
• 用户行为分析
• 提示词优化洞察

建议在生产环境中启用，
以便监控 AI 智能体表现
并有效优化成本。`
)

// Graphiti Integration constants
const (
	MonitoringGraphitiFormTitle       = "Graphiti 配置（测试版）"
	MonitoringGraphitiFormDescription = "配置 Graphiti 知识图谱集成"
	MonitoringGraphitiFormName        = "Graphiti（测试版）"
	MonitoringGraphitiFormOverview    = `⚠️  测试功能：该功能仍在积极开发中，请关注后续更新以获取功能改进与稳定性修复。

Graphiti 提供时序知识图谱能力：
• 实体与关系抽取
• 面向 AI 智能体的语义记忆
• 时序上下文追踪
• 跨任务流复用已有知识

⚠️  前置要求：Graphiti 需要已配置的 OpenAI 提供商（LLM 提供商 → OpenAI）来完成实体抽取。

请选择使用内置实例还是连接外部服务。`

	// Deployment types
	MonitoringGraphitiEmbedded = "内置组件栈"
	MonitoringGraphitiExternal = "外部服务"
	MonitoringGraphitiDisabled = "已禁用"

	// Form fields
	MonitoringGraphitiDeploymentType     = "部署方式"
	MonitoringGraphitiDeploymentTypeDesc = "选择 Graphiti 的部署方式"
	MonitoringGraphitiURL                = "Graphiti 服务地址"
	MonitoringGraphitiURLDesc            = "Graphiti API 服务器地址"
	MonitoringGraphitiTimeout            = "请求超时"
	MonitoringGraphitiTimeoutDesc        = "Graphiti 操作的超时时间（秒）"
	MonitoringGraphitiModelName          = "抽取模型"
	MonitoringGraphitiModelNameDesc      = "用于实体抽取的 LLM 模型（使用 LLM 提供商配置中的 OpenAI 提供商）"
	MonitoringGraphitiNeo4jUser          = "Neo4j 用户名"
	MonitoringGraphitiNeo4jUserDesc      = "访问 Neo4j 数据库的用户名"
	MonitoringGraphitiNeo4jPassword      = "Neo4j 密码"
	MonitoringGraphitiNeo4jPasswordDesc  = "访问 Neo4j 数据库的密码"
	MonitoringGraphitiNeo4jDatabase      = "Neo4j 数据库"
	MonitoringGraphitiNeo4jDatabaseDesc  = "Neo4j 数据库名称"

	// Help text
	MonitoringGraphitiModeGuide    = "选择部署方式：内置（本地 Neo4j）、外部（已有 Graphiti）、禁用（不启用知识图谱）"
	MonitoringGraphitiEmbeddedHelp = `⚠️  测试功能：该功能仍在积极开发中，请关注后续更新以获取改进。

内置方式会部署完整的 Graphiti 组件栈：
• Neo4j 图数据库
• Graphiti API 服务
• 自动从智能体交互中抽取实体
• 时序关系追踪
• 知识图谱私有化保存在你的服务器上

前置条件：
• 必须已配置 OpenAI 提供商（LLM 提供商 → OpenAI）
• 实体抽取会使用 OpenAI API 密钥
• 已配置的模型将用于知识图谱相关操作

资源需求：
• 至少约 1.5GB 内存、3GB 磁盘空间
• Neo4j 界面：http://localhost:7474
• Graphiti API：http://localhost:8000
• 自动完成部署与维护

适用场景：希望使用知识图谱能力，同时完全掌控数据与隐私的团队。`
	MonitoringGraphitiExternalHelp = `⚠️  测试功能：该功能仍在积极开发中，请关注后续更新以获取改进。

外部方式连接到你已有的 Graphiti 服务器：

• 无需本地基础设施
• 更新与维护由对方托管
• 团队之间可共享知识图谱
• 数据保存在外部服务方

配置要求：
• 需要 Graphiti 服务器地址及访问权限
• 需要网络连通
• 外部服务器必须已配置 OpenAI API 密钥
• 模型与抽取相关设置需在外部服务器上配置

适用场景：已有 Graphiti 部署或使用云服务的团队。`
	MonitoringGraphitiDisabledHelp = `Graphiti 已禁用。你将无法使用：

• 时序知识图谱
• 实体与关系抽取
• 智能体的语义记忆
• 跨任务流复用知识
• 高级上下文检索

注意：Graphiti 目前处于测试阶段。
生产环境建议启用，以便将渗透测试
结果沉淀为知识库。`
)

// Observability Integration constants
const (
	MonitoringObservabilityFormTitle       = "Observability 配置"
	MonitoringObservabilityFormDescription = "配置监控与可观测性组件栈"
	MonitoringObservabilityFormName        = "可观测性"
	MonitoringObservabilityFormOverview    = `可观测性组件栈包含：
• Grafana —— 可视化仪表盘
• VictoriaMetrics —— 时序数据存储
• Jaeger —— 分布式链路追踪
• Loki —— 日志聚合
• OpenTelemetry —— 数据采集

用于监控 PentAGI 的性能与系统健康状况。`

	// Deployment types
	MonitoringObservabilityEmbedded = "内置组件栈"
	MonitoringObservabilityExternal = "外部采集器"
	MonitoringObservabilityDisabled = "已禁用"

	// Form fields
	MonitoringObservabilityDeploymentType     = "部署方式"
	MonitoringObservabilityDeploymentTypeDesc = "选择监控组件的部署方式"
	MonitoringObservabilityOTelHost           = "OpenTelemetry 主机"
	MonitoringObservabilityOTelHostDesc       = "外部 OpenTelemetry 采集器的地址"

	// embedded listen fields
	MonitoringObservabilityGrafanaListenIP        = "Grafana 监听 IP"
	MonitoringObservabilityGrafanaListenIPDesc    = "Docker 端口映射所用的绑定地址（例如 0.0.0.0 表示在所有网络接口上开放）"
	MonitoringObservabilityGrafanaListenPort      = "Grafana 监听端口"
	MonitoringObservabilityGrafanaListenPortDesc  = "Docker 为 Grafana 网页界面开放的外部 TCP 端口"
	MonitoringObservabilityOTelGrpcListenIP       = "OTel gRPC 监听 IP"
	MonitoringObservabilityOTelGrpcListenIPDesc   = "Docker 端口映射所用的绑定地址（例如 0.0.0.0 表示在所有网络接口上开放）"
	MonitoringObservabilityOTelGrpcListenPort     = "OTel gRPC 监听端口"
	MonitoringObservabilityOTelGrpcListenPortDesc = "Docker 为 OTel gRPC 接收端开放的外部 TCP 端口"
	MonitoringObservabilityOTelHttpListenIP       = "OTel HTTP 监听 IP"
	MonitoringObservabilityOTelHttpListenIPDesc   = "Docker 端口映射所用的绑定地址（例如 0.0.0.0 表示在所有网络接口上开放）"
	MonitoringObservabilityOTelHttpListenPort     = "OTel HTTP 监听端口"
	MonitoringObservabilityOTelHttpListenPortDesc = "Docker 为 OTel HTTP 接收端开放的外部 TCP 端口"

	// Help text
	MonitoringObservabilityModeGuide    = "选择监控方式：内置（完整组件栈）、外部（已有基础设施）、禁用（不启用监控）"
	MonitoringObservabilityEmbeddedHelp = `内置方式会部署完整的监控体系：
• Grafana 仪表盘与告警
• VictoriaMetrics 时序数据库
• Jaeger 分布式追踪界面
• Loki 日志聚合系统
• ClickHouse 分析型数据库
• Node Exporter + cAdvisor 指标采集
• OpenTelemetry 数据采集

各组件已自动接入监控，并预置了用于
系统健康、性能分析与问题排查的仪表盘。

资源需求：
• 至少约 1.5GB 内存、3GB 磁盘空间
• Grafana 界面：http://localhost:3000
• 性能分析：http://localhost:7777

适用场景：需要完整的系统可见性、
问题排查与性能调优。`
	MonitoringObservabilityExternalHelp = `外部方式会将遥测数据发送到你已有的监控基础设施：

• 使用基于 HTTP/2 的 OTLP 协议（不启用 TLS）
• 你的采集器必须支持：
  - OTLP HTTP 接收端（端口 4318）
  - OTLP gRPC 接收端（端口 8148）
  - tls: insecure: true 配置
• 发送指标、链路追踪与日志
• 兼容各类企业级平台：
  Datadog、New Relic、Splunk 等

OTEL_HOST 示例：
your-collector:4318

采集器配置要求：
tls: insecure: true

适用场景：已有监控基础设施或
统一可观测性平台的组织。`
	MonitoringObservabilityDisabledHelp = `可观测性已禁用。你将无法使用：

• 系统性能监控
• 分布式请求追踪
• 结构化日志聚合
• 资源使用分析
• 错误追踪与告警
• 性能瓶颈分析

生产环境建议启用，以便监控系统健康、
排查问题并有效优化性能。`
)

// Summarizer Screen
const (
	SummarizerTitle       = "摘要器配置"
	SummarizerDescription = "启用对话摘要，以降低 LLM 成本并改善上下文管理"
	SummarizerName        = "摘要器"
	SummarizerOverview    = `优化上下文占用、降低 LLM 成本，并与模型能力相匹配。

何时需要调整摘要设置：
• Token 成本偏高：缩减上下文规模（4K-12K，而非 22K+ tokens）
• 出现"上下文过长"错误：按模型的上限进行配置
• 对话连贯性差：提升上下文保留量以改善质量
• 模型类型不同：针对短上下文与长上下文模型分别调优

General（通用摘要）：成本控制力最强、参数可精细调节，适合研究与分析类任务
Assistant（助手摘要）：对话质量最佳，具备智能上下文管理，适合交互式会话

快速见效的做法：
• 降低成本：使用 General，将"保留最近分段数"减至 1-2
• 上下文报错：将各项上限与你的模型对齐（8K/32K/128K）
• 优先保质量：使用 Assistant 并适当提高上限`

	SummarizerTypeGeneralName = "通用摘要（General）"
	SummarizerTypeGeneralDesc = "用于对话上下文管理的全局摘要设置"

	SummarizerTypeGeneralInfo = `适用于追求最强成本控制、并需兼容短上下文模型的场景。

在以下情况下最为合适：
• 需要大幅降低成本：可精调每个参数，将 token 用量压到最低
• 短上下文模型（8K-32K）：通过精确上限避免溢出错误
• 研究/分析类任务：在不丢失关键数据的前提下受控压缩
• 自定义问答处理：完全掌控问答对的处理方式

典型效果：
• 相比默认设置可降低 40-70% 成本
• 上下文规模 4K-12K tokens（Assistant 模式为 22K+）
• 在 GPT-3.5、Claude Instant 等较小模型上表现更好
• 可精确权衡"保留对话记忆"与"重置上下文"

最佳实践：
• 从 1-2 个"保留最近分段数"开始，以获得最大成本节省
• 启用"尺寸管理"以自动防止溢出
• 仅在关键推理任务中才关闭问答压缩`

	SummarizerTypeAssistantName = "助手摘要（Assistant）"
	SummarizerTypeAssistantDesc = "面向 AI 助手场景的专用摘要设置"

	SummarizerTypeAssistantInfo = `适用于追求最佳对话质量与连贯性的场景。

在以下情况下最为合适：
• 长推理链路：为复杂的多步思考维持上下文
• 高质量对话：保留对话流程与助手的表达风格
• 长上下文模型（64K+）：高效发挥模型的完整能力
• 交互式会话：更好地记住用户偏好与历史对话

典型效果：
• 上下文规模 8K-40K tokens，并具备智能压缩
• 对话连贯性显著优于手工设置
• 针对推理任务自动优化上下文
• 成本与质量兼顾（上下文约为 General 模式的 3 倍）

最佳实践：
• 多数场景直接使用默认值即可，它们已经过预优化
• 仅在任务非常复杂时才增加"保留最近分段数"
• 关注上下文用量，成本随 token 数线性增长
• 非常适合 GPT-4、Claude 等大上下文模型`
)

// Summarizer Form Screen
const (
	SummarizerFormGeneralTitle   = "通用摘要器配置"
	SummarizerFormAssistantTitle = "助手摘要器配置"
	SummarizerFormDescription    = "配置 %s 设置"

	// Field Labels and Descriptions
	SummarizerFormPreserveLast     = "尺寸管理"
	SummarizerFormPreserveLastDesc = "控制最后一个分段的压缩行为。启用：分段大小受 LastSecBytes 限制（上下文更小）。禁用：分段自由增长（上下文更大）"

	SummarizerFormUseQA     = "问答摘要"
	SummarizerFormUseQADesc = "当问答内容总量超过 MaxQABytes 或 MaxQASections 上限时，启用问答对压缩"

	SummarizerFormSumHumanInQA     = "压缩用户消息"
	SummarizerFormSumHumanInQADesc = "将用户消息一并纳入问答压缩。禁用：保留用户原文（多数场景推荐）"

	SummarizerFormLastSecBytes     = "分段大小上限"
	SummarizerFormLastSecBytesDesc = "启用尺寸管理时，每个最近分段的最大字节数。值越大：单个分段细节越多，token 用量越高"

	SummarizerFormMaxBPBytes     = "回复大小上限"
	SummarizerFormMaxBPBytesDesc = "单条 AI 回复在被压缩前的最大字节数。避免单条超长回复占据整个上下文"

	SummarizerFormMaxQASections     = "问答分段数上限"
	SummarizerFormMaxQASectionsDesc = "触发问答压缩前允许的最大问答分段数。与 MaxQABytes 配合，共同控制问答记忆总量"

	SummarizerFormMaxQABytes     = "问答记忆总量"
	SummarizerFormMaxQABytesDesc = "所有问答分段合计的最大字节数。超出后（与 MaxQASections 共同判定）触发问答压缩以压回上限内"

	SummarizerFormKeepQASections     = "保留最近分段数"
	SummarizerFormKeepQASectionsDesc = "不做压缩、原样保留的最近对话分段数量。影响上下文规模的首要参数"

	// Enhanced Help Text - General (common principles)
	SummarizerFormGeneralHelp = `上下文规模估算：常见 4K-22K tokens，参数拉满时可达 94K。

关键影响因素：
• 保留最近分段数：影响最大，每增加 1 个约增加 1.5-9K tokens
• 关闭尺寸管理：上下文增大 2-3 倍（压缩更少）
• 分段/回复上限：控制单个组成部分的大小
• 问答记忆：在超出上限时管理整体对话历史

参数之间的相互作用：
• 只有 MaxQABytes 与 MaxQASections 同时超限，才会触发问答压缩
• 禁用尺寸管理后，分段可增长到上限的 2 倍
• 回复上限可避免单条超长输出占满上下文
• 压缩用户消息（SummHumanInQA）约节省 5%，但会丢失原始措辞

小上下文模型建议下调：
• 保留最近分段数：1-2（默认为 3 以上）
• 分段上限：25-35KB（默认 50KB 以上）
• 简单对话可禁用尺寸管理

常见误区：
• 保留最近分段数设置过高（上下文溢出的主要原因）
• 启用尺寸管理的同时把分段上限设得过低（过度压缩）
• 问答上限不匹配（字节数高但分段数低，等于无效）

当前算法会压缩较早的内容，同时保持最近上下文的质量。`

	// Enhanced Help Text - Assistant specific (interactive conversations)
	SummarizerFormAssistantHelp = `针对需要保持上下文连续性的交互式对话进行了优化。

默认配置（3 个保留区段、75KB 上限）：
• 典型区间：8K-40K tokens
• 适用场景：长对话、推理链、依赖上下文的任务
• 模型要求：在 32K 以上上下文的模型上表现良好

按模型类型调整：
• 短上下文（≤16K）：保留区段数=1-2，区段大小上限=45KB
• 长上下文（128K+）：可将保留区段数提高到 5-7
• 高频对话：将保留区段数降到 2，以获得更快响应

进阶调优：
• 文档分析类对话可将问答记忆上限设为 200KB 以上
• 详尽的技术回答可将回复大小上限设为 24-32KB
• 保持用户消息不被压缩（SummHumanInQA=false）以获得更好的上下文

性能优化：
• 在助手模式下，每个保留区段约占 9-18KB
• 启用大小管理可减少约 20% 的增长，但可能丢失细节
• 由于默认上限较大，问答压缩触发的频率更低

大小管理默认启用，在防止上下文溢出的同时保持对话流畅。
建议先观察实际 token 用量，优先调整保留区段数，再调整各项上限。`

	// Context size estimation
	SummarizerContextEstimatedSize    = "预计上下文规模：%s\n%s"
	SummarizerContextTokenRange       = "约 %s tokens"
	SummarizerContextTokenRangeMinMax = "约 %s-%s tokens"
	SummarizerContextRequires256K     = "需要 256K 以上上下文的模型"
	SummarizerContextRequires128K     = "需要 128K 以上上下文的模型"
	SummarizerContextRequires64K      = "需要 64K 以上上下文的模型"
	SummarizerContextRequires32K      = "需要 32K 以上上下文的模型"
	SummarizerContextRequires16K      = "需要 16K 以上上下文的模型"
	SummarizerContextFitsIn8K         = "可在 8K 以上上下文的模型中运行"
)

// Tools screen strings
const (
	ToolsTitle       = "工具配置"
	ToolsDescription = "通过附加工具和选项增强智能体能力"
	ToolsName        = "工具"
	ToolsOverview    = `为 AI 智能体配置附加工具与能力。
每个工具都可按需启用并单独配置。

可用设置：
• Human-in-the-loop（人工介入）- 允许测试过程中由用户参与决策
• AI Agents Settings（智能体设置）- 配置 AI 智能体的全局行为
• Search Engines（搜索引擎）- 配置外部搜索服务商
• Scraper（网页抓取）- 网页内容提取与分析
• Graphiti (beta)（知识图谱）- 用于语义记忆的时序知识图谱
• Docker（容器环境）- 容器运行环境配置`
)

// Server Settings screen strings
const (
	ServerSettingsFormTitle       = "服务器设置"
	ServerSettingsFormDescription = "配置 PentAGI 服务器的网络访问与对外路由"
	ServerSettingsFormName        = "服务器设置"
	ServerSettingsFormOverview    = `• 网络绑定 - 控制 PentAGI 监听的网卡与端口
• 公开 URL - 用于重定向的外部地址及可选的基础路径
• CORS - 允许通过浏览器访问的来源
• 代理 - 访问 LLM/搜索服务商时使用的 HTTP/HTTPS 代理
• SSL 目录 - 存放 server.crt 与 server.key（PEM 格式）的自定义证书目录
• 数据目录 - 智能体产物与任务流工作区的持久化存储`

	// Field labels and descriptions
	ServerSettingsLicenseKey     = "许可证密钥"
	ServerSettingsLicenseKeyDesc = "PentAGI 许可证密钥，格式为 XXXX-XXXX-XXXX-XXXX"

	ServerSettingsHost     = "服务器主机（监听 IP）"
	ServerSettingsHostDesc = "Docker 端口映射使用的绑定地址（例如 0.0.0.0 表示在所有网卡上暴露）"

	ServerSettingsPort     = "服务器端口（监听端口）"
	ServerSettingsPortDesc = "Docker 为 PentAGI Web 界面对外暴露的 TCP 端口"

	ServerSettingsPublicURL     = "公开 URL"
	ServerSettingsPublicURLDesc = "用于重定向和链接的公开基础 URL（支持基础路径，例如 https://example.com/pentagi/）"

	ServerSettingsCORSOrigins     = "CORS 允许来源"
	ServerSettingsCORSOriginsDesc = "以逗号分隔的允许来源列表（例如 https://localhost:8443,https://localhost）"

	ServerSettingsProxyURL     = "HTTP/HTTPS 代理"
	ServerSettingsProxyURLDesc = "访问 LLM 和外部工具时使用的代理（不用于 Docker API 访问）"

	ServerSettingsProxyUsername     = "代理用户名"
	ServerSettingsProxyUsernameDesc = "代理认证用户名（可选）"
	ServerSettingsProxyPassword     = "代理密码"
	ServerSettingsProxyPasswordDesc = "代理认证密码（可选）"

	ServerSettingsHTTPClientTimeout       = "HTTP 客户端超时"
	ServerSettingsHTTPClientTimeoutDesc   = "调用外部 API（LLM 服务商、搜索引擎等）的超时秒数"
	ServerSettingsTerminalToolTimeout     = "终端工具超时"
	ServerSettingsTerminalToolTimeoutDesc = "终端命令的默认超时秒数（填 0 或负数表示使用 3 小时上限）"

	ServerSettingsExternalSSLCAPath     = "自定义 CA 证书路径"
	ServerSettingsExternalSSLCAPathDesc = "容器内自定义根 CA 证书的路径（例如 /opt/pentagi/ssl/ca-bundle.pem）"

	ServerSettingsExternalSSLInsecure     = "跳过 SSL 校验"
	ServerSettingsExternalSSLInsecureDesc = "禁用 SSL/TLS 证书校验（仅用于自签名证书的测试场景）"

	ServerSettingsSSLDir     = "SSL 目录"
	ServerSettingsSSLDirDesc = "存放 PEM 格式 server.crt 与 server.key 的目录（server.crt 可包含完整证书链）"

	ServerSettingsDataDir     = "数据目录"
	ServerSettingsDataDirDesc = "存放所有智能体生成文件的目录；其中的 flow-N 子目录会挂载为工作容器内的 /work"

	ServerSettingsCookieSigningSalt     = "Cookie 签名盐值"
	ServerSettingsCookieSigningSaltDesc = "用于签名 Cookie 的密钥（请妥善保密）"

	// Hints for fields overview
	ServerSettingsLicenseKeyHint          = "许可证密钥"
	ServerSettingsHostHint                = "监听 IP"
	ServerSettingsPortHint                = "监听端口"
	ServerSettingsPublicURLHint           = "公开 URL"
	ServerSettingsCORSOriginsHint         = "CORS 来源"
	ServerSettingsProxyURLHint            = "代理地址"
	ServerSettingsProxyUsernameHint       = "代理用户名"
	ServerSettingsProxyPasswordHint       = "代理密码"
	ServerSettingsHTTPClientTimeoutHint   = "HTTP 超时"
	ServerSettingsTerminalToolTimeoutHint = "终端超时"
	ServerSettingsExternalSSLCAPathHint   = "自定义 CA 路径"
	ServerSettingsExternalSSLInsecureHint = "跳过 SSL 校验"
	ServerSettingsSSLDirHint              = "SSL 目录"
	ServerSettingsDataDirHint             = "数据目录"

	// Help texts per-field
	ServerSettingsGeneralHelp = `PentAGI 通过 Docker 对外提供网页界面，主机地址与端口均可配置。

公开 URL 必须与用户实际访问服务器的方式一致。如果使用子路径（例如 /pentagi/），请一并填入。CORS 控制来自指定来源的浏览器访问。代理会影响发往 LLM/搜索提供商以及工具所用其他外部服务的出站流量。

SSL 目录用于提供自定义证书。设置后，服务器将使用该目录下的 server.crt 和 server.key。数据目录用于存放任务流的产物与工作文件。`

	ServerSettingsLicenseKeyHelp = `PentAGI 许可证密钥，格式为 XXXX-XXXX-XXXX-XXXX，用于与 PentAGI Cloud API 通信。`

	ServerSettingsHostHelp = `docker-compose 端口映射中对外发布端口所用的绑定地址。

示例：
• 127.0.0.1 —— 仅本机访问
• 0.0.0.0 —— 在所有网络接口上开放`

	ServerSettingsPortHelp = `PentAGI 界面对外使用的端口，必须在主机上未被占用。示例：8443。`

	ServerSettingsPublicURLHelp = `设置重定向和链接中使用的公开基础 URL。

示例：
• http://localhost:8443
• https://example.com/
• https://example.com/pentagi/（带子路径）`

	ServerSettingsCORSOriginsHelp = `允许浏览器访问的来源列表，以逗号分隔。`

	ServerSettingsProxyURLHelp = `用于发往 LLM 提供商和外部工具的出站请求的 HTTP 或 HTTPS 代理。不用于 Docker API 通信。`

	ServerSettingsHTTPClientTimeoutHelp = `所有外部 HTTP/HTTPS API 调用的超时时间（秒），包括：
• LLM 提供商请求（OpenAI、Anthropic、Bedrock 等）
• 搜索引擎查询（Google、Tavily、Perplexity 等）
• 外部工具集成
• 向量嵌入生成请求

默认值：600 秒（10 分钟）
设为 0 表示不限制超时（不建议在生产环境使用）
数值过低可能导致正常的长耗时请求失败。`

	ServerSettingsTerminalToolTimeoutHelp = `当智能体请求 timeout=0 或负数超时值时所采用的默认超时时间（秒）。

该设置影响通过隔离终端容器执行的命令，包括各类扫描器和命令行工具。

默认值：1200 秒（20 分钟）
允许范围：1–10800 秒（最多 3 小时）
小于等于 0 或大于 10800 的值会被限制为最大值（10800 秒 = 3 小时）；智能体永远不允许无限期运行。
工具调用显式提供的超时值若在 1–10800 秒范围内，将覆盖此默认值。`

	ServerSettingsExternalSSLCAPathHelp = `容器内自定义 CA 证书文件（PEM 格式）的路径。

必须指向 /opt/pentagi/ssl/ 目录，该目录由主机上的 pentagi-ssl 卷挂载而来。

示例：
• /opt/pentagi/ssl/ca-bundle.pem
• /opt/pentagi/ssl/corporate-ca.pem

该文件可包含多个根证书和中间证书。`

	ServerSettingsExternalSSLInsecureHelp = `禁用与 LLM 提供商及外部服务连接时的 SSL/TLS 证书校验。

⚠ 警告：仅可用于自签名证书的测试场景，切勿在生产环境启用。

启用后将完全跳过证书校验，连接会面临中间人攻击风险。`

	ServerSettingsSSLDirHelp = `存放 PEM 格式 server.crt 与 server.key 的目录路径。server.crt 可包含完整证书链。此设置会覆盖默认的自动生成证书行为。`

	ServerSettingsDataDirHelp = `用于存放持久化数据的主机目录。PentAGI 会将智能体产物存放在 flow-N 子目录下，这些目录会映射到工作容器内的 /work。`

	ServerSettingsCookieSigningSaltHelp = `用于签名 Cookie 的密钥盐值，请妥善保密。`
)

// Human-in-the-loop screen strings
const (
	// AI Agents Settings screen strings
	ToolsAIAgentsSettingsFormTitle       = "AI 智能体设置"
	ToolsAIAgentsSettingsFormDescription = "配置 AI 智能体的全局行为"
	ToolsAIAgentsSettingsFormName        = "AI 智能体设置"
	ToolsAIAgentsSettingsFormOverview    = `本节配置 PentAGI 中 AI 智能体的全局行为。

基础设置：
• 启用用户交互：允许智能体在需要时请求用户输入
• 启用多智能体模式：允许助手调度多个专职智能体

执行监控（⚠️  测试版）：
• 启用执行监控：由导师智能体自动监督并分析执行模式
• 相同工具调用阈值：触发导师复查前允许的连续相同工具调用次数
• 累计工具调用阈值：触发导师复查前允许的累计工具调用次数

工具调用上限：
• 最大工具调用数（通用智能体）：防止 Assistant、Primary Agent、Pentester、Coder、Installer 失控执行
• 最大工具调用数（受限智能体）：防止 Searcher、Enricher、Memorist 等失控执行

任务规划（⚠️  测试版）：
• 启用任务规划：为专职智能体生成结构化执行计划

⚠️  测试版功能仍在积极开发中，建议仅用于测试。`

	// field labels and descriptions
	ToolsAIAgentsSettingHumanInTheLoop          = "启用用户交互"
	ToolsAIAgentsSettingHumanInTheLoopDesc      = "允许智能体在需要时请求用户输入"
	ToolsAIAgentsSettingUseAgents               = "启用多智能体模式"
	ToolsAIAgentsSettingUseAgentsDesc           = "允许助手调度多个专职智能体"
	ToolsAIAgentsSettingExecutionMonitor        = "启用执行监控（测试版）"
	ToolsAIAgentsSettingExecutionMonitorDesc    = "自动调用导师智能体分析执行模式"
	ToolsAIAgentsSettingSameToolLimit           = "相同工具调用阈值"
	ToolsAIAgentsSettingSameToolLimitDesc       = "触发导师复查前允许的连续相同工具调用次数"
	ToolsAIAgentsSettingTotalToolLimit          = "累计工具调用阈值"
	ToolsAIAgentsSettingTotalToolLimitDesc      = "触发导师复查前允许的累计工具调用次数"
	ToolsAIAgentsSettingMaxGeneralToolCalls     = "最大工具调用数（通用智能体）"
	ToolsAIAgentsSettingMaxGeneralToolCallsDesc = "Assistant、Primary Agent、Pentester、Coder、Installer 的最大工具调用次数"
	ToolsAIAgentsSettingMaxLimitedToolCalls     = "最大工具调用数（受限智能体）"
	ToolsAIAgentsSettingMaxLimitedToolCallsDesc = "Searcher、Enricher、Memorist 等的最大工具调用次数"
	ToolsAIAgentsSettingTaskPlanning            = "启用任务规划（测试版）"
	ToolsAIAgentsSettingTaskPlanningDesc        = "为专职智能体生成结构化执行计划"

	// help content
	ToolsAIAgentsSettingsHelp = `AI 智能体设置决定智能体如何协作、如何与用户交互以及如何进行执行控制。

基础设置：
• 启用用户交互：允许智能体在需要时请求用户输入
• 启用多智能体模式：允许助手针对复杂任务调度专职智能体

执行监控（⚠️  测试版）：
自动调用 adviser（导师）分析执行模式、检测循环、提出替代策略，避免智能体死守单一思路。阈值：连续相同调用次数（默认 5）与累计调用次数（默认 10）。

任务规划（⚠️  测试版）：
在专职智能体开始工作前生成 3-7 步执行计划，可抑制范围蔓延并提升成功率。当 adviser 采用增强配置（更强模型或最高推理档位）时效果最佳。

工具调用上限（始终生效）：
硬性上限用于防止无限循环：通用智能体默认 100，受限智能体默认 20。该限制独立于测试版功能。

参数量低于 32B 的开源模型（Qwen3.5-27B、DeepSeek-V3、Llama-3.1-70B）：
✓ 建议同时启用两项测试版功能——这对结果质量至关重要
✓ 实测结果质量相比基线提升约 2 倍
✓ 为 adviser 配置增强设置可获得最佳表现
✓ 非常适合使用本地 LLM 推理的隔离网络部署

性能影响：token 消耗与耗时增加 2-3 倍；对低于 32B 的模型质量提升约 2 倍。

⚠️  测试版提醒：功能仍在积极开发中。尽管处于测试阶段，仍推荐参数量低于 32B 的开源模型启用；若使用云端大模型 API，建议保持关闭。

注意：修改后需重启服务生效。`
)

// Search Engines screen strings
const (
	ToolsSearchEnginesFormTitle       = "搜索引擎配置"
	ToolsSearchEnginesFormDescription = "配置搜索引擎，供 AI 智能体在测试过程中收集情报"
	ToolsSearchEnginesFormName        = "搜索引擎"
	ToolsSearchEnginesFormOverview    = `可用的搜索引擎：
• DuckDuckGo —— 免费搜索引擎（无需 API 密钥）
• Sploitus —— 漏洞利用与漏洞情报库（无需 API 密钥）
• Perplexity —— 带推理能力的 AI 搜索
• Tavily —— 面向 AI 应用的搜索 API
• Traversaal —— 网页抓取与搜索
• Google Search —— 需要 API 密钥和自定义搜索引擎 ID
• Searxng —— 互联网聚合搜索引擎

API 密钥获取地址：
• Perplexity: https://www.perplexity.ai/
• Tavily: https://tavily.com/
• Traversaal: https://traversaal.ai/
• Google: https://developers.google.com/custom-search/v1/introduction`

	ToolsSearchEnginesDuckDuckGo                = "DuckDuckGo 搜索"
	ToolsSearchEnginesDuckDuckGoName            = "DuckDuckGo"
	ToolsSearchEnginesDuckDuckGoDesc            = "启用 DuckDuckGo 搜索（无需 API 密钥）"
	ToolsSearchEnginesDuckDuckGoRegion          = "DuckDuckGo 区域"
	ToolsSearchEnginesDuckDuckGoRegionDesc      = "DuckDuckGo 区域代码（例如 us-en、uk-en、cn-zh）"
	ToolsSearchEnginesDuckDuckGoSafeSearch      = "DuckDuckGo 安全搜索"
	ToolsSearchEnginesDuckDuckGoSafeSearchDesc  = "DuckDuckGo 安全搜索级别（strict 严格、moderate 中等、off 关闭）"
	ToolsSearchEnginesDuckDuckGoTimeRange       = "DuckDuckGo 时间范围"
	ToolsSearchEnginesDuckDuckGoTimeRangeDesc   = "DuckDuckGo 时间范围（d: 天、w: 周、m: 月、y: 年）"
	ToolsSearchEnginesSploitus                  = "Sploitus 搜索"
	ToolsSearchEnginesSploitusName              = "Sploitus"
	ToolsSearchEnginesSploitusDesc              = "启用 Sploitus 漏洞利用与漏洞信息搜索（无需 API 密钥）"
	ToolsSearchEnginesPerplexityKey             = "Perplexity API 密钥"
	ToolsSearchEnginesPerplexityName            = "Perplexity"
	ToolsSearchEnginesPerplexityKeyDesc         = "Perplexity AI 搜索的 API 密钥"
	ToolsSearchEnginesPerplexityModel           = "Perplexity 模型"
	ToolsSearchEnginesPerplexityModelDesc       = "选择 Perplexity 使用的模型"
	ToolsSearchEnginesPerplexityContextSize     = "Perplexity 上下文大小"
	ToolsSearchEnginesPerplexityContextSizeDesc = "选择 Perplexity 使用的上下文大小"
	ToolsSearchEnginesTavilyKey                 = "Tavily API 密钥"
	ToolsSearchEnginesTavilyName                = "Tavily"
	ToolsSearchEnginesTavilyKeyDesc             = "Tavily 搜索服务的 API 密钥"
	ToolsSearchEnginesTraversaalKey             = "Traversaal API 密钥"
	ToolsSearchEnginesTraversaalName            = "Traversaal"
	ToolsSearchEnginesTraversaalKeyDesc         = "Traversaal 网页抓取的 API 密钥"
	ToolsSearchEnginesGoogleKey                 = "Google 搜索 API 密钥"
	ToolsSearchEnginesGoogleKeyDesc             = "Google 自定义搜索 API 密钥"
	ToolsSearchEnginesGoogleName                = "Google 搜索"
	ToolsSearchEnginesGoogleCX                  = "Google 搜索引擎 ID"
	ToolsSearchEnginesGoogleCXDesc              = "Google 自定义搜索引擎 ID"
	ToolsSearchEnginesGoogleLR                  = "Google 语言限制"
	ToolsSearchEnginesGoogleLRDesc              = "Google 搜索引擎的语言限制（例如 lang_en、lang_cn 等）"
	ToolsSearchEnginesSearxngURL                = "Searxng 搜索地址"
	ToolsSearchEnginesSearxngName               = "Searxng"
	ToolsSearchEnginesSearxngURLDesc            = "Searxng 搜索引擎的 URL"
	ToolsSearchEnginesSearxngCategories         = "Searxng 搜索分类"
	ToolsSearchEnginesSearxngCategoriesDesc     = "Searxng 搜索引擎分类（例如 general、it、web、news、technology、science、health、other）"
	ToolsSearchEnginesSearxngLanguage           = "Searxng 搜索语言"
	ToolsSearchEnginesSearxngLanguageDesc       = "Searxng 搜索引擎语言（en、ch、fr、de、it、es、pt、ru、zh，留空表示所有语言）"
	ToolsSearchEnginesSearxngSafeSearch         = "Searxng 安全搜索"
	ToolsSearchEnginesSearxngSafeSearchDesc     = "Searxng 搜索引擎安全搜索级别（0: 关闭、1: 中等、2: 严格）"
	ToolsSearchEnginesSearxngTimeRange          = "Searxng 时间范围"
	ToolsSearchEnginesSearxngTimeRangeDesc      = "Searxng 搜索引擎时间范围（day 天、month 月、year 年）"
	ToolsSearchEnginesSearxngTimeout            = "Searxng 超时"
	ToolsSearchEnginesSearxngTimeoutDesc        = "Searxng 请求超时时间（秒）"
)

// Scraper screen strings
const (
	ToolsScraperFormTitle       = "网页抓取服务配置"
	ToolsScraperFormDescription = "配置网页内容抓取服务"
	ToolsScraperFormName        = "网页抓取"
	ToolsScraperFormOverview    = `网页抓取服务使用 vxcontrol/scraper Docker 镜像提取并分析网页内容。

运行模式：
• 内置 —— 运行本地抓取容器（推荐）
• 外部 —— 使用独立部署的抓取服务
• 禁用 —— 关闭网页抓取功能

Docker 镜像：https://hub.docker.com/r/vxcontrol/scraper

支持的功能：
• 通过公网地址访问外部链接
• 通过内网地址访问内部或本地链接
• 提取并分析网页内容
• 多种输出格式`

	ToolsScraperModeTitle                 = "抓取服务模式"
	ToolsScraperModeDesc                  = "选择抓取服务的运行方式"
	ToolsScraperEmbedded                  = "内置容器"
	ToolsScraperExternal                  = "外部服务"
	ToolsScraperDisabled                  = "禁用"
	ToolsScraperPublicURL                 = "公网抓取地址"
	ToolsScraperPublicURLDesc             = "抓取公网网站所用的服务地址；留空时使用内网抓取地址"
	ToolsScraperPublicURLEmbeddedDesc     = "内置抓取服务的可选自定义地址；留空时使用内网抓取地址"
	ToolsScraperPrivateURL                = "内网抓取地址"
	ToolsScraperPrivateURLDesc            = "抓取内网网站所用的服务地址"
	ToolsScraperPublicUsername            = "公网地址用户名"
	ToolsScraperPublicUsernameDesc        = "访问公网抓取服务所用的用户名"
	ToolsScraperPublicPassword            = "公网地址密码"
	ToolsScraperPublicPasswordDesc        = "访问公网抓取服务所用的密码"
	ToolsScraperPrivateUsername           = "内网地址用户名"
	ToolsScraperPrivateUsernameDesc       = "访问内网抓取服务所用的用户名"
	ToolsScraperPrivatePassword           = "内网地址密码"
	ToolsScraperPrivatePasswordDesc       = "访问内网抓取服务所用的密码"
	ToolsScraperLocalUsername             = "本地地址用户名"
	ToolsScraperLocalUsernameDesc         = "内置抓取服务所用的用户名"
	ToolsScraperLocalPassword             = "本地地址密码"
	ToolsScraperLocalPasswordDesc         = "内置抓取服务所用的密码"
	ToolsScraperMaxConcurrentSessions     = "最大并发会话数"
	ToolsScraperMaxConcurrentSessionsDesc = "抓取会话的最大并发数量"
	ToolsScraperEmbeddedHelp              = "内置模式会运行一个本地抓取容器，可同时访问公网与内网资源。默认配置使用 https://someuser:somepass@scraper/。"
	ToolsScraperExternalHelp              = "外部模式使用独立的抓取服务。可按需为公网访问和内网访问分别配置不同的地址。"
	ToolsScraperDisabledHelp              = "抓取服务已禁用。网页内容提取与分析功能将不可用。"
)

// Docker Environment screen strings
const (
	ToolsDockerFormTitle       = "Docker 环境配置"
	ToolsDockerFormDescription = "配置工作容器所使用的 Docker 环境"
	ToolsDockerFormName        = "Docker 环境"
	ToolsDockerFormOverview    = `• 工作容器隔离 —— 容器为任务提供安全边界
• 网络能力 —— 为渗透测试启用特权网络操作
• 容器管理 —— 控制工作容器如何访问 Docker 守护进程
• 存储配置 —— 定义工作区与产物的存储位置
• 镜像选择 —— 为不同类型的任务设置默认镜像

对于需要网络扫描、自定义工具和安全任务隔离的渗透测试流程，这些配置至关重要。`

	// General help text
	ToolsDockerGeneralHelp = `每个 AI 智能体任务都运行在独立的 Docker 容器中，每个任务流会自动分配两个端口（28000-32000 范围）。工作容器按需创建，使用默认镜像或由智能体选定的镜像。

基础配置需要启用相应能力：Docker 访问权限允许创建额外容器以运行专用工具；网络管理权限则授予底层网络操作权限，这是 nmap 等扫描工具所必需的。

存储默认通过 Docker 卷实现；若指定了工作目录，则改用主机目录。连接设置决定 Docker 守护进程的位置 —— 标准部署使用本地套接字，分布式环境则可使用带 TLS 的远程 TCP 连接。

默认镜像作为回退选项：通用任务使用标准镜像，安全测试则默认使用渗透测试专用容器。公网 IP 通过为工作容器提供可达地址来支持反弹 shell 攻击，便于目标回连。通常填写运行工作容器的 Docker 守护进程所在主机的本地接口地址。

可根据场景组合配置：完整渗透测试需同时启用两项能力；需要保留产物文件时使用工作目录；面向隔离的 Docker 环境时配置远程连接。`

	// Container capabilities
	ToolsDockerInside       = "Docker 访问权限"
	ToolsDockerInsideDesc   = "允许工作容器管理 Docker 容器"
	ToolsDockerNetAdmin     = "网络管理权限"
	ToolsDockerNetAdminDesc = "授予 NET_ADMIN 能力，供 nmap 等网络扫描工具使用"

	// Connection settings
	ToolsDockerSocket       = "Docker 套接字"
	ToolsDockerSocketDesc   = "主机文件系统上 Docker 套接字的路径"
	ToolsDockerNetwork      = "Docker 网络"
	ToolsDockerNetworkDesc  = "工作容器使用的自定义网络名称，或填 'host' 以直接使用主机网络"
	ToolsDockerPublicIP     = "公网 IP 地址"
	ToolsDockerPublicIPDesc = "用于带外（OOB）攻击中反向连接的公网 IP"

	// Storage configuration
	ToolsDockerWorkDir     = "工作目录"
	ToolsDockerWorkDirDesc = "工作容器文件系统所在的主机目录（默认使用 Docker 卷）"

	// Default images
	ToolsDockerDefaultImage               = "默认镜像"
	ToolsDockerDefaultImageDesc           = "通用任务使用的默认 Docker 镜像"
	ToolsDockerDefaultImageForPentest     = "渗透测试镜像"
	ToolsDockerDefaultImageForPentestDesc = "安全测试任务使用的默认 Docker 镜像"

	// TLS connection settings (optional)
	ToolsDockerHost          = "Docker 主机"
	ToolsDockerHostDesc      = "Docker 守护进程连接地址（unix:// 或 tcp://）"
	ToolsDockerTLSVerify     = "TLS 验证"
	ToolsDockerTLSVerifyDesc = "为 Docker 连接启用 TLS 验证"
	ToolsDockerCertPath      = "TLS 证书"
	ToolsDockerCertPathDesc  = "包含 ca.pem、cert.pem、key.pem 文件的目录"

	// Help content for specific configurations
	ToolsDockerInsideHelp = `Docker 访问权限允许工作容器为专用工具或运行环境创建额外容器。当任务需要默认镜像中未包含的软件时，必须启用此项。

启用后，工作容器可以拉取并运行任意 Docker 镜像，以满足复杂测试场景的需要。`

	ToolsDockerNetAdminHelp = `网络管理权限允许工作容器执行渗透测试所必需的底层网络操作。

以下场景需要该权限：
• 使用 nmap、masscan 进行网络扫描
• 自定义数据包构造
• 网络接口操作
• 原始套接字操作

对全面的安全评估至关重要。`

	ToolsDockerSocketHelp = `Docker 套接字路径决定工作容器如何访问 Docker 守护进程。此处只填套接字文件的路径。需与「Docker 访问权限」配合使用才能管理容器。

为提升安全性，建议使用 docker-in-docker（DinD），而不是把主 Docker 守护进程直接暴露给工作容器。
使用 DinD 时，请填写 DinD 容器中已绑定到主机文件系统的 Docker 套接字文件路径。

示例：/var/run/docker.sock`

	ToolsDockerNetworkHelp = `Docker 网络控制工作容器的网络隔离模式：

桥接模式（自定义网络名）：
• 容器之间隔离通信
• 从容器到主机的端口转发
• 更强的安全边界
• 支持基于网络的监控与过滤
• 推荐用于大多数场景

主机模式（值为 'host'）：
• 直接访问主机网络接口
• 无端口转发，端口直接绑定到主机
• 原始数据包操作所必需
• 支持高级网络测试能力
• 隔离性较低，请谨慎使用

示例：
• 'pentagi-network' —— 创建隔离的桥接网络
• 'host' —— 启用直接的主机网络访问

安全提示：主机网络模式会降低容器隔离性。仅在高级渗透测试任务确实需要直接访问网络栈时使用。`

	ToolsDockerPublicIPHelp = `公网 IP 地址用于启用带外（OOB）攻击技术，为工作容器提供可供反向连接的可达地址。

工作容器会自动获得两个映射到该 IP 的随机端口（28000-32000 范围），用于接收来自被利用目标的回连。

默认情况下，智能体会尝试通过 api.ipify.org、ipinfo.io/ip 或 ifconfig.me 服务获取公网地址。`

	ToolsDockerWorkDirHelp = `工作目录指定工作容器存储所在的主机文件系统位置。设置后将以主机目录挂载替代默认的 Docker 卷。

优势：
• 重启后数据仍然保留
• 可直接访问文件系统
• 更便于管理产物文件
• 可自定义备份策略

默认情况下为每个工作容器使用独立的 Docker 卷。

示例：/path/to/workdir/`

	ToolsDockerDefaultImageHelp = `当任务需求未指定特定容器镜像时，默认镜像作为工作容器的回退选项。

应包含通用任务所需的基础工具与实用程序。默认值：debian:latest`

	ToolsDockerDefaultImageForPentestHelp = `渗透测试镜像作为安全测试任务的默认镜像，应包含完备的安全工具与实用程序。

推荐使用 Kali Linux、Parrot Security 或自定义的安全专用容器。默认值：vxcontrol/kali-linux`

	ToolsDockerHostHelp = `Docker 主机地址用于启动主工作容器，并覆盖默认的 Docker 守护进程连接。支持 Unix 套接字和 TCP 连接。

示例：
• unix:///var/run/docker.sock（本地）
• tcp://docker-host:2376（远程）

远程连接请启用 TLS。`

	ToolsDockerTLSVerifyHelp = `TLS 验证用于保护通过 TCP 建立的 Docker 守护进程连接。强烈建议对远程 Docker 主机启用。

需要在指定的证书目录中提供有效证书。`

	ToolsDockerCertPathHelp = `TLS 证书目录必须包含：
• ca.pem —— 证书颁发机构证书
• cert.pem —— 客户端证书
• key.pem —— 私钥

使用 TLS 管理工作容器时，安全的远程 Docker 连接需要这些文件。

示例：/path/to/certs`
)

// Embedder form strings
const (
	EmbedderFormTitle       = "嵌入模型配置"
	EmbedderFormDescription = "配置用于语义搜索和知识存储的文本向量化"
	EmbedderFormName        = "嵌入模型"
	EmbedderFormOverview    = `文本嵌入将文档转换为向量，用于语义搜索和知识存储。
不同提供商提供的模型在能力和价格上各有差异。

请谨慎选择：更换提供商需要对所有已存储数据重新建立索引。`

	EmbedderFormProvider     = "嵌入提供商"
	EmbedderFormProviderDesc = "选择用于文本向量化的提供商。嵌入用于语义搜索和知识存储。"

	EmbedderFormURL     = "API 端点地址"
	EmbedderFormURLDesc = "自定义 API 端点（留空则使用默认值）"

	EmbedderFormAPIKey     = "API 密钥"
	EmbedderFormAPIKeyDesc = "提供商的身份验证密钥（Ollama 无需填写）"

	EmbedderFormModel     = "模型名称"
	EmbedderFormModelDesc = "要使用的具体嵌入模型（留空则使用提供商默认值）"

	EmbedderFormBatchSize     = "批处理大小"
	EmbedderFormBatchSizeDesc = "单批处理的文档数量（1-1000）"

	EmbedderFormStripNewLines     = "去除换行符"
	EmbedderFormStripNewLinesDesc = "嵌入前移除文本中的换行符（true/false）"

	EmbedderFormMaxTextBytes     = "文本字节上限"
	EmbedderFormMaxTextBytesDesc = "发送到嵌入 API 的每个文本块的最大字节数（例如 8192）"

	EmbedderFormHelpTitle   = "嵌入配置"
	EmbedderFormHelpContent = `配置用于语义搜索和知识存储的文本向量化。

若未配置具体的嵌入设置，系统将使用 OpenAI 嵌入，并采用 LLM 提供商中配置的 API 密钥。

更换提供商需谨慎：不同嵌入模型生成的向量互不兼容，需要重建数据库索引。`

	EmbedderFormHelpOpenAI      = "OpenAI：最可靠的选择，质量优异。若此处未填写，将使用 LLM 提供商中的 API 密钥。"
	EmbedderFormHelpOllama      = "Ollama：本地嵌入，无需 API 密钥。需要运行 Ollama 服务器。"
	EmbedderFormHelpHuggingFace = "HuggingFace：开源模型，需要 API 密钥。"
	EmbedderFormHelpGoogleAI    = "Google AI：高质量嵌入，需要 API 密钥。"

	// Provider names and descriptions
	EmbedderProviderDefault         = "默认（OpenAI）"
	EmbedderProviderDefaultDesc     = "使用 OpenAI 嵌入，并采用 LLM 提供商配置中的 API 密钥"
	EmbedderProviderOpenAI          = "OpenAI"
	EmbedderProviderOpenAIDesc      = "OpenAI 文本嵌入 API（text-embedding-3-small、ada-002）"
	EmbedderProviderOllama          = "Ollama"
	EmbedderProviderOllamaDesc      = "本地 Ollama 服务器，运行开源嵌入模型"
	EmbedderProviderMistral         = "Mistral"
	EmbedderProviderMistralDesc     = "Mistral AI 嵌入模型"
	EmbedderProviderJina            = "Jina"
	EmbedderProviderJinaDesc        = "Jina AI 嵌入 API"
	EmbedderProviderHuggingFace     = "HuggingFace"
	EmbedderProviderHuggingFaceDesc = "HuggingFace 推理 API，用于嵌入模型"
	EmbedderProviderGoogleAI        = "Google AI"
	EmbedderProviderGoogleAIDesc    = "Google AI 嵌入模型（embedding-001）"
	EmbedderProviderVoyageAI        = "VoyageAI"
	EmbedderProviderVoyageAIDesc    = "VoyageAI 嵌入 API"
	EmbedderProviderDisabled        = "已禁用"
	EmbedderProviderDisabledDesc    = "完全禁用嵌入功能"

	// Provider-specific placeholders and help
	EmbedderURLPlaceholderOpenAI      = "https://api.openai.com/v1"
	EmbedderURLPlaceholderOllama      = "http://localhost:11434"
	EmbedderURLPlaceholderMistral     = "https://api.mistral.ai/v1"
	EmbedderURLPlaceholderJina        = "https://api.jina.ai/v1"
	EmbedderURLPlaceholderHuggingFace = "https://api-inference.huggingface.co"
	EmbedderURLPlaceholderGoogleAI    = "不支持自定义，使用默认端点"
	EmbedderURLPlaceholderVoyageAI    = "不支持自定义，使用默认端点"

	EmbedderAPIKeyPlaceholderOllama      = "本地模型无需填写"
	EmbedderAPIKeyPlaceholderMistral     = "Mistral API 密钥"
	EmbedderAPIKeyPlaceholderJina        = "Jina API 密钥"
	EmbedderAPIKeyPlaceholderHuggingFace = "HuggingFace API 密钥"
	EmbedderAPIKeyPlaceholderGoogleAI    = "Google AI API 密钥"
	EmbedderAPIKeyPlaceholderVoyageAI    = "VoyageAI API 密钥"
	EmbedderAPIKeyPlaceholderDefault     = "该提供商的 API 密钥"

	EmbedderModelPlaceholderOpenAI      = "text-embedding-3-small"
	EmbedderModelPlaceholderOllama      = "nomic-embed-text"
	EmbedderModelPlaceholderMistral     = "mistral-embed"
	EmbedderModelPlaceholderJina        = "jina-embeddings-v2-base-en"
	EmbedderModelPlaceholderHuggingFace = "sentence-transformers/all-MiniLM-L6-v2"
	EmbedderModelPlaceholderGoogleAI    = "gemini-embedding-001"
	EmbedderModelPlaceholderVoyageAI    = "voyage-2"
	EmbedderModelPlaceholderDefault     = "模型名称"

	// Provider IDs for internal use
	EmbedderProviderIDDefault     = "default"
	EmbedderProviderIDOpenAI      = "openai"
	EmbedderProviderIDOllama      = "ollama"
	EmbedderProviderIDMistral     = "mistral"
	EmbedderProviderIDJina        = "jina"
	EmbedderProviderIDHuggingFace = "huggingface"
	EmbedderProviderIDGoogleAI    = "googleai"
	EmbedderProviderIDVoyageAI    = "voyageai"
	EmbedderProviderIDDisabled    = "none"

	EmbedderHelpGeneral = `嵌入会将文本转换为向量，用于语义检索与知识存储。这让 PentAGI 能够理解语义而不只是匹配关键词，使检索结果更贴切、更智能。

主要收益：
• 按语义查找文档，而非精确词匹配
• 基于渗透测试结果构建智能知识库
• 让 AI 智能体快速定位相关信息
• 借助上下文数据支撑更深入的推理

若希望完全本地处理，请选择 Ollama —— 数据不会离开你的基础设施。其他提供商提供云端处理，模型能力与价格各不相同。

请谨慎配置，更换提供商需要重建整个知识库。`

	EmbedderHelpAttentionPrefix = "重要提示："
	EmbedderHelpAttention       = `不同嵌入提供商生成的向量互不兼容。更换提供商或模型会破坏现有的语义检索。

你必须使用 etester 工具清空或重建整个知识库：
• 执行 'etester flush' 清除旧的嵌入数据
• 执行 'etester reindex' 使用新提供商重建索引
• 数据量较大时，该过程可能耗时较长`

	EmbedderHelpAttentionSuffix = `除非确有必要，否则不要更换提供商。`

	// Provider help texts
	EmbedderHelpDefault = `默认模式使用 OpenAI 嵌入，并采用在 LLM 提供商中配置的 API 密钥。

若你已配置好 OpenAI，这是大多数用户的推荐选项，无需额外配置。`

	EmbedderHelpOpenAI = `直接通过 OpenAI API 生成嵌入。

在此获取 API 密钥：
https://platform.openai.com/api-keys

推荐模型：
• text-embedding-3-small（性价比高，1536 维）
• text-embedding-3-large（质量最高，3072 维）
• text-embedding-ada-002（旧版，仍受支持）`

	EmbedderHelpOllama = `使用本地 Ollama 服务器运行开源嵌入模型。

常用嵌入模型：
• nomic-embed-text（推荐，768 维）
• mxbai-embed-large（大模型，1024 维）
• snowflake-arctic-embed（支持多语言）

在此安装 Ollama：
https://ollama.com/

从这条命令开始：ollama pull nomic-embed-text`

	EmbedderHelpMistral = `通过 API 使用 Mistral AI 嵌入模型。

在此获取 API 密钥：
https://console.mistral.ai/

使用 Mistral 的嵌入模型，配置固定。
无需选择模型 —— 直接使用默认嵌入模型。`

	EmbedderHelpJina = `Jina AI 嵌入 API，提供多种专用模型。

在此获取 API 密钥：
https://jina.ai/

推荐模型：
• jina-embeddings-v2-base-en（通用，768 维）
• jina-embeddings-v2-small-en（轻量，512 维）
• jina-embeddings-v2-base-code（代码专用嵌入）`

	EmbedderHelpHuggingFace = `HuggingFace 推理 API，可使用开源模型。

在此获取 API 密钥：
https://huggingface.co/settings/tokens

常用模型：
• sentence-transformers/all-MiniLM-L6-v2（384 维）
• sentence-transformers/all-mpnet-base-v2（768 维）
• intfloat/e5-large-v2（1024 维）`

	EmbedderHelpGoogleAI = `Google AI 嵌入模型（Gemini）。

在此获取 API 密钥：
https://aistudio.google.com/app/apikey

可用模型：
• gemini-embedding-001（最新模型，768 维）
• text-embedding-004（旧版 Vertex AI 模型）

使用 Google 固定的接口地址 —— 不支持自定义 URL。`

	EmbedderHelpVoyageAI = `VoyageAI 嵌入 API，针对检索场景做了优化。

在此获取 API 密钥：
https://www.voyageai.com/

推荐模型：
• voyage-2（通用，1024 维）
• voyage-large-2（质量最高，1536 维）
• voyage-code-2（代码嵌入，1536 维）`

	EmbedderHelpDisabled = `禁用全部嵌入功能。

这会：
• 关闭语义检索能力
• 停止知识存储的向量化
• 降低内存与计算资源需求

仅当你的使用场景不需要嵌入时才建议选择。`
)

// Development and Mock Screen constants
const (
	MockScreenTitle                  = "开发中界面"
	MockScreenDescription            = "此界面尚在开发中"
	MockScreenUnderDevelopment       = "🚧 此界面正在开发中"
	MockScreenAvailableLater         = "该配置界面将在后续更新中提供。"
	MockScreenBackInstruction        = "按 Enter 或 Esc 返回主菜单。"
	MockScreenMigrationPending       = "⏳ 配置界面正在迁移"
	MockScreenDevelopmentNotice      = "开发提示"
	MockScreenMigrationDescription   = "当前正在将此配置界面迁移到新版界面。"
	MockScreenExpectedFeatures       = "预计提供："
	MockScreenModernForm             = "• 现代化表单界面"
	MockScreenImprovedValidation     = "• 更完善的输入校验"
	MockScreenEnhancedUserExperience = "• 更流畅的操作体验"
	MockScreenCheckLater             = "请关注后续更新。"
)

// Apply Changes screen constants
const (
	ApplyChangesFormTitle       = "应用配置更改"
	ApplyChangesFormName        = "应用更改"
	ApplyChangesFormDescription = "检查并应用你的配置更改"

	// Apply Changes overview and help
	ApplyChangesFormOverview = `在此界面可以查看所有待应用的配置更改，并将它们应用到你的 PentAGI 安装中。

应用更改时，系统将：
• 把所有修改过的环境变量保存到 .env 文件
• 使用新配置重启受影响的服务
• 按需安装额外组件`

	// Apply Changes status messages
	ApplyChangesNotStarted     = "配置更改已准备就绪，可以应用"
	ApplyChangesInProgress     = "正在应用配置更改...\n"
	ApplyChangesCompleted      = "配置更改已成功应用\n"
	ApplyChangesFailed         = "配置更改执行失败"
	ApplyChangesResetCompleted = "配置更改已成功重置\n"

	ApplyChangesTerminalIsNotInitialized = "终端尚未初始化"

	// Apply Changes instructions
	ApplyChangesInstructions = `按 Enter 开始应用配置更改。`

	ApplyChangesNoChanges = "没有待应用的配置更改"

	// Apply Changes installation status
	ApplyChangesInstallNotFound = `本系统当前尚未安装 PentAGI。

将执行以下操作：
• 设置并校验 Docker 环境
• 创建 docker-compose.yml 文件
• 安装并启动 PentAGI 核心服务`

	ApplyChangesInstallFoundLangfuse      = `• 安装 Langfuse 可观测性组件栈（docker-compose-langfuse.yml）`
	ApplyChangesInstallFoundObservability = `• 安装包含 Grafana、VictoriaMetrics 和 Jaeger 的完整可观测性组件栈（docker-compose-observability.yml）`

	ApplyChangesUpdateFound = `本系统当前已安装 PentAGI。

将执行以下操作：
• 更新 .env 文件中的环境变量
• 重建并重启受影响的 Docker 容器
• 将新配置应用到正在运行的服务`

	// Apply Changes warnings and notes
	ApplyChangesWarningCritical = "⚠️  检测到关键更改 —— 服务将被重启"
	ApplyChangesWarningSecrets  = "🔒 检测到敏感信息 —— 将以安全方式存储"
	ApplyChangesNoteBackup      = "💾 更改前会先备份当前配置"
	ApplyChangesNoteTime        = "⏱️  视所选组件而定，此过程通常不超过一分钟"

	// Apply Changes progress messages
	ApplyChangesStageValidation = "正在校验环境与依赖..."
	ApplyChangesStageBackup     = "正在创建配置备份..."
	ApplyChangesStageEnvFile    = "正在更新环境配置文件..."
	ApplyChangesStageCompose    = "正在生成 Docker Compose 文件..."
	ApplyChangesStageDocker     = "正在管理 Docker 容器..."
	ApplyChangesStageServices   = "正在启动服务..."
	ApplyChangesStageComplete   = "配置更改已成功应用"

	// Apply Changes change list headers
	ApplyChangesChangesTitle  = "待应用的配置更改"
	ApplyChangesChangesCount  = "更改总数：%d"
	ApplyChangesChangesMasked = "（出于安全已隐藏）"
	ApplyChangesChangesEmpty  = "没有需要应用的更改"

	// Apply Changes help content
	ApplyChangesHelpTitle   = "应用配置更改"
	ApplyChangesHelpContent = `应用更改前请务必确认当前配置。`
)

// apply changes integrity prompt
const (
	ApplyChangesIntegrityPromptTitle   = "文件完整性检查"
	ApplyChangesIntegrityPromptMessage = "检测到过期文件。\n是否将它们更新到最新版本？"
	ApplyChangesIntegrityOutdatedList  = "过期文件：\n%s\n确认更新？(y/n)"
	ApplyChangesIntegrityChecking      = "正在收集文件完整性信息..."
	ApplyChangesIntegrityNoOutdated    = "未发现过期文件，继续应用更改。"
)

// Maintenance Screen constants
const (
	MaintenanceTitle       = "系统维护"
	MaintenanceDescription = "管理 PentAGI 服务并执行维护操作"
	MaintenanceName        = "维护"
	MaintenanceOverview    = `对 PentAGI 执行系统维护操作。

可用操作取决于当前系统状态，只有适用时才会显示。

包含的操作：
• 服务生命周期管理（启动/停止/重启）
• 组件更新与下载
• 系统重置与清理
• 容器与镜像管理

每项操作都会实时反馈状态，并在需要时请求确认。`

	// Maintenance menu items
	MaintenanceStartPentagi            = "启动 PentAGI"
	MaintenanceStartPentagiDesc        = "启动所有已配置的 PentAGI 服务"
	MaintenanceStopPentagi             = "停止 PentAGI"
	MaintenanceStopPentagiDesc         = "停止所有正在运行的 PentAGI 服务"
	MaintenanceRestartPentagi          = "重启 PentAGI"
	MaintenanceRestartPentagiDesc      = "重启所有 PentAGI 服务"
	MaintenanceDownloadWorkerImage     = "下载工作容器镜像"
	MaintenanceDownloadWorkerImageDesc = "下载用于工作任务的渗透测试容器镜像"
	MaintenanceUpdateWorkerImage       = "更新工作容器镜像"
	MaintenanceUpdateWorkerImageDesc   = "将渗透测试容器镜像更新到最新版本"
	MaintenanceUpdatePentagi           = "更新 PentAGI"
	MaintenanceUpdatePentagiDesc       = "将 PentAGI 更新到最新版本"
	MaintenanceUpdateInstaller         = "更新安装程序"
	MaintenanceUpdateInstallerDesc     = "将本安装程序更新到最新版本"
	MaintenanceFactoryReset            = "恢复出厂设置"
	MaintenanceFactoryResetDesc        = "将 PentAGI 重置为出厂默认状态"
	MaintenanceRemovePentagi           = "移除 PentAGI"
	MaintenanceRemovePentagiDesc       = "移除 PentAGI 容器但保留数据"
	MaintenancePurgePentagi            = "彻底清除 PentAGI"
	MaintenancePurgePentagiDesc        = "完全移除 PentAGI，包括所有数据"
	MaintenanceResetPassword           = "重置管理员密码"
	MaintenanceResetPasswordDesc       = "重置 PentAGI 的管理员密码"
)

// Reset Password Screen constants
const (
	ResetPasswordFormTitle       = "重置管理员密码"
	ResetPasswordFormDescription = "重置 PentAGI 的管理员密码"
	ResetPasswordFormName        = "重置密码"
	ResetPasswordFormOverview    = `重置默认管理员账号（admin@pentagi.com）的密码。

此操作要求 PentAGI 处于运行状态，并将在 PostgreSQL 数据库中更新密码。

请两次输入新密码以确认，然后按 Enter 应用更改。

密码要求：
• 至少 5 个字符
• 两次输入的密码必须一致`

	// Form fields
	ResetPasswordNewPassword         = "新密码"
	ResetPasswordNewPasswordDesc     = "输入新的管理员密码"
	ResetPasswordConfirmPassword     = "确认密码"
	ResetPasswordConfirmPasswordDesc = "再次输入新密码以确认"

	// Status messages
	ResetPasswordNotAvailable = "重置密码需要 PentAGI 处于运行状态"
	ResetPasswordAvailable    = "可以重置密码"
	ResetPasswordInProgress   = "正在重置密码..."
	ResetPasswordSuccess      = "密码已成功重置"
	ResetPasswordErrorPrefix  = "错误："

	// Validation errors
	ResetPasswordErrorEmptyPassword = "密码不能为空"
	ResetPasswordErrorShortPassword = "密码长度至少为 5 个字符"
	ResetPasswordErrorMismatch      = "两次输入的密码不一致"

	// Help content
	ResetPasswordHelpContent = `重置用于登录 PentAGI 的管理员密码。

此操作会：
• 更新 admin@pentagi.com 账号的密码
• 将该用户状态设为 'active'
• 需要能够访问 PentAGI 数据库
• 不影响其他用户账号

操作成功完成后，密码更改立即生效。

在两个字段中输入相同的密码，然后按 Enter 确认更改。`
)

// Processor Operation Form constants
const (
	// Dynamic title templates
	ProcessorOperationFormTitle       = "%s"
	ProcessorOperationFormDescription = "执行 %s 操作"
	ProcessorOperationFormName        = "%s"

	// Common status messages
	ProcessorOperationNotStarted = "已就绪，可执行 %s 操作"
	ProcessorOperationInProgress = "正在执行 %s 操作...\n"
	ProcessorOperationCompleted  = "%s 操作已成功完成\n"
	ProcessorOperationFailed     = "执行 %s 操作失败"

	// Confirmation messages
	ProcessorOperationConfirmation = "确定要%s吗？"
	ProcessorOperationPressEnter   = "按 Enter 即可%s"
	ProcessorOperationPressYN      = "按 Y 确认，按 N 取消"
	// Short notice without hotkeys (for static help panel)
	ProcessorOperationRequiresConfirmationShort = "此操作需要确认"
	// Additional terminal messages
	ProcessorOperationCancelled = "操作已取消"
	ProcessorOperationUnknown   = "未知操作：%s"

	// Operation specific messages
	ProcessorOperationStarting    = "正在启动服务..."
	ProcessorOperationStopping    = "正在停止服务..."
	ProcessorOperationRestarting  = "正在重启服务..."
	ProcessorOperationDownloading = "正在下载镜像..."
	ProcessorOperationUpdating    = "正在更新组件..."
	ProcessorOperationResetting   = "正在恢复出厂默认设置..."
	ProcessorOperationRemoving    = "正在移除容器..."
	ProcessorOperationPurging     = "正在清除所有数据..."
	ProcessorOperationInstalling  = "正在安装 PentAGI 服务..."

	// Help text templates
	ProcessorOperationHelpTitle           = "%s 操作"
	ProcessorOperationHelpContent         = "此操作将%s。"
	ProcessorOperationHelpContentDownload = "此操作将下载 %s 组件。"
	ProcessorOperationHelpContentUpdate   = "此操作将更新 %s 组件。"
	// Generic title/description/builders for dynamic operations
	OperationTitleInstallPentagi    = "安装 PentAGI"
	OperationDescInstallPentagi     = "安装并配置 PentAGI 服务"
	OperationTitleDownload          = "下载 %s"
	OperationDescDownloadComponents = "下载 %s 组件"
	OperationTitleUpdate            = "更新 %s"
	OperationDescUpdateToLatest     = "将 %s 更新到最新版本"
	OperationTitleExecute           = "执行 %s"
	OperationDescExecuteOn          = "执行 %s（作用于 %s）"
	OperationProgressExecuting      = "正在执行 %s..."

	// Terminal not initialized
	ProcessorOperationTerminalNotInitialized = "终端未初始化"
)

// Operation-specific help texts
const (
	ProcessorHelpInstallPentagi = `此操作将：
• 为所选服务部署 Docker 容器
• 配置网络与数据卷
• 启动所有已启用的服务
• 如已配置，则设置监控组件

安装将使用你当前的配置设置。`

	ProcessorHelpStartPentagi = `此操作将启动：
• PentAGI 核心 API 与 Web 界面
• 已配置的 Langfuse 分析组件（如已启用）
• Observability 可观测性栈（如已启用）

服务将按正确的依赖顺序启动。`

	ProcessorHelpStopPentagi = `此操作将：
• 优雅地关闭容器
• 保留所有数据与配置
• 关闭网络连接

稍后可以重新启动服务，不会丢失任何数据。`

	ProcessorHelpRestartPentagi = `此操作将：
• 停止正在运行的容器
• 应用所有配置更改
• 以全新状态启动服务

在更新配置后或需要排查问题时很有用。`

	ProcessorHelpDownloadWorkerImage = `该大型镜像（6GB 以上）包含：
• Kali Linux 工具与实用程序
• 安全测试框架
• 网络分析软件

执行渗透测试操作所必需。`

	ProcessorHelpUpdateWorkerImage = `此操作将：
• 拉取最新的渗透测试镜像
• 更新安全工具与框架
• 保留现有的工作容器

注意：这是一次大体积下载（6GB 以上）。`

	ProcessorHelpUpdatePentagi = `此操作将：
• 下载最新的容器镜像
• 对服务执行滚动更新
• 保留所有数据与配置

更新期间服务会短暂不可用。`

	ProcessorHelpUpdateInstaller = `此操作将：
• 下载最新的安装程序二进制文件
• 替换当前的安装程序
• 退出，等待手动重启

更新完成后你需要重新启动安装程序。`

	ProcessorHelpFactoryReset = `⚠️  警告：此操作将：
• 移除所有容器与网络
• 删除所有配置文件
• 清除已存储的数据与数据卷
• 恢复默认设置

此操作无法撤销！`

	ProcessorHelpRemovePentagi = `此操作将：
• 停止并移除所有容器
• 移除 Docker 网络
• 保留数据卷与数据
• 保留配置文件

稍后可以重新安装，不会丢失数据。`

	ProcessorHelpPurgePentagi = `⚠️  警告：此操作将永久删除：
• 所有容器与镜像
• 所有数据卷
• 所有配置文件
• 所有已保存的结果

此操作无法撤销！`
)

// environment variable descriptions (centralized)
const (
	EnvDesc_OPEN_AI_KEY                       = "OpenAI API 密钥"
	EnvDesc_OPEN_AI_SERVER_URL                = "OpenAI 服务器地址"
	EnvDesc_ANTHROPIC_API_KEY                 = "Anthropic API 密钥"
	EnvDesc_ANTHROPIC_SERVER_URL              = "Anthropic 服务器地址"
	EnvDesc_GEMINI_API_KEY                    = "Google Gemini API 密钥"
	EnvDesc_GEMINI_SERVER_URL                 = "Gemini 服务器地址"
	EnvDesc_BEDROCK_DEFAULT_AUTH              = "AWS Bedrock 使用默认凭证链"
	EnvDesc_BEDROCK_BEARER_TOKEN              = "AWS Bedrock Bearer 令牌"
	EnvDesc_BEDROCK_ACCESS_KEY_ID             = "AWS Bedrock 访问密钥 ID"
	EnvDesc_BEDROCK_SECRET_ACCESS_KEY         = "AWS Bedrock 秘密访问密钥"
	EnvDesc_BEDROCK_SESSION_TOKEN             = "AWS Bedrock 会话令牌"
	EnvDesc_BEDROCK_REGION                    = "AWS Bedrock 区域"
	EnvDesc_BEDROCK_SERVER_URL                = "AWS Bedrock 自定义端点地址"
	EnvDesc_OLLAMA_SERVER_URL                 = "Ollama 服务器地址"
	EnvDesc_OLLAMA_SERVER_API_KEY             = "Ollama 服务器 API 密钥（云端）"
	EnvDesc_OLLAMA_SERVER_MODEL               = "Ollama 默认模型"
	EnvDesc_OLLAMA_SERVER_CONFIG_PATH         = "Ollama 容器内配置文件路径"
	EnvDesc_OLLAMA_SERVER_PULL_MODELS_TIMEOUT = "Ollama 模型下载超时"
	EnvDesc_OLLAMA_SERVER_PULL_MODELS_ENABLED = "Ollama 自动下载模型"
	EnvDesc_OLLAMA_SERVER_LOAD_MODELS_ENABLED = "Ollama 加载模型列表"
	EnvDesc_DEEPSEEK_API_KEY                  = "DeepSeek API 密钥"
	EnvDesc_DEEPSEEK_SERVER_URL               = "DeepSeek 服务器地址"
	EnvDesc_DEEPSEEK_PROVIDER                 = "DeepSeek 提供商名称前缀（用于 LiteLLM，例如 'deepseek'）"
	EnvDesc_GLM_API_KEY                       = "GLM API 密钥"
	EnvDesc_GLM_SERVER_URL                    = "GLM 服务器地址"
	EnvDesc_GLM_PROVIDER                      = "GLM 提供商名称前缀（用于 LiteLLM，例如 'zai'）"
	EnvDesc_KIMI_API_KEY                      = "Kimi API 密钥"
	EnvDesc_KIMI_SERVER_URL                   = "Kimi 服务器地址"
	EnvDesc_KIMI_PROVIDER                     = "Kimi 提供商名称前缀（用于 LiteLLM，例如 'moonshot'）"
	EnvDesc_QWEN_API_KEY                      = "Qwen API 密钥"
	EnvDesc_QWEN_SERVER_URL                   = "Qwen 服务器地址"
	EnvDesc_QWEN_PROVIDER                     = "Qwen 提供商名称前缀（用于 LiteLLM，例如 'dashscope'）"
	EnvDesc_LLM_SERVER_URL                    = "自定义 LLM 服务器地址"
	EnvDesc_LLM_SERVER_KEY                    = "自定义 LLM API 密钥"
	EnvDesc_LLM_SERVER_MODEL                  = "自定义 LLM 模型"
	EnvDesc_LLM_SERVER_CONFIG_PATH            = "自定义 LLM 容器内配置文件路径"
	EnvDesc_LLM_SERVER_LEGACY_REASONING       = "自定义 LLM 旧版推理模式"
	EnvDesc_LLM_SERVER_PRESERVE_REASONING     = "自定义 LLM 保留推理内容"
	EnvDesc_LLM_SERVER_PROVIDER               = "自定义 LLM 提供商名称"

	EnvDesc_LANGFUSE_LISTEN_IP   = "Langfuse 监听 IP"
	EnvDesc_LANGFUSE_LISTEN_PORT = "Langfuse 监听端口"
	EnvDesc_LANGFUSE_BASE_URL    = "Langfuse 基础地址"
	EnvDesc_LANGFUSE_PROJECT_ID  = "Langfuse 项目 ID"
	EnvDesc_LANGFUSE_PUBLIC_KEY  = "Langfuse 公开密钥"
	EnvDesc_LANGFUSE_SECRET_KEY  = "Langfuse 私密密钥"

	// langfuse init variables
	EnvDesc_LANGFUSE_INIT_PROJECT_ID         = "Langfuse 初始化项目 ID"
	EnvDesc_LANGFUSE_INIT_PROJECT_PUBLIC_KEY = "Langfuse 初始化项目公开密钥"
	EnvDesc_LANGFUSE_INIT_PROJECT_SECRET_KEY = "Langfuse 初始化项目私密密钥"
	EnvDesc_LANGFUSE_INIT_USER_EMAIL         = "Langfuse 初始化用户邮箱"
	EnvDesc_LANGFUSE_INIT_USER_NAME          = "Langfuse 初始化用户名"
	EnvDesc_LANGFUSE_INIT_USER_PASSWORD      = "Langfuse 初始化用户密码"

	EnvDesc_LANGFUSE_OTEL_EXPORTER_OTLP_ENDPOINT = "Langfuse 的 OpenTelemetry 导出器 OTLP 端点"

	EnvDesc_GRAFANA_LISTEN_IP     = "Grafana 监听 IP"
	EnvDesc_GRAFANA_LISTEN_PORT   = "Grafana 监听端口"
	EnvDesc_OTEL_GRPC_LISTEN_IP   = "OTel gRPC 监听 IP"
	EnvDesc_OTEL_GRPC_LISTEN_PORT = "OTel gRPC 监听端口"
	EnvDesc_OTEL_HTTP_LISTEN_IP   = "OTel HTTP 监听 IP"
	EnvDesc_OTEL_HTTP_LISTEN_PORT = "OTel HTTP 监听端口"
	EnvDesc_OTEL_HOST             = "OpenTelemetry 主机地址"

	EnvDesc_SUMMARIZER_PRESERVE_LAST       = "摘要器：末段大小管理"
	EnvDesc_SUMMARIZER_USE_QA              = "摘要器：启用问答摘要"
	EnvDesc_SUMMARIZER_SUM_MSG_HUMAN_IN_QA = "摘要器：压缩问答中的用户消息"
	EnvDesc_SUMMARIZER_LAST_SEC_BYTES      = "摘要器：末段字节上限"
	EnvDesc_SUMMARIZER_MAX_BP_BYTES        = "摘要器：单条回复字节上限"
	EnvDesc_SUMMARIZER_MAX_QA_BYTES        = "摘要器：问答总字节上限"
	EnvDesc_SUMMARIZER_MAX_QA_SECTIONS     = "摘要器：问答段落数上限"
	EnvDesc_SUMMARIZER_KEEP_QA_SECTIONS    = "摘要器：保留的最近段落数"

	EnvDesc_ASSISTANT_SUMMARIZER_PRESERVE_LAST    = "助手摘要器：末段大小管理"
	EnvDesc_ASSISTANT_SUMMARIZER_LAST_SEC_BYTES   = "助手摘要器：末段字节上限"
	EnvDesc_ASSISTANT_SUMMARIZER_MAX_BP_BYTES     = "助手摘要器：单条回复字节上限"
	EnvDesc_ASSISTANT_SUMMARIZER_MAX_QA_BYTES     = "助手摘要器：问答总字节上限"
	EnvDesc_ASSISTANT_SUMMARIZER_MAX_QA_SECTIONS  = "助手摘要器：问答段落数上限"
	EnvDesc_ASSISTANT_SUMMARIZER_KEEP_QA_SECTIONS = "助手摘要器：保留的最近段落数"

	EnvDesc_EMBEDDING_PROVIDER        = "向量化提供商"
	EnvDesc_EMBEDDING_URL             = "向量化服务地址"
	EnvDesc_EMBEDDING_KEY             = "向量化 API 密钥"
	EnvDesc_EMBEDDING_MODEL           = "向量化模型"
	EnvDesc_EMBEDDING_BATCH_SIZE      = "向量化批处理大小"
	EnvDesc_EMBEDDING_STRIP_NEW_LINES = "向量化前移除换行符"
	EnvDesc_EMBEDDING_MAX_TEXT_BYTES  = "向量化单块文本字节上限"

	EnvDesc_ASK_USER = "人工介入（Human-in-the-loop）"

	EnvDesc_ASSISTANT_USE_AGENTS = "为助手启用多智能体模式"

	EnvDesc_EXECUTION_MONITOR_ENABLED          = "启用执行监控（beta）"
	EnvDesc_EXECUTION_MONITOR_SAME_TOOL_LIMIT  = "相同工具调用阈值"
	EnvDesc_EXECUTION_MONITOR_TOTAL_TOOL_LIMIT = "工具调用总数阈值"
	EnvDesc_MAX_GENERAL_AGENT_TOOL_CALLS       = "通用智能体的工具调用上限"
	EnvDesc_MAX_LIMITED_AGENT_TOOL_CALLS       = "受限智能体的工具调用上限"
	EnvDesc_AGENT_PLANNING_STEP_ENABLED        = "启用任务规划（beta）"

	EnvDesc_SCRAPER_PUBLIC_URL                    = "Scraper 公网地址"
	EnvDesc_SCRAPER_PRIVATE_URL                   = "Scraper 内网地址"
	EnvDesc_LOCAL_SCRAPER_USERNAME                = "本地 Scraper 用户名"
	EnvDesc_LOCAL_SCRAPER_PASSWORD                = "本地 Scraper 密码"
	EnvDesc_LOCAL_SCRAPER_MAX_CONCURRENT_SESSIONS = "Scraper 最大并发会话数"

	EnvDesc_DUCKDUCKGO_ENABLED    = "DuckDuckGo 搜索"
	EnvDesc_DUCKDUCKGO_REGION     = "DuckDuckGo 地区"
	EnvDesc_DUCKDUCKGO_SAFESEARCH = "DuckDuckGo 安全搜索"
	EnvDesc_DUCKDUCKGO_TIME_RANGE = "DuckDuckGo 时间范围"
	EnvDesc_SPLOITUS_ENABLED      = "Sploitus 搜索"
	EnvDesc_PERPLEXITY_API_KEY    = "Perplexity API 密钥"
	EnvDesc_TAVILY_API_KEY        = "Tavily API 密钥"
	EnvDesc_TRAVERSAAL_API_KEY    = "Traversaal API 密钥"
	EnvDesc_GOOGLE_API_KEY        = "Google 搜索 API 密钥"
	EnvDesc_GOOGLE_CX_KEY         = "Google 搜索引擎 ID（CX）"
	EnvDesc_GOOGLE_LR_KEY         = "Google 搜索语言限制（LR）"

	EnvDesc_DOCKER_INSIDE                    = "允许工作容器管理 Docker"
	EnvDesc_DOCKER_NET_ADMIN                 = "Docker 网络管理权限"
	EnvDesc_DOCKER_SOCKET                    = "Docker 套接字路径"
	EnvDesc_DOCKER_NETWORK                   = "Docker 网络"
	EnvDesc_DOCKER_PUBLIC_IP                 = "Docker 公网 IP"
	EnvDesc_DOCKER_WORK_DIR                  = "Docker 工作目录"
	EnvDesc_DOCKER_DEFAULT_IMAGE             = "Docker 默认镜像"
	EnvDesc_DOCKER_DEFAULT_IMAGE_FOR_PENTEST = "Docker 渗透测试镜像"
	EnvDesc_DOCKER_HOST                      = "Docker 主机地址"
	EnvDesc_DOCKER_TLS_VERIFY                = "Docker TLS 验证"
	EnvDesc_DOCKER_CERT_PATH                 = "Docker 证书路径"

	EnvDesc_LICENSE_KEY                       = "PentAGI 许可证密钥"
	EnvDesc_PENTAGI_LISTEN_IP                 = "PentAGI 服务监听 IP"
	EnvDesc_PENTAGI_LISTEN_PORT               = "PentAGI 服务监听端口"
	EnvDesc_PUBLIC_URL                        = "PentAGI 公开访问地址"
	EnvDesc_CORS_ORIGINS                      = "PentAGI 允许的 CORS 来源"
	EnvDesc_COOKIE_SIGNING_SALT               = "PentAGI Cookie 签名盐值"
	EnvDesc_PROXY_URL                         = "HTTP/HTTPS 代理地址"
	EnvDesc_HTTP_CLIENT_TIMEOUT               = "HTTP 客户端超时（秒）"
	EnvDesc_TERMINAL_TOOL_TIMEOUT             = "终端工具超时（秒）"
	EnvDesc_EXTERNAL_SSL_CA_PATH              = "自定义 CA 证书路径"
	EnvDesc_EXTERNAL_SSL_INSECURE             = "跳过 SSL 验证"
	EnvDesc_PENTAGI_SSL_DIR                   = "PentAGI SSL 证书目录"
	EnvDesc_PENTAGI_DATA_DIR                  = "PentAGI 数据目录"
	EnvDesc_PENTAGI_DOCKER_SOCKET             = "挂载的 Docker 套接字路径"
	EnvDesc_PENTAGI_DOCKER_CERT_PATH          = "挂载的 Docker 证书路径"
	EnvDesc_PENTAGI_LLM_SERVER_CONFIG_PATH    = "自定义 LLM 主机侧配置路径"
	EnvDesc_PENTAGI_OLLAMA_SERVER_CONFIG_PATH = "Ollama 主机侧配置路径"

	EnvDesc_STATIC_DIR     = "前端静态资源目录"
	EnvDesc_STATIC_URL     = "前端静态资源地址"
	EnvDesc_SERVER_PORT    = "后端服务端口"
	EnvDesc_SERVER_HOST    = "后端服务主机"
	EnvDesc_SERVER_SSL_CRT = "后端服务 SSL 证书路径"
	EnvDesc_SERVER_SSL_KEY = "后端服务 SSL 私钥路径"
	EnvDesc_SERVER_USE_SSL = "后端服务启用 SSL"

	EnvDesc_PERPLEXITY_MODEL        = "Perplexity 模型"
	EnvDesc_PERPLEXITY_CONTEXT_SIZE = "Perplexity 上下文大小"

	EnvDesc_SEARXNG_URL        = "Searxng 搜索地址"
	EnvDesc_SEARXNG_CATEGORIES = "Searxng 搜索分类"
	EnvDesc_SEARXNG_LANGUAGE   = "Searxng 搜索语言"
	EnvDesc_SEARXNG_SAFESEARCH = "Searxng 安全搜索"
	EnvDesc_SEARXNG_TIME_RANGE = "Searxng 时间范围"
	EnvDesc_SEARXNG_TIMEOUT    = "Searxng 超时时间"

	EnvDesc_OAUTH_GOOGLE_CLIENT_ID     = "Google OAuth 客户端 ID"
	EnvDesc_OAUTH_GOOGLE_CLIENT_SECRET = "Google OAuth 客户端密钥"
	EnvDesc_OAUTH_GITHUB_CLIENT_ID     = "GitHub OAuth 客户端 ID"
	EnvDesc_OAUTH_GITHUB_CLIENT_SECRET = "GitHub OAuth 客户端密钥"

	EnvDesc_LANGFUSE_EE_LICENSE_KEY   = "Langfuse 企业版许可证密钥"
	EnvDesc_PENTAGI_POSTGRES_PASSWORD = "PentAGI PostgreSQL 密码"

	EnvDesc_GRAPHITI_URL        = "Graphiti 服务地址"
	EnvDesc_GRAPHITI_TIMEOUT    = "Graphiti 请求超时"
	EnvDesc_GRAPHITI_MODEL_NAME = "Graphiti 实体抽取模型"
	EnvDesc_NEO4J_USER          = "Neo4j 用户名"
	EnvDesc_NEO4J_DATABASE      = "Neo4j 数据库名"
	EnvDesc_NEO4J_PASSWORD      = "Neo4j 数据库密码"
)

// dynamic, contextual sections used in processor operation forms
const (
	// section headers
	ProcessorSectionCurrentState = "当前状态"
	ProcessorSectionPlanned      = "计划执行的操作"
	ProcessorSectionEffects      = "影响说明"

	// component labels
	ProcessorComponentPentagi       = "PentAGI"
	ProcessorComponentLangfuse      = "Langfuse"
	ProcessorComponentObservability = "可观测性"

	ProcessorComponentWorkerImage           = "工作容器镜像"
	ProcessorComponentComposeStacks         = "compose 服务栈"
	ProcessorComponentDefaultFiles          = "默认文件"
	ProcessorItemComposeFiles               = "compose 文件"
	ProcessorItemComposeStacksImagesVolumes = "compose 服务栈、镜像、数据卷"

	// common states
	ProcessorStateInstalled = "已安装"
	ProcessorStateMissing   = "未安装"
	ProcessorStateRunning   = "运行中"
	ProcessorStateStopped   = "已停止"
	ProcessorStateEmbedded  = "内置部署"
	ProcessorStateExternal  = "外部服务"
	ProcessorStateConnected = "已连接"
	ProcessorStateDisabled  = "已禁用"
	ProcessorStateUnknown   = "状态未知"

	// planned action bullet prefixes
	PlannedWillStart    = "将启动："
	PlannedWillStop     = "将停止："
	PlannedWillRestart  = "将重启："
	PlannedWillUpdate   = "将更新："
	PlannedWillSkip     = "将跳过："
	PlannedWillRemove   = "将移除："
	PlannedWillPurge    = "将彻底清除："
	PlannedWillDownload = "将下载："
	PlannedWillRestore  = "将还原："

	// effect notes per operation (concise and practical)
	EffectsStart           = "PentAGI 网页界面将变为可用。后台服务会按所需顺序依次启动。"
	EffectsStop            = "网页界面将不可用。进行中的任务流会安全暂停。再次启动 PentAGI 后，任务流会自动恢复，但当前智能体步骤的少量进度可能会丢失。"
	EffectsRestart         = "服务将停止并以干净状态重新启动。预计会有短暂的服务中断，之后任务流会自动恢复。"
	EffectsUpdateAll       = "将拉取镜像，并在需要时重建服务。外部或已禁用的组件会被跳过。预计会有临时的服务中断。"
	EffectsDownloadWorker  = "不会影响正在运行的工作容器。新建的任务流将使用已下载的镜像。若要让现有任务流改用新镜像，请先结束该任务流，然后新建任务或新建助手。"
	EffectsUpdateWorker    = "拉取最新的工作容器镜像。正在运行的工作容器仍使用旧镜像，新建的容器将使用更新后的镜像。"
	EffectsUpdateInstaller = "安装程序二进制文件将被更新，随后程序会退出。请重新启动安装程序以继续。"
	EffectsFactoryReset    = "移除容器、数据卷和网络，恢复默认的 .env 与内置文件，得到一个干净的初始状态。此操作无法撤销。"
	EffectsRemove          = "停止并移除容器，但保留数据卷和镜像。数据不会丢失。在再次启动之前，网页界面将不可用。"
	EffectsPurge           = "彻底清理：删除容器、镜像、数据卷和配置文件。不可恢复。"
	EffectsInstall         = "创建所需文件并启动服务。已检测到的外部组件会被跳过。"
)
