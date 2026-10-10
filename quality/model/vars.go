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
	// ErrJobIdRequired 任务 id 缺失
	ErrJobIdRequired = errors.New("job_id 不能为空")
)

// 任务最近一次执行状态（dc_quality_job.last_status）
const (
	JobStatusSuccess = "success" // 全部规则通过
	JobStatusFailed  = "failed"  // 全部规则不通过
	JobStatusPartial = "partial" // 部分规则通过
)

// 规则状态（dc_quality_rule.status）
const (
	RuleStatusEnabled  = 1 // 启用
	RuleStatusDisabled = 0 // 停用
)

// 分页缺省值与上限
const (
	defaultPageSize = 10
	maxPageSize     = 200
	idQueryChunk    = 500 // id in 查询分片大小
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

// dedupeIds id 去重并保持原有顺序
func dedupeIds(ids []int64) []int64 {
	seen := make(map[int64]struct{}, len(ids))
	uniq := make([]int64, 0, len(ids))
	for _, id := range ids {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		uniq = append(uniq, id)
	}
	return uniq
}
