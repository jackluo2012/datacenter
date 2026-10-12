// signproxy 独立签名代理：解决 DataEase v1.18 API 数据源只能配静态 header、
// 而平台 /datasvc/execute 需要 HMAC 动态签名（ts ±300 秒有效）的矛盾。
//
// 链路（docs/DEPLOY-DATAEASE.md 方案 c，平台零改码）：
//
//	DataEase API 数据源（静态 body：path/params_json，header 留空）
//	  → 本代理（现算 ts/sign，注入 body 与 header）→ 平台网关 /datasvc/execute → 原样返回
//
// 签名算法与 datasvc/rpc/internal/logic/credential.go 的 genSign 逐字一致：
//
//	sign = hex(hmac-sha256(key=app_secret, msg=app_key+ts))，ts 为秒级 unix 时间戳。
//
// 说明：网关 ExecuteApiReq（internal/types）从 JSON body 解析 app_key/ts/sign 三字段，
// 故代理将签名注入 body；同时按对外 OpenAPI 文档契约回写 X-App-Key/X-Timestamp/X-Sign
// 三个 header，两种姿势均兼容。
package main

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

// 环境变量名
const (
	envListen    = "PROXY_LISTEN"
	envUpstream  = "UPSTREAM"
	envAppKey    = "APP_KEY"
	envAppSecret = "APP_SECRET"
)

// 默认值与运行常量
const (
	// defListen 默认监听 :8077（8090 已被 metric rpc 占用，8091 亦冲突，故取 8077）
	defListen = ":8077"
	// defUpstream 默认上游平台网关（datacenter api 服务）
	defUpstream = "http://127.0.0.1:8887"
	// upstreamPath 平台统一执行入口（唯一免 JWT 入口）
	upstreamPath = "/datasvc/execute"
	// maxBodyBytes 请求体上限 1MB：DataEase 侧仅传 path/params_json，足够且防滥用
	maxBodyBytes = 1 << 20
	// upstreamTimeout 上游执行超时：datasvc 侧可能真实跑 SQL，放宽到 60 秒
	upstreamTimeout = 60 * time.Second
	// shutdownTimeout 收到退出信号后排空在途请求的等待时长
	shutdownTimeout = 5 * time.Second
)

// genSign 计算执行签名 —— 逐字对齐 datasvc/rpc/internal/logic/credential.go 的 genSign：
// hex(hmac-sha256(key=app_secret, msg=app_key+ts))；appSecret 即 key，appKey+ts 拼接即消息。
func genSign(appKey, ts, appSecret string) string {
	mac := hmac.New(sha256.New, []byte(appSecret))
	mac.Write([]byte(appKey + ts))
	return hex.EncodeToString(mac.Sum(nil))
}

// proxyHandler 代理主处理器
type proxyHandler struct {
	appKey    string
	appSecret string
	upstream  string // 形如 http://127.0.0.1:8887（无尾斜杠）
	client    *http.Client
}

// newProxyHandler 组装处理器（测试同构复用：httptest 直接挂它即可）
func newProxyHandler(appKey, appSecret, upstream string) http.Handler {
	return &proxyHandler{
		appKey:    appKey,
		appSecret: appSecret,
		upstream:  strings.TrimRight(upstream, "/"),
		client:    &http.Client{Timeout: upstreamTimeout},
	}
}

// ServeHTTP 路由：GET /healthz → 健康检查；其余仅接受 POST（DataEase API 数据源固定 POST）。
func (p *proxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	if r.URL.Path == "/healthz" {
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, "ok")
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "signproxy 仅支持 POST（健康检查为 GET /healthz）", http.StatusMethodNotAllowed)
		return
	}
	status := p.forward(w, r)
	// 访问日志：时间（logx 自带）/path/耗时/上游状态码
	logx.Infof("signproxy %s %s 耗时=%v 上游状态=%d", r.Method, r.URL.Path, time.Since(start), status)
}

// forward 读 body → 现算签名注入 → 转发上游 → 原样回写响应；返回上游状态码（失败路径返回 0）。
func (p *proxyHandler) forward(w http.ResponseWriter, r *http.Request) int {
	// 读请求体（限 1MB，超限直接 413，不转发）
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxBodyBytes))
	if err != nil {
		var mbe *http.MaxBytesError
		if errors.As(err, &mbe) {
			http.Error(w, fmt.Sprintf("请求体超过 %d 字节上限", maxBodyBytes), http.StatusRequestEntityTooLarge)
			return 0
		}
		http.Error(w, "读取请求体失败: "+err.Error(), http.StatusBadRequest)
		return 0
	}

	// 现算动态签名（ts 秒级 unix，网关侧 ±300 秒强校验，必须每次现算）
	ts := strconv.FormatInt(time.Now().Unix(), 10)
	sign := genSign(p.appKey, ts, p.appSecret)

	// body 注入签名三字段（覆盖 DataEase 静态配置里的残留旧值）；
	// 非 JSON 体无法注入则原样透传，仅靠 header 携带（打日志提示）。
	injected := false
	var obj map[string]any
	if json.Unmarshal(body, &obj) == nil {
		obj["app_key"] = p.appKey
		obj["ts"] = ts // 网关契约 ts 为字符串（ExecuteApiReq.Ts json:"ts" → string）
		obj["sign"] = sign
		if b, merr := json.Marshal(obj); merr == nil {
			body = b
			injected = true
		}
	}
	if !injected {
		logx.Errorf("请求体非合法 JSON 对象，签名未注入 body（仅 header 携带）: %s", r.URL.Path)
	}

	// 构造上游请求：POST UPSTREAM/datasvc/execute
	upReq, err := http.NewRequestWithContext(r.Context(), http.MethodPost, p.upstream+upstreamPath, bytes.NewReader(body))
	if err != nil {
		http.Error(w, "构造上游请求失败: "+err.Error(), http.StatusBadGateway)
		return 0
	}
	upReq.Header.Set("Content-Type", "application/json")
	// 签名 header 三件套（与对外 OpenAPI 文档契约一致；body 与 header 双通道）
	upReq.Header.Set("X-App-Key", p.appKey)
	upReq.Header.Set("X-Timestamp", ts)
	upReq.Header.Set("X-Sign", sign)

	resp, err := p.client.Do(upReq)
	if err != nil {
		http.Error(w, "上游请求失败: "+err.Error(), http.StatusBadGateway)
		return 0
	}
	defer resp.Body.Close()

	// 原样返回上游响应：状态码 + body + Content-Type
	if ct := resp.Header.Get("Content-Type"); ct != "" {
		w.Header().Set("Content-Type", ct)
	}
	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, resp.Body)
	return resp.StatusCode
}

func main() {
	listen := strings.TrimSpace(os.Getenv(envListen))
	if listen == "" {
		listen = defListen
	}
	upstream := strings.TrimSpace(os.Getenv(envUpstream))
	if upstream == "" {
		upstream = defUpstream
	}
	appKey := strings.TrimSpace(os.Getenv(envAppKey))
	appSecret := strings.TrimSpace(os.Getenv(envAppSecret))
	// 凭据启动强校验：缺失立即退出，避免起一个必然 403 的空转进程
	if appKey == "" || appSecret == "" {
		logx.Errorf("启动失败：环境变量 %s 与 %s 均不能为空", envAppKey, envAppSecret)
		os.Exit(1)
	}

	srv := &http.Server{
		Addr:              listen,
		Handler:           newProxyHandler(appKey, appSecret, upstream),
		ReadHeaderTimeout: 10 * time.Second,
	}

	// 优雅退出：SIGINT/SIGTERM → 停止接新请求，最多等 shutdownTimeout 排空在途请求
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go func() {
		<-ctx.Done()
		logx.Info("收到退出信号，开始优雅退出……")
		sctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		_ = srv.Shutdown(sctx)
	}()

	logx.Infof("signproxy 启动：监听 %s → 上游 %s%s（app_key=%s）", listen, upstream, upstreamPath, appKey)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logx.Errorf("服务异常退出: %v", err)
		os.Exit(1)
	}
	logx.Info("signproxy 已退出")
}
