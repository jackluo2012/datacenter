# AI 问数 LLM 接入指南（ai-rpc）

AI 问数（`Ask` / `AskAndRun`）走 **OpenAI 兼容** `POST {BaseUrl}/chat/completions`，
Bearer 鉴权、温度固定 0（问数要求确定性输出）、单次请求 30s 超时、响应体上限 1MB。
凡兼容 OpenAI 协议的供应商/网关（DeepSeek、通义千问兼容模式、OpenAI、本地 vLLM/Ollama 等）均可直接接入。

## 1. 配置项逐字段说明（ai/rpc/etc/ai.yaml 的 Llm 段）

| 字段 | 必填 | 说明 |
|---|---|---|
| `Enabled` | 是 | LLM 总开关。`false` 或调用失败时**自动降级为语义检索**（见第 4 节），不报错 |
| `BaseUrl` | 启用时必填 | OpenAI 兼容接口根地址，**无需**自带 `/chat/completions`（代码自动拼接），末尾多余 `/` 自动清理 |
| `ApiKey` | 启用时必填 | 供应商密钥，以 `Authorization: Bearer <ApiKey>` 头发送 |
| `Model` | 否 | 模型名，缺省 `gpt-4o-mini`；须用供应商提供的对话模型名 |

## 2. 三家供应商配置示例（逐字段可复制）

### 2.1 DeepSeek（推荐演示用：便宜、中文好、完全 OpenAI 兼容）

```yaml
Llm:
  Enabled: true
  BaseUrl: https://api.deepseek.com        #也兼容 https://api.deepseek.com/v1
  ApiKey: sk-xxxxxxxxxxxxxxxxxxxxxxxx      #DeepSeek 开放平台 → API Keys 页创建
  Model: deepseek-chat                     #V3 对话模型；推理模型为 deepseek-reasoner(不推荐问数用,慢)
```

- 控制台：https://platform.deepseek.com ；计费为Token 计量，问数单次 prompt 约 1-2K token，演示一整天成本可忽略。

### 2.2 通义千问（阿里云百炼 DashScope，OpenAI 兼容模式）

```yaml
Llm:
  Enabled: true
  BaseUrl: https://dashscope.aliyuncs.com/compatible-mode/v1   #必须带 compatible-mode 路径
  ApiKey: sk-xxxxxxxxxxxxxxxxxxxxxxxx      #阿里云百炼 → API-KEY 管理
  Model: qwen-plus                         #性价比首选；更快选 qwen-turbo，更强选 qwen-max
```

- 控制台：https://bailian.console.aliyun.com ；注意是**兼容模式**地址（`/compatible-mode/v1`），
  不要填原生 DashScope 地址（原生协议非 OpenAI 格式）。

### 2.3 OpenAI

```yaml
Llm:
  Enabled: true
  BaseUrl: https://api.openai.com/v1
  ApiKey: sk-xxxxxxxxxxxxxxxxxxxxxxxx      #platform.openai.com → API keys
  Model: gpt-4o-mini                       #便宜够用；追求质量换 gpt-4o
```

> 国内网络环境建议给 `BaseUrl` 挂自建网关/代理（同样保持 OpenAI 兼容路径），代码无需任何改动。
> 本地离线演示也可用 Ollama：`BaseUrl: http://127.0.0.1:11434/v1`、`Model: qwen2.5:7b`、`ApiKey` 留空即可。

## 3. 修改配置后生效方式

LLM 配置**不做热更新**：改完 `ai.yaml` 重启 ai-rpc（`ai` 服务）即可。
建议先 `cp ai.example.yaml ai.yaml` 再改，`ai.yaml` 已被 .gitignore 排除，不会把 key 提交进仓库。

## 4. 无 Key / 调用失败时的降级行为（无需额外配置）

| 场景 | 行为 |
|---|---|
| `Enabled: false`（无 key 的默认态） | `Ask` 返回**语义检索摘要**：命中候选清单 + 口径引导，`llm_enabled=false`，接口不报错 |
| `Enabled: true` 但 key 无效/欠费/网络不通 | 同上降级，`answer` 尾部附`[LLM 调用失败已降级为纯检索] 错误摘要: ...`，方便现场排障 |
| LLM 生成了非 SELECT / 多语句内容 | 发送侧安全校验拦截：`generated_sql` 置空，`answer` 附`[安全校验未通过]`说明 |
| `AskAndRun` 在以上任一降级下 | 不执行 SQL，`run_error` 写明原因（未启用 LLM / 安全校验拦截），`executed=false` |

即使 LLM 侧被绕过，执行侧仍有**二次防护**：AI 生成的 SQL 先存临时脚本，再经 ide-rpc 安全闸
（仅放行单条 SELECT、行数保护截断）真实执行——两级独立，写语句到不了数据库。

## 5. 演示建议（对照演示库 demo_data.sql 的确定性数据）

演示前置：元数据已采集（`fqn=datacenter.dm_xxx`，beid=2）、AI 问数候选来自元数据实体 + 指标。

| 问法 | 期望生成 SQL 形态 | 亮点 |
|---|---|---|
| 总销售额是多少 | `SELECT SUM(amount) ... WHERE status<>'cancelled'` | 命中 few-shot 示例，口径标准（不含取消单），金额约 300 万+ |
| 哪个省份用户最多 | `SELECT province, COUNT(*) FROM dm_users GROUP BY province ORDER BY ...` | 命中第二个示例，华东第一（200 人） |
| 最近90天每天订单量 | `SELECT DATE(created_at), COUNT(*) ... GROUP BY dt ORDER BY dt` | 逐日曲线不断点，还有两个大促峰值日（昨天与 8 天前，约 121 单/天） |
| 每种订单状态各多少单 | `SELECT status, COUNT(*) ... GROUP BY status` | paid/shipped/completed/cancelled 加权分布一目了然 |
| 卖得最好的5个商品 | `dm_order_items` 按 product_id 汇总 qty | 跨表口径，体现指标登记价值 |
| （追问）那按月份拆分呢 | 基于上一轮 SQL 改写 | 多轮对话：`Ask` 传 `context_json={"question":"上一问","sql":"上一轮SQL"}`，历史 SQL 会注入 prompt |

**演示效果最好的两问**：`总销售额是多少`（数字大、口径正）+ `哪个省份用户最多`（有排名有区分度）。
`AskAndRun`（`auto_run=true` + `datasource_id` 指向演示库数据源）能直接把数字跑出来，
配合 `rows_json` 现场出数据，是演示的 wow 点。

演示注意：
- `dm_orders.amount` ≈ 338 万（含 cancelled）；LLM 按 few-shot 口径剔除取消单，数字略小于手工全量 SUM，属**正确口径**，演示话术提前对齐；
- 生成的 SQL 里表名跟随元数据 `fqn`（如 `datacenter.dm_orders`），执行数据源必须是该库，否则报表不存在；
- 问句尽量带"销售额/订单/省份/分布"等实体词，语义检索才命中候选；候选为空时 LLM 会按规则拒答而不是编造表名。

## 6. 本地验证（无 key 环境）

```bash
export PATH=$HOME/.local/go/bin:$HOME/go/bin:$PATH
go build ./ai/... && go vet ./ai/... && go test -count=1 ./ai/...
```

单测覆盖（无需真实 key）：
- `internal/llm/llmclient_test.go`：httptest 模拟 OpenAI 响应，校验 URL/Bearer/30s 超时/非 200；
- `internal/logic/e2e_test.go`：httptest 模拟 OpenAI 全链路（system/user 双消息 → 生成 SQL 提取 → 安全校验 → ide 执行桩），含多轮上下文注入与写语句拦截、401 降级用例；
- `internal/logic/prompt_test.go`、`sqlguard_test.go`：prompt 结构与 SQL 安全校验的纯函数用例。
