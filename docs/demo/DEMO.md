# 数据中台演示剧本（电商演示数据包）

> 配套数据：[demo_data.sql](demo_data.sql)。本文是 15 分钟标准演示动线，每步给出
> curl 命令（或 grpcurl / 前端点击路径）与**预期输出要点**，照着敲即可走完全程。
>
> 演示登录账号：**`13800138000` / `123456`**；演示站点（多租户 beid）：**`2`（演示电商平台）**。

---

## 0. 前置准备（一次性，约 10 分钟）

### 0.1 起依赖中间件

```bash
cd /home/jackluo/data/datacenter
sh server.sh        # 起 mysql(3306, root/admin) / redis(6379, 密码 admin) / etcd(2379) / es(9200, elastic/admin)
```

> 注意：`server.sh` 会先 `docker rm -f` 掉名为 `mysql/redis/etcd/es` 的旧容器再重建。

### 0.2 建库并导入 schema + 演示数据

`demo_data.sql` 内部**不含 CREATE DATABASE**（调用方先建库），且**不写 USE**（跟谁 source 进哪个库就建在哪个库）：

```bash
mysql -uroot -padmin
```

```sql
CREATE DATABASE IF NOT EXISTS datacenter DEFAULT CHARSET utf8mb4;
USE datacenter;
SOURCE sql.sql;                        -- 平台基座表（base_member / base_app 等）
SOURCE docs/schema/phase1.sql;         -- Phase 1（数据源/元数据/数据服务/资产目录）——以下按顺序依次 source
SOURCE docs/schema/phase2.sql;         -- Phase 2（质量/指标/标签）
SOURCE docs/schema/phase3.sql;         -- Phase 3/4（调度/IDE/AI 问数审计）
SOURCE docs/schema/phase5.sql;         -- Phase 5（标准/主数据/运维日志；标准核对与主数据演示需要）
SOURCE docs/schema/phase6.sql;         -- Phase 6（数仓建模/安全/标签圈群；圈群演示需要）
SOURCE docs/demo/demo_data.sql;        -- 演示数据包（本仓库新增）
```

导入成功的自检（可选）：

```sql
SELECT COUNT(*) FROM dm_users;         -- 500
SELECT COUNT(*) FROM dm_products;      -- 80
SELECT COUNT(*) FROM dm_orders;        -- 1500
SELECT COUNT(*) FROM dm_order_items;   -- 2500
SELECT province, COUNT(*) FROM dm_users GROUP BY province ORDER BY 2 DESC LIMIT 1;   -- 华东 200
SELECT SUM(amount) FROM dm_orders;     -- 约 338 万（随演示日小幅变化）
```

### 0.3 各服务配置文件

仓库各 `rpc/etc/` 下已有现成 `*.yaml`（与 `*.example.yaml` 同目录）；若从零部署，参照 README
把每个 `*.example.yaml` 复制为正式名并核对 mysql/redis/etcd 连接信息即可。

### 0.4 编译并启动全部服务

```bash
sh restart.sh       # 编译 + 拉起网关与各 RPC，网关端口 8887
tail -F nohup.out   # 看到 "Starting server at 0.0.0.0:8887..." 即网关就绪
```

> 演示全程网关入口统一为 `http://127.0.0.1:8887`；管理面接口都要带 JWT（见步骤 1），
> 唯一免 JWT 的是 `POST /datasvc/execute`（走 app_key + 签名，见步骤 6）。

---

## 1.（第 1 分钟）登录拿 token

```bash
curl -s -XPOST http://127.0.0.1:8887/user/login \
  -H 'Content-Type: application/json' \
  -d '{"mobile":"13800138000","type":1,"password":"123456"}'
```

**预期输出要点**：

```json
{"uid":1,"username":"demo","mobile":"13800138000",
 "access_token":"eyJhbGciOi...","access_expire":178...,"refresh_after":...}
```

保存 token 供后续所有管理面接口使用：

```bash
TK=$(curl -s -XPOST http://127.0.0.1:8887/user/login -d '{"mobile":"13800138000","type":1,"password":"123456"}' | sed 's/.*"access_token":"\([^"]*\)".*/\1/')
echo $TK    # 非空即成功；后续请求统一带 -H "Authorization: Bearer $TK"
```

> 排错：报"用户未注册/密码错误"→ 确认步骤 0.2 里 `SELECT * FROM base_member WHERE mobile='13800138000';`
> 有记录且 password 列为明文 `123456`（登录逻辑是明文比对）。

## 2.（第 2-4 分钟）登记演示库数据源 → 采集元数据

```bash
# 2.1 登记演示库（演示数据就建在 datacenter 库）
curl -s -XPOST http://127.0.0.1:8887/datasource/create \
  -H "Authorization: Bearer $TK" -H 'Content-Type: application/json' \
  -d '{"name":"演示电商平台业务库","type":"mysql","host":"127.0.0.1","port":3306,
       "db_name":"datacenter","username":"root","password":"admin","owner":"演示用户","beid":2}'
# 预期：返回 {"id":1,...}，记住该 id（下文以 ds_id=1 为例；若非 1 请替换）

# 2.2 连接探活
curl -s "http://127.0.0.1:8887/datasource/test?id=1" -H "Authorization: Bearer $TK"
# 预期：ok=true（或 status 正常 + 时延毫秒数）

# 2.3 采集元数据（整库采集，dm_* 演示表随平台表一起入资产）
curl -s -XPOST http://127.0.0.1:8887/metadata/collect \
  -H "Authorization: Bearer $TK" -H 'Content-Type: application/json' \
  -d '{"ds_id":1,"beid":2}'
# 预期：{"tables_created":N,"columns_created":M}（N 为库内全部表数，>=4 张 dm_* 表）

# 2.4 查看入资产的演示表（关键字过滤）
curl -s "http://127.0.0.1:8887/metadata/entity/list?beid=2&keyword=dm&page=1&size=10" \
  -H "Authorization: Bearer $TK"
```

**预期输出要点**：list 中可见 `dm_users(500 行)/dm_products(80)/dm_orders(1500)/dm_order_items(2500)`，
每条带 `fqn=datacenter.dm_xxx`、`comment` 即建表注释（演示口径写在注释里，资产目录直接可读）。

> 前端点击路径：数据源管理 → 新建数据源（填同上表单）→ 测试连接 → 元数据 → 立即采集 → 元数据列表搜 `dm`。

## 3.（第 5-7 分钟）AI 问数三连（POST /ai/ask）

```bash
ask() { curl -s -XPOST http://127.0.0.1:8887/ai/ask \
  -H "Authorization: Bearer $TK" -H 'Content-Type: application/json' -d "$1"; echo; }

ask '{"question":"最近90天每天订单量","beid":2}'
ask '{"question":"哪个省份用户最多","beid":2}'
ask '{"question":"总销售额是多少","beid":2}'
```

**预期输出要点**（`candidates[].type/name`）：

| 问句 | 应命中候选 | 说明 |
|---|---|---|
| 最近90天每天订单量 | `entity: dm_orders` | 命中表注释中的"订单量"关键词 |
| 哪个省份用户最多 | `entity: dm_users` | 命中"用户/省份" |
| 总销售额是多少 | `entity: dm_orders` | 命中"销售额"（分词 bigram"销售"） |

- `llm_enabled=false` 时（未配 LLM），`answer` 为检索摘要（候选清单 + 口径引导），仍能演示"问数命中了哪张表"；
- `llm_enabled=true` 时（`ai/rpc/etc/ai.yaml` 配好 LLM），额外返回 `generated_sql`，例如
  `SELECT DATE(created_at) AS dt, COUNT(*) AS order_cnt FROM datacenter.dm_orders WHERE ... GROUP BY dt`；
- 想看"真跑出数字"，用问一问执行版（auto_run 走 IDE 安全闸）：

```bash
curl -s -XPOST http://127.0.0.1:8887/ai/ask_and_run \
  -H "Authorization: Bearer $TK" -H 'Content-Type: application/json' \
  -d '{"question":"最近90天每天订单量","beid":2,"auto_run":true,"datasource_id":1,"max_rows":100}'
# 预期：rows_json 含近 90 天逐日订单量（平常日约 15 单/天，两个大促峰值日——昨天与 8 天前——各约 121 单）
```

> 手工 SQL 对照（可在 mysql 里同步演示）：
> `SELECT DATE(created_at) d, COUNT(*) FROM dm_orders WHERE created_at >= DATE_SUB(CURDATE(), INTERVAL 90 DAY) GROUP BY d ORDER BY d;`
> `SELECT province, COUNT(*) FROM dm_users GROUP BY province ORDER BY 2 DESC;`（华东 200 第一）
> `SELECT ROUND(SUM(amount),2) FROM dm_orders;`（约 338 万）

## 4.（第 8 分钟）数据质量：not_null 规则 + 执行 + 看结果

```bash
# 4.1 建规则：dm_orders.amount 非空
curl -s -XPOST http://127.0.0.1:8887/quality/rule/create \
  -H "Authorization: Bearer $TK" -H 'Content-Type: application/json' \
  -d '{"name":"订单金额非空","rule_type":"not_null","target_table":"dm_orders","target_column":"amount","owner":"演示用户","beid":2}'
# 预期：返回 {"id":1,...}（rule id，下文 rule_ids 用它）

# 4.2 建质量任务（绑定数据源 + 规则）
curl -s -XPOST http://127.0.0.1:8887/quality/job/create \
  -H "Authorization: Bearer $TK" -H 'Content-Type: application/json' \
  -d '{"name":"电商演示质量任务","datasource_id":1,"db_name":"datacenter","rule_ids":"1"}'
# 预期：返回 {"id":1,...}（job id）

# 4.3 执行
curl -s -XPOST http://127.0.0.1:8887/quality/job/run \
  -H "Authorization: Bearer $TK" -H 'Content-Type: application/json' -d '{"id":1}'
# 预期：{"job_id":1,"total_rules":1,"passed":1,"last_status":"success"}

# 4.4 看结果明细
curl -s "http://127.0.0.1:8887/quality/result/list?job_id=1" -H "Authorization: Bearer $TK"
# 预期：pass_flag=1、异常行数=0
```

**加分项（演示"不通过"长什么样）**——再建一条必然失败的值域规则后重跑任务：

```bash
curl -s -XPOST http://127.0.0.1:8887/quality/rule/create \
  -H "Authorization: Bearer $TK" -H 'Content-Type: application/json' \
  -d '{"name":"用户等级必须为5(故意失败演示)","rule_type":"value_range","target_table":"dm_users","target_column":"level","params_json":"{\"min\":5,\"max\":5}","beid":2}'
curl -s -XPOST http://127.0.0.1:8887/quality/job/run \
  -H "Authorization: Bearer $TK" -H 'Content-Type: application/json' -d '{"id":1}'
# 预期：last_status=partial（1 通过 1 失败），result/list 中失败规则 pass_flag=0、异常行数≈449（非 L5 的 449 人）
```

## 5.（第 9-10 分钟）标签圈群：dm_users 建"高价值用户"标签 → 圈群 → 成员数

### 5.1 HTTP：建标签对象 → 建标签 → 打标（网关 `/tag/*`）

```bash
# 5.1.1 建标签对象（绑定演示表 dm_users，主键 id）
curl -s -XPOST http://127.0.0.1:8887/tag/object/create \
  -H "Authorization: Bearer $TK" -H 'Content-Type: application/json' \
  -d '{"object_type":"user","name":"电商用户","ds_id":1,"db_name":"datacenter","table_name":"dm_users","pk_column":"id","beid":2}'
# 预期：返回 {"id":1,...}（object_id）

# 5.1.2 建"高价值用户"规则标签（rule_sql 必须含 as object_id / as tag_value 两别名、无分号）
curl -s -XPOST http://127.0.0.1:8887/tag/def/create \
  -H "Authorization: Bearer $TK" -H 'Content-Type: application/json' \
  -d '{"object_id":1,"tag_key":"high_value","tag_name":"高价值用户","tag_type":"rule",
       "rule_sql":"select id as object_id, level as tag_value from dm_users where level >= 4","beid":2}'
# 预期：返回 {"id":1,...}（tag_id）

# 5.1.3 执行打标（在目标库真实跑 rule_sql 并写 dc_tag_value）
curl -s -XPOST http://127.0.0.1:8887/tag/tagging \
  -H "Authorization: Bearer $TK" -H 'Content-Type: application/json' -d '{"tag_id":1,"object_id":1,"beid":2}'
# 预期：{"tagged_rows":151,"tag_id":1} —— 高价值用户(L4 100 人 + L5 51 人)共 151

# 5.1.4 抽查单个成员画像（object_pk=成员主键值；uid=2 是 L4，应命中高价值）
curl -s "http://127.0.0.1:8887/tag/object/tags?object_id=1&object_pk=2&beid=2" -H "Authorization: Bearer $TK"
# 预期：tags=[{tag_key:"high_value",tag_name:"高价值用户",tag_value:"4"}]
# （换个非高价值用户试，如 object_pk=1（L2），预期 tags=[] 空数组）
```

### 5.2 圈群（Phase 6 tag-rpc 能力，网关暂未挂 HTTP 路由，用 grpcurl 直连 tag RPC :8091）

```bash
# grpcurl 安装： go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest
# 5.2.1 建人群（圈选 SQL 必须以 select 开头且含 as object_id 别名）
grpcurl -plaintext -d '{"name":"高价值用户人群","object_id":1,
  "circle_rule_sql":"select id as object_id from dm_users where level >= 4","beid":2}' \
  127.0.0.1:8091 tag.Tag/CreateGroup
# 预期：返回人群 id（下文 group_id=1）

# 5.2.2 执行圈选（跑 SQL 物化成员到 dc_tag_group_member_1）
grpcurl -plaintext -d '{"group_id":1,"beid":2}' 127.0.0.1:8091 tag.Tag/RunCircle
# 预期：{"member_count":151,"msg":"..."} —— 成员数 151

# 5.2.3 分页看成员
grpcurl -plaintext -d '{"group_id":1,"beid":2,"page":1,"size":10}' 127.0.0.1:8091 tag.Tag/ListGroupMembers
# 预期：list 为 member 主键(1-500 中的 151 个) + tag_value(等级)
```

> 没带 grpcurl 时，HTTP 打标返回的 `tagged_rows=151` 就是同一口径的成员数，可直接收尾本节。

## 6.（第 11-13 分钟）数据服务：向导建 API → 注册应用 → 授权 → 发布 → 签名执行

```bash
# 6.1 向导建"查询订单"API（mode=wizard 强制 SELECT 模板；占位符 :name；params_json 为参数定义数组）
curl -s -XPOST http://127.0.0.1:8887/datasvc/api/create \
  -H "Authorization: Bearer $TK" -H 'Content-Type: application/json' \
  -d '{"name":"查询订单","path":"/api/v1/demo/orders","mode":"wizard","datasource_id":1,
       "sql_text":"select order_no, amount, status, created_at from dm_orders where user_id = :user_id order by created_at desc limit :limit",
       "params_json":"[{\"name\":\"user_id\",\"type\":\"number\",\"required\":true},{\"name\":\"limit\",\"type\":\"number\",\"required\":false,\"default\":10}]",
       "cache_ttl":60,"owner":"演示用户","beid":2}'
# 预期：返回 {"id":1,...}（api id）
# 注意：sql_text 里不要用 DATE_FORMAT(x,'%H:%i:%s') 之类带冒号的格式串，:%i 会被当成未声明占位符拒绝

# 6.2 注册应用（app_secret 仅此一次返回，演示时抄下来）
curl -s -XPOST http://127.0.0.1:8887/datasvc/app/register \
  -H "Authorization: Bearer $TK" -H 'Content-Type: application/json' \
  -d '{"name":"demo-app","rate_limit_qps":10,"beid":2}'
# 预期：{"app_id":1,"app_key":"...","app_secret":"..."}
APP_ID=1; APP_KEY=<上一步 app_key>; APP_SECRET=<上一步 app_secret>

# 6.3 授权：把 API 1 授给应用 1（可加 field_mask 脱敏，如 "amount"）
curl -s -XPOST http://127.0.0.1:8887/datasvc/grant \
  -H "Authorization: Bearer $TK" -H 'Content-Type: application/json' \
  -d '{"api_id":1,"app_id":1,"beid":2}'
# 预期：{} 无错误

# 6.4 发布 API
curl -s -XPOST http://127.0.0.1:8887/datasvc/api/publish \
  -H "Authorization: Bearer $TK" -H 'Content-Type: application/json' -d '{"id":1}'
# 预期：{} 无错误（status=published）

# 6.5 免 JWT 签名执行（sign = hex(hmac-sha256(key=app_secret, msg=app_key+ts))，ts 与服务器偏差 <=300s）
TS=$(date +%s)
SIGN=$(printf '%s' "${APP_KEY}${TS}" | openssl dgst -sha256 -hmac "$APP_SECRET" | awk '{print $NF}')
curl -s -XPOST http://127.0.0.1:8887/datasvc/execute \
  -H 'Content-Type: application/json' \
  -d "{\"app_key\":\"$APP_KEY\",\"ts\":\"$TS\",\"sign\":\"$SIGN\",\"path\":\"/api/v1/demo/orders\",\"params_json\":\"{\\\"user_id\\\":19}\"}"
# 预期：{"rows_json":"[{\"order_no\":\"ORD2026...\",\"amount\":...}, ...]","cost_ms":...,"cache_hit":false,"affected":0}
#       rows_json 为 user_id=19 名下 3 笔订单（每用户恰 3 单）；再执行一次 cache_hit=true（60 秒结果缓存）
```

> 前端点击路径：数据服务 → API 向导（选数据源/写 SQL/配参数）→ 应用管理注册 → 授权 → 发布 → 接口测试。
> 反向演示：改坏 sign 或复用 6 分钟前的 ts → 预期返回"签名不匹配/ts 过期"。

## 7.（第 14-15 分钟）标准核对 / 主数据 / 物化视图（各一步点到即止）

### 7.1 数据标准核对（登记数据元 → 建标准模型 → 发布 → 核对 dm_orders）

```bash
# 登记两个数据元
curl -s -XPOST http://127.0.0.1:8887/standard/element/save -H "Authorization: Bearer $TK" \
  -H 'Content-Type: application/json' \
  -d '{"element_code":"order_no","element_name":"订单号","data_type":"varchar","data_length":32,"beid":2}'
curl -s -XPOST http://127.0.0.1:8887/standard/element/save -H "Authorization: Bearer $TK" \
  -H 'Content-Type: application/json' \
  -d '{"element_code":"amount","element_name":"金额","data_type":"decimal","data_precision":10,"beid":2}'
# 建标准模型并发布
curl -s -XPOST http://127.0.0.1:8887/standard/model/save -H "Authorization: Bearer $TK" \
  -H 'Content-Type: application/json' \
  -d '{"model_code":"dm_orders","model_name":"演示订单表标准模型",
       "elements_json":"[{\"element_code\":\"order_no\",\"required\":true},{\"element_code\":\"amount\",\"required\":true}]","beid":2}'
curl -s -XPOST http://127.0.0.1:8887/standard/model/publish -H "Authorization: Bearer $TK" \
  -H 'Content-Type: application/json' -d '{"id":1}'
# 核对（读 information_schema 实际列 vs 标准模型）
curl -s "http://127.0.0.1:8887/standard/check?model_id=1&datasource_id=1&db_name=datacenter&table_name=dm_orders" \
  -H "Authorization: Bearer $TK"
```

**预期输出要点**：`report.diff_json` 给出实际列与标准的差异清单（如多出的列 / 类型偏差），
`pass_flag` 汇总通过与否——现场讲"标准是尺子，核对是体检"。

### 7.2 主数据（建模型 → 录一条会员主数据 → 看自动码表）

```bash
curl -s -XPOST http://127.0.0.1:8887/mdm/model/save -H "Authorization: Bearer $TK" \
  -H 'Content-Type: application/json' \
  -d '{"model_code":"member","model_name":"会员主数据","table_name":"dm_users",
       "fields_json":"[{\"name\":\"user_name\",\"type\":\"string\"},{\"name\":\"level\",\"type\":\"number\"}]","beid":2}'
curl -s -XPOST http://127.0.0.1:8887/mdm/record/save -H "Authorization: Bearer $TK" \
  -H 'Content-Type: application/json' \
  -d '{"model_id":1,"data_json":"{\"user_name\":\"139****0001\",\"level\":5}","beid":2}'
# 预期：返回 {"id":1,"biz_code":"MEM...","status":"draft",...} —— biz_code 即码段自动生成的业务编码
curl -s "http://127.0.0.1:8887/mdm/record/list?beid=2&model_id=1&page=1&size=10" -H "Authorization: Bearer $TK"
```

### 7.3 物化视图（Phase 6 theme-rpc，网关暂未挂 HTTP 路由；grpcurl 直连 theme RPC :8080）

```bash
# 先手工起 theme RPC（Phase 6 骨架暂未纳入 restart.sh；8080 被占时改 theme/rpc/etc/theme.yaml 的 ListenOn）
cd theme/rpc && go build -o theme-server theme.go && nohup ./theme-server -f etc/theme.yaml &

# 保存物化视图定义（按 dm_orders 按天汇总）
grpcurl -plaintext -d '{"name":"订单日汇总","datasource_id":1,"db_name":"datacenter",
  "view_name":"dm_mv_order_daily",
  "select_sql":"select date(created_at) as dt, count(*) as order_cnt, sum(amount) as gmv from dm_orders group by dt",
  "refresh_mode":"manual","beid":2}' 127.0.0.1:8080 theme.Theme/SaveMaterializedView

# 刷新（真实执行 drop+create as select）
grpcurl -plaintext -d '{"id":1}' 127.0.0.1:8080 theme.Theme/RefreshMaterializedView
# 预期：{"ok":true,"msg":"...","cost_ms":<毫秒>}

# 在 MySQL 里直接看到刷新结果
mysql -uroot -padmin datacenter -e "SELECT * FROM dm_mv_order_daily ORDER BY dt DESC LIMIT 5;"
```

---

## 附录 A：演示数据口径速查（讲给评委听的数字）

| 表 | 行数 | 口径 |
|---|---|---|
| dm_users | 500 | 省份：华东 200 / 华北 100 / 华南 75 / 西南 50 / 华中 50 / 东北 25（"哪个省份用户最多"→华东）；等级：L1 74 / L2 126 / L3 149 / L4 100 / L5 51（**高价值 level>=4 = 151 人**）；注册时间 2023-2025（30%/40%/30%）；user_name 为手机号掩码（脱敏演示点） |
| dm_products | 80 | 数码/家居/美妆/服饰各 20 款，价格 .99 结尾，库存 50-999 |
| dm_orders | 1500 | 状态：completed 675 / paid 375 / shipped 300 / cancelled 150；近 90 天**每天都有订单**（平常日约 15 单/天，两个大促峰值日——昨天与 8 天前——各约 121 单）；总金额约 **338 万**（剔除 cancelled 约 **306 万**，随演示日小幅浮动）；order_no=ORD+日期+当日序号 |
| dm_order_items | 2500 | 每单至少 1 条明细（前 1000 单各 2 条）；明细金额之和与订单 amount **严格一致**；含 9 折促销价快照 |
| base_member | +1 | 演示账号 `13800138000 / 123456`（username=demo） |
| base_app | +1 | id=2「演示电商平台」，即全篇 beid=2 |

数据全部由递归 CTE + 确定性算术伪随机生成（无 RAND()），重复 source 结果一致；
脚本可重复执行（先 DROP 演示表、清插演示账号）。

## 附录 B：常见问题

- **登录 401/未注册**：步骤 0.2 漏了 source；或 demo_data.sql source 进了别的库。
- **采集/问数/质量查不到 dm_ 表**：`beid` 不一致——演示统一用 beid=2，各接口记得带上。
- **AI 问数 candidates 为空**：先做步骤 2 的元数据采集（候选来自元数据实体+指标）；LLM 未配置时 `llm_enabled=false` 属正常降级。
- **/datasvc/execute 报签名不匹配**：ts 必须是当前 unix 秒（偏差 ≤300s）、sign 按 `app_key+ts` 用 app_secret 计算。
- **8887 起不来**：`tail -F nohup.out` 看网关日志；常见是 /etc/datacenter-api.yaml 缺失或 mysql/etcd 未就绪。
