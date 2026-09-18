# PentAGI Optimizer

<div align="center" style="font-size: 1.5em; margin: 20px 0;">
    <strong>P</strong>enetration testing <strong>A</strong>rtificial <strong>G</strong>eneral <strong>I</strong>ntelligence
</div>

<div align="center">

[简体中文](README.md) | [English](README.en.md) | [汉化维护与上游同步](UPSTREAM_SYNC.zh-CN.md)

</div>

## 这是什么

**PentAGI Optimizer** 是 PentAGI 的唯一维护主线：完整保留简体中文默认、英文可选的界面与全部修复（原独立汉化版 [fulaoaz/pentagi](https://github.com/fulaoaz/pentagi) 已并入本仓库、停止独立维护），并在此基础上叠加外部接入协议和漏洞情报工具链。

- 网页首次打开默认显示简体中文，登录页和侧边栏用户菜单可随时切换 English；
- 保持与上游 [vxcontrol/pentagi](https://github.com/vxcontrol/pentagi) 的兼容性；
- 镜像发布到 GitHub Container Registry：`ghcr.io/fulaoaz/pentagi-optimizer`。

项目发布说明、界面预览与部署体验见：[浮潦の小窝 - PentAGI Optimizer 发布](https://fulao.cc/2026/08/27/pentagi-optimizer/)。

## 新增能力

### MCP 协议桥

原生支持 [Model Context Protocol](https://modelcontextprotocol.io) 客户端接入：

- `GET /mcp/sse`：SSE 流式传输，兼容 Claude Desktop 等 MCP 客户端；
- `POST /mcp/message`：JSON-RPC 消息端点；
- `POST /mcp`：单次会话端点；
- 默认要求配置 `MCP_API_KEY`，所有 MCP 端点通过 `Authorization: Bearer <key>` 认证，密钥比较使用常量时间算法；查询参数不再传递密钥。
- 默认只暴露查询型工具；`submit_flow_input` 和 `stop_flow` 仅在显式设置 `MCP_ENABLE_WRITE_TOOLS=true` 后注册。可额外配置 `MCP_WRITE_API_KEY`，将写工具凭据与读取凭据分离。
- 工具参数启用严格 Schema 校验：Flow ID 使用整数并限制为正数，输入文本限制长度；工具元数据会标注只读与破坏性属性，便于 MCP 客户端执行审批。

未配置 Key 时 MCP 端点保持不可用。仅在隔离的本机客户端场景下，才显式设置 `MCP_ALLOW_ANONYMOUS=true`；该选项不会启用写操作。

浏览器客户端还必须在 `MCP_ALLOWED_ORIGINS` 中显式列出来源。该变量默认留空：无 `Origin` 请求的桌面 MCP 客户端可正常使用，而跨域浏览器请求会被拒绝。

如需进一步缩小 AI 客户端的能力面，可在 `MCP_ALLOWED_TOOLS` 中填写逗号分隔的精确工具名。配置后只有列出的工具会出现在 `tools/list` 中；写工具仍需同时满足 `MCP_ENABLE_WRITE_TOOLS=true` 和写范围认证。

安装器的“服务器设置”页面也提供上述 MCP 参数的读取、保存与重置；密钥会以掩码显示，来源和工具白名单会在保存时规范化并校验。

MCP HTTP 请求体默认限制为 `1 MiB`，可通过 `MCP_MAX_REQUEST_BYTES` 调整，避免异常大消息占用过多内存和解析资源。

`get_flow_status` 返回的动态流数据固定限制在 `64 KiB` 内；流标题、任务标题和任务结果均按不可信数据标记，MCP 客户端应将其视为数据而非可执行指令。超出上限时响应会明确标记为截断。

`MCP_READ_TOOL_RATE_LIMIT` 和 `MCP_WRITE_TOOL_RATE_LIMIT` 按身份、工具分别限制每分钟调用次数，默认值分别为 `60` 和 `10`，设为 `0` 可关闭对应限流。`MCP_APPROVAL_MODE` 支持 `scope`、`destructive` 和 `write`：后两者会要求请求额外携带 `X-MCP-Approval: confirm`，分别覆盖破坏性工具或全部写工具。

浏览器 CORS 预检请求只在来源白名单通过后放行，实际消息请求仍必须携带 Bearer 密钥；认证比较使用固定长度摘要，避免因密钥长度不同产生直接比较差异。

每次 MCP 工具调用都会输出结构化审计日志，仅包含工具名、认证模式、读写属性、决策、结果和耗时，以及请求、主体、会话、已校验来源和客户端地址的关联短哈希；`allowed`、`rate_limited`、`approval_required` 与 `write_scope_required` 可区分授权与拒绝原因。调用参数、返回正文、凭据、原始浏览器来源和原始客户端地址不会写入日志。HTTP 访问日志会脱敏 `sessionId`、令牌、密钥及 OAuth 一次性参数。

当 `submit_flow_input` 经 MCP 创建或恢复异步任务时，还会记录 `action=mcp_task_trigger`，其中包含 `flow_id`、`task_id`、`trigger` 和上述关联字段。队列只复制这个精简关联对象，不保留 HTTP 请求上下文；后台工作日志仅保留 `input_bytes`，不会记录输入正文或输入指纹。

携带 MCP 关联的 `terminal` 与 `file` 沙箱动作会额外记录 `action=mcp_sandbox_action`，其中包含流程、任务和子任务 ID、工具、操作类型、结果、耗时及同一组关联字段。该事件和终端失败运行日志不记录命令、文件路径、参数或输出正文，后者仅保留 `result_bytes`。

当流工具在 MCP 关联上下文中预置或移除主容器时，还会记录 `action=mcp_container_lifecycle`，其中包含流程 ID、容器类型、操作、结果、耗时和同一组关联字段。该事件不记录容器 ID、名称、镜像、工作目录、主机路径或 Docker API 细节。

当持久化的主容器状态与 Docker 实际状态不一致时，流工具还会记录通用的 `action=container_recovery` 恢复事件。事件只包含流程 ID、容器类型、恢复原因（`runtime_unavailable`、`runtime_stopped` 或 `stale_status`）、结果和耗时；MCP 关联上下文会附加同一组短哈希。移除或重新预置失败也会记录 `outcome=error`，不会写入容器 ID、名称、镜像、路径、Docker 错误正文或其他运行时详情。

### 浏览器与网络搜索边界

- 浏览器目标地址只接受 `http`/`https`，必须包含有效主机，拒绝 URL 用户信息、无效端口和超过 `8 KiB` 的地址；日志会移除凭据、查询参数和片段。
- 目标主机的 IPv4、IPv6 以及所有 DNS 解析结果都会参与分类；回环、私有、链路本地、未指定、共享地址、保留/文档网段和多播地址统一走内网抓取器，明确的 `.local`、`.internal`、Docker 服务名等本地主机名也不会被当作公网目标。
- Scraper 单次响应体硬上限为 `16 MiB`、响应头上限为 `64 KiB`，网络搜索供应商响应体硬上限为 `4 MiB`，EPSS/CVE/KEV 情报接口响应体硬上限为 `8 MiB`；分块传输没有 `Content-Length` 时同样受限，超限会在解析前终止。
- 进入总结器、日志或长期记忆的浏览器和搜索结果会标记为不可信外部数据；搜索器内部送入总结器的输入进一步限制为 `256 KiB`，文本截断按 UTF-8 边界执行，避免恶意内容改变后续工具策略或生成无效编码。

### 漏洞情报工具链

为智能体新增四个互补的漏洞评估工具，全部可独立开关：

| 工具 | 用途 | 配置 |
| --- | --- | --- |
| `eppss` | 查询 FIRST.org EPSS，获取 CVE 在野利用概率与百分位 | `EPSS_ENABLED` |
| `cvss` | 本地 CVSS 评分：v3.0/3.1 精确计算，v4.0 近似评分 | `CVSS_ENABLED` |
| `cve` | 查询 NVD CVE 详情（描述、CVSS、CWE、参考链接） | `CVE_ENABLED` |
| `kev` | 查询 CISA KEV 已知被利用漏洞目录，按 CVE/厂商/产品过滤 | `KEV_ENABLED` |

工具调用记录会按新引擎类型写入搜索日志，前端筛选与统计无需改动即可识别。

### 发布与 CI

- Docker 镜像发布到 `ghcr.io/fulaoaz/pentagi-optimizer`；
- lint、单元测试、E2E 与镜像构建全链路在 GitHub Actions 自动执行；
- 修复 Windows 环境下脚本可执行位丢失导致的 CI 失败。

## 快速开始

```bash
git clone https://github.com/fulaoaz/pentagi-optimizer.git
cd pentagi-optimizer
cp .env.example .env
# 编辑 .env：配置数据库密码、LLM 提供商密钥等
docker compose up -d
```

浏览器打开 `http://localhost:8443`，注册并登录后即可新建任务流。

## MCP 客户端接入示例

```json
{
  "mcpServers": {
    "pentagi": {
      "type": "http",
      "url": "http://localhost:8443/mcp/sse",
      "headers": {
        "Authorization": "Bearer <MCP_API_KEY>"
      }
    }
  }
}
```

## 环境变量

| 变量 | 默认值 | 说明 |
| --- | --- | --- |
| `MCP_ENABLED` | `true` | 启用原生 MCP 桥 |
| `MCP_API_KEY` | 空 | MCP 端点认证密钥；默认必须配置，使用 Bearer 请求头传递 |
| `MCP_WRITE_API_KEY` | 空 | 可选的写工具专用 Bearer 密钥；设置后读取密钥不能调用写工具，留空时兼容单密钥模式 |
| `MCP_ALLOW_ANONYMOUS` | `false` | 仅隔离本机场景可显式允许无密钥只读接入 |
| `MCP_ALLOWED_ORIGINS` | 空 | 允许访问 MCP 的浏览器来源（逗号分隔）；留空时仅允许无 `Origin` 的桌面客户端 |
| `MCP_ALLOWED_TOOLS` | 空 | 可选的精确工具白名单（逗号分隔）；留空时保留所有已启用工具 |
| `MCP_MAX_REQUEST_BYTES` | `1048576` | MCP HTTP 请求体上限（字节） |
| `MCP_READ_TOOL_RATE_LIMIT` | `60` | 单个身份对单个读取工具的每分钟调用上限；`0` 表示不限制 |
| `MCP_WRITE_TOOL_RATE_LIMIT` | `10` | 单个身份对单个写入工具的每分钟调用上限；`0` 表示不限制 |
| `MCP_APPROVAL_MODE` | `scope` | 写操作审批模式：兼容、破坏性工具或全部写工具；审批请求头为 `X-MCP-Approval: confirm` |
| `MCP_ENABLE_WRITE_TOOLS` | `false` | 显式暴露 `submit_flow_input` 与 `stop_flow` 写工具 |
| `EPSS_ENABLED` | `true` | 启用 EPSS 查询工具 |
| `CVSS_ENABLED` | `true` | 启用本地 CVSS 评分工具 |
| `CVE_ENABLED` | `true` | 启用 NVD CVE 详情工具 |
| `KEV_ENABLED` | `true` | 启用 CISA KEV 目录工具 |

其余配置与上游一致，详见上游 [docs/config.md](https://github.com/vxcontrol/pentagi/blob/main/backend/docs/config.md)。

## 开发

```bash
cd backend
go build ./...
go test ./pkg/tools/...
```

## 更新日志

### 2026-08-31

- 安装器“服务器设置”新增 MCP 完整配置面，支持读写密钥掩码、来源与工具白名单规范化、请求体上限校验及安全默认值恢复；
- MCP 写工具在安装器中增加读取密钥门槛，避免配置界面显示已启用但后端实际未开放；
- MCP 动态流数据增加 64 KiB 输出边界与不可信内容提示，审计日志会明确记录写范围拒绝，并以短哈希关联请求、主体、会话、已校验来源和客户端地址；
- 浏览器目标地址增加协议、主机、端口和 URL 用户信息校验，私有/特殊 IP 与多 DNS 地址统一走内网抓取器；Scraper、搜索器和 EPSS/CVE/KEV 情报接口新增响应体硬上限和 UTF-8 安全截断；
- Docker 守护进程恢复新增 `container_recovery` 最小化审计事件，区分运行时不可用、运行时已停止和持久化状态过期，并排除容器标识与底层错误正文；
- 同步更新 MCP 配置文档与博客维护记录。

### 2026-08-30

- 收口响应式页面操作：任务流、模板、知识库、资源和设置页的页头操作按钮在窄屏下保留可访问名称并扩展为更易触控的按钮尺寸；
- 统一空状态、加载状态和创建入口的卡片结构，并为共享 `HeaderButton`、`StatusCard` 增加回归测试；
- 原生 MCP 桥新增不含参数、结果正文或凭据的结构化调用审计，并复用 SSE 会话服务，确保消息端点能识别客户端已建立的会话；
- MCP 浏览器访问新增独立来源白名单，默认拒绝跨域页面调用，桌面 MCP 客户端保持兼容；
- MCP 请求体新增可配置大小上限，覆盖固定长度和分块传输请求；
- MCP 工具改用强类型参数与严格 Schema 校验，补充正整数、空输入和超长输入边界，并发布只读/破坏性审批提示；
- MCP CORS 预检与实际请求分离处理，认证失败请求仍保持保护；令牌比较改为固定长度摘要比较；
- MCP 支持可选读写双密钥范围，读取密钥不能调用流程变更工具；未设置写密钥时保持既有单密钥兼容行为；
- MCP 支持工具级精确白名单，按最小权限原则缩小 `tools/list` 与可调用工具面；
- 后续上游同步、汉化跟进、增强功能和发布均只在本仓库进行。

### 2026-08-27

- 合并汉化版 `main`：中文默认、英文可选的界面与安装器本地化全部纳入 Optimizer；
- 新增 `kev` 工具：CISA KEV 已知被利用漏洞目录查询（CVE/厂商/产品过滤、按收录日期排序、勒索软件标记）；
- 新增 `cve` 工具：NVD CVE 详情查询；
- 新增 `cvss` 工具：本地 CVSS 3.0/3.1 精确评分与 4.0 近似评分；
- 新增 `eppss` 工具：FIRST.org EPSS 利用概率查询；
- 新增 MCP 协议桥（SSE + JSON-RPC）与可配置 Bearer 认证；
- CI 迁移至 GHCR 发布，修复脚本执行位问题。

## 许可证

沿用上游 [Apache-2.0](LICENSE)；原项目版权归 [vxcontrol](https://github.com/vxcontrol) 所有。
