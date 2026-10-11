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

// sys_menu.menu_type 菜单类型
const (
	MenuTypeDir    = 1 // 目录
	MenuTypeMenu   = 2 // 菜单
	MenuTypeButton = 3 // 按钮
)

// sys_user.status / sys_role.status 通用状态
const (
	StatusEnabled  = 1 // 启用
	StatusDisabled = 0 // 停用
)

// 分页缺省值与上限
const (
	defaultPageSize = 10
	maxPageSize     = 200
)

// execCtx 优先在事务 session 上执行写语句，session 为空时退回普通连接
func execCtx(ctx context.Context, conn sqlx.SqlConn, session sqlx.Session, query string, args ...any) (sql.Result, error) {
	if session != nil {
		return session.ExecCtx(ctx, query, args...)
	}
	return conn.ExecCtx(ctx, query, args...)
}

// queryRowCtx 优先在事务 session 上执行单行查询，session 为空时退回普通连接
func queryRowCtx(ctx context.Context, conn sqlx.SqlConn, session sqlx.Session, v any, query string, args ...any) error {
	if session != nil {
		return session.QueryRowCtx(ctx, v, query, args...)
	}
	return conn.QueryRowCtx(ctx, v, query, args...)
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

// dedupeInt64 int64 去重并保持原有顺序（过滤非正数，关联表主键 id 均为正）
func dedupeInt64(ids []int64) []int64 {
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
