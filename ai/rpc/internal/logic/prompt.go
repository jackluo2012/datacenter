package logic

import (
	"encoding/json"
	"regexp"
	"strings"

	"datacenter/ai/rpc/ai"
)

// askSystemPrompt 系统提示：约束 LLM 基于语义层生成只读 SQL
const askSystemPrompt = "你是数据助手，基于以下元数据/指标语义层生成只读 SQL：" +
	"只能生成 select 查询，禁止 insert/update/delete/drop 等任何写操作；" +
	"优先使用候选中的表（fqn 为全限定名）与指标口径；无法确定时明确说明，不要臆造表名或字段。"

// sqlFenceRe ```sql ... ``` 代码块（忽略大小写，跨行）
var sqlFenceRe = regexp.MustCompile(`(?is)` + "```sql\\s*(.*?)```")

// anyFenceRe 任意 ``` ... ``` 代码块兜底
var anyFenceRe = regexp.MustCompile(`(?is)` + "```\\s*(.*?)```")

// buildAskPrompt 构造问数 prompt：系统提示 + 候选 JSON + 用户问题（送 OpenAI 兼容 chat/completions）
func buildAskPrompt(question string, candidates []*ai.Candidate) string {
	var b strings.Builder
	b.WriteString(askSystemPrompt)
	b.WriteString("\n\n候选元数据(JSON)：\n")
	b.WriteString(marshalCandidates(candidates))
	b.WriteString("\n\n用户问题：")
	b.WriteString(question)
	b.WriteString("\n请基于候选元数据给出 SQL，并放入 ```sql 代码块中。")
	return b.String()
}

// marshalCandidates 候选序列化为 JSON（prompt 与审计共用；序列化失败兜底空数组）
func marshalCandidates(candidates []*ai.Candidate) string {
	list := make([]candidateJson, 0, len(candidates))
	for _, c := range candidates {
		if c == nil {
			continue
		}
		list = append(list, candidateJson{
			Type:    c.Type,
			Id:      c.Id,
			Name:    c.Name,
			Detail:  c.Detail,
			Comment: c.Comment,
		})
	}
	data, err := json.Marshal(list)
	if err != nil {
		return "[]"
	}
	return string(data)
}

// extractSQLBlock 从 LLM 回复中提取 SQL：
// 优先 ```sql ... ``` 代码块，其次任意 ``` ... ``` 代码块，都没有则裸文本整段兜底
func extractSQLBlock(content string) string {
	if m := sqlFenceRe.FindStringSubmatch(content); len(m) > 1 {
		return strings.TrimSpace(m[1])
	}
	if m := anyFenceRe.FindStringSubmatch(content); len(m) > 1 {
		return strings.TrimSpace(m[1])
	}
	//裸文本兜底：剥离未闭合围栏标记后原样返回
	t := strings.TrimSpace(content)
	t = strings.TrimPrefix(t, "```sql")
	t = strings.TrimPrefix(t, "```")
	t = strings.TrimSuffix(t, "```")
	return strings.TrimSpace(t)
}
