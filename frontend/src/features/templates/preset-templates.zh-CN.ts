interface LocalizedPresetTemplate {
    text: string;
    title: string;
}

/** Chinese display and editor content keyed by the upstream English preset title. */
export const zhCNPresetTemplates: Record<string, LocalizedPresetTemplate> = {
    'Active Directory Penetration Test': {
        text: `对域 {{DOMAIN_NAME}} 开展 Active Directory 安全评估

行动计划：
1. 初始访问：测试密码喷洒，检查 AS-REP Roasting，并寻找可执行 Kerberoasting 的账号
2. 域枚举：枚举用户、组、计算机、GPO 和域信任关系
3. 权限提升：查找错误配置的 ACL、可利用的组成员关系和委派问题
4. 凭据收集：在 SYSVOL 中搜索凭据，检查 AD 属性中的密码；条件允许时提取 NTDS.dit
5. 横向移动：测试 Pass-the-Hash、Pass-the-Ticket 和 Overpass-the-Hash
6. 持久化：寻找利用 Golden Ticket、Silver Ticket 和 DCSync 权限的机会
7. 域管理员路径：绘制从当前权限到域管理员权限的攻击路径
8. 报告：记录攻击链、受影响账号和 AD 配置中的安全缺口`,
        title: 'Active Directory 渗透测试',
    },
    'API Security Testing': {
        text: `对 API {{API_BASE_URL}} 进行全面安全评估

行动计划：
1. API 发现：识别所有端点、HTTP 方法和参数
2. 身份认证测试：检查身份认证缺陷、令牌篡改和 JWT 漏洞
3. 授权测试：检查对象级授权缺陷（BOLA/IDOR）和功能级授权绕过
4. 输入校验：测试 SQL、NoSQL、命令和 XXE 注入，以及批量赋值漏洞
5. 速率限制：检查速率限制和暴力破解防护是否缺失
6. 业务逻辑：检查数据过度暴露、资源限制缺失和 API 不安全调用
7. 安全配置：检查 CORS 策略、安全响应头和过于详细的错误信息
8. GraphQL 专项测试（如适用）：检查内省、查询深度限制和批处理攻击
9. 报告：记录 API 漏洞，并提供 curl 或 Postman 概念验证`,
        title: 'API 安全测试',
    },
    'Cloud Infrastructure Security Audit (AWS)': {
        text: `对 AWS 基础设施 {{AWS_ACCOUNT_ID or DOMAIN}} 开展安全审计

行动计划：
1. 侦察：识别 S3 存储桶、EC2 实例和公网端点，并通过 DNS 枚举服务
2. S3 安全：检查存储桶权限、公开访问、ACL 错误配置和存储桶策略
3. IAM 评估：审查角色和策略，检查权限过宽及闲置凭据
4. EC2 安全：查找开放的安全组，测试实例元数据服务（169.254.169.254），检查 IMDSv2
5. 网络安全：审查 VPC 配置、安全组、NACL 和公有子网
6. 数据库暴露：检查 RDS 公网可访问性、安全组和加密设置
7. Lambda 函数：检查函数 URL 暴露、环境变量泄露和 IAM 角色权限
8. CloudTrail 与日志：确认日志记录已启用，并检查安全监控缺口
9. 报告：按优先级记录云安全问题，并给出适用于 AWS 的修复建议`,
        title: 'AWS 云基础设施安全审计',
    },
    'Database Security Assessment': {
        text: `对 {{DATABASE_TYPE}} 数据库（{{HOST:PORT}}）开展安全评估

行动计划：
1. 访问测试：检查默认凭据、弱密码和匿名访问
2. 网络暴露：确认数据库是否应对公网开放，并检查防火墙规则
3. 身份认证：测试身份认证机制、用户枚举和密码策略
4. 授权：审查用户权限，测试权限提升并检查授权过度问题
5. 注入测试：检查应用层 SQL 注入，并测试存储过程注入
6. 配置审查：检查 xp_cmdshell、LOAD DATA、file_priv 等危险配置项
7. 加密：确认静态数据加密和连接 SSL/TLS，检查明文敏感数据
8. 备份安全：测试备份文件访问，检查备份加密并验证恢复流程
9. 审计日志：确认审计日志已启用，测试日志篡改并检查保留策略
10. 报告：记录数据库专项安全问题并给出加固建议`,
        title: '数据库安全评估',
    },
    'DevOps & CI/CD Pipeline Security': {
        text: `评估组织 {{ORGANIZATION}} 的 DevOps 基础设施和 CI/CD 流水线安全

行动计划：
1. 代码仓库安全：扫描 GitHub/GitLab 中暴露的密钥、API Key 以及提交历史中的凭据
2. CI/CD 配置：审查 Jenkins、GitLab CI 和 GitHub Actions 配置，测试流水线定义注入
3. 容器安全：扫描 Docker 镜像漏洞，测试容器逃逸并检查镜像来源
4. 密钥管理：测试 HashiCorp Vault、AWS Secrets Manager 等密钥存储，检查硬编码密钥
5. 访问控制：审查代码仓库、流水线、部署密钥和服务账号的权限
6. 制品安全：扫描构建制品，测试 Nexus、Artifactory 等制品仓库的访问控制
7. Kubernetes 安全：审查 Pod 安全策略、RBAC、网络策略和暴露的控制面板
8. 基础设施即代码：审查 Terraform/Ansible 配置错误和权限过宽的 IAM 角色
9. 监控与日志：确认安全日志正常记录，测试日志篡改并检查监控缺口
10. 报告：记录 DevOps 安全问题并给出流水线加固建议`,
        title: 'DevOps 与 CI/CD 流水线安全评估',
    },
    'External Attack Surface Assessment': {
        text: `评估组织 {{ORGANIZATION_NAME or DOMAIN}} 的外部攻击面

行动计划：
1. 资产发现：枚举所有域名、子域名（subfinder、amass）、IP 网段和 ASN 信息
2. 证书透明度：通过 crt.sh 搜索子域名，找出被遗忘的资产
3. 端口扫描：扫描所有已发现资产的开放端口和服务
4. Web 应用指纹识别：确定所用技术、CMS、框架和服务器版本
5. 邮件安全：检查 SPF、DKIM 和 DMARC 记录，评估邮件伪造风险
6. 云资产发现：搜索暴露的 S3 存储桶、Azure Blob 和云数据库
7. 敏感数据暴露：在 GitHub、GitLab 和 Pastebin 中搜索泄露的凭据与 API 密钥
8. 第三方集成：识别 SaaS 应用、API 端点和合作方集成
9. 漏洞优先级：识别面向互联网的严重漏洞
10. 报告：绘制完整的外部攻击面，并按风险对发现项排序`,
        title: '外部攻击面评估',
    },
    'Internal Network Penetration Test': {
        text: `从 {{INITIAL_ACCESS_LEVEL}} 访问条件出发开展内网渗透测试

行动计划：
1. 网络侦察：执行 ARP 扫描，识别网段并梳理内部基础设施
2. 服务发现：全面扫描内网主机端口，识别关键服务器
3. SMB/NetBIOS 枚举：测试空会话、枚举共享并检查匿名访问
4. 凭据攻击：测试 LLMNR/NBT-NS 投毒（Responder）、中继攻击和密码喷洒
5. 漏洞利用：利用未修补服务，测试默认凭据和已知 CVE
6. 权限提升：利用本地漏洞、错误配置的服务和薄弱权限
7. 横向移动：测试 Pass-the-Hash、令牌模拟和信任关系利用
8. 数据外传：确定敏感数据位置并测试数据防泄漏控制
9. 持久化：建立持久访问机制
10. 报告：记录内网安全状况、攻击路径图和修复优先级`,
        title: '内网渗透测试',
    },
    'Mobile Application Security Testing (API Backend)': {
        text: `对移动应用后端 API {{API_URL}} 开展安全测试

行动计划：
1. 流量拦截：分析移动应用流量，提取 API 端点和身份认证方式
2. 身份认证机制：测试 OAuth 流程、JWT 实现、刷新令牌处理和证书锁定绕过
3. API 端点测试：检查所有已发现端点中的 BOLA/IDOR 和功能级授权缺陷
4. 数据校验：测试 API 参数注入和文件上传端点
5. 业务逻辑：测试高级功能绕过、订阅校验和应用内购买验证
6. 会话管理：测试令牌过期、并发会话处理和会话固定
7. 敏感数据：检查个人身份信息暴露、响应数据过量和硬编码密钥
8. 速率限制：测试登录暴力破解防护、API 速率限制和账号锁定
9. 深层链接：测试深层链接劫持、Intent 重定向（Android）和 URL Scheme 滥用（iOS）
10. 报告：记录移动端专项漏洞并给出缓解建议`,
        title: '移动应用安全测试（API 后端）',
    },
    'Network Infrastructure Discovery & Mapping': {
        text: `对目标网络 {{TARGET_NETWORK}} 开展网络基础设施侦察

行动计划：
1. 网络发现：使用 nmap Ping 扫描识别在线主机并梳理网络拓扑
2. 端口扫描：扫描全部端口（1-65535），识别所有开放服务
3. 服务枚举：识别服务版本和操作系统信息
4. 漏洞扫描：对发现的服务运行自动化漏洞扫描
5. SSL/TLS 分析：检查证书有效性、弱密码套件和协议漏洞
6. Banner 抓取：收集详细的服务信息，用于检索可用漏洞利用
7. 网络图：绘制已发现基础设施的可视化拓扑图
8. 报告：按优先级列出主机、服务和潜在攻击入口`,
        title: '网络基础设施发现与测绘',
    },
    'Web Application Security Assessment': {
        text: `对 Web 应用 {{TARGET_URL}} 进行全面安全评估

行动计划：
1. 应用探索：浏览所有页面，测试各项功能，找出端点和输入入口
2. 按端点开展漏洞测试：
   - 路径遍历：尝试读取 /etc/passwd，重点检查文件下载和上传功能
   - XSS：注入唯一标记、扫描响应，并根据上下文构造载荷
   - SQL 注入：使用 sqlmap 测试输入点，必要时使用 tamper 脚本绕过 WAF
   - 命令注入：使用基于时间的方式检测，并尝试 commix
   - SSRF：使用 Interactsh 进行带外检测，重点检查文件上传和 PDF 生成功能
   - XXE：测试 XML 上传和 Office 文档处理功能
   - 不安全文件上传：测试可执行扩展名、双扩展名和空字节注入
   - CSRF：检查令牌校验，并测试将 POST 请求转换为 GET 请求
3. 身份认证与会话：检查身份认证缺陷、会话固定和弱密码策略
4. 业务逻辑：寻找权限提升、价格篡改和业务流程绕过问题
5. 报告：记录所有发现，并提供复现步骤和概念验证代码`,
        title: 'Web 应用安全评估',
    },
    'WordPress Security Assessment': {
        text: `对 WordPress 站点 {{WORDPRESS_URL}} 开展安全评估

行动计划：
1. 版本识别：确定 WordPress 核心、主题和已启用插件的版本
2. 插件漏洞：枚举已安装插件，使用 WPScan 和 Sploitus 检查已知 CVE
3. 主题漏洞：识别主题版本并搜索已知漏洞利用
4. 用户枚举：通过 REST API、作者归档页和登录响应枚举有效用户名
5. 身份认证测试：检查弱密码、暴力破解防护和 2FA 绕过
6. 文件上传：测试媒体上传限制和任意文件上传漏洞
7. XML-RPC：检查是否启用，并测试 Pingback SSRF 和暴力破解放大
8. SQL 注入：测试搜索功能、自定义查询参数和插件提供的输入点
9. XSS 测试：测试评论、搜索、联系表单和自定义字段
10. 配置问题：检查 wp-config.php 暴露、目录浏览和敏感文件访问
11. 报告：记录 WordPress 专项漏洞及其利用步骤`,
        title: 'WordPress 安全评估',
    },
};
