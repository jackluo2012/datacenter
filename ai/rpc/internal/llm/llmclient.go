// Package llm OpenAI 兼容 chat/completions 客户端封装（标准库 http 实现，可注入 http.Client 便于 httptest 单测）
package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// RequestTimeout 单次 LLM 请求超时（契约要求 30s）
const RequestTimeout = 30 * time.Second

// maxRespBodySize 响应体读取上限（1MB），防止异常超大响应拖垮内存
const maxRespBodySize = 1 << 20

// Client OpenAI 兼容 LLM 客户端
type Client struct {
	baseUrl    string       //形如 https://api.openai.com/v1，自动拼 /chat/completions
	apiKey     string       //Bearer 鉴权
	model      string       //模型名，如 gpt-4o-mini
	httpClient *http.Client //可注入，便于 httptest 测试
}

// NewClient 创建 LLM 客户端；httpClient 为 nil 时使用带 30s 超时的默认客户端
func NewClient(baseUrl, apiKey, model string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: RequestTimeout}
	}
	return &Client{
		baseUrl:    strings.TrimRight(baseUrl, "/"),
		apiKey:     apiKey,
		model:      model,
		httpClient: httpClient,
	}
}

// chatRequest OpenAI 兼容请求体（仅取所需字段）
type chatRequest struct {
	Model       string        `json:"model"`
	Messages    []chatMessage `json:"messages"`
	Temperature float64       `json:"temperature"` //固定 0，问数场景要求确定性输出
}

// chatMessage 对话消息
type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// chatResponse OpenAI 兼容响应体（仅解析所需字段）
type chatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

// ChatCompletion 发送一次对话补全，返回 choices[0].message.content
// 单 prompt 入参：调用方把系统提示+候选 JSON+用户问题拼成完整 prompt 后传入
func (c *Client) ChatCompletion(ctx context.Context, prompt string) (string, error) {
	//30s 超时兜底：即使注入的 http.Client 未设置 Timeout 也不会无限等待
	ctx, cancel := context.WithTimeout(ctx, RequestTimeout)
	defer cancel()

	body, err := json.Marshal(chatRequest{
		Model:       c.model,
		Messages:    []chatMessage{{Role: "user", Content: prompt}},
		Temperature: 0,
	})
	if err != nil {
		return "", fmt.Errorf("LLM 请求体序列化失败: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseUrl+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("LLM 请求构造失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("LLM 请求失败: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(io.LimitReader(resp.Body, maxRespBodySize))
	if err != nil {
		return "", fmt.Errorf("LLM 响应读取失败: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("LLM 接口返回非 200: status=%d body=%s", resp.StatusCode, truncateForLog(string(data)))
	}

	var cr chatResponse
	if err := json.Unmarshal(data, &cr); err != nil {
		return "", fmt.Errorf("LLM 响应解析失败: %w", err)
	}
	if len(cr.Choices) == 0 {
		return "", errors.New("LLM 响应无 choices")
	}
	return cr.Choices[0].Message.Content, nil
}

// truncateForLog 错误信息里的响应体截断（避免日志超长）
func truncateForLog(s string) string {
	runes := []rune(s)
	if len(runes) <= 256 {
		return s
	}
	return string(runes[:256]) + "..."
}
