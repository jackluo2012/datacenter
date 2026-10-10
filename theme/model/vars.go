package model

import (
	"context"
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
	// ErrDupDomainCode 域编码重复（beid+code 唯一）
	ErrDupDomainCode = errors.New("域编码已存在")
	// ErrDupThemeCode 主题编码重复（beid+code 唯一）
	ErrDupThemeCode = errors.New("主题编码已存在")
	// ErrDupTableName 模型表名重复（beid+table_name 唯一）
	ErrDupTableName = errors.New("模型表名已存在")
)

// 分页边界
const (
	defaultPageSize = 10
	maxPageSize     = 200
)

// placeholders 生成 n 个 ? 占位符（逗号分隔）
func placeholders(n int) string {
	return strings.TrimSuffix(strings.Repeat("?,", n), ",")
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

// isDuplicateEntry 识别 MySQL 唯一键冲突（错误码 1062），查重后并发插入撞唯一键的兜底
func isDuplicateEntry(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "duplicate entry") || strings.Contains(msg, "1062")
}

// ensureBeid beid 防御：缺失直接拒绝（多租户隔离第一道闸）
func ensureBeid(beid int64) error {
	if beid <= 0 {
		return ErrBeidRequired
	}
	return nil
}

// queryOne 单行查询统一包装：not found 归一为 ErrNotFound
func queryOne(ctx context.Context, conn sqlx.SqlConn, dest any, query string, args ...any) error {
	err := conn.QueryRowCtx(ctx, dest, query, args...)
	if errors.Is(err, sqlx.ErrNotFound) {
		return ErrNotFound
	}
	return err
}
