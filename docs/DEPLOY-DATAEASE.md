# DataEase 深度集成部署指南（从零到出图）

> 定位：可视化不自研（ROADMAP 决策），旁挂 DataEase v1.18.27 官方镜像。
> 本指南覆盖 **独立 compose 片段部署 → 初始化登录 → 对接平台数据（直连 / API 两条路线）→
> 典型仪表板 → 前端「数据可视化」页面与菜单** 的完整闭环。
> 配套文件：`deploy/dataease/docker-compose.dataease.yml`、`deploy/dataease/dataease.properties.template`、
> `docs/dataease-menu-seed.sql`、`web/src/views/dataease/DataeaseEmbed.vue`。
> 背景阅读：[INTEGRATION.md](INTEGRATION.md) 第 1 节（viz profile 外挂方式，与本指南二选一）。

---

## 0. 架构与两种部署方式（二选一）

| 方式 | 命令 | 适用 |
|------|------|------|
| A. 平台编排 viz profile | `docker compose --profile viz up -d dataease` | 一键全栈，连接参数已内置（见根 `docker-compose.yml`） |
| B. 独立 compose 片段 | 见下文第 1 步 | 单独部署/升级 DataEase，或平台编排之外维护可视化组件 |

两种方式**二选一**（容器名同为 `datacenter-dataease`，同时启动会报容器名冲突）。

独立片段要点（`deploy/dataease/docker-compose.dataease.yml`）：

- 镜像 `registry.cn-qingdao.aliyuncs.com/dataease/dataease:v1.18.27`（v1.18 稳定线最后一版；
  官方镜像仅发布在 DataEase 阿里云仓库，Docker Hub 无官方库，2026-10 已核实）；
- 端口：宿主机 **8088** → 容器固定 8081；
- **不含 mysql 服务**：元数据库复用平台 MySQL（`datacenter_datacenter` 网络内用服务名 `mysql:3306` 直连）；
- 健康检查：`nc -z 127.0.0.1 8081`（start_period 180s，首启跑 Flyway 建表较慢属正常）；
- 持久化：与平台编排共用仓库根 `./data/dataease/`（conf 之外的 logs/data/custom-drivers 子目录挂载，
  对齐官方安装器目录结构）；连接配置由 `dataease.properties`（官方安装器同款）挂载 +
  `SPRING_DATASOURCE_*` 环境变量双通道注入，环境变量优先。

可覆盖的环境变量（`up -d` 前设置，重建生效）：

| 变量 | 默认 | 说明 |
|------|------|------|
| `DATAEASE_DB_HOST` | `mysql` | 元数据库主机（平台网络内=服务名；独立部署=内网 IP / `host.docker.internal`） |
| `DATAEASE_DB_PORT` | `3306` | 元数据库端口 |
| `DATAEASE_DB_NAME` | `dataease` | 元数据库库名 |
| `DATAEASE_DB_USER` / `DATAEASE_DB_PASSWORD` | `root` / `admin` | 平台 MySQL 口令 |
| `DATAEASE_PORT` | `8088` | 宿主机端口 |
| `DATACENTER_NETWORK` | `datacenter_datacenter` | 外部网络名（`docker network ls \| grep datacenter` 核对） |

---

## 1. 启动 DataEase

```bash
# 仓库根目录执行
cp deploy/dataease/dataease.properties.template deploy/dataease/dataease.properties
docker compose -f deploy/dataease/docker-compose.dataease.yml up -d
docker compose -f deploy/dataease/docker-compose.dataease.yml logs -f dataease
# 看到 Tomcat/Spring 启动完成、8081 监听即就绪（首启约 1-3 分钟）
docker compose -f deploy/dataease/docker-compose.dataease.yml ps   # healthy 即就绪
```

> 独立部署（平台编排未启动）：先 `docker network create datacenter_datacenter`，
> 并设置 `DATAEASE_DB_HOST` 指向可到达的 MySQL 地址（Docker Desktop 用
> `host.docker.internal`，Linux 宿主机用 `172.17.0.1` 或内网 IP）。

## 2. 创建 dataease 元数据库（一次性）

v1.18 官方镜像**不内嵌元数据库**（官方安装器配套独立 mysql 容器），本方案复用平台 MySQL，
需一次性建独立库（平台业务库 `datacenter` 不受影响）：

```bash
docker compose exec mysql mysql -uroot -padmin \
  -e "CREATE DATABASE IF NOT EXISTS dataease DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;"
```

- **时机不限**：先建库再启动最佳；若 dataease 已因连不上库反复重启，补建后
  `docker compose -f deploy/dataease/docker-compose.dataease.yml restart dataease` 自动恢复；
- 首次启动应用经 **Flyway 自动建全量表结构**，无需手工导 SQL。

## 3. 登录

浏览器打开 **http://localhost:8088**：

- 默认账号 **`admin` / `dataease`**（v1.18 官方默认，首次登录后立即修改）；
- 系统内**新建用户**的初始密码默认 `DataEase123456`（`dataease.init_password` 可改，改后重启生效）；
- v1.18 自带 `demo` 演示账号，建议在「系统管理 → 用户管理」禁用或删除。

## 4.（可选）导入演示数据

典型仪表板模板（第 7 节）基于电商演示数据包，需先导入 `dm_*` 表
（已导入可跳过，步骤见 [demo/DEMO.md](demo/DEMO.md) 第 0.2 节）：

```bash
docker compose exec -T mysql mysql -uroot -padmin datacenter < docs/demo/demo_data.sql
# 自检：SELECT COUNT(*) FROM dm_orders;  -- 1500
```

---

## 5. 对接平台数据两条路线

### 5.0 路线怎么选

| | 路线 A：直连库 | 路线 B：API 数据源 |
|---|---|---|
| 接法 | DataEase 数据源 → 平台 MySQL（`datacenter` 库） | DataEase 数据源 → `POST /datasvc/execute`（app_key + 签名） |
| 版本要求 | v1.18 原生支持（推荐） | **API 数据源是 v2.x 内置能力，v1.18 无**（见 5.2 的落表替代） |
| 治理能力 | 直读库，绕过服务层 | 走平台统一入口：鉴权/限流/结果缓存/字段脱敏/调用审计 |
| 签名成本 | 无 | 有（HMAC-SHA256，ts 容差 ±300 秒，见 5.2） |
| 适用场景 | 内部 BI 分析、临时探索、大数据量出图 | 对外/跨网交付、口径收敛（SQL 收敛在数据服务里）、需要脱敏与配额 |

结论：**内部 BI 起步用路线 A；要把「已发布的数据 API」原样出图（口径/脱敏/限流收口在服务层）才走路线 B。**

### 5.1 路线 A：直连平台库（step by step）

1. 登录 DataEase → 顶部「**数据源**」→「新建数据源」→ 类型选 **MySQL**，配置：

   | 字段 | 值 | 说明 |
   |------|----|------|
   | 主机名/IP | `mysql` | compose 服务名即主机名（dataease 与平台同在 datacenter 网络）；独立部署填内网 IP |
   | 端口 | `3306` | |
   | 数据库 | `datacenter` | 平台初始化 SQL 所在库（含 `dc_*` 平台表与 `dm_*` 演示表） |
   | 用户名 / 密码 | `root` / `admin` | 平台 MySQL 口令 |
   | 额外参数 | 留空 | 连接池参数走默认 |

2. 点「**校验**」，提示校验通过后保存（命名为 `datacenter-mysql`）；
3. 「**数据集**」→「新建数据集」→「**SQL 数据集**」→ 选刚才的数据源 → 贴 SQL（第 7 节有现成模板）
   → 确认后字段列表自动生成（数值/时间字段类型可手工校正）；
4. 「**仪表板**」→「新建仪表板」→「添加视图」→ 选数据集 → 选图表类型 → 拖维度/指标字段 → 保存。

> 同一数据源也可直连 `dataease` 库（看 DataEase 自身元数据）或后续接入的 Doris（见 INTEGRATION.md 第 3 节）。

### 5.2 路线 B：API 数据源（指向 /datasvc/execute）

**前置：在平台侧准备一个已发布的 API + 一个应用**（详见 [demo/DEMO.md](demo/DEMO.md) 第 6 节）：

```bash
TK=<管理面 access_token>   # POST /user/login 获取（管理面接口全部挂 JWT）
# 1) 建 API（wizard 模式强制 SELECT，:name 占位符）
curl -s -XPOST http://127.0.0.1:8887/datasvc/api/create -H "Authorization: Bearer $TK" \
  -H 'Content-Type: application/json' \
  -d '{"name":"查询订单","path":"/api/v1/demo/orders","mode":"wizard","datasource_id":1,
       "sql_text":"select order_no, amount, status, created_at from dm_orders where user_id = :user_id order by created_at desc limit :limit",
       "params_json":"[{\"name\":\"user_id\",\"type\":\"number\",\"required\":true},{\"name\":\"limit\",\"type\":\"number\",\"required\":false,\"default\":10}]",
       "cache_ttl":60,"owner":"demo","beid":2}'
# 2) 注册应用（app_secret 仅此一次返回）
curl -s -XPOST http://127.0.0.1:8887/datasvc/app/register -H "Authorization: Bearer $TK" \
  -H 'Content-Type: application/json' -d '{"name":"dataease-app","rate_limit_qps":10,"beid":2}'
# 3) 授权 + 4) 发布
curl -s -XPOST http://127.0.0.1:8887/datasvc/grant   -H "Authorization: Bearer $TK" -H 'Content-Type: application/json' -d '{"api_id":1,"app_id":1,"beid":2}'
curl -s -XPOST http://127.0.0.1:8887/datasvc/api/publish -H "Authorization: Bearer $TK" -H 'Content-Type: application/json' -d '{"id":1}'
```

**签名机制（对照 `datasvc/rpc/internal/logic/credential.go`）**：请求体五字段
`app_key / ts / sign / path / params_json`，其中

- `ts`：unix 秒，与服务器偏差 **≤300 秒**；
- `sign`：`hex(hmac-sha256(key=app_secret, msg=app_key+ts))`；
- 返回：`{"rows_json":"[...]","cost_ms":12,"cache_hit":false,"affected":0}`（注意：该接口是唯一
  免 JWT 入口，返回体**没有** `{code,msg,data}` 包装壳）。

```bash
# 命令行验证一次（openssl 现算签名）
TS=$(date +%s)
SIGN=$(printf '%s' "${APP_KEY}${TS}" | openssl dgst -sha256 -hmac "$APP_SECRET" | awk '{print $NF}')
curl -s -XPOST http://127.0.0.1:8887/datasvc/execute -H 'Content-Type: application/json' \
  -d "{\"app_key\":\"$APP_KEY\",\"ts\":\"$TS\",\"sign\":\"$SIGN\",\"path\":\"/api/v1/demo/orders\",\"params_json\":\"{\\\"user_id\\\":19}\"}"
```

**DataEase v2 的 API 数据源配置**（v1.18.27 无内置 API 数据源，此配置仅适用于升级 v2 后；
v2 容器内端口变 8080，需同步调整 compose 端口映射）：

| 配置项 | 值 |
|--------|----|
| 数据源类型 | API |
| 请求方法 / URL | `POST` / `http://datacenter-api:8887/datasvc/execute`（容器网络内直连服务名） |
| Headers | `Content-Type: application/json` |
| 请求体 | `{"app_key":"<app_key>","ts":"<unix秒>","sign":"<hex签名>","path":"/api/v1/demo/orders","params_json":"{\"user_id\":19,\"limit\":10}"}` |
| 数据格式 / 取数 | JSON，从 `rows_json` 解析数组（JSONPath：`$.rows_json` 或按 v2 界面指引） |

**签名时效的工程判断（重要）**：ts ±300 秒强校验，而 DataEase 数据源的 Headers/Body 是**静态**
配置——ts/sign 会过期，仪表板自动刷新会在 5 分钟后开始报「签名不匹配」。三种落地方式按场景选：

1. **演示/验收窗口**：打开仪表板前现场算一次签名（上方 openssl 两行，或用前端
   「数据可视化 → 部署引导 → API 链路自测」现算现验），5 分钟内出图没问题；
2. **生产推荐：API 结果落表后直连**（回到路线 A）——用 job 内置同步引擎（Phase 8，源 SQL→目标表
   分批事务）或数据 IDE，把 API 背后的同源查询定时固化成快表，DataEase 直连快表；
   治理能力（限流/脱敏/审计）保留在平台侧的落表任务里；
3. **跨系统强隔离场景**：部署一个带动态签名的小代理（10 行级服务，转发时现算 ts/sign），
   DataEase 指向代理——平台侧无需改代码（本仓库不含该代理，按需自建）。

---

## 6. 前端「数据可视化」页面

`web/src/views/dataease/DataeaseEmbed.vue`（组件名 `dataease/DataeaseEmbed`，自包含、零新依赖、
**不改 router**），两种模式：

- **仪表板嵌入**：iframe 加载 DataEase 地址；建议填「仪表板 → 分享 → 开启公共链接」生成的免登录链接
  （嵌整站地址也可，在框内直接登录）。地址优先级：localStorage 覆盖 > `VITE_DATAEASE_URL` > 默认
  `http://localhost:8088`；构建期配置在 `web/.env.local` 加一行
  `VITE_DATAEASE_URL=http://localhost:8088` 即可；
- **部署引导**：未部署时展示一键 compose/建库/种子命令（复制即用，与本文档一致）+ API 签名链路自测
  （浏览器 WebCrypto 现算 HMAC，直调 `/datasvc/execute` 验证 app_key 链路，供路线 B 配置前自检）。

**菜单入口**（动态菜单按 用户→角色→菜单 链路下发，组件字符串 → `web/src/views/**/*.vue` 自动解析）：

```bash
# 幂等可重跑；先按脚本头部说明把 @beid 改成与现有菜单一致
docker compose exec -T mysql mysql -uroot -padmin datacenter < docs/dataease-menu-seed.sql
```

> 主线合入建议：在 `sys/model/seed.go` 的 `menuSeeds()` 里把「数据可视化」加为顶层菜单（与「仪表盘」同级）：
> `{Name: "数据可视化", Path: "/dataease", Component: "dataease/DataeaseEmbed", Icon: "DataBoard", MenuType: MenuTypeMenu, Sort: 11}`，
> 合入后 `docs/dataease-menu-seed.sql` 即可退役（种子幂等，两者共存也不冲突）。

---

## 7. 典型仪表板模板建议（对应演示数据）

三块模板均基于路线 A 的 SQL 数据集（`datacenter` 库演示表；`dc_*` 平台表随 schema 自带）。
每块给出：数据集 SQL → 视图类型 → 维度/指标。

### 7.1 数据源健康（对应 `dc_datasource` 登记表；调用日志 `dc_api_call_log` 需导入 phase7.sql）

| 视图 | 类型 | SQL 数据集 | 维度 / 指标 |
|------|------|-----------|------------|
| 数据源总览 | 指标卡 ×2 | `SELECT COUNT(*) AS total, COALESCE(SUM(status=1),0) AS enabled FROM dc_datasource WHERE deleted_at IS NULL` | total / enabled |
| 类型分布 | 饼图 | `SELECT type AS 类型, COUNT(*) AS 数量 FROM dc_datasource WHERE deleted_at IS NULL GROUP BY type` | 维度=类型，指标=数量 |
| API 调用健康（近 7 天） | 堆叠柱/折线 | `SELECT DATE(created_at) AS 日期, SUM(status='ok') AS 成功, SUM(status<>'ok') AS 失败 FROM dc_api_call_log GROUP BY DATE(created_at) ORDER BY 日期` | 维度=日期，指标=成功/失败 |

### 7.2 订单趋势（对应 `dm_orders` / `dm_order_items` / `dm_products`）

| 视图 | 类型 | SQL 数据集 | 维度 / 指标 |
|------|------|-----------|------------|
| 经营 KPI | 指标卡 ×3 | `SELECT COUNT(*) AS 订单数, SUM(amount) AS GMV, ROUND(SUM(amount)/COUNT(*),2) AS 客单价 FROM dm_orders WHERE status<>'cancelled' AND created_at >= DATE_SUB(CURDATE(), INTERVAL 30 DAY)` | 订单数 / GMV / 客单价 |
| 按日趋势 | 双轴折线 | `SELECT DATE(created_at) AS 日期, COUNT(*) AS 订单量, SUM(amount) AS GMV FROM dm_orders WHERE status<>'cancelled' GROUP BY DATE(created_at) ORDER BY 日期` | 维度=日期，指标=订单量/GMV |
| 状态占比 | 环形图 | `SELECT status AS 状态, COUNT(*) AS 单量 FROM dm_orders GROUP BY status` | 维度=状态，指标=单量 |
| 类目销售额 | 条形图 | `SELECT p.category AS 类目, SUM(i.qty*i.price) AS 销售额 FROM dm_order_items i JOIN dm_products p ON p.id=i.product_id JOIN dm_orders o ON o.id=i.order_id WHERE o.status<>'cancelled' GROUP BY p.category ORDER BY 销售额 DESC` | 维度=类目，指标=销售额 |

### 7.3 用户画像（对应 `dm_users`，可联 `dm_orders`）

| 视图 | 类型 | SQL 数据集 | 维度 / 指标 |
|------|------|-----------|------------|
| 大区分布 | 地图/饼图 | `SELECT province AS 大区, COUNT(*) AS 用户数 FROM dm_users GROUP BY province` | 维度=大区，指标=用户数 |
| 会员等级分布 | 柱形图 | `SELECT CONCAT('L', level) AS 等级, COUNT(*) AS 用户数 FROM dm_users GROUP BY level ORDER BY level` | 维度=等级，指标=用户数 |
| 注册趋势 | 折线 | `SELECT DATE_FORMAT(registered_at,'%Y-%m') AS 月份, COUNT(*) AS 注册数 FROM dm_users GROUP BY DATE_FORMAT(registered_at,'%Y-%m') ORDER BY 月份` | 维度=月份，指标=注册数 |
| 分层价值（ARPU） | 表格/柱形 | `SELECT CONCAT('L', u.level) AS 等级, COUNT(DISTINCT u.id) AS 用户数, ROUND(COALESCE(SUM(o.amount),0)/COUNT(DISTINCT u.id),2) AS ARPU FROM dm_users u LEFT JOIN dm_orders o ON o.user_id=u.id AND o.status<>'cancelled' GROUP BY u.level ORDER BY u.level` | 维度=等级，指标=用户数/ARPU |

---

## 8. 停止 / 卸载 / 迁移

```bash
docker compose -f deploy/dataease/docker-compose.dataease.yml stop      # 停止（不影响平台基座）
docker compose -f deploy/dataease/docker-compose.dataease.yml down      # 移除容器（数据保留）
```

- 持久化在仓库根 `./data/dataease/`（日志/静态资源/自定义驱动等）与平台 MySQL `dataease` 库；
  迁移机器 = 备份这两处 + 新机照第 1-2 步重来；
- 升级 v2：改 compose 的 image tag 与端口映射（v2 容器内 8080），元数据库由 v2 自动迁移
  （升级前务必备份 `dataease` 库）。

## 9. 常见问题

| 现象 | 原因与处理 |
|------|-----------|
| 容器反复重启，日志报连不上 `dataease` 库 | 未建元数据库：执行第 2 步建库后 `restart dataease` |
| `ps` 长时间 `health: starting` | 首启跑 Flyway 建表 1-3 分钟属正常；超过 5 分钟看 `logs -f dataease` |
| 启动报容器名冲突（`/datacenter-dataease` already in use） | 与平台编排 viz profile 重复部署：二选一，先 `docker compose --profile viz stop dataease` 或 `docker rm -f datacenter-dataease` |
| 网络报错（network datacenter_datacenter not found） | 平台编排未启动：`docker network ls` 核对真实网络名，用 `DATACENTER_NETWORK` 覆盖，或按第 0 节改独立模式 |
| 8088 端口被占 | `DATAEASE_PORT=8089 docker compose -f deploy/dataease/docker-compose.dataease.yml up -d` |
| 前端 iframe 空白，控制台报 X-Frame-Options / frame-ancestors 拒绝 | DataEase 官方镜像默认不加该头，通常来自自加的 nginx 反代：去掉 `X-Frame-Options` 或配置 `Content-Security-Policy: frame-ancestors 'self' http://<前端地址>` |
| 忘记 admin 密码 | 由管理员在「系统管理 → 用户管理」重置或新建账号；或改 `dataease.init_password` 后新建用户 |
| API 执行报「签名不匹配 / ts 过期」 | ts 必须为当前 unix 秒（±300s）、sign 按 `app_key+ts` 用 app_secret 计算；详见 DEMO.md 排错表 |
| 仪表板数据不更新 | 直连数据集默认实时查询；若开了「定时抽取」则按抽取计划更新，可改回直连或调 cron |
