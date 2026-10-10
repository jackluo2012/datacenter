package logic

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// DolphinScheduler 配置来源：config 结构体为脚手架生成不可改，故从环境变量读取。
// 未配置 = 调度集成未启用（部署/触发返回业务结果 ok=false，而非进程报错）：
//
//	DS_BASE_URL  DS OpenAPI 基址，如 http://ds:12345/dolphinscheduler
//	DS_TOKEN     API Token（DS 3.2+ 推荐，请求头 token: xxx）
//	DS_PROJECT   目标项目：纯数字按项目编码使用，否则作为项目名按名解析编码
const (
	envDSBaseURL = "DS_BASE_URL"
	envDSToken   = "DS_TOKEN"
	envDSProject = "DS_PROJECT"

	// dsStatusSuccess DS OpenAPI 统一响应壳的业务成功码（StatusSuccess）
	dsStatusSuccess = 0
	// maxRespBytes 响应体读取上限（防异常实例拖爆内存）
	maxRespBytes = 1 << 20
	// defaultDSTimeout 单请求超时
	defaultDSTimeout = 10 * time.Second
)

// DS OpenAPI v3 路径模板（%s=项目编码）
const (
	dsPathProjects          = "/projects"                                     // GET 项目列表（ping / 按名解析项目编码）
	dsPathProcessDefinition = "/projects/%s/process-definition"               // GET 查询 / POST 创建工作流定义
	dsPathRelease           = "/projects/%s/process-definition/%d/release"    // POST 上线工作流（releaseState=ONLINE）
	dsPathStart             = "/projects/%s/executors/start-process-instance" // POST 手动运行工作流
	dsPathProcessInstances  = "/projects/%s/process-instances"                // GET 工作流实例列表（查最新状态）
)

// dsEnvelope DS OpenAPI 统一响应壳：{code, msg, data}
type dsEnvelope struct {
	Code int             `json:"code"`
	Msg  string          `json:"msg"`
	Data json.RawMessage `json:"data"`
}

// dsClient DolphinScheduler OpenAPI 轻量客户端（纯标准库 net/http + encoding/json）
type dsClient struct {
	base    string       // OpenAPI 基址（末尾无 /）
	token   string       // API Token（请求头 token）
	project string       // 项目编码或项目名（env DS_PROJECT）
	hc      *http.Client // 可注入，单测用 httptest.Server
}

// enabled 是否已启用（DS_BASE_URL 未配置 = 未启用，client 为 nil）
func (c *dsClient) enabled() bool {
	return c != nil && c.base != ""
}

// dsClientFromEnv 按环境变量装配客户端；DS_BASE_URL 未配置返回 nil（未启用）。
// getenv 与 http.Client 均可注入，便于单测
func dsClientFromEnv(getenv func(string) string, hc *http.Client) *dsClient {
	base := strings.TrimRight(strings.TrimSpace(getenv(envDSBaseURL)), "/")
	if base == "" {
		return nil
	}
	if hc == nil {
		hc = &http.Client{Timeout: defaultDSTimeout}
	}
	return &dsClient{
		base:    base,
		token:   strings.TrimSpace(getenv(envDSToken)),
		project: strings.TrimSpace(getenv(envDSProject)),
		hc:      hc,
	}
}

// ds 进程内单例：环境变量驱动的默认客户端（未配置返回 nil）
var (
	dsOnce    sync.Once
	dsDefault *dsClient
)

func ds() *dsClient {
	dsOnce.Do(func() {
		dsDefault = dsClientFromEnv(os.Getenv, &http.Client{Timeout: defaultDSTimeout})
	})
	return dsDefault
}

// call 发起一次 DS OpenAPI 请求：构造 URL/请求头/请求体并校验 HTTP 状态码。
// 返回原始响应体（HTTP 层校验），业务码校验在 envelope
func (c *dsClient) call(method, path string, query url.Values, body any) ([]byte, error) {
	var buf io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("DS 请求体序列化失败(%s %s): %w", method, path, err)
		}
		buf = bytes.NewReader(b)
	}
	u := c.base + path
	if len(query) > 0 {
		u += "?" + query.Encode()
	}
	req, err := http.NewRequest(method, u, buf)
	if err != nil {
		return nil, fmt.Errorf("DS 请求构造失败(%s %s): %w", method, path, err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	// TODO(联调): 需真实 DolphinScheduler 3.x 实例验证 —— 3.2+ 推荐 token 请求头鉴权；
	// 更早版本可用 POST {base}/login(userName/userPassword 表单) 换 sessionId(cookie) 替代
	if c.token != "" {
		req.Header.Set("token", c.token)
	}
	resp, err := c.hc.Do(req)
	if err != nil {
		return nil, fmt.Errorf("DS 请求失败(%s %s): %w", method, path, err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, maxRespBytes))
	if err != nil {
		return nil, fmt.Errorf("DS 响应读取失败(%s %s): %w", method, path, err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("DS 响应状态码异常(%s %s): http %d: %s",
			method, path, resp.StatusCode, dsExcerpt(raw))
	}
	return raw, nil
}

// envelope 解析统一响应壳并校验业务码（DS 约定 code=0 为成功），返回 data 段
func (c *dsClient) envelope(method, path string, raw []byte) (json.RawMessage, error) {
	var env dsEnvelope
	if err := json.Unmarshal(raw, &env); err != nil {
		return nil, fmt.Errorf("DS 响应解析失败(%s %s): %w: %s", method, path, err, dsExcerpt(raw))
	}
	if env.Code != dsStatusSuccess {
		return nil, fmt.Errorf("DS 业务失败(%s %s): code=%d msg=%s", method, path, env.Code, env.Msg)
	}
	return env.Data, nil
}

// ping 连通性探测：拉取项目列表第一页（同时校验地址可达与 token 有效）
// TODO(联调): 需真实 DolphinScheduler 3.x 实例验证
func (c *dsClient) ping() error {
	query := url.Values{"pageNo": {"1"}, "pageSize": {"1"}}
	raw, err := c.call(http.MethodGet, dsPathProjects, query, nil)
	if err != nil {
		return err
	}
	_, err = c.envelope(http.MethodGet, dsPathProjects, raw)
	return err
}

// projectItem 项目/工作流列表条目（仅取需要的字段）
type projectItem struct {
	Code int64  `json:"code"`
	Name string `json:"name"`
}

// resolveProject 解析项目编码：纯数字直接使用；否则作为项目名查询解析。
// TODO(联调): 需真实 DolphinScheduler 3.x 实例验证项目列表接口的过滤参数与分页结构
func (c *dsClient) resolveProject() (string, error) {
	p := strings.TrimSpace(c.project)
	if p == "" {
		return "", errors.New("调度器项目未配置：需设置环境变量 DS_PROJECT(项目编码或项目名)")
	}
	if isAllDigits(p) {
		return p, nil
	}
	query := url.Values{"pageNo": {"1"}, "pageSize": {"100"}, "searchVal": {p}}
	raw, err := c.call(http.MethodGet, dsPathProjects, query, nil)
	if err != nil {
		return "", err
	}
	data, err := c.envelope(http.MethodGet, dsPathProjects, raw)
	if err != nil {
		return "", err
	}
	var page struct {
		TotalList []projectItem `json:"totalList"`
	}
	if err := json.Unmarshal(data, &page); err != nil {
		return "", fmt.Errorf("DS 项目列表解析失败: %w: %s", err, dsExcerpt(data))
	}
	for _, item := range page.TotalList {
		if item.Name == p && item.Code > 0 {
			return strconv.FormatInt(item.Code, 10), nil
		}
	}
	return "", fmt.Errorf("调度器项目 %q 不存在", p)
}

// ensureWorkflow 幂等确保工作流定义存在：同名（精确匹配）直接复用其编码，
// 否则创建空任务壳并返回新编码
// TODO(联调): 需真实 DolphinScheduler 3.x 实例验证 —— 3.x 官方创建接口的参数为
// name/description/locations + taskDefinitionJson/taskRelationJson（表单或 JSON），
// 此处按管控面骨架先提交 name/location/description/tasks 空任务壳，联调时校正字段名
func (c *dsClient) ensureWorkflow(name string) (int, error) {
	project, err := c.resolveProject()
	if err != nil {
		return 0, err
	}
	path := fmt.Sprintf(dsPathProcessDefinition, project)

	// 1) 按名查询已有工作流（精确匹配，保证幂等）
	query := url.Values{"pageNo": {"1"}, "pageSize": {"50"}, "searchVal": {name}}
	raw, err := c.call(http.MethodGet, path, query, nil)
	if err != nil {
		return 0, err
	}
	data, err := c.envelope(http.MethodGet, path, raw)
	if err != nil {
		return 0, err
	}
	var page struct {
		TotalList []projectItem `json:"totalList"`
	}
	if err := json.Unmarshal(data, &page); err != nil {
		return 0, fmt.Errorf("DS 工作流列表解析失败(%s): %w: %s", path, err, dsExcerpt(data))
	}
	for _, item := range page.TotalList {
		if item.Name == name && item.Code > 0 {
			return int(item.Code), nil
		}
	}

	// 2) 不存在则创建空任务壳（任务定义后续由执行面推送）
	body := map[string]any{
		"name":        name,
		"location":    "{}",
		"description": "由数据中台 job 服务自动创建的同步任务工作流",
		"tasks":       []any{},
	}
	raw, err = c.call(http.MethodPost, path, nil, body)
	if err != nil {
		return 0, err
	}
	data, err = c.envelope(http.MethodPost, path, raw)
	if err != nil {
		return 0, err
	}
	return parseCreatedCode(data)
}

// release 上线工作流（releaseState=ONLINE，未上线的工作流不允许被调度/手动运行）
// TODO(联调): 需真实 DolphinScheduler 3.x 实例验证
func (c *dsClient) release(code int) error {
	project, err := c.resolveProject()
	if err != nil {
		return err
	}
	path := fmt.Sprintf(dsPathRelease, project, code)
	raw, err := c.call(http.MethodPost, path, nil, map[string]string{"releaseState": "ONLINE"})
	if err != nil {
		return err
	}
	_, err = c.envelope(http.MethodPost, path, raw)
	return err
}

// start 手动运行工作流，返回新实例 id
// TODO(联调): 需真实 DolphinScheduler 3.x 实例验证（补齐 failureStrategy 等运行策略参数）
func (c *dsClient) start(code int) (int, error) {
	project, err := c.resolveProject()
	if err != nil {
		return 0, err
	}
	path := fmt.Sprintf(dsPathStart, project)
	raw, err := c.call(http.MethodPost, path, nil, map[string]any{"processDefinitionCode": code})
	if err != nil {
		return 0, err
	}
	data, err := c.envelope(http.MethodPost, path, raw)
	if err != nil {
		return 0, err
	}
	return parseInstanceId(data)
}

// latestInstanceStatus 查询工作流最新一次实例状态（无实例返回空串）
func (c *dsClient) latestInstanceStatus(code int) (string, error) {
	state, _, err := c.latestInstance(code)
	return state, err
}

// latestInstance 查询工作流最新一次实例的状态与实例 id（无实例返回空串/0）
// TODO(联调): 需真实 DolphinScheduler 3.x 实例验证 processDefinitionCode 过滤参数
func (c *dsClient) latestInstance(code int) (string, int, error) {
	project, err := c.resolveProject()
	if err != nil {
		return "", 0, err
	}
	path := fmt.Sprintf(dsPathProcessInstances, project)
	query := url.Values{
		"pageNo":                {"1"},
		"pageSize":              {"1"},
		"processDefinitionCode": {strconv.Itoa(code)},
	}
	raw, err := c.call(http.MethodGet, path, query, nil)
	if err != nil {
		return "", 0, err
	}
	data, err := c.envelope(http.MethodGet, path, raw)
	if err != nil {
		return "", 0, err
	}
	var page struct {
		TotalList []struct {
			Id    int64  `json:"id"`
			State string `json:"state"`
		} `json:"totalList"`
	}
	if err := json.Unmarshal(data, &page); err != nil {
		return "", 0, fmt.Errorf("DS 实例列表解析失败(%s): %w: %s", path, err, dsExcerpt(data))
	}
	if len(page.TotalList) == 0 {
		return "", 0, nil
	}
	first := page.TotalList[0]
	return first.State, int(first.Id), nil
}

// parseCreatedCode 解析创建工作流响应中的工作流编码（不同版本 data 为对象 {code:n} 或裸数字）
func parseCreatedCode(data json.RawMessage) (int, error) {
	var obj struct {
		Code int64 `json:"code"`
	}
	if err := json.Unmarshal(data, &obj); err == nil && obj.Code > 0 {
		return int(obj.Code), nil
	}
	var n int64
	if err := json.Unmarshal(data, &n); err == nil && n > 0 {
		return int(n), nil
	}
	return 0, fmt.Errorf("DS 创建工作流响应缺少工作流编码: %s", dsExcerpt(data))
}

// parseInstanceId 解析启动实例响应中的实例 id（不同版本 data 为对象 {id:n} 或裸数字）
func parseInstanceId(data json.RawMessage) (int, error) {
	var obj struct {
		Id int64 `json:"id"`
	}
	if err := json.Unmarshal(data, &obj); err == nil && obj.Id > 0 {
		return int(obj.Id), nil
	}
	var n int64
	if err := json.Unmarshal(data, &n); err == nil && n > 0 {
		return int(n), nil
	}
	return 0, fmt.Errorf("DS 启动实例响应缺少实例 id: %s", dsExcerpt(data))
}

// isAllDigits 是否为纯数字串
func isAllDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// dsExcerpt 响应体摘要（截断到 200 字节，错误信息透出用）
func dsExcerpt(raw []byte) string {
	const limit = 200
	if len(raw) > limit {
		return string(raw[:limit]) + "...(截断)"
	}
	return string(raw)
}
