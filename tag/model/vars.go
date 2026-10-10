package model

import (
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
	// ErrDupTagKey 同一对象下 tag_key 重复（object_id+tag_key+beid 唯一）
	ErrDupTagKey = errors.New("同对象下 tag_key 已存在")
)

// placeholders 生成 n 个 ? 占位符（逗号分隔）
func placeholders(n int) string {
	return strings.TrimSuffix(strings.Repeat("?,", n), ",")
}

// isDuplicateEntry 识别 MySQL 唯一键冲突（驱动报错文本含 duplicate entry / 错误码 1062），
// 用于 Create 查重后并发插入撞 uk_obj_key 的兜底
func isDuplicateEntry(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "duplicate entry") || strings.Contains(msg, "1062")
}
