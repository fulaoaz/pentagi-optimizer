# PentAGI × ARTEX 联合部署、备份与回滚

本文对应当前本地合并提交，不包含任何密钥。正式启动前必须把 ARTEX 镜像固定到已经验证的 tag 或 digest，并在同一台机器上保存 ARTEX 的 JWT 签名材料。

## 启动前检查

```powershell
docker version
docker compose version
docker compose -f docker-compose.yml -f docker-compose.artex.yml config --quiet
```

如果 `docker version` 的 Server 部分不可用，先启动 Docker Desktop 的 Linux 引擎。不要在 Server 不可用时执行迁移或删除卷命令。

复制 `.env.example` 为本地环境文件，并填写：

- `ARTEX_ENABLED=true`
- 固定的 `ARTEX_URL`、`ARTEX_API_KEY` 和 `ARTEX_IMAGE`
- `ARTEX_POSTGRES_PASSWORD`
- PentAGI 的数据库、模型提供商和加密密钥

ARTEX 控制令牌和模型提供商密钥分别配置。环境文件只保存在部署主机，不进入 Git、前端或日志。

## 联合启动

```powershell
docker compose -f docker-compose.yml -f docker-compose.artex.yml pull
docker compose -f docker-compose.yml -f docker-compose.artex.yml up -d
docker compose -f docker-compose.yml -f docker-compose.artex.yml ps
```

启动后检查 PentAGI 健康接口、ARTEX 健康接口和统一任务中心。第一次使用 ARTEX 前，在 PentAGI 模型提供商页面同步一个受管理模型绑定，再从统一任务中心选择 ARTEX 引擎创建任务。

## 备份

先记录当前镜像和迁移版本，再备份 PostgreSQL。数据库备份与 ARTEX 数据卷必须使用同一时间点的维护窗口：

```powershell
$stamp = Get-Date -Format 'yyyyMMdd-HHmmss'
New-Item -ItemType Directory -Force "backups/$stamp" | Out-Null
docker compose -f docker-compose.yml -f docker-compose.artex.yml config > "backups/$stamp/compose.rendered.yml"
docker compose -f docker-compose.yml -f docker-compose.artex.yml images > "backups/$stamp/images.txt"
docker compose -f docker-compose.yml -f docker-compose.artex.yml exec -T artex-postgres pg_dump -U artex -d artex -Fc > "backups/$stamp/artex.dump"
```

另外保存 ARTEX 数据卷中的 `jwt.key` 和 `data` 内容，确保文件权限只允许部署账号读取。只备份 PostgreSQL 或只备份 Docker 卷都不足以保证旧控制令牌在恢复后继续有效。

## 回滚

回滚前停止写入任务并确认没有正在运行的 ARTEX 任务：

```powershell
docker compose -f docker-compose.yml -f docker-compose.artex.yml stop pentagi artex
```

恢复同一备份时间点的 ARTEX PostgreSQL dump、`jwt.key` 和 `data` 卷后，先启动数据库，再启动 ARTEX，最后启动 PentAGI：

```powershell
docker compose -f docker-compose.yml -f docker-compose.artex.yml up -d artex-postgres
docker compose -f docker-compose.yml -f docker-compose.artex.yml exec -T artex-postgres pg_restore -U artex -d artex --clean --if-exists < backups/<stamp>/artex.dump
docker compose -f docker-compose.yml -f docker-compose.artex.yml up -d artex pentagi
```

应用镜像回滚必须和数据库迁移兼容。先恢复数据库，再恢复镜像；如果迁移不可逆，使用同版本备份恢复，不要只替换容器镜像。

## 验收记录

联合部署验收至少记录：

1. PentAGI 与 ARTEX 健康状态和版本。
2. 一个受管理模型绑定及其 revision。
3. 一个 PentAGI 原生任务和一个 ARTEX 任务在统一任务中心的状态。
4. ARTEX 任务暂停、恢复、容器重启后的状态和控制令牌。
5. 资产、发现、复测和证据快照的来源、版本与内容指纹。
6. 备份恢复后旧令牌、模型绑定和证据快照是否仍可用。

本机没有运行 Docker Server 时，只能完成源码构建和静态/定向测试，不能把这些结果写成联合部署或真实模型验收通过。
