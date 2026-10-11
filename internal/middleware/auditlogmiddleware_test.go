package middleware

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"datacenter/ops/rpc/opsClient"

	"google.golang.org/grpc"
)

// fakeOpsRpc 仅实现 WriteAuditLog 的 Ops 桩，记录审计入参供断言
type fakeOpsRpc struct {
	opsClient.Ops //其余方法不会被调用，内嵌接口以省略空实现
	mu            sync.Mutex
	got           []*opsClient.WriteAuditLogReq
	notify        chan struct{}
	fail          bool
}

func (f *fakeOpsRpc) WriteAuditLog(ctx context.Context, in *opsClient.WriteAuditLogReq, opts ...grpc.CallOption) (*opsClient.OkReply, error) {
	f.mu.Lock()
	f.got = append(f.got, in)
	f.mu.Unlock()
	if f.notify != nil {
		f.notify <- struct{}{}
	}
	if f.fail {
		return nil, io.ErrUnexpectedEOF
	}
	return &opsClient.OkReply{Ok: true}, nil
}

func (f *fakeOpsRpc) reqs() []*opsClient.WriteAuditLogReq {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.got
}

// waitCalled 等待异步审计落库完成（最多 2 秒）
func waitCalled(t *testing.T, f *fakeOpsRpc) {
	t.Helper()
	select {
	case <-f.notify:
	case <-time.After(2 * time.Second):
		t.Fatalf("等待审计落库超时")
	}
}

// TestAuditLogMiddlewareWriteOps POST/PUT/DELETE 应自动落审计，GET 不记
func TestAuditLogMiddlewareWriteOps(t *testing.T) {
	fake := &fakeOpsRpc{notify: make(chan struct{}, 8)}
	m := NewAuditLogMiddleware(fake)
	okHandler := func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) }

	// POST /datasource/create：带 JWT claims（uid/beid）
	r := httptest.NewRequest("POST", "/datasource/create?from=console", strings.NewReader(`{"name":"mysql","password":"p@ss"}`))
	r = r.WithContext(context.WithValue(r.Context(), "uid", json.Number("42")))
	r = r.WithContext(context.WithValue(r.Context(), "beid", json.Number("7")))
	w := httptest.NewRecorder()
	m.Handle(okHandler)(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("业务请求应正常放行, got %d", w.Code)
	}
	waitCalled(t, fake)

	got := fake.reqs()
	last := got[len(got)-1]
	if last.Action != "create" || last.Resource != "datasource" || last.ResourceId != "0" {
		t.Errorf("action/resource/resource_id 推断错误: %+v", last)
	}
	if last.Uid != 42 || last.Beid != 7 {
		t.Errorf("uid/beid 应取自 JWT ctx: %+v", last)
	}
	if last.Ip == "" {
		t.Errorf("ip 不应为空")
	}
	// detail_json 摘要应包含 path/query，且 password 已脱敏
	var detail map[string]string
	if err := json.Unmarshal([]byte(last.DetailJson), &detail); err != nil {
		t.Fatalf("detail_json 应为合法 JSON: %v", err)
	}
	if detail["path"] != "/datasource/create" || detail["query"] != "from=console" {
		t.Errorf("detail_json path/query 错误: %+v", detail)
	}
	if !strings.Contains(detail["body"], auditMaskValue) || strings.Contains(detail["body"], "p@ss") {
		t.Errorf("body 摘要中的密码应脱敏: %s", detail["body"])
	}

	// DELETE：action=delete
	r2 := httptest.NewRequest("DELETE", "/datasource/12", nil)
	w2 := httptest.NewRecorder()
	m.Handle(okHandler)(w2, r2)
	waitCalled(t, fake)
	last = fake.reqs()[len(fake.reqs())-1]
	if last.Action != "delete" || last.ResourceId != "12" {
		t.Errorf("DELETE 应记 delete 且 resource_id=12: %+v", last)
	}

	// PUT：action=update
	r3 := httptest.NewRequest("PUT", "/datasource/update", strings.NewReader(`{"id":1}`))
	w3 := httptest.NewRecorder()
	m.Handle(okHandler)(w3, r3)
	waitCalled(t, fake)
	if last := fake.reqs()[len(fake.reqs())-1]; last.Action != "update" {
		t.Errorf("PUT 应记 update: %+v", last)
	}

	// GET：不审计
	r4 := httptest.NewRequest("GET", "/datasource/list?page=1", nil)
	w4 := httptest.NewRecorder()
	m.Handle(okHandler)(w4, r4)
	select {
	case req := <-fake.notify:
		t.Errorf("GET 不应落审计日志: %+v", req)
	case <-time.After(100 * time.Millisecond):
	}
}

// TestAuditLogMiddlewareBodyPassthrough body 摘要截断不应破坏下游 handler 读取完整请求体
func TestAuditLogMiddlewareBodyPassthrough(t *testing.T) {
	fake := &fakeOpsRpc{notify: make(chan struct{}, 8)}
	m := NewAuditLogMiddleware(fake)

	// 构造超过 AuditBodyLimit 的请求体
	longBody := `{"data":"` + strings.Repeat("x", 2000) + `"}`
	var downstream string
	okHandler := func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		downstream = string(b)
	}
	r := httptest.NewRequest("POST", "/datasource/import", strings.NewReader(longBody))
	w := httptest.NewRecorder()
	m.Handle(okHandler)(w, r)
	waitCalled(t, fake)

	if downstream != longBody {
		t.Errorf("下游应读到完整请求体（截断长度 %d, 实际 %d）", len(longBody), len(downstream))
	}
	// body 摘要应为前 500 字节
	last := fake.reqs()[len(fake.reqs())-1]
	var detail map[string]string
	_ = json.Unmarshal([]byte(last.DetailJson), &detail)
	if len(detail["body"]) > AuditBodyLimit+64 { //脱敏重排后允许少量偏差，只验证不是全量
		t.Errorf("body 摘要应被限制在 %d 字节左右, got %d", AuditBodyLimit, len(detail["body"]))
	}
}

// TestAuditLogMiddlewareRPCFail RPC 失败不影响业务请求（仅记日志不阻断）
func TestAuditLogMiddlewareRPCFail(t *testing.T) {
	fake := &fakeOpsRpc{notify: make(chan struct{}, 8), fail: true}
	m := NewAuditLogMiddleware(fake)
	called := false
	okHandler := func(w http.ResponseWriter, r *http.Request) { called = true }
	r := httptest.NewRequest("POST", "/datasource/create", strings.NewReader(`{}`))
	w := httptest.NewRecorder()
	m.Handle(okHandler)(w, r)
	if !called || w.Code != http.StatusOK {
		t.Fatalf("RPC 失败不应影响业务放行")
	}
	waitCalled(t, fake) //确有调用尝试即可，错误已被中间件吞掉
}

// TestAuditHelpers resource/resource_id/action 推断的单测
func TestAuditHelpers(t *testing.T) {
	if auditResource("/datasource/create") != "datasource" {
		t.Errorf("/datasource/create 应取 datasource")
	}
	if auditResource("/") != "unknown" {
		t.Errorf("根路径应返回 unknown")
	}
	if auditResourceID("/datasource/12") != "12" || auditResourceID("/datasource/update") != "0" ||
		auditResourceID("/sys/user/create") != "0" {
		t.Errorf("resource_id 推断错误")
	}
	if auditAction(http.MethodPost) != "create" || auditAction(http.MethodPut) != "update" ||
		auditAction(http.MethodDelete) != "delete" {
		t.Errorf("action 推断错误")
	}
	if needAudit("GET") || needAudit("HEAD") || !needAudit("POST") || !needAudit("PUT") || !needAudit("DELETE") {
		t.Errorf("仅 POST/PUT/DELETE 需要审计")
	}
}
