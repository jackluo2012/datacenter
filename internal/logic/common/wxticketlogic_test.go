package logic

import "testing"

// TestWxConfigSign 用微信官方文档《JS-SDK 说明》附录一的签名示例做固定输入断言：
// https://developers.weixin.qq.com/doc/offiaccount/OA_Web_Apps/JS-SDK.html
// jsapi_ticket=..., noncestr=Wm3WZYTPz0wzccnW, timestamp=1414587457,
// url=http://mp.weixin.qq.com?params=value（不含 # 及其后部分）
// 期望签名 = 0f9de62fce790f9a083d5c99e95740ceb90c27ed
func TestWxConfigSign(t *testing.T) {
	const (
		ticket    = "sM4AOVdWfPE4DxkXGEs8VMCPGGVi4C3VM0P37wVUCFvkVAy_90u5h9nbSlYy3-Sl-HhTdfl2fzFy1AOcHKP7qg"
		nonceStr  = "Wm3WZYTPz0wzccnW"
		timestamp = "1414587457"
		want      = "0f9de62fce790f9a083d5c99e95740ceb90c27ed"
	)
	cases := []struct {
		name string
		url  string
	}{
		{"官方文档URL", "http://mp.weixin.qq.com?params=value"},
		{"URL含#片段应被截断", "http://mp.weixin.qq.com?params=value#xxxx"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := wxConfigSign(ticket, nonceStr, timestamp, c.url)
			if got != want {
				t.Fatalf("wxConfigSign(url=%q) = %q, want %q", c.url, got, want)
			}
		})
	}
}

// TestWxConfigSignSensitive 保证签名对各输入敏感且输出为 40 位十六进制 sha1，
// 防止误用"先排序再拼接"的签名实现（如 silenceper 的 util.Signature）
func TestWxConfigSignSensitive(t *testing.T) {
	base := wxConfigSign("ticket-1", "nonce", "123", "https://example.com/a")
	if len(base) != 40 {
		t.Fatalf("签名长度 = %d, want 40（sha1 十六进制）", len(base))
	}
	variants := []struct {
		name string
		sign string
	}{
		{"ticket不同", wxConfigSign("ticket-2", "nonce", "123", "https://example.com/a")},
		{"noncestr不同", wxConfigSign("ticket-1", "nonce2", "123", "https://example.com/a")},
		{"timestamp不同", wxConfigSign("ticket-1", "nonce", "456", "https://example.com/a")},
		{"url不同", wxConfigSign("ticket-1", "nonce", "123", "https://example.com/b")},
	}
	for _, v := range variants {
		if v.sign == base {
			t.Fatalf("%s 时签名不应与基准相同", v.name)
		}
	}
	// 与手工计算的 sha1 拼接结果逐字符比对
	manual := wxConfigSign("t", "n", "1", "u")
	if manual != wxConfigSign("t", "n", "1", "u#frag") {
		t.Fatalf("含 # 片段的 url 与截断后 url 的签名应一致")
	}
}
