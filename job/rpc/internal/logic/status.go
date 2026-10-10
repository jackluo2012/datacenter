package logic

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"datacenter/job/model"
)

// 同步任务状态流转纯函数（对齐 quality 服务 checker 纯函数风格，全部可单测）

// maxNameLen 任务名长度上限（dc_sync_task.name 为 varchar(128)）
const maxNameLen = 128

// IsValidMode 同步模式是否在白名单内
func IsValidMode(mode string) bool {
	return mode == model.TaskModeOffline || mode == model.TaskModeCdc
}

// ValidateSyncTask 创建同步任务前的静态校验（纯函数，内部做首尾空白防御）：
// 名称非空且 ≤128 字符、mode 白名单（offline/cdc）、source_ds_id>0、
// mapping_json 允许为空，非空时须为 JSON 对象
func ValidateSyncTask(name, mode string, sourceDsId int64, mappingJson string) error {
	name = strings.TrimSpace(name)
	if utf8.RuneCountInString(name) == 0 {
		return ErrTaskNameRequired
	}
	if utf8.RuneCountInString(name) > maxNameLen {
		return ErrTaskNameTooLong
	}
	if !IsValidMode(mode) {
		return ErrInvalidMode
	}
	if sourceDsId <= 0 {
		return ErrSourceDsRequired
	}
	if s := strings.TrimSpace(mappingJson); s != "" {
		var m map[string]any
		if err := json.Unmarshal([]byte(s), &m); err != nil {
			return fmt.Errorf("%w: %v", ErrInvalidMappingJson, err)
		}
	}
	return nil
}

// CanDelete 删除前置检查（纯函数）：仅草稿可直接删除；
// 已部署/运行中/失败的任务在调度器侧仍存在工作流，需先下线再删。
// 返回是否允许删除与拒绝原因
func CanDelete(status string) (bool, string) {
	if status == model.TaskStatusDraft {
		return true, ""
	}
	return false, "任务已部署到调度器，请先在调度器侧下线后再删除"
}

// DolphinScheduler 工作流实例状态（3.x ExecutionStatus 枚举）→ 本地状态词汇映射：
// 运行中 / 成功（终态回落 deployed）/ 失败（终态回落 failed）；无法识别返回空串
var (
	dsRunningStates = map[string]struct{}{
		"RUNNING":              {}, // 兼容旧版本
		"RUNNING_EXECUTION":    {},
		"SUBMITTED_SUCCESS":    {},
		"READY_PAUSE":          {},
		"READY_STOP":           {},
		"NEED_FAULT_TOLERANCE": {},
		"DELAY_EXECUTION":      {},
		"SERIAL_WAIT":          {},
	}
	dsSuccessStates = map[string]struct{}{
		"SUCCESS":        {},
		"FORCED_SUCCESS": {},
	}
	dsFailedStates = map[string]struct{}{
		"FAILURE": {},
		"KILL":    {},
		"STOP":    {},
		"PAUSE":   {},
	}
)

// mapDSState DS 实例状态 → 本地任务状态（纯函数）：
// 运行中→running、成功→deployed（任务仍在调度，仅实例终态为成功）、失败→failed；
// 无法识别返回空串，由调用方保持本地状态
func mapDSState(state string) string {
	switch {
	case containsKey(dsRunningStates, state):
		return model.TaskStatusRunning
	case containsKey(dsSuccessStates, state):
		return model.TaskStatusDeployed
	case containsKey(dsFailedStates, state):
		return model.TaskStatusFailed
	default:
		return ""
	}
}

// AggregateStatus 本地状态 + DS 最新实例状态聚合（纯函数，GetTaskStatus 展示口径，
// 不落库）：DS 状态可识别时以调度器为准，否则回落本地状态
func AggregateStatus(local, dsState string) string {
	if mapped := mapDSState(dsState); mapped != "" {
		return mapped
	}
	return local
}

// loadTask 按 id+beid 加载同步任务（beid/id 缺失校验 + not found 翻译为业务错误）
func loadTask(ctx context.Context, m model.SyncTaskModel, id, beid int64) (*model.SyncTask, error) {
	if beid <= 0 {
		return nil, model.ErrBeidRequired
	}
	if id <= 0 {
		return nil, ErrTaskIdRequired
	}
	task, err := m.FindOne(ctx, id, beid)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, ErrTaskNotFound
		}
		return nil, err
	}
	return task, nil
}

// dsErrMsg 统一拼接 DS 失败原因（部署/触发失败是业务结果而非 error，msg 需可读）
func dsErrMsg(step string, err error) string {
	return fmt.Sprintf("%s: %v", step, err)
}

// containsKey 集合是否包含 key
func containsKey(set map[string]struct{}, key string) bool {
	_, ok := set[key]
	return ok
}
