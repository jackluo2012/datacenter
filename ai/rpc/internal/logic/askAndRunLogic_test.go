package logic

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"datacenter/ai/model"
	"datacenter/ai/rpc/ai"
	"datacenter/ai/rpc/internal/config"
	"datacenter/ai/rpc/internal/svc"
	"datacenter/ide/rpc/ideClient"
	"datacenter/metadata/rpc/metadataClient"
	"datacenter/metric/rpc/metricsvc"

	"google.golang.org/grpc"
)

// ---------- 测试桩 ----------

// fakeIde ideClient.Ide 测试桩：内嵌接口兜底未触达方法，SaveScript/RunScript 按预置返回并记录入参
type fakeIde struct {
	ideClient.Ide

	saveErr  error                     //SaveScript 预置错误（nil=成功）
	runErr   error                     //RunScript 预置错误（nil=成功）
	runReply *ideClient.RunScriptReply //RunScript 预置成功响应（nil=默认一行结果）

	savedReq  *ideClient.SaveScriptReq //记录 SaveScript 入参
	runReq    *ideClient.RunScriptReq  //记录 RunScript 入参
	saveCalls int
	runCalls  int
}

func (f *fakeIde) SaveScript(ctx context.Context, in *ideClient.SaveScriptReq, opts ...grpc.CallOption) (*ideClient.ScriptReply, error) {
	f.saveCalls++
	f.savedReq = in
	if f.saveErr != nil {
		return nil, f.saveErr
	}
	return &ideClient.ScriptReply{Script: &ideClient.Script{Id: 101, Name: in.Name}}, nil
}

func (f *fakeIde) RunScript(ctx context.Context, in *ideClient.RunScriptReq, opts ...grpc.CallOption) (*ideClient.RunScriptReply, error) {
	f.runCalls++
	f.runReq = in
	if f.runErr != nil {
		return nil, f.runErr
	}
	if f.runReply != nil {
		return f.runReply, nil
	}
	return &ideClient.RunScriptReply{ColumnsJson: `["cnt"]`, RowsJson: `[{"cnt":42}]`, RowCount: 1}, nil
}

// fakeMetadata 元数据 rpc 测试桩：ListEntities 固定返回一张订单表实体
type fakeMetadata struct {
	metadataClient.Metadata
}

func (f *fakeMetadata) ListEntities(ctx context.Context, in *metadataClient.EntityListReq, opts ...grpc.CallOption) (*metadataClient.EntityListReply, error) {
	return &metadataClient.EntityListReply{List: []*metadataClient.MetaEntity{
		{Id: 1, EntityType: "table", Name: "订单表", Fqn: "mysql://mall.orders", Comment: "订单主表"},
	}}, nil
}

// fakeMetricSvc 指标 rpc 测试桩：ListMetrics 固定返回一个 GMV 指标
type fakeMetricSvc struct {
	metricsvc.MetricSvc
}

func (f *fakeMetricSvc) ListMetrics(ctx context.Context, in *metricsvc.MetricListReq, opts ...grpc.CallOption) (*metricsvc.MetricListReply, error) {
	return &metricsvc.MetricListReply{List: []*metricsvc.Metric{
		{Id: 2, NameEn: "gmv", NameCn: "成交额", AggType: "sum", BizField: "amount", Comment: "成交金额口径"},
	}}, nil
}

// fakeAskLogModel 审计 model 测试桩：记录落库的审计行
type fakeAskLogModel struct {
	model.AskLogModel
	logs []*model.AskLog
}

func (f *fakeAskLogModel) Insert(ctx context.Context, data *model.AskLog) (int64, error) {
	f.logs = append(f.logs, data)
	return int64(len(f.logs)), nil
}

// stubLLMTransport LLM HTTP 桩：任意请求返回固定 content 的 OpenAI 兼容响应
type stubLLMTransport struct {
	content string
}

func (s *stubLLMTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	raw, err := json.Marshal(s.content)
	if err != nil {
		return nil, err
	}
	body := fmt.Sprintf(`{"choices":[{"message":{"content":%s}}]}`, raw)
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     http.Header{},
	}, nil
}

const llmSqlContent = "按订单表统计如下：\n```sql\nSELECT count(*) AS cnt FROM mall.orders\n```"

// newAskAndRunSvcCtx 构造带全套桩的 ServiceContext（LLM 可关，ide 桩由调用方注入）
func newAskAndRunSvcCtx(llmEnabled bool, ide ideClient.Ide) (*svc.ServiceContext, *fakeAskLogModel) {
	var cfg config.Config
	cfg.Llm.Enabled = llmEnabled
	cfg.Llm.BaseUrl = "http://llm.fake"
	cfg.Llm.ApiKey = "test-key"
	cfg.Llm.Model = "gpt-4o-mini"
	audit := &fakeAskLogModel{}
	return &svc.ServiceContext{
		Config:      cfg,
		AskLogModel: audit,
		MetadataRpc: &fakeMetadata{},
		MetricRpc:   &fakeMetricSvc{},
		HttpClient:  &http.Client{Transport: &stubLLMTransport{content: llmSqlContent}},
		IdeRpc:      ide,
	}, audit
}

// ---------- 入参校验 ----------

// 测试入参校验：question 为空、auto_run 缺 datasource_id/beid 均拒绝；
// 校验先于任何下游依赖，svcCtx 空结构体即可
func TestAskAndRunValidation(t *testing.T) {
	l := NewAskAndRunLogic(context.Background(), &svc.ServiceContext{})
	cases := []struct {
		name    string
		in      *ai.AskAndRunReq
		wantErr string
	}{
		{
			name:    "question 为空",
			in:      &ai.AskAndRunReq{Question: "   ", AutoRun: true, DatasourceId: 1, Beid: 1},
			wantErr: "question",
		},
		{
			name:    "auto_run 缺 datasource_id",
			in:      &ai.AskAndRunReq{Question: "订单表有多少数据", AutoRun: true, DatasourceId: 0, Beid: 1},
			wantErr: "datasource_id",
		},
		{
			name:    "auto_run 缺 beid",
			in:      &ai.AskAndRunReq{Question: "订单表有多少数据", AutoRun: true, DatasourceId: 3, Beid: 0},
			wantErr: "beid",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := l.AskAndRun(c.in)
			if err == nil || !strings.Contains(err.Error(), c.wantErr) {
				t.Fatalf("AskAndRun(%+v) err = %v, 期望包含 %q", c.in, err, c.wantErr)
			}
		})
	}
}

// ---------- 纯函数 ----------

// 测试执行行数归一化：<=0 取默认 20，超上限截断
func TestNormalizeRunRows(t *testing.T) {
	cases := []struct {
		in   int32
		want int32
	}{
		{in: 0, want: 20},
		{in: -5, want: 20},
		{in: 7, want: 7},
		{in: 100000, want: 100},
	}
	for _, c := range cases {
		if got := normalizeRunRows(c.in); got != c.want {
			t.Errorf("normalizeRunRows(%d) = %d, 期望 %d", c.in, got, c.want)
		}
	}
}

// 测试执行摘要追加：成功/失败/未开启三种形态
func TestAppendRunSummary(t *testing.T) {
	//成功：摘要含数据源与行数
	got := appendRunSummary("答案", true, "", 3, 20)
	if !strings.Contains(got, "答案") || !strings.Contains(got, "[自动执行]") ||
		!strings.Contains(got, "数据源 3") || !strings.Contains(got, "20 行") {
		t.Errorf("appendRunSummary 成功态 = %q, 缺少预期片段", got)
	}
	//失败：摘要带原因
	got = appendRunSummary("答案", false, "SQL 不合法", 3, 0)
	if !strings.Contains(got, "[自动执行未成功]") || !strings.Contains(got, "SQL 不合法") {
		t.Errorf("appendRunSummary 失败态 = %q, 缺少预期片段", got)
	}
	//未开启自动执行：原样返回
	got = appendRunSummary("答案", false, "", 0, 0)
	if got != "答案" {
		t.Errorf("appendRunSummary 未执行态 = %q, 期望原样返回", got)
	}
}

// ---------- ide 执行链路 ----------

// 测试执行链路成功：先存脚本（带 datasource_id/beid、名称前缀 ai-gen-）再执行（带脚本 id 与行数上限）
func TestRunGeneratedSqlViaIdeSuccess(t *testing.T) {
	fake := &fakeIde{}
	run, err := runGeneratedSqlViaIde(context.Background(), fake,
		"SELECT count(*) FROM mall.orders", 7, 3, 25)
	if err != nil {
		t.Fatalf("runGeneratedSqlViaIde err = %v, 期望成功", err)
	}
	if run == nil || run.RowCount != 1 {
		t.Fatalf("run = %+v, 期望返回结果集", run)
	}
	//存脚本入参：临时脚本名前缀 + 调用方指定的数据源
	if fake.savedReq == nil || !strings.HasPrefix(fake.savedReq.Name, "ai-gen-") {
		t.Errorf("SaveScript 入参 = %+v, 名称应带 ai-gen- 前缀", fake.savedReq)
	}
	if fake.savedReq.SqlText != "SELECT count(*) FROM mall.orders" || fake.savedReq.DatasourceId != 3 || fake.savedReq.Beid != 7 {
		t.Errorf("SaveScript 入参 = %+v, SQL/数据源/租户不符", fake.savedReq)
	}
	//执行入参：脚本 id 取自保存回包，行数上限原样透传
	if fake.runReq == nil || fake.runReq.Id != 101 || fake.runReq.Beid != 7 || fake.runReq.MaxRows != 25 {
		t.Errorf("RunScript 入参 = %+v, 脚本 id/租户/行数不符", fake.runReq)
	}
}

// 测试执行链路降级：存脚本失败与执行失败均带出原因
func TestRunGeneratedSqlViaIdeDegraded(t *testing.T) {
	//存脚本失败
	fake := &fakeIde{saveErr: errors.New("etcd 无 ide 节点")}
	if _, err := runGeneratedSqlViaIde(context.Background(), fake, "select 1", 1, 2, 20); err == nil ||
		!strings.Contains(err.Error(), "保存 AI 生成脚本失败") {
		t.Fatalf("存脚本失败 err = %v, 期望带保存失败原因", err)
	}
	//执行失败（如安全闸拒绝非 SELECT）
	fake = &fakeIde{runErr: errors.New("SQL 不合法: 仅放行单条 SELECT")}
	_, err := runGeneratedSqlViaIde(context.Background(), fake, "delete from t", 1, 2, 20)
	if err == nil || !strings.Contains(err.Error(), "ide 执行失败") || !strings.Contains(err.Error(), "SELECT") {
		t.Fatalf("执行失败 err = %v, 期望带执行失败原因", err)
	}
	//未装配客户端
	if _, err := runGeneratedSqlViaIde(context.Background(), nil, "select 1", 1, 2, 20); err == nil {
		t.Fatal("nil 客户端应返回错误")
	}
}

// ---------- AskAndRun 全流程 ----------

// 全链路（LLM/元数据/指标/ide/审计全桩）：auto_run=true 时生成 SQL 真实执行并返回数据
func TestAskAndRunFullFlow(t *testing.T) {
	fake := &fakeIde{}
	svcCtx, audit := newAskAndRunSvcCtx(true, fake)
	l := NewAskAndRunLogic(context.Background(), svcCtx)

	reply, err := l.AskAndRun(&ai.AskAndRunReq{
		Question:      "订单表有多少数据",
		Beid:          7,
		AutoRun:       true,
		MaxRows:       0, //缺省走默认 20
		DatasourceId:  3,
		MaxCandidates: 5,
	})
	if err != nil {
		t.Fatalf("AskAndRun err = %v, 期望成功", err)
	}
	//生成侧：LLM 命中并提取到 SQL
	if !reply.LlmEnabled || !strings.Contains(reply.GeneratedSql, "SELECT count(*)") {
		t.Errorf("LlmEnabled=%v GeneratedSql=%q, 期望生成 SELECT", reply.LlmEnabled, reply.GeneratedSql)
	}
	if len(reply.Candidates) == 0 {
		t.Error("候选为空, 期望元数据/指标桩命中")
	}
	//执行侧：executed=true，结果集透传，行数上限取默认 20
	if !reply.Executed || reply.RunError != "" {
		t.Errorf("Executed=%v RunError=%q, 期望执行成功", reply.Executed, reply.RunError)
	}
	if reply.ColumnsJson != `["cnt"]` || reply.RowsJson != `[{"cnt":42}]` || reply.RowCount != 1 {
		t.Errorf("执行结果 = %s/%s/%d, 与 ide 桩不符", reply.ColumnsJson, reply.RowsJson, reply.RowCount)
	}
	if fake.runReq == nil || fake.runReq.MaxRows != 20 {
		t.Errorf("RunScript MaxRows = %+v, 期望缺省 20", fake.runReq)
	}
	//回答尾部带执行摘要；审计落库同文（不改表，执行信息随 answer 记录）
	if !strings.Contains(reply.Answer, "[自动执行]") || !strings.Contains(reply.Answer, "数据源 3") {
		t.Errorf("answer = %q, 缺少执行摘要", reply.Answer)
	}
	if len(audit.logs) != 1 || !strings.Contains(audit.logs[0].Answer, "[自动执行]") ||
		audit.logs[0].GeneratedSql != reply.GeneratedSql || audit.logs[0].Beid != 7 {
		t.Errorf("审计落库 = %+v, 期望含执行摘要且与回复一致", audit.logs)
	}
}

// 降级路径：ide 执行失败不返回 error，run_error 带原因，回答附未成功摘要
func TestAskAndRunIdeErrorDegraded(t *testing.T) {
	fake := &fakeIde{runErr: errors.New("SQL 不合法: 仅放行单条 SELECT")}
	svcCtx, audit := newAskAndRunSvcCtx(true, fake)
	l := NewAskAndRunLogic(context.Background(), svcCtx)

	reply, err := l.AskAndRun(&ai.AskAndRunReq{
		Question:     "订单表有多少数据",
		Beid:         7,
		AutoRun:      true,
		DatasourceId: 3,
	})
	if err != nil {
		t.Fatalf("执行失败应降级不返回 error, 实际 err = %v", err)
	}
	if reply.Executed {
		t.Error("Executed=true, 期望未执行")
	}
	if !strings.Contains(reply.RunError, "仅放行单条 SELECT") {
		t.Errorf("RunError = %q, 期望带安全闸拒绝原因", reply.RunError)
	}
	if !strings.Contains(reply.Answer, "[自动执行未成功]") {
		t.Errorf("answer = %q, 缺少未成功摘要", reply.Answer)
	}
	if len(audit.logs) != 1 || !strings.Contains(audit.logs[0].Answer, "[自动执行未成功]") {
		t.Errorf("审计落库 = %+v, 期望记录未成功摘要", audit.logs)
	}
}

// 未启用 LLM 时 auto_run 不触发执行，run_error 说明原因
func TestAskAndRunNoLlmNoRun(t *testing.T) {
	fake := &fakeIde{}
	svcCtx, audit := newAskAndRunSvcCtx(false, fake) //Llm.Enabled=false 零值配置
	l := NewAskAndRunLogic(context.Background(), svcCtx)

	reply, err := l.AskAndRun(&ai.AskAndRunReq{
		Question:     "订单表有多少数据",
		Beid:         7,
		AutoRun:      true,
		DatasourceId: 3,
	})
	if err != nil {
		t.Fatalf("AskAndRun err = %v, 期望成功", err)
	}
	if reply.LlmEnabled || reply.Executed {
		t.Errorf("LlmEnabled=%v Executed=%v, 期望均 false", reply.LlmEnabled, reply.Executed)
	}
	if !strings.Contains(reply.RunError, "LLM") {
		t.Errorf("RunError = %q, 期望说明无生成 SQL 可执行", reply.RunError)
	}
	if fake.saveCalls != 0 || fake.runCalls != 0 {
		t.Errorf("ide 调用 save=%d run=%d, 期望均未触达", fake.saveCalls, fake.runCalls)
	}
	if !strings.Contains(reply.Answer, "[自动执行未成功]") {
		t.Errorf("answer = %q, 缺少未成功摘要", reply.Answer)
	}
	if len(audit.logs) != 1 {
		t.Fatalf("审计落库 %d 条, 期望 1 条", len(audit.logs))
	}
}

// auto_run=false：纯问数，不触达 ide，answer 不追加执行摘要
func TestAskAndRunWithoutAutoRun(t *testing.T) {
	fake := &fakeIde{}
	svcCtx, audit := newAskAndRunSvcCtx(true, fake)
	l := NewAskAndRunLogic(context.Background(), svcCtx)

	reply, err := l.AskAndRun(&ai.AskAndRunReq{Question: "订单表有多少数据", Beid: 7})
	if err != nil {
		t.Fatalf("AskAndRun err = %v, 期望成功", err)
	}
	if reply.Executed || reply.RunError != "" {
		t.Errorf("Executed=%v RunError=%q, 期望未开启执行", reply.Executed, reply.RunError)
	}
	if strings.Contains(reply.Answer, "[自动执行") {
		t.Errorf("answer = %q, 不应带执行摘要", reply.Answer)
	}
	if fake.saveCalls != 0 || fake.runCalls != 0 {
		t.Errorf("ide 调用 save=%d run=%d, 期望均未触达", fake.saveCalls, fake.runCalls)
	}
	if len(audit.logs) != 1 || strings.Contains(audit.logs[0].Answer, "[自动执行") {
		t.Errorf("审计落库 = %+v, 期望无执行摘要", audit.logs)
	}
}
