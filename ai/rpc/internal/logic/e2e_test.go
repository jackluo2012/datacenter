package logic

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"datacenter/ai/rpc/ai"
	"datacenter/ai/rpc/internal/config"
	"datacenter/ai/rpc/internal/svc"
	"datacenter/ide/rpc/ideClient"
)

// e2e_test.go 端到端单测：httptest 模拟 OpenAI 兼容 /chat/completions，
// 全链路走 真实 llm.Client（HTTP 层）→ askCore（检索桩→LLM→SQL 提取→安全校验）→ AskAndRun（ide 执行桩）。
// 与 askAndRunLogic_test.go 的 RoundTrip 桩不同，这里连 HTTP 报文都真实收发，可校验请求体契约
// （system/user 双消息、model、Bearer 鉴权）。

// openAIMessage OpenAI 兼容请求体消息（测试侧解析用）
type openAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// capturedLLM 记录 httptest 服务端收到的请求要素
type capturedLLM struct {
	auth      string          //Authorization 头
	model     string          //请求体 model
	messages  []openAIMessage //请求体 messages
	recvCount int             //收到请求次数
}

// newOpenAIStub 启动 httptest 服务模拟 OpenAI 兼容 /chat/completions：
// 校验路径/方法/Content-Type，按入参 content 返回 choices[0].message.content；
// capture 非 nil 时记录请求要素供断言
func newOpenAIStub(t *testing.T, content string, capture *capturedLLM) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/completions" || r.Method != http.MethodPost {
			t.Errorf("LLM 请求路径/方法不符: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		data, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("读取请求体失败: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if capture != nil {
			capture.auth = r.Header.Get("Authorization")
			var req struct {
				Model    string          `json:"model"`
				Messages []openAIMessage `json:"messages"`
			}
			if err := json.Unmarshal(data, &req); err != nil {
				t.Errorf("请求体非合法 JSON: %v", err)
			}
			capture.model = req.Model
			capture.messages = req.Messages
			capture.recvCount++
		}
		//content 经 json.Marshal 转义（内容含反引号/换行不能直接拼字面量）
		contentJson, _ := json.Marshal(content)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"chatcmpl-test","object":"chat.completion","choices":[{"index":0,"message":{"role":"assistant","content":` +
			string(contentJson) + `},"finish_reason":"stop"}]}`))
	}))
	t.Cleanup(srv.Close)
	return srv
}

// newE2ESvcCtx 构造端到端 ServiceContext：LLM 指向 httptest 桩 + 元数据/指标/审计/ide 全桩
func newE2ESvcCtx(llmBaseURL string, ide ideClient.Ide) (*svc.ServiceContext, *fakeAskLogModel) {
	var cfg config.Config
	cfg.Llm.Enabled = true
	cfg.Llm.BaseUrl = llmBaseURL
	cfg.Llm.ApiKey = "sk-demo-key"
	cfg.Llm.Model = "deepseek-chat"
	audit := &fakeAskLogModel{}
	return &svc.ServiceContext{
		Config:      cfg,
		AskLogModel: audit,
		MetadataRpc: &fakeMetadata{},
		MetricRpc:   &fakeMetricSvc{},
		HttpClient:  http.DefaultClient, //真实 HTTP 客户端，走 httptest 服务
		IdeRpc:      ide,
	}, audit
}

// e2eSQLContent 模拟 LLM 按输出格式约定返回的回复（首行"根据<表>分析：" + ```sql 块）
const e2eSQLContent = "根据订单表分析：\n```sql\nSELECT SUM(amount) AS total_amount FROM dm_orders WHERE status <> 'cancelled' LIMIT 100\n```"

// 端到端（AskAndRun auto_run=true）：httptest OpenAI 桩 → 生成 SQL 提取 → ide 执行桩返回数据，
// 同时断言请求体契约（system/user 双消息、few-shot 注入、Bearer、model）
func TestE2EAskAndRunWithHttptestOpenAI(t *testing.T) {
	capture := &capturedLLM{}
	srv := newOpenAIStub(t, e2eSQLContent, capture)
	//ide 执行桩返回聚合结果形态（与生成 SQL 的 SUM 别名一致）
	fake := &fakeIde{runReply: &ideClient.RunScriptReply{
		ColumnsJson: `["total_amount"]`,
		RowsJson:    `[{"total_amount":123456.78}]`,
		RowCount:    1,
	}}
	svcCtx, audit := newE2ESvcCtx(srv.URL, fake)
	l := NewAskAndRunLogic(context.Background(), svcCtx)

	reply, err := l.AskAndRun(&ai.AskAndRunReq{
		Question:     "总销售额是多少",
		Beid:         7,
		AutoRun:      true,
		DatasourceId: 3,
	})
	if err != nil {
		t.Fatalf("AskAndRun err = %v, 期望成功", err)
	}

	//LLM 请求契约：Bearer 鉴权、model 名、system+user 双消息
	if capture.auth != "Bearer sk-demo-key" {
		t.Errorf("Authorization = %q, 期望 Bearer sk-demo-key", capture.auth)
	}
	if capture.model != "deepseek-chat" {
		t.Errorf("model = %q, 期望 deepseek-chat", capture.model)
	}
	if len(capture.messages) != 2 || capture.messages[0].Role != "system" || capture.messages[1].Role != "user" {
		t.Fatalf("messages = %+v, 期望 [system user] 双消息", capture.messages)
	}
	sysMsg := capture.messages[0].Content
	for _, want := range []string{"数据分析师", "禁止编造", "LIMIT 100", "dm_orders", "GROUP BY province"} {
		if !strings.Contains(sysMsg, want) {
			t.Errorf("system 消息缺少 %q", want)
		}
	}
	userMsg := capture.messages[1].Content
	//候选表 fqn/指标口径取自元数据/指标桩（mall.orders / sum(amount)）
	for _, want := range []string{"【候选表】", "mall.orders", "【候选指标】", "sum(amount)", "用户问题：总销售额是多少"} {
		if !strings.Contains(userMsg, want) {
			t.Errorf("user 消息缺少 %q", want)
		}
	}

	//生成侧：SQL 提取正确（含口径与 LIMIT）
	if !reply.LlmEnabled {
		t.Error("LlmEnabled=false, 期望 LLM 生成成功")
	}
	for _, want := range []string{"SUM(amount)", "status <> 'cancelled'", "LIMIT 100"} {
		if !strings.Contains(reply.GeneratedSql, want) {
			t.Errorf("GeneratedSql 缺少 %q: %s", want, reply.GeneratedSql)
		}
	}

	//执行侧：经 ide 桩真实执行并返回结果集，SQL 原样透传
	if !reply.Executed || reply.RunError != "" {
		t.Errorf("Executed=%v RunError=%q, 期望执行成功", reply.Executed, reply.RunError)
	}
	if fake.savedReq == nil || fake.savedReq.SqlText != reply.GeneratedSql {
		t.Errorf("ide 收到的 SQL = %+v, 应与生成 SQL 一致", fake.savedReq)
	}
	if reply.RowCount != 1 || !strings.Contains(reply.RowsJson, "total_amount") {
		t.Errorf("执行结果 = %d 行 %s, 与桩不符", reply.RowCount, reply.RowsJson)
	}

	//回答与审计：回答含 LLM 原文与执行摘要，审计落库含生成 SQL
	if !strings.Contains(reply.Answer, "根据订单表分析") || !strings.Contains(reply.Answer, "[自动执行]") {
		t.Errorf("answer 缺少分析首行/执行摘要: %s", reply.Answer)
	}
	if len(audit.logs) != 1 || audit.logs[0].LlmEnabled != 1 ||
		audit.logs[0].GeneratedSql != reply.GeneratedSql {
		t.Errorf("审计落库 = %+v, 期望 llm_enabled=1 且记录生成 SQL", audit.logs)
	}
}

// 端到端（Ask 多轮）：context_json 携带上一轮问题与 SQL，注入 user 消息辅助追问
func TestE2EAskMultiTurnContext(t *testing.T) {
	capture := &capturedLLM{}
	srv := newOpenAIStub(t, e2eSQLContent, capture)
	svcCtx, _ := newE2ESvcCtx(srv.URL, &fakeIde{})
	l := NewAskLogic(context.Background(), svcCtx)

	reply, err := l.Ask(&ai.AskReq{
		Question:      "那按省份拆分呢",
		Beid:          7,
		ContextJson:   `{"question":"总销售额是多少","sql":"SELECT SUM(amount) FROM dm_orders"}`,
		MaxCandidates: 5,
	})
	if err != nil {
		t.Fatalf("Ask err = %v, 期望成功", err)
	}
	if capture.recvCount != 1 {
		t.Fatalf("LLM 收到 %d 次请求, 期望 1 次", capture.recvCount)
	}
	userMsg := capture.messages[1].Content
	for _, want := range []string{
		"【多轮上下文】",
		"上一轮问题：总销售额是多少",
		"SELECT SUM(amount) FROM dm_orders",
		"用户问题：那按省份拆分呢",
	} {
		if !strings.Contains(userMsg, want) {
			t.Errorf("多轮 user 消息缺少 %q", want)
		}
	}
	if !reply.LlmEnabled || !strings.Contains(reply.GeneratedSql, "SUM(amount)") {
		t.Errorf("LlmEnabled=%v GeneratedSql=%q, 期望生成成功", reply.LlmEnabled, reply.GeneratedSql)
	}
}

// 端到端（安全拦截）：LLM 幻觉输出写语句时，发送侧校验拦截——generatedSql 清空、
// AskAndRun 不触达 ide、run_error 带拦截原因、answer 与审计均有记录
func TestE2EAskAndRunSqlGuardBlocksWriteSql(t *testing.T) {
	malicious := "根据订单表分析：\n```sql\nUPDATE dm_orders SET amount = 0\n```"
	capture := &capturedLLM{}
	srv := newOpenAIStub(t, malicious, capture)
	fake := &fakeIde{}
	svcCtx, audit := newE2ESvcCtx(srv.URL, fake)
	l := NewAskAndRunLogic(context.Background(), svcCtx)

	reply, err := l.AskAndRun(&ai.AskAndRunReq{
		Question:     "把订单金额清零",
		Beid:         7,
		AutoRun:      true,
		DatasourceId: 3,
	})
	if err != nil {
		t.Fatalf("安全拦截应降级不返回 error, 实际 err = %v", err)
	}
	//生成侧：llm_enabled=1（生成动作成功）但 SQL 被拦截清空
	if !reply.LlmEnabled {
		t.Error("LlmEnabled=false, 期望生成动作成功")
	}
	if reply.GeneratedSql != "" {
		t.Errorf("GeneratedSql = %q, 期望被安全校验清空", reply.GeneratedSql)
	}
	//执行侧：不触达 ide，run_error 带拦截原因
	if reply.Executed || fake.saveCalls != 0 || fake.runCalls != 0 {
		t.Errorf("Executed=%v ide 调用 save=%d run=%d, 期望被拦截不执行", reply.Executed, fake.saveCalls, fake.runCalls)
	}
	if !strings.Contains(reply.RunError, "安全校验") || !strings.Contains(reply.RunError, "SELECT") {
		t.Errorf("RunError = %q, 期望带安全校验拦截原因", reply.RunError)
	}
	if !strings.Contains(reply.Answer, "[安全校验未通过]") {
		t.Errorf("answer = %q, 期望带拦截说明", reply.Answer)
	}
	if len(audit.logs) != 1 || !strings.Contains(audit.logs[0].Answer, "[安全校验未通过]") {
		t.Errorf("审计落库 = %+v, 期望记录拦截说明", audit.logs)
	}
}

// 端到端（合规拒答）：候选不相关时 LLM 按规则只给文字说明、不输出 sql 块——
// 属正常业务态而非安全拦截：run_error 提示未生成 SQL，answer 无拦截标记
func TestE2EAskAndRunRefusalNoSqlBlock(t *testing.T) {
	refusal := "缺少候选元数据，无法确定对应的数据表，请先在元数据服务登记。"
	srv := newOpenAIStub(t, refusal, nil)
	fake := &fakeIde{}
	svcCtx, audit := newE2ESvcCtx(srv.URL, fake)
	l := NewAskAndRunLogic(context.Background(), svcCtx)

	reply, err := l.AskAndRun(&ai.AskAndRunReq{
		Question:     "公司今年营收多少", //与演示库无关的问法
		Beid:         7,
		AutoRun:      true,
		DatasourceId: 3,
	})
	if err != nil {
		t.Fatalf("合规拒答应降级不返回 error, 实际 err = %v", err)
	}
	if reply.LlmEnabled != true || reply.GeneratedSql != "" {
		t.Errorf("LlmEnabled=%v GeneratedSql=%q, 期望生成成功但无 SQL", reply.LlmEnabled, reply.GeneratedSql)
	}
	if reply.Executed || fake.saveCalls != 0 || fake.runCalls != 0 {
		t.Errorf("Executed=%v ide 调用 save=%d run=%d, 期望不执行", reply.Executed, fake.saveCalls, fake.runCalls)
	}
	if !strings.Contains(reply.RunError, "未生成 SQL") {
		t.Errorf("RunError = %q, 期望提示未生成 SQL", reply.RunError)
	}
	if strings.Contains(reply.Answer, "[安全校验未通过]") {
		t.Errorf("answer = %q, 合规拒答不应有安全拦截标记", reply.Answer)
	}
	if len(audit.logs) != 1 || !strings.Contains(audit.logs[0].Answer, "缺少候选元数据") {
		t.Errorf("审计落库 = %+v, 期望记录拒答原文", audit.logs)
	}
}

// 端到端（降级）：httptest 服务返回 500（如 key 无效/额度用尽），Ask 降级为纯检索不返回 error
func TestE2EAskLLMErrorDegrades(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":{"message":"Incorrect API key"}}`))
	}))
	t.Cleanup(srv.Close)
	svcCtx, audit := newE2ESvcCtx(srv.URL, &fakeIde{})
	l := NewAskLogic(context.Background(), svcCtx)

	reply, err := l.Ask(&ai.AskReq{Question: "总销售额是多少", Beid: 7})
	if err != nil {
		t.Fatalf("LLM 失败应降级不返回 error, 实际 err = %v", err)
	}
	if reply.LlmEnabled || reply.GeneratedSql != "" {
		t.Errorf("LlmEnabled=%v GeneratedSql=%q, 期望降级", reply.LlmEnabled, reply.GeneratedSql)
	}
	for _, want := range []string{"[LLM 调用失败已降级为纯检索]", "401"} {
		if !strings.Contains(reply.Answer, want) {
			t.Errorf("answer 缺少 %q: %s", want, reply.Answer)
		}
	}
	if len(audit.logs) != 1 || audit.logs[0].LlmEnabled != 0 {
		t.Errorf("审计落库 = %+v, 期望 llm_enabled=0", audit.logs)
	}
}
