package llm

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// 测试正常请求/响应：URL、方法、Bearer 鉴权头、请求体、choices[0].message.content 解析
func TestChatCompletionOK(t *testing.T) {
	var gotAuth, gotPath, gotMethod, gotContentType string
	var gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotAuth = r.Header.Get("Authorization")
		gotContentType = r.Header.Get("Content-Type")
		data, _ := io.ReadAll(r.Body)
		gotBody = string(data)
		//注意：JSON 里的 ```sql 围栏含反引号，不能放进 Go 原始字符串字面量
		w.Write([]byte("{\"choices\":[{\"message\":{\"role\":\"assistant\",\"content\":\"```sql\\nselect 1\\n```\"}}]}"))
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "sk-test", "gpt-4o-mini", srv.Client())
	content, err := c.ChatCompletion(context.Background(), "上月gmv是多少")
	if err != nil {
		t.Fatalf("ChatCompletion 报错: %v", err)
	}
	if content != "```sql\nselect 1\n```" {
		t.Errorf("content = %q, 期望 %q", content, "```sql\nselect 1\n```")
	}
	if gotPath != "/chat/completions" {
		t.Errorf("请求路径 = %q, 期望 /chat/completions", gotPath)
	}
	if gotMethod != http.MethodPost {
		t.Errorf("请求方法 = %q, 期望 POST", gotMethod)
	}
	if gotAuth != "Bearer sk-test" {
		t.Errorf("Authorization = %q, 期望 %q", gotAuth, "Bearer sk-test")
	}
	if !strings.HasPrefix(gotContentType, "application/json") {
		t.Errorf("Content-Type = %q, 期望 application/json", gotContentType)
	}
	//请求体：model / temperature=0 / prompt
	for _, want := range []string{`"model":"gpt-4o-mini"`, `"temperature":0`, `"role":"user"`, `"上月gmv是多少"`} {
		if !strings.Contains(gotBody, want) {
			t.Errorf("请求体缺少 %s: %s", want, gotBody)
		}
	}
}

// 测试非 200 响应：返回错误且带状态码
func TestChatCompletionNon200(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("upstream error"))
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "sk-test", "gpt-4o-mini", srv.Client())
	_, err := c.ChatCompletion(context.Background(), "问题")
	if err == nil {
		t.Fatal("期望非 200 返回错误，实际为 nil")
	}
	if !strings.Contains(err.Error(), "非 200") || !strings.Contains(err.Error(), "500") {
		t.Errorf("错误信息应包含状态码: %v", err)
	}
}

// 测试超时：注入短超时 http.Client，服务端 sleep 超过客户端超时
func TestChatCompletionTimeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(300 * time.Millisecond)
		w.Write([]byte(`{"choices":[{"message":{"content":"ok"}}]}`))
	}))
	defer srv.Close()

	slowClient := &http.Client{Timeout: 50 * time.Millisecond}
	c := NewClient(srv.URL, "sk-test", "gpt-4o-mini", slowClient)
	_, err := c.ChatCompletion(context.Background(), "问题")
	if err == nil {
		t.Fatal("期望超时返回错误，实际为 nil")
	}
	if !strings.Contains(err.Error(), "LLM 请求失败") {
		t.Errorf("错误信息应含降级包装前缀: %v", err)
	}
}

// 测试非法响应体：choices 为空返回错误
func TestChatCompletionNoChoices(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"choices":[]}`))
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "sk-test", "gpt-4o-mini", srv.Client())
	_, err := c.ChatCompletion(context.Background(), "问题")
	if err == nil || !strings.Contains(err.Error(), "choices") {
		t.Errorf("期望空 choices 报错，实际: %v", err)
	}
}

// 测试 baseUrl 尾斜杠清理
func TestNewClientTrimBaseUrl(t *testing.T) {
	c := NewClient("https://api.example.com/v1/", "k", "m", http.DefaultClient)
	if c.baseUrl != "https://api.example.com/v1" {
		t.Errorf("baseUrl = %q, 期望尾斜杠被清理", c.baseUrl)
	}
	if c.httpClient != http.DefaultClient {
		t.Errorf("httpClient 应保留注入实例")
	}
}
