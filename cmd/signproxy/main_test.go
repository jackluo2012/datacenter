package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

// refSign 参考实现 —— 与 datasvc/rpc/internal/logic/credential_test.go 的 mustSign
// 完全相同的算法代码（该 logic 包是 internal 包无法 import，故此处复制对拍）。
func refSign(appKey, ts, appSecret string) string {
	mac := hmac.New(sha256.New, []byte(appSecret))
	mac.Write([]byte(appKey + ts))
	return hex.EncodeToString(mac.Sum(nil))
}

// TestGenSignFixedVectors 固定向量对拍：期望值均由
// `printf '%s' "<appKey><ts>" | openssl dgst -sha256 -hmac "<secret>"` 现算得出，
// 与 datasvc credential.go genSign（及 credential_test.go mustSign）同源算法必须逐字一致。
func TestGenSignFixedVectors(t *testing.T) {
	cases := []struct {
		appKey, ts, secret, want string
	}{
		{
			appKey: "app_0123456789abcdef", ts: "1700000000",
			secret: "0123456789abcdef0123456789abcdef",
			want:   "026f50db22bd10779ad922dfa4d23be1b92c1adf96f82cd029fe188e3c69f972",
		},
		{
			appKey: "demo-key", ts: "1759804800",
			secret: "demo-secret",
			want:   "1d768e07ea4d92652494b89e99185619d1fe8c390dd4ad40aedbe6f2e831a103",
		},
	}
	for _, c := range cases {
		if got := genSign(c.appKey, c.ts, c.secret); got != c.want {
			t.Errorf("genSign 固定向量不一致: got=%s want=%s", got, c.want)
		}
		// 与参考实现（datasvc 同款算法代码）对拍
		if got := refSign(c.appKey, c.ts, c.secret); got != genSign(c.appKey, c.ts, c.secret) {
			t.Errorf("genSign 与 datasvc 参考实现不一致: got=%s ref=%s",
				genSign(c.appKey, c.ts, c.secret), refSign(c.appKey, c.ts, c.secret))
		}
	}
}

// newTestProxy 起「代理 → httptest 上游」双跳测试链路，上游行为由 upHandler 定义，
// 返回代理地址与捕获器（header/body）。
func newTestProxy(t *testing.T, upHandler http.HandlerFunc) (proxyURL string, got *capture) {
	t.Helper()
	got = &capture{}
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 仅接受打向上游统一入口的 POST
		if r.URL.Path != upstreamPath {
			t.Errorf("上游收到非预期路径: %s", r.URL.Path)
		}
		got.header = r.Header.Clone()
		b, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("读上游 body 失败: %v", err)
		}
		got.body = b
		upHandler(w, r)
	}))
	t.Cleanup(upstream.Close)

	proxy := httptest.NewServer(newProxyHandler("app_e2e", "s3cr3t", upstream.URL))
	t.Cleanup(proxy.Close)
	return proxy.URL, got
}

// capture 上游捕获：请求 header 与 body
type capture struct {
	header http.Header
	body   []byte
}

// execBody 代理侧收到并注入后的 body 结构（网关 ExecuteApiReq 同构）
type execBody struct {
	AppKey     string `json:"app_key"`
	Sign       string `json:"sign"`
	Ts         string `json:"ts"`
	Path       string `json:"path"`
	ParamsJson string `json:"params_json"`
}

// TestProxyForwardSignatureAndBody 端到端断言：三 header 正确 / body 透传且签名注入 / 上游响应原样返回。
// 上游模拟 denied 场景（403 + 错误体），验证透传链路完整。
func TestProxyForwardSignatureAndBody(t *testing.T) {
	logx.SetLevel(logx.ErrorLevel)
	proxyURL, got := newTestProxy(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		_, _ = io.WriteString(w, `{"error":"denied"}`)
	})

	// DataEase 侧发出的静态 body（不含签名字段，header 留空）
	req, err := http.NewRequest(http.MethodPost, proxyURL+upstreamPath,
		strings.NewReader(`{"path":"/api/v1/demo/orders","params_json":"{\"limit\":10}"}`))
	if err != nil {
		t.Fatalf("构造请求失败: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("代理请求失败: %v", err)
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)

	// 上游响应原样返回：状态码 + body + Content-Type
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("上游 403 应原样透传, got: %d", resp.StatusCode)
	}
	if string(respBody) != `{"error":"denied"}` {
		t.Errorf("上游 body 应原样透传, got: %s", respBody)
	}
	if ct := resp.Header.Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type 应透传, got: %q", ct)
	}

	// header 三件套：X-App-Key / X-Timestamp / X-Sign
	tsv, err := strconv.ParseInt(got.header.Get("X-Timestamp"), 10, 64)
	if err != nil {
		t.Fatalf("X-Timestamp 应为秒级整数: %v", err)
	}
	if diff := time.Now().Unix() - tsv; diff < -5 || diff > 5 {
		t.Errorf("X-Timestamp 应为当前时间（±5s 内）, got: %d", tsv)
	}
	if got.header.Get("X-App-Key") != "app_e2e" {
		t.Errorf("X-App-Key 应为配置值, got: %q", got.header.Get("X-App-Key"))
	}
	// X-Sign 与 datasvc genSign 同款算法现算对拍（用捕获的 ts 复算）
	wantSign := refSign("app_e2e", got.header.Get("X-Timestamp"), "s3cr3t")
	if got.header.Get("X-Sign") != wantSign {
		t.Errorf("X-Sign 与 datasvc genSign 算法不一致: got=%s want=%s",
			got.header.Get("X-Sign"), wantSign)
	}

	// body 透传且签名三字段已注入：业务字段原样、签名与 header 一致
	eb := gotBody(t, got)
	if eb.Path != "/api/v1/demo/orders" || eb.ParamsJson != `{"limit":10}` {
		t.Errorf("业务字段应原样透传, got: %s", got.body)
	}
	if eb.AppKey != "app_e2e" {
		t.Errorf("body app_key 应注入, got: %q", eb.AppKey)
	}
	if eb.Ts != got.header.Get("X-Timestamp") {
		t.Errorf("body ts 应与 X-Timestamp 一致, got: %q", eb.Ts)
	}
	if eb.Sign != got.header.Get("X-Sign") {
		t.Errorf("body sign 应与 X-Sign 一致, got: %q", eb.Sign)
	}
}

// gotBody 便捷取捕获 body（解析失败直接退测试）
func gotBody(t *testing.T, c *capture) execBody {
	t.Helper()
	var eb execBody
	if err := json.Unmarshal(c.body, &eb); err != nil {
		t.Fatalf("上游 body 应为合法 JSON: %v (%s)", err, c.body)
	}
	return eb
}

// TestProxyOverwritesStaleStaticSign DataEase 静态配置里若残留旧签名/旧 ts/旧 key，
// 代理必须覆盖为本次动态值（否则 5 分钟后必然签名过期）。
func TestProxyOverwritesStaleStaticSign(t *testing.T) {
	logx.SetLevel(logx.ErrorLevel)
	proxyURL, got := newTestProxy(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, `{}`)
	})
	stale := `{"app_key":"stale-key","ts":"1600000000","sign":"deadbeef","path":"/api/v1/x"}`
	resp, err := http.Post(proxyURL+upstreamPath, "application/json", strings.NewReader(stale))
	if err != nil {
		t.Fatalf("代理请求失败: %v", err)
	}
	resp.Body.Close()

	eb := gotBody(t, got)
	if eb.AppKey != "app_e2e" || eb.Ts == "1600000000" || eb.Sign == "deadbeef" {
		t.Errorf("静态残留签名应被动态值覆盖, got: %s", got.body)
	}
	if eb.Path != "/api/v1/x" {
		t.Errorf("业务字段应保留, got: %s", got.body)
	}
}

// TestProxyNonJSONBodyPassthrough 非 JSON body：原样透传给上游（header 仍携带签名）。
func TestProxyNonJSONBodyPassthrough(t *testing.T) {
	logx.SetLevel(logx.ErrorLevel)
	proxyURL, got := newTestProxy(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	resp, err := http.Post(proxyURL+upstreamPath, "text/plain", strings.NewReader("not-json"))
	if err != nil {
		t.Fatalf("代理请求失败: %v", err)
	}
	resp.Body.Close()

	if string(got.body) != "not-json" {
		t.Errorf("非 JSON body 应原样透传, got: %q", got.body)
	}
	if got.header.Get("X-App-Key") != "app_e2e" || got.header.Get("X-Sign") == "" {
		t.Error("header 签名三件套仍应携带")
	}
}

// TestHealthzAndMethodGuard 健康检查与方法守卫。
func TestHealthzAndMethodGuard(t *testing.T) {
	logx.SetLevel(logx.ErrorLevel)
	srv := httptest.NewServer(newProxyHandler("k", "s", "http://127.0.0.1:1"))
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/healthz")
	if err != nil {
		t.Fatalf("healthz 请求失败: %v", err)
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK || string(b) != "ok" {
		t.Errorf("healthz 应 200 ok, got: %d %s", resp.StatusCode, b)
	}

	resp2, err := http.Get(srv.URL + "/other")
	if err != nil {
		t.Fatalf("GET 请求失败: %v", err)
	}
	defer resp2.Body.Close()
	if resp2.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("非 healthz 的 GET 应 405, got: %d", resp2.StatusCode)
	}
}

// TestBodySizeLimit 超 1MB 请求体应 413 且不触达上游。
func TestBodySizeLimit(t *testing.T) {
	logx.SetLevel(logx.ErrorLevel)
	touched := false
	proxyURL, _ := newTestProxy(t, func(_ http.ResponseWriter, _ *http.Request) {
		touched = true
	})
	// 1MB+1 的合法 JSON（可解析但超限，先卡在大小检查）
	big := `{"pad":"` + strings.Repeat("a", maxBodyBytes) + `"}`
	resp, err := http.Post(proxyURL+upstreamPath, "application/json", strings.NewReader(big))
	if err != nil {
		t.Fatalf("代理请求失败: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusRequestEntityTooLarge {
		t.Errorf("超限 body 应 413, got: %d", resp.StatusCode)
	}
	if touched {
		t.Error("超限 body 不应转发到上游")
	}
}
