package logic

import (
	"errors"
	"testing"
)

// TestNextApplyStatus 审批状态机纯函数测试：
// 仅 pending 可审批；approved/rejected 再审批报 ErrApplyAlreadyReviewed
func TestNextApplyStatus(t *testing.T) {
	cases := []struct {
		name       string
		current    string // 当前申请状态
		approve    bool   // 是否通过
		wantStatus string // 期望的下一状态
		wantErr    error  // 期望的业务错误（哨兵）
		wantErrMsg string // 期望的错误信息（非哨兵错误按信息比对）
	}{
		{name: "待审批通过", current: ApplyStatusPending, approve: true, wantStatus: ApplyStatusApproved},
		{name: "待审批驳回", current: ApplyStatusPending, approve: false, wantStatus: ApplyStatusRejected},
		{name: "已通过再审批", current: ApplyStatusApproved, approve: true, wantErr: ErrApplyAlreadyReviewed},
		{name: "已驳回再审批", current: ApplyStatusRejected, approve: false, wantErr: ErrApplyAlreadyReviewed},
		{name: "已通过再审批_驳回操作", current: ApplyStatusApproved, approve: false, wantErr: ErrApplyAlreadyReviewed},
		{name: "已驳回再审批_通过操作", current: ApplyStatusRejected, approve: true, wantErr: ErrApplyAlreadyReviewed},
		{name: "非法状态", current: "unknown", approve: true, wantErrMsg: "非法的申请状态: unknown"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := nextApplyStatus(c.current, c.approve)
			// 校验错误
			switch {
			case c.wantErr == nil && c.wantErrMsg == "":
				if err != nil {
					t.Fatalf("期望成功，实际错误: %v", err)
				}
			case c.wantErr != nil:
				if err == nil {
					t.Fatalf("期望错误 %v，实际成功，状态: %s", c.wantErr, got)
				}
				if !errors.Is(err, c.wantErr) {
					t.Fatalf("期望错误 %v，实际错误: %v", c.wantErr, err)
				}
			default:
				if err == nil {
					t.Fatalf("期望错误 %q，实际成功，状态: %s", c.wantErrMsg, got)
				}
				if err.Error() != c.wantErrMsg {
					t.Fatalf("期望错误 %q，实际错误: %q", c.wantErrMsg, err.Error())
				}
			}
			// 校验状态
			if got != c.wantStatus {
				t.Fatalf("期望状态 %q，实际状态 %q", c.wantStatus, got)
			}
		})
	}
}
