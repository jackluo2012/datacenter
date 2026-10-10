# AD部分
有整套商业版数据中台源码出售，基于doris实现的数据仓库，支持PB级别的数据存储分析，功能模块如下：

## 功能模块
### 1、数据源管理
支持各类关系型数据库、非关系型数据库、MQ、文件源等。

### 2、元数据管理
元模型、最新元数据、定版元数据、数据全景图、数据血缘、数据影响分析、多版本元数据对比

### 3、数据标准管理
标准词根、标准字典、数据元、标准模型、发布、多版本维护、数据标准核对

### 4、数据仓库管理
支持主题域、主题、数仓集群、维度建模、模型运维、模型审计、模型数据查、物化视图

### 5、数据质量
规则定义，任务执行，结果查看，统计分析，质量问题修复日志

### 6、数据指标
数据指标在线开发，数据指标地图、数据指标看板

### 7、数据标签
标签对象、标签管理、置标任务、标签圈群、标签画像

### 8、数据安全
分级分类、数据脱敏、分级分类授权

### 9、数据生命周期
数据归档，数据恢复

### 10、数据服务
接口在线开发(支持通过JS脚本对数据进行处理后返回，支持动态SQL)，API注册(手写代码开发的接口注册到数据服务中)，接口测试，接口发布，应用管理，应用授权

### 11、数据资产
数据资产标签、数据资产目录、数据资产门户、数据资产申请使用、AI问数

### 10、数据集成/开发
实时开发(支持CDC)、离线开发(支持数据库、csv、excel、接口、ftp、kafka、mqtt、mongodb等数据源接入)，在线拖拉拽生成数据集成任务

### 11、数据可视化
数据集，报表管理，报表设计，报表查看、知识图谱构建、知识图谱查看

### 12、数据运维
微服务监控、中间件监控、服务器监控、数据仓库定时备份恢复、数据服务接口监控。

### 13、系统日志
syslog查询、syslog实时查看、登录日志、审计日志

### 14、主数据
主数据模型管理、自动生成代码码段管理、主数据管理、主数据分发

技术栈    
后端：Java springboot2.7 springcloud/alibaba mybatis plus hutool 等常见技术    
前端：vue elementui vite 等常见技术    
中间件：doris，mysql，redis，rabbitmq，minio，zookeeper。    
开源版后端：Go 1.26 + go-zero v1.10.3 + MySQL 8.4 + Redis 8 + etcd 3.5 + Elasticsearch 7.17    
 
有演示环境和方案PPT，需要加V:jackluo1989    

# 以下为本开源项目正文

# datacenter：基于 go-zero 的开源数据中台

> **2026-10 全栈升级**：Go 1.26 + go-zero v1.10.3 + MySQL 8.4 + Redis 8 + etcd 3.5 + ES 7.17，升级明细见 [docs/UPGRADE.md](docs/UPGRADE.md)。
> 项目正在从"活动业务中台"向真正的数据中台演进，路线图见 [docs/ROADMAP.md](docs/ROADMAP.md)。

### 环境要求
- Go **1.24+**（go.mod 声明 1.26）
- goctl v1.10.2（`go install github.com/zeromicro/go-zero/tools/goctl@latest`，仅重新生成代码时需要）
- protoc 36.x + protoc-gen-go + protoc-gen-go-grpc（仅重新生成 RPC 时需要）

### 详细介绍 [我是如何用go-zero实现一个中台系统的](https://www.cnblogs.com/jackluo/p/14148518.html)

## 架构图
![中台系统](https://img2020.cnblogs.com/blog/203395/202012/203395-20201217094615171-335437652.jpg "中台架构")

### 已完成的功能列表
- [x] 微信公众号登陆
- [x] 七牛上传获取token
- [x] 投票
    - [x] 报名
    - [x] 报名列表
    - [x] 投票
- [x] 抽奖问答
    - [x] 活动信息
    - [x] 问答列表
    - [x] 提交答案 
    - [x] 获取得分 
    - [x] 抽奖
    - [x] 填写中奖人信息    
- [x] 搜索
    - [x] 基于elasticsearch
- [x] 数据中台十服务（数据源 / 元数据 / 数据服务 API / 资产目录 / 质量 / 指标 / 标签 / 调度 / SQL IDE / AI 问数），详见 [数据中台能力](#数据中台能力)

### 未完成的        
- [ ] 微信支付宝登陆
- [ ] 微信支付宝支付

## 数据中台能力
> 项目正从上面的活动业务中台向真正的数据中台演进，Phase 1-4 共 10 个服务已交付。

### Phase 1 已完成（datasource / metadata / datasvc / catalog）
- **数据源管理**（datasource）：统一登记各类型数据源，凭证 AES-GCM 加密存储，支持连接探活（时延/错误信息）
- **元数据管理**（metadata）：基于 information_schema 自动采集库表结构，表级血缘登记与查询，元数据定版快照与多版本对比
- **数据服务 API**（datasvc，OneService 模式：配置出 API、统一管控）：HMAC 签名校验 → 应用限流 → 授权校验 → 参数白名单 → 结果缓存 → 出口字段脱敏
- **资产目录**（catalog）：资产分类、资产登记、资产申请与审批

### 快速体验
```shell
sh server.sh                  # 1. docker 起依赖（mysql 8.4 / redis 8 / etcd 3.5 / es 7.17）
mysql -uroot -padmin          # 2. 导入表结构：source sql.sql 后依次 source docs/schema/phase1.sql ~ phase3.sql
sh restart.sh                 # 3. 编译并重启全部服务（网关端口 8887）
```

下面给出关键步骤的 curl 示例（网关 `http://127.0.0.1:8887`），完整流程为：登记数据源 → 采集元数据 → 建 API（`/datasvc/api/create`）→ 注册应用 → 授权（`/datasvc/grant`）→ 发布（`/datasvc/api/publish`）→ 执行：

```shell
# 登记数据源（凭证 AES-GCM 加密入库）
curl -XPOST http://127.0.0.1:8887/datasource/create -d '{"name":"业务库","type":"mysql","host":"127.0.0.1","port":3306,"db_name":"datacenter","username":"root","password":"admin"}'

# 采集元数据（information_schema → 元数据实体）
curl -XPOST http://127.0.0.1:8887/metadata/collect -d '{"ds_id":1}'

# 注册应用（返回 app_key/app_secret，secret 仅注册时返回一次）
curl -XPOST http://127.0.0.1:8887/datasvc/app/register -d '{"name":"demo-app","rate_limit_qps":10}'

# 执行 API（sign = hex(hmac-sha256(app_secret, app_key+ts))，返回 rows_json/cost_ms/cache_hit）
curl -XPOST http://127.0.0.1:8887/datasvc/execute -d '{"app_key":"<app_key>","ts":"<unix秒>","sign":"<hex签名>","path":"/api/v1/demo","params_json":"{}"}'
```

### Phase 2 已完成（quality / metric / tag）
- **数据质量**（quality）：规则定义、质量任务执行、结果查看（`docs/schema/phase2.sql`）
- **数据指标**（metric）：原子/派生指标定义与指标关系，指标可被数据服务引用成 API
- **数据标签**（tag）：标签对象、标签定义、标签值（规则圈群、画像）

### Phase 3/4 已交付（job / ide / ai）
- **调度集成**（job）：同步任务登记与 DolphinScheduler 管控面对接，真实环境联调见 [#21](https://github.com/jackluo2012/datacenter/issues/21)
- **SQL IDE**（ide）：SQL 脚本在线开发与版本管理
- **AI 问数**（ai）：语义层（NL2Metric）问数编排与问数日志，LLM 对接待真实环境联调，见 [#23](https://github.com/jackluo2012/datacenter/issues/23)；表结构见 `docs/schema/phase3.sql`

### 相关文档
- [docs/ROADMAP.md](docs/ROADMAP.md)：演进路线
- [docs/UPGRADE.md](docs/UPGRADE.md)：升级记录（Go 1.26 / go-zero v1.10.3 全栈现代化）
- [docs/PHASE1.md](docs/PHASE1.md)：Phase 1 开发契约（服务端口、表结构、架构约定）

## 项目结构

| 目录 | 职责 |
|---|---|
| `datacenter.go` + `internal/` | 网关（:8887）：路由聚合、JWT 登录校验、防越权中间件、跨服务编排 |
| `datasource/` | 数据源管理：登记、凭证 AES-GCM 加密、连接探活 |
| `metadata/` | 元数据：库表自动采集、表级血缘、定版快照与版本对比 |
| `datasvc/` | 数据服务 API（OneService）：配置出 API，签名/限流/授权/缓存/脱敏 |
| `catalog/` | 资产目录：资产分类、登记、申请与审批 |
| `quality/` | 数据质量：规则、任务执行、结果 |
| `metric/` | 数据指标：原子/派生指标定义与关系 |
| `tag/` | 数据标签：标签对象、定义、圈群画像 |
| `job/` | 调度集成：同步任务登记、对接 DolphinScheduler 管控面 |
| `ide/` | SQL IDE：脚本在线开发与版本管理 |
| `ai/` | AI 问数：语义层（NL2Metric）问数编排 |
| `user/` `common/` `votes/` `search/` `questions/` | 遗留业务 RPC：登录、公共、投票、搜索、问答抽奖 |
| `docs/` | [ROADMAP](docs/ROADMAP.md)（路线图）、[UPGRADE](docs/UPGRADE.md)（升级记录）、[PHASE1](docs/PHASE1.md)（开发契约）、[schema](docs/schema/phase1.sql)（phase1~3 表结构） |

## 如何运行

### 方式一（推荐）：docker compose 一键编排
```shell
docker compose build          # 首次构建镜像，比较久
docker compose up -d          # 4 个依赖（mysql/redis/etcd/es）+ 11 个服务，mysql 健康后按序启动
docker compose ps             # 查看状态；日志：docker compose logs -f datacenter-api
```
- 仅网关 `8887` 对外（http://localhost:8887），其余容器网络内互通；
- 首次启动空库时自动按文件名顺序执行初始化 SQL：`sql.sql` → `docs/schema/phase1.sql` → `phase2.sql` → `phase3.sql`，无需手工导入；
- `user/common/votes/search/questions` 五个遗留 RPC 不在编排内（etcd 懒加载，网关可正常启动）：需要 `/user/login` 等接口时，按方式二另起对应服务后再调用。

### 方式二：restart.sh 本地编译启动

先用 docker 启动 mysql redis etcd elasticsearch 服务
```shell
sh server.sh
```
### 输出如下 就显示 成功了
``` 
mysql
mysql
8d5d4b381ab7abe8947f532422255cd172f214ab4a6b0533da1619259e1cc4a5
redis
redis
Start Redis Service...
1fc187a9d82f0942dd60cac76c723a5bc531e1b67424384d04e7a69dad1362f0
etcd
etcd
98f88d81e1e218d4d53c608e5a68cd70254df221bdb34c9beab37a7473971ba0
es
es
04b37b58f10411fa8ab5894c917266cc7bc7f9a96988908fff9b3734e6259ad4
```
### elasticsearch 的帐号密码已经通过 server.sh 内置（`ELASTIC_PASSWORD=admin`）
### 即帐号: elastic ,密码: admin，将其填入 search/rpc/etc/search.yaml 文件中即可



    接着导入 sql.sql到 mysql数据中 ,如果有工具自行导入,下面仅参考
```
mysql -uroot -padmin
mysql > set names utf8
mysql > create database datacenter
mysql > use datacenter
mysql > source sql.sql
mysql > source docs/schema/phase1.sql
mysql > source docs/schema/phase2.sql
mysql > source docs/schema/phase3.sql
```
### 然后分别把配置文件 ,文件下面分别对应了一个rpc.example.yaml的文件，复制，基本就没有问题

```
vi etc/datacenter-api.yaml #网关配置
vi user/rpc/etc/rpc.yaml #用户信息配置
vi common/rpc/etc/rpc.yaml #公共配置
vi votes/rpc/etc/rpc.yaml #投票配置
vi search/rpc/etc/search.yaml #搜索配置
vi questions/rpc/etc/questions.yaml #抽奖配置
```
### 然后启动 服务 ,应该我们要启动
```
sh restart.sh
```
### 输出如下
```
➜  datacenter.bak git:(master) ✗ sh restart.sh              
appending output to nohup.out
appending output to nohup.out
appending output to nohup.out
appending output to nohup.out 
appending output to nohup.out 
appending output to nohup.out    
```
### 可以分别查看是否启动成功
```
tail -F nohup.out  #网关的服务
tail -F user/rpc/nohup.out #用户的rpc服务
tail -F common/rpc/nohup.out #公共的
tail -F votes/rpc/nohup.out #投票的
tail -F search/rpc/nohup.out #搜索的
tail -F questions/rpc/nohup.out #问答抽奖
```

### 在postman 导入数据 数据中台中心.postman_collection.json  就可以很愉快的玩耍了

### 登录与鉴权（受保护接口需要 JWT）
- 先调 `POST /user/login` 获取 `access_token`（登录返回 `access_token` / `access_expire`）；
- 请求受保护接口时请求头带 `Authorization: Bearer <access_token>`；
- 防越权接口还需带请求头 `x-user-id: <用户ID>`，且必须与 token 中的 uid 一致（见 `internal/middleware/usercheckmiddleware.go`）。
