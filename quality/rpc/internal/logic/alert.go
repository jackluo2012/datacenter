package logic

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"datacenter/quality/model"
	"datacenter/quality/rpc/quality"
)

// 告警相关业务错误（通道静态校验）
var (
	// ErrChannelNameRequired 通道名缺失
	ErrChannelNameRequired = errors.New("通道名不能为空")
	// ErrInvalidChannelType 通道类型不在白名单
	ErrInvalidChannelType = errors.New("通道类型非法：仅支持 dingtalk/feishu")
	// ErrInvalidWebhookURL webhook 地址非法（必须为 http/https 且带主机）
	ErrInvalidWebhookURL = errors.New("webhook_url 非法：必须为 http(s) 地址")
	// ErrInvalidAlertLevel 触发级别不在白名单
	ErrInvalidAlertLevel = errors.New("触发级别非法：仅支持 failed/partial")
	// ErrChannelIdRequired 通道 id 缺失
	ErrChannelIdRequired = errors.New("通道 id 不能为空")
)

const (
	// alertHTTPTimeout 机器人 webhook 发送超时（避免拖慢任务执行链路）
	alertHTTPTimeout = 5 * time.Second
	// alertContentMaxRunes 告警文本上限（字符数；dc_quality_alert_log.content varchar(1024) 字符）
	alertContentMaxRunes = 1000
)

// alertHTTPClient webhook 发送客户端（包级复用，测试可用 httptest 服务端替换目标地址）
var alertHTTPClient = &http.Client{Timeout: alertHTTPTimeout}

// alertFailedRule 失败规则摘要（告警内容构造的输入载体）
type alertFailedRule struct {
	Name     string // 规则名
	RuleType string // 规则类型
	Detail   string // detail_json（阈值/实测/差异）
}

// alertChannelToPb 通道 model -> pb（status 为 int32 收窄转换）
func alertChannelToPb(c *model.AlertChannel) *quality.AlertChannel {
	return &quality.AlertChannel{
		Id:         c.Id,
		Name:       c.Name,
		Type:       c.Type,
		WebhookUrl: c.WebhookUrl,
		Secret:     c.Secret,
		MinLevel:   c.MinLevel,
		Status:     int32(c.Status),
		Beid:       c.Beid,
	}
}

// ValidateChannel 保存通道前的静态校验：名称必填 + 类型白名单 +
// webhook 必须为 http(s) 且带主机 + 触发级别白名单（空串视为缺省 failed，由调用方回填）
func ValidateChannel(name, channelType, webhookURL, minLevel string) error {
	if strings.TrimSpace(name) == "" {
		return ErrChannelNameRequired
	}
	if !model.IsValidChannelType(channelType) {
		return ErrInvalidChannelType
	}
	u, err := url.Parse(strings.TrimSpace(webhookURL))
	if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") {
		return ErrInvalidWebhookURL
	}
	if minLevel != "" && !model.IsValidAlertLevel(minLevel) {
		return ErrInvalidAlertLevel
	}
	return nil
}

// LevelAllowed 当前任务状态是否达到通道的触发级别（纯函数）：
// min_level=failed 仅任务全部不通过才触发；min_level=partial 部分不通过即触发（含 failed）
func LevelAllowed(minLevel, status string) bool {
	switch minLevel {
	case model.AlertLevelPartial:
		return status == model.JobStatusPartial || status == model.JobStatusFailed
	default:
		return status == model.JobStatusFailed
	}
}

// BuildAlertContent 构造告警文本（纯函数）：任务名/状态/通过率/失败规则明细，
// 结果按字符数截断到 alertContentMaxRunes（保证落日志列 varchar(1024) 不超限，且不截断半个汉字）
func BuildAlertContent(jobName string, jobId, total, passed int64, status string, failed []alertFailedRule) string {
	rate := 0.0
	if total > 0 {
		rate = float64(passed) / float64(total) * 100
	}
	var b strings.Builder
	fmt.Fprintf(&b, "【数据质量告警】任务「%s」执行异常\n", jobName)
	fmt.Fprintf(&b, "任务ID：%d | 状态：%s | 通过率：%.1f%%（%d/%d）\n", jobId, status, rate, passed, total)
	if len(failed) > 0 {
		b.WriteString("失败规则明细：\n")
		for i, f := range failed {
			detail := strings.TrimSpace(f.Detail)
			if detail == "" {
				detail = "-"
			}
			fmt.Fprintf(&b, "%d. [%s] %s %s\n", i+1, f.RuleType, f.Name, detail)
		}
	}
	return TruncateRunes(b.String(), alertContentMaxRunes)
}

// TruncateRunes 按 rune 数截断字符串（避免中文按字节截断出现乱码），n<=0 返回空串
func TruncateRunes(s string, n int) string {
	if n <= 0 {
		return ""
	}
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n])
}

// DingTalkSign 钉钉机器人加签（纯函数，ts 为毫秒时间戳）：
// sign = base64(hmac-sha256(key=secret, msg="{ts}\n{secret}"))，
// 返回拼接好 timestamp/sign 参数的 webhook 地址（无 secret 时原样返回）
func DingTalkSign(webhook, secret string, tsMilli int64) string {
	if strings.TrimSpace(secret) == "" {
		return webhook
	}
	stringToSign := fmt.Sprintf("%d\n%s", tsMilli, secret)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(stringToSign))
	sign := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	sep := "&"
	if !strings.Contains(webhook, "?") {
		sep = "?"
	}
	return webhook + sep + "timestamp=" + strconv.FormatInt(tsMilli, 10) +
		"&sign=" + url.QueryEscape(sign)
}

// dingTalkPayload 钉钉机器人 text 消息体
type dingTalkPayload struct {
	MsgType string       `json:"msgtype"`
	Text    dingTalkText `json:"text"`
}

type dingTalkText struct {
	Content string `json:"content"`
}

// feishuPayload 飞书机器人 text 消息体
type feishuPayload struct {
	MsgType string        `json:"msg_type"`
	Content feishuContent `json:"content"`
}

type feishuContent struct {
	Text string `json:"text"`
}

// BuildAlertPayload 按通道类型构造发送体 json（纯函数）：
//
//	dingtalk → {"msgtype":"text","text":{"content":"..."}}
//	feishu   → {"msg_type":"text","content":{"text":"..."}}
func BuildAlertPayload(channelType, content string) (string, error) {
	switch channelType {
	case model.ChannelTypeDingtalk:
		b, err := json.Marshal(dingTalkPayload{MsgType: "text", Text: dingTalkText{Content: content}})
		if err != nil {
			return "", fmt.Errorf("构造钉钉消息体失败: %w", err)
		}
		return string(b), nil
	case model.ChannelTypeFeishu:
		b, err := json.Marshal(feishuPayload{MsgType: "text", Content: feishuContent{Text: content}})
		if err != nil {
			return "", fmt.Errorf("构造飞书消息体失败: %w", err)
		}
		return string(b), nil
	default:
		return "", ErrInvalidChannelType
	}
}

// alertResp 机器人响应（钉钉 errcode/errmsg；飞书 code/msg）
type alertResp struct {
	Errcode int64  `json:"errcode"`
	Errmsg  string `json:"errmsg"`
	Code    int64  `json:"code"`
	Msg     string `json:"msg"`
}

// ParseAlertResp 解析机器人响应（纯函数）：钉钉 errcode==0 / 飞书 code==0 视为成功
func ParseAlertResp(channelType, body string) error {
	var resp alertResp
	if err := json.Unmarshal([]byte(body), &resp); err != nil {
		return fmt.Errorf("机器人响应解析失败: %w", err)
	}
	switch channelType {
	case model.ChannelTypeDingtalk:
		if resp.Errcode != 0 {
			return fmt.Errorf("钉钉机器人返回错误: errcode=%d errmsg=%s", resp.Errcode, resp.Errmsg)
		}
	case model.ChannelTypeFeishu:
		if resp.Code != 0 {
			return fmt.Errorf("飞书机器人返回错误: code=%d msg=%s", resp.Code, resp.Msg)
		}
	default:
		return ErrInvalidChannelType
	}
	return nil
}

// sendAlertToChannel 向单个通道发送告警文本：构造发送体 → 加签 → POST → 校验响应。
// 返回 error 供调用方落告警日志（失败不阻断任务主流程）
func sendAlertToChannel(ch *model.AlertChannel, content string) error {
	payload, err := BuildAlertPayload(ch.Type, content)
	if err != nil {
		return err
	}
	target := DingTalkSign(ch.WebhookUrl, ch.Secret, time.Now().UnixMilli())
	return postWebhook(alertHTTPClient, target, payload, ch.Type)
}

// postWebhook POST json 到机器人并校验响应（响应体最多读 4KB 防大报文）
func postWebhook(client *http.Client, targetURL, payload, channelType string) error {
	resp, err := client.Post(targetURL, "application/json", strings.NewReader(payload))
	if err != nil {
		return fmt.Errorf("webhook 请求失败: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if err != nil {
		return fmt.Errorf("webhook 响应读取失败: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("webhook 响应码异常: http %d body=%s", resp.StatusCode, string(body))
	}
	return ParseAlertResp(channelType, string(body))
}
