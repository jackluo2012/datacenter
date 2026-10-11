package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"datacenter/ops/rpc/opsClient"

	"github.com/zeromicro/go-zero/core/logx"
)

// 审计切面常量
const (
	// AuditBodyLimit body 摘要最大字节数（等保三级：操作详情留痕，防止大请求体撑爆日志）
	AuditBodyLimit = 500
	// auditRPCTimeout 异步写审计日志的 RPC 超时保护
	auditRPCTimeout = 3 * time.Second
	// ClientIPCtxKey 客户端 IP 注入 ctx 的 key（供登录日志等下游 logic 取用）
	ClientIPCtxKey = "x-client-ip"
	// UserAgentCtxKey 客户端 UserAgent 注入 ctx 的 key
	UserAgentCtxKey = "x-user-agent"
	// 敏感字段脱敏后的占位值
	auditMaskValue = "******"
)

// needAudit 仅管理面写操作需要审计：POST/PUT/DELETE 落审计，GET 不记
func needAudit(method string) bool {
	switch method {
	case http.MethodPost, http.MethodPut, http.MethodDelete:
		return true
	default:
		return false
	}
}

// ctxInt64Value 从 ctx 取 JWT claims 注入的整数值
// （go-zero jwt 中间件将 claims 逐 key 注入 ctx，值经 json 解码可能是 json.Number/float64，无则返回 0）
func ctxInt64Value(ctx context.Context, key string) int64 {
	val, _ := json.Number(fmt.Sprintf("%v", ctx.Value(key))).Int64()
	return val
}

// ctxStringValue 从 ctx 取字符串值（不存在时返回空串）
func ctxStringValue(ctx context.Context, key string) string {
	val, _ := ctx.Value(key).(string)
	return val
}

// auditAction 按请求方法推断操作类型：POST=create、PUT=update、DELETE=delete
func auditAction(method string) string {
	switch method {
	case http.MethodPost:
		return "create"
	case http.MethodPut:
		return "update"
	case http.MethodDelete:
		return "delete"
	default:
		return "unknown"
	}
}

// auditResource 取路径首段资源名：/datasource/create → datasource（原始路径按 "/" 切分的第二段）
func auditResource(path string) string {
	parts := strings.Split(path, "/") // "/datasource/create" → ["", "datasource", "create"]
	if len(parts) > 1 && parts[1] != "" {
		return parts[1]
	}
	return "unknown"
}

// auditResourceID 取路径末段作为资源 id：纯数字则原样返回，否则视为 "0"
// （如 /datasource/update 路径本身不带 id，实体 id 落在 body/query 中，故记 0）
func auditResourceID(path string) string {
	parts := strings.Split(strings.TrimRight(path, "/"), "/")
	last := parts[len(parts)-1]
	if last == "" {
		return "0"
	}
	var id int64
	if _, err := fmt.Sscanf(last, "%d", &id); err == nil && fmt.Sprint(id) == last {
		return last
	}
	return "0"
}

// clientIP 解析客户端真实 IP：优先反向代理头（X-Real-IP / X-Forwarded-For 首个），
// 兜底 RemoteAddr 去端口
func clientIP(r *http.Request) string {
	if v := strings.TrimSpace(r.Header.Get("X-Real-IP")); v != "" {
		return v
	}
	if v := strings.TrimSpace(r.Header.Get("X-Forwarded-For")); v != "" {
		if i := strings.IndexByte(v, ','); i >= 0 {
			v = v[:i]
		}
		return strings.TrimSpace(v)
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// readBodySnippet 读取请求体前 AuditBodyLimit 字节做摘要，
// 已读部分与剩余 body 用 MultiReader 拼回，保证下游 handler 仍能解析完整请求体
func readBodySnippet(r *http.Request) string {
	if r.Body == nil {
		return ""
	}
	head, err := io.ReadAll(io.LimitReader(r.Body, AuditBodyLimit))
	if err != nil {
		// 读失败时重建空 body，避免下游拿到损坏的 reader
		r.Body = io.NopCloser(strings.NewReader(""))
		return ""
	}
	r.Body = io.NopCloser(io.MultiReader(bytes.NewReader(head), r.Body))
	return string(head)
}

// maskSensitive 对 body 摘要中的密码类字段脱敏（审计日志不落明文口令，等保要求）
// 无法按 JSON 解析时原样返回（仅作 500 字节内摘要，不含完整报文）
func maskSensitive(body string) string {
	trimmed := strings.TrimSpace(body)
	if trimmed == "" {
		return ""
	}
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(trimmed), &m); err != nil {
		return trimmed
	}
	for key := range m {
		lower := strings.ToLower(key)
		if strings.Contains(lower, "password") || strings.Contains(lower, "passwd") ||
			strings.Contains(lower, "secret") || strings.Contains(lower, "token") {
			m[key] = auditMaskValue
		}
	}
	out, err := json.Marshal(m)
	if err != nil {
		return trimmed
	}
	return string(out)
}

// auditEntry 待落库的一条审计记录（提取于请求生命周期内，goroutine 中只使用值拷贝，避免持有 Request）
type auditEntry struct {
	Uid        int64
	Username   string
	Beid       int64
	Action     string
	Resource   string
	ResourceID string
	IP         string
	Path       string
	Query      string
	Body       string
}

// AuditLogMiddleware 管理面审计切面：对全部写操作（POST/PUT/DELETE）自动落审计日志
// （对标等保三级"业务操作 100% 自动审计"），经 server.Use 全局挂载，GET 不记。
// 审计调用异步执行且带超时保护，失败仅记错误日志，绝不阻断业务请求。
type AuditLogMiddleware struct {
	opsRpc opsClient.Ops //数据运维与日志服务（审计日志落库）
}

// NewAuditLogMiddleware opsRpc: 已装配的 OpsRpc 客户端
func NewAuditLogMiddleware(opsRpc opsClient.Ops) *AuditLogMiddleware {
	return &AuditLogMiddleware{opsRpc: opsRpc}
}

func (m *AuditLogMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 所有请求统一注入客户端 ip/ua 到 ctx（登录日志等下游 logic 使用），再放行
		ip := clientIP(r)
		r = r.WithContext(context.WithValue(r.Context(), ClientIPCtxKey, ip))
		r = r.WithContext(context.WithValue(r.Context(), UserAgentCtxKey, r.UserAgent()))

		// GET 等读操作不审计，直接放行
		if !needAudit(r.Method) {
			next(w, r)
			return
		}

		// 请求生命周期内提取审计要素（含 body 摘要与敏感字段脱敏）
		entry := auditEntry{
			Uid:        ctxInt64Value(r.Context(), "uid"),       //JWT claims 注入，无则 0（go-zero jwt 先于 server.Use 执行）
			Username:   ctxStringValue(r.Context(), "username"), //claims 无 username 时为空
			Beid:       ctxInt64Value(r.Context(), "beid"),      //JWT claims 注入的应用 id
			Action:     auditAction(r.Method),
			Resource:   auditResource(r.URL.Path),
			ResourceID: auditResourceID(r.URL.Path),
			IP:         ip,
			Path:       r.URL.Path,
			Query:      r.URL.RawQuery,
			Body:       maskSensitive(readBodySnippet(r)),
		}

		// 放行业务请求
		next(w, r)

		// 异步落审计日志：不阻塞响应返回；失败仅记日志
		go m.writeAuditLog(entry)
	}
}

// writeAuditLog 异步写审计日志：3 秒超时保护 + panic 兜底，失败仅记录错误日志不重试不阻断
func (m *AuditLogMiddleware) writeAuditLog(entry auditEntry) {
	defer func() {
		// 审计属旁路逻辑，任何异常不得影响主流程
		if r := recover(); r != nil {
			logx.Errorf("[审计] 写审计日志 panic: %v, entry=%+v", r, entry)
		}
	}()
	detail, err := json.Marshal(map[string]string{
		"path":  entry.Path,
		"query": entry.Query,
		"body":  entry.Body,
	})
	if err != nil {
		logx.Errorf("[审计] 审计详情序列化失败 path=%s: %v", entry.Path, err)
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), auditRPCTimeout)
	defer cancel()
	if _, err := m.opsRpc.WriteAuditLog(ctx, &opsClient.WriteAuditLogReq{
		Uid:        entry.Uid,
		Username:   entry.Username,
		Action:     entry.Action,
		Resource:   entry.Resource,
		ResourceId: entry.ResourceID,
		DetailJson: string(detail),
		Ip:         entry.IP,
		Beid:       entry.Beid,
	}); err != nil {
		logx.Errorf("[审计] 写审计日志失败 uid=%d action=%s resource=%s path=%s: %v",
			entry.Uid, entry.Action, entry.Resource, entry.Path, err)
	}
}
