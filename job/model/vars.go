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
)

// 同步任务状态（dc_sync_task.status 状态机：draft --部署--> deployed --触发--> running --终态--> deployed/failed）
const (
	TaskStatusDraft    = "draft"    // 草稿（未部署）
	TaskStatusDeployed = "deployed" // 已部署到调度器并上线
	TaskStatusRunning  = "running"  // 运行中
	TaskStatusFailed   = "failed"   // 最近一次运行失败
)

// 同步模式（dc_sync_task.mode 白名单）
const (
	TaskModeOffline = "offline" // 离线批量同步
	TaskModeCdc     = "cdc"     // 实时增量同步
)

// taskStatusSet 任务状态白名单（列表过滤入参校验用）
var taskStatusSet = map[string]struct{}{
	TaskStatusDraft:    {},
	TaskStatusDeployed: {},
	TaskStatusRunning:  {},
	TaskStatusFailed:   {},
}

// IsValidTaskStatus 任务状态是否在白名单内
func IsValidTaskStatus(status string) bool {
	_, ok := taskStatusSet[status]
	return ok
}

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
