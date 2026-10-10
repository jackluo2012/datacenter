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
	// ErrInvalidDirection 血缘遍历方向非法（仅支持 upstream/downstream）
	ErrInvalidDirection = errors.New("direction 仅支持 upstream/downstream")
)

// 关系类型
const (
	RelTypeContains = "contains" // 包含关系：database->table、table->column
	RelTypeLineage  = "lineage"  // 血缘关系
	RelTypeDepends  = "depends"  // 依赖关系
)

// 血缘遍历方向
const (
	DirectionUpstream   = "upstream"   // 上游：沿 from_id -> to_id 反向找
	DirectionDownstream = "downstream" // 下游：沿 from_id -> to_id 正向找
)

// DefaultClassification 实体默认分级（L2 内部）
const DefaultClassification = "L2"

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

// escapeLike 转义 like 通配符（%/_/\），防止用户输入干扰模糊匹配
func escapeLike(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `%`, `\%`)
	s = strings.ReplaceAll(s, `_`, `\_`)
	return s
}
