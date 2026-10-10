package shared

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestHttpPostForm(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Errorf("解析表单失败: %v", err)
		}
		if got := r.FormValue("name"); got != "jackluo" {
			t.Errorf("表单字段 name = %q, 期望 %q", got, "jackluo")
		}
		w.Write([]byte("ok"))
	}))
	defer srv.Close()

	body, err := HttpPostForm(srv.URL, url.Values{"name": {"jackluo"}})
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != "ok" {
		t.Errorf("响应体 = %q, 期望 %q", string(body), "ok")
	}
}

func TestToLimitOffset(t *testing.T) {
	offset, limit := ToLimitOffset("3", "20")
	if offset != 40 || limit != 20 {
		t.Errorf("ToLimitOffset(3,20) = (%d, %d), 期望 (40, 20)", offset, limit)
	}
	//默认值：page<=0 归 1，limit=0 归 10
	offset, limit = ToLimitOffset("0", "")
	if offset != 0 || limit != 10 {
		t.Errorf("ToLimitOffset(0,\"\") = (%d, %d), 期望 (0, 10)", offset, limit)
	}
}
