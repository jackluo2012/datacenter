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
	// ErrTaskIdRequired 归档任务 id 缺失
	ErrTaskIdRequired = errors.New("任务 id 不能为空")
)

// 归档执行状态（dc_ops_archive_task.last_status / dc_ops_archive_log.status）
const (
	ArchiveStatusSuccess = "success" // 全部批次成功
	ArchiveStatusFailed  = "failed"  // 任一批失败中断
)

// 分页缺省值与上限
const (
	defaultPageSize = 10
	maxPageSize     = 200
)

// placeholders 生成 n 个 ? 占位符（逗号分隔）
func placeholders(n int) string {
	return strings.TrimSuffix(strings.Repeat("?,", n), ",")
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
