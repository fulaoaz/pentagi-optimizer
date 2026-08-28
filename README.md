# PentAGI Optimizer

<div align="center" style="font-size: 1.5em; margin: 20px 0;">
    <strong>P</strong>enetration testing <strong>A</strong>rtificial <strong>G</strong>eneral <strong>I</strong>ntelligence
</div>

<div align="center">

[简体中文](README.md) | [English](README.en.md) | [汉化维护与上游同步](UPSTREAM_SYNC.zh-CN.md)

</div>

## 这是什么

**PentAGI Optimizer** 基于 [fulaoaz/pentagi](https://github.com/fulaoaz/pentagi) 简体中文维护版独立演进：完整保留汉化版的中文默认、英文可选界面与全部修复，再叠加外部接入协议和漏洞情报工具链。

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
- 可配置 `MCP_API_KEY`：设置后所有 MCP 端点要求 `Authorization: Bearer <key>`，密钥比较使用常量时间算法防时序攻击。

默认未配置 Key 时保持匿名本地模式，适合只在本机使用 MCP 客户端的场景。

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
| `MCP_API_KEY` | 空 | MCP 端点认证密钥，设置后启用 Bearer 校验 |
| `EPSS_ENABLED` | `true` | 启用 EPSS 查询工具 |
| `CVSS_ENABLED` | `true` | 启用本地 CVSS 评分工具 |
| `CVE_ENABLED` | `true` | 启用 NVD CVE 详情工具 |
| `KEV_ENABLED` | `true` | 启用 CISA KEV 目录工具 |

其余配置与汉化版/上游一致，详见上游 [docs/config.md](https://github.com/vxcontrol/pentagi/blob/main/backend/docs/config.md)。

## 开发

```bash
cd backend
go build ./...
go test ./pkg/tools/...
```

## 更新日志

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
