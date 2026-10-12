# 数据中台演示剧本（电商演示数据包）

> 配套数据：[demo_data.sql](demo_data.sql)、一键脚本：[DEMO_AUTO.sh](DEMO_AUTO.sh)。
> 本文给出两条动线：**一键演示（约 2 分钟准备 + 1 分钟执行）**与 **15 分钟手动动线**，
> 每步给出 curl 命令（或 grpcurl / 前端点击路径）与**预期输出要点**。
>
> **双账号体系**（注意区分）：
> - **管理台（sys_user）**：`admin` / `admin123`，走 `POST /sys/login`，控制台 http://localhost:5173 用它登录；
> - **C 端会员（base_member）**：`13800138000` / `123456`，走 `POST /user/login`（需 user RPC 在线），
>   归属演示站点（多租户 beid）**`2`（演示电商平台）**。
> 全篇管理面接口以 `beid=2` 调用。

---

## 0. 一键演示（推荐，销售 10 分钟内出效果）

### 0.0 前置

一次性完成依赖与数据准备（已做过可跳过）：

```bash
cd /home/jackluo/data/datacenter
sh server.sh        # 起 mysql(3306, root/admin) / redis(16379, 密码 admin) / etcd(2379) / es(9200)
mysql -uroot -padmin   # 进入后依次：
#   CREATE DATABASE IF NOT EXISTS datacenter DEFAULT CHARSET utf8mb4; USE datacenter;
#   SOURCE sql.sql;  SOURCE sys.sql;                  -- 平台基座 + RBAC（admin 账号在 sys.sql）
#   SOURCE docs/schema/phase1.sql; ... 依次 phase2/3/5/6/7;
#   SOURCE docs/demo/demo_data.sql;                  -- 演示数据（含大促/异常/脏数据故事）
sh restart.sh       # 编译 + 拉起网关与各 RPC，网关端口 8887
```

### 0.1 一键跑通

```bash
bash docs/demo/DEMO_AUTO.sh
```

脚本自动完成 11 步并**彩色打印每步结果摘要**：

| 步骤 | 内容 | 预期输出要点 |
|---|---|---|
| 1 | 前置检查 | 网关可达、MySQL 可连、签名能力（openssl/python3）可用 |
| 2 | 管理台登录 | admin/admin123 换取 JWT；打印双账号体系说明 |
| 3 | 登记演示数据源 | 按名称幂等复用/新建，探活 `{"ok":true,...}` |
| 4 | 采集元数据 | `tables_created/columns_created`（重复执行为 0=幂等）；dm_ 表 5 张 |
| 5 | AI 问数三连 | 每问打印命中候选与回答摘要：订单量→dm_orders、销售额→dm_orders、用户反馈→dm_user_feedback |
| 6 | 质量规则与执行 | not_null+value_range 双规则，任务 `1/2（50%）`、`last_status=partial`，value_range 实测 12 |
| 7 | 脏数据错误捕获 | `amount<0` 共 12 行；系统留存错误样本 12 行（如 `ORD20261008BAD01 amount=-1.00`） |
| 8 | 数据服务签名调用 | 建 API→注册应用→授权→发布→**HMAC 签名调用**，打印返回 3 行真实订单 |
| 9 | 标签打标 | 高价值用户标签 `tagged_rows=151` + 单人画像抽查 |
| 10 | 人群圈选 | grpcurl 直连 tag RPC 圈选；命中已知兼容问题时自动降级为打标口径（tagged_rows=151） |
| 11 | 主数据 | NextCode 取号（`MEMBER20261012xxxxxx`）→ 录记录落库 |

收尾打印：**`演示完成，打开 http://localhost:5173 体验控制台（管理台 admin/admin123）`**。

脚本特性：
- 每步失败**不中断**（黄色 warning 继续）；仅前置检查/登录/数据源失败红色退出；
- 可重复执行：数据源/标签对象/标签/人群/主数据模型按名称复用；质量规则/任务、API、应用、主数据记录按次新建；
- 环境变量可覆盖：`GW`（网关）、`MYSQL_HOST/PORT/USER/PWD/DB`、`MYSQL_CONTAINER`（宿主机无 mysql 命令时 docker exec 兜底）、`BEID`。

> 10 分钟讲法建议：先开控制台登录 → 跑脚本 → 对着脚本每步输出在控制台里翻对应页面
> （资产目录看 dm_orders 注释 / 质量看 partial 任务与错误样本 / 数据服务看调用日志）。

---

## 1. 15 分钟手动动线

## 1.（第 1 分钟）登录拿 token（双体系）

管理台（控制台/管理面接口用）：

```bash
curl -s -XPOST http://127.0.0.1:8887/sys/login \
  -H 'Content-Type: application/json' -d '{"username":"admin","password":"admin123"}'
# 预期：{"token":"eyJ...","uid":1,"username":"admin","roles":["admin"],...}
TK=$(curl -s -XPOST http://127.0.0.1:8887/sys/login -H 'Content-Type: application/json' \
  -d '{"username":"admin","password":"admin123"}' | sed 's/.*"token":"\([^"]*\)".*/\1/')
echo $TK    # 非空即成功；后续请求统一带 -H "Authorization: Bearer $TK"
```

C 端会员（演示"多租户"叙事时用，需 user RPC 在线）：

```bash
curl -s -XPOST http://127.0.0.1:8887/user/login \
  -H 'Content-Type: application/json' -d '{"mobile":"13800138000","type":1,"password":"123456"}'
# 预期：{"uid":1,"username":"demo","mobile":"13800138000","access_token":"eyJ..."}
```

> 排错：
> - `/sys/login` 401/失败 → 确认 `sys.sql` 已 source，`SELECT username,password FROM sys_user;` 有 admin（bcrypt 密文）；
> - `/user/login` 报 "produced zero addresses" → user RPC 未启动，管理面演示不受影响。

## 2.（第 2-4 分钟）登记演示库数据源 → 采集元数据

```bash
# 2.1 登记演示库（演示数据就建在 datacenter 库）
curl -s -XPOST http://127.0.0.1:8887/datasource/create \
  -H "Authorization: Bearer $TK" -H 'Content-Type: application/json' \
  -d '{"name":"演示电商平台业务库","type":"mysql","host":"127.0.0.1","port":3306,
       "db_name":"datacenter","username":"root","password":"admin","owner":"演示用户","beid":2}'
# 预期：返回 {"id":N,...}，记住该 id（下文以 ds_id=1 为例；若非 1 请替换）

# 2.2 连接探活（注意带 beid）
curl -s "http://127.0.0.1:8887/datasource/test?id=1&beid=2" -H "Authorization: Bearer $TK"
# 预期：{"ok":true,"latency_ms":1,"err_msg":""}

# 2.3 采集元数据（整库采集，dm_* 演示表随平台表一起入资产；db_name 必填）
curl -s -XPOST http://127.0.0.1:8887/metadata/collect \
  -H "Authorization: Bearer $TK" -H 'Content-Type: application/json' \
  -d '{"ds_id":1,"beid":2,"db_name":"datacenter"}'
# 预期：{"tables_created":N,"columns_created":M}（重复采集按 fqn 幂等，无新增时为 0）

# 2.4 查看入资产的演示表（关键字过滤）
curl -s "http://127.0.0.1:8887/metadata/entity/list?beid=2&keyword=dm&page=1&size=10" \
  -H "Authorization: Bearer $TK"
```

**预期输出要点**：list 中可见 `dm_users(500 行)/dm_products(80)/dm_orders(1500)/dm_order_items(2500)/dm_user_feedback(200)`，
每条带 `fqn=datacenter.dm_xxx`、`comment` 即建表注释（**演示口径关键词已埋进注释**：
订单量/销售额/用户画像/用户反馈/好评率——AI 问数语义检索靠它命中，资产目录直接可读）。

> 前端点击路径：数据源管理 → 新建数据源（填同上表单）→ 测试连接 → 元数据 → 立即采集 → 元数据列表搜 `dm`。

## 3.（第 5-7 分钟）AI 问数四连（POST /ai/ask）

```bash
ask() { curl -s -XPOST http://127.0.0.1:8887/ai/ask \
  -H "Authorization: Bearer $TK" -H 'Content-Type: application/json' -d "$1"; echo; }

ask '{"question":"最近90天每天订单量","beid":2}'
ask '{"question":"哪个省份用户最多","beid":2}'
ask '{"question":"总销售额是多少","beid":2}'
ask '{"question":"用户反馈怎么样","beid":2}'
```

**预期输出要点**（`candidates[].type/name`）：

| 问句 | 应命中候选 | 说明 |
|---|---|---|
| 最近90天每天订单量 | `entity: dm_orders` | 命中表注释中的"订单量"关键词 |
| 哪个省份用户最多 | `entity: dm_users` | 命中"用户/省份" |
| 总销售额是多少 | `entity: dm_orders` | 命中"销售额"（分词 bigram"销售"） |
| 用户反馈怎么样 | `entity: dm_user_feedback` | 命中"用户反馈"（新反馈表演示） |

- `llm_enabled=false` 时（未配 LLM），`answer` 为检索摘要（候选清单 + 口径引导），仍能演示"问数命中了哪张表"；
- `llm_enabled=true` 时（`ai/rpc/etc/ai.yaml` 配好 LLM），额外返回 `generated_sql`；
- 想看"真跑出数字"，用问一问执行版（auto_run 走 IDE 安全闸）：

```bash
curl -s -XPOST http://127.0.0.1:8887/ai/ask_and_run \
  -H "Authorization: Bearer $TK" -H 'Content-Type: application/json' \
  -d '{"question":"最近90天每天订单量","beid":2,"auto_run":true,"datasource_id":1,"max_rows":100}'
# 预期：rows_json 含近 90 天逐日订单量——三段故事曲线：
#   平常日约 16 单/天；60 天前大促峰值日 48 单（3 倍，且集中在 20-24 点晚高峰）；
#   每周三会员日 23 单；3/4/5 天前异常骤降到 8 单/天（-50%，其中还混着 4 笔 -1 元脏单）
```

> 手工 SQL 对照（可在 mysql 里同步演示）：
> `SELECT DATE(created_at) d, COUNT(*) FROM dm_orders WHERE created_at >= DATE_SUB(CURDATE(), INTERVAL 90 DAY) GROUP BY d ORDER BY d;`
> `SELECT province, COUNT(*) FROM dm_users GROUP BY province ORDER BY 2 DESC;`（华东 200 第一）
> `SELECT COUNT(*), SUM(amount) FROM dm_orders WHERE amount < 0;`（12 行脏单，故事钩子）

## 4.（第 8 分钟）数据质量：not_null + value_range 双规则 + 执行 + 错误数据捕获

```bash
# 4.1 建规则1：dm_orders.order_no 非空（应通过）
curl -s -XPOST http://127.0.0.1:8887/quality/rule/create \
  -H "Authorization: Bearer $TK" -H 'Content-Type: application/json' \
  -d '{"name":"订单号非空","rule_type":"not_null","target_table":"dm_orders","target_column":"order_no","owner":"演示用户","beid":2}'

# 4.2 建规则2：dm_orders.amount 值域 [0,1000000]（命中 12 行 -1 元脏单，应失败）
curl -s -XPOST http://127.0.0.1:8887/quality/rule/create \
  -H "Authorization: Bearer $TK" -H 'Content-Type: application/json' \
  -d '{"name":"订单金额值域0-100万","rule_type":"value_range","target_table":"dm_orders","target_column":"amount","params_json":"{\"min\":0,\"max\":1000000}","beid":2}'
# 预期：两条规则各返回 {"id":M,...}（rule id，下文 rule_ids 用它们）

# 4.3 建质量任务（绑定数据源 + 规则）
curl -s -XPOST http://127.0.0.1:8887/quality/job/create \
  -H "Authorization: Bearer $TK" -H 'Content-Type: application/json' \
  -d '{"name":"电商演示质量任务","datasource_id":1,"db_name":"datacenter","rule_ids":"1,2"}'

# 4.4 执行（注意带 beid）
curl -s -XPOST http://127.0.0.1:8887/quality/job/run \
  -H "Authorization: Bearer $TK" -H 'Content-Type: application/json' -d '{"id":1,"beid":2}'
# 预期：{"job_id":1,"total_rules":2,"passed":1,"last_status":"partial"} —— 通过率 50%

# 4.5 看结果明细
curl -s "http://127.0.0.1:8887/quality/result/list?job_id=1" -H "Authorization: Bearer $TK"
# 预期：not_null pass_flag=1 实测 0；value_range pass_flag=0 实测 12（12 行脏单被点名）
```

**错误数据捕获（Phase 7 能力，讲"系统不仅发现，还留了证据"）**：
执行失败规则自动抓取违规行样本写入 `dc_quality_error_data`（上限 100 行），数据库直查：

```sql
SELECT rule_id, sample_count, JSON_EXTRACT(rows_json,'$.rows[0].order_no')  AS bad_order,
       JSON_EXTRACT(rows_json,'$.rows[0].amount')  AS bad_amount, run_at
FROM dc_quality_error_data WHERE job_id=1 ORDER BY id DESC LIMIT 1;
-- 预期：sample_count=12，bad_order="ORD20261008BAD01"，bad_amount=-1.00
```

故事讲法：*3/4/5 天前订单量腰斩——上游系统故障；故障期间混入 12 笔 -1 元脏单。
质量任务 partial 告警 + 错误样本下钻，10 分钟定位问题源头。*

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

### 5.2 圈群（tag RPC 能力；网关 `/tag/group/*` 为预留空壳，用 grpcurl 直连 tag RPC :8091）

```bash
# grpcurl 安装： go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest
# tag RPC 未开 reflection，需 -import-path/-proto 指向 tag/rpc/tag.proto
# 5.2.1 建人群（圈选 SQL 必须以 select 开头且含 as object_id 别名）
grpcurl -plaintext -import-path tag/rpc -proto tag.proto -d '{"name":"高价值用户人群","object_id":1,
  "circle_rule_sql":"select id as object_id from dm_users where level >= 4","beid":2}' \
  127.0.0.1:8091 tag.Tag/CreateGroup
# 预期：返回人群 id（下文 group_id=1）

# 5.2.2 执行圈选（跑 SQL 物化成员到 dc_tag_group_member_1）
grpcurl -plaintext -import-path tag/rpc -proto tag.proto -d '{"group_id":1,"beid":2}' 127.0.0.1:8091 tag.Tag/RunCircle
# 预期：{"member_count":151,"msg":"..."} —— 成员数 151

# 5.2.3 分页看成员
grpcurl -plaintext -import-path tag/rpc -proto tag.proto -d '{"group_id":1,"beid":2,"page":1,"size":10}' \
  127.0.0.1:8091 tag.Tag/ListGroupMembers
# 预期：list 为 member 主键(1-500 中的 151 个) + tag_value(等级)
```

> **已知问题**：当前版本 tag RPC 的 Group 模型把 `dc_tag_group.last_run_at`（datetime(3)）按 int64 扫描，
> CreateGroup/RunCircle 回读人群行时会报
> `Scan error on column index 6, name "last_run_at"`（产品侧待修复）。
> 演示兜底：**HTTP 打标返回的 `tagged_rows=151` 与圈选 SQL 是同一成员口径**，DEMO_AUTO.sh 已自动降级到该口径收尾。

## 6.（第 11-13 分钟）数据服务：向导建 API → 注册应用 → 授权 → 发布 → 签名执行

```bash
# 6.1 向导建"查询订单"API（mode=wizard 强制 SELECT 模板；占位符 :name；params_json 为参数定义数组）
#     注意：LIMIT 写死 3——不要用 :limit 占位符（LIMIT ? 绑定 float64 会触发 MySQL
#     Error 1210 Incorrect arguments）；sql_text 里也不要用 DATE_FORMAT(x,'%H:%i:%s')
#     之类带冒号的格式串（:%i 会被当成未声明占位符拒绝）
curl -s -XPOST http://127.0.0.1:8887/datasvc/api/create \
  -H "Authorization: Bearer $TK" -H 'Content-Type: application/json' \
  -d '{"name":"查询订单","path":"/api/v1/demo/orders","mode":"wizard","datasource_id":1,
       "sql_text":"select order_no, amount, status, created_at from dm_orders where user_id = :user_id order by created_at desc limit 3",
       "params_json":"[{\"name\":\"user_id\",\"type\":\"number\",\"required\":true}]",
       "cache_ttl":60,"owner":"演示用户","beid":2}'
# 预期：返回 {"id":1,...}（api id）

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
#       rows_json 为 user_id=19 名下订单（至多 3 行）；再执行一次 cache_hit=true（60 秒结果缓存）
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

### 7.2 主数据（建模型 → NextCode 取号 → 录一条会员主数据）

```bash
# 注意 fields_json 类型白名单是 string/int/decimal/bool/date（不是 number）
curl -s -XPOST http://127.0.0.1:8887/mdm/model/save -H "Authorization: Bearer $TK" \
  -H 'Content-Type: application/json' \
  -d '{"model_code":"member","model_name":"会员主数据","table_name":"dm_users",
       "fields_json":"[{\"name\":\"user_name\",\"type\":\"string\",\"required\":true},{\"name\":\"level\",\"type\":\"int\",\"required\":true}]","beid":2}'
# 码表预览：NextCode 取一个业务编码（号段模式：前缀+日期+6位流水）
curl -s "http://127.0.0.1:8887/mdm/nextcode?id=1&beid=2" -H "Authorization: Bearer $TK"
# 预期：{"biz_code":"MEMBER20261012000001"}（重复调用流水递增）
# 录记录（biz_code 传 NextCode 拿到的号，或留空自动取号）
curl -s -XPOST http://127.0.0.1:8887/mdm/record/save -H "Authorization: Bearer $TK" \
  -H 'Content-Type: application/json' \
  -d '{"model_id":1,"data_json":"{\"user_name\":\"139****0001\",\"level\":5}","beid":2}'
# 预期：返回 {"id":1,"biz_code":"MEM...","status":"draft",...} —— biz_code 即码段自动生成的业务编码
curl -s "http://127.0.0.1:8887/mdm/record/list?beid=2&model_id=1&page=1&size=10" -H "Authorization: Bearer $TK"
```

### 7.3 物化视图（Phase 6 theme-rpc，网关暂未挂 HTTP 路由；grpcurl 直连 theme RPC :8080）

```bash
# theme RPC 已随 restart.sh 拉起；8080 被占时改 theme/rpc/etc/theme.yaml 的 ListenOn
# 保存物化视图定义（按 dm_orders 按天汇总）
grpcurl -plaintext -d '{"name":"订单日汇总","datasource_id":1,"db_name":"datacenter",
  "view_name":"dm_mv_order_daily",
  "select_sql":"select date(created_at) as dt, count(*) as order_cnt, sum(amount) as gmv from dm_orders group by dt",
  "refresh_mode":"manual","beid":2}' 127.0.0.1:8080 theme.Theme/SaveMaterializedView

# 刷新（真实执行 drop+create as select）
grpcurl -plaintext -d '{"id":1}' 127.0.0.1:8080 theme.Theme/RefreshMaterializedView
# 预期：{"ok":true,"msg":"...","cost_ms":<毫秒>}

# 在 MySQL 里直接看到刷新结果（含大促峰值与异常 3 天的故事曲线）
mysql -uroot -padmin datacenter -e "SELECT * FROM dm_mv_order_daily ORDER BY dt DESC LIMIT 10;"
```

---

## 附录 A：演示数据口径速查（讲给评委听的数字）

| 表 | 行数 | 口径 |
|---|---|---|
| dm_users | 500 | 省份：华东 200 / 华北 100 / 华南 75 / 西南 50 / 华中 50 / 东北 25（"哪个省份用户最多"→华东）；等级：L1 74 / L2 126 / L3 149 / L4 100 / L5 51（**高价值 level>=4 = 151 人**）；注册时间 2023-2025（30%/40%/30%）；user_name 为手机号掩码（脱敏演示点） |
| dm_products | 80 | 数码/家居/美妆/服饰四类各 20 款；每类内再分**入门/中端/高端/旗舰** 4 个价格带（全表每带 20 款，如数码 99-199/299-599/899-1599/1999-2999）；价格 .99 结尾，库存 50-999 |
| dm_orders | 1500（1488 正常 + 12 脏） | 状态：paid/shipped/completed/cancelled ≈ 25/20/45/10%；近 90 天**每天都有订单**，三段故事曲线：**平常日约 16 单/天；60 天前大促峰值日 48 单（3 倍，20-24 点晚高峰）；每周三会员日 23 单；3/4/5 天前异常骤降到 8 单/天（-50%）**；12 笔脏单 amount=-1（`where amount<0` 可查，order_no 带 BAD 段，全部落在异常 3 天）；总金额约 **277 万**（剔除 cancelled 与脏单约 249 万，随演示日小幅浮动）；order_no=ORD+日期+当日序号 |
| dm_order_items | 2500 | 1488 正常单全覆盖（前 1012 单各 2 条明细）；明细金额之和与订单 amount **严格一致**；含 9 折促销价快照 |
| dm_user_feedback | 200 | 近 30 天；满意度 5 星 40%/4 星 30%/3 星 15%/2 星 10%/1 星 5%（**好评率(4 星+)约 70%**）；content 为词典组合短句；**末尾 3 行是脏行（content=''、score=0）**，供 not_null/值域演示命中 |
| base_member | +1 | C 端演示账号 `13800138000 / 123456`（username=demo，需 user RPC） |
| base_app | +1 | id=2「演示电商平台」，即全篇 beid=2 |
| sys_user | sql.sql+sys.sql 初始化 | 管理台账号 `admin / admin123`（bcrypt），控制台与 /sys/login 用 |

数据全部由递归 CTE + 确定性算术伪随机生成（无 RAND()），重复 source 结果一致；
脚本可重复执行（先 DROP 演示表、清插演示账号）。

**三个"可被系统发现"的故事钩子**（DEMO_AUTO.sh 第 5/6/7 步自动呈现）：
1. 大促峰值：`60 天前 = 48 单 = 平日 3 倍` → 讲"报表/问数一眼看到大促"；
2. 异常骤降：`3/4/5 天前 = 8 单/天（-50%）` → 讲"趋势异常被发现"；
3. 脏数据：`12 笔 amount=-1 混在异常 3 天` → 讲"质量规则 value_range 命中 + 错误样本留存，证据链闭环"。

## 附录 B：常见问题

- **一键脚本某步 warning**：看黄色提示；常见是元数据未采集（先跑步骤 4）、user RPC 未启动（C 端登录降级，不影响主流程）。
- **登录 401/未注册**：C 端是 base_member（demo_data.sql 插入）、管理台是 sys_user（sys.sql 初始化）——确认对应文件已 source 且 source 进了 datacenter 库。
- **采集/问数/质量查不到 dm_ 表**：`beid` 不一致——演示统一用 beid=2，各接口记得带上（`/datasource/test`、`/quality/job/run` 也需要）。
- **AI 问数 candidates 为空**：先做元数据采集（候选来自元数据实体+指标）；LLM 未配置时 `llm_enabled=false` 属正常降级。
- **/datasvc/execute 报签名不匹配**：ts 必须是当前 unix 秒（偏差 ≤300s）、sign 按 `app_key+ts` 用 app_secret 计算。
- **/datasvc/execute 报 Error 1210**：SQL 模板里用了 `:limit` 之类占位符在 LIMIT 位置（绑定 float64 触发），把 LIMIT 写死即可。
- **人群圈选报 last_run_at 扫描错误**：tag RPC 已知兼容问题（datetime(3)→int64），见 5.2 说明，用打标口径收尾。
- **8887 起不来**：`restart.sh` 后看网关日志；常见是 /etc/datacenter-api.yaml 缺失或 mysql/etcd 未就绪。
