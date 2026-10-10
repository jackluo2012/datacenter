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
	// ErrDupModelCode 模型编码重复（beid+model_code 唯一）
	ErrDupModelCode = errors.New("模型编码已存在")
	// ErrDupBizCode 业务编码重复（model_id+biz_code 唯一）
	ErrDupBizCode = errors.New("业务编码已存在")
	// ErrSegmentMissing 码段不存在（理论上 FindOrCreate 后不应出现）
	ErrSegmentMissing = errors.New("码段不存在")
	// ErrSegmentStepInvalid 码段步长非法（<=0 无法推进流水）
	ErrSegmentStepInvalid = errors.New("码段步长必须大于 0")
)

// 分页边界
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

// normalizePage 分页参数防御：page 缺省 1，size 缺省 10、上限 200
func normalizePage(page, size int64) (int64, int64) {
	if page <= 0 {
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

// isDuplicateEntry 识别 MySQL 唯一键冲突（错误码 1062），用于查重后并发插入撞唯一键的兜底
func isDuplicateEntry(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "duplicate entry") || strings.Contains(msg, "1062")
}
