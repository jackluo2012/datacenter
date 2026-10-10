package model

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// 通用错误
var (
	// ErrNotFound 记录不存在
	ErrNotFound = sqlx.ErrNotFound
	// ErrBeidRequired 多租户 beid 缺失
	ErrBeidRequired = errors.New("beid 不能为空")
	// ErrModelIdRequired 标准模型 id 缺失
	ErrModelIdRequired = errors.New("model_id 不能为空")
)

// 标准模型状态（dc_std_model.status）
const (
	ModelStatusDraft      = "draft"      // 草稿
	ModelStatusPublished  = "published"  // 已发布
	ModelStatusDeprecated = "deprecated" // 已废弃
)

// 字典状态（dc_std_dict.status）
const (
	DictStatusEnabled  = 1 // 启用
	DictStatusDisabled = 0 // 停用
)

// 分页缺省值与上限
const (
	defaultPageSize = 10
	maxPageSize     = 200
	codeQueryChunk  = 500 // code in 查询分片大小
)

// execCtx 优先在事务 session 上执行写语句，session 为空时退回普通连接
func execCtx(ctx context.Context, conn sqlx.SqlConn, session sqlx.Session, query string, args ...any) (sql.Result, error) {
	if session != nil {
		return session.ExecCtx(ctx, query, args...)
	}
	return conn.ExecCtx(ctx, query, args...)
}

// queryRowsCtx 优先在事务 session 上执行列表查询，session 为空时退回普通连接
func queryRowsCtx(ctx context.Context, conn sqlx.SqlConn, session sqlx.Session, v any, query string, args ...any) error {
	if session != nil {
		return session.QueryRowsCtx(ctx, v, query, args...)
	}
	return conn.QueryRowsCtx(ctx, v, query, args...)
}

// placeholders 生成 n 个 ? 占位符（逗号分隔）
func placeholders(n int) string {
	return strings.TrimSuffix(strings.Repeat("?,", n), ",")
}

// batchValues 生成批量插入的 values 片段：rowCount 组括号占位符（每组 fieldSize 个 ?）
func batchValues(fieldSize, rowCount int) string {
	return strings.TrimSuffix(strings.Repeat("("+placeholders(fieldSize)+"),", rowCount), ",")
}

// escapeLike 转义 like 通配符（%/_/\），防止用户输入干扰模糊匹配
func escapeLike(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `%`, `\%`)
	s = strings.ReplaceAll(s, `_`, `\_`)
	return s
}

// normalizePage 分页参数防御：page 缺省 1，size 缺省 10、封顶 200
func normalizePage(page, size int64) (int64, int64) {
	if page < 1 {
		page = 1
	}
	if size <= 0 {
		size = defaultPageSize
	}
	if size > maxPageSize {
		size = maxPageSize
	}
	return page, size
}

// dedupeStrings 字符串去重并保持原有顺序（过滤空串）
func dedupeStrings(items []string) []string {
	seen := make(map[string]struct{}, len(items))
	uniq := make([]string, 0, len(items))
	for _, item := range items {
		if item == "" {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		uniq = append(uniq, item)
	}
	return uniq
}

// queryRowCtx 优先在事务 session 上执行单行查询，session 为空时退回普通连接
func queryRowCtx(ctx context.Context, conn sqlx.SqlConn, session sqlx.Session, v any, query string, args ...any) error {
	if session != nil {
		return session.QueryRowCtx(ctx, v, query, args...)
	}
	return conn.QueryRowCtx(ctx, v, query, args...)
}
