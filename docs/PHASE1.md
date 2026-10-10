# Phase 1 开发契约（数据源 / 元数据 / 数据服务 / 资产目录）

> 本文档是 4 个服务并行开发的**契约**。骨架（proto → pb → logic 桩 → config → example yaml）已生成并提交，各服务实现**只允许修改自己目录内的文件**。

## 目录与端口分配

| 服务 | 目录 | 端口 | etcd key | 表 |
|---|---|---|---|---|
| 数据源 | `datasource/rpc` | :8085 | `datasource.rpc` | dc_datasource |
| 元数据 | `metadata/rpc` | :8086 | `metadata.rpc` | dc_meta_entity / dc_meta_relation / dc_meta_version |
| 数据服务 | `datasvc/rpc` | :8087 | `datasvc.rpc` | dc_api / dc_api_app / dc_api_grant |
| 资产目录 | `catalog/rpc` | :8088 | `catalog.rpc` | dc_asset / dc_asset_category / dc_asset_apply |

接口契约：各 `*/rpc/*.proto`（已定稿，不得改 message/rpc 签名）。表结构：`docs/schema/phase1.sql`（已定稿）。

## 架构约定

1. **网关编排，rpc 纯业务**：跨服务依赖（如 metadata 采集需要 DSN）由网关 logic 先调 datasource-rpc 拿 DSN 再传参，rpc 服务之间不互相调用。唯一例外：`datasvc.ExecuteApi` 内部 import `datacenter/datasource/rpc/dataSourceClient` 获取 DSN（同 module 直连 client）。
2. **model 层**：对齐现有风格（`*/model/` 目录 + go-zero sqlx/sqlc）。新服务直接用 `sqlx.NewMysql` + 手写 model（不需要 cache 的表用 sqlx；dc_api 可用 sqlc 缓存）。
3. **密码/密钥加密**：AES-GCM，key 取 `config.EncryptKey`（32 字节）。实现在各自服务内部（datasource：`internal/logic/credential.go`；datasvc 同理），不建共享包。加密结果存 varbinary 列。
4. **多租户**：所有查询带 `beid` 条件；软删仅 dc_datasource（deleted_at）。
5. **SQL 占位符**：datasvc 的 SQL 模板用 `:name` 风格参数，执行前校验参数白名单（防注入），仅允许 SELECT 模板执行向导查询；sql_text 若含非 SELECT 前缀（INSERT/UPDATE/DELETE/DROP 等）必须在 CreateApi 时拒绝（mode=sql 允许写语句，写语句不缓存）。
6. **注释与命名**：中文注释，命名对齐现有代码风格（goZero style）。

## 交付要求（每个服务）

- [ ] `*/model/`：表 model（FindOne/List/Insert/Update/Delete + 业务方法）
- [ ] `internal/logic/*`：填充全部 logic 桩（错误用 `errors.New` 业务错误 + `xzrpc` 语义；对齐 votes 服务的写法）
- [ ] `internal/svc/serviceContext.go`：装配 model 连接
- [ ] 至少 1 个核心 logic 的单元测试（sqlmock 或纯函数测试；datasvc 重点测 ExecuteApi 的参数校验/签名/脱敏/缓存）
- [ ] `go build ./<svc>/...` 与 `go vet ./<svc>/...` 通过
- [ ] 不修改：proto、pb、client、server、main、config.go、etc/*.example.yaml、shared/、其他服务目录

## 网关接入（主线收口，各代理不管）

`datasource.api` / `metadata.api` / `datasvc.api` / `catalog.api` 四个网关文件与编排 logic 由主线统一完成。
