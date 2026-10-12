package logic

import (
	"strings"
	"testing"

	"datacenter/ai/rpc/ai"
)

// ---------- 系统提示 ----------

// 测试系统提示：角色、严格规则（只读/禁编造/SUM/LIMIT 100）、输出格式、few-shot 齐全
func TestBuildAskSystemPrompt(t *testing.T) {
	prompt := buildAskSystemPrompt()
	for _, want := range []string{
		"数据分析师",                 //角色锚定
		"SELECT",                //只读约束
		"禁止编造",                  //禁止幻觉表/列
		"SUM",                   //金额聚合口径
		"LIMIT 100",             //行数上限
		"根据<表名>分析：",             //输出格式首行
		"```sql",                //代码块约定（extractSQLBlock 依赖）
		"【示例】",                  //few-shot 区块
		"dm_orders",             //示例 1：总销售额
		"SUM(amount)",           //示例 1 口径
		"status <> 'cancelled'", //示例 1 取消单过滤
		"dm_users",              //示例 2：省份分布
		"GROUP BY province",     //示例 2 分组
	} {
		if !strings.Contains(prompt, want) {
			t.Errorf("系统提示缺少 %q", want)
		}
	}
	//few-shot 示例必须是单条无分号 SELECT（与安全闸口径一致，示范即是约束）
	if strings.Count(prompt, "```sql") != 3 { //输出格式 1 处 + 示例 2 处
		t.Errorf("```sql 出现 %d 处, 期望 3 处(格式约定+两个示例)", strings.Count(prompt, "```sql"))
	}
}

// ---------- 用户消息 ----------

// 测试用户消息构造：候选表/指标上下文（含口径与说明）+ 用户问题 + 输出提醒
func TestBuildAskUserPrompt(t *testing.T) {
	candidates := []*ai.Candidate{
		{Type: "entity", Id: 1, Name: "订单表", Detail: "dm_orders", Comment: "演示-电商订单表(amount=订单金额)"},
		{Type: "metric", Id: 2, Name: "GMV", Detail: "sum(amount)", Comment: "成交金额口径"},
	}
	prompt := buildAskUserPrompt("上月gmv是多少", candidates, "")

	//候选表区块：表名/fqn/说明
	for _, want := range []string{"【候选表】", "表名:订单表", "fqn: dm_orders", "说明: 演示-电商订单表"} {
		if !strings.Contains(prompt, want) {
			t.Errorf("用户消息缺少候选表片段 %q: %s", want, prompt)
		}
	}
	//候选指标区块：口径表达式与说明
	for _, want := range []string{"【候选指标】", "指标:GMV", "expr: sum(amount)", "成交金额口径"} {
		if !strings.Contains(prompt, want) {
			t.Errorf("用户消息缺少候选指标片段 %q: %s", want, prompt)
		}
	}
	//用户问题与输出提醒
	if !strings.Contains(prompt, "用户问题：上月gmv是多少") {
		t.Errorf("用户消息缺少问题: %s", prompt)
	}
	if !strings.Contains(prompt, "```sql") {
		t.Errorf("用户消息缺少 sql 代码块提醒: %s", prompt)
	}
	//未传多轮上下文：不应出现上下文区块
	if strings.Contains(prompt, "【多轮上下文】") {
		t.Errorf("未传 context_json 不应出现多轮上下文区块: %s", prompt)
	}
}

// 测试多轮上下文注入：合法 JSON 的问题与 SQL 均进 prompt
func TestBuildAskUserPromptWithContext(t *testing.T) {
	prompt := buildAskUserPrompt("那按省份拆分呢", nil,
		`{"question":"总销售额是多少","sql":"SELECT SUM(amount) FROM dm_orders"}`)
	for _, want := range []string{
		"【多轮上下文】",
		"上一轮问题：总销售额是多少",
		"SELECT SUM(amount) FROM dm_orders",
	} {
		if !strings.Contains(prompt, want) {
			t.Errorf("多轮 prompt 缺少 %q: %s", want, prompt)
		}
	}
}

// 测试多轮上下文容错：空/非法 JSON/空对象均视为无上下文，不注入区块
func TestParseAskContextTolerant(t *testing.T) {
	for _, raw := range []string{"", "   ", "not-json", `{"other":1}`, "[1,2]"} {
		if ctx := parseAskContext(raw); ctx.Question != "" || ctx.Sql != "" {
			t.Errorf("parseAskContext(%q) = %+v, 期望零值", raw, ctx)
		}
	}
	//合法载荷正常解析
	ctx := parseAskContext(`{"question":"q1","sql":"select 1"}`)
	if ctx.Question != "q1" || ctx.Sql != "select 1" {
		t.Errorf("parseAskContext 合法载荷 = %+v, 解析不符", ctx)
	}
}

// 测试空候选用户消息：给出明确的空候选说明（引导模型按规则 6 拒答而非编造）
func TestBuildAskUserPromptEmptyCandidates(t *testing.T) {
	prompt := buildAskUserPrompt("有多少用户", nil, "")
	if !strings.Contains(prompt, "未命中任何候选") || !strings.Contains(prompt, "禁止编造表名") {
		t.Errorf("空候选 prompt 应含未命中说明与禁编造提示: %s", prompt)
	}
}

// 测试候选文本截断：超长 comment 被 clip 到上限内
func TestClipPromptText(t *testing.T) {
	long := strings.Repeat("列", maxCandidateTextRunes+50)
	clipped := clipPromptText(long)
	if got := len([]rune(strings.TrimSuffix(clipped, "…"))); got != maxCandidateTextRunes {
		t.Errorf("clipPromptText 截断后 %d 字符, 期望 %d", got, maxCandidateTextRunes)
	}
}

// ---------- SQL 提取 ----------

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
			name:    "纯文字拒答返回空串",
			content: "缺少候选元数据，无法确定对应的数据表，请先完善登记。",
			want:    "",
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
