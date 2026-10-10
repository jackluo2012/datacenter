package logic

import (
	"strings"
	"testing"

	"datacenter/ai/rpc/ai"
)

// 测试 prompt 构造：系统提示 + 候选 JSON + 用户问题
func TestBuildAskPrompt(t *testing.T) {
	candidates := []*ai.Candidate{
		{Type: "entity", Id: 1, Name: "订单表", Detail: "mysql://mall.orders", Comment: "订单主表"},
		{Type: "metric", Id: 2, Name: "GMV", Detail: "sum(amount)", Comment: "成交金额口径"},
	}
	prompt := buildAskPrompt("上月gmv是多少", candidates)

	//系统提示（含只读 SQL 约束）
	if !strings.Contains(prompt, "你是数据助手") || !strings.Contains(prompt, "只读") {
		t.Errorf("prompt 缺少系统提示: %s", prompt)
	}
	//候选 JSON（含字段名与值）
	for _, want := range []string{`"type":"entity"`, `"name":"订单表"`, `"detail":"mysql://mall.orders"`, `"name":"GMV"`, `"detail":"sum(amount)"`} {
		if !strings.Contains(prompt, want) {
			t.Errorf("prompt 缺少候选 JSON 片段 %s: %s", want, prompt)
		}
	}
	//用户问题
	if !strings.Contains(prompt, "用户问题：上月gmv是多少") {
		t.Errorf("prompt 缺少用户问题: %s", prompt)
	}
	//SQL 代码块约定（供 extractSQLBlock 提取）
	if !strings.Contains(prompt, "```sql") {
		t.Errorf("prompt 缺少 sql 代码块约定: %s", prompt)
	}
}

// 测试空候选 prompt：JSON 为空数组
func TestBuildAskPromptEmptyCandidates(t *testing.T) {
	prompt := buildAskPrompt("有多少用户", nil)
	if !strings.Contains(prompt, "[]") {
		t.Errorf("空候选 prompt 应含空数组 JSON: %s", prompt)
	}
}

// 测试 SQL 代码块提取：```sql 块 > 任意 ``` 块 > 裸文本兜底
func TestExtractSQLBlock(t *testing.T) {
	cases := []struct {
		name    string
		content string
		want    string
	}{
		{
			name:    "标准 sql 代码块",
			content: "查询如下：\n```sql\nselect count(*) from mall.orders\n```\n以上。",
			want:    "select count(*) from mall.orders",
		},
		{
			name:    "SQL 大写围栏与多行",
			content: "```SQL\nselect 1\nfrom t\n```",
			want:    "select 1\nfrom t",
		},
		{
			name:    "无语言标注代码块兜底",
			content: "```\nselect 2\n```",
			want:    "select 2",
		},
		{
			name:    "裸文本兜底",
			content: "select 3 from t where id = 1",
			want:    "select 3 from t where id = 1",
		},
		{
			name:    "未闭合 sql 围栏剥离",
			content: "```sql\nselect 4",
			want:    "select 4",
		},
		{
			name:    "空内容",
			content: "   ",
			want:    "",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := extractSQLBlock(c.content); got != c.want {
				t.Errorf("extractSQLBlock(%q) = %q, 期望 %q", c.content, got, c.want)
			}
		})
	}
}

// 测试检索摘要文本
func TestBuildSearchSummary(t *testing.T) {
	if s := buildSearchSummary(nil); !strings.Contains(s, "未命中任何候选") {
		t.Errorf("空候选摘要 = %q, 期望包含未命中提示", s)
	}
	cands := []*ai.Candidate{
		{Type: "entity", Id: 1, Name: "订单表", Detail: "mysql://mall.orders"},
		{Type: "metric", Id: 2, Name: "GMV", Detail: "sum(amount)"},
	}
	s := buildSearchSummary(cands)
	for _, want := range []string{"命中 2 个候选", "[entity] 订单表(mysql://mall.orders)", "[metric] GMV(sum(amount))"} {
		if !strings.Contains(s, want) {
			t.Errorf("摘要缺少 %q: %s", want, s)
		}
	}
}
