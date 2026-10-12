# PentAGI × ARTEX 工作台

这一版在 PentAGI 内增加 ARTEX 工作台（`/artex`）：使用 PentAGI 登录，在同一界面查看 ARTEX 任务、资产、发现、证据和活动，创建任务、暂停/恢复任务、发起复测，并下载单条发现及其复测记录的 JSON 快照。任务提交同时写入 PentAGI 的 `engine_runs` 本地运行记录，保存引擎、模型绑定版本、远端任务 ID 和提交状态。

ARTEX 仍是独立进程，使用自己的数据库、模型配置、工具环境和智能体调度。PentAGI 后端通过固定接口连接它。此版本没有迁移两个项目的数据库或执行引擎，没有互相派生子任务，也不会把 ARTEX 任务伪装成 PentAGI 原生 Flow。原生任务入口保持可用。

## 连接已有 ARTEX

1. 启动 ARTEX，完成它自己的初始密码和 LLM 配置。
2. 在 PentAGI 的环境变量或管理员「运行时配置」中设置下表参数，随后重启 PentAGI。已有运行时覆盖值会优先于环境变量生效，应使用同一个来源修改配置。
3. 使用带 `flows.admin` 权限的用户登录 PentAGI，从侧栏打开「ARTEX 工作台」。初始为只读模式；需要创建、控制和复测时设置 `ARTEX_ALLOW_WRITES=true` 并重启。

| 参数 | 默认值 | 含义 |
| --- | --- | --- |
| `ARTEX_ENABLED` | `false` | 启用连接 |
| `ARTEX_URL` | 空 | PentAGI 后端可访问的服务根地址，不带 `/api` |
| `ARTEX_API_KEY` | 空 | ARTEX 登录接口返回的 JWT，不是模型 API Key，也不是 `jwt.key` 签名密钥 |
| `ARTEX_ALLOW_WRITES` | `false` | 是否允许通过工作台提交写操作 |
| `ARTEX_TIMEOUT` | `30` | 单次请求超时（秒），有效上限 120；连接检测总超时 10 秒 |

ARTEX 登录接口为 `POST /api/auth/login`，请求包含 `username: "ARTEX"` 和已设置的 `password`，响应的 `token` 即为所需 JWT。核对的 ARTEX 版本签发 7 天令牌；过期后重新登录获取令牌、更新 PentAGI 配置并重启。适配器不保存 ARTEX 登录密码、不自动续期，不把令牌交给前端。

如果两个服务都运行在宿主机上，地址可为 `http://127.0.0.1:8787`；PentAGI 位于容器中时，要使用该容器可访问的地址。同一 Compose 网络使用 `http://artex:8787`，不是容器自身的 `localhost`。

## 可选联合部署

仓库提供 `docker-compose.artex.yml` 作为可选覆盖文件，给 ARTEX 分配独立 PostgreSQL 和数据卷。现有 ARTEX 实例不需要使用此文件。

在 `.env` 中设置 `ARTEX_IMAGE`（已确认的 ARTEX 镜像版本或 digest）和 `ARTEX_POSTGRES_PASSWORD`（使用 URL 安全字符的随机密码）。该覆盖文件只绑定本机 `127.0.0.1:8787` 管理端口；模型和 ARTEX 管理密码在 ARTEX 自己的设置中配置。

```bash
docker compose -f docker-compose.yml -f docker-compose.artex.yml up -d artex-postgres artex
# 在 http://localhost:8787 完成 ARTEX 初始化，获取登录令牌。
# 填写 ARTEX_ENABLED=true、ARTEX_API_KEY，按需打开 ARTEX_ALLOW_WRITES。
docker compose -f docker-compose.yml -f docker-compose.artex.yml up -d pentagi
```

本地代码需先构建成镜像，并通过 `PENTAGI_IMAGE` 指向该镜像，才能使用尚未发布的工作台。不要把仓库的已发布镜像当作已包含本次改动。

## 权限与数据含义

- 一个 ARTEX JWT 对应整个实例，因此本版本仅允许 PentAGI 的 `flows.admin` 用户访问。创建任务还需要 `flows.create`，暂停/恢复及复测需要 `flows.edit`；服务端重复检查，不依赖按钮隐藏。
- 状态检查同时调用健康检查和受保护的任务接口。ARTEX 认证失败会显示独立连接错误，不会注销 PentAGI 会话。
- 资产和发现按明确任务 ID 查询，每页 50 项。资产传 `task_id`；发现使用独立发现表的分页接口和 `task_id`，以获得可用于详情、复测的全局发现 ID。此处不展示来源任务继承的探索节点。
- 活动页面按时间顺序展示，每页至多 100 条，使用游标翻页；刷新按钮更新当前页。当前不提供 SSE 或统一 Token 计费。
- 新建任务会立即由 ARTEX 调度。任务时限默认 1800 秒；目标和范围写入任务要求。PentAGI 的容器、MCP 审批和工具策略不会自动传入 ARTEX，执行限制仍由 ARTEX 配置负责。
- 复测使用 ARTEX 的独立 retester，会显示原始证据与历次复测结论。下载内容仅为当前这条发现和已返回复测记录的快照，不是全任务流量归档。
- 写请求不自动重试。提交后若连接中断，应先刷新核对任务或复测记录，避免人工重复提交。
- 创建任务会先持久化本地运行记录，再调用 ARTEX；成功响应记录远端 ID，响应丢失或连接中断会标记为 `submit_state=uncertain` / `status=sync_unknown`，不会自动再次创建。
- 网络客户端只访问固定接口，拒绝重定向、非法 ID 和不带任务的资产/发现查询；请求体上限 64 KiB，响应体上限 16 MiB。错误和操作日志不记录上游响应正文或凭据。

## 接口与兼容基线

接口位于 `/api/v1/artex`，全部使用 PentAGI 会话/API 令牌认证及上述权限检查：

| 方法 | 路径 | 功能 |
| --- | --- | --- |
| GET | `/status` | 连接、认证、写入能力 |
| GET / POST | `/tasks` | 列表 / 创建 |
| GET | `/runs` | 当前用户的 PentAGI 本地运行映射、引擎来源、模型绑定版本和提交状态 |
| GET | `/tasks/:id` | 任务详情 |
| POST | `/tasks/:id/control` | `{"action":"pause"}` 或 `{"action":"resume"}` |
| GET | `/assets?task=ID&page=1` | 任务资产 |
| GET | `/findings?task=ID&page=1` | 任务发现 |
| GET | `/findings/:id` | 发现详情与证据 |
| GET | `/findings/:id/retests` | 复测记录 |
| POST | `/findings/:id/retest` | `{"notes":"补充说明"}`，最多 4000 字符 |
| GET | `/activity?task=ID&since=0` | 活动记录，后续页传上一页的 cursor |

接口契约核对基线：`Autumn-27/ARTEX@bf7f414477832b77d2152539c0723dc691086522`（2026-09-24 核对），重点参照 `server/server.go`、`server/auth.go`、`server/finding_retests.go`、`server/findings_groups.go` 和 `web/src/lib/api.ts`。升级 ARTEX 后应重跑契约和本地靶场验收。

本适配器独立实现 HTTP 交互，没有引入 ARTEX 的运行时源码或依赖。ARTEX 的发行包继续保留自身许可证与声明；这里没有为两者混合分发作许可结论。

## 验证

```bash
cd backend
go test ./pkg/integrations/artex ./pkg/server/services -run 'Test(Client|Artex)'
cd ../frontend
corepack pnpm exec vitest run src/pages/integrations/artex-workbench.test.tsx e2e/route-manifest.unit.test.ts
corepack pnpm run typescript
```

自动化测试使用 HTTP 测试服务与界面模拟响应，不发起真实扫描。生产启用前，应使用本地靶场走完创建、切换、暂停/恢复、发现详情、复测和令牌失效检查。

## 备份与回滚

升级前分别备份 PentAGI 数据库（包含 `engine_provider_bindings` 与 `engine_runs`）、ARTEX PostgreSQL 数据库、ARTEX 数据卷，以及 ARTEX 的 JWT 签名文件（联合 Compose 中挂载的 `jwt.key`）。不要只备份容器或 ARTEX 数据目录，否则已有登录令牌可能在重建后失效。

回滚时先停止 ARTEX 写入口并保留 PentAGI 原生 Flow，核对 `engine_runs` 中是否有 `pending` 或 `uncertain` 记录；逐条在 ARTEX 任务列表中确认后再决定恢复或人工处理，禁止盲目重放创建请求。确认备份可读后回退应用镜像和适配器配置；数据库迁移只允许按已验证的向前兼容路径恢复，`engine_provider_bindings` 的 down 迁移在仍有数据时会主动拒绝，避免删除模型轮换审计。
