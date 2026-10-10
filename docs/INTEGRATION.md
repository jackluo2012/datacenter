# 数据中台集成联调指南（DataEase / DolphinScheduler / Doris）

本指南覆盖 `docker-compose.yml` 中两个 **profile 外挂组件**与 **Doris 兼容接入**的联调步骤。
外挂组件仅在对应 profile 激活时启动，默认 `docker compose up` 完全不受影响：

| Profile | 服务 | 定位 | 宿主机入口 | 官方默认凭据 |
|---------|------|------|-----------|-------------|
| `viz` | `dataease` | 可视化（Phase 4，可视化不自研，外挂官方镜像） | http://localhost:8088 | admin / DataEase@123456 |
| `ds` | `dolphinscheduler` | 调度联调环境（Phase 3，job-rpc 经 OpenAPI 驱动） | http://localhost:12345/dolphinscheduler/ui | admin / dolphinscheduler123 |

配套代码位置：

- job-rpc 对接 DS 的客户端：`job/rpc/internal/logic/dsclient.go`（配置走环境变量）
- 数据服务（dc_api）执行入口：网关 `POST /datasvc/execute`（`datasvc.api`）
- 数据源登记/探活：网关 `/datasource/create`、`/datasource/test`（`datasource.api`）

---

## 0. 前置说明

### 0.1 启动基座

```bash
export PATH=$HOME/.local/go/bin:$HOME/go/bin:$PATH   # 如需本机 go 工具链
docker compose build      # 首次构建 11 个服务镜像（见 Dockerfile）
docker compose up -d      # 启动 4 依赖 + 11 服务（不含外挂）
docker compose ps
```

网关唯一对外端口：`http://localhost:8887`（容器名 `datacenter-api`）。

### 0.2 管理面接口鉴权（JWT）

数据源/元数据/任务等管理面接口均挂网关 JWT 保护（未带 token 返回 401）。请求头统一：

```bash
TK=<access_token>
curl -H "Authorization: Bearer $TK" http://localhost:8887/datasource/list
```

token 获取：`POST /user/login`（遗留 user 服务不在本编排内，按 `restart.sh` 另起后再调，
见 `README.md`「登录与鉴权」一节）。免鉴权入口仅 `POST /datasvc/execute`
（走 `app_key + 签名`，见 1.4）。

### 0.3 容器网络内的主机名

compose 中所有服务同在 `datacenter` bridge 网络内，**服务名即主机名**：
`mysql:3306`、`redis:6379`、`etcd:2379`、`datacenter-api:8887`、`dolphinscheduler:12345`。
因此 DataEase / job-rpc 访问平台组件时直接填服务名，不需要 IP。

---

## 1. DataEase 可视化外挂（Phase 4）

镜像：`registry.cn-qingdao.aliyuncs.com/dataease/dataease:v1.18.27`
（v1.18.x 稳定线最后一版；官方镜像仅发布在 DataEase 阿里云仓库，
Docker Hub 无 `dataease/dataease` 官方库——2026-10 已核实，
故 compose 中使用完整仓库地址）。容器内固定监听 **8081**，映射宿主机 **8088**。

### 1.1 启动

```bash
# 激活 viz profile 启动 dataease（depends_on 会连带把 mysql 拉起并等其健康）
docker compose --profile viz up -d dataease
docker compose --profile viz ps
docker compose logs -f dataease     # 看到 Tomcat/Spring 启动完成、8081 监听即就绪
```

### 1.2 一次性准备：创建 DataEase 元数据库

v1.18 官方镜像**不内嵌元数据库**（官方安装器是配套一个独立 mysql 容器）。
本编排复用平台 `mysql` 服务，但需一次性为 DataEase 建独立库（平台业务库 `datacenter`
不受影响）：

```bash
docker compose exec mysql mysql -uroot -padmin \
  -e "CREATE DATABASE IF NOT EXISTS dataease DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;"
```

> 首次启动 dataease 时应用会经 Flyway 自动建全量表结构，无需手工导 SQL。
> 若漏建库，dataease 会因连不上 `dataease` 库反复重启，补建后
> `docker compose --profile viz restart dataease` 即可。

连接配置已由 compose 内置（容器内 `/opt/dataease/conf/dataease.properties` +
Spring 环境变量双通道，指向 `jdbc:mysql://mysql:3306/dataease`），无需再手工配置。

### 1.3 登录

浏览器打开 **http://localhost:8088**：

- 默认账号 `admin`，默认密码 `DataEase@123456`（**首次登录后立即修改**）；
- v1.18 自带 `demo` 演示账号，建议在「系统管理 → 用户管理」中禁用或删除；
- 新建用户的初始密码默认 `DataEase123456`（可在 `./data/dataease/conf/dataease.properties`
  的 `dataease.init_password` 修改，改后 `docker compose --profile viz restart dataease` 生效）。

### 1.4 把平台数据接入 DataEase（数据源配置）

**路径 A：直连库（v1.18 推荐）**

DataEase 与平台同网络，直接把平台 MySQL 配为数据源：

1. 登录 DataEase → 「数据源」→「新建数据源」→ 类型选 **MySQL**；
2. 配置（主机名即 compose 服务名）：
   - 主机名：`mysql`
   - 端口：`3306`
   - 数据库：`datacenter`（平台初始化 SQL 所在库）
   - 用户名/密码：`root` / `admin`
3. 「校验通过」后保存，即可在其上建 SQL 数据集/数据库数据集做仪表板。

也可直连 DataEase 自己的库（`mysql:3306/dataease`）或后续接入的 Doris（见第 3 节）。

**路径 B：API 数据源——把数据服务 API（dc_api）当数据源**

把 `datasvc` 发布的数据 API 作为 DataEase 数据源，走网关统一执行入口：

- 接口：`POST http://datacenter-api:8887/datasvc/execute`（容器网络内可直连服务名）
- 请求体（JSON）：

  ```json
  {
    "app_key": "<注册应用返回的 app_key>",
    "ts": "<unix 秒>",
    "sign": "<hex(hmac-sha256(app_secret, app_key+ts))>",
    "path": "/api/v1/<已发布的 API 路径>",
    "params_json": "{}"
  }
  ```

- 返回：`rows_json`（结果集 JSON 串）/ `cost_ms` / `cache_hit`。
  前置步骤：`/datasvc/api/create` → `/datasvc/app/register`（返回 app_key/app_secret，
  secret 仅返回一次）→ `/datasvc/grant` → `/datasvc/api/publish`。

> 注意：**API 类型数据源是 DataEase v2.x 的内置能力**，v1.18.27 未内置 REST API
> 数据源，v1.18 下请走路径 A 直连，或先在数据服务侧把 API 结果落表后直连该表。
> 若升级到 DataEase v2（容器内端口变为 8080，需同步调整 compose 端口映射），
> 可将上述接口直接配为 API 数据源，用返回的 `rows_json` 解析出 JSON 数据集。

### 1.5 持久化与停止

- 持久化目录：`./data/dataease/`（conf / logs / 数据子目录 / custom-drivers，
  子目录挂载方式对齐官方安装器，镜像内置文件不被覆盖）；
- 停止不影响基座：`docker compose --profile viz stop dataease`；
- 完整移除：`docker compose --profile viz down` 只移除 profile 内服务。

---

## 2. DolphinScheduler 调度联调（Phase 3）

镜像：`apache/dolphinscheduler-standalone-server:3.2.1`（3.2.x 稳定线）。
standalone 单进程内含 master/worker/api/alert + **内嵌 zookeeper + 内嵌 H2 库**，
因此**不需要** `POSTGRESQL_*` 等外部数据库环境变量；端口 `12345`（API+UI）。

> standalone 的元数据是 **H2 内存库，容器重启即丢失**（项目/工作流/令牌需重建）；
> compose 挂载的 `./data/dolphinscheduler:/opt/dolphinscheduler/logs` 仅持久化日志。
> 联调够用，生产化请另用 DB 持久化部署（非本编排范围）。

### 2.1 启动与 DS 初始化

```bash
docker compose --profile ds up -d dolphinscheduler
# 就绪探针（standalone 启动约 1 分钟）：
curl -s http://localhost:12345/dolphinscheduler/actuator/health
```

1. 浏览器打开 **http://localhost:12345/dolphinscheduler/ui**，
   默认凭据 `admin / dolphinscheduler123`（首次登录后修改）；
2. 「项目管理」→ 新建项目，名称与后续 `DS_PROJECT` 一致（示例：`datacenter`）；
3. 「安全中心 → 令牌管理」→ 创建令牌（选用户 admin、设置失效时间），
   生成的 **Token 串** 即 `DS_TOKEN`。

### 2.2 job-rpc 侧环境变量配置

job-rpc 的 DS 集成不走 yaml 配置（go-zero 脚手架 config 结构体不可改），
从环境变量读取（实现见 `job/rpc/internal/logic/dsclient.go`）：

| 环境变量 | 含义 | 联调取值 |
|---------|------|---------|
| `DS_BASE_URL` | DS OpenAPI 基址 | `http://dolphinscheduler:12345/dolphinscheduler` |
| `DS_TOKEN` | API Token（请求头 `token`，DS 3.2+ 推荐方式） | 「安全中心 → 令牌管理」生成的串 |
| `DS_PROJECT` | 目标项目：纯数字按项目编码，否则按项目名解析 | `datacenter` |

在 `docker-compose.yml` 的 `job` 服务上追加 `environment`（覆盖模板时保留 `TZ`）：

```yaml
  job:
    <<: *go-service
    container_name: datacenter-job
    command: ["./job-server", "-f", "/app/etc/job.yaml"]
    environment:
      TZ: Asia/Shanghai
      DS_BASE_URL: http://dolphinscheduler:12345/dolphinscheduler
      DS_TOKEN: <在 DS UI 创建的 token>
      DS_PROJECT: datacenter
```

然后重建使其生效（环境变量在容器创建时固化）：

```bash
docker compose up -d job
docker compose exec job env | grep ^DS_
```

> 网络说明：`job` 与 `dolphinscheduler` 同在 `datacenter` 网络，服务名可直接直连。
> 若希望 job 随 `--profile ds` 一起启停，给 `job` 服务加 `profiles: ["ds"]` 即可，
> 效果等同（同网络服务名互通）；默认编排保持 job 常驻、服务定义结构不变。
> 未配置 `DS_BASE_URL` 时 = 调度集成未启用：部署/触发接口返回业务结果 `ok=false`
> 而非进程报错，不影响其余功能。

### 2.3 DeployToScheduler 验证步骤（对照 issue #21 验收）

按顺序执行（管理面接口需 0.2 节的 JWT 请求头）：

```bash
BASE=http://localhost:8887
TK=<access_token>
AUTH="Authorization: Bearer $TK"

# 1) 调度器连通性：configured=true 且 connected=true 即通
#    （内部等价于 dsclient.ping()：GET /projects 分页接口，同时校验地址与 token）
curl -H "$AUTH" "$BASE/job/scheduler/info?id=0"
# 期望：{"configured":true,"connected":true,"base_url":"http://dolphinscheduler:12345/dolphinscheduler"}

# 2) 创建同步任务（source_ds_id 指向已登记的数据源 id，见 /datasource/list）
curl -XPOST -H "$AUTH" "$BASE/job/task/create" \
  -d '{"name":"ods_demo_sync","source_ds_id":1,"mode":"full","cron":"0 0 * * * ?"}'

# 3) 部署到调度器：ok=true 且 ds_workflow_code>0 即通
#    （内部：解析/确保项目 -> 确保工作流定义存在 -> 上线 ONLINE）
curl -XPOST -H "$AUTH" "$BASE/job/task/deploy" -d '{"id":1}'
# 期望：{"ok":true,"ds_workflow_code":<code>}；DS UI「工作流定义」出现同名工作流且已上线

# 4) 手动触发：ok=true 且 ds_process_instance_id>0 即通
curl -XPOST -H "$AUTH" "$BASE/job/task/trigger" -d '{"id":1}'
# 期望：DS UI「工作流实例」出现新实例（空任务壳，状态跑完即成功）

# 5) 状态回查：status 为 DS 实例状态映射
curl -H "$AUTH" "$BASE/job/task/status?id=1"
```

验收口径（对照 issue #21）：`scheduler/info` 连通 → `deploy` 幂等部署成功
（重复 deploy 返回同一 workflow code）→ `trigger` 拉起实例 → `status` 能取回实例状态，
四步全通即 DS 集成闭环成立。

### 2.4 已知待验证点（对照 dsclient.go 头部 TODO，联调时重点核对）

1. **Token 鉴权方式**：3.2+ 为请求头 `token: <token>`；若对旧版本（<3.1.x）联调，
   需改为 `POST {base}/login`（userName/userPassword 表单）换 sessionId(cookie)；
2. **项目/工作流列表分页结构**：按 `totalList` 数组 + `searchVal` 过滤解析，
   需与 3.2.1 实际返回结构核对；
3. **创建工作流参数**：当前提交 `name/location/description/tasks` 空任务壳；
   3.x 官方接口参数为 `name/description/locations + taskDefinitionJson/taskRelationJson`
   （表单或 JSON），联调时按实际接口校正字段名；
4. **上线接口**：`POST .../release`，body `{"releaseState":"ONLINE"}`，核对参数位置
   （path 上的工作流编码）与大小写；
5. **启动实例**：当前仅传 `processDefinitionCode`，需补齐 `failureStrategy` 等运行策略参数；
6. **实例列表过滤**：`processDefinitionCode` 查询参数需验证 3.2.1 是否支持按此过滤；
7. **空任务壳的调度语义**：创建的空壳工作流仅用于打通管控面链路，
   真实任务定义（SQL/SeaTunnel 节点）由执行面推送——联调时确认空壳可正常上线/跑完。

另注意：standalone（H2 内存库）重启后项目/令牌丢失，`DS_PROJECT` 指向的项目与
`DS_TOKEN` 需重建，再重跑 2.3 步骤。

### 2.5 常见问题

- `scheduler/info` 返回 `connected=false, msg=...`：先看 msg——连接拒绝多为
  dolphinscheduler 未就绪（等 healthcheck 通过）；`code=1030c` 类业务错误多为 token
  无效/过期，回 2.1 重建令牌；
- deploy 报「调度器项目 "datacenter" 不存在」：DS UI 项目名与 `DS_PROJECT` 不一致，
  或 H2 重启后项目未重建；
- 容器内验证网络：`docker compose exec job wget -qO- http://dolphinscheduler:12345/dolphinscheduler/actuator/health`。

---

## 3. Doris 接入说明（MySQL 协议兼容）

### 3.1 原理

Doris FE 的 **9030 端口是 MySQL 协议查询端口**。平台数据源模块按 type=mysql 走
JDBC/SQL 直连（探活、元数据采集 `information_schema`、IDE 查询、质量/指标取数均如此），
因此 **兼容 MySQL 协议的 Doris 天然可接入，无需任何代码改动**：登记时 `type` 填
`mysql`、端口填 `9030` 即可完成 探活/采集/IDE 全链路。

### 3.2 登记 Doris 并验证

> host 按实际部署填写「容器网络内可达」的地址：Doris 与 compose 同机时可用
> 宿主机地址（如 `host.docker.internal`，Docker Desktop/WSL2 默认支持）或 Doris 所在
> 机器 IP；Doris 若也容器化并接入 `datacenter` 网络，则直接填其服务名。

```bash
BASE=http://localhost:8887
AUTH="Authorization: Bearer <access_token>"

# 1) 登记 Doris（type=mysql，9030 查询端口；凭证 AES-GCM 加密入库）
curl -XPOST -H "$AUTH" "$BASE/datasource/create" \
  -d '{"name":"doris-olap","type":"mysql","host":"host.docker.internal","port":9030,"db_name":"ods","username":"root","password":"<doris密码>"}'

# 2) 探活（GET /datasource/test?id=<返回的 id>）
curl -H "$AUTH" "$BASE/datasource/test?id=2"

# 3) 采集元数据（information_schema → 元数据实体，与 MySQL 库同流程）
curl -XPOST -H "$AUTH" "$BASE/metadata/collect" -d '{"ds_id":2}'

# 4) IDE 中直接对该数据源执行 SQL 查询（9030 MySQL 协议）
```

登记成功后，DataEase（第 1 节路径 A）也可以把同一 Doris 直连为数据源做可视化。

---

## 4. 速查表

| 项 | 值 |
|----|----|
| 网关 | http://localhost:8887（容器内 `datacenter-api:8887`） |
| 数据 API 执行 | `POST /datasvc/execute`，`sign=hex(hmac-sha256(app_secret, app_key+ts))` |
| DataEase 启动 | `docker compose --profile viz up -d dataease` |
| DataEase 入口/凭据 | http://localhost:8088 ，admin / DataEase@123456（首登改密） |
| DataEase 元数据库 | 平台 `mysql:3306` 库 `dataease`（首次需手工 CREATE DATABASE，见 1.2） |
| DataEase 持久化 | `./data/dataease/`（conf/logs/data/custom-drivers） |
| DS 启动 | `docker compose --profile ds up -d dolphinscheduler` |
| DS 入口/凭据 | http://localhost:12345/dolphinscheduler/ui ，admin / dolphinscheduler123 |
| DS OpenAPI 基址 | `http://dolphinscheduler:12345/dolphinscheduler`（容器内） |
| DS 健康检查 | `GET /dolphinscheduler/actuator/health` |
| job-rpc DS 环境变量 | `DS_BASE_URL` / `DS_TOKEN` / `DS_PROJECT`（见 2.2） |
| DS 持久化 | `./data/dolphinscheduler/`（仅日志；H2 元数据重启丢失） |
| Doris 登记 | `type=mysql` + FE 查询端口 `9030`，无代码改动 |
| compose 校验 | `docker compose [--profile viz|--profile ds] config -q` |
