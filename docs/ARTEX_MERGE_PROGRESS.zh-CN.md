# PentAGI × ARTEX 实施进度

更新日期：2026-10-13。对应已确认方案：[合并规划](ARTEX_MERGE_PLAN.zh-CN.md)。

## 当前状态

| 阶段 | 状态 | 已完成／下一步 |
| --- | --- | --- |
| P0 实施基线 | 完成初始记录 | 52 个已有文件的路径与 SHA-256；28 项原有修改、12 项混合文件、12 项 ARTEX 原型。检查确认本轮验证前这些文件均未改动 |
| P1 真实接口验证 | ARTEX 协议与持久性检查通过 | 真实 ARTEX + PostgreSQL 16：基线 20/20、容器重建 12/12。模型使用本地协议 fixture，不能等同于真实模型或完整产品端到端验收 |
| P2 模型配置共享 | 首版完成并通过真实联调 | PentAGI 提供商归属、配置版本、受管理 profile、远端 ID、显式密钥轮换已实现；联合 QA 中同一提供商生成 ARTEX profile 并被任务实际使用 |
| P3 统一任务中心 | 首版完成并通过双引擎联调 | 统一入口可创建 PentAGI/ARTEX 任务；`engine_runs` 支持提交前持久化、未知状态核对、状态规范化和 `Idempotency-Key` 去重；已实测 ARTEX 创建、幂等重放、暂停/恢复与原生 Flow 创建 |
| P4 统一资产／证据／报告 | 首版完成并通过报告验收 | ARTEX 资产/发现/活动/证据历史和复测沿用来源快照；统一任务列表提供两引擎深链，统一报告接口分别返回 ARTEX 资产/发现/证据和 PentAGI 报告来源 |
| P5 联合发行验收 | 首版联合 QA、备份恢复和重建完成 | PentAGI + ARTEX 隔离联合运行、迁移、PentAGI 数据库备份恢复、容器重建和 ARTEX 绑定持久性均已验证；部署与回滚说明已具备交付条件 |

首版目标已完成。外部真实模型的效果和真实目标上的漏洞发现率不在本轮本地靶场验收范围内，不能由协议 fixture 结果推导。

## 本轮证据

- [文件基线](merge-baseline/2026-09-24-files.json)：只保存源文件路径、分组与哈希，不复制 `.env`、凭据或运行数据。
- [真实 ARTEX 基线结果](merge-baseline/2026-09-24-artex-live-bootstrap.json)：20 项检查通过。
- [容器重建结果](merge-baseline/2026-09-24-artex-live-recreate.json)：12 项检查通过，包含重建前登录令牌继续有效。
- [2026-10-12 容器重建结果](merge-baseline/2026-10-12-artex-live-recreate.json)：12 项检查通过，包含控制令牌、任务、模型绑定、轮换 Key 和复测历史持久化。
- [2026-10-12 备份恢复结果](merge-baseline/2026-10-12-artex-backup-restore.json)：PostgreSQL custom dump 恢复到临时库，48 张公共表恢复后清理临时库。
- [2026-10-13 PentAGI × ARTEX 联合 QA](merge-baseline/2026-10-13-pentagi-artex-joint-qa.json)：迁移、双引擎任务、模型绑定、幂等、暂停/恢复、统一报告、PentAGI 数据库备份恢复和容器重建均通过。
- [暂停与截止时间观察](merge-baseline/2026-09-24-artex-deadline-observation.json)：保留一次失败及其解释，未用新的成功结果掩盖它。
- [可复现验证脚本](../scripts/integration/artex/README.md)：启动、探测、重建及停止步骤。
- 2026-09-25 后端模型共享与运行映射回归：`go test -p 2 ./pkg/integrations/artex ./pkg/server/services ./pkg/server -run 'Test(ModelSharing|ArtexModelSource|Client|Artex|Router)' -count=1` 通过。
- 2026-09-25 前端模型共享回归：`artex-workbench.test.tsx` 9/9、TypeScript 检查和 Vite production build 通过。
- 2026-09-25 统一任务中心前端：路由完整性与 ARTEX 工作台共 15 项测试通过，TypeScript 和 Vite production build 通过。
- 2026-09-25 真实模型共享联调（阶段性记录）：隔离容器内 `TestSQLModelSharingLive` 的模型绑定、轮换、越权和实际模型请求检查通过；后续联合 QA 已在 2026-10-13 收口双引擎任务、报告和重建验收。

验证环境固定到 ARTEX `bf7f414477832b77d2152539c0723dc691086522`，源码检出后保持干净；以 Go 1.26.5 构建 Linux/amd64 后端。OCI 基础镜像和数据库镜像在验证 Compose 中按 digest 固定。

环境在 `F:/CTF/PentAGI/.integration-lab`。测试凭据仅保存在该目录下受限的运行时目录中，不进入 Git；数据库及工作空间使用独立 Docker 卷。ARTEX 只连接 internal 网络，没有挂载宿主机工作目录或 Docker socket。PentAGI QA 容器通过合同实验室的固定网络别名访问本地模型 fixture，控制代理只发布宿主机回环端口。

本轮 QA 使用独立的 PentAGI 数据卷、网络和端口，并重建了本地 `pentagi` 容器以验证新镜像；不涉及其他项目容器。测试任务结束后处于终态，环境保留供复核。

## 已验证的接口行为

1. 健康接口可公开读取，任务接口需要有效 ARTEX 登录令牌。
2. 模型 profile 可以创建、按 ID 更新；读取不返回原始 `api_key`。
3. 显式任务 profile 会用于 worker 和 planner 的实际模型请求；没有把全局默认模型偷偷切换为任务模型。
4. 更新绑定 profile 的 Key 后，任务恢复时使用新 Key，远端 profile ID 不变。
5. 暂停会持久化状态并取消在途模型请求；恢复后的任务仍使用所选 profile。
6. retester 使用它自己的 agent 级模型绑定。测试中原任务用模型 B，而复测明确绑定模型 A，实际请求与此一致；不能宣称复测自动继承任务模型。
7. 对合成发现执行的复测会保存独立记录，原始证据不变；`inconclusive` 不会把发现改成“已修复”。
8. 重建 ARTEX 容器后，暂停任务、绑定 profile、轮换后的 Key 和复测记录保留，任务可再次恢复、请求模型并暂停。

上述“模型请求”指真实 ARTEX 向本地 OpenAI 协议 fixture 发出的请求。fixture 不执行外部工具、不评估真实漏洞，不能作为真实模型效果证据。

## 需要带入后续实现的边界

### 配置版本与密钥轮换

ARTEX 更新 profile 会使其引用任务在后续调用中使用更新值。因此模型、端点和请求参数应形成版本化绑定；Key 轮换则作为单独、显式的操作验证。不能把修改共享 profile 当成只影响新任务的操作。

同一显示名称不是授权依据。后续受管理映射必须包含 PentAGI 实例、提供商归属、提供商 ID、配置版本和远端 profile ID；不能覆盖用户在 ARTEX 手工创建的同名配置。

### 暂停不是冻结任务截止时间

第一次重建检查发生在 90 秒任务的截止时间之后。ARTEX 返回 `status=timeout`、`paused=true`，完成时间正好对应任务截止时间。后续持久性测试改用 3600 秒时限，并在时限内完成重建。

统一任务中心应保存原始状态及截止时间，正确展示这种组合，不能因 `paused=true` 就覆盖已经发生的超时。暂停也不能被当成模型凭据撤销或完全停止后续收尾调用的保证。

### JWT 签名材料必须单独持久化

该 ARTEX 源码版本将 `jwt.key` 放在可执行文件同级，而非可浏览的工作空间。只持久化 `data` 卷不足以保证容器重建后登录令牌继续有效。

验证环境已使用单独的只读签名文件挂载，并验证旧令牌在容器重建后仍有效。正式联合部署仍必须按[联合部署说明](ARTEX_JOINT_DEPLOYMENT.zh-CN.md)单独保存 ARTEX 签名材料、数据卷和数据库备份；不能仅凭数据卷存在就宣称认证材料已备份。

## 首版联合验收记录（2026-10-13）

- 运行镜像 `pentagi-merge-local:20261013` 已构建并启动；`/api/v1/info` 返回 200，容器重建后仍保持运行。
- Goose 迁移版本已统一为连续 14 位版本号，最新 `20260925140000` 成功应用；此前按日期前缀造成的重复版本已消除。
- PentAGI 的 `default:openai` 提供商目录显示兼容；同一配置同步出 ARTEX 受管理 profile `remoteProfileId=12`，ARTEX 任务使用该 profile 实际向本地协议 fixture 发出 `gpt-5.4-nano` 请求。
- 原生 PentAGI Flow 也使用同一模型端点发出真实请求；统一任务列表同时包含 `artex:...` 和 `pentagi:1`，详情分别进入 `/artex?task=13` 与 `/flows/1`。
- 重复使用同一 `Idempotency-Key` 返回同一个 ARTEX 远端任务 ID；ARTEX 暂停和恢复均返回对应状态；两引擎统一报告接口均返回 200 并带来源 artifact。
- PentAGI PostgreSQL custom dump 已恢复到临时库，29 张公共表恢复后清理；随后强制重建 PentAGI 容器，ARTEX 健康状态和 2 个 ready 模型绑定仍可读取。

上述模型请求来自隔离协议 fixture，只证明请求路由、模型绑定和任务生命周期，不代表真实模型效果。没有对未知外部目标执行测试。

## 首版之后的边界

1. 如果需要真实模型效果评测，应在用户明确授权的模型账号和本地靶场上固定预算、范围和版本后另行执行。
2. 如果需要多人强隔离或统一模型网关，应先更新已确认方案；首版仍是单一信任域、配置同步而非统一网关。
3. 继续跟进上游时，保留迁移版本唯一性检查、ARTEX 签名材料备份和本地报告来源校验。

不自动推送、公开发布或部署生产环境；本地首版提交保留在当前分支供复核。

## 2026-10-12 上游合并候选验证

已在隔离候选树合并上游 `vxcontrol/pentagi` 提交 `843dfeb4bfc809c09d1bc5022e567c602d8fc9dd`。合并基线保留上游的 Mistral/xAI、模型上下文和推理能力、版本更新查询、时区统计及安装器更新，同时恢复本地的模型连接配置、密钥加密、在线模型发现、运行时配置、搜索引擎状态、MCP 审计、威胁模型和安全覆盖工具。

已完成的验证：后端 `go build ./...` 通过；GraphQL 后端定向测试通过；前端 GraphQL 类型生成、TypeScript 检查和 Vite production build 通过；提供商、容器恢复、MCP 审计和工具分发定向测试通过。Windows 环境下安装器嵌入文件和部分旧测试仍需在 CI/Linux 生成环境复验，不能把当前本机构建等同于完整 CI 绿灯。

2026-10-12 又补齐了远端 ARTEX task 到本地 run 的证据历史入口：`/artex/tasks/:id/snapshots` 和详情接口会先解析受管理映射，再读取脱敏快照；前端任务详情新增“证据历史”页签。未找到本地 run、快照不存在或归档存储不可用都会返回明确状态，不回退为未归档成功。

2026-10-12 又补齐统一任务创建入口和远端创建幂等：统一任务中心的“新建执行任务”会选择引擎；ARTEX 创建请求支持 `Idempotency-Key`，重复请求复用已创建任务，`pending/uncertain` 状态必须先核对。对应迁移为 `20260925140000_engine_run_idempotency.sql`。

2026-10-12 在隔离 ARTEX 环境完成 bootstrap 20/20、容器重建 12/12 和数据库备份恢复演练。模型请求仍使用本地协议 fixture，因此这些结果证明真实 ARTEX 协议和持久性，不等同于真实外部模型效果。
