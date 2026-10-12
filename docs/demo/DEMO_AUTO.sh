#!/usr/bin/env bash
# ============================================================================
# 数据中台一键演示脚本 DEMO_AUTO.sh
# ----------------------------------------------------------------------------
# 用途    : 10 分钟内对着一套已启动的数据中台环境跑通"有故事感"的完整演示——
#           登录 → 登记数据源 → 采集元数据 → AI 问数三连 → 质量规则与脏数据捕获
#           → 数据服务签名调用 → 标签打标 → 人群圈选 → 主数据编码 → 收尾。
#           demo_data.sql 里埋的大促峰值/异常骤降/脏数据故事由质量步骤呈现。
# 用法    : bash docs/demo/DEMO_AUTO.sh
# 前置    : 网关已启动（默认 http://127.0.0.1:8887）、演示库可连
#           （默认 root:admin@127.0.0.1:3306，库 datacenter），
#           且已 source demo_data.sql（脏数据/反馈表等故事数据在其中）。
# 环境变量: GW / BEID / ADMIN_USER / ADMIN_PASS
#           MYSQL_HOST / MYSQL_PORT / MYSQL_USER / MYSQL_PWD / MYSQL_DB
#           MYSQL_CONTAINER（宿主机无 mysql 命令时用 docker exec 该容器代替，默认 dc-run-mysql）
# 幂等    : 数据源/标签对象/标签/人群/主数据模型按名称查找复用，重复执行安全；
#           质量规则/任务、API、应用与主数据记录每次新建（可重复演示）。
# 每步失败不中断（黄色 warning 继续）；仅前置检查/登录/数据源失败红色退出。
# ============================================================================
set -euo pipefail

# ---------- 可配置项（环境变量覆盖） ----------------------------------------
GW="${GW:-http://127.0.0.1:8887}"
BEID="${BEID:-2}"
ADMIN_USER="${ADMIN_USER:-admin}"
ADMIN_PASS="${ADMIN_PASS:-admin123}"
MYSQL_HOST="${MYSQL_HOST:-127.0.0.1}"
MYSQL_PORT="${MYSQL_PORT:-3306}"
MYSQL_USER="${MYSQL_USER:-root}"
MYSQL_PWD="${MYSQL_PWD:-admin}"
MYSQL_DB="${MYSQL_DB:-datacenter}"
MYSQL_CONTAINER="${MYSQL_CONTAINER:-dc-run-mysql}"
FRONT_URL="${FRONT_URL:-http://localhost:5173}"
DS_NAME="演示电商平台业务库"
API_PATH="/api/v1/demo/orders"
TAG_OBJ_NAME="电商用户"
TAG_KEY="high_value"
GROUP_NAME="高价值用户人群"
MDM_CODE="member"

# ---------- 彩色输出 ---------------------------------------------------------
if [ -t 1 ]; then
  C_G=$'\e[32m' C_Y=$'\e[33m' C_R=$'\e[31m' C_C=$'\e[36m' C_B=$'\e[1m' C_D=$'\e[2m' C_RST=$'\e[0m'
else
  C_G='' C_Y='' C_R='' C_C='' C_B='' C_D='' C_RST=''
fi
TOTAL_STEP=11
STEP_NO=0

say()   { printf '%s\n' "$*"; }
phead() { STEP_NO=$((STEP_NO+1)); printf '\n%s═══ [%s/%s] %s %s═══%s\n' "$C_C$C_B" "$STEP_NO" "$TOTAL_STEP" "$*" "$C_C" "$C_RST"; }
ok()    { printf '%s  ✔ %s%s\n' "$C_G" "$*" "$C_RST"; }
warn()  { printf '%s  ⚠ %s%s\n' "$C_Y" "$*" "$C_RST"; }
bad()   { printf '%s  ✘ %s%s\n' "$C_R" "$*" "$C_RST"; }
note()  { printf '%s  · %s%s\n' "$C_D" "$*" "$C_RST"; }
clip()  { printf '%s' "$1" | tr -d '\n' | cut -c1-"$2"; }   # clip <串> <长度>

# ---------- 依赖探测 ---------------------------------------------------------
HAVE_PY=0;    command -v python3 >/dev/null 2>&1 && HAVE_PY=1
HAVE_SSL=0;   command -v openssl >/dev/null 2>&1 && HAVE_SSL=1
HAVE_MYSQL=0; command -v mysql   >/dev/null 2>&1 && HAVE_MYSQL=1

# mq [mysql 参数...]：宿主机有 mysql 客户端直连；否则 docker exec 到容器内执行
mq() {
  if [ "$HAVE_MYSQL" = 1 ]; then
    MYSQL_PWD="$MYSQL_PWD" mysql -h "$MYSQL_HOST" -P "$MYSQL_PORT" -u "$MYSQL_USER" --default-character-set=utf8mb4 "$@"
  else
    docker exec -i -e MYSQL_PWD="$MYSQL_PWD" "$MYSQL_CONTAINER" \
      mysql -u "$MYSQL_USER" --default-character-set=utf8mb4 "$@"
  fi
}

# jget <json串> <python表达式(变量d)> ：解析并打印；解析失败输出空串
jget() {
  if [ "$HAVE_PY" != 1 ]; then printf '%s' "$1" | cut -c1-200; return 0; fi
  printf '%s' "$1" | python3 -c '
import json, sys
try:
    d = json.load(sys.stdin)
    v = eval(sys.argv[1], {"d": d})
    if v is None:
        sys.stdout.write("")
    elif isinstance(v, (list, dict)):
        sys.stdout.write(json.dumps(v, ensure_ascii=False))
    else:
        sys.stdout.write(str(v))
except Exception:
    pass' "$2"
}

# 管理面接口封装（自动带 JWT）
api_post() { curl -s -m 20 -XPOST "$GW$1" -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' -d "$2"; }
api_get()  { curl -s -m 20 "$GW$1" -H "Authorization: Bearer $TOKEN"; }

# run_step <标题> <函数...>：软步骤，失败 warning 继续不退出
run_step() {
  local title="$1"; shift
  phead "$title"
  if "$@"; then return 0; fi
  warn "「$title」未完整走通（跳过，继续下一步）"
  return 0
}
# must_step <标题> <函数...>：关键步骤，失败红色提示并退出 1
must_step() {
  local title="$1"; shift
  phead "$title"
  if "$@"; then return 0; fi
  bad "「$title」失败——关键步骤，演示无法继续（排查后重跑：bash $0）"
  exit 1
}

# ---------- [1] 前置检查 ------------------------------------------------------
precheck() {
  say "环境探测：python3=$([ $HAVE_PY = 1 ] && echo 有 || echo 无)  openssl=$([ $HAVE_SSL = 1 ] && echo 有 || echo 无)  mysql客户端=$([ $HAVE_MYSQL = 1 ] && echo 有 || echo "无→docker exec $MYSQL_CONTAINER")"
  local pass=1 http_code
  http_code=$(curl -s -m 5 -o /dev/null -w '%{http_code}' -XPOST "$GW/sys/login" -H 'Content-Type: application/json' -d '{}' 2>/dev/null) || http_code=000
  if [ "$http_code" = "000" ]; then
    bad "网关 $GW 不可达（请先 sh restart.sh 启动；或在别的机器用 GW=... 指定地址）"
    pass=0
  else
    ok "网关可达 $GW（HTTP $http_code）"
  fi
  if mq -N -B -e "SELECT 1" >/dev/null 2>&1; then
    ok "MySQL 可连 $MYSQL_HOST:$MYSQL_PORT/$MYSQL_DB"
  else
    bad "MySQL 不可连（$MYSQL_USER@${MYSQL_HOST}:${MYSQL_PORT}，容器兜底 $MYSQL_CONTAINER 也失败）——核对 server.sh 依赖或 MYSQL_* 环境变量"
    pass=0
  fi
  if [ "$HAVE_SSL" = 1 ] || [ "$HAVE_PY" = 1 ]; then
    ok "签名计算可用（$([ "$HAVE_SSL" = 1 ] && echo openssl || echo python3-hmac)）"
  else
    warn "openssl 与 python3 均不可用——数据服务签名调用步骤将跳过"
  fi
  if [ "$pass" != "1" ]; then bad "前置检查未通过，演示中止"; return 1; fi
  return 0
}

# ---------- [2] 登录（管理台 admin/admin123，双账号体系说明） ------------------
do_login() {
  local resp token cend
  resp=$(curl -s -m 10 -XPOST "$GW/sys/login" -H 'Content-Type: application/json' \
    -d "{\"username\":\"$ADMIN_USER\",\"password\":\"$ADMIN_PASS\"}")
  token=$(jget "$resp" "d.get('token','')")
  if [ -z "$token" ]; then
    bad "管理台登录失败：$(clip "$resp" 160)（确认 sys.sql 已导入且 sys_user 有 $ADMIN_USER）"
    return 1
  fi
  TOKEN="$token"
  ok "管理台登录成功（sys_user：$ADMIN_USER，角色 $(jget "$resp" "(d.get('roles') or [''])[0]" )，token ${#token} 字符）"
  note "双账号体系：管理面走 sys_user（$ADMIN_USER/$ADMIN_PASS）；C 端会员走 base_member（13800138000/123456，需 user RPC 在线）"
  cend=$(curl -s -m 5 -XPOST "$GW/user/login" -H 'Content-Type: application/json' \
    -d '{"mobile":"13800138000","type":1,"password":"123456"}') || true
  if [ -n "$(jget "$cend" "d.get('access_token','')")" ]; then
    note "C 端会员登录可用（13800138000 @ 演示电商平台 beid=$BEID）"
  else
    note "C 端会员登录未验证（user RPC 未启动即属正常降级，不影响管理面演示）"
  fi
  return 0
}

# ---------- [3] 登记演示数据源 ------------------------------------------------
do_datasource() {
  local resp ds_id test_resp
  resp=$(api_get "/datasource/list?beid=$BEID&page=1&size=100")
  ds_id=$(jget "$resp" "next((x['id'] for x in d.get('list',[]) if x.get('name')=='$DS_NAME'), '')")
  if [ -n "$ds_id" ]; then
    ok "复用已有数据源 id=$ds_id（$DS_NAME）"
  else
    resp=$(api_post "/datasource/create" "{\"name\":\"$DS_NAME\",\"type\":\"mysql\",\"host\":\"$MYSQL_HOST\",\"port\":$MYSQL_PORT,\"db_name\":\"$MYSQL_DB\",\"username\":\"$MYSQL_USER\",\"password\":\"$MYSQL_PWD\",\"owner\":\"演示用户\",\"beid\":$BEID}")
    ds_id=$(jget "$resp" "d.get('id','')")
    if [ -z "$ds_id" ]; then
      bad "数据源创建失败：$(clip "$resp" 160)"
      return 1
    fi
    ok "新建数据源 id=$ds_id（$DS_NAME → $MYSQL_HOST:$MYSQL_PORT/$MYSQL_DB）"
  fi
  DS_ID="$ds_id"
  test_resp=$(api_get "/datasource/test?id=$DS_ID&beid=$BEID")
  note "连接探活：$(clip "$test_resp" 120)"
  return 0
}

# ---------- [4] 采集元数据 ----------------------------------------------------
do_collect() {
  local resp t c dm_n ent_total
  resp=$(api_post "/metadata/collect" "{\"ds_id\":$DS_ID,\"beid\":$BEID,\"db_name\":\"$MYSQL_DB\"}")
  t=$(jget "$resp" "d.get('tables_created','')")
  c=$(jget "$resp" "d.get('columns_created','')")
  if [ -z "$t" ]; then
    bad "采集失败：$(clip "$resp" 160)"
    return 1
  fi
  ok "整库采集完成：tables_created=$t columns_created=$c（首次全量入库；重复采集按 fqn 幂等，无新增时为 0）"
  ent_total=$(jget "$(api_get "/metadata/entity/list?beid=$BEID&page=1&size=1")" "d.get('total', 0)")
  dm_n=$(jget "$(api_get "/metadata/entity/list?beid=$BEID&entity_type=table&keyword=dm&page=1&size=100")" "sum(1 for x in d.get('list',[]) if x.get('name','').startswith('dm_'))")
  note "元数据实体 $ent_total 个（库/表/列），其中 dm_ 演示表 ${dm_n:-5} 张（users/products/orders/order_items/user_feedback），COMMENT 已埋演示关键词"
  return 0
}

# ---------- [5] AI 问数三连 ---------------------------------------------------
ai_ask() {
  local q="$1" expect="$2" resp cand ans ms
  resp=$(api_post "/ai/ask" "{\"question\":\"$q\",\"beid\":$BEID}")
  cand=$(jget "$resp" "'; '.join((x.get('type','')+':'+x.get('name','')) for x in d.get('candidates',[])[:3])")
  ans=$(jget "$resp" "(d.get('answer','') or '')[:70].replace(chr(10),' ')")
  ms=$(jget "$resp" "d.get('cost_ms',0)")
  say "    问：$q"
  say "    命中：${cand:-（无候选）}  ${C_D}cost=${ms}ms$C_RST"
  say "    答：$ans"
  if printf '%s' "$cand" | grep -q "$expect"; then
    ok "命中预期表 $expect"
  else
    warn "未命中预期表 $expect（候选：${cand:-空}——确认上一步元数据已采集）"
  fi
  return 0
}
do_ai() {
  ai_ask "最近90天每天订单量" "dm_orders"
  ai_ask "总销售额是多少" "dm_orders"
  ai_ask "用户反馈怎么样" "dm_user_feedback"
  note "故事：问“最近90天每天订单量”可见 60 天前大促峰值（3 倍）与近 3 天骤降（-50%）"
  return 0
}

# ---------- [6] 数据质量：2 条规则 + 任务 + 执行 ------------------------------
do_quality() {
  local r1 r2 j run total pass rate rl
  r1=$(api_post "/quality/rule/create" "{\"name\":\"订单号非空\",\"rule_type\":\"not_null\",\"target_table\":\"dm_orders\",\"target_column\":\"order_no\",\"owner\":\"演示用户\",\"beid\":$BEID}")
  R1=$(jget "$r1" "d.get('id','')")
  if [ -z "$R1" ]; then bad "规则1创建失败：$(clip "$r1" 160)"; return 1; fi
  r2=$(api_post "/quality/rule/create" "{\"name\":\"订单金额值域0-100万\",\"rule_type\":\"value_range\",\"target_table\":\"dm_orders\",\"target_column\":\"amount\",\"params_json\":\"{\\\"min\\\":0,\\\"max\\\":1000000}\",\"owner\":\"演示用户\",\"beid\":$BEID}")
  R2=$(jget "$r2" "d.get('id','')")
  if [ -z "$R2" ]; then bad "规则2创建失败：$(clip "$r2" 160)"; return 1; fi
  ok "规则创建：#$R1 订单号非空(not_null) / #$R2 金额值域(value_range 0~100万)"

  j=$(api_post "/quality/job/create" "{\"name\":\"电商演示质量任务\",\"datasource_id\":$DS_ID,\"db_name\":\"$MYSQL_DB\",\"rule_ids\":\"$R1,$R2\",\"beid\":$BEID}")
  JOB_ID=$(jget "$j" "d.get('id','')")
  if [ -z "$JOB_ID" ]; then bad "任务创建失败：$(clip "$j" 160)"; return 1; fi
  ok "质量任务 #$JOB_ID 已绑定规则 $R1,$R2，执行中…"

  run=$(api_post "/quality/job/run" "{\"id\":$JOB_ID,\"beid\":$BEID}")
  total=$(jget "$run" "d.get('total_rules',0)")
  pass=$(jget "$run" "d.get('passed',0)")
  if [ -z "$(jget "$run" "d.get('total_rules','')")" ]; then
    bad "任务执行失败：$(clip "$run" 160)"
    return 1
  fi
  if [ "${total:-0}" -gt 0 ] 2>/dev/null; then
    rate="$pass/$total（$(( pass * 100 / total ))%）"
  else
    rate="0/0"
  fi
  say "    通过率：$rate  last_status=$(jget "$run" "d.get('last_status','')")"
  rl=$(api_get "/quality/result/list?job_id=$JOB_ID&page=1&size=10")
  if [ "$HAVE_PY" = 1 ]; then
    printf '%s' "$rl" | python3 -c '
import json, sys
try:
    d = json.load(sys.stdin)
    for x in d.get("list", [])[:2]:
        print("    规则#%s %s：pass_flag=%s 实测=%s" % (x.get("rule_id"), x.get("rule_type"), x.get("pass_flag"), x.get("check_value")))
except Exception:
    pass'
  fi
  note "口径：not_null 通过（订单号完备）；value_range 命中脏数据 → 任务 partial——下一步看错误数据捕获"
  return 0
}

# ---------- [7] 脏数据错误捕获 ------------------------------------------------
do_errordata() {
  local cnt line
  cnt=$(mq -N -B -D "$MYSQL_DB" -e "SELECT COUNT(*) FROM dm_orders WHERE amount < 0" 2>/dev/null) || cnt=""
  if [ -z "$cnt" ]; then
    warn "无法直连演示库统计脏数据（mysql 不可用）"
    return 1
  fi
  ok "源表脏数据：dm_orders 中 amount<0 共 $cnt 行（order_no 带 BAD 段，模拟上游故障混入的 -1 元订单）"
  if [ "$HAVE_PY" = 1 ]; then
    line=$(mq -N -B -D "$MYSQL_DB" -e "SELECT sample_count, rows_json FROM dc_quality_error_data WHERE job_id=$JOB_ID ORDER BY id DESC LIMIT 1" 2>/dev/null | head -1) || line=""
    if [ -n "$line" ]; then
      printf '%s' "$line" | python3 -c '
import json, sys
raw = sys.stdin.read().strip()
cnt, _, payload = raw.partition("\t")
try:
    rows = json.loads(payload).get("rows", [])
except Exception:
    rows = []
if rows:
    print("    系统捕获：错误样本 %s 行（上限100），如 %s amount=%s" % (cnt, rows[0].get("order_no"), rows[0].get("amount")))
    for r in rows[1:3]:
        print("             … %s amount=%s" % (r.get("order_no"), r.get("amount")))
else:
    print("    （样本行解析为空——确认质量任务已执行且 value_range 失败）")'
    else
      note "（未查到错误样本行——确认质量任务已执行且 value_range 规则失败）"
    fi
  fi
  note "故事：上游故障那 3 天（3/4/5 天前）单量骤降 50%，同时混入 -1 元脏单——规则发现 + 样本留存，可下钻定位"
  return 0
}

# ---------- [8] 数据服务：建 API → 应用 → 授权 → 发布 → 签名调用 --------------
do_datasvc() {
  local resp api_id app_id app_key app_secret ts sign rows n first
  # SQL 模板：LIMIT 写死 3（:limit 占位符经 prepared statement 绑定 float64 会触发
  # MySQL Error 1210，演示 API 用固定行数绕开；参数仅 user_id）
  local sql_text='select order_no, amount, status, created_at from dm_orders where user_id = :user_id order by created_at desc limit 3'
  local params_json='[{"name":"user_id","type":"number","required":true}]'
  local params_json_esc=${params_json//\"/\\\"}   # 内嵌 JSON 需转义双引号
  resp=$(api_get "/datasvc/api/info?path=$API_PATH&beid=$BEID")
  api_id=$(jget "$resp" "d.get('id','')")
  if [ -n "$api_id" ]; then
    ok "复用已有 API #$api_id（$API_PATH），同步刷新 SQL 模板"
    api_post "/datasvc/api/update" "{\"id\":$api_id,\"beid\":$BEID,\"sql_text\":\"$sql_text\",\"params_json\":\"$params_json_esc\",\"cache_ttl\":60}" >/dev/null
  else
    resp=$(api_post "/datasvc/api/create" "{\"name\":\"查询订单\",\"path\":\"$API_PATH\",\"mode\":\"wizard\",\"datasource_id\":$DS_ID,\"sql_text\":\"$sql_text\",\"params_json\":\"$params_json_esc\",\"cache_ttl\":60,\"owner\":\"演示用户\",\"beid\":$BEID}")
    api_id=$(jget "$resp" "d.get('id','')")
    if [ -z "$api_id" ]; then bad "API 创建失败：$(clip "$resp" 160)"; return 1; fi
    ok "向导创建 API #$api_id（$API_PATH，按 user_id 查订单，缓存 60s）"
  fi
  API_ID="$api_id"

  resp=$(api_post "/datasvc/app/register" "{\"name\":\"demo-app-$(date +%s)\",\"rate_limit_qps\":10,\"beid\":$BEID}")
  app_id=$(jget "$resp" "d.get('app_id','')")
  app_key=$(jget "$resp" "d.get('app_key','')")
  app_secret=$(jget "$resp" "d.get('app_secret','')")
  if [ -z "$app_key" ]; then bad "应用注册失败：$(clip "$resp" 160)"; return 1; fi
  ok "注册应用 #$app_id app_key=$app_key（secret 已下发，仅此一次）"

  api_post "/datasvc/grant" "{\"api_id\":$API_ID,\"app_id\":$app_id,\"beid\":$BEID}" >/dev/null
  api_post "/datasvc/api/publish" "{\"id\":$API_ID}" >/dev/null
  ok "已授权 API#$API_ID → 应用#$app_id，并发布（status=published）"

  ts=$(date +%s)
  if [ "$HAVE_SSL" = 1 ]; then
    sign=$(printf '%s' "${app_key}${ts}" | openssl dgst -sha256 -hmac "$app_secret" | awk '{print $NF}')
  elif [ "$HAVE_PY" = 1 ]; then
    sign=$(python3 -c 'import hmac,hashlib,sys; print(hmac.new(sys.argv[1].encode(), sys.argv[2].encode(), hashlib.sha256).hexdigest())' "$app_secret" "${app_key}${ts}")
  else
    warn "无 openssl/python3，跳过签名调用"
    return 0
  fi
  resp=$(curl -s -m 20 -XPOST "$GW/datasvc/execute" -H 'Content-Type: application/json' \
    -d "{\"app_key\":\"$app_key\",\"ts\":\"$ts\",\"sign\":\"$sign\",\"path\":\"$API_PATH\",\"params_json\":\"{\\\"user_id\\\":19}\"}")
  rows=$(jget "$resp" "d.get('rows_json','')")
  if [ -z "$rows" ]; then
    bad "签名调用失败：$(clip "$resp" 160)（核对 sign=hex(hmac-sha256(key=app_secret, msg=app_key+ts))、ts 偏差≤300s）"
    return 1
  fi
  n="?"; first=""
  if [ "$HAVE_PY" = 1 ]; then
    n=$(printf '%s' "$rows" | python3 -c 'import json,sys; print(len(json.load(sys.stdin)))' 2>/dev/null) || n="?"
    first=$(printf '%s' "$rows" | python3 -c 'import json,sys; r=json.load(sys.stdin)[0]; print({k: r[k] for k in list(r)[:4]})' 2>/dev/null) || first=""
  fi
  ok "签名调用成功（免 JWT）：cost=$(jget "$resp" "d.get('cost_ms',0)")ms 返回 $n 行真实订单数据"
  if [ -n "$first" ]; then printf '    首行：%s\n' "$first"; fi
  return 0
}

# ---------- [9] 标签打标（高价值用户） ----------------------------------------
do_tag() {
  local resp obj_id tag_id tagged probe
  resp=$(api_get "/tag/object/list?beid=$BEID&object_type=user&page=1&size=100")
  obj_id=$(jget "$resp" "next((x['id'] for x in d.get('list',[]) if x.get('name')=='$TAG_OBJ_NAME'), '')")
  if [ -n "$obj_id" ]; then
    ok "复用标签对象 #$obj_id（$TAG_OBJ_NAME → dm_users.id）"
  else
    resp=$(api_post "/tag/object/create" "{\"object_type\":\"user\",\"name\":\"$TAG_OBJ_NAME\",\"ds_id\":$DS_ID,\"db_name\":\"$MYSQL_DB\",\"table_name\":\"dm_users\",\"pk_column\":\"id\",\"beid\":$BEID}")
    obj_id=$(jget "$resp" "d.get('id','')")
    if [ -z "$obj_id" ]; then bad "标签对象创建失败：$(clip "$resp" 160)"; return 1; fi
    ok "创建标签对象 #$obj_id（$TAG_OBJ_NAME → dm_users.id）"
  fi
  TAG_OBJ_ID="$obj_id"

  resp=$(api_get "/tag/def/list?beid=$BEID&object_id=$obj_id&page=1&size=100")
  tag_id=$(jget "$resp" "next((x['id'] for x in d.get('list',[]) if x.get('tag_key')=='$TAG_KEY'), '')")
  if [ -n "$tag_id" ]; then
    ok "复用标签定义 #$tag_id（$TAG_KEY 高价值用户：level>=4）"
  else
    resp=$(api_post "/tag/def/create" "{\"object_id\":$obj_id,\"tag_key\":\"$TAG_KEY\",\"tag_name\":\"高价值用户\",\"tag_type\":\"rule\",\"rule_sql\":\"select id as object_id, level as tag_value from dm_users where level >= 4\",\"beid\":$BEID}")
    tag_id=$(jget "$resp" "d.get('id','')")
    if [ -z "$tag_id" ]; then bad "标签定义创建失败：$(clip "$resp" 160)"; return 1; fi
    ok "创建规则标签 #$tag_id（高价值用户 = level>=4）"
  fi
  TAG_ID="$tag_id"

  resp=$(api_post "/tag/tagging" "{\"tag_id\":$tag_id,\"object_id\":$obj_id,\"beid\":$BEID}")
  tagged=$(jget "$resp" "d.get('tagged_rows','')")
  if [ -z "$tagged" ]; then bad "打标失败：$(clip "$resp" 160)"; return 1; fi
  ok "执行打标：tagged_rows=$tagged（L4+L5 高价值用户写入标签值表）"
  probe=$(api_get "/tag/object/tags?object_id=$obj_id&object_pk=2&beid=$BEID")
  note "画像抽查：用户 id=2 → $(clip "$probe" 150)"
  return 0
}

# ---------- [10] 人群圈选（tag RPC gRPC；网关 /tag/group/* 为预留空壳） --------
do_group() {
  local grpc_bin proto_dir proto_out gid run_out circle_sql
  circle_sql="select id as object_id from dm_users where level >= 4"
  grpc_bin="${GRPCURL:-}"
  [ -z "$grpc_bin" ] && grpc_bin="$(command -v grpcurl || true)"
  [ -z "$grpc_bin" ] && [ -x "$HOME/go/bin/grpcurl" ] && grpc_bin="$HOME/go/bin/grpcurl"
  proto_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" 2>/dev/null && pwd)/../../tag/rpc"
  if [ -x "$grpc_bin" ] && [ -f "$proto_dir/tag.proto" ]; then
    # 建人群（幂等：先查库，无则创建；CreateGroup 回读 last_run_at 可能报扫描错，但插入已落库）
    gid=$(mq -N -B -D "$MYSQL_DB" -e "SELECT id FROM dc_tag_group WHERE name='$GROUP_NAME' AND beid=$BEID ORDER BY id DESC LIMIT 1" 2>/dev/null) || gid=""
    if [ -n "$gid" ]; then
      ok "复用人群 #$gid（$GROUP_NAME，object_id=$TAG_OBJ_ID）"
    else
      proto_out=$("$grpc_bin" -plaintext -import-path "$proto_dir" -proto tag.proto \
        -d "{\"name\":\"$GROUP_NAME\",\"object_id\":$TAG_OBJ_ID,\"circle_rule_sql\":\"$circle_sql\",\"beid\":$BEID}" \
        127.0.0.1:8091 tag.Tag/CreateGroup 2>&1) || true
      gid=$(mq -N -B -D "$MYSQL_DB" -e "SELECT id FROM dc_tag_group WHERE name='$GROUP_NAME' AND beid=$BEID ORDER BY id DESC LIMIT 1" 2>/dev/null) || gid=""
      if [ -z "$gid" ]; then
        warn "tag RPC CreateGroup 未落库：$(clip "$proto_out" 140)"
        return 1
      fi
      ok "tag RPC 创建人群 #$gid（$GROUP_NAME，圈选 SQL：level>=4）"
    fi
    # 执行圈选
    run_out=$("$grpc_bin" -plaintext -import-path "$proto_dir" -proto tag.proto \
      -d "{\"group_id\":$gid,\"beid\":$BEID}" 127.0.0.1:8091 tag.Tag/RunCircle 2>&1) || true
    if printf '%s' "$run_out" | grep -q '"member_count"'; then
      ok "执行圈选：$(clip "$run_out" 120)（成员已物化到人群表）"
      local mem
      mem=$("$grpc_bin" -plaintext -import-path "$proto_dir" -proto tag.proto \
        -d "{\"group_id\":$gid,\"beid\":$BEID,\"page\":1,\"size\":3}" 127.0.0.1:8091 tag.Tag/ListGroupMembers 2>&1) || true
      note "成员抽样：$(clip "$mem" 160)"
      return 0
    fi
    if printf '%s' "$run_out" | grep -q "last_run_at"; then
      warn "圈选回读失败（已知兼容问题：dc_tag_group.last_run_at 为 datetime(3)，RPC 模型按 int64 扫描报错，需产品侧修复）"
    else
      warn "tag RPC RunCircle 失败：$(clip "$run_out" 140)"
    fi
  else
    note "未找到 grpcurl（go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest）或 tag/rpc/tag.proto，跳过 gRPC 直连"
  fi
  # 降级口径：打标结果即同一圈选 SQL 的成员数
  local tagged
  tagged=$(jget "$(api_post "/tag/tagging" "{\"tag_id\":$TAG_ID,\"object_id\":$TAG_OBJ_ID,\"beid\":$BEID}")" "d.get('tagged_rows','')")
  note "降级口径：HTTP 打标 tagged_rows=${tagged:-?} —— 与人群圈选 SQL（level>=4）同一成员集合，可直接用于收尾讲解"
  return 0
}

# ---------- [11] 主数据：模型 + NextCode + 录记录 ------------------------------
do_mdm() {
  local resp model_id code saved
  resp=$(api_get "/mdm/model/list?beid=$BEID&keyword=$MDM_CODE&page=1&size=50")
  model_id=$(jget "$resp" "next((x['id'] for x in d.get('list',[]) if x.get('model_code')=='$MDM_CODE'), '')")
  if [ -n "$model_id" ]; then
    ok "复用主数据模型 #$model_id（$MDM_CODE 会员主数据）"
  else
    resp=$(api_post "/mdm/model/save" "{\"model_code\":\"$MDM_CODE\",\"model_name\":\"会员主数据\",\"table_name\":\"dm_users\",\"fields_json\":\"[{\\\"name\\\":\\\"user_name\\\",\\\"type\\\":\\\"string\\\",\\\"required\\\":true},{\\\"name\\\":\\\"level\\\",\\\"type\\\":\\\"int\\\",\\\"required\\\":true}]\",\"beid\":$BEID}")
    model_id=$(jget "$resp" "d.get('id','')")
    if [ -z "$model_id" ]; then bad "模型创建失败：$(clip "$resp" 160)"; return 1; fi
    ok "创建主数据模型 #$model_id（$MDM_CODE：user_name/level，码段编码策略）"
  fi
  resp=$(api_get "/mdm/nextcode?id=$model_id&beid=$BEID")
  code=$(jget "$resp" "d.get('biz_code','')")
  if [ -z "$code" ]; then bad "NextCode 失败：$(clip "$resp" 160)"; return 1; fi
  ok "号段取号 NextCode → $code"
  resp=$(api_post "/mdm/record/save" "{\"model_id\":$model_id,\"biz_code\":\"$code\",\"data_json\":\"{\\\"user_name\\\":\\\"139****8888\\\",\\\"level\\\":5}\",\"beid\":$BEID}")
  saved=$(jget "$resp" "d.get('id','')")
  if [ -z "$saved" ]; then bad "记录保存失败：$(clip "$resp" 160)"; return 1; fi
  ok "录入主数据记录 #$saved（biz_code=$code 即 NextCode 取号，status=draft）"
  return 0
}

# ---------- 主流程 -----------------------------------------------------------
say "${C_B}════════════════════════════════════════════════════════════"
say "  数据中台一键演示  ·  网关 $GW  ·  租户 beid=$BEID  ·  $(date '+%F %T')"
say "════════════════════════════════════════════${C_RST}"

phead "前置检查"
if ! precheck; then
  bad "前置检查未通过，演示中止"
  exit 1
fi
must_step "管理台登录"       do_login
must_step "登记演示数据源"   do_datasource
run_step  "采集元数据"       do_collect
run_step  "AI 问数三连"      do_ai
run_step  "质量规则与执行"   do_quality
run_step  "脏数据错误捕获"   do_errordata
run_step  "数据服务签名调用" do_datasvc
run_step  "标签打标"         do_tag
run_step  "人群圈选"         do_group
run_step  "主数据编码与录入" do_mdm

say ""
say "${C_B}══════════════ 演示完成 ══════════════${C_RST}"
say "  ${C_G}✔ 演示完成，打开 ${C_B}$FRONT_URL${C_G} 体验控制台（管理台 $ADMIN_USER/$ADMIN_PASS）${C_RST}"
say "  数据故事：60 天前大促峰值 3 倍 · 近 3 天订单骤降 50% · 12 笔 -1 元脏单被质量规则捕获"
say "  （脚本可重复执行：数据源/标签/人群/模型按名复用；质量/API/应用/记录按次新建）"
say ""
