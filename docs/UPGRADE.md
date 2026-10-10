# 升级记录：go-zero v1.3.0 → v1.10.3（2026-10-10）

## 版本对照

| 组件 | 升级前 | 升级后 |
|---|---|---|
| Go | 1.17 | **1.26.0**（go.mod）/ 工具链 1.27.2 |
| go-zero | v1.3.0 (2022) | **v1.10.3** (2026-08) |
| goctl | v1.3.0 | **v1.10.2** |
| grpc | v1.43.0 | v1.84.0 |
| protobuf | v1.27.1 | v1.36.12（protoc 36.2） |
| jwt | dgrijalva/jwt-go v3.2.0（停维护，**CVE-2020-26160**） | **golang-jwt/jwt/v4 v4.5.2**（与 go-zero 依赖对齐） |
| 七牛 SDK | qiniu/api.v7/v7 v7.8.2（仓库已归档） | **qiniu/go-sdk/v7 v7.29.0**（仅换 import，业务代码零修改） |
| elastic | olivere/elastic/v7 v7.0.28 | v7.0.32 |
| MySQL 镜像 | mysql:8.0.21 (2020) | **mysql:8.4** |
| Redis | sameersbn/redis（陈旧） | **redis:8-alpine** |
| etcd | quay.io/coreos/etcd（v3.3 时代） | **etcd v3.5.21** |
| Elasticsearch | spencezhou/elasticsearch:7.6.2（个人镜像） | **docker.elastic.co 官方 7.17.28**（密码用 `ELASTIC_PASSWORD=admin` 一键设置，不再需要手工 setup-passwords） |
| MySQL 字符集 | utf8(mb3) 与 utf8mb4 混用 | **全库 utf8mb4** |

## 代码改动清单

1. **合并双 main**：删除 user/votes/common 三个 rpc 目录下与手写 `rpc.go` 重复的 goctl 生成 main（`user.go`/`votes.go`/`common.go`），`go build ./...` 恢复可用。
2. **`redis.NewRedis` → `redis.MustNewRedis(c.CacheRedis[0].RedisConf)`**（3 处 ServiceContext，签名变更的编译级必改）。
3. **jwt import** 换 `golang-jwt/jwt/v4`（2 处，签发代码零修改）。
4. **`ioutil.ReadAll` → `io.ReadAll`**（shared/tools.go）。
5. **存量 bug 修复**（vet 暴露）：
   - `votes/model/appvotesmodel.go:83`：Sprintf 占位符与参数不匹配，**按次数计票 SQL 一直是坏的**（`from <字段列表>` 而非 `from app_votes`）；
   - `search/rpc/dao/article.go:70`：`logx.Info` 误用格式化占位符 → `logx.Infof`；
   - `shared/tools_test.go`：写死占位 URL 的必挂测试 → 改为 httptest 常规测试，并补充 `ToLimitOffset` 用例。
6. **gofmt 全量格式化**（10 个历史未格式化文件）。
7. **`.api` 文件**：全部补 `syntax = "v1"` 头（新版 goctl 硬要求）；`type Xxx struct {` → `type Xxx {`（新语法）。
8. **server.sh**：镜像全面升级（见上表），并给 stop/rm 加 `2>/dev/null` 容错。

## ⚠️ 最重要的破坏性变更：参数解析 tag 规则（v1.3.0 → v1.8+）

go-zero 新版 `httpx.Parse` 逐阶段（path → form → header → json）用**各自专用 tag** 解析，且任何字段被某一阶段认领后就做 required 校验：

- 旧版：`json` tag 可同时用于 query/form 解析；
- 新版：**query/form 只认 `form` tag**；而 GET 请求虽无 body，json 阶段仍会对 `json` tag 的 required 字段做空 map 校验 → json-only tag 的 GET 接口**必挂** `field "xxx" is not set`（官方 goctl 最新版生成的代码同样触发，属上游已知行为）。

**迁移格式**（本仓库已统一采用）：

```go
// GET query 参数（.api 与 types.go 保持一致）
Beid int64 `json:"beid,optional" form:"beid,optional"`
// POST JSON body 参数：保持 json-only，required 语义由 body 提供
Mobile string `json:"mobile"`
```

> 禁止 `json:"xxx"` + `form:"xxx"` 双 required 组合（POST JSON 请求会挂在 form 阶段）。
> GET 参数 required 校验由 form 阶段承担；多方法共用类型（如 `VoteReq` 同时用于 GET/POST）统一 `,optional`，必填校验下沉到业务层（voteslogic 本就有完整校验）。

涉及类型：`Beid`、`Actid`、`VoteReq`、`SearchReq`、`UserReq`（其余 POST-only 类型未动）。

## 其他行为变更（影响感知，无需改码）

- `zrpc.RpcClientConf.NonBlock` 默认值 false → **true**：网关启动不再等待后端 RPC 就绪（服务可先起）。
- trace 的 `Batcher: jaeger` 已移除（v1.10.0），配置 jaeger 需走 `otlpgrpc`；本项目未配置 trace，无影响。
- MySQL 8.4：DSN 保持 `charset=utf8mb4&parseTime=true&loc=Asia/Shanghai`；如遇 `Public Key Retrieval is not allowed`，DSN 加 `allowPublicKeyRetrieval=true`（生产建议 TLS）。

## 验证结果（本机冒烟，Docker：MySQL 8.4 + Redis 8 + etcd v3.5.21）

- `go build ./...` / `go vet ./...` / `go test ./...` ✅
- GET `/common/appinfo?beid=1`：网关→etcd→common.rpc→MySQL→Redis 全链路 ✅，缓存键 `cache#BaseApp#id#1` 生成 ✅
- POST `/common/snsinfo`：JSON body 解析 + 微信 login_url 组装 ✅
- POST `/user/login`：user.rpc → `FindOneByMobile` 链路 ✅（返回业务态"用户未注册"）
- 参数校验语义保持：GET 缺参仍返回 `field "beid" is not set` ✅

## 遗留事项

| 事项 | 说明 |
|---|---|
| `chanxuehong/wechat` 僵尸依赖 | 编译可用但 5 年未更新。迁移方案：`silenceper/wechat/v2 v2.1.14`（OAuth：`officialaccount/oauth`；JS-SDK：`officialaccount/js`），涉及 `wxticketlogic.go` + `snsloginlogic.go` 约百行胶水重写，**需真实公众号凭据回归**，故本次未动 |
| pb 未重新生成 | 旧 pb.go（protoc-gen-go v1.25）与新 grpc v1.84 兼容，全链路验证通过；待需要新字段时用 `goctl rpc protoc` 整套重生成（勿新旧混放） |
| vet "lock by value" 24 处 | 网关 logic/handler 以值传递 pb 结构的历史模式，随 pb 重生成一并改为指针传递 |
| 鉴权中间件被注释 | `usercheckmiddleware.go` 校验逻辑整段注释（升级前即如此），生产前必须恢复 |
| 全量 regen | goctl 不会覆盖已存在文件；如需 regen 须先备份删除旧生成物再 diff 合入 |

## 工具链安装（本机，用户空间）

```bash
# Go 1.27.2
curl -o /tmp/go.tar.gz https://go.dev/dl/go1.27.2.linux-amd64.tar.gz
mkdir -p ~/.local && tar -C ~/.local -xzf /tmp/go.tar.gz
export PATH=$HOME/.local/go/bin:$HOME/go/bin:$PATH

# goctl + protoc 工具链
go install github.com/zeromicro/go-zero/tools/goctl@latest
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
# protoc 36.2 位于 ~/.local/protoc/bin
```
