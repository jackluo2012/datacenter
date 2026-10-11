package logic

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"datacenter/quality/model"
)

// 测试告警内容构造：任务名/状态/通过率/失败规则明细齐全，明细按序编号
func TestBuildAlertContent(t *testing.T) {
	failed := []alertFailedRule{
		{Name: "user_id非空", RuleType: "not_null", Detail: `{"actual":3,"expect":0}`},
		{Name: "order_no唯一", RuleType: "unique", Detail: `{"actual":1,"expect":0}`},
	}
	content := BuildAlertContent("订单日检", 12, 6, 4, "partial", failed)
	for _, want := range []string{
		"订单日检",
		"任务ID：12",
		"状态：partial",
		"通过率：66.7%（4/6）",
		"1. [not_null] user_id非空",
		"2. [unique] order_no唯一",
		`{"actual":3,"expect":0}`,
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("告警内容缺少 %q: \n%s", want, content)
		}
	}
	// 全部不通过：通过率 0.0%
	content = BuildAlertContent("订单日检", 12, 6, 0, "failed", failed)
	if !strings.Contains(content, "通过率：0.0%（0/6）") {
		t.Fatalf("全部不通过通过率应为 0.0%%: \n%s", content)
	}
	// 无失败明细（如规则执行出错）不出现明细段
	content = BuildAlertContent("订单日检", 12, 1, 0, "failed", nil)
	if strings.Contains(content, "失败规则明细") {
		t.Fatalf("无失败规则不应出现明细段: \n%s", content)
	}
}

// 测试告警文本截断：超长按 rune 截断（中文不出现半个字），上限 1000 字符
func TestBuildAlertContentTruncate(t *testing.T) {
	failed := make([]alertFailedRule, 0, 200)
	for i := 0; i < 200; i++ {
		failed = append(failed, alertFailedRule{Name: strings.Repeat("长规则名", 30), RuleType: "not_null", Detail: "x"})
	}
	content := BuildAlertContent("订单日检", 1, 200, 0, "failed", failed)
	if got := len([]rune(content)); got != alertContentMaxRunes {
		t.Fatalf("截断后应为 %d 字符, got %d", alertContentMaxRunes, got)
	}
	// 未超长原样返回
	short := BuildAlertContent("t", 1, 1, 1, "partial", nil)
	if len([]rune(short)) >= alertContentMaxRunes {
		t.Fatalf("短内容不应被截断")
	}
	// TruncateRunes 边界
	if TruncateRunes("中文abc", 2) != "中文" {
		t.Fatalf("rune 截断结果不符")
	}
	if TruncateRunes("abc", 0) != "" || TruncateRunes("abc", -1) != "" {
		t.Fatalf("n<=0 应返回空串")
	}
	if TruncateRunes("abc", 10) != "abc" {
		t.Fatalf("n 大于长度应原样返回")
	}
}

// 测试钉钉加签算法：固定输入对比独立生成的真值向量，
// 并校验 URL 拼接格式（timestamp 原值 + sign urlencode）
func TestDingTalkSign(t *testing.T) {
	cases := []struct {
		webhook    string
		secret     string
		ts         int64
		wantSign   string // base64(hmac-sha256(secret, "{ts}\n{secret}"))
		wantEscape string // url.QueryEscape 后的签名
		wantPrefix string // 非空时按前缀校验拼接 URL
	}{
		{
			webhook:    "https://oapi.dingtalk.com/robot/send?access_token=abc",
			secret:     "SEC7d9cf2ed6ac34b28",
			ts:         1700000000000,
			wantSign:   "y6UUwfwPxdvVmbJxsbEGqoBYupP90YSnElsngb5X974=",
			wantEscape: "y6UUwfwPxdvVmbJxsbEGqoBYupP90YSnElsngb5X974%3D",
		},
		{
			webhook:    "http://127.0.0.1:8080/hook",
			secret:     "SECabc+XYZ/123=",
			ts:         1700000000000,
			wantSign:   "827wsAFD4HF/Eu6ieZu4v/mASLNKrj3f0DXzGYx+7SE=",
			wantEscape: "827wsAFD4HF%2FEu6ieZu4v%2FmASLNKrj3f0DXzGYx%2B7SE%3D",
			wantPrefix: "http://127.0.0.1:8080/hook?timestamp=1700000000000&sign=", // 无查询串以 ? 拼接
		},
	}
	for i, c := range cases {
		got := DingTalkSign(c.webhook, c.secret, c.ts)
		if c.wantPrefix != "" {
			if !strings.HasPrefix(got, c.wantPrefix) {
				t.Fatalf("case %d 加签 URL 不符:\n got %s\nwant 前缀 %s", i, got, c.wantPrefix)
			}
			if !strings.HasSuffix(got, c.wantEscape) {
				t.Fatalf("case %d 签名未 urlencode, got %s", i, got)
			}
			continue
		}
		want := c.webhook + "&timestamp=" + strconv.FormatInt(c.ts, 10) + "&sign=" + c.wantEscape
		if got != want {
			t.Fatalf("case %d 加签 URL 不符:\n got %s\nwant %s", i, got, want)
		}
		if !strings.Contains(got, "sign="+c.wantEscape) {
			t.Fatalf("case %d 签名未 urlencode", i)
		}
	}
	// 无 secret 原样返回
	if got := DingTalkSign("https://oapi.dingtalk.com/robot/send?a=1", "", 1); got != "https://oapi.dingtalk.com/robot/send?a=1" {
		t.Fatalf("无 secret 应原样返回, got %s", got)
	}
	// webhook 不带查询串时用 ? 拼接
	got := DingTalkSign("https://oapi.dingtalk.com/robot/send", "s", 1)
	if !strings.HasPrefix(got, "https://oapi.dingtalk.com/robot/send?timestamp=1&sign=") {
		t.Fatalf("无查询串应拼接 ?, got %s", got)
	}
}

// 测试发送体构造：钉钉/飞书两种固定 json 结构
func TestBuildAlertPayload(t *testing.T) {
	ding, err := BuildAlertPayload(model.ChannelTypeDingtalk, "hello")
	if err != nil {
		t.Fatalf("BuildAlertPayload error: %v", err)
	}
	if want := `{"msgtype":"text","text":{"content":"hello"}}`; ding != want {
		t.Fatalf("钉钉消息体 = %s, want %s", ding, want)
	}
	feishu, err := BuildAlertPayload(model.ChannelTypeFeishu, "hello")
	if err != nil {
		t.Fatalf("BuildAlertPayload error: %v", err)
	}
	if want := `{"msg_type":"text","content":{"text":"hello"}}`; feishu != want {
		t.Fatalf("飞书消息体 = %s, want %s", feishu, want)
	}
	// 非法类型报错
	if _, err := BuildAlertPayload("email", "hello"); err == nil {
		t.Fatalf("非法通道类型应报错")
	}
}

// 测试触发级别闸门：failed 仅全挂才报，partial 部分失败即报（含全挂）
func TestLevelAllowed(t *testing.T) {
	cases := []struct {
		minLevel, status string
		want             bool
	}{
		{"failed", "failed", true},
		{"failed", "partial", false},
		{"failed", "success", false},
		{"partial", "failed", true},
		{"partial", "partial", true},
		{"partial", "success", false},
		{"", "failed", true},   // 缺省按 failed 处理
		{"", "partial", false}, // 缺省部分失败不报
		{"hack", "failed", true},
		{"hack", "partial", false},
	}
	for _, c := range cases {
		if got := LevelAllowed(c.minLevel, c.status); got != c.want {
			t.Fatalf("LevelAllowed(%q, %q) = %v, want %v", c.minLevel, c.status, got, c.want)
		}
	}
}

// 测试机器人响应解析：钉钉 errcode / 飞书 code
func TestParseAlertResp(t *testing.T) {
	if err := ParseAlertResp(model.ChannelTypeDingtalk, `{"errcode":0,"errmsg":"ok"}`); err != nil {
		t.Fatalf("钉钉成功响应不应报错: %v", err)
	}
	if err := ParseAlertResp(model.ChannelTypeDingtalk, `{"errcode":310000,"errmsg":"sign not match"}`); err == nil {
		t.Fatalf("钉钉错误响应应报错")
	}
	if err := ParseAlertResp(model.ChannelTypeFeishu, `{"code":0,"msg":"success"}`); err != nil {
		t.Fatalf("飞书成功响应不应报错: %v", err)
	}
	if err := ParseAlertResp(model.ChannelTypeFeishu, `{"code":19021,"msg":"sign match fail"}`); err == nil {
		t.Fatalf("飞书错误响应应报错")
	}
	if err := ParseAlertResp(model.ChannelTypeDingtalk, `not-json`); err == nil {
		t.Fatalf("非 json 响应应报错")
	}
	if err := ParseAlertResp("email", `{}`); err == nil {
		t.Fatalf("非法通道类型应报错")
	}
}

// 测试通道静态校验：类型白名单 + webhook http(s) + 触发级别白名单
func TestValidateChannel(t *testing.T) {
	valid := [][]string{
		{"质量群", "dingtalk", "https://oapi.dingtalk.com/robot/send?access_token=x", "failed"},
		{"质量群", "feishu", "https://open.feishu.cn/open-apis/bot/v2/hook/x", "partial"},
		{"质量群", "feishu", "http://10.0.0.1:9090/hook", ""}, // 内网 http 允许，级别缺省
	}
	for _, c := range valid {
		if err := ValidateChannel(c[0], c[1], c[2], c[3]); err != nil {
			t.Fatalf("合法用例不应报错: %v, got %v", c, err)
		}
	}
	invalid := []struct {
		label          string // 用例名
		chName, chType string // 通道名/类型
		url, level     string
		wantCheck      func(error) bool
	}{
		{"缺通道名", " ", "dingtalk", "https://a.com/x", "", isErrChannelNameRequired},
		{"非法类型", "质量群", "群", "https://a.com/x", "", isErrInvalidChannelType},
		{"email 未开放", "质量群", "email", "https://a.com/x", "", isErrInvalidChannelType},
		{"缺 webhook", "质量群", "dingtalk", "", "", isErrInvalidWebhookURL},
		{"ftp 协议", "质量群", "dingtalk", "ftp://a.com/x", "", isErrInvalidWebhookURL},
		{"无主机", "质量群", "dingtalk", "https://", "", isErrInvalidWebhookURL},
		{"纯文本非 URL", "质量群", "dingtalk", "not a url", "", isErrInvalidWebhookURL},
		{"非法级别", "质量群", "dingtalk", "https://a.com/x", "warn", isErrInvalidAlertLevel},
	}
	for _, c := range invalid {
		err := ValidateChannel(c.chName, c.chType, c.url, c.level)
		if err == nil || !c.wantCheck(err) {
			t.Fatalf("用例 %q 应报对应错误, got %v", c.label, err)
		}
	}
}

func isErrChannelNameRequired(err error) bool { return err == ErrChannelNameRequired }
func isErrInvalidChannelType(err error) bool  { return err == ErrInvalidChannelType }
func isErrInvalidWebhookURL(err error) bool   { return err == ErrInvalidWebhookURL }
func isErrInvalidAlertLevel(err error) bool   { return err == ErrInvalidAlertLevel }

// 测试 webhook 发送闭环（httptest 模拟机器人）：
// 钉钉校验请求体格式与加签参数；飞书校验消息体；服务端报错时发送返回错误
func TestSendAlertToChannel(t *testing.T) {
	// 钉钉：校验 body 与加签参数
	var dingBody, dingSign string
	dingSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		dingBody = string(b)
		dingSign = r.URL.Query().Get("sign")
		w.Write([]byte(`{"errcode":0,"errmsg":"ok"}`))
	}))
	defer dingSrv.Close()

	ch := &model.AlertChannel{
		Type:       model.ChannelTypeDingtalk,
		WebhookUrl: dingSrv.URL + "/robot/send?access_token=x",
		Secret:     "SEC7d9cf2ed6ac34b28",
	}
	if err := sendAlertToChannel(ch, "测试内容"); err != nil {
		t.Fatalf("钉钉发送不应报错: %v", err)
	}
	if want := `{"msgtype":"text","text":{"content":"测试内容"}}`; dingBody != want {
		t.Fatalf("钉钉请求体 = %s, want %s", dingBody, want)
	}
	if dingSign == "" {
		t.Fatalf("带 secret 的钉钉请求应携带加签参数")
	}

	// 飞书：无 secret 不加签，body 为飞书格式
	var feishuBody string
	feishuSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		feishuBody = string(b)
		w.Write([]byte(`{"code":0,"msg":"success"}`))
	}))
	defer feishuSrv.Close()

	ch2 := &model.AlertChannel{Type: model.ChannelTypeFeishu, WebhookUrl: feishuSrv.URL + "/hook"}
	if err := sendAlertToChannel(ch2, "测试内容"); err != nil {
		t.Fatalf("飞书发送不应报错: %v", err)
	}
	if want := `{"msg_type":"text","content":{"text":"测试内容"}}`; feishuBody != want {
		t.Fatalf("飞书请求体 = %s, want %s", feishuBody, want)
	}

	// 机器人业务报错 → 发送返回错误
	ch3 := &model.AlertChannel{Type: model.ChannelTypeDingtalk, WebhookUrl: dingSrv.URL}
	// 复用 dingSrv 但返回错误码需要独立服务
	errSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"errcode":310000,"errmsg":"sign not match"}`))
	}))
	defer errSrv.Close()
	ch3.WebhookUrl = errSrv.URL
	if err := sendAlertToChannel(ch3, "x"); err == nil {
		t.Fatalf("机器人报错时发送应返回错误")
	}

	// 目标不可达 → 发送返回错误（不 panic、带超时）
	ch4 := &model.AlertChannel{Type: model.ChannelTypeFeishu, WebhookUrl: "http://127.0.0.1:1/hook"}
	if err := sendAlertToChannel(ch4, "x"); err == nil {
		t.Fatalf("不可达目标应返回错误")
	}
}
