# signproxy — DataEase 动态签名代理

解决 DataEase v1.18/v2 API 数据源只能配**静态** header，而平台 `/datasvc/execute`
要求 **HMAC 动态签名**（`ts` ±300 秒强校验，静态配置 5 分钟后必然过期）的矛盾。
即 `docs/DEPLOY-DATAEASE.md` 三种落地方式中的**方案 c**：平台零改码，代理转发时现算签名。

```
DataEase API 数据源（静态 body，header 留空）
  → signproxy（每次请求现算 ts/sign 注入 body + header）
    → 平台网关 POST /datasvc/execute → 原样返回
```

签名算法与 `datasvc/rpc/internal/logic/credential.go` 的 `genSign` 逐字一致：
`sign = hex(hmac-sha256(key=app_secret, msg=app_key+ts))`，`ts` 为秒级 unix 时间戳。
网关 `ExecuteApiReq` 从 JSON body 解析 `app_key/ts/sign`，故代理将签名注入 body；
同时按对外 OpenAPI 文档契约回写 `X-App-Key` / `X-Timestamp` / `X-Sign` 三个 header，两种姿势均兼容。

## 编译与运行

```bash
# 仓库根目录
go build -o signproxy ./cmd/signproxy

export APP_KEY="<datasvc 注册应用返回的 app_key>"
export APP_SECRET="<对应的 app_secret>"
# 可选：PROXY_LISTEN（默认 :8077，8090/8091 已被 metric 等占用）、UPSTREAM（默认 http://127.0.0.1:8887）
./signproxy
```

环境变量一览：

| 变量 | 必填 | 默认值 | 说明 |
|------|------|--------|------|
| `APP_KEY` | 是 | — | datasvc 应用 app_key，缺失启动即退出 |
| `APP_SECRET` | 是 | — | 对应 app_secret，缺失启动即退出 |
| `PROXY_LISTEN` | 否 | `:8077` | 代理监听地址 |
| `UPSTREAM` | 否 | `http://127.0.0.1:8887` | 平台网关地址（datacenter api） |

## DataEase 侧配置（API 数据源）

数据源类型选 API / 请求方法 POST，**base url 指向本代理**，header 留空即可（签名与
Content-Type 均由代理注入）：

| 配置项 | 值 |
|--------|----|
| 请求方法 / URL | `POST` / `http://<proxy主机>:8077/datasvc/execute` |
| Headers | 留空（代理注入 `Content-Type: application/json` 与签名三件套） |
| 请求体 | `{"path":"/api/v1/demo/orders","params_json":"{\"user_id\":19,\"limit\":10}"}` |
| 数据格式 / 取数 | JSON，从 `rows_json` 解析数组（JSONPath：`$.rows_json`） |

注意：请求体里**不要再写** `app_key/ts/sign`（写了也会被代理覆盖为动态值）；
`app_key/app_secret` 由代理进程的环境变量统一持有，不下发到 DataEase。

## 部署前自测（上游在线）

```bash
# 1) 健康检查
curl -s http://127.0.0.1:8077/healthz        # → ok

# 2) 假 key 走通透传链路：上游应返回 403/denied 类响应（证明链路通、仅凭据不对）
curl -is -XPOST http://127.0.0.1:8077/datasvc/execute \
  -H 'Content-Type: application/json' \
  -d '{"path":"/api/v1/demo/orders","params_json":"{\"limit\":10}"}'

# 3) 手工核对签名（与代理注入值一致性）
TS=$(date +%s)
printf '%s' "${APP_KEY}${TS}" | openssl dgst -sha256 -hmac "$APP_SECRET" | awk '{print $NF}'
```

换上真实 app_key/app_secret 后重跑 2)，应返回 `{"rows_json":...}`（该接口是唯一免
JWT 入口，返回体无 `{code,msg,data}` 包装壳）。

## systemd 单元示例

`/etc/systemd/system/signproxy.service`：

```ini
[Unit]
Description=DataEase dynamic sign proxy (datacenter datasvc)
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
# 建议以专用低权限用户运行；凭据放 EnvironmentFile（0600 权限）
User=datacenter
EnvironmentFile=/etc/datacenter/signproxy.env
ExecStart=/opt/datacenter/bin/signproxy
Restart=on-failure
RestartSec=3
LimitNOFILE=65535

[Install]
WantedBy=multi-user.target
```

`/etc/datacenter/signproxy.env`（chmod 600）：

```ini
APP_KEY=<app_key>
APP_SECRET=<app_secret>
PROXY_LISTEN=:8077
UPSTREAM=http://127.0.0.1:8887
```

启用：

```bash
systemctl daemon-reload && systemctl enable --now signproxy && systemctl status signproxy
journalctl -u signproxy -f   # 访问日志：时间/path/耗时/上游状态码
```
