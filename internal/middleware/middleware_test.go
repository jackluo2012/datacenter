package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUserCheckMiddleware(t *testing.T) {
	m := NewUserCheckMiddleware()
	okHandler := func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) }

	// 一致：放行
	r := httptest.NewRequest("POST", "/votes/vote", nil)
	r = r.WithContext(context.WithValue(r.Context(), "uid", int64(42)))
	r.Header.Set("x-user-id", "42")
	w := httptest.NewRecorder()
	m.Handle(okHandler)(w, r)
	if w.Code != http.StatusOK {
		t.Errorf("uid 一致应放行, got %d", w.Code)
	}

	// 不一致：拒绝
	r2 := httptest.NewRequest("POST", "/votes/vote", nil)
	r2 = r2.WithContext(context.WithValue(r2.Context(), "uid", int64(42)))
	r2.Header.Set("x-user-id", "43")
	w2 := httptest.NewRecorder()
	m.Handle(okHandler)(w2, r2)
	if w2.Code == http.StatusOK {
		t.Errorf("uid 不一致应拒绝")
	}

	// 缺失 header：拒绝
	r3 := httptest.NewRequest("POST", "/votes/vote", nil)
	r3 = r3.WithContext(context.WithValue(r3.Context(), "uid", int64(42)))
	w3 := httptest.NewRecorder()
	m.Handle(okHandler)(w3, r3)
	if w3.Code == http.StatusOK {
		t.Errorf("缺失 x-user-id 应拒绝")
	}

	// 无 JWT claims（未带 token）：拒绝
	r4 := httptest.NewRequest("POST", "/votes/vote", nil)
	r4.Header.Set("x-user-id", "42")
	w4 := httptest.NewRecorder()
	m.Handle(okHandler)(w4, r4)
	if w4.Code == http.StatusOK {
		t.Errorf("无 claims 应拒绝")
	}
}

func TestAdminCheckMiddleware(t *testing.T) {
	t.Setenv("ADMIN_USER", "boss")
	t.Setenv("ADMIN_PASS", "secret123")
	m := NewAdminCheckMiddleware()
	okHandler := func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) }

	req := func(user, pass string) int {
		r := httptest.NewRequest("GET", "/search/article", nil)
		if user != "" || pass != "" {
			r.SetBasicAuth(user, pass)
		}
		w := httptest.NewRecorder()
		m.Handle(okHandler)(w, r)
		return w.Code
	}

	if code := req("boss", "secret123"); code != http.StatusOK {
		t.Errorf("正确凭据应放行, got %d", code)
	}
	if code := req("boss", "wrong"); code == http.StatusOK {
		t.Errorf("错误密码应拒绝")
	}
	if code := req("", ""); code == http.StatusOK {
		t.Errorf("无凭据应拒绝")
	}

	// 未配置环境变量：fail-closed 拒绝所有
	t.Setenv("ADMIN_USER", "")
	t.Setenv("ADMIN_PASS", "")
	m2 := NewAdminCheckMiddleware()
	r := httptest.NewRequest("GET", "/x", nil)
	r.SetBasicAuth("boss", "secret123")
	w := httptest.NewRecorder()
	m2.Handle(okHandler)(w, r)
	if w.Code == http.StatusOK {
		t.Errorf("未配置凭据应拒绝所有请求")
	}
}
