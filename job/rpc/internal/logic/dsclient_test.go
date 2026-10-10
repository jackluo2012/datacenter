package logic

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
)

// recorded 单个请求的录制信息（请求构造断言用）
type recorded struct {
	Method string
	Path   string
	Query  url.Values
	Token  string
	Body   string
}

// recorder 捕获 DS 模拟服务收到的请求（handler 在服务 goroutine 执行，须加锁）
type recorder struct {
	mu sync.Mutex
	rs []recorded
}

// record 从请求中提取断言所需字段
func (r *recorder) record(req *http.Request) {
	var body string
	if req.Body != nil {
		b, _ := io.ReadAll(req.Body)
		body = string(b)
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.rs = append(r.rs, recorded{
		Method: req.Method,
		Path:   req.URL.Path,
		Query:  req.URL.Query(),
		Token:  req.Header.Get("token"),
		Body:   body,
	})
}

// last 取第 n（从 0 开始）个请求
func (r *recorder) last(n int) recorded {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.rs[n]
}

// count 已录制请求数
func (r *recorder) count() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.rs)
}

// dsJSON 模拟 DS 统一响应壳
func dsJSON(code int, msg string, data any) string {
	b, _ := json.Marshal(map[string]any{"code": code, "msg": msg, "data": data})
	return string(b)
}

// newDSTestEnv 启动模拟 DS 服务并返回注入了测试地址的客户端
func newDSTestEnv(t *testing.T, handler http.HandlerFunc) (*dsClient, *recorder) {
	t.Helper()
	rec := &recorder{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		rec.record(req)
		handler(w, req)
	}))
	t.Cleanup(srv.Close)
	client := &dsClient{
		base:    srv.URL,
		token:   "test-token",
		project: "888", // 纯数字：直接作为项目编码，不触发项目解析
		hc:      srv.Client(),
	}
	return client, rec
}

// okRoute 固定返回一个成功响应
func okRoute(payload string) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(payload))
	}
}

// mustNoErr 致命错误断言
func mustNoErr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestDSClientFromEnv 环境变量装配：未配置=未启用(nil)，配置=字段透传
func TestDSClientFromEnv(t *testing.T) {
	if c := dsClientFromEnv(func(string) string { return "" }, nil); c != nil {
		t.Fatalf("DS_BASE_URL 未配置时应返回 nil, got %+v", c)
	}
	env := map[string]string{
		"DS_BASE_URL": "http://ds:12345/dolphinscheduler/",
		"DS_TOKEN":    " abc ",
		"DS_PROJECT":  "中台同步",
	}
	c := dsClientFromEnv(func(k string) string { return env[k] }, nil)
	if c == nil {
		t.Fatal("已配置时应返回客户端")
	}
	if c.base != "http://ds:12345/dolphinscheduler" {
		t.Fatalf("base 应去尾部斜杠, got %q", c.base)
	}
	if c.token != "abc" || c.project != "中台同步" {
		t.Fatalf("token/project 未正确透传: %q %q", c.token, c.project)
	}
	if !c.enabled() {
		t.Fatal("enabled 应为 true")
	}
}

// TestDSPing ping 的请求构造（GET /projects + token 头）与成功解析
func TestDSPing(t *testing.T) {
	client, rec := newDSTestEnv(t, okRoute(dsJSON(0, "success", map[string]any{
		"totalList": []any{}, "total": 0,
	})))
	mustNoErr(t, client.ping())
	r := rec.last(0)
	if r.Method != http.MethodGet || r.Path != "/projects" {
		t.Fatalf("ping 应 GET /projects, got %s %s", r.Method, r.Path)
	}
	if r.Query.Get("pageNo") != "1" || r.Query.Get("pageSize") != "1" {
		t.Fatalf("ping 分页参数缺失: %v", r.Query)
	}
	if r.Token != "test-token" {
		t.Fatalf("token 请求头缺失: %q", r.Token)
	}
}

// TestDSPingTokenInvalid 鉴权失败（HTTP 200 但 DS 业务码非 0）须透出 DS msg
func TestDSPingTokenInvalid(t *testing.T) {
	client, _ := newDSTestEnv(t, okRoute(dsJSON(1001, "token 无效", nil)))
	err := client.ping()
	if err == nil || !strings.Contains(err.Error(), "token 无效") || !strings.Contains(err.Error(), "1001") {
		t.Fatalf("应透出 DS 业务失败信息, got %v", err)
	}
}

// TestDSNon200 HTTP 非 200 的错误透出
func TestDSNon200(t *testing.T) {
	client, _ := newDSTestEnv(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("boom"))
	})
	err := client.ping()
	if err == nil || !strings.Contains(err.Error(), "500") || !strings.Contains(err.Error(), "boom") {
		t.Fatalf("应透出 http 状态码与响应摘要, got %v", err)
	}
}

// TestDSUnreachable 不可达（连接拒绝）须返回错误
func TestDSUnreachable(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	srv.Close() // 立即关闭制造不可达
	client := &dsClient{base: srv.URL, hc: srv.Client()}
	if err := client.ping(); err == nil {
		t.Fatal("不可达时应返回错误")
	}
}

// TestDSResolveProjectNumeric 纯数字项目编码直接使用，不触发额外解析请求
func TestDSResolveProjectNumeric(t *testing.T) {
	client, rec := newDSTestEnv(t, okRoute(dsJSON(0, "success", nil)))
	code, err := client.resolveProject()
	mustNoErr(t, err)
	if code != "888" {
		t.Fatalf("纯数字应直接作为项目编码, got %q", code)
	}
	if rec.count() != 0 {
		t.Fatalf("纯数字项目编码不应发起请求, 请求数=%d", rec.count())
	}
}

// TestDSResolveProjectByName 项目名解析：按名精确匹配出编码
func TestDSResolveProjectByName(t *testing.T) {
	client, rec := newDSTestEnv(t, okRoute(dsJSON(0, "success", map[string]any{
		"totalList": []any{
			map[string]any{"code": 123, "name": "其它项目"},
			map[string]any{"code": 888, "name": "中台同步"},
		},
	})))
	client.project = "中台同步"
	code, err := client.resolveProject()
	mustNoErr(t, err)
	if code != "888" {
		t.Fatalf("应解析出项目编码 888, got %q", code)
	}
	r := rec.last(0)
	if r.Method != http.MethodGet || r.Path != "/projects" || r.Query.Get("searchVal") != "中台同步" {
		t.Fatalf("按名解析应 GET /projects 且带 searchVal 过滤, got %s %s %v", r.Method, r.Path, r.Query)
	}
}

// TestDSResolveProjectNotFound 项目名不存在须报错
func TestDSResolveProjectNotFound(t *testing.T) {
	client, _ := newDSTestEnv(t, okRoute(dsJSON(0, "success", map[string]any{
		"totalList": []any{}, "total": 0,
	})))
	client.project = "不存在"
	if err := func() error { _, err := client.resolveProject(); return err }(); err == nil || !strings.Contains(err.Error(), "不存在") {
		t.Fatalf("项目不存在应报错, got %v", err)
	}
}

// TestDSResolveProjectMissing 未配置项目须报业务错误
func TestDSResolveProjectMissing(t *testing.T) {
	client, _ := newDSTestEnv(t, okRoute(dsJSON(0, "success", nil)))
	client.project = ""
	_, err := client.resolveProject()
	if err == nil || !strings.Contains(err.Error(), "DS_PROJECT") {
		t.Fatalf("未配置项目应提示 DS_PROJECT, got %v", err)
	}
}

// TestDSEnsureWorkflowCreate 工作流不存在时创建：先查后建，请求构造与编码解析
func TestDSEnsureWorkflowCreate(t *testing.T) {
	client, rec := newDSTestEnv(t, func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case req.Method == http.MethodGet && strings.HasPrefix(req.URL.Path, "/projects/888/process-definition"):
			// 查询：无同名工作流
			_, _ = w.Write([]byte(dsJSON(0, "success", map[string]any{"totalList": []any{}, "total": 0})))
		case req.Method == http.MethodPost && req.URL.Path == "/projects/888/process-definition":
			// 创建：返回工作流编码（对象形态 data.code）
			_, _ = w.Write([]byte(dsJSON(0, "success", map[string]any{"code": 123, "name": "订单同步"})))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})
	code, err := client.ensureWorkflow("订单同步")
	mustNoErr(t, err)
	if code != 123 {
		t.Fatalf("应返回新工作流编码 123, got %d", code)
	}
	if rec.count() != 2 {
		t.Fatalf("应先查后建共 2 次请求, got %d", rec.count())
	}
	q := rec.last(0)
	if q.Method != http.MethodGet || q.Path != "/projects/888/process-definition" || q.Query.Get("searchVal") != "订单同步" {
		t.Fatalf("查询请求构造不符: %+v", q)
	}
	c := rec.last(1)
	if c.Method != http.MethodPost || c.Path != "/projects/888/process-definition" {
		t.Fatalf("创建请求构造不符: %+v", c)
	}
	var body map[string]any
	if err := json.Unmarshal([]byte(c.Body), &body); err != nil {
		t.Fatalf("创建请求体应为合法 JSON: %v", c.Body)
	}
	if body["name"] != "订单同步" {
		t.Fatalf("创建请求体 name 不符: %v", body["name"])
	}
	if tasks, ok := body["tasks"].([]any); !ok || len(tasks) != 0 {
		t.Fatalf("创建请求体应为空任务壳 tasks:[], got %v", body["tasks"])
	}
	if _, ok := body["location"]; !ok {
		t.Fatal("创建请求体缺少 location")
	}
}

// TestDSEnsureWorkflowExists 同名工作流已存在时直接复用（幂等，不发起创建）
func TestDSEnsureWorkflowExists(t *testing.T) {
	client, rec := newDSTestEnv(t, func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(dsJSON(0, "success", map[string]any{
			"totalList": []any{
				map[string]any{"code": 77, "name": "别的任务"},
				map[string]any{"code": 66, "name": "订单同步"},
			},
		})))
	})
	code, err := client.ensureWorkflow("订单同步")
	mustNoErr(t, err)
	if code != 66 {
		t.Fatalf("应复用已有工作流编码 66, got %d", code)
	}
	if rec.count() != 1 {
		t.Fatalf("已存在时不应发起创建请求, 请求数=%d", rec.count())
	}
}

// TestDSEnsureWorkflowCreateBareCode 创建响应 data 为裸数字形态时的编码解析
func TestDSEnsureWorkflowCreateBareCode(t *testing.T) {
	client, _ := newDSTestEnv(t, func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if req.Method == http.MethodGet {
			_, _ = w.Write([]byte(dsJSON(0, "success", map[string]any{"totalList": []any{}, "total": 0})))
			return
		}
		_, _ = w.Write([]byte(dsJSON(0, "success", 456)))
	})
	code, err := client.ensureWorkflow("t")
	mustNoErr(t, err)
	if code != 456 {
		t.Fatalf("裸数字形态应解析出 456, got %d", code)
	}
}

// TestDSEnsureWorkflowBizError 创建接口业务失败须透出 DS msg
func TestDSEnsureWorkflowBizError(t *testing.T) {
	client, _ := newDSTestEnv(t, func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if req.Method == http.MethodGet {
			_, _ = w.Write([]byte(dsJSON(0, "success", map[string]any{"totalList": []any{}, "total": 0})))
			return
		}
		_, _ = w.Write([]byte(dsJSON(10005, "工作流名称已存在", nil)))
	})
	_, err := client.ensureWorkflow("t")
	if err == nil || !strings.Contains(err.Error(), "工作流名称已存在") {
		t.Fatalf("应透出创建失败原因, got %v", err)
	}
}

// TestDSRelease 上线请求构造（POST .../release + releaseState=ONLINE）
func TestDSRelease(t *testing.T) {
	client, rec := newDSTestEnv(t, okRoute(dsJSON(0, "success", nil)))
	mustNoErr(t, client.release(123))
	r := rec.last(0)
	if r.Method != http.MethodPost || r.Path != "/projects/888/process-definition/123/release" {
		t.Fatalf("release 请求构造不符: %s %s", r.Method, r.Path)
	}
	if !strings.Contains(r.Body, `"releaseState":"ONLINE"`) {
		t.Fatalf("release 应提交 releaseState=ONLINE, got %s", r.Body)
	}
}

// TestDSReleaseBizError 上线失败透出原因
func TestDSReleaseBizError(t *testing.T) {
	client, _ := newDSTestEnv(t, okRoute(dsJSON(10017, "工作流未保存", nil)))
	err := client.release(123)
	if err == nil || !strings.Contains(err.Error(), "工作流未保存") {
		t.Fatalf("应透出上线失败原因, got %v", err)
	}
}

// TestDSStart 启动实例：请求构造与实例 id 解析（对象形态）
func TestDSStart(t *testing.T) {
	client, rec := newDSTestEnv(t, okRoute(dsJSON(0, "success", map[string]any{"id": 42})))
	id, err := client.start(123)
	mustNoErr(t, err)
	if id != 42 {
		t.Fatalf("应返回实例 id 42, got %d", id)
	}
	r := rec.last(0)
	if r.Method != http.MethodPost || r.Path != "/projects/888/executors/start-process-instance" {
		t.Fatalf("start 请求构造不符: %s %s", r.Method, r.Path)
	}
	var body map[string]any
	if err := json.Unmarshal([]byte(r.Body), &body); err != nil {
		t.Fatalf("start 请求体应为合法 JSON: %v", r.Body)
	}
	if body["processDefinitionCode"] != float64(123) {
		t.Fatalf("start 应提交 processDefinitionCode=123, got %v", body["processDefinitionCode"])
	}
}

// TestDSStartBareId 启动响应 data 为裸数字形态时的实例 id 解析
func TestDSStartBareId(t *testing.T) {
	client, _ := newDSTestEnv(t, okRoute(dsJSON(0, "success", 43)))
	id, err := client.start(123)
	mustNoErr(t, err)
	if id != 43 {
		t.Fatalf("裸数字形态应解析出 43, got %d", id)
	}
}

// TestDSStartMissingId 启动响应缺实例 id 须报错
func TestDSStartMissingId(t *testing.T) {
	client, _ := newDSTestEnv(t, okRoute(dsJSON(0, "success", map[string]any{})))
	if _, err := client.start(123); err == nil || !strings.Contains(err.Error(), "实例 id") {
		t.Fatalf("缺实例 id 应报错, got %v", err)
	}
}

// TestDSLatestInstance 实例状态查询：请求构造、状态与实例 id 解析
func TestDSLatestInstance(t *testing.T) {
	client, rec := newDSTestEnv(t, okRoute(dsJSON(0, "success", map[string]any{
		"totalList": []any{map[string]any{"id": 9, "state": "RUNNING_EXECUTION", "name": "订单同步-1-2024"}},
		"total":     1,
	})))
	state, id, err := client.latestInstance(123)
	mustNoErr(t, err)
	if state != "RUNNING_EXECUTION" || id != 9 {
		t.Fatalf("应解析出 RUNNING_EXECUTION/9, got %s/%d", state, id)
	}
	r := rec.last(0)
	if r.Method != http.MethodGet || r.Path != "/projects/888/process-instances" {
		t.Fatalf("实例查询请求构造不符: %s %s", r.Method, r.Path)
	}
	if r.Query.Get("processDefinitionCode") != "123" || r.Query.Get("pageSize") != "1" {
		t.Fatalf("实例查询过滤参数不符: %v", r.Query)
	}
	stateOnly, err := client.latestInstanceStatus(123)
	mustNoErr(t, err)
	if stateOnly != "RUNNING_EXECUTION" {
		t.Fatalf("latestInstanceStatus 应返回 %q, got %q", "RUNNING_EXECUTION", stateOnly)
	}
}

// TestDSLatestInstanceEmpty 无运行实例时返回空串（不视为错误）
func TestDSLatestInstanceEmpty(t *testing.T) {
	client, _ := newDSTestEnv(t, okRoute(dsJSON(0, "success", map[string]any{
		"totalList": []any{}, "total": 0,
	})))
	state, id, err := client.latestInstance(123)
	mustNoErr(t, err)
	if state != "" || id != 0 {
		t.Fatalf("无实例应返回空串/0, got %s/%d", state, id)
	}
}
