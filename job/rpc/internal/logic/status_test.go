package logic

import (
	"errors"
	"testing"

	"datacenter/job/model"
)

// TestValidateSyncTask 创建任务的静态校验（名称/mode 白名单/source_ds_id/mapping_json）
func TestValidateSyncTask(t *testing.T) {
	cases := []struct {
		name        string
		taskName    string
		mode        string
		sourceDsId  int64
		mappingJson string
		wantErr     error
	}{
		{name: "离线任务合法", taskName: "订单同步", mode: "offline", sourceDsId: 1},
		{name: "cdc 任务合法", taskName: "订单同步", mode: "cdc", sourceDsId: 1, mappingJson: `{"table":"orders"}`},
		{name: "名称为空", taskName: "  ", mode: "offline", sourceDsId: 1, wantErr: ErrTaskNameRequired},
		{name: "名称超长", taskName: repeat("名", 129), mode: "offline", sourceDsId: 1, wantErr: ErrTaskNameTooLong},
		{name: "mode 不在白名单", taskName: "t", mode: "stream", sourceDsId: 1, wantErr: ErrInvalidMode},
		{name: "mode 为空", taskName: "t", mode: "", sourceDsId: 1, wantErr: ErrInvalidMode},
		{name: "源数据源缺失", taskName: "t", mode: "offline", sourceDsId: 0, wantErr: ErrSourceDsRequired},
		{name: "mapping_json 非法", taskName: "t", mode: "offline", sourceDsId: 1, mappingJson: "{bad", wantErr: ErrInvalidMappingJson},
		{name: "mapping_json 须为对象", taskName: "t", mode: "offline", sourceDsId: 1, mappingJson: "[1,2]", wantErr: ErrInvalidMappingJson},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := ValidateSyncTask(c.taskName, c.mode, c.sourceDsId, c.mappingJson)
			if !errors.Is(err, c.wantErr) {
				t.Fatalf("ValidateSyncTask = %v, want %v", err, c.wantErr)
			}
		})
	}
}

// repeat 生成 n 次重复字符串（测试辅助，避免 strings.Builder 样板）
func repeat(s string, n int) string {
	b := make([]byte, 0, len(s)*n)
	for i := 0; i < n; i++ {
		b = append(b, s...)
	}
	return string(b)
}

// TestCanDelete 删除前置检查：仅草稿可删，已部署/运行中/失败需先下线
func TestCanDelete(t *testing.T) {
	for _, status := range []string{model.TaskStatusDraft} {
		if ok, reason := CanDelete(status); !ok || reason != "" {
			t.Fatalf("状态 %s 应可直接删除, got ok=%v reason=%q", status, ok, reason)
		}
	}
	for _, status := range []string{model.TaskStatusDeployed, model.TaskStatusRunning, model.TaskStatusFailed} {
		if ok, reason := CanDelete(status); ok || reason == "" {
			t.Fatalf("状态 %s 应拒绝删除并给出原因, got ok=%v reason=%q", status, ok, reason)
		}
	}
}

// TestMapDSState DS 实例状态 → 本地状态映射
func TestMapDSState(t *testing.T) {
	cases := []struct {
		state string
		want  string
	}{
		{"RUNNING_EXECUTION", model.TaskStatusRunning},
		{"RUNNING", model.TaskStatusRunning}, // 兼容旧版本状态名
		{"SUBMITTED_SUCCESS", model.TaskStatusRunning},
		{"DELAY_EXECUTION", model.TaskStatusRunning},
		{"SERIAL_WAIT", model.TaskStatusRunning},
		{"SUCCESS", model.TaskStatusDeployed},
		{"FORCED_SUCCESS", model.TaskStatusDeployed},
		{"FAILURE", model.TaskStatusFailed},
		{"KILL", model.TaskStatusFailed},
		{"STOP", model.TaskStatusFailed},
		{"PAUSE", model.TaskStatusFailed},
		{"UNKNOWN_STATE", ""}, // 无法识别回落空串
		{"", ""},              // 无实例
	}
	for _, c := range cases {
		if got := mapDSState(c.state); got != c.want {
			t.Fatalf("mapDSState(%q) = %q, want %q", c.state, got, c.want)
		}
	}
}

// TestAggregateStatus 本地状态 + DS 实例状态聚合：DS 可识别时以调度器为准，否则回落本地
func TestAggregateStatus(t *testing.T) {
	cases := []struct {
		local   string
		dsState string
		want    string
	}{
		// DS 运行中 → running
		{model.TaskStatusDeployed, "RUNNING_EXECUTION", model.TaskStatusRunning},
		// DS 成功终态 → deployed（任务仍在调度）
		{model.TaskStatusRunning, "SUCCESS", model.TaskStatusDeployed},
		// DS 失败终态 → failed
		{model.TaskStatusDeployed, "FAILURE", model.TaskStatusFailed},
		// DS 状态无法识别/无实例 → 保持本地
		{model.TaskStatusDraft, "", model.TaskStatusDraft},
		{model.TaskStatusDeployed, "WHAT_EVER", model.TaskStatusDeployed},
	}
	for _, c := range cases {
		if got := AggregateStatus(c.local, c.dsState); got != c.want {
			t.Fatalf("AggregateStatus(%q, %q) = %q, want %q", c.local, c.dsState, got, c.want)
		}
	}
}
