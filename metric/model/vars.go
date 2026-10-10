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
)

// execCtx 优先在事务 session 上执行写语句，session 为空时退回普通连接
func execCtx(ctx context.Context, conn sqlx.SqlConn, session sqlx.Session, query string, args ...any) (sql.Result, error) {
	if session != nil {
		return session.ExecCtx(ctx, query, args...)
	}
	return conn.ExecCtx(ctx, query, args...)
}

// placeholders 生成 n 个 ? 占位符（逗号分隔）
func placeholders(n int) string {
	return strings.TrimSuffix(strings.Repeat("?,", n), ",")
}

// escapeLike 转义 like 通配符（%/_/\），防止用户输入干扰模糊匹配
func escapeLike(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `%`, `\%`)
	s = strings.ReplaceAll(s, `_`, `\_`)
	return s
}

// UniqInt64s int64 切片去重（保持原顺序，跳过非正值）
func UniqInt64s(ids []int64) []int64 {
	seen := make(map[int64]struct{}, len(ids))
	uniq := make([]int64, 0, len(ids))
	for _, id := range ids {
		if id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		uniq = append(uniq, id)
	}
	return uniq
}
