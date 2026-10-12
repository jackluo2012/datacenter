package logic

import (
	"encoding/json"
	"regexp"
	"strings"

	"datacenter/ai/rpc/ai"
)

// prompt 相关常量：候选文本截断上限（防异常超长元数据撑爆 prompt/token）
const (
	maxCandidateTextRunes = 300  //候选单条 name/detail/comment 文本截断上限（字符）
	maxContextJsonRunes   = 2000 //context_json 入参截断上限（字符），防历史上下文无限增长
)

// ---------- 系统提示（角色 + 严格规则 + 输出格式 + few-shot，对标成熟 ChatBI 提示词结构） ----------

// askRolePrompt 角色定义：让模型锚定"数据分析师"身份与目标（自然语言→一条只读查询）
const askRolePrompt = "你是企业数据中台的 ChatBI 数据分析师，负责把用户的自然语言问题改写为一条可在 MySQL 上执行的只读查询。"

// askRulesPrompt 严格规则：只读/只准用候选列/聚合口径/LIMIT/禁止编造，逐条编号方便模型对齐
const askRulesPrompt = "\n\n【严格规则】\n" +
	"1. 只能生成一条 SELECT 查询语句，禁止 INSERT/UPDATE/DELETE/DROP/ALTER/TRUNCATE 等任何写操作，禁止多条语句；\n" +
	"2. 只能使用下方上下文中给出的候选表、候选指标及其列信息，禁止编造任何表名或列名；候选未覆盖到的表或列，在回答文字中说明，不要硬写进 SQL；\n" +
	"3. 金额、数量等可加总的指标一律用聚合函数（金额类优先 SUM(amount)），已登记的指标以候选指标的口径表达式为准，不要自行改写口径；\n" +
	"4. 查询必须带 LIMIT 100（明细防刷库，聚合/分组结果同样补 LIMIT 100）；\n" +
	"5. 过滤取消状态时约定使用 status <> 'cancelled'（演示库订单表口径）；\n" +
	"6. 问题与候选完全不相关、或无法确定表与列时，不要输出 sql 代码块，用一两句话说明缺少什么元数据。"

// askOutputPrompt 输出格式约束：首行"根据<表>分析：" + ```sql 代码块（供 extractSQLBlock 稳定提取）
const askOutputPrompt = "\n\n【输出格式】\n" +
	"严格遵守：第一行写\"根据<表名>分析：\"，随后紧跟一个 ```sql 代码块，代码块内只放这一条 SELECT，" +
	"代码块外不要输出第二条语句或其它 SQL。"

// askFewShots 少样本示例（对照演示库 dm_orders/dm_users，两个最常用问法：聚合/分组），
// 演示数据确定性生成，模型照抄口径即可得到可执行 SQL
const askFewShots = "\n\n【示例】\n" +
	"问：总销售额是多少\n" +
	"根据dm_orders分析：\n" +
	"```sql\n" +
	"SELECT SUM(amount) AS total_amount FROM dm_orders WHERE status <> 'cancelled' LIMIT 100\n" +
	"```\n" +
	"\n" +
	"问：用户的省份分布\n" +
	"根据dm_users分析：\n" +
	"```sql\n" +
	"SELECT province, COUNT(*) AS user_cnt FROM dm_users GROUP BY province ORDER BY user_cnt DESC LIMIT 100\n" +
	"```"

// buildAskSystemPrompt 组装系统提示：角色 + 严格规则 + 输出格式 + few-shot（独立函数便于单测）
func buildAskSystemPrompt() string {
	return askRolePrompt + askRulesPrompt + askOutputPrompt + askFewShots
}

// ---------- 用户消息（候选上下文 + 多轮上下文 + 问题） ----------

// askContext AskReq.context_json 的载荷：携带上一轮问题与生成 SQL，辅助追问/改写类多轮场景
type askContext struct {
	Question string `json:"question"` //上一轮用户问题（可空）
	Sql      string `json:"sql"`      //上一轮生成的 SQL（可空）
}

// parseAskContext 解析多轮上下文 JSON（容错设计：空串/非法 JSON/字段全空均视为无上下文，
// 不报错不阻断问数主流程——多轮是锦上添花能力，坏了就当首轮问）
func parseAskContext(raw string) askContext {
	var c askContext
	raw = strings.TrimSpace(truncateRunes(raw, maxContextJsonRunes))
	if raw == "" {
		return c
	}
	if err := json.Unmarshal([]byte(raw), &c); err != nil {
		return askContext{}
	}
	return c
}

// buildAskUserPrompt 组装用户消息：候选表/指标上下文（含口径）+ 可选多轮上下文 + 用户问题
func buildAskUserPrompt(question string, candidates []*ai.Candidate, contextJson string) string {
	var b strings.Builder

	//1. 候选上下文：表与指标分组罗列，detail(fqn/口径表达式)与 comment 全量注入（单条截断防爆）
	entities, metrics := splitCandidates(candidates)
	if len(entities) > 0 {
		b.WriteString("【候选表】（fqn=库表全限定名，comment=表用途/列信息说明）\n")
		for _, c := range entities {
			b.WriteString("- 表名:")
			b.WriteString(clipPromptText(c.Name))
			b.WriteString(" | fqn: ")
			b.WriteString(clipPromptText(c.Detail))
			if c.Comment != "" {
				b.WriteString(" | 说明: ")
				b.WriteString(clipPromptText(c.Comment))
			}
			b.WriteString("\n")
		}
	}
	if len(metrics) > 0 {
		b.WriteString("\n【候选指标】（expr=口径表达式，comment=口径说明，SQL 中按口径引用）\n")
		for _, c := range metrics {
			b.WriteString("- 指标:")
			b.WriteString(clipPromptText(c.Name))
			b.WriteString(" | expr: ")
			b.WriteString(clipPromptText(c.Detail))
			if c.Comment != "" {
				b.WriteString(" | 说明: ")
				b.WriteString(clipPromptText(c.Comment))
			}
			b.WriteString("\n")
		}
	}
	if len(entities) == 0 && len(metrics) == 0 {
		b.WriteString("【候选表/候选指标】\n（语义检索未命中任何候选，请按规则第 2/6 条处理，禁止编造表名）\n")
	}

	//2. 多轮上下文（可选）：注入上一轮问题与 SQL，支持"那按省份拆分呢"之类的追问改写
	if ctx := parseAskContext(contextJson); ctx.Question != "" || ctx.Sql != "" {
		b.WriteString("\n【多轮上下文】（上一轮对话，供追问参考；本轮问题可能基于上一轮省略主语）\n")
		if ctx.Question != "" {
			b.WriteString("上一轮问题：" + clipPromptText(ctx.Question) + "\n")
		}
		if ctx.Sql != "" {
			b.WriteString("上一轮 SQL：\n```sql\n" + clipPromptText(ctx.Sql) + "\n```\n")
		}
	}

	//3. 本轮用户问题 + 输出提醒
	b.WriteString("\n用户问题：")
	b.WriteString(question)
	b.WriteString("\n请严格按系统规则输出：第一行\"根据<表>分析：\"，随后一个 ```sql 代码块。")
	return b.String()
}

// splitCandidates 候选按类型分组（entity 表 / metric 指标，其余类型归入表组兜底）
func splitCandidates(candidates []*ai.Candidate) (entities, metrics []*ai.Candidate) {
	for _, c := range candidates {
		if c == nil {
			continue
		}
		if c.Type == "metric" {
			metrics = append(metrics, c)
			continue
		}
		entities = append(entities, c)
	}
	return entities, metrics
}

// clipPromptText 候选文本截断（按字符，超长补省略号），防止异常超长元数据撑爆 prompt
func clipPromptText(s string) string {
	s = strings.TrimSpace(s)
	if r := []rune(s); len(r) > maxCandidateTextRunes {
		return string(r[:maxCandidateTextRunes]) + "…"
	}
	return s
}

// ---------- SQL 提取 ----------

// sqlFenceRe ```sql ... ``` 代码块（忽略大小写，跨行）
var sqlFenceRe = regexp.MustCompile(`(?is)` + "```sql\\s*(.*?)```")

// anyFenceRe 任意 ``` ... ``` 代码块兜底
var anyFenceRe = regexp.MustCompile(`(?is)` + "```\\s*(.*?)```")

// extractSQLBlock 从 LLM 回复中提取 SQL：
// 优先 ```sql ... ``` 代码块，其次任意 ``` ... ``` 代码块；都没有时仅当裸文本形似 SQL
// （剥掉围栏残留后以 select 开头）才采信，否则返回空串——模型按输出契约合规拒答的
// 纯文字说明（如"缺少候选元数据"）不应被误当 SQL 送安全校验/执行
func extractSQLBlock(content string) string {
	if m := sqlFenceRe.FindStringSubmatch(content); len(m) > 1 {
		return strings.TrimSpace(m[1])
	}
	if m := anyFenceRe.FindStringSubmatch(content); len(m) > 1 {
		return strings.TrimSpace(m[1])
	}
	//裸文本兜底：剥离未闭合围栏标记后，仅保留形似 SQL 的文本
	t := strings.TrimSpace(content)
	t = strings.TrimPrefix(t, "```sql")
	t = strings.TrimPrefix(t, "```")
	t = strings.TrimSuffix(t, "```")
	t = strings.TrimSpace(t)
	if !strings.HasPrefix(strings.ToLower(t), "select") {
		return ""
	}
	return t
}
